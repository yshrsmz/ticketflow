package commands

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/cli/commands/testharness"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestWorktreeCommand_Execute_List_Integration(t *testing.T) {
	// Cannot run in parallel due to os.Chdir
	env := testharness.NewTestEnvironment(t)

	// Create a ticket and its worktree
	_ = env.CreateTicket("test-ticket", ticket.StatusDoing)
	env.RunGit("add", ".")
	env.RunGit("commit", "-m", "Add test ticket")

	// Create a worktree
	worktreeDir := filepath.Join(filepath.Dir(env.RootDir), "test-worktrees", "test-ticket")
	env.RunGit("worktree", "add", worktreeDir, "-b", "test-ticket")

	// Execute worktree list command
	env.WithWorkingDirectory(t, func() {
		cmd := NewWorktreeCommand()
		err := cmd.Execute(context.Background(), nil, []string{"list"})
		require.NoError(t, err)
	})
}

func TestWorktreeCommand_Execute_List_JSON_Integration(t *testing.T) {
	// Cannot run in parallel due to os.Chdir
	env := testharness.NewTestEnvironment(t)

	// Create a ticket and its worktree
	_ = env.CreateTicket("test-ticket", ticket.StatusDoing)
	env.RunGit("add", ".")
	env.RunGit("commit", "-m", "Add test ticket")

	// Create a worktree
	worktreeDir := filepath.Join(filepath.Dir(env.RootDir), "test-worktrees", "test-ticket")
	env.RunGit("worktree", "add", worktreeDir, "-b", "test-ticket")

	// Capture JSON output
	outputStr := testharness.CaptureOutput(t, func() {
		env.WithWorkingDirectory(t, func() {
			cmd := NewWorktreeCommand()
			err := cmd.Execute(context.Background(), nil, []string{"list", "--format", "json"})
			require.NoError(t, err)
		})
	})

	// Parse and validate JSON properly
	jsonData := testharness.ValidateJSON(t, outputStr)
	
	// Verify structure
	testharness.AssertJSONFieldExists(t, jsonData, "worktrees")
	
	// Get worktrees array
	worktrees, ok := jsonData["worktrees"].([]interface{})
	require.True(t, ok, "worktrees should be an array")
	require.Greater(t, len(worktrees), 0, "should have at least one worktree")
	
	// Validate first worktree fields
	firstWorktree, ok := worktrees[0].(map[string]interface{})
	require.True(t, ok, "worktree entry should be a map")
	testharness.AssertJSONFieldExists(t, firstWorktree, "Path")
	testharness.AssertJSONFieldExists(t, firstWorktree, "Branch")
	testharness.AssertJSONFieldExists(t, firstWorktree, "Head")
}

func TestWorktreeCommand_Execute_Clean_Integration(t *testing.T) {
	// Cannot run in parallel due to os.Chdir
	env := testharness.NewTestEnvironment(t)

	// Create a ticket
	_ = env.CreateTicket("test-ticket", ticket.StatusDone)
	env.RunGit("add", ".")
	env.RunGit("commit", "-m", "Add test ticket")

	// Execute worktree clean command
	env.WithWorkingDirectory(t, func() {
		cmd := NewWorktreeCommand()
		err := cmd.Execute(context.Background(), nil, []string{"clean"})
		require.NoError(t, err)
	})
}

func TestWorktreeCommand_Execute_InvalidSubcommandFlag_Integration(t *testing.T) {
	// Cannot run in parallel due to os.Chdir
	env := testharness.NewTestEnvironment(t)

	// Try to execute list with invalid flag format
	env.WithWorkingDirectory(t, func() {
		cmd := NewWorktreeCommand()
		err := cmd.Execute(context.Background(), nil, []string{"list", "--format", "invalid"})
		// Should get validation error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid format")
	})
}

func TestWorktreeCommand_Execute_SubcommandWithExtraArgs_Integration(t *testing.T) {
	// Cannot run in parallel due to os.Chdir
	env := testharness.NewTestEnvironment(t)

	// Try to execute clean with unexpected arguments
	env.WithWorkingDirectory(t, func() {
		cmd := NewWorktreeCommand()
		err := cmd.Execute(context.Background(), nil, []string{"clean", "extra", "args"})
		// Should get validation error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no arguments")
	})
}

func TestWorktreeCommand_Execute_ListWithInvalidFormat_Integration(t *testing.T) {
	// Cannot run in parallel due to os.Chdir
	env := testharness.NewTestEnvironment(t)

	// Try list with invalid format
	env.WithWorkingDirectory(t, func() {
		cmd := NewWorktreeCommand()
		// Test invalid format value
		err := cmd.Execute(context.Background(), nil, []string{"list", "--format=yaml"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "format")
	})
}
