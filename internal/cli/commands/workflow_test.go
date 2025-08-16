package commands

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowCommand_Metadata(t *testing.T) {
	cmd := NewWorkflowCommand()

	assert.Equal(t, "workflow", cmd.Name())
	assert.Nil(t, cmd.Aliases())
	assert.Equal(t, "Print development workflow guide", cmd.Description())
	assert.Equal(t, "workflow", cmd.Usage())
}

func TestWorkflowCommand_SetupFlags(t *testing.T) {
	cmd := NewWorkflowCommand()
	result := cmd.SetupFlags(nil)
	assert.Nil(t, result)
}

func TestWorkflowCommand_Validate(t *testing.T) {
	cmd := NewWorkflowCommand()

	// Should accept no arguments
	err := cmd.Validate(nil, []string{})
	assert.NoError(t, err)

	// Should also accept arguments (they're ignored)
	err = cmd.Validate(nil, []string{"arg1", "arg2"})
	assert.NoError(t, err)
}

func TestWorkflowCommand_Execute(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := &WorkflowCommand{output: &buf}

		err := cmd.Execute(context.Background(), nil, nil)
		require.NoError(t, err)

		output := buf.String()

		// Verify the output starts with the expected header
		assert.True(t, strings.HasPrefix(output, "# TicketFlow Development Workflow"))

		// Verify key sections are present
		assert.Contains(t, output, "## Overview")
		assert.Contains(t, output, "## Development Workflow for New Features")
		assert.Contains(t, output, "### 1. Create a Feature Ticket")
		assert.Contains(t, output, "### 2. Start Work on the Ticket")
		assert.Contains(t, output, "### 3. Navigate to the Worktree")

		// Verify commands are present
		assert.Contains(t, output, "ticketflow new")
		assert.Contains(t, output, "ticketflow start")
		assert.Contains(t, output, "ticketflow close")
		assert.Contains(t, output, "ticketflow cleanup")

		// Verify it's valid markdown
		assert.Contains(t, output, "```bash")
		assert.Contains(t, output, "```")

		// Verify AI integration section
		assert.Contains(t, output, "## Integration with AI Tools")
		assert.Contains(t, output, "ticketflow workflow > CLAUDE.md")
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel the context immediately

		var buf bytes.Buffer
		cmd := &WorkflowCommand{output: &buf}

		err := cmd.Execute(ctx, nil, nil)
		assert.ErrorIs(t, err, context.Canceled)

		// Should not write anything when context is cancelled
		assert.Empty(t, buf.String())
	})

	t.Run("output write error", func(t *testing.T) {
		// Create a writer that always fails
		failWriter := &failingWriter{}
		cmd := &WorkflowCommand{output: failWriter}

		err := cmd.Execute(context.Background(), nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write workflow content")
	})
}

func TestWorkflowCommand_ContentStructure(t *testing.T) {
	var buf bytes.Buffer
	cmd := &WorkflowCommand{output: &buf}

	err := cmd.Execute(context.Background(), nil, nil)
	require.NoError(t, err)

	output := buf.String()
	lines := strings.Split(output, "\n")

	// Verify the document has reasonable structure
	assert.Greater(t, len(lines), 100, "Workflow content should have substantial content")

	// Count sections
	sectionCount := 0
	subsectionCount := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			sectionCount++
		}
		if strings.HasPrefix(line, "### ") {
			subsectionCount++
		}
	}

	assert.Greater(t, sectionCount, 5, "Should have multiple main sections")
	assert.Greater(t, subsectionCount, 10, "Should have multiple subsections")

	// Verify code blocks are properly closed
	codeBlockStarts := strings.Count(output, "```")
	assert.Equal(t, 0, codeBlockStarts%2, "Code blocks should be properly closed")
}

// failingWriter is a test helper that always returns an error when writing
type failingWriter struct{}

func (f *failingWriter) Write(p []byte) (n int, err error) {
	return 0, assert.AnError
}
