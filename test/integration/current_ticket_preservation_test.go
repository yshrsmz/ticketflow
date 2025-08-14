package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestCurrentTicketPreservation(t *testing.T) {
	// Cannot run in parallel - uses os.Chdir
	ctx := context.Background()

	// Setup test repository
	repoPath := setupTestRepo(t)

	// Change to test directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalWd)
		require.NoError(t, err)
	}()
	require.NoError(t, os.Chdir(repoPath))

	// Initialize ticketflow
	err = cli.InitCommandWithWorkingDir(ctx, repoPath)
	require.NoError(t, err)

	// Create app instance
	app, err := cli.NewAppWithWorkingDir(ctx, t, repoPath)
	require.NoError(t, err)

	// Disable worktrees for simpler testing
	app.Config.Worktree.Enabled = false

	t.Run("preserve current-ticket.md when closing different ticket", func(t *testing.T) {
		// Create two test tickets
		// First ticket - will become current
		_, err := app.NewTicket(ctx, "current-ticket", "")
		require.NoError(t, err)

		// Second ticket - will be closed
		_, err = app.NewTicket(ctx, "other-ticket", "")
		require.NoError(t, err)

		// Get the tickets
		tickets, err := app.Manager.List(ctx, ticket.StatusFilterTodo)
		require.NoError(t, err)
		require.Len(t, tickets, 2, "Should have created 2 tickets")

		var currentTicket, otherTicket *ticket.Ticket
		for i := range tickets {
			if tickets[i].Slug == "current-ticket" {
				currentTicket = &tickets[i]
			} else if tickets[i].Slug == "other-ticket" {
				otherTicket = &tickets[i]
			}
		}
		require.NotNil(t, currentTicket, "Could not find current-ticket")
		require.NotNil(t, otherTicket, "Could not find other-ticket")

		// Start the current ticket (this creates current-ticket.md symlink)
		_, err = app.StartTicket(ctx, currentTicket.ID, false)
		require.NoError(t, err)

		// Verify current-ticket.md exists and points to the correct ticket
		currentTicketPath := filepath.Join(repoPath, "current-ticket.md")
		target, err := os.Readlink(currentTicketPath)
		require.NoError(t, err, "current-ticket.md should exist as symlink")
		assert.Contains(t, target, currentTicket.ID, "current-ticket.md should point to current ticket")

		// Close the OTHER ticket (not the current one)
		_, err = app.CloseTicketByID(ctx, otherTicket.ID, "Test reason", false)
		require.NoError(t, err)

		// CRITICAL: Verify current-ticket.md still exists and points to the same ticket
		targetAfterClose, err := os.Readlink(currentTicketPath)
		require.NoError(t, err, "current-ticket.md should still exist after closing different ticket")
		assert.Equal(t, target, targetAfterClose, "current-ticket.md should not have changed")
		assert.Contains(t, targetAfterClose, currentTicket.ID, "current-ticket.md should still point to original current ticket")

		// Verify the other ticket was closed
		closedTicket, err := app.Manager.Get(ctx, otherTicket.ID)
		require.NoError(t, err)
		assert.Equal(t, ticket.StatusDone, closedTicket.Status())
		assert.Equal(t, "Test reason", closedTicket.ClosureReason)
	})

	t.Run("remove current-ticket.md when closing current ticket", func(t *testing.T) {
		// Create a test ticket
		_, err := app.NewTicket(ctx, "ticket-to-close", "")
		require.NoError(t, err)

		// Get the ticket
		tickets, err := app.Manager.List(ctx, ticket.StatusFilterTodo)
		require.NoError(t, err)

		var testTicket *ticket.Ticket
		for i := range tickets {
			if tickets[i].Slug == "ticket-to-close" {
				testTicket = &tickets[i]
				break
			}
		}
		require.NotNil(t, testTicket, "Could not find ticket-to-close")

		// Start the ticket (this creates current-ticket.md symlink)
		_, err = app.StartTicket(ctx, testTicket.ID, false)
		require.NoError(t, err)

		// Verify current-ticket.md exists
		currentTicketPath := filepath.Join(repoPath, "current-ticket.md")
		_, err = os.Readlink(currentTicketPath)
		require.NoError(t, err, "current-ticket.md should exist as symlink")

		// Close the current ticket
		_, err = app.CloseTicketWithReason(ctx, "Completed", false)
		require.NoError(t, err)

		// Verify current-ticket.md was removed
		_, err = os.Readlink(currentTicketPath)
		assert.Error(t, err, "current-ticket.md should be removed when closing current ticket")
		assert.True(t, os.IsNotExist(err), "current-ticket.md should not exist")
	})

	t.Run("handle missing current-ticket.md gracefully", func(t *testing.T) {
		// Create a test ticket
		_, err := app.NewTicket(ctx, "ticket-without-current", "")
		require.NoError(t, err)

		// Get the ticket
		tickets, err := app.Manager.List(ctx, ticket.StatusFilterTodo)
		require.NoError(t, err)

		var testTicket *ticket.Ticket
		for i := range tickets {
			if tickets[i].Slug == "ticket-without-current" {
				testTicket = &tickets[i]
				break
			}
		}
		require.NotNil(t, testTicket, "Could not find ticket-without-current")

		// Ensure current-ticket.md doesn't exist
		currentTicketPath := filepath.Join(repoPath, "current-ticket.md")
		os.Remove(currentTicketPath) // Ignore error if doesn't exist

		// Close the ticket - should not error even though current-ticket.md doesn't exist
		_, err = app.CloseTicketByID(ctx, testTicket.ID, "Test reason", false)
		require.NoError(t, err, "Should handle missing current-ticket.md gracefully")

		// Verify the ticket was closed
		closedTicket, err := app.Manager.Get(ctx, testTicket.ID)
		require.NoError(t, err)
		assert.Equal(t, ticket.StatusDone, closedTicket.Status())
	})
}