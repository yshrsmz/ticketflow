package commands

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/cli/commands/testharness"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestStatusCommand_Execute_TextOutput_Integration(t *testing.T) {
	// Cannot run in parallel due to os.Chdir
	env := testharness.NewTestEnvironment(t)

	// Create a ticket in doing status (current ticket)
	_ = env.CreateTicket("current-ticket", ticket.StatusDoing,
		testharness.WithDescription("Test ticket for status"))
	env.RunGit("add", ".")
	env.RunGit("commit", "-m", "Add current ticket")

	// Execute status command with text output
	env.WithWorkingDirectory(t, func() {
		cmd := NewStatusCommand()
		ctx := context.Background()
		flags := &statusFlags{format: "text"}

		err := cmd.Execute(ctx, flags, []string{})
		require.NoError(t, err)
	})
}

func TestStatusCommand_Execute_JSONOutput_Integration(t *testing.T) {
	// Cannot run in parallel due to os.Chdir
	env := testharness.NewTestEnvironment(t)

	// Create a ticket in doing status (current ticket)
	_ = env.CreateTicket("current-ticket", ticket.StatusDoing,
		testharness.WithDescription("Test ticket for JSON"))
	env.RunGit("add", ".")
	env.RunGit("commit", "-m", "Add current ticket")

	// Capture JSON output
	var output bytes.Buffer
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	env.WithWorkingDirectory(t, func() {
		cmd := NewStatusCommand()
		ctx := context.Background()
		flags := &statusFlags{format: "json"}

		err := cmd.Execute(ctx, flags, []string{})
		require.NoError(t, err)
	})

	w.Close()
	_, _ = io.Copy(&output, r)
	os.Stdout = origStdout

	// Verify JSON output structure
	outputStr := output.String()
	assert.Contains(t, outputStr, `"current_ticket":`)
	assert.Contains(t, outputStr, `"id": "current-ticket"`)
	assert.Contains(t, outputStr, `"description": "Test ticket for JSON"`)
	assert.Contains(t, outputStr, `"status": "doing"`)
}

func TestStatusCommand_Execute_NoCurrentTicket_Integration(t *testing.T) {
	// Cannot run in parallel due to os.Chdir
	env := testharness.NewTestEnvironment(t)

	// Don't create any ticket in doing status
	_ = env.CreateTicket("todo-ticket", ticket.StatusTodo)
	env.RunGit("add", ".")
	env.RunGit("commit", "-m", "Add todo ticket")

	// Execute status command - should succeed even with no current ticket
	env.WithWorkingDirectory(t, func() {
		cmd := NewStatusCommand()
		ctx := context.Background()
		flags := &statusFlags{format: "text"}

		// Status command succeeds but shows warning for no active ticket
		err := cmd.Execute(ctx, flags, []string{})
		require.NoError(t, err)
	})
}

func TestStatusCommand_Execute_CancelledContext_Integration(t *testing.T) {
	// Cannot run in parallel due to os.Chdir
	env := testharness.NewTestEnvironment(t)

	// Execute with cancelled context
	env.WithWorkingDirectory(t, func() {
		cmd := NewStatusCommand()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		flags := &statusFlags{format: "text"}

		err := cmd.Execute(ctx, flags, []string{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}
