package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func setupTestRepo(t *testing.T) string {
	// Create temp directory
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := git.New(tmpDir)
	_, err := cmd.Exec(context.Background(), "init")
	require.NoError(t, err)

	// Set git config
	_, err = cmd.Exec(context.Background(), "config", "user.name", "Test User")
	require.NoError(t, err)
	_, err = cmd.Exec(context.Background(), "config", "user.email", "test@example.com")
	require.NoError(t, err)
	// Disable GPG signing for test commits
	_, err = cmd.Exec(context.Background(), "config", "commit.gpgSign", "false")
	require.NoError(t, err)

	// Create initial commit
	readmePath := filepath.Join(tmpDir, "README.md")
	err = os.WriteFile(readmePath, []byte("# Test Project\n"), 0644)
	require.NoError(t, err)

	_, err = cmd.Exec(context.Background(), "add", "README.md")
	require.NoError(t, err)
	_, err = cmd.Exec(context.Background(), "commit", "-m", "Initial commit")
	require.NoError(t, err)

	// Ensure we're on the main branch (important for CI where tests may run on PR branches)
	_, err = cmd.Exec(context.Background(), "checkout", "main")
	if err != nil {
		// Branch does not exist, create it
		_, err = cmd.Exec(context.Background(), "checkout", "-b", "main")
		require.NoError(t, err)
	}

	return tmpDir
}

func TestCompleteWorkflow(t *testing.T) {
	t.Parallel()

	// Setup test repository
	repoPath := setupTestRepo(t)

	// 1. Initialize ticketflow
	err := cli.InitCommandWithWorkingDir(context.Background(), repoPath)
	require.NoError(t, err)

	// Verify config exists
	configPath := filepath.Join(repoPath, ".ticketflow.yaml")
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	// Verify tickets directory exists
	ticketsPath := filepath.Join(repoPath, "tickets")
	_, err = os.Stat(ticketsPath)
	require.NoError(t, err)

	// Disable worktrees for this test
	app, err := cli.NewAppWithWorkingDir(context.Background(), t, repoPath)
	require.NoError(t, err)
	app.Config.Worktree.Enabled = false

	// Commit the config file to avoid dirty workspace
	gitCmd := git.New(repoPath)
	_, err = gitCmd.Exec(context.Background(), "add", ".ticketflow.yaml", ".gitignore")
	require.NoError(t, err)
	_, err = gitCmd.Exec(context.Background(), "commit", "-m", "Initialize ticketflow")
	require.NoError(t, err)

	// 2. Create a new ticket

	_, err = app.NewTicket(context.Background(), "test-feature", "")
	require.NoError(t, err)

	// 3. List tickets
	tickets, err := app.Manager.List(context.Background(), ticket.StatusFilterActive)
	require.NoError(t, err)
	assert.Len(t, tickets, 1)
	assert.Equal(t, "test-feature", tickets[0].Slug)
	assert.Equal(t, ticket.StatusTodo, tickets[0].Status())

	// Commit the new ticket to avoid dirty workspace
	_, err = gitCmd.Exec(context.Background(), "add", "tickets/")
	require.NoError(t, err)
	_, err = gitCmd.Exec(context.Background(), "commit", "-m", "Add test ticket")
	require.NoError(t, err)

	// 4. Start work on ticket
	ticketID := tickets[0].ID
	_, err = app.StartTicket(context.Background(), ticketID, false)
	require.NoError(t, err)

	// Verify ticket status changed
	updatedTicket, err := app.Manager.Get(context.Background(), ticketID)
	require.NoError(t, err)
	assert.Equal(t, ticket.StatusDoing, updatedTicket.Status())
	assert.NotNil(t, updatedTicket.StartedAt)

	// Verify branch created
	currentBranch, err := app.Git.CurrentBranch(context.Background())
	require.NoError(t, err)
	assert.Equal(t, ticketID, currentBranch)

	// Verify current-ticket symlink
	linkPath := filepath.Join(repoPath, "current-ticket.md")
	_, err = os.Lstat(linkPath)
	require.NoError(t, err)

	// The ticket move is already committed by StartTicket, no need to commit again

	// 5. Make some changes
	testFile := filepath.Join(repoPath, "test.txt")
	err = os.WriteFile(testFile, []byte("test content\n"), 0644)
	require.NoError(t, err)

	err = app.Git.Add(context.Background(), "test.txt")
	require.NoError(t, err)
	err = app.Git.Commit(context.Background(), "Add test file")
	require.NoError(t, err)

	// 6. Close ticket
	_, err = app.CloseTicket(context.Background(), false)
	require.NoError(t, err)

	// Branch should still be on ticket branch (no automatic switch)
	currentBranch, err = app.Git.CurrentBranch(context.Background())
	require.NoError(t, err)
	assert.Equal(t, ticketID, currentBranch)

	// Verify ticket status changed
	closedTicket, err := app.Manager.Get(context.Background(), ticketID)
	require.NoError(t, err)
	assert.Equal(t, ticket.StatusDone, closedTicket.Status())
	assert.NotNil(t, closedTicket.ClosedAt)

	// Verify current-ticket symlink removed
	_, err = os.Lstat(linkPath)
	assert.True(t, os.IsNotExist(err))
}

func TestRestoreWorkflow(t *testing.T) {
	t.Parallel()

	// Setup test repository
	repoPath := setupTestRepo(t)

	// Initialize and create ticket
	err := cli.InitCommandWithWorkingDir(context.Background(), repoPath)
	require.NoError(t, err)

	// Disable worktrees for this test
	app, err := cli.NewAppWithWorkingDir(context.Background(), t, repoPath)
	require.NoError(t, err)
	app.Config.Worktree.Enabled = false

	// Commit the config file to avoid dirty workspace
	gitCmd := git.New(repoPath)
	_, err = gitCmd.Exec(context.Background(), "add", ".ticketflow.yaml", ".gitignore")
	require.NoError(t, err)
	_, err = gitCmd.Exec(context.Background(), "commit", "-m", "Initialize ticketflow")
	require.NoError(t, err)

	_, err = app.NewTicket(context.Background(), "restore-test", "")
	require.NoError(t, err)

	tickets, err := app.Manager.List(context.Background(), ticket.StatusFilterActive)
	require.NoError(t, err)
	ticketID := tickets[0].ID

	// Commit the new ticket to avoid dirty workspace
	_, err = gitCmd.Exec(context.Background(), "add", "tickets/")
	require.NoError(t, err)
	_, err = gitCmd.Exec(context.Background(), "commit", "-m", "Add test ticket")
	require.NoError(t, err)

	// Start work on ticket
	_, err = app.StartTicket(context.Background(), ticketID, false)
	require.NoError(t, err)

	// Remove current-ticket link
	linkPath := filepath.Join(repoPath, "current-ticket.md")
	err = os.Remove(linkPath)
	require.NoError(t, err)

	// Restore link
	_, err = app.RestoreCurrentTicket(context.Background())
	require.NoError(t, err)

	// Verify link restored
	_, err = os.Lstat(linkPath)
	require.NoError(t, err)

	// Verify correct ticket linked
	current, err := app.Manager.GetCurrentTicket(context.Background())
	require.NoError(t, err)
	require.NotNil(t, current)
	assert.Equal(t, ticketID, current.ID)
}
