package cli

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestCleanupResultPrintable(t *testing.T) {
	t.Parallel()

	t.Run("TextRepresentation without errors", func(t *testing.T) {
		result := &CleanupResult{
			OrphanedWorktrees: 3,
			StaleBranches:     2,
			Errors:            []string{},
		}

		text := result.TextRepresentation()
		assert.Contains(t, text, "Cleanup Summary")
		assert.Contains(t, text, "Orphaned worktrees removed: 3")
		assert.Contains(t, text, "Stale branches removed: 2")
		assert.NotContains(t, text, "Errors encountered")
	})

	t.Run("TextRepresentation with errors", func(t *testing.T) {
		result := &CleanupResult{
			OrphanedWorktrees: 1,
			StaleBranches:     0,
			Errors:            []string{"error1", "error2"},
		}

		text := result.TextRepresentation()
		assert.Contains(t, text, "Cleanup Summary")
		assert.Contains(t, text, "Orphaned worktrees removed: 1")
		assert.Contains(t, text, "Stale branches removed: 0")
		assert.Contains(t, text, "Errors encountered")
		assert.Contains(t, text, "error1")
		assert.Contains(t, text, "error2")
	})

	t.Run("StructuredData", func(t *testing.T) {
		result := &CleanupResult{
			OrphanedWorktrees: 5,
			StaleBranches:     3,
			Errors:            []string{"warning"},
		}

		data := result.StructuredData().(map[string]interface{})
		assert.Equal(t, 5, data["orphaned_worktrees"])
		assert.Equal(t, 3, data["stale_branches"])
		assert.Equal(t, []string{"warning"}, data["errors"])
		assert.Equal(t, true, data["has_errors"])
	})

	t.Run("StructuredData JSON serialization", func(t *testing.T) {
		result := &CleanupResult{
			OrphanedWorktrees: 2,
			StaleBranches:     1,
			Errors:            []string{},
		}

		data := result.StructuredData()
		jsonBytes, err := json.Marshal(data)
		require.NoError(t, err)

		var parsed map[string]interface{}
		err = json.Unmarshal(jsonBytes, &parsed)
		require.NoError(t, err)

		assert.Equal(t, float64(2), parsed["orphaned_worktrees"])
		assert.Equal(t, float64(1), parsed["stale_branches"])
		assert.Equal(t, false, parsed["has_errors"])
	})
}

func TestTicketListResultPrintable(t *testing.T) {
	t.Parallel()

	t.Run("TextRepresentation with empty list", func(t *testing.T) {
		result := &TicketListResult{
			Tickets: []ticket.Ticket{},
			Count:   nil,
		}

		text := result.TextRepresentation()
		assert.Equal(t, "No tickets found\n", text)
	})

	t.Run("TextRepresentation with tickets", func(t *testing.T) {
		now := time.Now()
		result := &TicketListResult{
			Tickets: []ticket.Ticket{
				{
					ID:          "ticket-123",
					Priority:    2,
					Description: "Test ticket",
					CreatedAt:   ticket.NewRFC3339Time(now),
				},
				{
					ID:          "very-long-ticket-id-that-should-be-truncated",
					Priority:    1,
					Description: "Another ticket",
					StartedAt:   ticket.NewRFC3339TimePtr(&now),
				},
			},
			Count: map[string]int{
				"todo":  1,
				"doing": 1,
				"done":  0,
				"total": 2,
			},
		}

		text := result.TextRepresentation()

		// Check header with dynamic width
		assert.Contains(t, text, "ID")
		assert.Contains(t, text, "STATUS")
		assert.Contains(t, text, "PRI")
		assert.Contains(t, text, "DESCRIPTION")
		assert.Contains(t, text, "---")

		// Check tickets - IDs are NOT truncated with dynamic width
		assert.Contains(t, text, "ticket-123")
		assert.Contains(t, text, "Test ticket")
		assert.Contains(t, text, "very-long-ticket-id-that-should-be-truncated")
		assert.Contains(t, text, "Another ticket")

		// Summary is no longer included in TextRepresentation
	})

	t.Run("TextRepresentation with nil time fields", func(t *testing.T) {
		result := &TicketListResult{
			Tickets: []ticket.Ticket{
				{
					ID:          "test-1",
					Priority:    3,
					Description: "Ticket with nil times",
					CreatedAt:   ticket.RFC3339Time{Time: time.Time{}},
					StartedAt:   ticket.RFC3339TimePtr{Time: nil},
					ClosedAt:    ticket.RFC3339TimePtr{Time: nil},
				},
			},
			Count: nil,
		}

		text := result.TextRepresentation()
		assert.Contains(t, text, "test-1")
		assert.Contains(t, text, "todo") // Should default to todo status
		assert.Contains(t, text, "Ticket with nil times")
		assert.NotContains(t, text, "Summary") // No summary if Count is nil
	})

	t.Run("TextRepresentation with done ticket", func(t *testing.T) {
		now := time.Now()
		result := &TicketListResult{
			Tickets: []ticket.Ticket{
				{
					ID:          "done-1",
					Priority:    1,
					Description: "Completed ticket",
					ClosedAt:    ticket.NewRFC3339TimePtr(&now),
				},
			},
			Count: nil,
		}

		text := result.TextRepresentation()
		assert.Contains(t, text, "done-1")
		assert.Contains(t, text, "done")
		assert.Contains(t, text, "Completed ticket")
	})

	t.Run("StructuredData", func(t *testing.T) {
		now := time.Now()
		result := &TicketListResult{
			Tickets: []ticket.Ticket{
				{
					ID:          "test-123",
					Priority:    2,
					Description: "Test ticket",
					CreatedAt:   ticket.NewRFC3339Time(now),
				},
			},
			Count: map[string]int{
				"todo":  1,
				"doing": 0,
				"done":  0,
				"total": 1,
			},
		}

		data := result.StructuredData().(map[string]interface{})

		tickets := data["tickets"].([]map[string]interface{})
		assert.Len(t, tickets, 1)
		assert.Equal(t, "test-123", tickets[0]["id"])
		assert.Equal(t, 2, tickets[0]["priority"])

		summary := data["summary"].(map[string]int)
		assert.Equal(t, 1, summary["todo"])
		assert.Equal(t, 0, summary["doing"])
		assert.Equal(t, 1, summary["total"])
	})

	t.Run("StructuredData JSON serialization", func(t *testing.T) {
		result := &TicketListResult{
			Tickets: []ticket.Ticket{
				{
					ID:          "json-test",
					Priority:    1,
					Description: "JSON test ticket",
				},
			},
			Count: map[string]int{
				"todo":  1,
				"doing": 0,
				"done":  0,
				"total": 1,
			},
		}

		data := result.StructuredData()
		jsonBytes, err := json.Marshal(data)
		require.NoError(t, err)

		var parsed map[string]interface{}
		err = json.Unmarshal(jsonBytes, &parsed)
		require.NoError(t, err)

		tickets := parsed["tickets"].([]interface{})
		assert.Len(t, tickets, 1)

		firstTicket := tickets[0].(map[string]interface{})
		assert.Equal(t, "json-test", firstTicket["id"])
		assert.Equal(t, "JSON test ticket", firstTicket["description"])
	})
}

func TestPrintableInterfaceCompliance(t *testing.T) {
	t.Parallel()

	// This test verifies that our types implement the Printable interface
	// The compile-time checks in printable.go ensure this, but we can also
	// verify at runtime

	t.Run("CleanupResult implements Printable", func(t *testing.T) {
		var p Printable = &CleanupResult{}
		assert.NotNil(t, p)

		// Should be able to call interface methods
		_ = p.TextRepresentation()
		_ = p.StructuredData()
	})

	t.Run("TicketListResult implements Printable", func(t *testing.T) {
		var p Printable = &TicketListResult{}
		assert.NotNil(t, p)

		// Should be able to call interface methods
		_ = p.TextRepresentation()
		_ = p.StructuredData()
	})
}

func TestTextRepresentationPerformance(t *testing.T) {
	// Test that TextRepresentation uses strings.Builder efficiently

	t.Run("CleanupResult with many errors", func(t *testing.T) {
		errors := make([]string, 100)
		for i := 0; i < 100; i++ {
			errors[i] = strings.Repeat("error", 20) // Long error messages
		}

		result := &CleanupResult{
			OrphanedWorktrees: 10,
			StaleBranches:     20,
			Errors:            errors,
		}

		// Should not panic or have performance issues
		text := result.TextRepresentation()
		assert.NotEmpty(t, text)
		assert.Contains(t, text, "Errors encountered")

		// Verify all errors are included
		for _, err := range errors {
			assert.Contains(t, text, err)
		}
	})

	t.Run("TicketListResult with many tickets", func(t *testing.T) {
		tickets := make([]ticket.Ticket, 100)
		for i := 0; i < 100; i++ {
			tickets[i] = ticket.Ticket{
				ID:          strings.Repeat("a", 50), // Long ID
				Priority:    i % 5,
				Description: strings.Repeat("desc", 25), // Long description
			}
		}

		result := &TicketListResult{
			Tickets: tickets,
			Count: map[string]int{
				"todo":  50,
				"doing": 30,
				"done":  20,
				"total": 100,
			},
		}

		// Should not panic or have performance issues
		text := result.TextRepresentation()
		assert.NotEmpty(t, text)
		// Summary is no longer included in TextRepresentation
		assert.Contains(t, text, "ID")
		assert.Contains(t, text, "STATUS")
	})
}

func TestTicketResultPrintable(t *testing.T) {
	t.Parallel()

	t.Run("TextRepresentation with complete ticket", func(t *testing.T) {
		now := time.Now()
		startTime := now.Add(-2 * time.Hour)
		closeTime := now.Add(-30 * time.Minute)

		result := &TicketResult{
			Ticket: &ticket.Ticket{
				ID:          "test-123",
				Priority:    2,
				Description: "Test ticket description",
				CreatedAt:   ticket.NewRFC3339Time(now),
				StartedAt:   ticket.RFC3339TimePtr{Time: &startTime},
				ClosedAt:    ticket.RFC3339TimePtr{Time: &closeTime},
				Related:     []string{"parent:test-parent", "blocks:test-blocked"},
				Content:     "# Test Content\n\nThis is the ticket content.",
			},
		}

		text := result.TextRepresentation()
		assert.Contains(t, text, "ID: test-123")
		assert.Contains(t, text, "Priority: 2")
		assert.Contains(t, text, "Description: Test ticket description")
		assert.Contains(t, text, "Created:")
		assert.Contains(t, text, "Started:")
		assert.Contains(t, text, "Closed:")
		assert.Contains(t, text, "Related: parent:test-parent, blocks:test-blocked")
		assert.Contains(t, text, "# Test Content")
	})

	t.Run("TextRepresentation with nil ticket", func(t *testing.T) {
		result := &TicketResult{
			Ticket: nil,
		}

		text := result.TextRepresentation()
		assert.Equal(t, "No ticket found\n", text)
	})

	t.Run("StructuredData", func(t *testing.T) {
		result := &TicketResult{
			Ticket: &ticket.Ticket{
				ID:          "test-456",
				Priority:    1,
				Description: "Another test",
			},
		}

		data := result.StructuredData()
		m, ok := data.(map[string]interface{})
		assert.True(t, ok)

		ticketData, ok := m["ticket"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "test-456", ticketData["id"])
		assert.Equal(t, 1, ticketData["priority"])
	})
}

func TestWorktreeListResultPrintable(t *testing.T) {
	t.Parallel()

	t.Run("TextRepresentation with worktrees", func(t *testing.T) {
		result := &WorktreeListResult{
			Worktrees: []git.WorktreeInfo{
				{
					Path:   "/path/to/worktree1",
					Branch: "feature-1",
					HEAD:   "abc123def456789012345678901234567890123",
				},
				{
					Path:   "/path/to/worktree2",
					Branch: "feature-2",
					HEAD:   "short",
				},
			},
		}

		text := result.TextRepresentation()
		assert.Contains(t, text, "PATH")
		assert.Contains(t, text, "BRANCH")
		assert.Contains(t, text, "HEAD")
		assert.Contains(t, text, "/path/to/worktree1")
		assert.Contains(t, text, "feature-1")
		assert.Contains(t, text, "abc123d") // Should be truncated to 7 chars
		assert.Contains(t, text, "feature-2")
		assert.Contains(t, text, "short") // Short hash not truncated
	})

	t.Run("TextRepresentation with empty list", func(t *testing.T) {
		result := &WorktreeListResult{
			Worktrees: []git.WorktreeInfo{},
		}

		text := result.TextRepresentation()
		assert.Equal(t, "No worktrees found\n", text)
	})

	t.Run("StructuredData", func(t *testing.T) {
		result := &WorktreeListResult{
			Worktrees: []git.WorktreeInfo{
				{Path: "/test", Branch: "main", HEAD: "abc"},
			},
		}

		data := result.StructuredData()
		m, ok := data.(map[string]interface{})
		assert.True(t, ok)
		assert.Contains(t, m, "worktrees")
	})
}

func TestStatusResultPrintable(t *testing.T) {
	t.Parallel()

	t.Run("TextRepresentation with active ticket", func(t *testing.T) {
		startTime := time.Now().Add(-1 * time.Hour)
		result := &StatusResult{
			CurrentBranch: "feature-branch",
			CurrentTicket: &ticket.Ticket{
				ID:          "active-123",
				Description: "Active ticket",
				StartedAt:   ticket.RFC3339TimePtr{Time: &startTime},
			},
			WorktreePath: "/path/to/worktree",
			Summary: map[string]int{
				"todo":  5,
				"doing": 2,
				"done":  10,
			},
			TotalTickets: 17,
		}

		text := result.TextRepresentation()
		assert.Contains(t, text, "Current branch: feature-branch")
		assert.Contains(t, text, "Active ticket: active-123")
		assert.Contains(t, text, "Description: Active ticket")
		assert.Contains(t, text, "Duration:")
		assert.Contains(t, text, "Worktree: /path/to/worktree")
		assert.Contains(t, text, "Todo:  5")
		assert.Contains(t, text, "Doing: 2")
		assert.Contains(t, text, "Done:  10")
		assert.Contains(t, text, "Total: 17")
	})

	t.Run("TextRepresentation without active ticket", func(t *testing.T) {
		result := &StatusResult{
			CurrentBranch: "main",
			CurrentTicket: nil,
			Summary: map[string]int{
				"todo":  3,
				"doing": 0,
				"done":  7,
			},
			TotalTickets: 10,
		}

		text := result.TextRepresentation()
		assert.Contains(t, text, "No active ticket")
		assert.Contains(t, text, "Start a ticket with: ticketflow start")
	})

	t.Run("StructuredData", func(t *testing.T) {
		result := &StatusResult{
			CurrentBranch: "test",
			CurrentTicket: &ticket.Ticket{ID: "test-123"},
			Summary:       map[string]int{"todo": 1},
		}

		data := result.StructuredData()
		m, ok := data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "test", m["current_branch"])
		assert.Contains(t, m, "summary")
		assert.NotNil(t, m["current_ticket"])
	})
}

func TestStartResultPrintable(t *testing.T) {
	t.Parallel()

	t.Run("TextRepresentation with worktree enabled", func(t *testing.T) {
		result := &StartResult{
			StartTicketResult: &StartTicketResult{
				Ticket: &ticket.Ticket{
					ID:          "new-feature",
					Description: "New feature implementation",
				},
				WorktreePath:         "/worktrees/new-feature",
				ParentBranch:         "main",
				InitCommandsExecuted: true,
			},
			WorktreeEnabled: true,
		}

		text := result.TextRepresentation()
		assert.Contains(t, text, "Started work on ticket: new-feature")
		assert.Contains(t, text, "Description: New feature implementation")
		assert.Contains(t, text, "Worktree created: /worktrees/new-feature")
		assert.Contains(t, text, "Parent ticket: main")
		assert.Contains(t, text, "Navigate to worktree:")
		assert.Contains(t, text, "cd /worktrees/new-feature")
	})

	t.Run("TextRepresentation with worktree disabled", func(t *testing.T) {
		result := &StartResult{
			StartTicketResult: &StartTicketResult{
				Ticket: &ticket.Ticket{
					ID:          "branch-feature",
					Description: "Branch mode feature",
				},
			},
			WorktreeEnabled: false,
		}

		text := result.TextRepresentation()
		assert.Contains(t, text, "Switched to branch: branch-feature")
		assert.NotContains(t, text, "Worktree created")
		assert.NotContains(t, text, "Navigate to worktree")
	})

	t.Run("StructuredData", func(t *testing.T) {
		result := &StartResult{
			StartTicketResult: &StartTicketResult{
				Ticket:               &ticket.Ticket{ID: "test-123"},
				WorktreePath:         "/test/path",
				ParentBranch:         "develop",
				InitCommandsExecuted: true,
			},
		}

		data := result.StructuredData()
		m, ok := data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "test-123", m["ticket_id"])
		assert.Equal(t, "/test/path", m["worktree_path"])
		assert.Equal(t, "develop", m["parent_branch"])
		assert.Equal(t, true, m["init_commands_executed"])
	})
}

func TestNewTicketResult_TextRepresentation(t *testing.T) {
	tests := []struct {
		name     string
		result   *NewTicketResult
		contains []string
	}{
		{
			name: "simple ticket creation",
			result: &NewTicketResult{
				Ticket: &ticket.Ticket{
					ID:   "240101-123456-feature",
					Path: "tickets/todo/240101-123456-feature.md",
				},
			},
			contains: []string{
				"🎫 Created new ticket: 240101-123456-feature",
				"File: tickets/todo/240101-123456-feature.md",
				"Next steps:",
				"$EDITOR tickets/todo/240101-123456-feature.md",
				"git add tickets/todo/240101-123456-feature.md",
				"git commit -m \"Add ticket: 240101-123456-feature\"",
				"ticketflow start 240101-123456-feature",
			},
		},
		{
			name: "sub-ticket with parent",
			result: &NewTicketResult{
				Ticket: &ticket.Ticket{
					ID:   "240101-123456-subfeature",
					Path: "tickets/todo/240101-123456-subfeature.md",
				},
				ParentTicket: "240101-100000-parent",
			},
			contains: []string{
				"🎫 Created new ticket: 240101-123456-subfeature",
				"Parent ticket: 240101-100000-parent",
				"Type: Sub-ticket",
			},
		},
		{
			name: "nil ticket",
			result: &NewTicketResult{
				Ticket: nil,
			},
			contains: []string{
				"Error: No ticket created",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.result.TextRepresentation()
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestNewTicketResult_StructuredData(t *testing.T) {
	tests := []struct {
		name   string
		result *NewTicketResult
		verify func(t *testing.T, data interface{})
	}{
		{
			name: "simple ticket",
			result: &NewTicketResult{
				Ticket: &ticket.Ticket{
					ID:   "240101-123456-feature",
					Path: "tickets/todo/240101-123456-feature.md",
				},
			},
			verify: func(t *testing.T, data interface{}) {
				m, ok := data.(map[string]interface{})
				require.True(t, ok)

				ticketData, ok := m["ticket"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "240101-123456-feature", ticketData["id"])
				assert.Equal(t, "tickets/todo/240101-123456-feature.md", ticketData["path"])

				_, hasParent := m["parent_ticket"]
				assert.False(t, hasParent)
			},
		},
		{
			name: "sub-ticket with parent",
			result: &NewTicketResult{
				Ticket: &ticket.Ticket{
					ID:   "240101-123456-subfeature",
					Path: "tickets/todo/240101-123456-subfeature.md",
				},
				ParentTicket: "240101-100000-parent",
			},
			verify: func(t *testing.T, data interface{}) {
				m, ok := data.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "240101-100000-parent", m["parent_ticket"])
			},
		},
		{
			name: "nil ticket",
			result: &NewTicketResult{
				Ticket: nil,
			},
			verify: func(t *testing.T, data interface{}) {
				assert.Nil(t, data)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := tt.result.StructuredData()
			tt.verify(t, data)
		})
	}
}

func TestCloseTicketResult_TextRepresentation(t *testing.T) {
	closedAt := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)
	startedAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		result   *CloseTicketResult
		contains []string
	}{
		{
			name: "close current ticket with duration",
			result: &CloseTicketResult{
				Ticket: &ticket.Ticket{
					ID:        "240101-123456-feature",
					StartedAt: ticket.RFC3339TimePtr{Time: &startedAt},
					ClosedAt:  ticket.RFC3339TimePtr{Time: &closedAt},
				},
				Mode:          "current",
				Duration:      2*time.Hour + 30*time.Minute,
				CommitCreated: true,
			},
			contains: []string{
				"✅ Closed current ticket: 240101-123456-feature",
				"Duration: 2h 30m",
				"Status: doing → done",
				"Committed: \"Close ticket: 240101-123456-feature\"",
				"Push your branch to create/update PR",
				"git push",
				"ticketflow cleanup 240101-123456-feature",
			},
		},
		{
			name: "close by ID with force and reason",
			result: &CloseTicketResult{
				Ticket: &ticket.Ticket{
					ID:       "240101-123456-feature",
					ClosedAt: ticket.RFC3339TimePtr{Time: &closedAt},
				},
				Mode:          "by_id",
				ForceUsed:     true,
				CloseReason:   "Task completed",
				CommitCreated: true,
				Branch:        "240101-123456-feature",
			},
			contains: []string{
				"✅ Closed ticket: 240101-123456-feature",
				"⚠️  Force flag used to bypass validation",
				"Reason: Task completed",
				"Committed: \"Close ticket: 240101-123456-feature - Task completed\"",
				"git push origin 240101-123456-feature",
			},
		},
		{
			name: "close with parent ticket",
			result: &CloseTicketResult{
				Ticket: &ticket.Ticket{
					ID: "240101-123456-subfeature",
				},
				Mode:         "current",
				ParentTicket: "240101-100000-parent",
			},
			contains: []string{
				"Parent ticket: 240101-100000-parent",
			},
		},
		{
			name: "nil ticket",
			result: &CloseTicketResult{
				Ticket: nil,
			},
			contains: []string{
				"Error: No ticket to close",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.result.TextRepresentation()
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestCloseTicketResult_StructuredData(t *testing.T) {
	closedAt := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name   string
		result *CloseTicketResult
		verify func(t *testing.T, data interface{})
	}{
		{
			name: "current ticket with all fields",
			result: &CloseTicketResult{
				Ticket: &ticket.Ticket{
					ID:       "240101-123456-feature",
					ClosedAt: ticket.RFC3339TimePtr{Time: &closedAt},
				},
				Mode:          "current",
				ForceUsed:     true,
				CommitCreated: true,
				Duration:      2*time.Hour + 30*time.Minute,
				ParentTicket:  "parent-123",
				WorktreePath:  "/path/to/worktree",
				CloseReason:   "Completed",
			},
			verify: func(t *testing.T, data interface{}) {
				m, ok := data.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, true, m["success"])
				assert.Equal(t, "240101-123456-feature", m["ticket_id"])
				assert.Equal(t, "done", m["status"])
				assert.Equal(t, "current", m["mode"])
				assert.Equal(t, true, m["force_used"])
				assert.Equal(t, true, m["commit_created"])
				assert.Equal(t, "2h30m", m["duration"])
				assert.Equal(t, "parent-123", m["parent_ticket"])
				assert.Equal(t, "/path/to/worktree", m["worktree_path"])
				assert.Equal(t, "Completed", m["close_reason"])
				assert.Equal(t, closedAt.Format(time.RFC3339), m["closed_at"])
			},
		},
		{
			name: "by_id mode with branch",
			result: &CloseTicketResult{
				Ticket: &ticket.Ticket{
					ID: "240101-123456-feature",
				},
				Mode:   "by_id",
				Branch: "feature-branch",
			},
			verify: func(t *testing.T, data interface{}) {
				m, ok := data.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "by_id", m["mode"])
				assert.Equal(t, "feature-branch", m["branch"])
				_, hasWorktree := m["worktree_path"]
				assert.False(t, hasWorktree)
			},
		},
		{
			name: "nil ticket",
			result: &CloseTicketResult{
				Ticket: nil,
			},
			verify: func(t *testing.T, data interface{}) {
				assert.Nil(t, data)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := tt.result.StructuredData()
			tt.verify(t, data)
		})
	}
}

func TestRestoreTicketResult_TextRepresentation(t *testing.T) {
	result := &RestoreTicketResult{
		Ticket: &ticket.Ticket{
			ID: "240101-123456-feature",
		},
	}

	output := result.TextRepresentation()
	assert.Equal(t, "✅ Current ticket symlink restored\n", output)
}

func TestRestoreTicketResult_StructuredData(t *testing.T) {
	tests := []struct {
		name   string
		result *RestoreTicketResult
		verify func(t *testing.T, data interface{})
	}{
		{
			name: "full restore result",
			result: &RestoreTicketResult{
				Ticket: &ticket.Ticket{
					ID: "240101-123456-feature",
				},
				SymlinkPath:  "current-ticket.md",
				TargetPath:   "tickets/doing/240101-123456-feature.md",
				ParentTicket: "parent-123",
				WorktreePath: "/path/to/worktree",
			},
			verify: func(t *testing.T, data interface{}) {
				m, ok := data.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, true, m["success"])
				assert.Equal(t, "240101-123456-feature", m["ticket_id"])
				assert.Equal(t, "doing", m["status"])
				assert.Equal(t, true, m["symlink_restored"])
				assert.Equal(t, "current-ticket.md", m["symlink_path"])
				assert.Equal(t, "tickets/doing/240101-123456-feature.md", m["target_path"])
				assert.Equal(t, "Current ticket symlink restored", m["message"])
				assert.Equal(t, "parent-123", m["parent_ticket"])
				assert.Equal(t, "/path/to/worktree", m["worktree_path"])
			},
		},
		{
			name: "minimal restore result",
			result: &RestoreTicketResult{
				Ticket: &ticket.Ticket{
					ID: "240101-123456-feature",
				},
				SymlinkPath: "current-ticket.md",
				TargetPath:  "tickets/doing/240101-123456-feature.md",
			},
			verify: func(t *testing.T, data interface{}) {
				m, ok := data.(map[string]interface{})
				require.True(t, ok)
				_, hasParent := m["parent_ticket"]
				assert.False(t, hasParent)
				_, hasWorktree := m["worktree_path"]
				assert.False(t, hasWorktree)
			},
		},
		{
			name: "nil ticket",
			result: &RestoreTicketResult{
				Ticket: nil,
			},
			verify: func(t *testing.T, data interface{}) {
				assert.Nil(t, data)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := tt.result.StructuredData()
			tt.verify(t, data)
		})
	}
}
