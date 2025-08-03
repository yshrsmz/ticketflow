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

func TestStartTicketWithExistingBranch(t *testing.T) {
	// Setup test repository
	repoPath := setupTestRepo(t)
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalWd)
		require.NoError(t, err)
	}()
	err = os.Chdir(repoPath)
	require.NoError(t, err)

	// Initialize ticketflow with worktree enabled
	err = cli.InitCommand(context.Background())
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
	app, err := cli.NewApp(ctx)
	require.NoError(t, err)

	// Create a test ticket
	err = app.NewTicket(ctx, "existing-branch-test", cli.FormatText)
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

	// Verify branch exists using git command (since BranchExists is not in interface)
	_, err = gitCmd.Exec(ctx, "show-ref", "--verify", "--quiet", "refs/heads/"+ticketID)
	// Command returns error if branch doesn't exist, which is what we check
	assert.NoError(t, err, "Branch should exist")

	// Verify worktree doesn't exist yet
	hasWorktree, err := gitCmd.HasWorktree(ctx, ticketID)
	require.NoError(t, err)
	assert.False(t, hasWorktree, "Worktree should not exist yet")

	// Now try to start the ticket - this should succeed even though branch exists
	err = app.StartTicket(ctx, ticketID)
	require.NoError(t, err, "Starting ticket with existing branch should succeed")

	// Verify worktree was created
	hasWorktree, err = gitCmd.HasWorktree(ctx, ticketID)
	require.NoError(t, err)
	assert.True(t, hasWorktree, "Worktree should exist after starting ticket")

	// Verify worktree is on the correct branch
	worktrees, err := gitCmd.ListWorktrees(ctx)
	require.NoError(t, err)

	var foundWorktree *git.WorktreeInfo
	for _, wt := range worktrees {
		if wt.Branch == ticketID {
			foundWorktree = &wt
			break
		}
	}
	require.NotNil(t, foundWorktree, "Should find worktree for ticket")
	assert.Equal(t, ticketID, foundWorktree.Branch)

	// Verify ticket status changed to doing
	updatedTicket, err := app.Manager.Get(ctx, ticketID)
	require.NoError(t, err)
	assert.Equal(t, "doing", string(updatedTicket.Status()))
}

func TestStartTicketWithExistingBranchAndWorktree(t *testing.T) {
	// Setup test repository
	repoPath := setupTestRepo(t)
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalWd)
		require.NoError(t, err)
	}()
	err = os.Chdir(repoPath)
	require.NoError(t, err)

	// Initialize ticketflow
	err = cli.InitCommand(context.Background())
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
	app, err := cli.NewApp(ctx)
	require.NoError(t, err)

	// Create a test ticket
	err = app.NewTicket(ctx, "existing-both-test", cli.FormatText)
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
	err = app.StartTicket(ctx, ticketID)
	require.NoError(t, err)

	// Try to start the same ticket again - should fail
	err = app.StartTicket(ctx, ticketID)
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
