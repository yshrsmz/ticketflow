package commands

import (
	"bytes"
	"context"
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/cli/commands/testharness"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestCleanupCommand_Execute_AutoCleanup_Integration(t *testing.T) {
	// Cannot run in parallel due to os.Chdir
	env := testharness.NewTestEnvironment(t)

	// Create some done tickets
	_ = env.CreateTicket("done-ticket-1", ticket.StatusDone)
	_ = env.CreateTicket("done-ticket-2", ticket.StatusDone)
	env.RunGit("add", ".")
	env.RunGit("commit", "-m", "Add done tickets")

	// Run cleanup command in auto mode with text output
	env.WithWorkingDirectory(t, func() {
		cmd := NewCleanupCommand()
		fs := flag.NewFlagSet("cleanup", flag.ContinueOnError)
		flags := cmd.SetupFlags(fs).(*cleanupFlags)
		flags.format = "text"

		// Should execute without errors even if nothing to clean
		err := cmd.Execute(context.Background(), flags, []string{})
		require.NoError(t, err)
	})
}

func TestCleanupCommand_Execute_AutoCleanup_DryRun_Integration(t *testing.T) {
	// Cannot run in parallel due to os.Chdir
	env := testharness.NewTestEnvironment(t)

	// Create some tickets
	_ = env.CreateTicket("test-ticket", ticket.StatusDone)
	env.RunGit("add", ".")
	env.RunGit("commit", "-m", "Add test ticket")

	// Run cleanup command in dry-run mode
	env.WithWorkingDirectory(t, func() {
		cmd := NewCleanupCommand()
		fs := flag.NewFlagSet("cleanup", flag.ContinueOnError)
		flags := cmd.SetupFlags(fs).(*cleanupFlags)
		flags.dryRun = true
		flags.format = "text"

		// Should execute without errors in dry-run mode
		err := cmd.Execute(context.Background(), flags, []string{})
		require.NoError(t, err)
	})
}

func TestCleanupCommand_Execute_AutoCleanup_JSON_Integration(t *testing.T) {
	// Cannot run in parallel due to os.Chdir
	env := testharness.NewTestEnvironment(t)

	// Create some tickets
	_ = env.CreateTicket("test-ticket", ticket.StatusDone)
	env.RunGit("add", ".")
	env.RunGit("commit", "-m", "Add test ticket")

	// Capture JSON output
	var output bytes.Buffer
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	env.WithWorkingDirectory(t, func() {
		cmd := NewCleanupCommand()
		fs := flag.NewFlagSet("cleanup", flag.ContinueOnError)
		flags := cmd.SetupFlags(fs).(*cleanupFlags)
		flags.format = "json"

		err := cmd.Execute(context.Background(), flags, []string{})
		require.NoError(t, err)
	})

	w.Close()
	_, _ = io.Copy(&output, r)
	os.Stdout = origStdout

	// Verify JSON output structure
	outputStr := output.String()
	assert.Contains(t, outputStr, `"success":`)
	assert.Contains(t, outputStr, `"orphaned_worktrees":`)
	assert.Contains(t, outputStr, `"stale_branches":`)
}

func TestCleanupCommand_Execute_TicketCleanup_Integration(t *testing.T) {
	// Cannot run in parallel due to os.Chdir
	env := testharness.NewTestEnvironment(t)

	// Create a done ticket
	_ = env.CreateTicket("done-ticket", ticket.StatusDone)
	env.RunGit("add", ".")
	env.RunGit("commit", "-m", "Add done ticket")

	// Create a worktree for the ticket
	worktreeDir := filepath.Join(filepath.Dir(env.RootDir), "test-worktrees", "done-ticket")
	env.RunGit("worktree", "add", worktreeDir, "-b", "done-ticket")

	// Run cleanup for specific ticket with force flag to skip confirmation
	env.WithWorkingDirectory(t, func() {
		cmd := NewCleanupCommand()
		fs := flag.NewFlagSet("cleanup", flag.ContinueOnError)
		flags := cmd.SetupFlags(fs).(*cleanupFlags)
		flags.format = "text"
		flags.force = true // Skip confirmation
		flags.args = []string{"done-ticket"}

		err := cmd.Execute(context.Background(), flags, []string{"done-ticket"})
		require.NoError(t, err)

		// Verify worktree was removed
		worktrees := env.RunGit("worktree", "list")
		assert.NotContains(t, worktrees, "done-ticket")

		// Verify branch was removed
		branches := env.RunGit("branch", "-a")
		assert.NotContains(t, branches, "done-ticket")
	})
}

func TestCleanupCommand_Execute_TicketCleanup_Force_Integration(t *testing.T) {
	// Cannot run in parallel due to os.Chdir
	env := testharness.NewTestEnvironment(t)

	// Create a done ticket
	_ = env.CreateTicket("done-ticket-force", ticket.StatusDone)
	env.RunGit("add", ".")
	env.RunGit("commit", "-m", "Add done ticket")

	// Create a worktree for the ticket
	worktreeDir := filepath.Join(filepath.Dir(env.RootDir), "test-worktrees", "done-ticket-force")
	env.RunGit("worktree", "add", worktreeDir, "-b", "done-ticket-force")

	// Make a change in the worktree to simulate uncommitted work
	worktreeFile := filepath.Join(worktreeDir, "uncommitted.txt")
	require.NoError(t, os.WriteFile(worktreeFile, []byte("uncommitted changes"), 0644))

	// Try cleanup with force flag
	env.WithWorkingDirectory(t, func() {
		cmd := NewCleanupCommand()
		fs := flag.NewFlagSet("cleanup", flag.ContinueOnError)
		flags := cmd.SetupFlags(fs).(*cleanupFlags)
		flags.force = true
		flags.format = "text"
		flags.args = []string{"done-ticket-force"}

		err := cmd.Execute(context.Background(), flags, []string{"done-ticket-force"})
		require.NoError(t, err)

		// Verify worktree was removed even with uncommitted changes
		worktrees := env.RunGit("worktree", "list")
		assert.NotContains(t, worktrees, "done-ticket-force")
	})
}

func TestCleanupCommand_Execute_TicketCleanup_JSON_Integration(t *testing.T) {
	// Cannot run in parallel due to os.Chdir
	env := testharness.NewTestEnvironment(t)

	// Create a done ticket
	_ = env.CreateTicket("done-ticket", ticket.StatusDone,
		testharness.WithDescription("Test ticket for cleanup"))
	env.RunGit("add", ".")
	env.RunGit("commit", "-m", "Add done ticket")

	// Create a worktree for the ticket
	worktreeDir := filepath.Join(filepath.Dir(env.RootDir), "test-worktrees", "done-ticket")
	env.RunGit("worktree", "add", worktreeDir, "-b", "done-ticket")

	// Capture JSON output
	var output bytes.Buffer
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	env.WithWorkingDirectory(t, func() {
		cmd := NewCleanupCommand()
		fs := flag.NewFlagSet("cleanup", flag.ContinueOnError)
		flags := cmd.SetupFlags(fs).(*cleanupFlags)
		flags.format = "json"
		flags.force = true // Skip confirmation
		flags.args = []string{"done-ticket"}

		err := cmd.Execute(context.Background(), flags, []string{"done-ticket"})
		require.NoError(t, err)
	})

	w.Close()
	_, _ = io.Copy(&output, r)
	os.Stdout = origStdout

	// Verify JSON output structure
	outputStr := output.String()
	assert.Contains(t, outputStr, `"success": true`)
	assert.Contains(t, outputStr, `"ticket":`)
	assert.Contains(t, outputStr, `"id": "done-ticket"`)
	assert.Contains(t, outputStr, `"description": "Test ticket for cleanup"`)
}

func TestCleanupCommand_Execute_TicketCleanup_NotFound_Integration(t *testing.T) {
	// Cannot run in parallel due to os.Chdir
	env := testharness.NewTestEnvironment(t)

	// Try to cleanup non-existent ticket
	env.WithWorkingDirectory(t, func() {
		cmd := NewCleanupCommand()
		fs := flag.NewFlagSet("cleanup", flag.ContinueOnError)
		flags := cmd.SetupFlags(fs).(*cleanupFlags)
		flags.format = "text"
		flags.args = []string{"non-existent-ticket"}

		err := cmd.Execute(context.Background(), flags, []string{"non-existent-ticket"})
		require.Error(t, err)
		// Just check that we got an error about ticket not found
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestCleanupCommand_Execute_TicketCleanup_ErrorJSON_Integration(t *testing.T) {
	// Cannot run in parallel due to os.Chdir
	env := testharness.NewTestEnvironment(t)

	// Capture JSON error output
	var output bytes.Buffer
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	env.WithWorkingDirectory(t, func() {
		cmd := NewCleanupCommand()
		fs := flag.NewFlagSet("cleanup", flag.ContinueOnError)
		flags := cmd.SetupFlags(fs).(*cleanupFlags)
		flags.format = "json"
		flags.args = []string{"non-existent-ticket"}

		// This will error since ticket doesn't exist
		_ = cmd.Execute(context.Background(), flags, []string{"non-existent-ticket"})
	})

	w.Close()
	_, _ = io.Copy(&output, r)
	os.Stdout = origStdout

	// Verify JSON error output structure
	outputStr := output.String()
	assert.Contains(t, outputStr, `"success": false`)
	assert.Contains(t, outputStr, `"error":`)
}

func TestCleanupCommand_Execute_AutoCleanup_NoConfig_Integration(t *testing.T) {
	// Cannot run in parallel due to os.Chdir
	tmpDir := t.TempDir()

	// Change to temp dir without ticketflow config
	origWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(origWd))
	}()
	require.NoError(t, os.Chdir(tmpDir))

	// Try to run cleanup without config
	cmd := NewCleanupCommand()
	fs := flag.NewFlagSet("cleanup", flag.ContinueOnError)
	flags := cmd.SetupFlags(fs).(*cleanupFlags)
	flags.format = "text"

	err = cmd.Execute(context.Background(), flags, []string{})
	require.Error(t, err)
	// Should error about missing config or git repo
}
