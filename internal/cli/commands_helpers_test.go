package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/mocks"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestCountTicketsByStatus(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		tickets   []ticket.Ticket
		wantTodo  int
		wantDoing int
		wantDone  int
	}{
		{
			name:      "empty list",
			tickets:   []ticket.Ticket{},
			wantTodo:  0,
			wantDoing: 0,
			wantDone:  0,
		},
		{
			name: "mixed statuses",
			tickets: []ticket.Ticket{
				{Path: "/todo/ticket1.md"}, // Todo - no times set
				{Path: "/todo/ticket2.md"}, // Todo - no times set
				{Path: "/doing/ticket3.md", StartedAt: ticket.RFC3339TimePtr{Time: &now}},                                             // Doing - started
				{Path: "/done/ticket4.md", StartedAt: ticket.RFC3339TimePtr{Time: &now}, ClosedAt: ticket.RFC3339TimePtr{Time: &now}}, // Done
				{Path: "/done/ticket5.md", StartedAt: ticket.RFC3339TimePtr{Time: &now}, ClosedAt: ticket.RFC3339TimePtr{Time: &now}}, // Done
				{Path: "/done/ticket6.md", StartedAt: ticket.RFC3339TimePtr{Time: &now}, ClosedAt: ticket.RFC3339TimePtr{Time: &now}}, // Done
			},
			wantTodo:  2,
			wantDoing: 1,
			wantDone:  3,
		},
		{
			name: "all todo",
			tickets: []ticket.Ticket{
				{Path: "/todo/ticket1.md"},
				{Path: "/todo/ticket2.md"},
			},
			wantTodo:  2,
			wantDoing: 0,
			wantDone:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{}
			todoCount, doingCount, doneCount := app.countTicketsByStatus(tt.tickets)
			assert.Equal(t, tt.wantTodo, todoCount, "todo count mismatch")
			assert.Equal(t, tt.wantDoing, doingCount, "doing count mismatch")
			assert.Equal(t, tt.wantDone, doneCount, "done count mismatch")
		})
	}
}

func TestCalculateWorkDuration(t *testing.T) {
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour)

	tests := []struct {
		name   string
		ticket *ticket.Ticket
		want   string
	}{
		{
			name: "no started time",
			ticket: &ticket.Ticket{
				StartedAt: ticket.RFC3339TimePtr{},
				ClosedAt:  ticket.RFC3339TimePtr{Time: &now},
			},
			want: "",
		},
		{
			name: "no closed time",
			ticket: &ticket.Ticket{
				StartedAt: ticket.RFC3339TimePtr{Time: &oneHourAgo},
				ClosedAt:  ticket.RFC3339TimePtr{},
			},
			want: "",
		},
		{
			name: "one hour duration",
			ticket: &ticket.Ticket{
				StartedAt: ticket.RFC3339TimePtr{Time: &oneHourAgo},
				ClosedAt:  ticket.RFC3339TimePtr{Time: &now},
			},
			want: "1h",
		},
		{
			name: "two hours duration",
			ticket: &ticket.Ticket{
				StartedAt: ticket.RFC3339TimePtr{Time: &twoHoursAgo},
				ClosedAt:  ticket.RFC3339TimePtr{Time: &now},
			},
			want: "2h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{}
			got := app.calculateWorkDuration(tt.ticket)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractParentTicketID(t *testing.T) {
	tests := []struct {
		name   string
		ticket *ticket.Ticket
		want   string
	}{
		{
			name: "no related field",
			ticket: &ticket.Ticket{
				Related: []string{},
			},
			want: "",
		},
		{
			name: "has parent",
			ticket: &ticket.Ticket{
				Related: []string{"parent:parent-ticket-id"},
			},
			want: "parent-ticket-id",
		},
		{
			name: "multiple related with parent",
			ticket: &ticket.Ticket{
				Related: []string{"related:other-ticket", "parent:parent-ticket-id", "blocked-by:blocker"},
			},
			want: "parent-ticket-id",
		},
		{
			name: "no parent in related",
			ticket: &ticket.Ticket{
				Related: []string{"related:other-ticket", "blocked-by:blocker"},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{}
			got := app.extractParentTicketID(tt.ticket)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCheckExistingWorktree(t *testing.T) {
	tests := []struct {
		name            string
		worktreeEnabled bool
		hasWorktree     bool
		checkError      error
		wantErr         bool
		errContains     string
	}{
		{
			name:            "worktree disabled",
			worktreeEnabled: false,
			hasWorktree:     false,
			checkError:      nil,
			wantErr:         false,
		},
		{
			name:            "worktree does not exist",
			worktreeEnabled: true,
			hasWorktree:     false,
			checkError:      nil,
			wantErr:         false,
		},
		{
			name:            "worktree already exists",
			worktreeEnabled: true,
			hasWorktree:     true,
			checkError:      nil,
			wantErr:         true,
			errContains:     "Worktree already exists",
		},
		{
			name:            "check worktree error",
			worktreeEnabled: true,
			hasWorktree:     false,
			checkError:      assert.AnError,
			wantErr:         true,
			errContains:     "failed to check worktree",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGit := new(mocks.MockGitClient)
			mockManager := new(mocks.MockTicketManager)

			app := &App{
				Config: &config.Config{
					Worktree: config.WorktreeConfig{
						Enabled: tt.worktreeEnabled,
					},
				},
				Git:     mockGit,
				Manager: mockManager,
			}

			testTicket := &ticket.Ticket{ID: "test-ticket"}

			if tt.worktreeEnabled {
				mockGit.On("HasWorktree", "test-ticket").Return(tt.hasWorktree, tt.checkError)
			}

			err := app.checkExistingWorktree(testTicket)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockGit.AssertExpectations(t)
		})
	}
}
