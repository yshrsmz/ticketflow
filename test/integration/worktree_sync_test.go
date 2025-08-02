package integration

import (
	"context"
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

func TestStartTicket_WorktreeCreatedAfterCommit(t *testing.T) {
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

	// Initialize ticketflow with worktree enabled
	err = cli.InitCommand()
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
	app, err := cli.NewApp()
	require.NoError(t, err)

	// 1. Create a ticket
	err = app.NewTicket(context.Background(), "commit-first-test", cli.FormatText)
	require.NoError(t, err)

	// Commit the ticket
	_, err = gitCmd.Exec(context.Background(), "add", "tickets/")
	require.NoError(t, err)
	_, err = gitCmd.Exec(context.Background(), "commit", "-m", "Add test ticket")
	require.NoError(t, err)

	// Get ticket ID
	tickets, err := app.Manager.List(context.Background(), ticket.StatusFilterActive)
	require.NoError(t, err)
	require.Len(t, tickets, 1)
	ticketID := tickets[0].ID

	// 2. Start work on ticket
	err = app.StartTicket(context.Background(), ticketID)
	require.NoError(t, err)

	// 3. Verify parent branch state
	// Check that ticket is in doing directory
	parentDoingPath := filepath.Join(repoPath, "tickets", "doing", ticketID+".md")
	_, err = os.Stat(parentDoingPath)
	assert.NoError(t, err, "ticket should exist in parent branch doing directory")

	// Check that ticket is NOT in todo directory
	parentTodoPath := filepath.Join(repoPath, "tickets", "todo", ticketID+".md")
	_, err = os.Stat(parentTodoPath)
	assert.True(t, os.IsNotExist(err), "ticket should not exist in parent branch todo directory")

	// 4. Find the worktree
	worktrees, err := app.Git.ListWorktrees(context.Background())
	require.NoError(t, err)

	var ticketWorktree *git.WorktreeInfo
	for _, wt := range worktrees {
		if wt.Branch == ticketID {
			ticketWorktree = &wt
			break
		}
	}
	require.NotNil(t, ticketWorktree)

	// 5. Verify worktree state (should already have ticket in doing, NOT in todo)
	// Check that ticket is in doing directory
	worktreeDoingPath := filepath.Join(ticketWorktree.Path, "tickets", "doing", ticketID+".md")
	_, err = os.Stat(worktreeDoingPath)
	assert.NoError(t, err, "ticket should exist in worktree doing directory")

	// Check that ticket is NOT in todo directory
	worktreeTodoPath := filepath.Join(ticketWorktree.Path, "tickets", "todo", ticketID+".md")
	_, err = os.Stat(worktreeTodoPath)
	assert.True(t, os.IsNotExist(err), "ticket should not exist in worktree todo directory")

	// 6. Verify current-ticket.md symlink
	linkPath := filepath.Join(ticketWorktree.Path, "current-ticket.md")
	linkInfo, err := os.Lstat(linkPath)
	require.NoError(t, err)
	assert.True(t, linkInfo.Mode()&os.ModeSymlink != 0, "current-ticket.md should be a symlink")

	// Verify link target
	target, err := os.Readlink(linkPath)
	require.NoError(t, err)
	expectedTarget := filepath.Join("tickets", "doing", ticketID+".md")
	assert.Equal(t, expectedTarget, target)

	// 7. Verify worktree has clean status (no uncommitted changes)
	wtGit := git.New(ticketWorktree.Path)
	dirty, err := wtGit.HasUncommittedChanges(context.Background())
	require.NoError(t, err)
	assert.False(t, dirty, "worktree should have clean status with no uncommitted changes")
}
