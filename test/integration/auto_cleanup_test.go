package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestAutoCleanupStaleBranchesIntegration(t *testing.T) {
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

	// Create the app
	app, err := cli.NewApp(context.Background())
	require.NoError(t, err)

	// Create multiple tickets
	ticketSlugs := []string{"cleanup-test-1", "cleanup-test-2", "cleanup-test-3"}
	ticketIDs := make([]string, 0, len(ticketSlugs))

	for _, slug := range ticketSlugs {
		// Create ticket
		err = app.NewTicket(context.Background(), slug, cli.FormatText)
		require.NoError(t, err)

		// Find the ticket ID
		tickets, err := app.Manager.List(context.Background(), ticket.StatusFilterTodo)
		require.NoError(t, err)

		var ticketID string
		for _, t := range tickets {
			if t.Slug == slug {
				ticketID = t.ID
				break
			}
		}
		require.NotEmpty(t, ticketID, "Could not find ticket: "+slug)
		ticketIDs = append(ticketIDs, ticketID)

		// Start the ticket (creates branch and worktree)
		err = app.StartTicket(context.Background(), ticketID)
		require.NoError(t, err)
	}

	// Verify all branches exist
	output, err := app.Git.Exec(context.Background(), "branch", "--format=%(refname:short)")
	require.NoError(t, err)
	branches := strings.Split(strings.TrimSpace(output), "\n")
	for _, id := range ticketIDs {
		assert.Contains(t, branches, id)
	}

	// Close first two tickets (move to done)
	// We'll move them manually since CloseTicket requires being in the worktree
	for i := 0; i < 2; i++ {
		// Get the ticket
		tkt, err := app.Manager.Get(context.Background(), ticketIDs[i])
		require.NoError(t, err)

		// Update ticket to set closed time
		now := time.Now()
		tkt.ClosedAt = ticket.RFC3339TimePtr{Time: &now}
		err = app.Manager.Update(context.Background(), tkt)
		require.NoError(t, err)

		// Move ticket file to done directory
		oldPath := tkt.Path
		newPath := filepath.Join(repoPath, "tickets", "done", filepath.Base(oldPath))
		err = os.Rename(oldPath, newPath)
		require.NoError(t, err)

		// Verify ticket is done
		tkt, err = app.Manager.Get(context.Background(), ticketIDs[i])
		require.NoError(t, err)
		assert.Equal(t, ticket.StatusDone, tkt.Status())
	}

	// Run auto cleanup
	result, err := app.AutoCleanup(context.Background(), false)
	require.NoError(t, err)
	assert.Equal(t, 2, result.StaleBranches, "Should clean 2 stale branches (done tickets)")
	assert.Equal(t, 2, result.OrphanedWorktrees, "Should clean 2 orphaned worktrees for done tickets")

	// Verify stale branches were removed
	output, err = app.Git.Exec(context.Background(), "branch", "--format=%(refname:short)")
	require.NoError(t, err)
	branches = strings.Split(strings.TrimSpace(output), "\n")

	// First two should be gone (done tickets)
	assert.NotContains(t, branches, ticketIDs[0])
	assert.NotContains(t, branches, ticketIDs[1])

	// Third should still exist (still in doing)
	assert.Contains(t, branches, ticketIDs[2])

	// Main branch should still exist
	assert.Contains(t, branches, "main")
}

func TestAutoCleanupDryRun(t *testing.T) {
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

	// Create the app
	app, err := cli.NewApp(context.Background())
	require.NoError(t, err)

	// Create a ticket
	err = app.NewTicket(context.Background(), "dry-run-test", cli.FormatText)
	require.NoError(t, err)

	// Find the ticket ID
	tickets, err := app.Manager.List(context.Background(), ticket.StatusFilterTodo)
	require.NoError(t, err)
	require.NotEmpty(t, tickets)

	var ticketID string
	for _, t := range tickets {
		if t.Slug == "dry-run-test" {
			ticketID = t.ID
			break
		}
	}
	require.NotEmpty(t, ticketID)

	// Start the ticket
	err = app.StartTicket(context.Background(), ticketID)
	require.NoError(t, err)

	// Manually close the ticket
	tkt, err := app.Manager.Get(context.Background(), ticketID)
	require.NoError(t, err)

	// Update ticket to set closed time
	now := time.Now()
	tkt.ClosedAt = ticket.RFC3339TimePtr{Time: &now}
	err = app.Manager.Update(context.Background(), tkt)
	require.NoError(t, err)

	// Move ticket file to done directory
	oldPath := tkt.Path
	newPath := filepath.Join(repoPath, "tickets", "done", filepath.Base(oldPath))
	err = os.Rename(oldPath, newPath)
	require.NoError(t, err)

	// Verify branch exists before cleanup
	output, err := app.Git.Exec(context.Background(), "branch", "--format=%(refname:short)")
	require.NoError(t, err)
	assert.Contains(t, output, ticketID)

	// Run auto cleanup with dry run
	result, err := app.AutoCleanup(context.Background(), true)
	require.NoError(t, err)
	assert.Equal(t, 1, result.StaleBranches, "Should detect 1 stale branch in dry run")
	assert.Equal(t, 1, result.OrphanedWorktrees, "Should detect 1 orphaned worktree in dry run")

	// Verify branch still exists (dry run should not delete)
	output, err = app.Git.Exec(context.Background(), "branch", "--format=%(refname:short)")
	require.NoError(t, err)
	assert.Contains(t, output, ticketID)
}
