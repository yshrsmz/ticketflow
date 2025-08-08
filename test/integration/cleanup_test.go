package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestCleanupTicketWithForceFlag(t *testing.T) {
	t.Parallel()

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

func TestCleanupTicketWithoutForceFlag(t *testing.T) {
	// This test would require mocking stdin to simulate user input
	// For now, we're just testing the force flag functionality
	t.Skip("Testing without force flag requires stdin mocking")
}
