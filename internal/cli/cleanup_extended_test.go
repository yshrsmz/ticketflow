package cli

import (
	"context"
	"fmt"
	"testing"

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

				// Mock ticket lookups
				m.On("List", mock.Anything, ticket.StatusFilterDone).Return([]*ticket.Ticket{
					{ID: "250101-120000-old-feature"},
				}, nil)
				m.On("List", mock.Anything, ticket.StatusFilterDoing).Return([]*ticket.Ticket{}, nil)
				m.On("FindTicket", mock.Anything, "250102-120000-active-feature").Return(nil, fmt.Errorf("not found"))

				// Mock cleanup operations
				g.On("RemoveWorktree", mock.Anything, "/worktrees/active").Return(nil)
				g.On("Exec", mock.Anything, "branch", "-D", "250102-120000-active-feature").Return("", nil)
				g.On("PruneWorktrees", mock.Anything).Return(nil)

				// Mock stale branch cleanup
				g.On("Exec", mock.Anything, "branch", "--format=%(refname:short)").Return("main\n250101-120000-old-feature\n250103-120000-stale", nil)
				m.On("FindTicket", mock.Anything, "250103-120000-stale").Return(nil, fmt.Errorf("not found"))
				g.On("Exec", mock.Anything, "branch", "-D", "250103-120000-stale").Return("", nil)
			},
			expectedResult: &CleanupResult{
				OrphanedWorktrees: 1,
				StaleBranches:     1,
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
				m.On("List", mock.Anything, ticket.StatusFilterDone).Return([]*ticket.Ticket{}, nil)
				m.On("List", mock.Anything, ticket.StatusFilterDoing).Return([]*ticket.Ticket{}, nil)
				m.On("FindTicket", mock.Anything, "250101-120000-orphaned").Return(nil, fmt.Errorf("not found"))

				g.On("Exec", mock.Anything, "branch", "--format=%(refname:short)").Return("main\n250101-120000-orphaned", nil)
				g.On("PruneWorktrees", mock.Anything).Return(nil)

				// In dry run, no actual deletion should happen
			},
			expectedResult: &CleanupResult{
				OrphanedWorktrees: 1,
				StaleBranches:     1,
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
				g.On("Exec", mock.Anything, "branch", "--format=%(refname:short)").Return("main\n250101-120000-stale", nil)
				m.On("FindTicket", mock.Anything, "250101-120000-stale").Return(nil, fmt.Errorf("not found"))
				g.On("Exec", mock.Anything, "branch", "-D", "250101-120000-stale").Return("", nil)
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
		ticketLookups  map[string]error // ticketID -> error (nil means found)
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
			ticketLookups: map[string]error{
				"250101-120000-orphaned": fmt.Errorf("not found"),
			},
			expectedCount: 1,
			expectedError: false,
		},
		{
			name:   "skip main branch",
			dryRun: false,
			worktrees: []git.WorktreeInfo{
				{Branch: "main", Path: "/main"},
				{Branch: "master", Path: "/master"},
			},
			ticketLookups:  map[string]error{},
			expectedCount:  0,
			expectedError:  false,
		},
		{
			name:   "skip active ticket worktree",
			dryRun: false,
			worktrees: []git.WorktreeInfo{
				{Branch: "250101-120000-active", Path: "/worktrees/active"},
			},
			ticketLookups: map[string]error{
				"250101-120000-active": nil, // Found
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
			ticketLookups: map[string]error{
				"250101-120000-old1":   fmt.Errorf("not found"),
				"250102-120000-old2":   fmt.Errorf("not found"),
				"250103-120000-active": nil, // Found
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

			for ticketID, err := range tt.ticketLookups {
				if err != nil {
					mockManager.On("FindTicket", mock.Anything, ticketID).Return(nil, err)
				} else {
					mockManager.On("FindTicket", mock.Anything, ticketID).Return(&ticket.Ticket{ID: ticketID}, nil)
				}
			}

			// Setup removal mocks for orphaned worktrees
			if !tt.dryRun {
				for _, wt := range tt.worktrees {
					if err, found := tt.ticketLookups[wt.Branch]; found && err != nil {
						mockGit.On("RemoveWorktree", mock.Anything, wt.Path).Return(nil)
						mockGit.On("Exec", mock.Anything, "branch", "-D", wt.Branch).Return("", nil)
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
			expected: []string{},
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
				// Mock worktree stats
				g.On("ListWorktrees", mock.Anything).Return([]git.WorktreeInfo{
					{Branch: "250101-120000-orphaned", Path: "/worktrees/orphaned"},
					{Branch: "main", Path: "/main"},
				}, nil)
				m.On("FindTicket", mock.Anything, "250101-120000-orphaned").Return(nil, fmt.Errorf("not found"))

				// Mock branch stats
				g.On("Exec", mock.Anything, "branch", "--format=%(refname:short)").Return("main\n250101-120000-orphaned\n250102-120000-stale", nil)
				m.On("FindTicket", mock.Anything, "250102-120000-stale").Return(nil, fmt.Errorf("not found"))
			},
			expectedError: false,
		},
		{
			name: "error getting stats",
			setupMocks: func(g *mocks.MockGitClient, m *mocks.MockTicketManager) {
				g.On("ListWorktrees", mock.Anything).Return(nil, fmt.Errorf("git error"))
			},
			expectedError: true,
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