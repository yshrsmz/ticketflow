package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/git"
)

func TestWorktreeWorkflow(t *testing.T) {
	// Setup test repository
	repoPath := setupTestRepo(t)
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(repoPath)

	// Initialize ticketflow with worktree enabled
	err := cli.InitCommand()
	require.NoError(t, err)

	// Enable worktrees in config
	cfg, err := config.Load(repoPath)
	require.NoError(t, err)
	cfg.Worktree.Enabled = true
	cfg.Worktree.BaseDir = "./.worktrees"
	cfg.Worktree.AutoOperations.CreateOnStart = true
	cfg.Worktree.AutoOperations.RemoveOnClose = true
	cfg.Worktree.InitCommands = []string{} // No init commands for test
	err = cfg.Save(filepath.Join(repoPath, ".ticketflow.yaml"))
	require.NoError(t, err)

	// Commit config
	gitCmd := git.New(repoPath)
	_, err = gitCmd.Exec("add", ".ticketflow.yaml", ".gitignore")
	require.NoError(t, err)
	_, err = gitCmd.Exec("commit", "-m", "Initialize ticketflow with worktrees")
	require.NoError(t, err)

	// Create app instance
	app, err := cli.NewApp()
	require.NoError(t, err)

	// 1. Create a ticket
	err = app.NewTicket("worktree-test", cli.FormatText)
	require.NoError(t, err)

	// Commit the ticket
	_, err = gitCmd.Exec("add", "tickets/")
	require.NoError(t, err)
	_, err = gitCmd.Exec("commit", "-m", "Add test ticket")
	require.NoError(t, err)

	// 2. Start work on ticket (should create worktree)
	tickets, err := app.Manager.List("")
	require.NoError(t, err)
	require.Len(t, tickets, 1)

	ticketID := tickets[0].ID
	err = app.StartTicket(ticketID, true) // no-push
	require.NoError(t, err)

	// Verify worktree was created
	worktrees, err := app.Git.ListWorktrees()
	require.NoError(t, err)
	assert.Len(t, worktrees, 2) // main + ticket worktree

	// Find the ticket worktree
	var ticketWorktree *git.WorktreeInfo
	for _, wt := range worktrees {
		if wt.Branch == ticketID {
			ticketWorktree = &wt
			break
		}
	}
	require.NotNil(t, ticketWorktree)

	expectedPath := filepath.Join(repoPath, ".worktrees", ticketID)
	assert.Equal(t, expectedPath, ticketWorktree.Path)

	// Verify worktree directory exists
	_, err = os.Stat(ticketWorktree.Path)
	require.NoError(t, err)

	// 3. Make changes in worktree
	wtGit := git.New(ticketWorktree.Path)
	testFile := filepath.Join(ticketWorktree.Path, "worktree-test.txt")
	err = os.WriteFile(testFile, []byte("test content from worktree\n"), 0644)
	require.NoError(t, err)

	_, err = wtGit.Exec("add", "worktree-test.txt")
	require.NoError(t, err)
	_, err = wtGit.Exec("commit", "-m", "Add test file in worktree")
	require.NoError(t, err)

	// 4. Close ticket (should remove worktree)
	err = app.CloseTicket(true, false) // no-push
	require.NoError(t, err)

	// Verify worktree was removed
	worktrees, err = app.Git.ListWorktrees()
	require.NoError(t, err)
	assert.Len(t, worktrees, 1) // only main

	// Verify worktree directory was removed
	_, err = os.Stat(ticketWorktree.Path)
	assert.True(t, os.IsNotExist(err))

	// Verify we're back on main branch
	currentBranch, err := app.Git.CurrentBranch()
	require.NoError(t, err)
	assert.Equal(t, "main", currentBranch)
}

func TestWorktreeCleanCommand(t *testing.T) {
	// Setup test repository
	repoPath := setupTestRepo(t)
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(repoPath)

	// Initialize ticketflow with worktrees
	err := cli.InitCommand()
	require.NoError(t, err)

	cfg, err := config.Load(repoPath)
	require.NoError(t, err)
	cfg.Worktree.Enabled = true
	cfg.Worktree.BaseDir = "./.worktrees"
	err = cfg.Save(filepath.Join(repoPath, ".ticketflow.yaml"))
	require.NoError(t, err)

	gitCmd := git.New(repoPath)
	_, err = gitCmd.Exec("add", ".ticketflow.yaml", ".gitignore")
	require.NoError(t, err)
	_, err = gitCmd.Exec("commit", "-m", "Initialize ticketflow")
	require.NoError(t, err)

	app, err := cli.NewApp()
	require.NoError(t, err)

	// Create multiple tickets
	for i := 1; i <= 3; i++ {
		slug := fmt.Sprintf("ticket-%d", i)
		err = app.NewTicket(slug, cli.FormatText)
		require.NoError(t, err)
	}

	_, err = gitCmd.Exec("add", "tickets/")
	require.NoError(t, err)
	_, err = gitCmd.Exec("commit", "-m", "Add test tickets")
	require.NoError(t, err)

	tickets, err := app.Manager.List("")
	require.NoError(t, err)
	require.Len(t, tickets, 3)

	// Start work on first two tickets
	err = app.StartTicket(tickets[0].ID, true)
	require.NoError(t, err)
	err = app.StartTicket(tickets[1].ID, true)
	require.NoError(t, err)

	// Manually close the first ticket without removing worktree
	// (simulating orphaned worktree)
	ticket1, err := app.Manager.Get(tickets[0].ID)
	require.NoError(t, err)
	
	// Manually move the ticket to done directory
	oldPath := ticket1.Path
	donePath := filepath.Join(repoPath, "tickets", "done")
	newPath := filepath.Join(donePath, filepath.Base(ticket1.Path))
	os.MkdirAll(donePath, 0755)
	os.Rename(oldPath, newPath)
	
	// Update ticket status and path
	ticket1.Path = newPath
	err = ticket1.Close()
	require.NoError(t, err)
	err = app.Manager.Update(ticket1)
	require.NoError(t, err)

	// Verify we have 3 worktrees (main + 2 ticket worktrees)
	worktrees, err := app.Git.ListWorktrees()
	require.NoError(t, err)
	assert.Len(t, worktrees, 3)

	// Run clean command
	err = app.CleanWorktrees()
	require.NoError(t, err)

	// Should have removed the orphaned worktree
	worktrees, err = app.Git.ListWorktrees()
	require.NoError(t, err)
	assert.Len(t, worktrees, 2) // main + ticket2 (still active)

	// Verify the remaining worktree is for the active ticket
	for _, wt := range worktrees {
		if wt.Branch != "" && wt.Branch != "main" {
			assert.Equal(t, tickets[1].ID, wt.Branch)
		}
	}
}

func TestWorktreeListCommand(t *testing.T) {
	// Setup test repository
	repoPath := setupTestRepo(t)
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(repoPath)

	// Initialize and setup
	err := cli.InitCommand()
	require.NoError(t, err)

	cfg, err := config.Load(repoPath)
	require.NoError(t, err)
	cfg.Worktree.Enabled = true
	cfg.Worktree.BaseDir = "./.worktrees"
	err = cfg.Save(filepath.Join(repoPath, ".ticketflow.yaml"))
	require.NoError(t, err)

	gitCmd := git.New(repoPath)
	_, err = gitCmd.Exec("add", ".ticketflow.yaml", ".gitignore")
	require.NoError(t, err)
	_, err = gitCmd.Exec("commit", "-m", "Initialize")
	require.NoError(t, err)

	app, err := cli.NewApp()
	require.NoError(t, err)

	// Create and start a ticket
	err = app.NewTicket("list-test", cli.FormatText)
	require.NoError(t, err)

	_, err = gitCmd.Exec("add", "tickets/")
	require.NoError(t, err)
	_, err = gitCmd.Exec("commit", "-m", "Add ticket")
	require.NoError(t, err)

	tickets, err := app.Manager.List("")
	require.NoError(t, err)
	err = app.StartTicket(tickets[0].ID, true)
	require.NoError(t, err)

	// Test listing worktrees
	worktrees, err := app.Git.ListWorktrees()
	require.NoError(t, err)
	assert.Len(t, worktrees, 2)

	// Verify both worktrees are present
	branches := make(map[string]bool)
	for _, wt := range worktrees {
		if wt.Branch != "" {
			branches[wt.Branch] = true
		}
	}
	assert.True(t, branches["main"])
	assert.True(t, branches[tickets[0].ID])
}