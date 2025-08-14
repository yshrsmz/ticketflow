package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/config"
	ticketerrors "github.com/yshrsmz/ticketflow/internal/errors"
	"github.com/yshrsmz/ticketflow/internal/mocks"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestApp_NewTicket(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		slug          string
		setupManager  func(m *mocks.MockTicketManager)
		outputFormat  OutputFormat
		expectedError bool
		errorMessage  string
	}{
		{
			name: "successful ticket creation",
			slug: "test-feature",
			setupManager: func(m *mocks.MockTicketManager) {
				m.On("Create", mock.Anything, "test-feature").Return(&ticket.Ticket{
					ID:          "250131-120000-test-feature",
					Path:        "/path/to/ticket.md",
					Priority:    1,
					Description: "",
					CreatedAt:   ticket.RFC3339Time{Time: time.Now()},
					Content:     "",
				}, nil)
			},
			outputFormat:  FormatText,
			expectedError: false,
		},
		{
			name: "invalid slug",
			slug: "invalid slug with spaces",
			setupManager: func(m *mocks.MockTicketManager) {
				// Manager won't be called if slug is invalid
			},
			outputFormat:  FormatText,
			expectedError: true,
			errorMessage:  "Invalid slug format",
		},
		{
			name: "ticket already exists",
			slug: "existing-ticket",
			setupManager: func(m *mocks.MockTicketManager) {
				m.On("Create", mock.Anything, "existing-ticket").Return(nil, ticketerrors.ErrTicketExists)
			},
			outputFormat:  FormatText,
			expectedError: true,
			errorMessage:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock manager
			mockManager := new(mocks.MockTicketManager)
			if tt.setupManager != nil {
				tt.setupManager(mockManager)
			}

			// Create mock git client
			mockGit := new(mocks.MockGitClient)
			// Only expect CurrentBranch call for valid slugs
			if tt.slug == "" || ticket.IsValidSlug(tt.slug) {
				mockGit.On("CurrentBranch", mock.Anything).Return("main", nil)
			}

			// Create app with mock
			app := &App{
				Config:      config.Default(),
				Manager:     mockManager,
				Git:         mockGit,
				ProjectRoot: "/test/project",
				Output:      NewOutputWriter(nil, nil, FormatText),
			}

			// Execute
			_, err := app.NewTicket(context.Background(), tt.slug, "")

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify expectations
			mockManager.AssertExpectations(t)
			mockGit.AssertExpectations(t)
		})
	}
}

func TestApp_ListTickets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		status        ticket.Status
		count         int
		outputFormat  OutputFormat
		setupManager  func(m *mocks.MockTicketManager)
		expectedError bool
		errorMessage  string
	}{
		{
			name:         "list active tickets (default)",
			status:       "",
			count:        10,
			outputFormat: FormatText,
			setupManager: func(m *mocks.MockTicketManager) {
				tickets := []ticket.Ticket{
					{ID: "ticket1", Path: "", Priority: 1, Description: ""},
					{ID: "ticket2", Path: "", Priority: 2, Description: ""},
				}
				m.On("List", mock.Anything, ticket.StatusFilterActive).Return(tickets, nil)
			},
			expectedError: false,
		},
		{
			name:         "list todo tickets",
			status:       ticket.StatusTodo,
			count:        5,
			outputFormat: FormatJSON,
			setupManager: func(m *mocks.MockTicketManager) {
				tickets := []ticket.Ticket{
					{
						ID:          "todo1",
						Path:        "",
						Priority:    1,
						Description: "",
						CreatedAt:   ticket.RFC3339Time{Time: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
					},
				}
				m.On("List", mock.Anything, ticket.StatusFilterTodo).Return(tickets, nil)
				// JSON format also calls List with "all" to get summary
				m.On("List", mock.Anything, ticket.StatusFilterAll).Return(tickets, nil)
			},
			expectedError: false,
		},
		{
			name:         "list error",
			status:       "",
			count:        10,
			outputFormat: FormatText,
			setupManager: func(m *mocks.MockTicketManager) {
				m.On("List", mock.Anything, ticket.StatusFilterActive).Return(nil, assert.AnError)
			},
			expectedError: true,
		},
	}

	tmpDir := t.TempDir()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock manager
			mockManager := new(mocks.MockTicketManager)
			if tt.setupManager != nil {
				tt.setupManager(mockManager)
			}

			// Create mock git client
			mockGit := new(mocks.MockGitClient)

			// Create app with mock
			app := &App{
				Config:      config.Default(),
				Manager:     mockManager,
				Git:         mockGit,
				ProjectRoot: tmpDir,
				Output:      NewOutputWriter(nil, nil, FormatText),
			}

			// Execute
			err := app.ListTickets(context.Background(), tt.status, tt.count, tt.outputFormat)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify expectations
			mockManager.AssertExpectations(t)
			mockGit.AssertExpectations(t)
		})
	}
}

func TestApp_StartTicket_WithMocks(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	tests := []struct {
		name          string
		ticketID      string
		setupMocks    func(tm *mocks.MockTicketManager, gc *mocks.MockGitClient, tmpDir string)
		expectedError bool
		errorMessage  string
	}{
		{
			name:     "successful start",
			ticketID: "250131-120000-test",
			setupMocks: func(tm *mocks.MockTicketManager, gc *mocks.MockGitClient, tmpDir string) {
				// Create the required directory structure
				todoDir := filepath.Join(tmpDir, "tickets", "todo")
				doingDir := filepath.Join(tmpDir, "tickets", "doing")
				if err := os.MkdirAll(todoDir, 0755); err != nil {
					t.Fatalf("Failed to create todo dir: %v", err)
				}
				if err := os.MkdirAll(doingDir, 0755); err != nil {
					t.Fatalf("Failed to create doing dir: %v", err)
				}

				// Create the ticket file
				ticketPath := filepath.Join(todoDir, "250131-120000-test.md")
				content := `---
priority: 1
description: ""
created_at: "2021-01-31T12:00:00Z"
---

Test ticket content`
				if err := os.WriteFile(ticketPath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write ticket file: %v", err)
				}

				testTicket := &ticket.Ticket{
					ID:          "250131-120000-test",
					Path:        ticketPath,
					Priority:    1,
					Description: "",
					CreatedAt:   ticket.RFC3339Time{Time: time.Now()},
				}
				tm.On("Get", mock.Anything, "250131-120000-test").Return(testTicket, nil)
				tm.On("UpdateStatus", mock.Anything, testTicket, ticket.StatusDoing).Return(nil).Maybe()
				tm.On("Update", mock.Anything, mock.AnythingOfType("*ticket.Ticket")).Return(nil)
				tm.On("SetCurrentTicket", mock.Anything, mock.AnythingOfType("*ticket.Ticket")).Return(nil)
				gc.On("CurrentBranch", mock.Anything).Return("main", nil)
				gc.On("HasUncommittedChanges", mock.Anything).Return(false, nil)
				gc.On("BranchExists", mock.Anything, "250131-120000-test").Return(false, nil).Maybe()
				gc.On("CreateBranch", mock.Anything, "250131-120000-test").Return(nil)
				gc.On("Add", mock.Anything, "-A", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
				gc.On("Commit", mock.Anything, "Start ticket: 250131-120000-test").Return(nil)
				gc.On("Checkout", mock.Anything, "250131-120000-test").Return(nil).Maybe()
			},
			expectedError: false,
		},
		{
			name:     "ticket not found",
			ticketID: "nonexistent",
			setupMocks: func(tm *mocks.MockTicketManager, gc *mocks.MockGitClient, tmpDir string) {
				tm.On("Get", mock.Anything, "nonexistent").Return(nil, assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockManager := new(mocks.MockTicketManager)
			mockGit := new(mocks.MockGitClient)

			if tt.setupMocks != nil {
				tt.setupMocks(mockManager, mockGit, tmpDir)
			}

			// Create app with mocks
			cfg := config.Default()
			cfg.Worktree.Enabled = false // Disable worktrees for this unit test
			cfg.Git.DefaultBranch = "main"
			cfg.Tickets.Dir = "tickets"

			app := &App{
				Config:      cfg,
				Manager:     mockManager,
				Git:         mockGit,
				ProjectRoot: tmpDir,
				Output:      NewOutputWriter(nil, nil, FormatText),
			}

			// Execute
			_, err := app.StartTicket(context.Background(), tt.ticketID, false)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			mockManager.AssertExpectations(t)
			mockGit.AssertExpectations(t)
		})
	}
}

func TestApp_StartTicket_WorktreeMode_NoMainRepoSymlink(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create the required directory structure
	todoDir := filepath.Join(tmpDir, "tickets", "todo")
	doingDir := filepath.Join(tmpDir, "tickets", "doing")
	require.NoError(t, os.MkdirAll(todoDir, 0755))
	require.NoError(t, os.MkdirAll(doingDir, 0755))

	// Create the ticket file
	ticketID := "250131-120000-test"
	ticketPath := filepath.Join(todoDir, ticketID+".md")
	content := `---
priority: 1
description: "Test ticket"
created_at: "2021-01-31T12:00:00Z"
---

Test ticket content`
	require.NoError(t, os.WriteFile(ticketPath, []byte(content), 0644))

	// Create the worktree directory that AddWorktree would create
	worktreeBaseDir := filepath.Join(tmpDir, "../test.worktrees")
	worktreePath := filepath.Join(worktreeBaseDir, ticketID)
	require.NoError(t, os.MkdirAll(worktreePath, 0755))

	// Create mocks
	mockManager := new(mocks.MockTicketManager)
	mockGit := new(mocks.MockGitClient)

	testTicket := &ticket.Ticket{
		ID:          ticketID,
		Path:        ticketPath,
		Priority:    1,
		Description: "Test ticket",
		CreatedAt:   ticket.RFC3339Time{Time: time.Now()},
	}

	// Setup expectations
	mockManager.On("Get", mock.Anything, ticketID).Return(testTicket, nil)
	mockManager.On("Update", mock.Anything, mock.AnythingOfType("*ticket.Ticket")).Return(nil)
	// SetCurrentTicket should NOT be called when worktrees are enabled
	// We don't set any expectation for it, so the test will fail if it's called

	mockGit.On("CurrentBranch", mock.Anything).Return("main", nil)
	mockGit.On("HasUncommittedChanges", mock.Anything).Return(false, nil).Maybe()
	mockGit.On("BranchExists", mock.Anything, ticketID).Return(false, nil).Maybe()
	mockGit.On("CreateBranch", mock.Anything, ticketID).Return(nil).Maybe()
	mockGit.On("Checkout", mock.Anything, ticketID).Return(nil).Maybe()
	mockGit.On("HasWorktree", mock.Anything, ticketID).Return(false, nil) // Check for existing worktree
	mockGit.On("Add", mock.Anything, "-A", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	mockGit.On("Commit", mock.Anything, "Start ticket: "+ticketID).Return(nil)
	mockGit.On("Checkout", mock.Anything, "main").Return(nil).Maybe() // Switch back in worktree mode
	mockGit.On("AddWorktree", mock.Anything, worktreePath, ticketID).Return(nil)

	// Create app with worktrees ENABLED
	cfg := config.Default()
	cfg.Worktree.Enabled = true
	cfg.Worktree.BaseDir = "../test.worktrees"
	cfg.Git.DefaultBranch = "main"
	cfg.Tickets.Dir = "tickets"
	cfg.Worktree.InitCommands = []string{} // Disable init commands for test

	app := &App{
		Config:      cfg,
		Manager:     mockManager,
		Git:         mockGit,
		ProjectRoot: tmpDir,
		Output:      NewOutputWriter(nil, nil, FormatText),
	}

	// Execute
	_, err := app.StartTicket(context.Background(), ticketID, false)
	assert.NoError(t, err)

	// Verify that SetCurrentTicket was NOT called (by checking expectations)
	mockManager.AssertExpectations(t)
	mockGit.AssertExpectations(t)

	// Also verify that current-ticket.md does NOT exist in the main repo
	mainRepoSymlink := filepath.Join(tmpDir, "current-ticket.md")
	_, err = os.Lstat(mainRepoSymlink)
	assert.True(t, os.IsNotExist(err), "current-ticket.md should not exist in main repo when worktrees are enabled")
}

func TestApp_StartTicket_NonWorktreeMode_CreatesMainRepoSymlink(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create the required directory structure
	todoDir := filepath.Join(tmpDir, "tickets", "todo")
	doingDir := filepath.Join(tmpDir, "tickets", "doing")
	require.NoError(t, os.MkdirAll(todoDir, 0755))
	require.NoError(t, os.MkdirAll(doingDir, 0755))

	// Create the ticket file
	ticketID := "250131-120000-test"
	ticketPath := filepath.Join(todoDir, ticketID+".md")
	content := `---
priority: 1
description: "Test ticket"
created_at: "2021-01-31T12:00:00Z"
---

Test ticket content`
	require.NoError(t, os.WriteFile(ticketPath, []byte(content), 0644))

	// Create mocks
	mockManager := new(mocks.MockTicketManager)
	mockGit := new(mocks.MockGitClient)

	testTicket := &ticket.Ticket{
		ID:          ticketID,
		Path:        ticketPath,
		Priority:    1,
		Description: "Test ticket",
		CreatedAt:   ticket.RFC3339Time{Time: time.Now()},
	}

	// Setup expectations
	mockManager.On("Get", mock.Anything, ticketID).Return(testTicket, nil)
	mockManager.On("Update", mock.Anything, mock.AnythingOfType("*ticket.Ticket")).Return(nil)
	// SetCurrentTicket SHOULD be called when worktrees are disabled
	mockManager.On("SetCurrentTicket", mock.Anything, mock.AnythingOfType("*ticket.Ticket")).Return(nil)

	mockGit.On("CurrentBranch", mock.Anything).Return("main", nil)
	mockGit.On("HasUncommittedChanges", mock.Anything).Return(false, nil)
	mockGit.On("BranchExists", mock.Anything, ticketID).Return(false, nil).Maybe()
	mockGit.On("CreateBranch", mock.Anything, ticketID).Return(nil).Maybe()
	mockGit.On("Checkout", mock.Anything, ticketID).Return(nil).Maybe()
	mockGit.On("Add", mock.Anything, "-A", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	mockGit.On("Commit", mock.Anything, "Start ticket: "+ticketID).Return(nil)
	// No checkout back to main since worktrees are disabled

	// Create app with worktrees DISABLED
	cfg := config.Default()
	cfg.Worktree.Enabled = false
	cfg.Git.DefaultBranch = "main"
	cfg.Tickets.Dir = "tickets"

	app := &App{
		Config:      cfg,
		Manager:     mockManager,
		Git:         mockGit,
		ProjectRoot: tmpDir,
		Output:      NewOutputWriter(nil, nil, FormatText),
	}

	// Execute
	_, err := app.StartTicket(context.Background(), ticketID, false)
	assert.NoError(t, err)

	// Verify that SetCurrentTicket WAS called
	mockManager.AssertExpectations(t)
	mockGit.AssertExpectations(t)
}

func TestNewApp_DefaultWorkingDirectory(t *testing.T) {
	// Cannot use t.Parallel() - this test specifically validates behavior
	// when no working directory is specified, so it must use os.Chdir
	// Create a temporary directory and make it the current directory
	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalWd)
		if err != nil {
			t.Logf("Failed to restore working directory: %v", err)
		}
	}()

	// Change to temp directory
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	err = cmd.Run()
	require.NoError(t, err)

	// Configure git locally (not globally)
	ConfigureTestGit(t, tmpDir)

	// Initialize ticketflow
	err = InitCommand(context.Background())
	require.NoError(t, err)

	// Create app without specifying working directory - should use current directory
	app, err := NewApp(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, app)

	// Resolve symlinks for comparison (macOS /var -> /private/var)
	expectedPath, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)
	actualPath, err := filepath.EvalSymlinks(app.ProjectRoot)
	require.NoError(t, err)
	assert.Equal(t, expectedPath, actualPath)
}

// TestValidateTicketForClose_SymlinkError tests that symlink errors are properly detected
// using error type checking instead of string matching
func TestValidateTicketForClose_SymlinkError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		setupManager    func(m *mocks.MockTicketManager)
		expectedError   bool
		expectedCode    string
		checkSuggestion string
	}{
		{
			name: "symlink readlink error",
			setupManager: func(m *mocks.MockTicketManager) {
				// Simulate a readlink error from GetCurrentTicket
				// This happens when the symlink exists but cannot be read
				pathErr := &os.PathError{
					Op:   "readlink",
					Path: "/test/current-ticket.md",
					Err:  os.ErrPermission,
				}
				wrappedErr := fmt.Errorf("failed to read current ticket link: %w", pathErr)
				m.On("GetCurrentTicket", mock.Anything).Return(nil, wrappedErr)
			},
			expectedError:   true,
			expectedCode:    ErrTicketNotStarted,
			checkSuggestion: "Try restoring the current ticket link: ticketflow restore",
		},
		{
			name: "other file system error",
			setupManager: func(m *mocks.MockTicketManager) {
				// Simulate a different file system error (not readlink)
				pathErr := &os.PathError{
					Op:   "open",
					Path: "/test/some-file.md",
					Err:  os.ErrNotExist,
				}
				wrappedErr := fmt.Errorf("failed to open file: %w", pathErr)
				m.On("GetCurrentTicket", mock.Anything).Return(nil, wrappedErr)
			},
			expectedError: true,
			expectedCode:  "", // Should use ConvertError for non-readlink errors
		},
		{
			name: "successful get current ticket",
			setupManager: func(m *mocks.MockTicketManager) {
				now := time.Now()
				testTicket := &ticket.Ticket{
					ID:          "250131-120000-test",
					Priority:    1,
					Description: "Test ticket",
					StartedAt:   ticket.RFC3339TimePtr{Time: &now},
					Path:        "/test/tickets/doing/250131-120000-test.md",
				}
				m.On("GetCurrentTicket", mock.Anything).Return(testTicket, nil)
			},
			expectedError: false,
		},
		{
			name: "no current ticket",
			setupManager: func(m *mocks.MockTicketManager) {
				m.On("GetCurrentTicket", mock.Anything).Return(nil, nil)
			},
			expectedError:   true,
			expectedCode:    ErrTicketNotStarted,
			checkSuggestion: "Start a ticket first: ticketflow start <ticket-id>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock manager
			mockManager := new(mocks.MockTicketManager)
			if tt.setupManager != nil {
				tt.setupManager(mockManager)
			}

			// Create mock git client
			mockGit := new(mocks.MockGitClient)
			// Setup git mock only for successful case where we have a current ticket
			if tt.name == "successful get current ticket" {
				mockGit.On("FindWorktreeByBranch", mock.Anything, "250131-120000-test").Return(nil, nil).Maybe()
				mockGit.On("HasUncommittedChanges", mock.Anything).Return(false, nil).Maybe()
				mockGit.On("CurrentBranch", mock.Anything).Return("250131-120000-test", nil).Maybe()
			}

			// Create app with mocks
			app := &App{
				Config:      config.Default(),
				Manager:     mockManager,
				Git:         mockGit,
				ProjectRoot: "/test/project",
				Output:      NewOutputWriter(nil, nil, FormatText),
			}

			// Execute
			ticket, ticketPath, err := app.validateTicketForClose(context.Background(), false)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)

				// Check if it's a CLI error
				if cliErr, ok := err.(*CLIError); ok {
					if tt.expectedCode != "" {
						assert.Equal(t, tt.expectedCode, cliErr.Code, "Error code mismatch")
					}
					if tt.checkSuggestion != "" {
						found := false
						for _, suggestion := range cliErr.Suggestions {
							if suggestion == tt.checkSuggestion {
								found = true
								break
							}
						}
						assert.True(t, found, "Expected suggestion not found: %s", tt.checkSuggestion)
					}
				}

				assert.Nil(t, ticket)
				assert.Empty(t, ticketPath)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ticket)
				// ticketPath is empty when there's no worktree, which is expected
				assert.Empty(t, ticketPath)
			}

			mockManager.AssertExpectations(t)
			mockGit.AssertExpectations(t)
		})
	}
}

func TestApp_CloseTicketByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		ticketID      string
		reason        string
		setupMocks    func(*mocks.MockTicketManager, *mocks.MockGitClient, string)
		expectedError bool
		errorContains string
	}{
		{
			name:     "close ticket with reason when branch not merged",
			ticketID: "250131-120000-test-ticket",
			reason:   "Abandoned due to priority change",
			setupMocks: func(tm *mocks.MockTicketManager, gc *mocks.MockGitClient, tmpDir string) {
				// Setup ticket
				testTicket := &ticket.Ticket{
					ID:          "250131-120000-test-ticket",
					Path:        filepath.Join(tmpDir, "tickets/todo/250131-120000-test-ticket.md"),
					Priority:    2,
					Description: "Test ticket",
					CreatedAt:   ticket.RFC3339Time{Time: time.Now()},
					Content:     "# Test Ticket\n\nContent here.",
				}

				// Mock getting ticket
				tm.On("Get", mock.Anything, "250131-120000-test-ticket").Return(testTicket, nil)

				// Mock GetCurrentTicket (returns nil for not current)
				tm.On("GetCurrentTicket", mock.Anything).Return(nil, nil)

				// Mock SetCurrentTicket (remove current if it's the current ticket)
				tm.On("SetCurrentTicket", mock.Anything, (*ticket.Ticket)(nil)).Return(nil).Maybe()

				// Mock branch merge check (not merged)
				gc.On("IsBranchMerged", mock.Anything, "250131-120000-test-ticket", "main").Return(false, nil)

				// Mock updating ticket with reason (only once in moveTicketToDoneWithReason)
				tm.On("Update", mock.Anything, testTicket).Return(nil).Times(1)

				// Mock git operations
				gc.On("Add", mock.Anything, "-A", mock.Anything, mock.Anything).Return(nil)
				gc.On("Commit", mock.Anything, mock.MatchedBy(func(msg string) bool {
					return strings.Contains(msg, "Close ticket:")
				})).Return(nil)
			},
			expectedError: false,
		},
		{
			name:     "close ticket when branch already merged",
			ticketID: "250131-120000-merged-ticket",
			reason:   "",
			setupMocks: func(tm *mocks.MockTicketManager, gc *mocks.MockGitClient, tmpDir string) {
				// Setup ticket
				testTicket := &ticket.Ticket{
					ID:          "250131-120000-merged-ticket",
					Path:        filepath.Join(tmpDir, "tickets/doing/250131-120000-merged-ticket.md"),
					Priority:    2,
					Description: "Merged ticket",
					CreatedAt:   ticket.RFC3339Time{Time: time.Now()},
					StartedAt:   ticket.NewRFC3339TimePtr(&time.Time{}),
					Content:     "# Merged Ticket\n\nContent here.",
				}

				// Mock getting ticket
				tm.On("Get", mock.Anything, "250131-120000-merged-ticket").Return(testTicket, nil)

				// Mock GetCurrentTicket (returns nil for not current)
				tm.On("GetCurrentTicket", mock.Anything).Return(nil, nil)

				// Mock SetCurrentTicket (remove current if it's the current ticket)
				tm.On("SetCurrentTicket", mock.Anything, (*ticket.Ticket)(nil)).Return(nil).Maybe()

				// Mock branch merge check (merged)
				gc.On("IsBranchMerged", mock.Anything, "250131-120000-merged-ticket", "main").Return(true, nil)

				// Mock updating ticket (only once in moveTicketToDoneWithReason)
				tm.On("Update", mock.Anything, testTicket).Return(nil).Times(1)

				// Mock git operations
				gc.On("Add", mock.Anything, "-A", mock.Anything, mock.Anything).Return(nil)
				gc.On("Commit", mock.Anything, mock.MatchedBy(func(msg string) bool {
					return strings.Contains(msg, "Close ticket:")
				})).Return(nil)
			},
			expectedError: false,
		},
		{
			name:     "error when ticket not found",
			ticketID: "nonexistent-ticket",
			reason:   "Some reason",
			setupMocks: func(tm *mocks.MockTicketManager, gc *mocks.MockGitClient, tmpDir string) {
				// CloseTicketByID checks if this is the current ticket first
				tm.On("GetCurrentTicket", mock.Anything).Return(nil, nil)
				tm.On("Get", mock.Anything, "nonexistent-ticket").Return(nil, ticketerrors.ErrTicketNotFound)
			},
			expectedError: true,
			errorContains: "not found",
		},
		{
			name:     "error when ticket already closed",
			ticketID: "250131-120000-closed-ticket",
			reason:   "Some reason",
			setupMocks: func(tm *mocks.MockTicketManager, gc *mocks.MockGitClient, tmpDir string) {
				closedTime := time.Now()
				testTicket := &ticket.Ticket{
					ID:          "250131-120000-closed-ticket",
					Path:        filepath.Join(tmpDir, "tickets/done/250131-120000-closed-ticket.md"),
					Priority:    2,
					Description: "Closed ticket",
					CreatedAt:   ticket.RFC3339Time{Time: time.Now()},
					ClosedAt:    ticket.NewRFC3339TimePtr(&closedTime),
					Content:     "# Closed Ticket\n\nContent here.",
				}

				// CloseTicketByID checks if this is the current ticket first
				tm.On("GetCurrentTicket", mock.Anything).Return(nil, nil)
				tm.On("Get", mock.Anything, "250131-120000-closed-ticket").Return(testTicket, nil)
			},
			expectedError: true,
			errorContains: "already closed",
		},
		{
			name:     "error when reason missing for unmerged branch",
			ticketID: "250131-120000-test-ticket",
			reason:   "",
			setupMocks: func(tm *mocks.MockTicketManager, gc *mocks.MockGitClient, tmpDir string) {
				testTicket := &ticket.Ticket{
					ID:          "250131-120000-test-ticket",
					Path:        filepath.Join(tmpDir, "tickets/todo/250131-120000-test-ticket.md"),
					Priority:    2,
					Description: "Test ticket",
					CreatedAt:   ticket.RFC3339Time{Time: time.Now()},
					Content:     "# Test Ticket\n\nContent here.",
				}

				// CloseTicketByID checks if this is the current ticket first
				tm.On("GetCurrentTicket", mock.Anything).Return(nil, nil)
				tm.On("Get", mock.Anything, "250131-120000-test-ticket").Return(testTicket, nil)

				// Mock branch merge check (not merged)
				gc.On("IsBranchMerged", mock.Anything, "250131-120000-test-ticket", "main").Return(false, nil)
			},
			expectedError: true,
			errorContains: "Reason required",
		},
		{
			name:     "preserve current-ticket.md when closing non-current ticket",
			ticketID: "250131-120000-other-ticket",
			reason:   "Abandoned",
			setupMocks: func(tm *mocks.MockTicketManager, gc *mocks.MockGitClient, tmpDir string) {
				// Setup the ticket being closed
				otherTicket := &ticket.Ticket{
					ID:          "250131-120000-other-ticket",
					Path:        filepath.Join(tmpDir, "tickets/todo/250131-120000-other-ticket.md"),
					Priority:    2,
					Description: "Other ticket",
					CreatedAt:   ticket.RFC3339Time{Time: time.Now()},
					Content:     "# Other Ticket\n\nContent here.",
				}

				// Setup the current ticket (different from the one being closed)
				currentTicket := &ticket.Ticket{
					ID:          "250131-120000-current-ticket",
					Path:        filepath.Join(tmpDir, "tickets/doing/250131-120000-current-ticket.md"),
					Priority:    1,
					Description: "Current ticket",
					CreatedAt:   ticket.RFC3339Time{Time: time.Now()},
					StartedAt:   ticket.NewRFC3339TimePtr(&time.Time{}),
					Content:     "# Current Ticket\n\nContent here.",
				}

				// Mock getting the ticket to close
				tm.On("Get", mock.Anything, "250131-120000-other-ticket").Return(otherTicket, nil)

				// Mock GetCurrentTicket (returns the current ticket which is different)
				tm.On("GetCurrentTicket", mock.Anything).Return(currentTicket, nil)

				// IMPORTANT: SetCurrentTicket should NOT be called since we're not closing the current ticket
				// The fix ensures this method is not called when closing a non-current ticket

				// Mock branch merge check (not merged, so reason is required)
				gc.On("IsBranchMerged", mock.Anything, "250131-120000-other-ticket", "main").Return(false, nil)

				// Mock updating ticket with reason
				tm.On("Update", mock.Anything, otherTicket).Return(nil).Times(1)

				// Mock git operations
				gc.On("Add", mock.Anything, "-A", mock.Anything, mock.Anything).Return(nil)
				gc.On("Commit", mock.Anything, mock.MatchedBy(func(msg string) bool {
					return strings.Contains(msg, "Close ticket: 250131-120000-other-ticket")
				})).Return(nil)
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for tests
			tmpDir := t.TempDir()

			// Create required directories
			require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "tickets", "todo"), 0755))
			require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "tickets", "doing"), 0755))
			require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "tickets", "done"), 0755))

			// Setup mocks
			mockManager := &mocks.MockTicketManager{}
			mockGit := &mocks.MockGitClient{}

			// Setup test mocks
			tt.setupMocks(mockManager, mockGit, tmpDir)

			// Create the ticket files for tests that need them
			if !tt.expectedError || tt.errorContains == "already closed" {
				// Create the actual ticket file for non-error cases and already closed case
				switch tt.ticketID {
				case "250131-120000-test-ticket":
					if tt.reason == "" {
						// For error case where reason is missing, still create the file
						ticketFile := filepath.Join(tmpDir, "tickets", "todo", "250131-120000-test-ticket.md")
						require.NoError(t, os.WriteFile(ticketFile, []byte("test content"), 0644))
					} else {
						ticketFile := filepath.Join(tmpDir, "tickets", "todo", "250131-120000-test-ticket.md")
						require.NoError(t, os.WriteFile(ticketFile, []byte("test content"), 0644))
					}
				case "250131-120000-merged-ticket":
					ticketFile := filepath.Join(tmpDir, "tickets", "doing", "250131-120000-merged-ticket.md")
					require.NoError(t, os.WriteFile(ticketFile, []byte("test content"), 0644))
				case "250131-120000-closed-ticket":
					ticketFile := filepath.Join(tmpDir, "tickets", "done", "250131-120000-closed-ticket.md")
					require.NoError(t, os.WriteFile(ticketFile, []byte("test content"), 0644))
				case "250131-120000-other-ticket":
					// For the preservation test - create the ticket being closed
					ticketFile := filepath.Join(tmpDir, "tickets", "todo", "250131-120000-other-ticket.md")
					require.NoError(t, os.WriteFile(ticketFile, []byte("test content"), 0644))
					// Also create the current ticket file
					currentFile := filepath.Join(tmpDir, "tickets", "doing", "250131-120000-current-ticket.md")
					require.NoError(t, os.WriteFile(currentFile, []byte("current ticket content"), 0644))
				}
			}

			// Create app with mocks
			app := &App{
				Manager:     mockManager,
				Git:         mockGit,
				ProjectRoot: tmpDir,
				Config: &config.Config{
					Git: config.GitConfig{
						DefaultBranch: "main",
					},
					Tickets: config.TicketsConfig{
						Dir:      "tickets",
						TodoDir:  "todo",
						DoingDir: "doing",
						DoneDir:  "done",
					},
				},
				Output: NewOutputWriter(os.Stdout, os.Stderr, FormatText),
			}

			// Execute
			_, err := app.CloseTicketByID(context.Background(), tt.ticketID, tt.reason, false)

			// Verify
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockManager.AssertExpectations(t)
			mockGit.AssertExpectations(t)
		})
	}
}

func TestApp_CloseTicketWithReason(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		reason        string
		setupMocks    func(*mocks.MockTicketManager, *mocks.MockGitClient, string)
		expectedError bool
		errorContains string
	}{
		{
			name:   "close current ticket with reason",
			reason: "Cancelled due to requirements change",
			setupMocks: func(tm *mocks.MockTicketManager, gc *mocks.MockGitClient, tmpDir string) {
				testTicket := &ticket.Ticket{
					ID:          "250131-120000-current-ticket",
					Path:        filepath.Join(tmpDir, "tickets/doing/250131-120000-current-ticket.md"),
					Priority:    2,
					Description: "Current ticket",
					CreatedAt:   ticket.RFC3339Time{Time: time.Now()},
					StartedAt:   ticket.NewRFC3339TimePtr(&time.Time{}),
					Content:     "# Current Ticket\n\nWork in progress.",
				}

				// Mock getting current ticket
				tm.On("GetCurrentTicket", mock.Anything).Return(testTicket, nil)

				// Mock updating ticket with reason (only once in moveTicketToDoneWithReason)
				tm.On("Update", mock.Anything, testTicket).Return(nil).Times(1)

				// Mock removing current ticket symlink
				tm.On("SetCurrentTicket", mock.Anything, (*ticket.Ticket)(nil)).Return(nil)

				// Mock git operations
				gc.On("HasUncommittedChanges", mock.Anything).Return(false, nil)
				gc.On("CurrentBranch", mock.Anything).Return("250131-120000-current-ticket", nil)
				gc.On("Add", mock.Anything, "-A", mock.Anything, mock.Anything).Return(nil)
				gc.On("Commit", mock.Anything, mock.MatchedBy(func(msg string) bool {
					return strings.Contains(msg, "Close ticket:")
				})).Return(nil)
			},
			expectedError: false,
		},
		{
			name:   "error when no current ticket",
			reason: "Some reason",
			setupMocks: func(tm *mocks.MockTicketManager, gc *mocks.MockGitClient, tmpDir string) {
				tm.On("GetCurrentTicket", mock.Anything).Return(nil, nil)
			},
			expectedError: true,
			errorContains: "No active ticket",
		},
		{
			name:   "error when uncommitted changes",
			reason: "Some reason",
			setupMocks: func(tm *mocks.MockTicketManager, gc *mocks.MockGitClient, tmpDir string) {
				testTicket := &ticket.Ticket{
					ID:          "250131-120000-current-ticket",
					Path:        filepath.Join(tmpDir, "tickets/doing/250131-120000-current-ticket.md"),
					Priority:    2,
					Description: "Current ticket",
					CreatedAt:   ticket.RFC3339Time{Time: time.Now()},
					StartedAt:   ticket.NewRFC3339TimePtr(&time.Time{}),
					Content:     "# Current Ticket\n\nWork in progress.",
				}

				tm.On("GetCurrentTicket", mock.Anything).Return(testTicket, nil)
				gc.On("HasUncommittedChanges", mock.Anything).Return(true, nil)
			},
			expectedError: true,
			errorContains: "Uncommitted changes",
		},
		{
			name:   "error when empty reason",
			reason: "",
			setupMocks: func(tm *mocks.MockTicketManager, gc *mocks.MockGitClient, tmpDir string) {
				// No mocks needed, validation happens before any calls
			},
			expectedError: true,
			errorContains: "Empty reason",
		},
		{
			name:   "error when whitespace-only reason",
			reason: "   \t  ",
			setupMocks: func(tm *mocks.MockTicketManager, gc *mocks.MockGitClient, tmpDir string) {
				// No mocks needed, validation happens before any calls
			},
			expectedError: true,
			errorContains: "Empty reason",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for tests
			tmpDir := t.TempDir()

			// Create required directories
			require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "tickets", "todo"), 0755))
			require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "tickets", "doing"), 0755))
			require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "tickets", "done"), 0755))

			// Setup mocks
			mockManager := &mocks.MockTicketManager{}
			mockGit := &mocks.MockGitClient{}

			// Setup test mocks
			tt.setupMocks(mockManager, mockGit, tmpDir)

			// Create the ticket file for CloseTicketWithReason test
			if tt.name == "close current ticket with reason" {
				ticketFile := filepath.Join(tmpDir, "tickets", "doing", "250131-120000-current-ticket.md")
				require.NoError(t, os.WriteFile(ticketFile, []byte("test content"), 0644))
			}

			// Create app with mocks
			app := &App{
				Manager:     mockManager,
				Git:         mockGit,
				ProjectRoot: tmpDir,
				Config: &config.Config{
					Git: config.GitConfig{
						DefaultBranch: "main",
					},
					Tickets: config.TicketsConfig{
						Dir:      "tickets",
						TodoDir:  "todo",
						DoingDir: "doing",
						DoneDir:  "done",
					},
				},
				Output: NewOutputWriter(os.Stdout, os.Stderr, FormatText),
			}

			// Execute
			_, err := app.CloseTicketWithReason(context.Background(), tt.reason, false)

			// Verify
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockManager.AssertExpectations(t)
			mockGit.AssertExpectations(t)
		})
	}
}
