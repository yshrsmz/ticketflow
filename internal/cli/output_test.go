package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseOutputFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected OutputFormat
	}{
		{"json", FormatJSON},
		{"JSON", FormatJSON},
		{"text", FormatText},
		{"TEXT", FormatText},
		{"", FormatText},
		{"invalid", FormatText},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseOutputFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{0, "0m"},
		{-1 * time.Hour, "0s"},
		{30 * time.Minute, "30m"},
		{1 * time.Hour, "1h"},
		{90 * time.Minute, "1h 30m"},
		{25 * time.Hour, "1d 1h"},
		{49 * time.Hour, "2d 1h"},
		{24*time.Hour + 30*time.Minute, "1d 30m"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}
