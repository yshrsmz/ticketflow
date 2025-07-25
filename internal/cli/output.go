package cli

import (
	"encoding/json"
	"fmt"
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

// outputJSON outputs data as JSON
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
		"id":          t.ID,
		"path":        t.Path,
		"status":      string(t.Status()),
		"priority":    t.Priority,
		"description": t.Description,
		"created_at":  t.CreatedAt,
		"started_at":  t.StartedAt,
		"closed_at":   t.ClosedAt,
		"related":     t.Related,
		"has_worktree": t.HasWorktree(),
	}
	
	if worktreePath != "" {
		result["worktree_path"] = worktreePath
	}
	
	return result
}