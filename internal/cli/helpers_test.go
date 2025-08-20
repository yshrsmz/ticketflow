package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestCalculateDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ticket   *ticket.Ticket
		expected time.Duration
	}{
		{
			name: "normal case with both times",
			ticket: &ticket.Ticket{
				StartedAt: ticket.RFC3339TimePtr{Time: timePtr(time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC))},
				ClosedAt:  ticket.RFC3339TimePtr{Time: timePtr(time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC))},
			},
			expected: 2*time.Hour + 30*time.Minute,
		},
		{
			name: "no started time",
			ticket: &ticket.Ticket{
				ClosedAt: ticket.RFC3339TimePtr{Time: timePtr(time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC))},
			},
			expected: 0,
		},
		{
			name: "no closed time",
			ticket: &ticket.Ticket{
				StartedAt: ticket.RFC3339TimePtr{Time: timePtr(time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC))},
			},
			expected: 0,
		},
		{
			name:     "both times nil",
			ticket:   &ticket.Ticket{},
			expected: 0,
		},
		{
			name:     "nil ticket",
			ticket:   nil,
			expected: 0,
		},
		{
			name: "closed before started (invalid state)",
			ticket: &ticket.Ticket{
				StartedAt: ticket.RFC3339TimePtr{Time: timePtr(time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC))},
				ClosedAt:  ticket.RFC3339TimePtr{Time: timePtr(time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC))},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateDuration(tt.ticket)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractParentID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ticket   *ticket.Ticket
		expected string
	}{
		{
			name: "ticket with parent",
			ticket: &ticket.Ticket{
				Related: []string{"parent:parent-ticket-123", "related:other-ticket"},
			},
			expected: "parent-ticket-123",
		},
		{
			name: "ticket without parent",
			ticket: &ticket.Ticket{
				Related: []string{"related:other-ticket", "blocks:another-ticket"},
			},
			expected: "",
		},
		{
			name:     "ticket with no relations",
			ticket:   &ticket.Ticket{},
			expected: "",
		},
		{
			name: "ticket with multiple parents (takes first)",
			ticket: &ticket.Ticket{
				Related: []string{"parent:first-parent", "parent:second-parent"},
			},
			expected: "first-parent",
		},
		{
			name:     "nil ticket",
			ticket:   nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractParentID(tt.ticket)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHelperFormatDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "zero duration",
			duration: 0,
			expected: "",
		},
		{
			name:     "hours and minutes",
			duration: 2*time.Hour + 30*time.Minute,
			expected: "2h 30m",
		},
		{
			name:     "only hours",
			duration: 3 * time.Hour,
			expected: "3h 0m",
		},
		{
			name:     "only minutes",
			duration: 45 * time.Minute,
			expected: "0h 45m",
		},
		{
			name:     "more than 24 hours",
			duration: 25*time.Hour + 15*time.Minute,
			expected: "25h 15m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
