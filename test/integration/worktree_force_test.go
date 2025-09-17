package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestStartTicketWithForceFlag(t *testing.T) {
	// Cannot run in parallel due to os.Chdir in setupTestRepo

	// Setup test repository
	repoPath := setupTestRepo(t)

	// Initialize ticketflow with worktree enabled
	err := cli.InitCommandWithWorkingDir(context.Background(), repoPath)
	require.NoError(t, err)

	// Enable worktrees in config
	cfg, err := config.Load(repoPath)
	require.NoError(t, err)
	cfg.Worktree.Enabled = true
	cfg.Worktree.BaseDir = "./.worktrees"
	cfg.Worktree.InitCommands = []string{} // No init commands for test
	err = cfg.Save(filepath.Join(repoPath, ".ticketflow.yaml"))
	require.NoError(t, err)

	// Commit config
	gitCmd := git.New(repoPath)
	_, err = gitCmd.Exec(context.Background(), "add", ".ticketflow.yaml", ".gitignore")
	require.NoError(t, err)
	_, err = gitCmd.Exec(context.Background(), "commit", "-m", "Initialize ticketflow with worktrees")
	require.NoError(t, err)

	// Create app instance
	app, err := cli.NewAppWithWorkingDir(context.Background(), t, repoPath)
	require.NoError(t, err)

	// Create a ticket
	_, err = app.NewTicket(context.Background(), "force-worktree-test", "")
	require.NoError(t, err)

	// Get the ticket
	tickets, err := app.Manager.List(context.Background(), ticket.StatusFilterAll)
	require.NoError(t, err)
	require.Len(t, tickets, 1)
	ticketID := tickets[0].ID

	// Start the ticket (creates worktree)
	_, err = app.StartTicket(context.Background(), ticketID, false)
	require.NoError(t, err)

	// Verify worktree exists
	worktreePath := filepath.Join(repoPath, ".worktrees", ticketID)
	assert.DirExists(t, worktreePath)

	// Create a file in the worktree to verify it gets removed
	testFile := filepath.Join(worktreePath, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Try to start again without force (should fail with suggestion to use --force)
	_, err = app.StartTicket(context.Background(), ticketID, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Ticket already started")

	// Verify the error is a CLIError with the --force suggestion
	if cliErr, ok := err.(*cli.CLIError); ok {
		assert.Equal(t, cli.ErrTicketAlreadyStarted, cliErr.Code)
		assert.Len(t, cliErr.Suggestions, 1)
		assert.Contains(t, cliErr.Suggestions[0], "Use --force to recreate the worktree for this ticket")
	} else {
		t.Fatal("Expected error to be a CLIError")
	}

	// Verify test file exists before force recreation
	assert.FileExists(t, testFile)

	// Start with force flag (should succeed)
	result, err := app.StartTicket(context.Background(), ticketID, true)
	assert.NoError(t, err)

	// Verify the result indicates this was a recreation
	assert.Equal(t, ticket.StatusDoing, result.OriginalStatus, "Original status should be 'doing'")
	assert.True(t, result.IsRecreatingWorktree, "IsRecreatingWorktree should be true")

	// Verify worktree was recreated (test file should be gone)
	assert.DirExists(t, worktreePath)
	assert.NoFileExists(t, testFile)

	// Verify ticket is still in doing status
	updatedTicket, err := app.Manager.Get(context.Background(), ticketID)
	require.NoError(t, err)
	assert.Equal(t, "doing", string(updatedTicket.Status()))
}

func TestStartTicketForceWithoutWorktree(t *testing.T) {
	// Cannot run in parallel due to os.Chdir in setupTestRepo

	// Setup test repository
	repoPath := setupTestRepo(t)

	// Initialize ticketflow without worktree
	err := cli.InitCommandWithWorkingDir(context.Background(), repoPath)
	require.NoError(t, err)

	// Ensure worktrees are disabled
	cfg, err := config.Load(repoPath)
	require.NoError(t, err)
	cfg.Worktree.Enabled = false
	err = cfg.Save(filepath.Join(repoPath, ".ticketflow.yaml"))
	require.NoError(t, err)

	// Commit config
	gitCmd := git.New(repoPath)
	_, err = gitCmd.Exec(context.Background(), "add", ".ticketflow.yaml", ".gitignore")
	require.NoError(t, err)
	_, err = gitCmd.Exec(context.Background(), "commit", "-m", "Initialize ticketflow without worktrees")
	require.NoError(t, err)

	// Create app instance
	app, err := cli.NewAppWithWorkingDir(context.Background(), t, repoPath)
	require.NoError(t, err)

	// Create a ticket
	_, err = app.NewTicket(context.Background(), "no-worktree-force-test", "")
	require.NoError(t, err)

	// Get the ticket
	tickets, err := app.Manager.List(context.Background(), ticket.StatusFilterAll)
	require.NoError(t, err)
	require.Len(t, tickets, 1)
	ticketID := tickets[0].ID

	// Commit the new ticket file to avoid uncommitted changes error
	_, err = gitCmd.Exec(context.Background(), "add", ".")
	require.NoError(t, err)
	_, err = gitCmd.Exec(context.Background(), "commit", "-m", "Add test ticket")
	require.NoError(t, err)

	// Start the ticket (no worktree created)
	_, err = app.StartTicket(context.Background(), ticketID, false)
	require.NoError(t, err)

	// Start again with force flag (should fail since worktrees are disabled)
	_, err = app.StartTicket(context.Background(), ticketID, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Ticket already started")

	// Verify error suggests branch mode alternatives (not --force)
	if cliErr, ok := err.(*cli.CLIError); ok {
		assert.Equal(t, cli.ErrTicketAlreadyStarted, cliErr.Code)

		// Should NOT suggest --force when worktrees are disabled
		for _, suggestion := range cliErr.Suggestions {
			assert.NotContains(t, suggestion, "--force", "Should not suggest --force when worktrees are disabled")
		}

		// Should suggest git checkout and status commands
		assert.Len(t, cliErr.Suggestions, 2, "Should have 2 suggestions for branch mode")
		assert.Contains(t, cliErr.Suggestions[0], "git checkout", "Should suggest git checkout")
		assert.Contains(t, cliErr.Suggestions[0], ticketID, "Should include ticket ID in git checkout suggestion")
		assert.Contains(t, cliErr.Suggestions[1], "ticketflow status", "Should suggest ticketflow status")
	} else {
		t.Fatal("Expected error to be a CLIError")
	}

	// Verify no worktree directory was created
	worktreePath := filepath.Join(repoPath, ".worktrees", ticketID)
	assert.NoDirExists(t, worktreePath)
}
