package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
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
			err := app.NewTicket(context.Background(), tt.slug, tt.outputFormat)

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
			err := app.StartTicket(context.Background(), tt.ticketID)

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
