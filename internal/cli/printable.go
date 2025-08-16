package cli

import (
	"fmt"
	"strings"

	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// Printable represents a result that knows how to format itself
type Printable interface {
	// TextRepresentation returns human-readable format
	TextRepresentation() string
	// StructuredData returns data for JSON/YAML serialization
	StructuredData() interface{}
}

// Verify interface compliance at compile time
var (
	_ Printable = (*CleanupResult)(nil)
	_ Printable = (*TicketListResult)(nil)
)

// TextRepresentation returns human-readable format for CleanupResult
func (r *CleanupResult) TextRepresentation() string {
	var buf strings.Builder
	// Pre-allocate capacity for better performance
	buf.Grow(256)

	// strings.Builder.Write methods never return errors, so we can safely ignore them
	buf.WriteString("\nCleanup Summary:\n")
	buf.WriteString("  Orphaned worktrees removed: ")
	buf.WriteString(fmt.Sprintf("%d", r.OrphanedWorktrees))
	buf.WriteByte('\n')
	buf.WriteString("  Stale branches removed: ")
	buf.WriteString(fmt.Sprintf("%d", r.StaleBranches))
	buf.WriteByte('\n')

	if r.HasErrors() {
		buf.WriteString("\nErrors encountered:\n")
		for _, err := range r.Errors {
			buf.WriteString("  - ")
			buf.WriteString(err)
			buf.WriteByte('\n')
		}
	}

	return buf.String()
}

// StructuredData returns data for JSON serialization
func (r *CleanupResult) StructuredData() interface{} {
	return map[string]interface{}{
		"orphaned_worktrees": r.OrphanedWorktrees,
		"stale_branches":     r.StaleBranches,
		"errors":             r.Errors,
		"has_errors":         r.HasErrors(),
	}
}

// TicketListResult wraps ticket list to make it Printable
type TicketListResult struct {
	Tickets []ticket.Ticket
	Count   map[string]int
}

func (r *TicketListResult) TextRepresentation() string {
	var buf strings.Builder

	if len(r.Tickets) == 0 {
		return "No tickets found\n"
	}

	// Pre-allocate capacity
	buf.Grow(512)

	// Header
	buf.WriteString("ID       STATUS  PRI  DESCRIPTION\n")
	buf.WriteString("---------------------------------------------------------\n")

	// Tickets
	for _, t := range r.Tickets {
		status := "todo"
		if t.StartedAt.Time != nil && !t.StartedAt.Time.IsZero() {
			status = "doing"
		}
		if t.ClosedAt.Time != nil && !t.ClosedAt.Time.IsZero() {
			status = "done"
		}

		// Safely truncate ID
		idDisplay := t.ID
		if len(idDisplay) > 8 {
			idDisplay = idDisplay[:8]
		}

		buf.WriteString(fmt.Sprintf("%-8s %-7s %-4d %s\n",
			idDisplay,
			status,
			t.Priority,
			t.Description))
	}

	// Summary
	if r.Count != nil {
		buf.WriteString(fmt.Sprintf("\nSummary: %d todo, %d doing, %d done (Total: %d)\n",
			r.Count["todo"], r.Count["doing"], r.Count["done"], r.Count["total"]))
	}

	return buf.String()
}

func (r *TicketListResult) StructuredData() interface{} {
	tickets := make([]map[string]interface{}, len(r.Tickets))
	for i, t := range r.Tickets {
		tickets[i] = ticketToJSON(&t, "")
	}

	return map[string]interface{}{
		"tickets": tickets,
		"summary": r.Count,
	}
}
