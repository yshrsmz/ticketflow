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

// OutputWriter handles formatted output for CLI commands
type OutputWriter struct {
	stdout io.Writer
	stderr io.Writer
	format OutputFormat
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
	}
}

// PrintJSON writes JSON output to stdout
func (w *OutputWriter) PrintJSON(data interface{}) error {
	encoder := json.NewEncoder(w.stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// Printf writes formatted text to stdout
func (w *OutputWriter) Printf(format string, args ...interface{}) {
	fmt.Fprintf(w.stdout, format, args...)
}

// Println writes a line to stdout
func (w *OutputWriter) Println(args ...interface{}) {
	_, _ = fmt.Fprintln(w.stdout, args...)
}

// GetFormat returns the current output format
func (w *OutputWriter) GetFormat() OutputFormat {
	return w.format
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

	parts := []string{}
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
