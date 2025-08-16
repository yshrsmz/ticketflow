package cli

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		
		// Check header
		assert.Contains(t, text, "ID       STATUS  PRI  DESCRIPTION")
		assert.Contains(t, text, "---")
		
		// Check tickets
		assert.Contains(t, text, "ticket-1")  // ID truncated to 8 chars
		assert.Contains(t, text, "Test ticket")
		assert.Contains(t, text, "very-lon")  // Long ID truncated
		assert.Contains(t, text, "Another ticket")
		
		// Check summary
		assert.Contains(t, text, "Summary: 1 todo, 1 doing, 0 done (Total: 2)")
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
		assert.Contains(t, text, "todo")  // Should default to todo status
		assert.Contains(t, text, "Ticket with nil times")
		assert.NotContains(t, text, "Summary")  // No summary if Count is nil
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
		assert.Contains(t, text, "Summary: 50 todo, 30 doing, 20 done (Total: 100)")
	})
}