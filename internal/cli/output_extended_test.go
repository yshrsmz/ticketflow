package cli

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestTicketToJSON(t *testing.T) {
	now := time.Now()
	startTime := now.Add(-1 * time.Hour)
	closeTime := now

	tests := []struct {
		name     string
		ticket   *ticket.Ticket
		expected map[string]interface{}
	}{
		{
			name: "full ticket",
			ticket: &ticket.Ticket{
				ID:          "250101-120000-test-feature",
				Path:        "/tickets/done/250101-120000-test-feature.md",
				Priority:    1,
				Description: "Test feature implementation",
				CreatedAt:   ticket.NewRFC3339Time(now.Add(-2 * time.Hour)),
				StartedAt:   ticket.NewRFC3339TimePtr(&startTime),
				ClosedAt:    ticket.NewRFC3339TimePtr(&closeTime),
				Related:     []string{"parent:250101-110000-parent-feature"},
				Content:     "# Test Feature\n\nThis is the content",
			},
			expected: map[string]interface{}{
				"id":           "250101-120000-test-feature",
				"path":         "/tickets/done/250101-120000-test-feature.md",
				"status":       "done",
				"priority":     float64(1),
				"description":  "Test feature implementation",
				"created_at":   now.Add(-2 * time.Hour).Format(time.RFC3339),
				"started_at":   startTime.Format(time.RFC3339),
				"closed_at":    closeTime.Format(time.RFC3339),
				"related":      []interface{}{"parent:250101-110000-parent-feature"},
				"has_worktree": false,
			},
		},
		{
			name: "minimal ticket",
			ticket: &ticket.Ticket{
				ID:          "250101-120000-minimal",
				Path:        "/tickets/todo/250101-120000-minimal.md",
				Priority:    2,
				Description: "",
				CreatedAt:   ticket.NewRFC3339Time(now),
				Related:     nil,
			},
			expected: map[string]interface{}{
				"id":           "250101-120000-minimal",
				"path":         "/tickets/todo/250101-120000-minimal.md",
				"status":       "todo",
				"priority":     float64(2),
				"description":  "",
				"created_at":   now.Format(time.RFC3339),
				"started_at":   nil,
				"closed_at":    nil,
				"related":      nil,
				"has_worktree": false,
			},
		},
		{
			name: "ticket in doing status",
			ticket: &ticket.Ticket{
				ID:          "250101-120000-worktree",
				Path:        "/tickets/doing/250101-120000-worktree.md",
				Priority:    2,
				Description: "Feature in progress",
				CreatedAt:   ticket.NewRFC3339Time(now),
				StartedAt:   ticket.NewRFC3339TimePtr(&startTime),
			},
			expected: map[string]interface{}{
				"id":           "250101-120000-worktree",
				"path":         "/tickets/doing/250101-120000-worktree.md",
				"status":       "doing",
				"priority":     float64(2),
				"description":  "Feature in progress",
				"created_at":   now.Format(time.RFC3339),
				"started_at":   startTime.Format(time.RFC3339),
				"closed_at":    nil,
				"related":      nil,
				"has_worktree": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ticketToJSON(tt.ticket, "")

			// Convert to JSON and back to ensure proper comparison
			jsonBytes, err := json.Marshal(result)
			require.NoError(t, err)

			var actual map[string]interface{}
			err = json.Unmarshal(jsonBytes, &actual)
			require.NoError(t, err)

			// Compare expected fields
			for key, expectedValue := range tt.expected {
				actualValue, ok := actual[key]
				assert.True(t, ok, "Missing key: %s", key)
				
				// Special handling for time values
				if key == "created_at" || key == "started_at" || key == "closed_at" {
					if expectedValue != nil {
						// Parse both times to ensure they're equivalent
						expectedTime, err1 := time.Parse(time.RFC3339, expectedValue.(string))
						actualTime, err2 := time.Parse(time.RFC3339, actualValue.(string))
						if err1 == nil && err2 == nil {
							assert.Equal(t, expectedTime.Unix(), actualTime.Unix())
							continue
						}
					}
				}
				
				assert.Equal(t, expectedValue, actualValue, "Mismatch for key: %s", key)
			}
		})
	}
}

func TestFormatDuration_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "negative duration",
			duration: -5 * time.Hour,
			expected: "0s",
		},
		{
			name:     "zero duration",
			duration: 0,
			expected: "0m",
		},
		{
			name:     "less than minute",
			duration: 30 * time.Second,
			expected: "0m",
		},
		{
			name:     "exactly one minute",
			duration: 1 * time.Minute,
			expected: "1m",
		},
		{
			name:     "59 minutes",
			duration: 59 * time.Minute,
			expected: "59m",
		},
		{
			name:     "exactly one hour",
			duration: 1 * time.Hour,
			expected: "1h",
		},
		{
			name:     "23 hours",
			duration: 23 * time.Hour,
			expected: "23h",
		},
		{
			name:     "exactly one day",
			duration: 24 * time.Hour,
			expected: "1d",
		},
		{
			name:     "multiple days with hours",
			duration: 73 * time.Hour,
			expected: "3d 1h",
		},
		{
			name:     "days with minutes (no hours)",
			duration: 24*time.Hour + 30*time.Minute,
			expected: "1d 30m",
		},
		{
			name:     "only days",
			duration: 48 * time.Hour,
			expected: "2d",
		},
		{
			name:     "complex duration",
			duration: 50*time.Hour + 45*time.Minute,
			expected: "2d 2h 45m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOutputJSON_Function(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		expected string
	}{
		{
			name: "simple structure",
			data: map[string]string{
				"status": "success",
			},
			expected: `{"status":"success"}`,
		},
		{
			name: "nested structure",
			data: map[string]interface{}{
				"ticket": map[string]interface{}{
					"id":       "test-123",
					"priority": 1,
					"tags":     []string{"bug", "urgent"},
				},
			},
			expected: `{"ticket":{"id":"test-123","priority":1,"tags":["bug","urgent"]}}`,
		},
		{
			name:     "empty map",
			data:     map[string]interface{}{},
			expected: `{}`,
		},
		{
			name: "nil values",
			data: map[string]interface{}{
				"value": nil,
			},
			expected: `{"value":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.data)
			require.NoError(t, err)

			result := outputJSON(tt.data)
			require.NoError(t, result)

			// Parse both to ensure they represent the same data
			var expected, actual interface{}
			err = json.Unmarshal([]byte(tt.expected), &expected)
			require.NoError(t, err)

			err = json.Unmarshal(jsonBytes, &actual)
			require.NoError(t, err)

			assert.Equal(t, expected, actual)
		})
	}
}

// Benchmarks for performance-critical functions
func BenchmarkFormatDuration(b *testing.B) {
	durations := []time.Duration{
		30 * time.Second,
		5 * time.Minute,
		2 * time.Hour,
		25 * time.Hour,
		73*time.Hour + 45*time.Minute,
	}

	b.ReportAllocs() // Report memory allocations
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, d := range durations {
			_ = formatDuration(d)
		}
	}
}

func BenchmarkTicketToJSON(b *testing.B) {
	now := time.Now()
	startTime := now.Add(-1 * time.Hour)
	
	testTicket := &ticket.Ticket{
		ID:          "250101-120000-benchmark",
		Path:        "/tickets/doing/250101-120000-benchmark.md",
		Priority:    1,
		Description: "Benchmark test ticket",
		CreatedAt:   ticket.NewRFC3339Time(now),
		StartedAt:   ticket.NewRFC3339TimePtr(&startTime),
		Related:     []string{"parent:250101-110000-parent", "blocks:250101-130000-blocked"},
		Content:     "# Benchmark Test\n\nThis is a longer content to simulate real ticket data.\n\n## Tasks\n- [ ] Task 1\n- [ ] Task 2",
	}

	b.ReportAllocs() // Report memory allocations
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ticketToJSON(testTicket, "")
	}
}

func BenchmarkParseOutputFormat(b *testing.B) {
	formats := []string{"json", "JSON", "text", "TEXT", "", "invalid", "yaml"}

	b.ReportAllocs() // Report memory allocations
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, format := range formats {
			_ = ParseOutputFormat(format)
		}
	}
}