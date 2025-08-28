package ui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/mocks"
	"github.com/yshrsmz/ticketflow/internal/ticket"
	"github.com/yshrsmz/ticketflow/internal/ui/components"
)

func TestIsCurrentTicket(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		current  *ticket.Ticket
		target   *ticket.Ticket
		expected bool
	}{
		{
			name:     "both nil",
			current:  nil,
			target:   nil,
			expected: false,
		},
		{
			name:     "current nil",
			current:  nil,
			target:   &ticket.Ticket{ID: "test"},
			expected: false,
		},
		{
			name:     "target nil",
			current:  &ticket.Ticket{ID: "test"},
			target:   nil,
			expected: false,
		},
		{
			name:     "same ticket",
			current:  &ticket.Ticket{ID: "test"},
			target:   &ticket.Ticket{ID: "test"},
			expected: true,
		},
		{
			name:     "different tickets",
			current:  &ticket.Ticket{ID: "test1"},
			target:   &ticket.Ticket{ID: "test2"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCurrentTicket(tt.current, tt.target)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckBranchMerged(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		ticketID      string
		defaultBranch string
		isMerged      bool
		mergeError    error
		wantErr       bool
		errContains   string
	}{
		{
			name:          "branch is merged",
			ticketID:      "test-ticket",
			defaultBranch: "main",
			isMerged:      true,
			wantErr:       false,
		},
		{
			name:          "branch not merged",
			ticketID:      "test-ticket",
			defaultBranch: "main",
			isMerged:      false,
			wantErr:       false,
		},
		{
			name:          "no default branch configured",
			ticketID:      "test-ticket",
			defaultBranch: "",
			wantErr:       true,
			errContains:   "default branch not configured",
		},
		{
			name:          "git error",
			ticketID:      "test-ticket",
			defaultBranch: "main",
			mergeError:    assert.AnError,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGit := new(mocks.MockGitClient)
			m := &Model{
				config: &config.Config{
					Git: config.GitConfig{
						DefaultBranch: tt.defaultBranch,
					},
				},
				git: mockGit,
			}

			if tt.defaultBranch != "" {
				mockGit.On("IsBranchMerged", mock.Anything, tt.ticketID, tt.defaultBranch).
					Return(tt.isMerged, tt.mergeError)
			}

			result, err := m.checkBranchMerged(tt.ticketID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.isMerged, result)
			}

			mockGit.AssertExpectations(t)
		})
	}
}

func TestCheckCloseRequirements(t *testing.T) {
	t.Parallel()
	now := time.Now()

	tests := []struct {
		name            string
		targetTicket    *ticket.Ticket
		currentTicket   *ticket.Ticket
		getCurrentError error
		isMerged        bool
		mergeError      error
		defaultBranch   string
		expectRequired  bool
		expectIsCurrent bool
	}{
		{
			name: "current ticket - reason optional",
			targetTicket: &ticket.Ticket{
				ID:   "test-ticket",
				Path: "/doing/test-ticket.md",
			},
			currentTicket: &ticket.Ticket{
				ID:   "test-ticket",
				Path: "/doing/test-ticket.md",
			},
			defaultBranch:   "main",
			expectRequired:  false,
			expectIsCurrent: true,
		},
		{
			name: "non-current ticket merged - reason optional",
			targetTicket: &ticket.Ticket{
				ID:   "other-ticket",
				Path: "/todo/other-ticket.md",
			},
			currentTicket: &ticket.Ticket{
				ID:        "test-ticket",
				Path:      "/doing/test-ticket.md",
				StartedAt: ticket.RFC3339TimePtr{Time: &now},
			},
			defaultBranch:   "main",
			isMerged:        true,
			expectRequired:  false,
			expectIsCurrent: false,
		},
		{
			name: "non-current ticket not merged - reason required",
			targetTicket: &ticket.Ticket{
				ID:   "other-ticket",
				Path: "/todo/other-ticket.md",
			},
			currentTicket: &ticket.Ticket{
				ID:        "test-ticket",
				Path:      "/doing/test-ticket.md",
				StartedAt: ticket.RFC3339TimePtr{Time: &now},
			},
			defaultBranch:   "main",
			isMerged:        false,
			expectRequired:  true,
			expectIsCurrent: false,
		},
		{
			name: "no current ticket - check merge status",
			targetTicket: &ticket.Ticket{
				ID:   "test-ticket",
				Path: "/todo/test-ticket.md",
			},
			getCurrentError: assert.AnError,
			defaultBranch:   "main",
			isMerged:        false,
			expectRequired:  true,
			expectIsCurrent: false,
		},
		{
			name: "merge check fails - assume required",
			targetTicket: &ticket.Ticket{
				ID:   "other-ticket",
				Path: "/todo/other-ticket.md",
			},
			currentTicket: &ticket.Ticket{
				ID:   "test-ticket",
				Path: "/doing/test-ticket.md",
			},
			defaultBranch:   "main",
			mergeError:      assert.AnError,
			expectRequired:  true,
			expectIsCurrent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := new(mocks.MockTicketManager)
			mockGit := new(mocks.MockGitClient)

			m := &Model{
				config: &config.Config{
					Git: config.GitConfig{
						DefaultBranch: tt.defaultBranch,
					},
				},
				manager: mockManager,
				git:     mockGit,
			}

			// Setup mocks
			mockManager.On("GetCurrentTicket", mock.Anything).
				Return(tt.currentTicket, tt.getCurrentError)

			// Only mock IsBranchMerged if it's not the current ticket and default branch is configured
			if (tt.currentTicket == nil || tt.currentTicket.ID != tt.targetTicket.ID) && tt.defaultBranch != "" {
				mockGit.On("IsBranchMerged", mock.Anything, tt.targetTicket.ID, tt.defaultBranch).
					Return(tt.isMerged, tt.mergeError)
			}

			// Execute the command
			cmd := m.checkCloseRequirements(tt.targetTicket)
			msg := cmd()

			// Assert the result
			result, ok := msg.(closeRequirementsMsg)
			assert.True(t, ok, "expected closeRequirementsMsg type")
			assert.Equal(t, tt.targetTicket.ID, result.ticket.ID)
			assert.Equal(t, tt.expectRequired, result.requireReason)
			assert.Equal(t, tt.expectIsCurrent, result.isCurrent)

			mockManager.AssertExpectations(t)
			mockGit.AssertExpectations(t)
		})
	}
}

func TestCloseDialog_SetRequireReason(t *testing.T) {
	// This test verifies the SetRequireReason method works correctly
	// to prevent race conditions when updating requirements asynchronously

	// Import the components package to test dialog directly
	dialog := components.NewCloseDialogModel()

	// Show dialog with reason required initially (safer default)
	dialog.Show(true)
	assert.True(t, dialog.IsVisible(), "Dialog should be visible after Show")

	// Update requirement without hiding dialog
	dialog.SetRequireReason(false)
	assert.True(t, dialog.IsVisible(), "Dialog should remain visible after SetRequireReason")

	// Update requirement again
	dialog.SetRequireReason(true)
	assert.True(t, dialog.IsVisible(), "Dialog should still be visible after second SetRequireReason")

	// Hide dialog
	dialog.Hide()
	assert.False(t, dialog.IsVisible(), "Dialog should be hidden after Hide")
}
