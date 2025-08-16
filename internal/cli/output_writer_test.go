package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestJSONOutputFormatter(t *testing.T) {
	t.Parallel()

	t.Run("PrintResult with Printable", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewJSONOutputFormatter(&buf)

		result := &CleanupResult{
			OrphanedWorktrees: 5,
			StaleBranches:     3,
			Errors:            []string{"warning1"},
		}

		err := w.PrintResult(result)
		assert.NoError(t, err)

		// Verify JSON output
		var data map[string]interface{}
		err = json.Unmarshal(buf.Bytes(), &data)
		assert.NoError(t, err)
		assert.Equal(t, float64(5), data["orphaned_worktrees"])
		assert.Equal(t, float64(3), data["stale_branches"])
		assert.Equal(t, true, data["has_errors"])
	})

	t.Run("PrintResult with non-Printable", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewJSONOutputFormatter(&buf)

		data := map[string]string{"key": "value"}
		err := w.PrintResult(data)
		assert.NoError(t, err)

		// Verify JSON output
		var result map[string]string
		err = json.Unmarshal(buf.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "value", result["key"])
	})

	t.Run("PrintJSON backwards compatibility", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewJSONOutputFormatter(&buf)

		data := map[string]int{"count": 42}
		err := w.PrintJSON(data)
		assert.NoError(t, err)

		var result map[string]int
		err = json.Unmarshal(buf.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, 42, result["count"])
	})
}

func TestTextOutputFormatter(t *testing.T) {
	t.Parallel()

	t.Run("PrintResult with Printable", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewTextOutputFormatter(&buf)

		result := &CleanupResult{
			OrphanedWorktrees: 2,
			StaleBranches:     1,
		}

		err := w.PrintResult(result)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Cleanup Summary")
		assert.Contains(t, output, "Orphaned worktrees removed: 2")
		assert.Contains(t, output, "Stale branches removed: 1")
	})

	t.Run("PrintResult with ticket", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewTextOutputFormatter(&buf)

		tk := &ticket.Ticket{
			ID:          "test-123",
			Priority:    2,
			Description: "Test ticket",
		}

		err := w.PrintResult(tk)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Ticket: test-123")
		assert.Contains(t, output, "Priority: 2")
		assert.Contains(t, output, "Description: Test ticket")
	})

	t.Run("PrintResult with ticket list", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewTextOutputFormatter(&buf)

		tickets := []*ticket.Ticket{
			{ID: "ticket-1", Description: "First"},
			{ID: "ticket-2", Description: "Second"},
		}

		err := w.PrintResult(tickets)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "ticket-1")
		assert.Contains(t, output, "First")
		assert.Contains(t, output, "ticket-2")
		assert.Contains(t, output, "Second")
	})

	t.Run("PrintResult with empty ticket list", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewTextOutputFormatter(&buf)

		tickets := []*ticket.Ticket{}
		err := w.PrintResult(tickets)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "No tickets found")
	})

	t.Run("PrintResult with map", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewTextOutputFormatter(&buf)

		data := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}

		err := w.PrintResult(data)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "key1: value1")
		assert.Contains(t, output, "key2: 42")
	})

	t.Run("PrintResult with unknown type", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewTextOutputFormatter(&buf)

		err := w.PrintResult("plain string")
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "plain string")
	})
}

func TestTextOutputFormatterConcurrency(t *testing.T) {
	var buf bytes.Buffer
	w := NewTextOutputFormatter(&buf)

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Start 10 goroutines writing concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				result := &CleanupResult{
					OrphanedWorktrees: n,
					StaleBranches:     j,
				}
				if err := w.PrintResult(result); err != nil {
					errors <- err
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("Unexpected error during concurrent writes: %v", err)
	}

	// Verify output was written
	output := buf.String()
	assert.NotEmpty(t, output)
	// Should have 1000 occurrences of "Cleanup Summary"
	count := strings.Count(output, "Cleanup Summary")
	assert.Equal(t, 1000, count, "All writes should be recorded")
}

func TestNewOutputFormatter(t *testing.T) {
	t.Parallel()

	t.Run("creates JSON writer for JSON format", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewOutputFormatter(&buf, FormatJSON)

		// Verify it's a JSON writer by checking output format
		data := map[string]string{"test": "value"}
		err := w.PrintResult(data)
		require.NoError(t, err)

		// Should be valid JSON
		var result map[string]string
		err = json.Unmarshal(buf.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "value", result["test"])
	})

	t.Run("creates text writer for text format", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewOutputFormatter(&buf, FormatText)

		// Verify it's a text formatter
		_, ok := w.(*textOutputFormatter)
		assert.True(t, ok, "Should create textOutputFormatter for text format")
	})

	t.Run("creates text writer for unknown format", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewOutputFormatter(&buf, OutputFormat("unknown"))

		// Should default to text formatter
		_, ok := w.(*textOutputFormatter)
		assert.True(t, ok, "Should default to textOutputFormatter for unknown format")
	})
}

func TestLegacyOutputWriter(t *testing.T) {
	t.Parallel()

	t.Run("NewOutputWriter with nil writers", func(t *testing.T) {
		w := NewOutputWriter(nil, nil, FormatJSON)
		assert.NotNil(t, w.stdout)
		assert.NotNil(t, w.stderr)
		assert.Equal(t, FormatJSON, w.format)
	})

	t.Run("GetFormat", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewOutputWriter(&buf, &buf, FormatJSON)
		assert.Equal(t, FormatJSON, w.GetFormat())
	})

	t.Run("PrintResult delegates to result writer", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewOutputWriter(&buf, &buf, FormatJSON)

		data := map[string]int{"count": 5}
		err := w.PrintResult(data)
		assert.NoError(t, err)

		// Should output JSON
		var result map[string]int
		err = json.Unmarshal(buf.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, 5, result["count"])
	})

	t.Run("deprecated Printf and Println", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewOutputWriter(&buf, &buf, FormatText)

		w.Printf("Hello %s", "world")
		assert.Contains(t, buf.String(), "Hello world")

		buf.Reset()
		w.Println("Line", "output")
		assert.Contains(t, buf.String(), "Line output\n")
	})
}

