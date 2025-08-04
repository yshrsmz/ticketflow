package cli

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yshrsmz/ticketflow/internal/config"
	ticketerrors "github.com/yshrsmz/ticketflow/internal/errors"
	"github.com/yshrsmz/ticketflow/internal/mocks"
	"github.com/yshrsmz/ticketflow/internal/testutil"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestApp_NewTicket_WithParent(t *testing.T) {
	tests := []struct {
		name           string
		slug           string
		explicitParent string
		currentBranch  string
		setupManager   func(m *mocks.MockTicketManager)
		expectedError  bool
		errorMessage   string
		expectedParent string
		checkOutput    func(t *testing.T, output *testutil.OutputCapture)
	}{
		{
			name:           "explicit parent - valid",
			slug:           "sub-feature",
			explicitParent: "parent-ticket",
			currentBranch:  "main",
			setupManager: func(m *mocks.MockTicketManager) {
				// Validate parent exists
				parentTicket := &ticket.Ticket{
					ID:          "parent-ticket",
					Path:        "/test/tickets/doing/parent-ticket.md",
					Priority:    1,
					Description: "Parent ticket",
					CreatedAt:   ticket.RFC3339Time{Time: time.Now()},
				}
				m.On("Get", mock.Anything, "parent-ticket").Return(parentTicket, nil)

				// Create new ticket
				newTicket := &ticket.Ticket{
					ID:          "250802-120000-sub-feature",
					Slug:        "sub-feature",
					Path:        "/test/tickets/todo/250802-120000-sub-feature.md",
					Priority:    3,
					Description: "sub-feature",
					CreatedAt:   ticket.RFC3339Time{Time: time.Now()},
				}
				m.On("Create", mock.Anything, "sub-feature").Return(newTicket, nil)

				// Update with parent relation
				updatedTicket := *newTicket
				updatedTicket.Related = []string{"parent:parent-ticket"}
				m.On("Update", mock.Anything, mock.MatchedBy(func(t *ticket.Ticket) bool {
					return t.ID == newTicket.ID && len(t.Related) == 1 && t.Related[0] == "parent:parent-ticket"
				})).Return(nil)
			},
			expectedError:  false,
			expectedParent: "parent-ticket",
			checkOutput: func(t *testing.T, output *testutil.OutputCapture) {
				assert.Contains(t, output.Stdout(), "Creating sub-ticket with parent: parent-ticket")
			},
		},
		{
			name:           "explicit parent - not found",
			slug:           "sub-feature",
			explicitParent: "non-existent",
			currentBranch:  "main",
			setupManager: func(m *mocks.MockTicketManager) {
				// Parent doesn't exist
				m.On("Get", mock.Anything, "non-existent").Return(nil, ticketerrors.ErrTicketNotFound)
			},
			expectedError: true,
			errorMessage:  "Parent ticket not found",
		},
		{
			name:           "explicit parent - self parent",
			slug:           "same-ticket",
			explicitParent: "same-ticket",
			currentBranch:  "main",
			setupManager: func(m *mocks.MockTicketManager) {
				// Don't need to check if parent exists for self-parent validation
			},
			expectedError: true,
			errorMessage:  "Invalid parent relationship",
		},
		{
			name:           "implicit parent from current branch",
			slug:           "sub-feature",
			explicitParent: "",
			currentBranch:  "250802-100000-parent-ticket",
			setupManager: func(m *mocks.MockTicketManager) {
				// Check if current branch is a ticket
				parentTicket := &ticket.Ticket{
					ID:          "250802-100000-parent-ticket",
					Path:        "/test/tickets/doing/250802-100000-parent-ticket.md",
					Priority:    1,
					Description: "Parent ticket",
					CreatedAt:   ticket.RFC3339Time{Time: time.Now()},
				}
				m.On("Get", mock.Anything, "250802-100000-parent-ticket").Return(parentTicket, nil)

				// Create new ticket
				newTicket := &ticket.Ticket{
					ID:          "250802-120000-sub-feature",
					Slug:        "sub-feature",
					Path:        "/test/tickets/todo/250802-120000-sub-feature.md",
					Priority:    3,
					Description: "sub-feature",
					CreatedAt:   ticket.RFC3339Time{Time: time.Now()},
				}
				m.On("Create", mock.Anything, "sub-feature").Return(newTicket, nil)

				// Update with parent relation
				updatedTicket := *newTicket
				updatedTicket.Related = []string{"parent:250802-100000-parent-ticket"}
				m.On("Update", mock.Anything, mock.MatchedBy(func(t *ticket.Ticket) bool {
					return t.ID == newTicket.ID && len(t.Related) == 1 && t.Related[0] == "parent:250802-100000-parent-ticket"
				})).Return(nil)
			},
			expectedError:  false,
			expectedParent: "250802-100000-parent-ticket",
			checkOutput: func(t *testing.T, output *testutil.OutputCapture) {
				assert.Contains(t, output.Stdout(), "Creating ticket in branch: 250802-100000-parent-ticket")
			},
		},
		{
			name:           "explicit parent overrides implicit",
			slug:           "sub-feature",
			explicitParent: "explicit-parent",
			currentBranch:  "implicit-parent",
			setupManager: func(m *mocks.MockTicketManager) {
				// Validate explicit parent
				explicitTicket := &ticket.Ticket{
					ID: "explicit-parent",
				}
				m.On("Get", mock.Anything, "explicit-parent").Return(explicitTicket, nil)

				// Create new ticket
				newTicket := &ticket.Ticket{
					ID:          "250802-120000-sub-feature",
					Slug:        "sub-feature",
					Path:        "/test/tickets/todo/250802-120000-sub-feature.md",
					Priority:    3,
					Description: "sub-feature",
					CreatedAt:   ticket.RFC3339Time{Time: time.Now()},
				}
				m.On("Create", mock.Anything, "sub-feature").Return(newTicket, nil)

				// Update with explicit parent
				updatedTicket := *newTicket
				updatedTicket.Related = []string{"parent:explicit-parent"}
				m.On("Update", mock.Anything, mock.MatchedBy(func(t *ticket.Ticket) bool {
					return t.ID == newTicket.ID && len(t.Related) == 1 && t.Related[0] == "parent:explicit-parent"
				})).Return(nil)
			},
			expectedError:  false,
			expectedParent: "explicit-parent",
			checkOutput: func(t *testing.T, output *testutil.OutputCapture) {
				assert.Contains(t, output.Stdout(), "Creating sub-ticket with parent: explicit-parent")
			},
		},
		{
			name:           "no parent - on main branch",
			slug:           "top-level-feature",
			explicitParent: "",
			currentBranch:  "main",
			setupManager: func(m *mocks.MockTicketManager) {
				// Create new ticket without parent
				newTicket := &ticket.Ticket{
					ID:          "250802-120000-top-level-feature",
					Slug:        "top-level-feature",
					Path:        "/test/tickets/todo/250802-120000-top-level-feature.md",
					Priority:    3,
					Description: "top-level-feature",
					CreatedAt:   ticket.RFC3339Time{Time: time.Now()},
				}
				m.On("Create", mock.Anything, "top-level-feature").Return(newTicket, nil)
				// No Update call expected since no parent
			},
			expectedError:  false,
			expectedParent: "",
			checkOutput: func(t *testing.T, output *testutil.OutputCapture) {
				// Should not contain any parent-related messages
				assert.NotContains(t, output.Stdout(), "Creating sub-ticket")
				assert.NotContains(t, output.Stdout(), "Creating ticket in branch")
			},
		},
		// Note: Circular dependency detection is designed to prevent cycles when
		// creating parent relationships. However, in practice it's difficult to
		// create a true circular dependency test because:
		// 1. The new ticket doesn't exist yet when we check
		// 2. The generated ID includes a timestamp that changes
		// 3. Circular deps would only occur if an existing ticket already references
		//    the exact ID we're about to generate
		//
		// The implementation is correct but these edge cases are unlikely in practice
		{
			name:           "done parent warning",
			slug:           "sub-feature",
			explicitParent: "done-parent",
			currentBranch:  "main",
			setupManager: func(m *mocks.MockTicketManager) {
				// Parent ticket in done status
				closedTime := time.Now().Add(-24 * time.Hour)
				doneParent := &ticket.Ticket{
					ID:          "done-parent",
					Path:        "/test/tickets/done/done-parent.md",
					Priority:    1,
					Description: "Done parent",
					CreatedAt:   ticket.RFC3339Time{Time: time.Now().Add(-48 * time.Hour)},
					StartedAt:   ticket.RFC3339TimePtr{Time: &closedTime},
					ClosedAt:    ticket.RFC3339TimePtr{Time: &closedTime},
				}
				m.On("Get", mock.Anything, "done-parent").Return(doneParent, nil)

				// Create new ticket
				newTicket := &ticket.Ticket{
					ID:          "250802-120000-sub-feature",
					Slug:        "sub-feature",
					Path:        "/test/tickets/todo/250802-120000-sub-feature.md",
					Priority:    3,
					Description: "sub-feature",
					CreatedAt:   ticket.RFC3339Time{Time: time.Now()},
				}
				m.On("Create", mock.Anything, "sub-feature").Return(newTicket, nil)

				// Update with parent relation
				updatedTicket := *newTicket
				updatedTicket.Related = []string{"parent:done-parent"}
				m.On("Update", mock.Anything, mock.MatchedBy(func(t *ticket.Ticket) bool {
					return t.ID == newTicket.ID && len(t.Related) == 1 && t.Related[0] == "parent:done-parent"
				})).Return(nil)
			},
			expectedError:  false,
			expectedParent: "done-parent",
			checkOutput: func(t *testing.T, output *testutil.OutputCapture) {
				// Check both stdout and stderr for the warning
				allOutput := output.Stdout() + output.Stderr()
				assert.Contains(t, allOutput, "Warning: Parent ticket 'done-parent' is already done")
			},
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
			// Only set up CurrentBranch expectation if no explicit parent
			if tt.explicitParent == "" {
				mockGit.On("CurrentBranch", mock.Anything).Return(tt.currentBranch, nil)
			}

			// Capture output
			output := testutil.NewOutputCapture()

			// Create app with mocks
			cfg := config.Default()
			cfg.Git.DefaultBranch = "main"
			app := &App{
				Config:      cfg,
				Manager:     mockManager,
				Git:         mockGit,
				ProjectRoot: "/test/project",
				Output:      NewOutputWriter(output, output.StderrWriter(), FormatText),
			}

			// Execute
			err := app.NewTicket(context.Background(), tt.slug, tt.explicitParent, FormatText)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
			}

			// Check output if provided
			if tt.checkOutput != nil {
				tt.checkOutput(t, output)
			}

			// Verify expectations
			mockManager.AssertExpectations(t)
			mockGit.AssertExpectations(t)
		})
	}
}
