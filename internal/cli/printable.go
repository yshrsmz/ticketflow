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
	// This follows Git's default abbreviation length
	GitSHAShortLength = 7

	// Buffer pre-allocation sizes for strings.Builder
	// These are estimates based on typical output sizes to minimize allocations
	smallBufferSize  = 256  // For simple results with minimal text
	mediumBufferSize = 512  // For typical results with moderate text
	largeBufferSize  = 1024 // For complex results with extensive text
)

// Printable represents a result that knows how to format itself.
// This interface follows the pattern used by kubectl's ResourcePrinter,
// where each result type owns its formatting logic instead of having
// a central switch statement handle all types.
//
// Benefits of this pattern:
//   - Single Responsibility: Each result type manages its own formatting
//   - Open/Closed Principle: New result types can be added without modifying existing code
//   - Better testability: Each result type can be tested independently
//   - Reduced coupling: Business logic and presentation are cleanly separated
//
// Implementation guidelines:
//   - TextRepresentation should return formatted text for human consumption
//   - StructuredData should return data suitable for JSON/YAML serialization
//   - Pre-allocate strings.Builder capacity based on expected output size
//   - Use compile-time interface compliance checks (see examples below)
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
	_ Printable = (*NewTicketResult)(nil)
	_ Printable = (*CloseTicketResult)(nil)
	_ Printable = (*RestoreTicketResult)(nil)
)

// TextRepresentation returns human-readable format for CleanupResult
func (r *CleanupResult) TextRepresentation() string {
	var buf strings.Builder
	// Pre-allocate capacity for better performance
	buf.Grow(smallBufferSize)

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
	buf.Grow(mediumBufferSize)

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
	buf.Grow(mediumBufferSize)

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
	buf.Grow(mediumBufferSize)

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
	CurrentBranch string
	CurrentTicket *ticket.Ticket
	WorktreePath  string
	Summary       map[string]int
	TotalTickets  int
}

// TextRepresentation returns human-readable format for status
func (r *StatusResult) TextRepresentation() string {
	var buf strings.Builder
	buf.Grow(mediumBufferSize)

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
	buf.Grow(largeBufferSize)

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

// NewTicketResult represents the result of creating a new ticket
type NewTicketResult struct {
	Ticket       *ticket.Ticket
	ParentTicket string // Parent ticket ID if this is a sub-ticket
}

// TextRepresentation returns human-readable format for new ticket result
func (r *NewTicketResult) TextRepresentation() string {
	if r.Ticket == nil {
		return "Error: No ticket created\n"
	}

	var buf strings.Builder
	buf.Grow(mediumBufferSize)

	fmt.Fprintf(&buf, "\n🎫 Created new ticket: %s\n", r.Ticket.ID)
	fmt.Fprintf(&buf, "   File: %s\n", r.Ticket.Path)
	if r.ParentTicket != "" {
		fmt.Fprintf(&buf, "   Parent ticket: %s\n", r.ParentTicket)
		fmt.Fprintf(&buf, "   Type: Sub-ticket\n")
	}
	fmt.Fprintf(&buf, "\n📋 Next steps:\n")
	fmt.Fprintf(&buf, "1. Edit the ticket file to add details:\n")
	fmt.Fprintf(&buf, "   $EDITOR %s\n", r.Ticket.Path)
	fmt.Fprintf(&buf, "   \n")
	fmt.Fprintf(&buf, "2. Commit the ticket file:\n")
	fmt.Fprintf(&buf, "   git add %s\n", r.Ticket.Path)
	fmt.Fprintf(&buf, "   git commit -m \"Add ticket: %s\"\n", r.Ticket.ID)
	fmt.Fprintf(&buf, "   \n")
	fmt.Fprintf(&buf, "3. Start working on the ticket:\n")
	fmt.Fprintf(&buf, "   ticketflow start %s\n", r.Ticket.ID)

	return buf.String()
}

// StructuredData returns data for JSON serialization
func (r *NewTicketResult) StructuredData() interface{} {
	if r.Ticket == nil {
		return nil
	}

	output := map[string]interface{}{
		"ticket": map[string]interface{}{
			"id":   r.Ticket.ID,
			"path": r.Ticket.Path,
		},
	}

	if r.ParentTicket != "" {
		output["parent_ticket"] = r.ParentTicket
	}

	return output
}

// CloseTicketResult represents the result of closing a ticket
type CloseTicketResult struct {
	Ticket        *ticket.Ticket
	Mode          string        // "current" or "by_id"
	ForceUsed     bool          // Whether --force flag was used
	CommitCreated bool          // Whether a commit was created
	CloseReason   string        // Optional close reason
	Duration      time.Duration // Duration from start to close (for current ticket)
	ParentTicket  string        // Parent ticket ID if available
	WorktreePath  string        // Worktree path for current ticket
	Branch        string        // Branch name for by-ID mode
}

// TextRepresentation returns human-readable format for close result
func (r *CloseTicketResult) TextRepresentation() string {
	if r.Ticket == nil {
		return "Error: No ticket to close\n"
	}

	var buf strings.Builder
	buf.Grow(mediumBufferSize)

	// Success message
	if r.Mode == "current" {
		fmt.Fprintf(&buf, "\n✅ Closed current ticket: %s\n", r.Ticket.ID)
	} else {
		fmt.Fprintf(&buf, "\n✅ Closed ticket: %s\n", r.Ticket.ID)
	}

	// Show force flag usage if applicable
	if r.ForceUsed {
		fmt.Fprintf(&buf, "   ⚠️  Force flag used to bypass validation\n")
	}

	// Show close reason if provided
	if r.CloseReason != "" {
		fmt.Fprintf(&buf, "   Reason: %s\n", r.CloseReason)
	}

	// Show duration for current ticket
	if r.Mode == "current" && r.Duration > 0 {
		fmt.Fprintf(&buf, "   Duration: %s\n", formatDuration(r.Duration))
	}

	// Show parent ticket if available
	if r.ParentTicket != "" {
		fmt.Fprintf(&buf, "   Parent ticket: %s\n", r.ParentTicket)
	}

	// Show status transition
	fmt.Fprintf(&buf, "   Status: doing → done\n")

	// Show commit info
	if r.CommitCreated {
		fmt.Fprintf(&buf, "   Committed: \"Close ticket: %s", r.Ticket.ID)
		if r.CloseReason != "" {
			fmt.Fprintf(&buf, " - %s", r.CloseReason)
		}
		fmt.Fprintf(&buf, "\"\n")
	}

	// Next steps
	fmt.Fprintf(&buf, "\n📋 Next steps:\n")
	if r.Mode == "current" {
		fmt.Fprintf(&buf, "1. Push your branch to create/update PR:\n")
		fmt.Fprintf(&buf, "   git push\n")
		fmt.Fprintf(&buf, "   \n")
		fmt.Fprintf(&buf, "2. After PR is merged, clean up the worktree:\n")
		fmt.Fprintf(&buf, "   ticketflow cleanup %s\n", r.Ticket.ID)
	} else {
		fmt.Fprintf(&buf, "1. Push the branch if needed:\n")
		fmt.Fprintf(&buf, "   git push origin %s\n", r.Branch)
		fmt.Fprintf(&buf, "   \n")
		fmt.Fprintf(&buf, "2. Create a pull request if needed\n")
	}

	return buf.String()
}

// StructuredData returns data for JSON serialization
func (r *CloseTicketResult) StructuredData() interface{} {
	if r.Ticket == nil {
		return nil
	}

	output := map[string]interface{}{
		"success":        true,
		"ticket_id":      r.Ticket.ID,
		"status":         string(r.Ticket.Status()),
		"mode":           r.Mode,
		"force_used":     r.ForceUsed,
		"commit_created": r.CommitCreated,
	}

	if r.Ticket.ClosedAt.Time != nil {
		output["closed_at"] = r.Ticket.ClosedAt.Time.Format(time.RFC3339)
	}

	if r.Mode == "current" && r.Duration > 0 {
		hours := int(r.Duration.Hours())
		minutes := int(r.Duration.Minutes()) % 60
		output["duration"] = fmt.Sprintf("%dh%dm", hours, minutes)
	}

	if r.ParentTicket != "" {
		output["parent_ticket"] = r.ParentTicket
	}

	if r.WorktreePath != "" {
		output["worktree_path"] = r.WorktreePath
	}

	if r.Mode == "by_id" && r.Branch != "" {
		output["branch"] = r.Branch
	}

	if r.CloseReason != "" {
		output["close_reason"] = r.CloseReason
	}

	return output
}

// RestoreTicketResult represents the result of restoring a ticket symlink
type RestoreTicketResult struct {
	Ticket       *ticket.Ticket
	SymlinkPath  string // Path to the symlink (usually "current-ticket.md")
	TargetPath   string // Path the symlink points to
	ParentTicket string // Parent ticket ID if available
	WorktreePath string // Current working directory
}

// TextRepresentation returns human-readable format for restore result
func (r *RestoreTicketResult) TextRepresentation() string {
	return "✅ Current ticket symlink restored\n"
}

// StructuredData returns data for JSON serialization
func (r *RestoreTicketResult) StructuredData() interface{} {
	if r.Ticket == nil {
		return nil
	}

	output := map[string]interface{}{
		"success":          true,
		"ticket_id":        r.Ticket.ID,
		"status":           string(r.Ticket.Status()),
		"symlink_restored": true,
		"symlink_path":     r.SymlinkPath,
		"target_path":      r.TargetPath,
		"message":          "Current ticket symlink restored",
	}

	if r.WorktreePath != "" {
		output["worktree_path"] = r.WorktreePath
	}

	if r.ParentTicket != "" {
		output["parent_ticket"] = r.ParentTicket
	}

	return output
}
