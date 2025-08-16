package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// OutputFormat represents the output format
type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
)

// ParseOutputFormat parses output format from string
func ParseOutputFormat(format string) OutputFormat {
	switch strings.ToLower(format) {
	case "json":
		return FormatJSON
	default:
		return FormatText
	}
}

// ResultWriter handles structured data output for CLI commands.
// It formats the final results according to the selected output format.
type ResultWriter interface {
	PrintResult(data interface{}) error
	// Keep PrintJSON for backward compatibility during migration
	PrintJSON(data interface{}) error
}

// jsonResultWriter outputs data in JSON format
type jsonResultWriter struct {
	encoder *json.Encoder
}

// NewJSONResultWriter creates a result writer for JSON output
func NewJSONResultWriter(w io.Writer) ResultWriter {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return &jsonResultWriter{encoder: encoder}
}

func (w *jsonResultWriter) PrintResult(data interface{}) error {
	return w.encoder.Encode(data)
}

func (w *jsonResultWriter) PrintJSON(data interface{}) error {
	return w.encoder.Encode(data)
}

// textResultWriter outputs data in human-readable text format
type textResultWriter struct {
	w io.Writer
}

// NewTextResultWriter creates a result writer for text output
func NewTextResultWriter(w io.Writer) ResultWriter {
	return &textResultWriter{w: w}
}

func (w *textResultWriter) PrintResult(data interface{}) error {
	// Handle different data types with appropriate formatting
	switch v := data.(type) {
	case *CleanupResult:
		return w.printCleanupResult(v)
	case *ticket.Ticket:
		return w.printTicket(v)
	case []*ticket.Ticket:
		return w.printTicketList(v)
	case map[string]interface{}:
		// Handle generic map data
		return w.printMap(v)
	default:
		// Fallback to simple string representation
		_, err := fmt.Fprintf(w.w, "%v\n", v)
		return err
	}
}

func (w *textResultWriter) PrintJSON(data interface{}) error {
	// In text mode, pretty-print JSON-like data
	return w.PrintResult(data)
}

func (w *textResultWriter) printCleanupResult(r *CleanupResult) error {
	fmt.Fprintf(w.w, "\nCleanup Summary:\n")
	fmt.Fprintf(w.w, "  Orphaned worktrees removed: %d\n", r.OrphanedWorktrees)
	fmt.Fprintf(w.w, "  Stale branches removed: %d\n", r.StaleBranches)

	if len(r.Errors) > 0 {
		fmt.Fprintf(w.w, "\nWarnings:\n")
		for _, err := range r.Errors {
			fmt.Fprintf(w.w, "  - %s\n", err)
		}
	}
	return nil
}

func (w *textResultWriter) printTicket(t *ticket.Ticket) error {
	fmt.Fprintf(w.w, "Ticket: %s\n", t.ID)
	fmt.Fprintf(w.w, "Status: %s\n", t.Status())
	fmt.Fprintf(w.w, "Priority: %d\n", t.Priority)
	fmt.Fprintf(w.w, "Description: %s\n", t.Description)

	if !t.CreatedAt.Time.IsZero() {
		fmt.Fprintf(w.w, "Created: %s\n", t.CreatedAt.Time.Format(time.RFC3339))
	}
	if !t.StartedAt.Time.IsZero() {
		fmt.Fprintf(w.w, "Started: %s\n", t.StartedAt.Time.Format(time.RFC3339))
	}
	if !t.ClosedAt.Time.IsZero() {
		fmt.Fprintf(w.w, "Closed: %s\n", t.ClosedAt.Time.Format(time.RFC3339))
	}

	return nil
}

func (w *textResultWriter) printTicketList(tickets []*ticket.Ticket) error {
	if len(tickets) == 0 {
		fmt.Fprintln(w.w, "No tickets found")
		return nil
	}

	for _, t := range tickets {
		fmt.Fprintf(w.w, "[%s] %s - %s\n", t.Status(), t.ID, t.Description)
	}
	return nil
}

func (w *textResultWriter) printMap(m map[string]interface{}) error {
	// Simple key-value printing for generic maps
	for k, v := range m {
		fmt.Fprintf(w.w, "%s: %v\n", k, v)
	}
	return nil
}

// NewResultWriter creates the appropriate result writer based on the output format
func NewResultWriter(w io.Writer, format OutputFormat) ResultWriter {
	if format == FormatJSON {
		return NewJSONResultWriter(w)
	}
	return NewTextResultWriter(w)
}

// Legacy OutputWriter - kept for backward compatibility during migration
// This will be removed once all code is migrated to use ResultWriter
type OutputWriter struct {
	stdout io.Writer
	stderr io.Writer
	format OutputFormat
	result ResultWriter
}

// NewOutputWriter creates a new OutputWriter with the specified format
func NewOutputWriter(stdout, stderr io.Writer, format OutputFormat) *OutputWriter {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	return &OutputWriter{
		stdout: stdout,
		stderr: stderr,
		format: format,
		result: NewResultWriter(stdout, format),
	}
}

// PrintJSON writes JSON output to stdout
func (w *OutputWriter) PrintJSON(data interface{}) error {
	return w.result.PrintJSON(data)
}

// Printf writes formatted text to stdout - DEPRECATED
func (w *OutputWriter) Printf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(w.stdout, format, args...)
}

// Println writes a line to stdout - DEPRECATED
func (w *OutputWriter) Println(args ...interface{}) {
	_, _ = fmt.Fprintln(w.stdout, args...)
}

// GetFormat returns the current output format
func (w *OutputWriter) GetFormat() OutputFormat {
	return w.format
}

// PrintResult delegates to the result writer
func (w *OutputWriter) PrintResult(data interface{}) error {
	return w.result.PrintResult(data)
}

// Helper functions

// outputJSON outputs data as JSON - kept for backward compatibility
func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "0s"
	}

	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	// Pre-allocate parts slice with capacity 3 (days, hours, minutes)
	parts := make([]string, 0, 3)
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}

	return strings.Join(parts, " ")
}

// ticketToJSON converts a ticket to JSON representation
func ticketToJSON(t *ticket.Ticket, worktreePath string) map[string]interface{} {
	result := map[string]interface{}{
		"id":           t.ID,
		"path":         t.Path,
		"status":       string(t.Status()),
		"priority":     t.Priority,
		"description":  t.Description,
		"created_at":   t.CreatedAt.Time,
		"started_at":   t.StartedAt.Time,
		"closed_at":    t.ClosedAt.Time,
		"related":      t.Related,
		"has_worktree": t.HasWorktree(),
	}

	if worktreePath != "" {
		result["worktree_path"] = worktreePath
	}

	return result
}
