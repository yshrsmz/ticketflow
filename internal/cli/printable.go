package cli

import (
	"fmt"
	"strings"
	"time"

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
	_ Printable = (*TicketResult)(nil)
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

	// Find max ID length for alignment
	maxIDLen := 0
	for _, t := range r.Tickets {
		if len(t.ID) > maxIDLen {
			maxIDLen = len(t.ID)
		}
	}
	// Minimum width for ID column
	if maxIDLen < 2 {
		maxIDLen = 2
	}

	// Header
	buf.WriteString(fmt.Sprintf("%-*s  %-6s  %-3s  %s\n", maxIDLen, "ID", "STATUS", "PRI", "DESCRIPTION"))
	buf.WriteString(strings.Repeat("-", maxIDLen+50))
	buf.WriteString("\n")

	// Tickets
	for _, t := range r.Tickets {
		status := getTicketStatus(&t)

		// Truncate description if too long
		desc := t.Description
		maxDescLen := 50
		if len(desc) > maxDescLen {
			desc = desc[:maxDescLen-3] + "..."
		}

		buf.WriteString(fmt.Sprintf("%-*s  %-6s  %-3d  %s\n",
			maxIDLen,
			t.ID,
			status,
			t.Priority,
			desc))
	}

	return buf.String()
}

// getTicketStatus determines the status of a ticket based on its time fields
func getTicketStatus(t *ticket.Ticket) string {
	if isTimeSet(t.ClosedAt.Time) {
		return "done"
	}
	if isTimeSet(t.StartedAt.Time) {
		return "doing"
	}
	return "todo"
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

// TicketResult wraps a single ticket to make it Printable
type TicketResult struct {
	Ticket *ticket.Ticket
}

// TextRepresentation returns human-readable format for a single ticket
func (r *TicketResult) TextRepresentation() string {
	if r.Ticket == nil {
		return "No ticket found\n"
	}

	var buf strings.Builder
	buf.Grow(512)

	t := r.Ticket
	buf.WriteString(fmt.Sprintf("ID: %s\n", t.ID))
	buf.WriteString(fmt.Sprintf("Status: %s\n", t.Status()))
	buf.WriteString(fmt.Sprintf("Priority: %d\n", t.Priority))
	buf.WriteString(fmt.Sprintf("Description: %s\n", t.Description))
	buf.WriteString(fmt.Sprintf("Created: %s\n", t.CreatedAt.Format(time.RFC3339)))

	if t.StartedAt.Time != nil {
		buf.WriteString(fmt.Sprintf("Started: %s\n", t.StartedAt.Time.Format(time.RFC3339)))
	}

	if t.ClosedAt.Time != nil {
		buf.WriteString(fmt.Sprintf("Closed: %s\n", t.ClosedAt.Time.Format(time.RFC3339)))
	}

	if len(t.Related) > 0 {
		buf.WriteString(fmt.Sprintf("Related: %s\n", strings.Join(t.Related, ", ")))
	}

	buf.WriteString(fmt.Sprintf("\n%s\n", t.Content))

	return buf.String()
}

// StructuredData returns the ticket for JSON serialization
func (r *TicketResult) StructuredData() interface{} {
	if r.Ticket == nil {
		return nil
	}

	return map[string]interface{}{
		"ticket": map[string]interface{}{
			"id":          r.Ticket.ID,
			"path":        r.Ticket.Path,
			"status":      string(r.Ticket.Status()),
			"priority":    r.Ticket.Priority,
			"description": r.Ticket.Description,
			"created_at":  r.Ticket.CreatedAt.Time,
			"started_at":  r.Ticket.StartedAt.Time,
			"closed_at":   r.Ticket.ClosedAt.Time,
			"related":     r.Ticket.Related,
			"content":     r.Ticket.Content,
		},
	}
}
