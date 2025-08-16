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

// Make CleanupResult implement Printable
func (r *CleanupResult) TextRepresentation() string {
	var buf strings.Builder
	fmt.Fprintf(&buf, "\nCleanup Summary:\n")
	fmt.Fprintf(&buf, "  Orphaned worktrees removed: %d\n", r.OrphanedWorktrees)
	fmt.Fprintf(&buf, "  Stale branches removed: %d\n", r.StaleBranches)
	
	if r.HasErrors() {
		fmt.Fprintf(&buf, "\nErrors encountered:\n")
		for _, err := range r.Errors {
			fmt.Fprintf(&buf, "  - %s\n", err)
		}
	}
	
	return buf.String()
}

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
	
	// Header
	fmt.Fprintf(&buf, "ID       STATUS  PRI  DESCRIPTION\n")
	fmt.Fprintf(&buf, "---------------------------------------------------------\n")
	
	// Tickets
	for _, t := range r.Tickets {
		status := "todo"
		if t.StartedAt.Time != nil {
			status = "doing"
		}
		if t.ClosedAt.Time != nil {
			status = "done"
		}
		
		fmt.Fprintf(&buf, "%-8s %-7s %-4d %s\n", 
			t.ID[:min(8, len(t.ID))], 
			status, 
			t.Priority, 
			t.Description)
	}
	
	// Summary
	if r.Count != nil {
		fmt.Fprintf(&buf, "\nSummary: %d todo, %d doing, %d done (Total: %d)\n",
			r.Count["todo"], r.Count["doing"], r.Count["done"], r.Count["total"])
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

// Helper function (you probably already have this)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}