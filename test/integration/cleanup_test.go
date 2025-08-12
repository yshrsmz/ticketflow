package integration

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestCleanupTicketWithForceFlag(t *testing.T) {
	// Cannot use t.Parallel() - TestCleanupTicketWithWorktreeAndForceFlag in same file uses os.Chdir

	// Setup test repository
	repoPath := setupTestRepo(t)

	// Initialize ticketflow with worktrees disabled for this test
	// This test is about the force flag for cleanup, not about worktree symlink behavior
	err := cli.InitCommandWithWorkingDir(context.Background(), repoPath)
	require.NoError(t, err)

	// Create and test the app with custom config
	app, err := cli.NewAppWithWorkingDir(context.Background(), t, repoPath)
	require.NoError(t, err)

	// Disable worktrees for this test to simplify the workflow
	app.Config.Worktree.Enabled = false

	// Create a ticket
	err = app.NewTicket(context.Background(), "test-cleanup-force", "", cli.FormatText)
	require.NoError(t, err)

	// List tickets to get the actual ID
	tickets, err := app.Manager.List(context.Background(), ticket.StatusFilterTodo)
	require.NoError(t, err)
	require.NotEmpty(t, tickets)

	// Find our ticket
	var ticketID string
	for _, t := range tickets {
		if t.Slug == "test-cleanup-force" {
			ticketID = t.ID
			break
		}
	}
	require.NotEmpty(t, ticketID, "Could not find created ticket")

	// Commit the ticket file
	err = app.Git.Add(context.Background(), ".")
	require.NoError(t, err)
	err = app.Git.Commit(context.Background(), "Add ticket: test-cleanup-force")
	require.NoError(t, err)

	// Start the ticket (creates branch but no worktree since we disabled it)
	err = app.StartTicket(context.Background(), ticketID, false)
	require.NoError(t, err)

	// Get the ticket to verify it exists
	tkt, err := app.Manager.Get(context.Background(), ticketID)
	require.NoError(t, err)
	assert.Equal(t, ticket.StatusDoing, tkt.Status())

	// Close the ticket to move it to done status
	err = app.CloseTicket(context.Background(), true) // force close to skip uncommitted changes check
	require.NoError(t, err)

	// Verify ticket is now done
	tkt, err = app.Manager.Get(context.Background(), ticketID)
	require.NoError(t, err)
	assert.Equal(t, ticket.StatusDone, tkt.Status())

	// Test cleanup with force flag - should NOT prompt for confirmation
	// Note: In the actual CLI, the flag order matters: ticketflow cleanup --force <ticket-id>
	err = app.CleanupTicket(context.Background(), ticketID, true)
	require.NoError(t, err)

	// Verify branch was deleted
	branches, err := app.Git.Exec(context.Background(), "branch", "--list", ticketID)
	assert.NoError(t, err)
	assert.Empty(t, branches)
}

func TestCleanupTicketWithWorktreeAndForceFlag(t *testing.T) {
	// Cannot use t.Parallel() because this test uses os.Chdir to navigate to worktree
	// This test verifies cleanup behavior specifically with worktrees enabled,
	// complementing TestCleanupTicketWithForceFlag which tests non-worktree mode.

	// Setup test repository
	repoPath := setupTestRepo(t)

	// Initialize ticketflow with worktree enabled
	err := cli.InitCommandWithWorkingDir(context.Background(), repoPath)
	require.NoError(t, err)

	// Create and test the app with custom config
	app, err := cli.NewAppWithWorkingDir(context.Background(), t, repoPath)
	require.NoError(t, err)

	// Explicitly enable worktrees for this test
	app.Config.Worktree.Enabled = true
	app.Config.Worktree.BaseDir = "./.worktrees"
	app.Config.Worktree.InitCommands = []string{} // No init commands for test

	// Create a ticket
	err = app.NewTicket(context.Background(), "test-cleanup-worktree", "", cli.FormatText)
	require.NoError(t, err)

	// List tickets to get the actual ID
	tickets, err := app.Manager.List(context.Background(), ticket.StatusFilterTodo)
	require.NoError(t, err)
	require.NotEmpty(t, tickets)

	// Find our ticket
	var ticketID string
	for _, t := range tickets {
		if t.Slug == "test-cleanup-worktree" {
			ticketID = t.ID
			break
		}
	}
	require.NotEmpty(t, ticketID, "Could not find created ticket")

	// Commit the ticket file
	err = app.Git.Add(context.Background(), ".")
	require.NoError(t, err)
	err = app.Git.Commit(context.Background(), "Add ticket: test-cleanup-worktree")
	require.NoError(t, err)

	// Start the ticket (creates branch AND worktree since we enabled it)
	err = app.StartTicket(context.Background(), ticketID, false)
	require.NoError(t, err)

	// Verify worktree was created
	worktrees, err := app.Git.ListWorktrees(context.Background())
	require.NoError(t, err)
	assert.Len(t, worktrees, 2, "Expected 2 worktrees (main + ticket)")

	// Find the ticket worktree
	var ticketWorktree *git.WorktreeInfo
	for _, wt := range worktrees {
		if wt.Branch == ticketID {
			ticketWorktree = &wt
			break
		}
	}
	require.NotNil(t, ticketWorktree, "Ticket worktree should exist")

	// Verify worktree directory exists
	_, err = os.Stat(ticketWorktree.Path)
	require.NoError(t, err, "Worktree directory should exist")

	// Get the ticket to verify it exists
	tkt, err := app.Manager.Get(context.Background(), ticketID)
	require.NoError(t, err)
	assert.Equal(t, ticket.StatusDoing, tkt.Status())

	// Close the ticket from within the worktree (required when worktrees are enabled)
	// Change to worktree directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(ticketWorktree.Path)
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	// Create app instance from within worktree
	wtApp, err := cli.NewApp(context.Background())
	require.NoError(t, err)

	// Close the ticket to move it to done status
	err = wtApp.CloseTicket(context.Background(), true) // force close to skip uncommitted changes check
	require.NoError(t, err)

	// Change back to original directory
	err = os.Chdir(originalWd)
	require.NoError(t, err)

	// Simulate PR merge by merging the branch into main
	// This brings the ticket file changes (todo -> doing -> done) to main
	// In a real workflow, this would happen after the PR is merged on GitHub/GitLab
	err = app.Git.Checkout(context.Background(), "main")
	require.NoError(t, err)
	_, err = app.Git.Exec(context.Background(), "merge", "--no-ff", ticketID)
	require.NoError(t, err)

	// Now verify ticket is done from main's perspective
	tkt, err = app.Manager.Get(context.Background(), ticketID)
	require.NoError(t, err)
	assert.Equal(t, ticket.StatusDone, tkt.Status())

	// Test cleanup with force flag - should NOT prompt for confirmation
	// This should remove both the worktree and the branch
	err = app.CleanupTicket(context.Background(), ticketID, true)
	require.NoError(t, err)

	// Verify worktree was removed
	worktrees, err = app.Git.ListWorktrees(context.Background())
	require.NoError(t, err)
	assert.Len(t, worktrees, 1, "Only main worktree should remain after cleanup")

	// Verify worktree directory was removed
	_, err = os.Stat(ticketWorktree.Path)
	assert.True(t, os.IsNotExist(err), "Worktree directory should be removed")

	// Verify branch was deleted
	branches, err := app.Git.Exec(context.Background(), "branch", "--list", ticketID)
	assert.NoError(t, err)
	assert.Empty(t, branches, "Branch should be deleted")
}

func TestCleanupTicketWithoutForceFlag(t *testing.T) {
	// This test would require mocking stdin to simulate user input
	// For now, we're just testing the force flag functionality
	t.Skip("Testing without force flag requires stdin mocking")
}
