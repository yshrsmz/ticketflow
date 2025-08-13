package integration

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestStartTicketWithExistingBranch(t *testing.T) {
	// Cannot run in parallel due to os.Chdir

	// This test verifies that when a branch already exists but has diverged
	// from main, StartTicket will detect the divergence and prompt the user.
	// Since we can't provide input in tests, it will fail with an EOF error.

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
	err = cfg.Save(filepath.Join(repoPath, ".ticketflow.yaml"))
	require.NoError(t, err)

	// Commit config
	gitCmd := git.New(repoPath)
	ctx := context.Background()
	_, err = gitCmd.Exec(ctx, "add", ".ticketflow.yaml", ".gitignore")
	require.NoError(t, err)
	_, err = gitCmd.Exec(ctx, "commit", "-m", "Initialize ticketflow")
	require.NoError(t, err)

	// Create app instance
	app, err := cli.NewAppWithWorkingDir(ctx, t, repoPath)
	require.NoError(t, err)

	// Create a test ticket
	err = app.NewTicket(ctx, "existing-branch-test", "", cli.FormatText)
	require.NoError(t, err)

	// Commit the ticket
	_, err = gitCmd.Exec(ctx, "add", "tickets/")
	require.NoError(t, err)
	_, err = gitCmd.Exec(ctx, "commit", "-m", "Add test ticket")
	require.NoError(t, err)

	// Get the ticket ID
	tickets, err := app.Manager.List(ctx, ticket.StatusFilterActive)
	require.NoError(t, err)
	require.Len(t, tickets, 1)
	ticketID := tickets[0].ID

	// Create the branch manually (simulating the scenario where branch exists but worktree doesn't)
	err = gitCmd.CreateBranch(ctx, ticketID)
	require.NoError(t, err)

	// Switch back to main branch
	err = gitCmd.Checkout(ctx, "main")
	require.NoError(t, err)

	// Make a commit on main to ensure the branch will be behind when we start the ticket
	// (StartTicket will make a commit to change ticket status)
	_, err = gitCmd.Exec(ctx, "commit", "--allow-empty", "-m", "Another commit on main")
	require.NoError(t, err)

	// Verify branch exists using git command
	_, err = gitCmd.Exec(ctx, "show-ref", "--verify", "--quiet", "refs/heads/"+ticketID)
	assert.NoError(t, err, "Branch should exist")

	// Verify worktree doesn't exist yet
	hasWorktree, err := gitCmd.HasWorktree(ctx, ticketID)
	require.NoError(t, err)
	assert.False(t, hasWorktree, "Worktree should not exist yet")

	// Now try to start the ticket - this will:
	// 1. Move ticket to "doing" status and commit
	// 2. Try to create worktree with existing branch
	// 3. Detect that branch is behind main (missing status change commit)
	// 4. In non-interactive mode, automatically recreate the branch
	err = app.StartTicket(ctx, ticketID, false, cli.FormatText)
	require.NoError(t, err, "StartTicket should succeed in non-interactive mode")

	// Verify worktree was created successfully
	hasWorktree, err = gitCmd.HasWorktree(ctx, ticketID)
	require.NoError(t, err)
	assert.True(t, hasWorktree, "Worktree should exist after successful start")
}

func TestStartTicketWithExistingBranchAndWorktree(t *testing.T) {
	// Cannot use t.Parallel() - TestStartTicketWithExistingBranch in same file has comment about os.Chdir

	// Setup test repository
	repoPath := setupTestRepo(t)

	// Initialize ticketflow
	err := cli.InitCommandWithWorkingDir(context.Background(), repoPath)
	require.NoError(t, err)

	// Enable worktrees in config
	cfg, err := config.Load(repoPath)
	require.NoError(t, err)
	cfg.Worktree.Enabled = true
	cfg.Worktree.BaseDir = "./.worktrees"
	err = cfg.Save(filepath.Join(repoPath, ".ticketflow.yaml"))
	require.NoError(t, err)

	// Commit config
	gitCmd := git.New(repoPath)
	ctx := context.Background()
	_, err = gitCmd.Exec(ctx, "add", ".ticketflow.yaml", ".gitignore")
	require.NoError(t, err)
	_, err = gitCmd.Exec(ctx, "commit", "-m", "Initialize ticketflow")
	require.NoError(t, err)

	// Create app instance
	app, err := cli.NewAppWithWorkingDir(ctx, t, repoPath)
	require.NoError(t, err)

	// Create a test ticket
	err = app.NewTicket(ctx, "existing-both-test", "", cli.FormatText)
	require.NoError(t, err)

	// Commit the ticket
	_, err = gitCmd.Exec(ctx, "add", "tickets/")
	require.NoError(t, err)
	_, err = gitCmd.Exec(ctx, "commit", "-m", "Add test ticket")
	require.NoError(t, err)

	// Get the ticket ID
	tickets, err := app.Manager.List(ctx, ticket.StatusFilterActive)
	require.NoError(t, err)
	require.Len(t, tickets, 1)
	ticketID := tickets[0].ID

	// Start the ticket normally first time
	err = app.StartTicket(ctx, ticketID, false, cli.FormatText)
	require.NoError(t, err)

	// Try to start the same ticket again - should fail
	err = app.StartTicket(ctx, ticketID, false, cli.FormatText)
	require.Error(t, err)

	// The error is "ticket already started" which happens before worktree check
	// This is expected behavior - the ticket status check happens first
	assert.Contains(t, err.Error(), "already")

	// For CLI errors, check that it's a ticket error
	if cliErr, ok := err.(*cli.CLIError); ok {
		// Either TICKET_ALREADY_STARTED or WORKTREE_EXISTS are valid
		assert.True(t, cliErr.Code == cli.ErrTicketAlreadyStarted || cliErr.Code == cli.ErrWorktreeExists,
			"Expected ticket already started or worktree exists error, got: %s", cliErr.Code)
	}
}
