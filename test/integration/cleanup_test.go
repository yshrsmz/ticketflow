package integration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestCleanupTicketWithForceFlag(t *testing.T) {
	// Setup test repository
	repoPath := setupTestRepo(t)
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(repoPath)

	// Initialize ticketflow
	err := cli.InitCommand()
	require.NoError(t, err)

	// Create and test the app
	app, err := cli.NewApp()
	require.NoError(t, err)

	// Create a ticket
	err = app.NewTicket("test-cleanup-force", cli.FormatText)
	require.NoError(t, err)

	// List tickets to get the actual ID
	tickets, err := app.Manager.List("todo")
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

	// Start the ticket (creates worktree)
	err = app.StartTicket(ticketID)
	require.NoError(t, err)

	// Get the ticket to verify it exists
	tkt, err := app.Manager.Get(ticketID)
	require.NoError(t, err)
	assert.Equal(t, ticket.StatusDoing, tkt.Status())

	// Close the ticket to move it to done status
	err = app.CloseTicket(true) // force close to skip uncommitted changes check
	require.NoError(t, err)

	// Verify ticket is now done
	tkt, err = app.Manager.Get(ticketID)
	require.NoError(t, err)
	assert.Equal(t, ticket.StatusDone, tkt.Status())

	// Test cleanup with force flag - should NOT prompt for confirmation
	// Note: In the actual CLI, the flag order matters: ticketflow cleanup --force <ticket-id>
	err = app.CleanupTicket(ticketID, true)
	require.NoError(t, err)

	// Verify worktree was removed
	wt, err := app.Git.FindWorktreeByBranch(ticketID)
	assert.NoError(t, err)
	assert.Nil(t, wt)

	// Verify branch was deleted
	branches, err := app.Git.Exec("branch", "--list", ticketID)
	assert.NoError(t, err)
	assert.Empty(t, branches)
}

func TestCleanupTicketWithoutForceFlag(t *testing.T) {
	// This test would require mocking stdin to simulate user input
	// For now, we're just testing the force flag functionality
	t.Skip("Testing without force flag requires stdin mocking")
}
