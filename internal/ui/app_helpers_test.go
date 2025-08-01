package ui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/mocks"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestValidateTicketForStart(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name    string
		ticket  *ticket.Ticket
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid todo ticket",
			ticket: &ticket.Ticket{
				ID:   "test-ticket",
				Path: "/todo/test-ticket.md",
			},
			wantErr: false,
		},
		{
			name: "already doing ticket",
			ticket: &ticket.Ticket{
				ID:        "test-ticket",
				Path:      "/doing/test-ticket.md",
				StartedAt: ticket.RFC3339TimePtr{Time: &now},
			},
			wantErr: true,
			errMsg:  "already in progress",
		},
		{
			name: "done ticket",
			ticket: &ticket.Ticket{
				ID:        "test-ticket",
				Path:      "/done/test-ticket.md",
				StartedAt: ticket.RFC3339TimePtr{Time: &now},
				ClosedAt:  ticket.RFC3339TimePtr{Time: &now},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Model{}
			err := m.validateTicketForStart(tt.ticket)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckWorkspaceForStart(t *testing.T) {
	tests := []struct {
		name              string
		worktreeEnabled   bool
		hasUncommitted    bool
		checkError        error
		wantErr           bool
		errContains       string
	}{
		{
			name:            "worktree enabled - no check needed",
			worktreeEnabled: true,
			hasUncommitted:  true,
			wantErr:         false,
		},
		{
			name:            "worktree disabled - clean workspace",
			worktreeEnabled: false,
			hasUncommitted:  false,
			wantErr:         false,
		},
		{
			name:            "worktree disabled - dirty workspace",
			worktreeEnabled: false,
			hasUncommitted:  true,
			wantErr:         true,
			errContains:     "uncommitted changes detected",
		},
		{
			name:            "worktree disabled - check error",
			worktreeEnabled: false,
			checkError:      assert.AnError,
			wantErr:         true,
			errContains:     "failed to check git status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGit := new(mocks.MockGitClient)
			mockManager := new(mocks.MockTicketManager)

			m := &Model{
				config: &config.Config{
					Worktree: config.WorktreeConfig{
						Enabled: tt.worktreeEnabled,
					},
				},
				git:     mockGit,
				manager: mockManager,
			}

			if !tt.worktreeEnabled {
				mockGit.On("HasUncommittedChanges").Return(tt.hasUncommitted, tt.checkError)
			}

			err := m.checkWorkspaceForStart()

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

func TestValidateTicketForClose(t *testing.T) {
	tests := []struct {
		name          string
		currentTicket *ticket.Ticket
		targetTicket  *ticket.Ticket
		getError      error
		wantErr       bool
		errContains   string
	}{
		{
			name: "valid close - same ticket",
			currentTicket: &ticket.Ticket{
				ID: "test-ticket",
			},
			targetTicket: &ticket.Ticket{
				ID: "test-ticket",
			},
			wantErr: false,
		},
		{
			name:          "no current ticket",
			currentTicket: nil,
			targetTicket: &ticket.Ticket{
				ID: "test-ticket",
			},
			wantErr:     true,
			errContains: "can only close the current active ticket",
		},
		{
			name: "different ticket",
			currentTicket: &ticket.Ticket{
				ID: "other-ticket",
			},
			targetTicket: &ticket.Ticket{
				ID: "test-ticket",
			},
			wantErr:     true,
			errContains: "can only close the current active ticket",
		},
		{
			name:        "get current ticket error",
			getError:    assert.AnError,
			wantErr:     true,
			errContains: "failed to get current ticket",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := new(mocks.MockTicketManager)

			m := &Model{
				manager: mockManager,
			}

			mockManager.On("GetCurrentTicket").Return(tt.currentTicket, tt.getError)

			err := m.validateTicketForClose(tt.targetTicket)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockManager.AssertExpectations(t)
		})
	}
}