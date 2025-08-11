package ticket

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRFC3339TimeMarshalYAML(t *testing.T) {
	t.Parallel()
	// Test time with nanoseconds
	testTime := time.Date(2025, 1, 27, 19, 38, 54, 927166000, time.FixedZone("JST", 9*60*60))
	rt := NewRFC3339Time(testTime)

	// Marshal to YAML
	data, err := yaml.Marshal(rt)
	require.NoError(t, err)

	// Should format without subseconds
	expected := "\"2025-01-27T19:38:54+09:00\"\n"
	assert.Equal(t, expected, string(data))
}

func TestRFC3339TimeUnmarshalYAML(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		yaml     string
		expected time.Time
		isZero   bool
	}{
		{
			name:     "RFC3339 format",
			yaml:     "2025-01-27T19:38:54+09:00",
			expected: time.Date(2025, 1, 27, 19, 38, 54, 0, time.FixedZone("JST", 9*60*60)),
		},
		{
			name:     "RFC3339Nano format (backward compatibility)",
			yaml:     "2025-01-27T19:38:54.927166+09:00",
			expected: time.Date(2025, 1, 27, 19, 38, 54, 927166000, time.FixedZone("JST", 9*60*60)),
		},
		{
			name:   "null value",
			yaml:   "null",
			isZero: true,
		},
		{
			name:   "empty value",
			yaml:   "",
			isZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rt RFC3339Time
			err := yaml.Unmarshal([]byte(tt.yaml), &rt)
			require.NoError(t, err)

			if tt.isZero {
				assert.True(t, rt.IsZero())
			} else {
				assert.True(t, rt.Equal(tt.expected))
			}
		})
	}
}

func TestRFC3339TimePtrMarshalYAML(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		ptr      RFC3339TimePtr
		expected string
	}{
		{
			name:     "nil time",
			ptr:      RFC3339TimePtr{},
			expected: "null\n",
		},
		{
			name: "valid time",
			ptr: NewRFC3339TimePtr(func() *time.Time {
				t := time.Date(2025, 1, 27, 19, 38, 54, 927166000, time.FixedZone("JST", 9*60*60))
				return &t
			}()),
			expected: "\"2025-01-27T19:38:54+09:00\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(tt.ptr)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))
		})
	}
}

func TestRFC3339TimePtrUnmarshalYAML(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		yaml     string
		expected *time.Time
	}{
		{
			name:     "null value",
			yaml:     "null",
			expected: nil,
		},
		{
			name: "valid RFC3339",
			yaml: "2025-01-27T19:38:54+09:00",
			expected: func() *time.Time {
				t := time.Date(2025, 1, 27, 19, 38, 54, 0, time.FixedZone("JST", 9*60*60))
				return &t
			}(),
		},
		{
			name: "RFC3339Nano (backward compatibility)",
			yaml: "2025-01-27T19:38:54.927166+09:00",
			expected: func() *time.Time {
				t := time.Date(2025, 1, 27, 19, 38, 54, 927166000, time.FixedZone("JST", 9*60*60))
				return &t
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ptr RFC3339TimePtr
			err := yaml.Unmarshal([]byte(tt.yaml), &ptr)
			require.NoError(t, err)

			if tt.expected == nil {
				assert.Nil(t, ptr.Time)
			} else {
				require.NotNil(t, ptr.Time)
				assert.True(t, ptr.Time.Equal(*tt.expected))
			}
		})
	}
}

func TestTicketDateFormatting(t *testing.T) {
	t.Parallel()
	// Create a ticket with nanosecond precision times
	now := time.Date(2025, 1, 27, 19, 38, 54, 927166000, time.FixedZone("JST", 9*60*60))
	startTime := now.Add(5 * time.Minute)
	closeTime := now.Add(10 * time.Minute)

	ticket := &Ticket{
		Priority:    1,
		Description: "Test date formatting",
		CreatedAt:   NewRFC3339Time(now),
		StartedAt:   NewRFC3339TimePtr(&startTime),
		ClosedAt:    NewRFC3339TimePtr(&closeTime),
		Content:     "# Test",
	}

	// Convert to bytes
	data, err := ticket.ToBytes()
	require.NoError(t, err)

	content := string(data)

	// Verify dates are formatted without subseconds
	assert.Contains(t, content, "created_at: \"2025-01-27T19:38:54+09:00\"")
	assert.Contains(t, content, "started_at: \"2025-01-27T19:43:54+09:00\"")
	assert.Contains(t, content, "closed_at: \"2025-01-27T19:48:54+09:00\"")

	// Should NOT contain nanoseconds
	assert.False(t, strings.Contains(content, ".927166"))
}
