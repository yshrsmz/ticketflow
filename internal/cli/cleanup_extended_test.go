package cli

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/mocks"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestCleanupResult_HasErrors(t *testing.T) {
	tests := []struct {
		name     string
		result   CleanupResult
		expected bool
	}{
		{
			name: "no errors",
			result: CleanupResult{
				OrphanedWorktrees: 2,
				StaleBranches:     3,
				Errors:            []string{},
			},
			expected: false,
		},
		{
			name: "with errors",
			result: CleanupResult{
				OrphanedWorktrees: 1,
				StaleBranches:     0,
				Errors:            []string{"failed to remove worktree", "branch delete failed"},
			},
			expected: true,
		},
		{
			name:     "empty result",
			result:   CleanupResult{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.HasErrors())
		})
	}
}

func TestAutoCleanup(t *testing.T) {
	tests := []struct {
		name           string
		dryRun         bool
		worktreeEnabled bool
		setupMocks     func(*mocks.MockGitClient, *mocks.MockTicketManager)
		expectedResult *CleanupResult
		expectedError  bool
	}{
		{
			name:           "successful cleanup with worktrees enabled",
			dryRun:         false,
			worktreeEnabled: true,
			setupMocks: func(g *mocks.MockGitClient, m *mocks.MockTicketManager) {
				// Mock git operations for orphaned worktree cleanup
				g.On("ListWorktrees", mock.Anything).Return([]git.WorktreeInfo{
					{Branch: "250101-120000-old-feature", Path: "/worktrees/old-feature"},
					{Branch: "250102-120000-active-feature", Path: "/worktrees/active"},
					{Branch: "main", Path: "/main"},
				}, nil)

				// Mock ticket lookups - only doing tickets are considered "active"
				m.On("List", mock.Anything, ticket.StatusFilterDoing).Return([]ticket.Ticket{
					{ID: "250102-120000-active-feature"},
				}, nil)

				// Mock cleanup operations - only old-feature worktree will be removed since it's not in doing status
				g.On("RemoveWorktree", mock.Anything, "/worktrees/old-feature").Return(nil)
				g.On("PruneWorktrees", mock.Anything).Return(nil)

				// Mock stale branch cleanup
				g.On("Exec", mock.Anything, "branch", "--format=%(refname:short)").Return("main\n250101-120000-old-feature\n250102-120000-active-feature\n250103-120000-done-ticket", nil)
				// cleanStaleBranches calls List with "all" to get all tickets
				// Create tickets with proper Status() method
				doneTime1, _ := time.Parse(time.RFC3339, "2025-01-01T14:00:00Z")
				doneTime3, _ := time.Parse(time.RFC3339, "2025-01-01T15:00:00Z")
				startTime, _ := time.Parse(time.RFC3339, "2025-01-01T13:00:00Z")
				
				doneTicket1 := ticket.Ticket{ID: "250101-120000-old-feature"}
				doneTicket1.ClosedAt = ticket.RFC3339TimePtr{Time: &doneTime1}
				doneTicket3 := ticket.Ticket{ID: "250103-120000-done-ticket"}
				doneTicket3.ClosedAt = ticket.RFC3339TimePtr{Time: &doneTime3}
				activeTicket := ticket.Ticket{ID: "250102-120000-active-feature"}
				activeTicket.StartedAt = ticket.RFC3339TimePtr{Time: &startTime}
				
				m.On("List", mock.Anything, ticket.StatusFilterAll).Return([]ticket.Ticket{
					doneTicket1,
					activeTicket,
					doneTicket3,
				}, nil)
				g.On("Exec", mock.Anything, "branch", "-D", "250101-120000-old-feature").Return("", nil)
				g.On("Exec", mock.Anything, "branch", "-D", "250103-120000-done-ticket").Return("", nil)
			},
			expectedResult: &CleanupResult{
				OrphanedWorktrees: 1,  // Only old-feature worktree is orphaned (not in doing status)
				StaleBranches:     2,  // Two done tickets will have their branches removed
				Errors:            []string{},
			},
			expectedError: false,
		},
		{
			name:           "dry run mode",
			dryRun:         true,
			worktreeEnabled: true,
			setupMocks: func(g *mocks.MockGitClient, m *mocks.MockTicketManager) {
				g.On("ListWorktrees", mock.Anything).Return([]git.WorktreeInfo{
					{Branch: "250101-120000-orphaned", Path: "/worktrees/orphaned"},
				}, nil)
				m.On("List", mock.Anything, ticket.StatusFilterDoing).Return([]ticket.Ticket{}, nil)

				g.On("Exec", mock.Anything, "branch", "--format=%(refname:short)").Return("main\n250101-120000-orphaned\n250102-120000-done-ticket", nil)
				// Create a done ticket for branch cleanup
				doneTime, _ := time.Parse(time.RFC3339, "2025-01-01T14:00:00Z")
				doneTicket := ticket.Ticket{ID: "250102-120000-done-ticket"}
				doneTicket.ClosedAt = ticket.RFC3339TimePtr{Time: &doneTime}
				m.On("List", mock.Anything, ticket.StatusFilterAll).Return([]ticket.Ticket{
					doneTicket,
				}, nil)

				// In dry run, no actual deletion should happen
			},
			expectedResult: &CleanupResult{
				OrphanedWorktrees: 1,  // orphaned worktree would be removed
				StaleBranches:     1,  // done ticket branch would be removed
				Errors:            []string{},
			},
			expectedError: false,
		},
		{
			name:           "worktrees disabled",
			dryRun:         false,
			worktreeEnabled: false,
			setupMocks: func(g *mocks.MockGitClient, m *mocks.MockTicketManager) {
				// Only stale branch cleanup should run
				g.On("Exec", mock.Anything, "branch", "--format=%(refname:short)").Return("main\n250101-120000-done-ticket", nil)
				// Create a done ticket for branch cleanup
				doneTime, _ := time.Parse(time.RFC3339, "2025-01-01T14:00:00Z")
				doneTicket := ticket.Ticket{ID: "250101-120000-done-ticket"}
				doneTicket.ClosedAt = ticket.RFC3339TimePtr{Time: &doneTime}
				m.On("List", mock.Anything, ticket.StatusFilterAll).Return([]ticket.Ticket{
					doneTicket,
				}, nil)
				g.On("Exec", mock.Anything, "branch", "-D", "250101-120000-done-ticket").Return("", nil)
			},
			expectedResult: &CleanupResult{
				OrphanedWorktrees: 0,
				StaleBranches:     1,
				Errors:            []string{},
			},
			expectedError: false,
		},
		{
			name:           "with errors",
			dryRun:         false,
			worktreeEnabled: true,
			setupMocks: func(g *mocks.MockGitClient, m *mocks.MockTicketManager) {
				// Worktree list fails
				g.On("ListWorktrees", mock.Anything).Return(nil, fmt.Errorf("git error"))
				g.On("PruneWorktrees", mock.Anything).Return(nil)  // This still gets called

				// Branch cleanup continues
				g.On("Exec", mock.Anything, "branch", "--format=%(refname:short)").Return("", fmt.Errorf("branch list failed"))
			},
			expectedResult: &CleanupResult{
				OrphanedWorktrees: 0,
				StaleBranches:     0,
				Errors:            []string{"worktrees: git error", "branches: branch list failed"},
			},
			expectedError: false, // AutoCleanup returns errors in result, not as error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			
			mockGit := new(mocks.MockGitClient)
			mockManager := new(mocks.MockTicketManager)

			cfg := config.Default()
			cfg.Worktree.Enabled = tt.worktreeEnabled

			app := &App{
				Config:  cfg,
				Git:     mockGit,
				Manager: mockManager,
			}

			tt.setupMocks(mockGit, mockManager)

			result, err := app.AutoCleanup(ctx, tt.dryRun)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.OrphanedWorktrees, result.OrphanedWorktrees)
				assert.Equal(t, tt.expectedResult.StaleBranches, result.StaleBranches)
				assert.Equal(t, len(tt.expectedResult.Errors), len(result.Errors))
			}

			mockGit.AssertExpectations(t)
			mockManager.AssertExpectations(t)
		})
	}
}

func TestCleanOrphanedWorktrees(t *testing.T) {
	tests := []struct {
		name           string
		dryRun         bool
		worktrees      []git.WorktreeInfo
		activeTickets  []ticket.Ticket // Tickets in doing status
		expectedCount  int
		expectedError  bool
	}{
		{
			name:   "remove orphaned worktree",
			dryRun: false,
			worktrees: []git.WorktreeInfo{
				{Branch: "250101-120000-orphaned", Path: "/worktrees/orphaned"},
				{Branch: "main", Path: "/main"},
			},
			activeTickets: []ticket.Ticket{}, // No active tickets
			expectedCount: 1,
			expectedError: false,
		},
		{
			name:   "skip default and empty branches",
			dryRun: false,
			worktrees: []git.WorktreeInfo{
				{Branch: "main", Path: "/main"},
				{Branch: "", Path: "/empty"},
				{Branch: "250101-120000-feature", Path: "/feature"},
			},
			activeTickets: []ticket.Ticket{},
			expectedCount: 1, // Only feature branch gets removed
			expectedError: false,
		},
		{
			name:   "skip active ticket worktree",
			dryRun: false,
			worktrees: []git.WorktreeInfo{
				{Branch: "250101-120000-active", Path: "/worktrees/active"},
			},
			activeTickets: []ticket.Ticket{
				{ID: "250101-120000-active"},
			},
			expectedCount: 0,
			expectedError: false,
		},
		{
			name:   "multiple orphaned worktrees",
			dryRun: false,
			worktrees: []git.WorktreeInfo{
				{Branch: "250101-120000-old1", Path: "/worktrees/old1"},
				{Branch: "250102-120000-old2", Path: "/worktrees/old2"},
				{Branch: "250103-120000-active", Path: "/worktrees/active"},
			},
			activeTickets: []ticket.Ticket{
				{ID: "250103-120000-active"},
			},
			expectedCount: 2,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			
			mockGit := new(mocks.MockGitClient)
			mockManager := new(mocks.MockTicketManager)

			app := &App{
				Config:  config.Default(),
				Git:     mockGit,
				Manager: mockManager,
			}

			// Setup mocks
			mockGit.On("ListWorktrees", mock.Anything).Return(tt.worktrees, nil)
			mockGit.On("PruneWorktrees", mock.Anything).Return(nil)
			mockManager.On("List", mock.Anything, ticket.StatusFilterDoing).Return(tt.activeTickets, nil)

			// Setup removal mocks for orphaned worktrees
			if !tt.dryRun {
				// Build map of active tickets
				activeMap := make(map[string]bool)
				for _, t := range tt.activeTickets {
					activeMap[t.ID] = true
				}
				
				// Default branch from config
				defaultBranch := app.Config.Git.DefaultBranch
				
				for _, wt := range tt.worktrees {
					// Skip empty or default branch (matches the actual logic)
					if wt.Branch == "" || wt.Branch == defaultBranch {
						continue
					}
					// If not in active map, it will be removed
					if !activeMap[wt.Branch] {
						mockGit.On("RemoveWorktree", mock.Anything, wt.Path).Return(nil)
					}
				}
			}

			count, err := app.cleanOrphanedWorktrees(ctx, tt.dryRun)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
			}

			mockGit.AssertExpectations(t)
			mockManager.AssertExpectations(t)
		})
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil, // splitLines returns nil for empty string
		},
		{
			name:     "single line",
			input:    "main",
			expected: []string{"main"},
		},
		{
			name:     "multiple lines",
			input:    "main\nfeature-1\nfeature-2",
			expected: []string{"main", "feature-1", "feature-2"},
		},
		{
			name:     "lines with empty entries",
			input:    "main\n\nfeature\n\n",
			expected: []string{"main", "feature"},
		},
		{
			name:     "lines with whitespace",
			input:    "  main  \n\t feature \t\n  ",
			expected: []string{"main", "feature"},
		},
		{
			name:     "windows line endings",
			input:    "main\r\nfeature\r\n",
			expected: []string{"main", "feature"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLines(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCleanupStats(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mocks.MockGitClient, *mocks.MockTicketManager)
		expectedError bool
	}{
		{
			name: "display cleanup stats",
			setupMocks: func(g *mocks.MockGitClient, m *mocks.MockTicketManager) {
				// Mock done tickets
				doneTime, _ := time.Parse(time.RFC3339, "2025-01-01T14:00:00Z")
				doneTicket := ticket.Ticket{ID: "250101-120000-done-ticket"}
				doneTicket.ClosedAt = ticket.RFC3339TimePtr{Time: &doneTime}
				m.On("List", mock.Anything, ticket.StatusFilterDone).Return([]ticket.Ticket{doneTicket}, nil)

				// Mock worktree stats
				g.On("ListWorktrees", mock.Anything).Return([]git.WorktreeInfo{
					{Branch: "250101-120000-orphaned", Path: "/worktrees/orphaned"},
					{Branch: "main", Path: "/main"},
				}, nil)
				m.On("List", mock.Anything, ticket.StatusFilterDoing).Return([]ticket.Ticket{}, nil)

				// Mock branch stats
				g.On("Exec", mock.Anything, "branch", "--format=%(refname:short)").Return("main\n250101-120000-done-ticket\n250102-120000-orphaned", nil)
				m.On("List", mock.Anything, ticket.StatusFilterAll).Return([]ticket.Ticket{doneTicket}, nil)
			},
			expectedError: false,
		},
		{
			name: "partial error getting stats",
			setupMocks: func(g *mocks.MockGitClient, m *mocks.MockTicketManager) {
				// Mock done tickets succeeds
				m.On("List", mock.Anything, ticket.StatusFilterDone).Return([]ticket.Ticket{}, nil)
				
				// Mock worktree stats fails, but it still tries to get doing tickets
				g.On("ListWorktrees", mock.Anything).Return(nil, fmt.Errorf("git error"))
				m.On("List", mock.Anything, ticket.StatusFilterDoing).Return([]ticket.Ticket{}, nil)
				
				// Branch stats continues despite error
				g.On("Exec", mock.Anything, "branch", "--format=%(refname:short)").Return("main", nil)
				m.On("List", mock.Anything, ticket.StatusFilterAll).Return([]ticket.Ticket{}, nil)
			},
			expectedError: false, // CleanupStats doesn't return errors, just continues
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			
			mockGit := new(mocks.MockGitClient)
			mockManager := new(mocks.MockTicketManager)

			cfg := config.Default()
			cfg.Worktree.Enabled = true

			app := &App{
				Config:  cfg,
				Git:     mockGit,
				Manager: mockManager,
			}

			tt.setupMocks(mockGit, mockManager)

			err := app.CleanupStats(ctx)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockGit.AssertExpectations(t)
			mockManager.AssertExpectations(t)
		})
	}
}

// Test error scenarios
func TestAutoCleanup_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	
	mockGit := new(mocks.MockGitClient)
	mockManager := new(mocks.MockTicketManager)

	cfg := config.Default()
	cfg.Worktree.Enabled = true

	app := &App{
		Config:  cfg,
		Git:     mockGit,
		Manager: mockManager,
	}

	// Setup failing mocks
	mockGit.On("PruneWorktrees", mock.Anything).Return(nil) // This still gets called
	mockGit.On("ListWorktrees", mock.Anything).Return(nil, fmt.Errorf("permission denied"))
	mockGit.On("Exec", mock.Anything, "branch", "--format=%(refname:short)").Return("", fmt.Errorf("git not found"))

	result, err := app.AutoCleanup(ctx, false)

	// AutoCleanup should not return an error, but collect errors in result
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.HasErrors())
	assert.GreaterOrEqual(t, len(result.Errors), 2)

	mockGit.AssertExpectations(t)
}