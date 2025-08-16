package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAppWithFormat(t *testing.T) {
	ctx := context.Background()

	t.Run("creates app with JSON format", func(t *testing.T) {
		app, err := NewAppWithFormat(ctx, FormatJSON)
		require.NoError(t, err)
		require.NotNil(t, app)

		// Verify StatusWriter is null for JSON
		_, ok := app.StatusWriter.(*nullStatusWriter)
		assert.True(t, ok, "Should use nullStatusWriter for JSON format")

		// Verify Output format
		assert.Equal(t, FormatJSON, app.Output.GetFormat())

		// Test that status messages are suppressed
		app.StatusWriter.Printf("This should not appear")
		app.StatusWriter.Println("Neither should this")

		// Test that JSON output works
		var buf bytes.Buffer
		app.Output = NewOutputWriter(&buf, nil, FormatJSON)

		data := map[string]string{"test": "value"}
		err = app.Output.PrintResult(data)
		require.NoError(t, err)

		var result map[string]string
		err = json.Unmarshal(buf.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "value", result["test"])
	})

	t.Run("creates app with text format", func(t *testing.T) {
		app, err := NewAppWithFormat(ctx, FormatText)
		require.NoError(t, err)
		require.NotNil(t, app)

		// Verify StatusWriter is text for text format
		_, ok := app.StatusWriter.(*textStatusWriter)
		assert.True(t, ok, "Should use textStatusWriter for text format")

		// Verify Output format
		assert.Equal(t, FormatText, app.Output.GetFormat())

		// Test that status messages work
		var statusBuf bytes.Buffer
		app.StatusWriter = NewTextStatusWriter(&statusBuf)
		app.StatusWriter.Printf("Status: %s", "working")
		assert.Contains(t, statusBuf.String(), "Status: working")
	})

	t.Run("creates app with unknown format defaults to text", func(t *testing.T) {
		app, err := NewAppWithFormat(ctx, OutputFormat("unknown"))
		require.NoError(t, err)
		require.NotNil(t, app)

		// Should default to text behavior
		_, ok := app.StatusWriter.(*textStatusWriter)
		assert.True(t, ok, "Should default to textStatusWriter for unknown format")
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		t.Skip("Skipping test that requires git repository")
		// This test would need a test environment with a git repository
		// to properly test context cancellation
	})
}

func TestNewAppWithOptions(t *testing.T) {
	ctx := context.Background()

	t.Run("creates app with custom options", func(t *testing.T) {
		var statusBuf, outputBuf bytes.Buffer

		app, err := NewAppWithOptions(ctx,
			WithOutput(NewOutputWriter(&outputBuf, nil, FormatJSON)),
			WithStatusWriter(NewTextStatusWriter(&statusBuf)),
		)
		require.NoError(t, err)
		require.NotNil(t, app)

		// Test status writer
		app.StatusWriter.Printf("test status")
		assert.Contains(t, statusBuf.String(), "test status")

		// Test output writer
		data := map[string]int{"value": 42}
		err = app.Output.PrintResult(data)
		require.NoError(t, err)

		var result map[string]int
		err = json.Unmarshal(outputBuf.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, 42, result["value"])
	})

	t.Run("creates app with nil options uses defaults", func(t *testing.T) {
		app, err := NewAppWithOptions(ctx)
		require.NoError(t, err)
		require.NotNil(t, app)

		// Should have default writers
		assert.NotNil(t, app.StatusWriter)
		assert.NotNil(t, app.Output)
	})
}

func TestWithOutput(t *testing.T) {
	var buf bytes.Buffer
	writer := NewOutputWriter(&buf, nil, FormatJSON)

	opt := WithOutput(writer)
	assert.NotNil(t, opt)

	// Apply option to empty app
	app := &App{}
	opt(app)

	assert.Equal(t, writer, app.Output)
}

func TestWithStatusWriter(t *testing.T) {
	var buf bytes.Buffer
	writer := NewTextStatusWriter(&buf)

	opt := WithStatusWriter(writer)
	assert.NotNil(t, opt)

	// Apply option to empty app
	app := &App{}
	opt(app)

	assert.Equal(t, writer, app.StatusWriter)
}
