package cli

import (
	"strings"
	"time"

	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// CalculateDuration calculates the work duration for a ticket.
// Returns 0 if the ticket is nil, either timestamp is nil, or if closed time is before started time (invalid state).
func CalculateDuration(t *ticket.Ticket) time.Duration {
	if t == nil || t.StartedAt.Time == nil || t.ClosedAt.Time == nil {
		return 0
	}

	// Guard against invalid state where closed time is before started time
	if t.ClosedAt.Time.Before(*t.StartedAt.Time) {
		return 0
	}

	return t.ClosedAt.Time.Sub(*t.StartedAt.Time)
}

// ExtractParentID extracts the parent ticket ID from a ticket's Related field.
// Returns empty string if the ticket is nil or has no parent relationship.
// Only returns the first parent found if multiple exist (though this should not happen in practice).
func ExtractParentID(t *ticket.Ticket) string {
	if t == nil {
		return ""
	}
	for _, rel := range t.Related {
		if strings.HasPrefix(rel, "parent:") {
			return strings.TrimPrefix(rel, "parent:")
		}
	}
	return ""
}

// FormatDuration formats a duration as human-readable string (e.g., "2h 30m").
// Returns empty string for zero or negative durations.
//
// Deprecated: This function is maintained for backward compatibility.
// New code should use formatDuration which provides more comprehensive formatting
// including support for days and consistent behavior.
//
// Note: This uses space-separated format for better readability.
func FormatDuration(d time.Duration) string {
	if d <= 0 {
		return ""
	}
	// Delegate to the internal helper for consistent formatting
	// The internal helper returns "0s" for zero/negative, but we return ""
	result := formatDuration(d)
	if result == "0s" {
		return ""
	}
	return result
}
