package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/mocks"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestApp_NewTicket(t *testing.T) {
	tests := []struct {
		name          string
		slug          string
		outputFormat  OutputFormat
		setupMocks    func(*mocks.MockTicketManager, *mocks.MockGitClient)
		expectedError bool
		errorMessage  string
	}{
		{
			name:         "successful ticket creation",
			slug:         "test-feature",
			outputFormat: FormatText,
			setupMocks: func(m *mocks.MockTicketManager, g *mocks.MockGitClient) {
				// Git mocks for parent detection
				g.On("CurrentBranch", mock.Anything).Return("main", nil)

				// Manager mocks
				newTicket := &ticket.Ticket{
					ID:          "250131-120000-test-feature",
					Path:        "/path/to/ticket.md",
					Priority:    2,
					Description: "",
				}
				m.On("Create", mock.Anything, "test-feature").Return(newTicket, nil)
			},
			expectedError: false,
		},
		{
			name:         "invalid slug",
			slug:         "invalid slug with spaces",
			outputFormat: FormatText,
			setupMocks: func(m *mocks.MockTicketManager, g *mocks.MockGitClient) {
				// No mock setup needed as validation happens before manager call
			},
			expectedError: true,
			errorMessage:  "Invalid slug format",
		},
		{
			name:         "ticket already exists",
			slug:         "existing-feature",
			outputFormat: FormatText,
			setupMocks: func(m *mocks.MockTicketManager, g *mocks.MockGitClient) {
				g.On("CurrentBranch", mock.Anything).Return("main", nil)
				m.On("Create", mock.Anything, "existing-feature").Return(nil, assert.AnError)
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
				tt.setupMocks(mockManager, mockGit)
			}

			// Create app with mock
			app := &App{
				Config:      config.Default(),
				Manager:     mockManager,
				Git:         mockGit,
				ProjectRoot: "/test/project",
			}

			// Execute
			ctx := context.Background()
			err := app.NewTicket(ctx, tt.slug, tt.outputFormat)

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

func TestApp_ListTickets(t *testing.T) {
	tests := []struct {
		name          string
		status        ticket.Status
		count         int
		outputFormat  OutputFormat
		setupMocks    func(*mocks.MockTicketManager)
		expectedError bool
	}{
		{
			name:         "list active tickets (default)",
			status:       "",
			count:        10,
			outputFormat: FormatText,
			setupMocks: func(m *mocks.MockTicketManager) {
				tickets := []ticket.Ticket{
					{ID: "ticket1", Priority: 1},
					{ID: "ticket2", Priority: 2},
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
			setupMocks: func(m *mocks.MockTicketManager) {
				tickets := []ticket.Ticket{
					{ID: "todo1", Priority: 1},
				}
				m.On("List", mock.Anything, ticket.StatusFilterTodo).Return(tickets, nil)
				// JSON format also fetches all tickets for summary
				m.On("List", mock.Anything, ticket.StatusFilterAll).Return(tickets, nil)
			},
			expectedError: false,
		},
		{
			name:         "list error",
			status:       ticket.StatusDoing,
			count:        10,
			outputFormat: FormatText,
			setupMocks: func(m *mocks.MockTicketManager) {
				m.On("List", mock.Anything, ticket.StatusFilterDoing).Return(nil, assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock
			mockManager := new(mocks.MockTicketManager)
			if tt.setupMocks != nil {
				tt.setupMocks(mockManager)
			}

			// Create app with mock
			app := &App{
				Config:      config.Default(),
				Manager:     mockManager,
				ProjectRoot: "/test/project",
			}

			// Execute
			err := app.ListTickets(context.Background(), tt.status, tt.count, tt.outputFormat)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			mockManager.AssertExpectations(t)
		})
	}
}

func TestApp_StartTicket_WithMocks(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "ticketflow-test-*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tmpDir)
		require.NoError(t, err)
	}()

	// Create tickets directories
	todoDir := filepath.Join(tmpDir, "tickets", "todo")
	doingDir := filepath.Join(tmpDir, "tickets", "doing")
	require.NoError(t, os.MkdirAll(todoDir, 0755))
	require.NoError(t, os.MkdirAll(doingDir, 0755))

	tests := []struct {
		name          string
		ticketID      string
		setupMocks    func(*mocks.MockTicketManager, *mocks.MockGitClient, string)
		expectedError bool
		errorMessage  string
	}{
		{
			name:     "successful start",
			ticketID: "250131-120000-test",
			setupMocks: func(tm *mocks.MockTicketManager, gc *mocks.MockGitClient, tmpDir string) {
				// Create ticket file
				ticketPath := filepath.Join(tmpDir, "tickets", "todo", "250131-120000-test.md")
				require.NoError(t, os.WriteFile(ticketPath, []byte("test content"), 0644))

				testTicket := &ticket.Ticket{
					ID:       "250131-120000-test",
					Path:     ticketPath,
					Priority: 2,
				}

				// validateTicketForStart calls
				tm.On("Get", mock.Anything, "250131-120000-test").Return(testTicket, nil)
				tm.On("GetCurrentTicket", mock.Anything).Return(nil, nil).Maybe() // No current ticket - Maybe() allows 0 or more calls

				// checkWorkspaceForStart calls
				// HasUncommittedChanges is NOT called when worktree is enabled

				// detectParentBranch calls
				gc.On("CurrentBranch", mock.Anything).Return("main", nil)
				// No Manager.Get("main") call because we skip when currentBranch == defaultBranch

				// setupTicketBranch calls (worktree is enabled in the test config)
				// No CreateBranch call when worktree is enabled

				// moveTicketToDoing calls
				newPath := filepath.Join(tmpDir, "tickets", "doing", "250131-120000-test.md")
				// Update is called after Start() to save the started timestamp
				tm.On("Update", mock.Anything, mock.MatchedBy(func(t *ticket.Ticket) bool {
					return t.ID == "250131-120000-test" && t.Path == newPath && t.StartedAt.Time != nil
				})).Return(nil)
				// Git add with -A flag to handle rename
				todoDir := filepath.Join(tmpDir, "tickets", "todo")
				doingDir := filepath.Join(tmpDir, "tickets", "doing")
				gc.On("Add", mock.Anything, "-A", todoDir, doingDir).Return(nil)
				gc.On("Commit", mock.Anything, "Start ticket: 250131-120000-test").Return(nil)
				tm.On("SetCurrentTicket", mock.Anything, mock.MatchedBy(func(t *ticket.Ticket) bool {
					return t.ID == "250131-120000-test" && t.Path == newPath
				})).Return(nil)

				// For worktree mode, we need these additional calls
				gc.On("HasWorktree", mock.Anything, "250131-120000-test").Return(false, nil)
				gc.On("Checkout", mock.Anything, "main").Return(nil)

				// AddWorktree should create the directory
				worktreePath := filepath.Join(tmpDir, ".worktrees", "250131-120000-test")
				gc.On("AddWorktree", mock.Anything, mock.MatchedBy(func(path string) bool {
					// Create the worktree directory when this is called
					err := os.MkdirAll(path, 0755)
					require.NoError(t, err)
					// Also create the tickets/doing directory in the worktree
					ticketsDoingPath := filepath.Join(path, "tickets", "doing")
					err = os.MkdirAll(ticketsDoingPath, 0755)
					require.NoError(t, err)
					return path == worktreePath
				}), "250131-120000-test").Return(nil)
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
			cfg.Worktree.Enabled = true
			cfg.Worktree.BaseDir = filepath.Join(tmpDir, ".worktrees") // Use absolute path
			cfg.Worktree.InitCommands = []string{}                     // Disable init commands for test
			cfg.Git.DefaultBranch = "main"
			cfg.Tickets.Dir = "tickets"

			app := &App{
				Config:      cfg,
				Manager:     mockManager,
				Git:         mockGit,
				ProjectRoot: tmpDir,
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
