package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestCloseWithReason(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Setup test repository
	repoPath := setupTestRepo(t)

	// Initialize ticketflow
	err := cli.InitCommandWithWorkingDir(ctx, repoPath)
	require.NoError(t, err)

	// Create app instance
	app, err := cli.NewAppWithWorkingDir(ctx, t, repoPath)
	require.NoError(t, err)

	// Disable worktrees for simpler testing
	app.Config.Worktree.Enabled = false

	t.Run("close ticket with reason", func(t *testing.T) {
		// Create a test ticket
		err := app.NewTicket(ctx, "test-close-reason", "", cli.FormatText)
		require.NoError(t, err)

		// Get the created ticket
		tickets, err := app.Manager.List(ctx, ticket.StatusFilterTodo)
		require.NoError(t, err)
		require.NotEmpty(t, tickets)

		var testTicket *ticket.Ticket
		for _, tkt := range tickets {
			if tkt.Slug == "test-close-reason" {
				testTicket = &tkt
				break
			}
		}
		require.NotNil(t, testTicket, "Could not find created ticket")

		// Close the ticket with reason
		err = app.CloseTicketByID(ctx, testTicket.ID, "Duplicate of another ticket", false)
		require.NoError(t, err)

		// Verify ticket was moved to done
		updatedTicket, err := app.Manager.Get(ctx, testTicket.ID)
		require.NoError(t, err)
		assert.Equal(t, ticket.StatusDone, updatedTicket.Status())

		// Check that closure reason is set
		assert.Equal(t, "Duplicate of another ticket", updatedTicket.ClosureReason)

		// Check that closure note is in content
		assert.Contains(t, updatedTicket.Content, "## Closure Note")
		assert.Contains(t, updatedTicket.Content, "**Reason**: Duplicate of another ticket")

		// Verify git commit was created with reason
		output, err := app.Git.Exec(ctx, "log", "--oneline", "-1")
		require.NoError(t, err)
		assert.Contains(t, output, "Close ticket: "+testTicket.ID+" (Duplicate of another ticket)")
	})

	t.Run("close ticket without reason when branch merged", func(t *testing.T) {
		// Create a ticket
		err := app.NewTicket(ctx, "test-merged-branch", "", cli.FormatText)
		require.NoError(t, err)

		// Get the created ticket
		tickets, err := app.Manager.List(ctx, ticket.StatusFilterTodo)
		require.NoError(t, err)

		var testTicket *ticket.Ticket
		for _, tkt := range tickets {
			if tkt.Slug == "test-merged-branch" {
				testTicket = &tkt
				break
			}
		}
		require.NotNil(t, testTicket, "Could not find created ticket")

		// Commit the new ticket file
		err = app.Git.Add(ctx, ".")
		require.NoError(t, err)
		err = app.Git.Commit(ctx, "Add ticket: test-merged-branch")
		require.NoError(t, err)

		// Start the ticket (creates branch)
		err = app.StartTicket(ctx, testTicket.ID, false)
		require.NoError(t, err)

		// Make a change on the branch
		testFile := filepath.Join(repoPath, "test-file.txt")
		err = os.WriteFile(testFile, []byte("test content"), 0644)
		require.NoError(t, err)
		err = app.Git.Add(ctx, "test-file.txt")
		require.NoError(t, err)
		err = app.Git.Commit(ctx, "Add test file")
		require.NoError(t, err)

		// Switch back to main and merge
		err = app.Git.Checkout(ctx, "main")
		require.NoError(t, err)

		// Remove the current ticket symlink since we're not on the ticket branch anymore
		err = app.Manager.SetCurrentTicket(ctx, nil)
		require.NoError(t, err)
		// Use regular merge instead of squash merge so git considers the branch merged
		_, err = app.Git.Exec(ctx, "merge", testTicket.ID)
		require.NoError(t, err)

		// Verify branch is actually merged
		merged, err := app.Git.IsBranchMerged(ctx, testTicket.ID, "main")
		require.NoError(t, err)
		assert.True(t, merged, "Branch should be merged to main")

		// Close the ticket without reason (should work since branch is merged)
		err = app.CloseTicketByID(ctx, testTicket.ID, "", false)
		require.NoError(t, err)

		// Verify ticket was moved to done
		updatedTicket, err := app.Manager.Get(ctx, testTicket.ID)
		require.NoError(t, err)
		assert.Equal(t, ticket.StatusDone, updatedTicket.Status())

		// Check that no closure reason is added
		assert.Empty(t, updatedTicket.ClosureReason)
		assert.NotContains(t, updatedTicket.Content, "## Closure Note")

		// Verify git commit message has no reason
		output, err := app.Git.Exec(ctx, "log", "--oneline", "-1")
		require.NoError(t, err)
		assert.Contains(t, output, "Close ticket: "+testTicket.ID)
		assert.NotContains(t, output, "(") // No reason in parentheses
	})

	t.Run("error when closing unmerged ticket without reason", func(t *testing.T) {
		// Create a ticket
		err := app.NewTicket(ctx, "test-unmerged", "", cli.FormatText)
		require.NoError(t, err)

		// Get the created ticket
		tickets, err := app.Manager.List(ctx, ticket.StatusFilterTodo)
		require.NoError(t, err)

		var testTicket *ticket.Ticket
		for _, tkt := range tickets {
			if tkt.Slug == "test-unmerged" {
				testTicket = &tkt
				break
			}
		}
		require.NotNil(t, testTicket, "Could not find created ticket")

		// Commit the new ticket file
		err = app.Git.Add(ctx, ".")
		require.NoError(t, err)
		err = app.Git.Commit(ctx, "Add ticket: test-unmerged")
		require.NoError(t, err)

		// Start the ticket (creates branch)
		err = app.StartTicket(ctx, testTicket.ID, false)
		require.NoError(t, err)

		// Make a change on the branch but don't merge
		testFile := filepath.Join(repoPath, "unmerged-file.txt")
		err = os.WriteFile(testFile, []byte("unmerged content"), 0644)
		require.NoError(t, err)
		err = app.Git.Add(ctx, "unmerged-file.txt")
		require.NoError(t, err)
		err = app.Git.Commit(ctx, "Add unmerged file")
		require.NoError(t, err)

		// Switch back to main without merging
		err = app.Git.Checkout(ctx, "main")
		require.NoError(t, err)

		// Try to close the ticket without reason (should fail)
		err = app.CloseTicketByID(ctx, testTicket.ID, "", false)
		assert.Error(t, err)
		assert.Contains(t, strings.ToLower(err.Error()), "reason")

		// Verify ticket was NOT moved to done
		updatedTicket, err := app.Manager.Get(ctx, testTicket.ID)
		require.NoError(t, err)
		assert.NotEqual(t, ticket.StatusDone, updatedTicket.Status())

		// Now close with reason (should work)
		err = app.CloseTicketByID(ctx, testTicket.ID, "Abandoned due to priority change", false)
		require.NoError(t, err)

		// Verify ticket was moved to done
		updatedTicket, err = app.Manager.Get(ctx, testTicket.ID)
		require.NoError(t, err)
		assert.Equal(t, ticket.StatusDone, updatedTicket.Status())
		assert.Equal(t, "Abandoned due to priority change", updatedTicket.ClosureReason)
	})
}
