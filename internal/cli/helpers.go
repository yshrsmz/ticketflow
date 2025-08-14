package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// CalculateDuration calculates the work duration for a ticket
func CalculateDuration(t *ticket.Ticket) time.Duration {
	if t.StartedAt.Time == nil || t.ClosedAt.Time == nil {
		return 0
	}
	return t.ClosedAt.Time.Sub(*t.StartedAt.Time)
}

// ExtractParentID extracts the parent ticket ID from a ticket's Related field
func ExtractParentID(t *ticket.Ticket) string {
	for _, rel := range t.Related {
		if strings.HasPrefix(rel, "parent:") {
			return strings.TrimPrefix(rel, "parent:")
		}
	}
	return ""
}

// FormatDuration formats a duration as human-readable string (e.g., "2h30m")
func FormatDuration(d time.Duration) string {
	if d == 0 {
		return ""
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh%dm", hours, minutes)
}
