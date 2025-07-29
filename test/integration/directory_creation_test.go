package integration

import (
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

func TestDirectoryAutoCreation(t *testing.T) {
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
	err = cli.InitCommand()
	require.NoError(t, err)

	// Load the app
	app, err := cli.NewApp()
	require.NoError(t, err)
	app.Config.Worktree.Enabled = false

	// Commit initial setup
	gitCmd := git.New(repoPath)
	_, err = gitCmd.Exec("add", ".")
	require.NoError(t, err)
	_, err = gitCmd.Exec("commit", "-m", "Initial setup")
	require.NoError(t, err)

	// Test 1: Remove doing directory and verify it's recreated on start
	doingPath := app.Config.GetDoingPath(repoPath)
	err = os.RemoveAll(doingPath)
	require.NoError(t, err)

	// Verify doing directory doesn't exist
	_, err = os.Stat(doingPath)
	assert.True(t, os.IsNotExist(err), "doing directory should not exist")

	// Create a ticket
	err = app.NewTicket("test-auto-dir", cli.FormatText)
	require.NoError(t, err)

	// Get the ticket
	tickets, err := app.Manager.List(ticket.StatusFilterActive)
	require.NoError(t, err)
	require.Len(t, tickets, 1)

	// Commit the ticket
	_, err = gitCmd.Exec("add", "tickets/")
	require.NoError(t, err)
	_, err = gitCmd.Exec("commit", "-m", "Add test ticket")
	require.NoError(t, err)

	// Start the ticket (this should create the doing directory)
	err = app.StartTicket(tickets[0].ID)
	require.NoError(t, err)

	// Verify doing directory was created
	info, err := os.Stat(doingPath)
	require.NoError(t, err)
	assert.True(t, info.IsDir(), "doing directory should be created")

	// Test 2: Remove done directory and verify it's recreated on close
	donePath := app.Config.GetDonePath(repoPath)
	err = os.RemoveAll(donePath)
	require.NoError(t, err)

	// Verify done directory doesn't exist
	_, err = os.Stat(donePath)
	assert.True(t, os.IsNotExist(err), "done directory should not exist")

	// Close the ticket (this should create the done directory)
	err = app.CloseTicket(false)
	require.NoError(t, err)

	// Verify done directory was created
	info, err = os.Stat(donePath)
	require.NoError(t, err)
	assert.True(t, info.IsDir(), "done directory should be created")
}

func TestDirectoryCreationWithWorktrees(t *testing.T) {
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

	// Create custom config with worktrees enabled
	cfg := config.Default()
	cfg.Worktree.Enabled = true
	cfg.Worktree.BaseDir = "./.worktrees"
	cfg.Worktree.InitCommands = []string{"git status"}

	configPath := filepath.Join(repoPath, ".ticketflow.yaml")
	err = cfg.Save(configPath)
	require.NoError(t, err)

	// Create initial directories
	ticketsDir := filepath.Join(repoPath, cfg.Tickets.Dir)
	todoDir := filepath.Join(ticketsDir, cfg.Tickets.TodoDir)
	err = os.MkdirAll(todoDir, 0755)
	require.NoError(t, err)

	// Load the app
	app, err := cli.NewApp()
	require.NoError(t, err)

	// Commit initial setup
	gitCmd := git.New(repoPath)
	_, err = gitCmd.Exec("add", ".")
	require.NoError(t, err)
	_, err = gitCmd.Exec("commit", "-m", "Initial setup with worktrees")
	require.NoError(t, err)

	// Remove doing directory
	doingPath := app.Config.GetDoingPath(repoPath)
	err = os.RemoveAll(doingPath)
	require.NoError(t, err)

	// Create a ticket
	err = app.NewTicket("test-worktree-dir", cli.FormatText)
	require.NoError(t, err)

	// Get the ticket
	tickets, err := app.Manager.List(ticket.StatusFilterActive)
	require.NoError(t, err)
	require.Len(t, tickets, 1)

	// Commit the ticket
	_, err = gitCmd.Exec("add", "tickets/")
	require.NoError(t, err)
	_, err = gitCmd.Exec("commit", "-m", "Add test ticket")
	require.NoError(t, err)

	// Start the ticket (this should create the doing directory even with worktrees)
	err = app.StartTicket(tickets[0].ID)
	require.NoError(t, err)

	// Verify doing directory was created
	info, err := os.Stat(doingPath)
	require.NoError(t, err)
	assert.True(t, info.IsDir(), "doing directory should be created with worktrees enabled")
}
