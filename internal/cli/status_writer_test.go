package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNullStatusWriter(t *testing.T) {
	t.Parallel()
	w := NewNullStatusWriter()

	// Should not panic on any method call
	w.Printf("test %s %d", "message", 123)
	w.Println("test", "multiple", "args")

	// Verify it's actually a nullStatusWriter
	_, ok := w.(*nullStatusWriter)
	assert.True(t, ok, "Should be a nullStatusWriter")
}

func TestTextStatusWriter(t *testing.T) {
	t.Parallel()

	t.Run("Printf", func(t *testing.T) {
		var buf bytes.Buffer
		w := &textStatusWriter{w: &buf}

		w.Printf("Hello %s, number %d", "world", 42)
		assert.Equal(t, "Hello world, number 42", buf.String())
	})

	t.Run("Println", func(t *testing.T) {
		var buf bytes.Buffer
		w := &textStatusWriter{w: &buf}

		w.Println("Hello", "world")
		assert.Equal(t, "Hello world\n", buf.String())
	})

	t.Run("Println with no args", func(t *testing.T) {
		var buf bytes.Buffer
		w := &textStatusWriter{w: &buf}

		w.Println()
		assert.Equal(t, "\n", buf.String())
	})
}

func TestNewStatusWriter(t *testing.T) {
	t.Parallel()

	t.Run("text format", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewStatusWriter(&buf, FormatText)

		// Should be textStatusWriter
		_, ok := w.(*textStatusWriter)
		assert.True(t, ok, "Should create textStatusWriter for text format")

		// Should write to buffer
		w.Printf("test")
		assert.Equal(t, "test", buf.String())
	})

	t.Run("json format", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewStatusWriter(&buf, FormatJSON)

		// Should be nullStatusWriter
		_, ok := w.(*nullStatusWriter)
		assert.True(t, ok, "Should create nullStatusWriter for JSON format")

		// Should not write to buffer
		w.Printf("test")
		assert.Empty(t, buf.String(), "Should not write anything in JSON mode")
	})

	t.Run("unknown format defaults to text", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewStatusWriter(&buf, OutputFormat("unknown"))

		// Should default to textStatusWriter
		_, ok := w.(*textStatusWriter)
		assert.True(t, ok, "Should default to textStatusWriter for unknown format")
	})
}

func TestStatusWriterConcurrency(t *testing.T) {
	// Test that textStatusWriter is safe for concurrent use
	// Note: textStatusWriter doesn't have built-in thread safety,
	// so this test just verifies it doesn't panic during concurrent access
	var buf bytes.Buffer
	w := NewStatusWriter(&buf, FormatText)

	done := make(chan bool, 10)

	// Start 10 goroutines writing concurrently
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 100; j++ {
				w.Printf("goroutine %d iteration %d\n", n, j)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Just verify that some output was written and there was no panic
	output := buf.String()
	assert.NotEmpty(t, output, "Should have written some output")
	assert.Contains(t, output, "goroutine")
}
