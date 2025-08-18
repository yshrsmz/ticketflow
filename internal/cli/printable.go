package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

const (
	// GitSHAFullLength is the length of a full git SHA-1 hash
	GitSHAFullLength = 40
	// GitSHAShortLength is the standard length for abbreviated git commit hashes
	GitSHAShortLength = 7
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
	_ Printable = (*WorktreeListResult)(nil)
	_ Printable = (*StatusResult)(nil)
	_ Printable = (*StartResult)(nil)
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
	fmt.Fprintf(&buf, "%-*s  %-6s  %-3s  %s\n", maxIDLen, "ID", "STATUS", "PRI", "DESCRIPTION")
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

		fmt.Fprintf(&buf, "%-*s  %-6s  %-3d  %s\n",
			maxIDLen,
			t.ID,
			status,
			t.Priority,
			desc)
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
	fmt.Fprintf(&buf, "ID: %s\n", t.ID)
	fmt.Fprintf(&buf, "Status: %s\n", t.Status())
	fmt.Fprintf(&buf, "Priority: %d\n", t.Priority)
	fmt.Fprintf(&buf, "Description: %s\n", t.Description)
	fmt.Fprintf(&buf, "Created: %s\n", t.CreatedAt.Format(time.RFC3339))

	if t.StartedAt.Time != nil {
		fmt.Fprintf(&buf, "Started: %s\n", t.StartedAt.Time.Format(time.RFC3339))
	}

	if t.ClosedAt.Time != nil {
		fmt.Fprintf(&buf, "Closed: %s\n", t.ClosedAt.Time.Format(time.RFC3339))
	}

	if len(t.Related) > 0 {
		fmt.Fprintf(&buf, "Related: %s\n", strings.Join(t.Related, ", "))
	}

	fmt.Fprintf(&buf, "\n%s\n", t.Content)

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

// WorktreeListResult wraps worktree list to make it Printable
type WorktreeListResult struct {
	Worktrees []git.WorktreeInfo
}

// TextRepresentation returns human-readable format for worktree list
func (r *WorktreeListResult) TextRepresentation() string {
	if len(r.Worktrees) == 0 {
		return "No worktrees found\n"
	}

	var buf strings.Builder
	buf.Grow(512)

	// Header
	fmt.Fprintf(&buf, "%-50s %-30s %s\n", "PATH", "BRANCH", "HEAD")
	buf.WriteString(strings.Repeat("-", 100))
	buf.WriteString("\n")

	// Worktrees
	for _, wt := range r.Worktrees {
		head := wt.HEAD
		if len(head) > GitSHAFullLength {
			head = head[:GitSHAShortLength] // Short commit hash
		}
		fmt.Fprintf(&buf, "%-50s %-30s %s\n", wt.Path, wt.Branch, head)
	}

	return buf.String()
}

// StructuredData returns worktrees for JSON serialization
func (r *WorktreeListResult) StructuredData() interface{} {
	return map[string]interface{}{
		"worktrees": r.Worktrees,
	}
}

// StatusResult wraps status information to make it Printable
type StatusResult struct {
	CurrentBranch  string
	CurrentTicket  *ticket.Ticket
	WorktreePath   string
	Summary        map[string]int
	TotalTickets   int
}

// TextRepresentation returns human-readable format for status
func (r *StatusResult) TextRepresentation() string {
	var buf strings.Builder
	buf.Grow(512)

	fmt.Fprintf(&buf, "\n🌿 Current branch: %s\n", r.CurrentBranch)

	if r.CurrentTicket != nil {
		fmt.Fprintf(&buf, "\n🎯 Active ticket: %s\n", r.CurrentTicket.ID)
		fmt.Fprintf(&buf, "   Description: %s\n", r.CurrentTicket.Description)
		fmt.Fprintf(&buf, "   Status: %s\n", r.CurrentTicket.Status())
		if r.CurrentTicket.StartedAt.Time != nil {
			duration := time.Since(*r.CurrentTicket.StartedAt.Time)
			fmt.Fprintf(&buf, "   Duration: %s\n", formatDuration(duration))
		}
		if r.WorktreePath != "" {
			fmt.Fprintf(&buf, "   Worktree: %s\n", r.WorktreePath)
		}
	} else {
		buf.WriteString("\n⚠️  No active ticket\n")
		buf.WriteString("   Start a ticket with: ticketflow start <ticket-id>\n")
	}

	fmt.Fprintf(&buf, "\n📊 Ticket summary:\n")
	fmt.Fprintf(&buf, "   📘 Todo:  %d\n", r.Summary["todo"])
	fmt.Fprintf(&buf, "   🔨 Doing: %d\n", r.Summary["doing"])
	fmt.Fprintf(&buf, "   ✅ Done:  %d\n", r.Summary["done"])
	fmt.Fprintf(&buf, "   ─────────\n")
	fmt.Fprintf(&buf, "   🔢 Total: %d\n", r.TotalTickets)

	return buf.String()
}

// StructuredData returns status data for JSON serialization
func (r *StatusResult) StructuredData() interface{} {
	output := map[string]interface{}{
		"current_branch": r.CurrentBranch,
		"summary":        r.Summary,
	}

	if r.CurrentTicket != nil {
		output["current_ticket"] = ticketToJSON(r.CurrentTicket, r.WorktreePath)
	} else {
		output["current_ticket"] = nil
	}

	return output
}

// StartResult wraps StartTicketResult to make it Printable
type StartResult struct {
	*StartTicketResult
	WorktreeEnabled bool
}

// TextRepresentation returns human-readable format for start result
func (r *StartResult) TextRepresentation() string {
	var buf strings.Builder
	buf.Grow(1024)

	fmt.Fprintf(&buf, "\n✅ Started work on ticket: %s\n", r.Ticket.ID)
	fmt.Fprintf(&buf, "   Description: %s\n", r.Ticket.Description)

	if r.WorktreeEnabled {
		// Worktree mode
		fmt.Fprintf(&buf, "\n📁 Worktree created: %s\n", r.WorktreePath)
		if r.ParentBranch != "" {
			fmt.Fprintf(&buf, "   Parent ticket: %s\n", r.ParentBranch)
			fmt.Fprintf(&buf, "   Branch from: %s\n", r.ParentBranch)
		}
		fmt.Fprintf(&buf, "   Status: todo → doing\n")
		fmt.Fprintf(&buf, "   Committed: \"Start ticket: %s\"\n", r.Ticket.ID)
		fmt.Fprintf(&buf, "\n📋 Next steps:\n")
		fmt.Fprintf(&buf, "1. Navigate to worktree:\n")
		fmt.Fprintf(&buf, "   cd %s\n", r.WorktreePath)
		fmt.Fprintf(&buf, "   \n")
		fmt.Fprintf(&buf, "2. Make your changes and commit regularly\n")
		fmt.Fprintf(&buf, "   \n")
		fmt.Fprintf(&buf, "3. Push branch to create PR:\n")
		fmt.Fprintf(&buf, "   git push -u origin %s\n", r.Ticket.ID)
		fmt.Fprintf(&buf, "   \n")
		fmt.Fprintf(&buf, "4. When done, close the ticket:\n")
		fmt.Fprintf(&buf, "   ticketflow close\n")
	} else {
		// Branch mode
		fmt.Fprintf(&buf, "\n🌿 Switched to branch: %s\n", r.Ticket.ID)
		fmt.Fprintf(&buf, "   Status: todo → doing\n")
		fmt.Fprintf(&buf, "   Committed: \"Start ticket: %s\"\n", r.Ticket.ID)
		fmt.Fprintf(&buf, "\n📋 Next steps:\n")
		fmt.Fprintf(&buf, "1. Make your changes and commit regularly\n")
		fmt.Fprintf(&buf, "   \n")
		fmt.Fprintf(&buf, "2. Push branch to create PR:\n")
		fmt.Fprintf(&buf, "   git push -u origin %s\n", r.Ticket.ID)
		fmt.Fprintf(&buf, "   \n")
		fmt.Fprintf(&buf, "3. When done, close the ticket:\n")
		fmt.Fprintf(&buf, "   ticketflow close\n")
	}

	return buf.String()
}

// StructuredData returns start data for JSON serialization
func (r *StartResult) StructuredData() interface{} {
	return map[string]interface{}{
		"ticket_id":              r.Ticket.ID,
		"status":                 string(r.Ticket.Status()),
		"worktree_path":          r.WorktreePath,
		"branch":                 r.Ticket.ID,
		"parent_branch":          r.ParentBranch,
		"init_commands_executed": r.InitCommandsExecuted,
	}
}
