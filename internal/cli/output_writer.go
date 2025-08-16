package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
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

// OutputFormatter handles structured data output for CLI commands.
// It formats the final results according to the selected output format.
//
// OutputFormatter is responsible for the final data output (JSON or formatted text),
// while StatusWriter (in status_writer.go) handles progress messages during execution.
// This separation ensures that JSON output remains valid by suppressing status
// messages in JSON mode while still providing user feedback in text mode.
//
// See also: StatusWriter in status_writer.go for progress/status messages.
type OutputFormatter interface {
	PrintResult(data interface{}) error
	// Keep PrintJSON for backward compatibility during migration
	PrintJSON(data interface{}) error
}

// Verify interface compliance at compile time
var (
	_ OutputFormatter = (*jsonOutputFormatter)(nil)
	_ OutputFormatter = (*textOutputFormatter)(nil)
)

// jsonOutputFormatter outputs data in JSON format
type jsonOutputFormatter struct {
	encoder *json.Encoder
}

// NewJSONOutputFormatter creates an output formatter for JSON output
func NewJSONOutputFormatter(w io.Writer) OutputFormatter {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return &jsonOutputFormatter{encoder: encoder}
}

func (w *jsonOutputFormatter) PrintResult(data interface{}) error {
	// Check if data implements Printable interface
	if p, ok := data.(Printable); ok {
		return w.encoder.Encode(p.StructuredData())
	}
	// Fallback to encoding the data directly
	return w.encoder.Encode(data)
}

func (w *jsonOutputFormatter) PrintJSON(data interface{}) error {
	return w.encoder.Encode(data)
}

// textOutputFormatter outputs data in human-readable text format
type textOutputFormatter struct {
	mu sync.Mutex
	w  io.Writer
}

// NewTextOutputFormatter creates an output formatter for text output
func NewTextOutputFormatter(w io.Writer) OutputFormatter {
	return &textOutputFormatter{w: w}
}

func (w *textOutputFormatter) PrintResult(data interface{}) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check if data implements Printable interface first
	if p, ok := data.(Printable); ok {
		_, err := fmt.Fprint(w.w, p.TextRepresentation())
		return err
	}

	// Fallback to type switch for backward compatibility
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

func (w *textOutputFormatter) PrintJSON(data interface{}) error {
	// In text mode, pretty-print JSON-like data
	return w.PrintResult(data)
}

func (w *textOutputFormatter) printCleanupResult(r *CleanupResult) error {
	// Using fmt.Fprint to combine all output and check error once
	_, err := fmt.Fprintf(w.w, "\nCleanup Summary:\n  Orphaned worktrees removed: %d\n  Stale branches removed: %d\n",
		r.OrphanedWorktrees, r.StaleBranches)
	if err != nil {
		return err
	}

	if len(r.Errors) > 0 {
		_, err = fmt.Fprint(w.w, "\nWarnings:\n")
		if err != nil {
			return err
		}
		for _, errMsg := range r.Errors {
			_, err = fmt.Fprintf(w.w, "  - %s\n", errMsg)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *textOutputFormatter) printTicket(t *ticket.Ticket) error {
	_, err := fmt.Fprintf(w.w, "Ticket: %s\nStatus: %s\nPriority: %d\nDescription: %s\n",
		t.ID, t.Status(), t.Priority, t.Description)
	if err != nil {
		return err
	}

	if !t.CreatedAt.IsZero() {
		_, err = fmt.Fprintf(w.w, "Created: %s\n", t.CreatedAt.Format(time.RFC3339))
		if err != nil {
			return err
		}
	}
	if isTimeSet(t.StartedAt.Time) {
		_, err = fmt.Fprintf(w.w, "Started: %s\n", t.StartedAt.Time.Format(time.RFC3339))
		if err != nil {
			return err
		}
	}
	if isTimeSet(t.ClosedAt.Time) {
		_, err = fmt.Fprintf(w.w, "Closed: %s\n", t.ClosedAt.Time.Format(time.RFC3339))
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *textOutputFormatter) printTicketList(tickets []*ticket.Ticket) error {
	if len(tickets) == 0 {
		_, err := fmt.Fprintln(w.w, "No tickets found")
		return err
	}

	for _, t := range tickets {
		_, err := fmt.Fprintf(w.w, "[%s] %s - %s\n", t.Status(), t.ID, t.Description)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *textOutputFormatter) printMap(m map[string]interface{}) error {
	// Simple key-value printing for generic maps
	for k, v := range m {
		_, err := fmt.Fprintf(w.w, "%s: %v\n", k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewOutputFormatter creates the appropriate output formatter based on the output format
func NewOutputFormatter(w io.Writer, format OutputFormat) OutputFormatter {
	if format == FormatJSON {
		return NewJSONOutputFormatter(w)
	}
	return NewTextOutputFormatter(w)
}

// Legacy OutputWriter - kept for backward compatibility during migration
// This will be removed once all code is migrated to use OutputFormatter
type OutputWriter struct {
	stdout    io.Writer
	stderr    io.Writer
	format    OutputFormat
	formatter OutputFormatter
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
		stdout:    stdout,
		stderr:    stderr,
		format:    format,
		formatter: NewOutputFormatter(stdout, format),
	}
}

// PrintJSON writes JSON output to stdout
func (w *OutputWriter) PrintJSON(data interface{}) error {
	return w.formatter.PrintJSON(data)
}

// Printf writes formatted text to stdout
// Deprecated: Use StatusWriter for progress messages or OutputFormatter for structured output
func (w *OutputWriter) Printf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(w.stdout, format, args...)
}

// Println writes a line to stdout  
// Deprecated: Use StatusWriter for progress messages or OutputFormatter for structured output
func (w *OutputWriter) Println(args ...interface{}) {
	_, _ = fmt.Fprintln(w.stdout, args...)
}

// GetFormat returns the current output format
func (w *OutputWriter) GetFormat() OutputFormat {
	return w.format
}

// PrintResult delegates to the output formatter
func (w *OutputWriter) PrintResult(data interface{}) error {
	return w.formatter.PrintResult(data)
}

// Helper functions

// isTimeSet checks if a time pointer is not nil and not zero
func isTimeSet(t *time.Time) bool {
	return t != nil && !t.IsZero()
}

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
