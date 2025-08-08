package integration

import (
	"context"
	"fmt"
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

func TestWorktreeWorkflow(t *testing.T) {
	t.Parallel()

	// Setup test repository
	repoPath := setupTestRepo(t)

	// Initialize ticketflow with worktree enabled
	err := cli.InitCommandWithWorkingDir(context.Background(), repoPath)
	require.NoError(t, err)

	// Enable worktrees in config
	cfg, err := config.Load(repoPath)
	require.NoError(t, err)
	cfg.Worktree.Enabled = true
	cfg.Worktree.BaseDir = "./.worktrees"
	cfg.Worktree.InitCommands = []string{} // No init commands for test
	err = cfg.Save(filepath.Join(repoPath, ".ticketflow.yaml"))
	require.NoError(t, err)

	// Commit config
	gitCmd := git.New(repoPath)
	_, err = gitCmd.Exec(context.Background(), "add", ".ticketflow.yaml", ".gitignore")
	require.NoError(t, err)
	_, err = gitCmd.Exec(context.Background(), "commit", "-m", "Initialize ticketflow with worktrees")
	require.NoError(t, err)

	// Create app instance
	app, err := cli.NewAppWithWorkingDir(context.Background(), t, repoPath)
	require.NoError(t, err)

	// 1. Create a ticket
	err = app.NewTicket(context.Background(), "worktree-test", "", cli.FormatText)
	require.NoError(t, err)

	// Commit the ticket
	_, err = gitCmd.Exec(context.Background(), "add", "tickets/")
	require.NoError(t, err)
	_, err = gitCmd.Exec(context.Background(), "commit", "-m", "Add test ticket")
	require.NoError(t, err)

	// 2. Start work on ticket (should create worktree)
	tickets, err := app.Manager.List(context.Background(), ticket.StatusFilterActive)
	require.NoError(t, err)
	require.Len(t, tickets, 1)

	ticketID := tickets[0].ID
	err = app.StartTicket(context.Background(), ticketID, false)
	require.NoError(t, err)

	// Verify worktree was created
	worktrees, err := app.Git.ListWorktrees(context.Background())
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
	// Resolve symlinks to handle macOS /var -> /private/var symlink
	expectedPath, err = filepath.EvalSymlinks(expectedPath)
	require.NoError(t, err)
	actualPath, err := filepath.EvalSymlinks(ticketWorktree.Path)
	require.NoError(t, err)
	assert.Equal(t, expectedPath, actualPath)

	// Verify worktree directory exists
	_, err = os.Stat(ticketWorktree.Path)
	require.NoError(t, err)

	// 3. Make changes in worktree
	wtGit := git.New(ticketWorktree.Path)
	testFile := filepath.Join(ticketWorktree.Path, "worktree-test.txt")
	err = os.WriteFile(testFile, []byte("test content from worktree\n"), 0644)
	require.NoError(t, err)

	_, err = wtGit.Exec(context.Background(), "add", "worktree-test.txt")
	require.NoError(t, err)
	_, err = wtGit.Exec(context.Background(), "commit", "-m", "Add test file in worktree")
	require.NoError(t, err)

	// 4. Close ticket from within the worktree (as required when worktrees are enabled)
	// Change to worktree directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(ticketWorktree.Path)
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	// Create app instance - it will use the current directory (worktree)
	wtApp, err := cli.NewApp(context.Background())
	require.NoError(t, err)

	// Now close the ticket from within the worktree
	err = wtApp.CloseTicket(context.Background(), false)
	require.NoError(t, err)

	// Change back to original directory
	err = os.Chdir(originalWd)
	require.NoError(t, err)

	// Verify worktree still exists
	worktrees, err = app.Git.ListWorktrees(context.Background())
	require.NoError(t, err)
	assert.Len(t, worktrees, 2) // main + ticket worktree

	// Verify worktree directory still exists
	_, err = os.Stat(ticketWorktree.Path)
	require.NoError(t, err)

	// We should still be on main branch (test was run from main repo)
	currentBranch, err := app.Git.CurrentBranch(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "main", currentBranch)
}

func TestWorktreeCleanCommand(t *testing.T) {
	t.Parallel()

	// Setup test repository
	repoPath := setupTestRepo(t)

	// Initialize ticketflow with worktrees
	err := cli.InitCommandWithWorkingDir(context.Background(), repoPath)
	require.NoError(t, err)

	cfg, err := config.Load(repoPath)
	require.NoError(t, err)
	cfg.Worktree.Enabled = true
	cfg.Worktree.BaseDir = "./.worktrees"
	err = cfg.Save(filepath.Join(repoPath, ".ticketflow.yaml"))
	require.NoError(t, err)

	gitCmd := git.New(repoPath)
	_, err = gitCmd.Exec(context.Background(), "add", ".ticketflow.yaml", ".gitignore")
	require.NoError(t, err)
	_, err = gitCmd.Exec(context.Background(), "commit", "-m", "Initialize ticketflow")
	require.NoError(t, err)

	app, err := cli.NewAppWithWorkingDir(context.Background(), t, repoPath)
	require.NoError(t, err)

	// Create multiple tickets
	for i := 1; i <= 3; i++ {
		slug := fmt.Sprintf("ticket-%d", i)
		err = app.NewTicket(context.Background(), slug, "", cli.FormatText)
		require.NoError(t, err)
	}

	_, err = gitCmd.Exec(context.Background(), "add", "tickets/")
	require.NoError(t, err)
	_, err = gitCmd.Exec(context.Background(), "commit", "-m", "Add test tickets")
	require.NoError(t, err)

	tickets, err := app.Manager.List(context.Background(), ticket.StatusFilterActive)
	require.NoError(t, err)
	require.Len(t, tickets, 3)

	// Start work on first two tickets
	err = app.StartTicket(context.Background(), tickets[0].ID, false)
	require.NoError(t, err)
	err = app.StartTicket(context.Background(), tickets[1].ID, false)
	require.NoError(t, err)

	// Manually close the first ticket without removing worktree
	// (simulating orphaned worktree)
	ticket1, err := app.Manager.Get(context.Background(), tickets[0].ID)
	require.NoError(t, err)

	// Manually move the ticket to done directory
	oldPath := ticket1.Path
	donePath := filepath.Join(repoPath, "tickets", "done")
	newPath := filepath.Join(donePath, filepath.Base(ticket1.Path))
	err = os.MkdirAll(donePath, 0755)
	require.NoError(t, err)
	err = os.Rename(oldPath, newPath)
	require.NoError(t, err)

	// Update ticket status and path
	ticket1.Path = newPath
	err = ticket1.Close()
	require.NoError(t, err)
	err = app.Manager.Update(context.Background(), ticket1)
	require.NoError(t, err)

	// Verify we have 3 worktrees (main + 2 ticket worktrees)
	worktrees, err := app.Git.ListWorktrees(context.Background())
	require.NoError(t, err)
	assert.Len(t, worktrees, 3)

	// Run clean command
	err = app.CleanWorktrees(context.Background())
	require.NoError(t, err)

	// Should have removed the orphaned worktree
	worktrees, err = app.Git.ListWorktrees(context.Background())
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
	t.Parallel()

	// Setup test repository
	repoPath := setupTestRepo(t)

	// Initialize and setup
	err := cli.InitCommandWithWorkingDir(context.Background(), repoPath)
	require.NoError(t, err)

	cfg, err := config.Load(repoPath)
	require.NoError(t, err)
	cfg.Worktree.Enabled = true
	cfg.Worktree.BaseDir = "./.worktrees"
	err = cfg.Save(filepath.Join(repoPath, ".ticketflow.yaml"))
	require.NoError(t, err)

	gitCmd := git.New(repoPath)
	_, err = gitCmd.Exec(context.Background(), "add", ".ticketflow.yaml", ".gitignore")
	require.NoError(t, err)
	_, err = gitCmd.Exec(context.Background(), "commit", "-m", "Initialize")
	require.NoError(t, err)

	app, err := cli.NewAppWithWorkingDir(context.Background(), t, repoPath)
	require.NoError(t, err)

	// Create and start a ticket
	err = app.NewTicket(context.Background(), "list-test", "", cli.FormatText)
	require.NoError(t, err)

	_, err = gitCmd.Exec(context.Background(), "add", "tickets/")
	require.NoError(t, err)
	_, err = gitCmd.Exec(context.Background(), "commit", "-m", "Add ticket")
	require.NoError(t, err)

	tickets, err := app.Manager.List(context.Background(), ticket.StatusFilterActive)
	require.NoError(t, err)
	err = app.StartTicket(context.Background(), tickets[0].ID, false)
	require.NoError(t, err)

	// Test listing worktrees
	worktrees, err := app.Git.ListWorktrees(context.Background())
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
