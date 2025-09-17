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

// TestStartTicketFromWithinWorktree tests that creating a worktree from within
// another worktree creates it as a sibling, not nested
func TestStartTicketFromWithinWorktree(t *testing.T) {
	// Cannot run in parallel due to os.Chdir in setupTestRepo

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

	// Create first ticket
	app, err := cli.NewAppWithWorkingDir(context.Background(), t, repoPath)
	require.NoError(t, err)
	firstTicket, err := app.NewTicket(context.Background(), "first-ticket", "")
	require.NoError(t, err)
	require.NotNil(t, firstTicket)

	// Start first ticket (creates first worktree)
	startResult, err := app.StartTicket(context.Background(), firstTicket.ID, false)
	require.NoError(t, err)
	require.NotEmpty(t, startResult.WorktreePath)

	// Verify first worktree was created in the right location
	expectedFirstWorktreePath := filepath.Join(repoPath, ".worktrees", firstTicket.ID)
	// Resolve symlinks for comparison (macOS temp dirs have /private prefix)
	expectedFirstResolved, _ := filepath.EvalSymlinks(expectedFirstWorktreePath)
	actualFirstResolved, _ := filepath.EvalSymlinks(startResult.WorktreePath)
	assert.Equal(t, expectedFirstResolved, actualFirstResolved)
	assert.DirExists(t, expectedFirstWorktreePath)

	// Now change to the first worktree directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalWd)
		require.NoError(t, err)
	}()
	require.NoError(t, os.Chdir(expectedFirstWorktreePath))

	// Create a second ticket from within the first worktree
	worktreeApp, err := cli.NewApp(context.Background())
	require.NoError(t, err)
	secondTicket, err := worktreeApp.NewTicket(context.Background(), "second-ticket", "")
	require.NoError(t, err)
	require.NotNil(t, secondTicket)

	// Start second ticket from within the first worktree
	// This is the critical test - it should create a sibling worktree, not nested
	secondStartResult, err := worktreeApp.StartTicket(context.Background(), secondTicket.ID, false)
	require.NoError(t, err)
	require.NotEmpty(t, secondStartResult.WorktreePath)

	// Verify second worktree was created as a sibling, not nested
	expectedSecondWorktreePath := filepath.Join(repoPath, ".worktrees", secondTicket.ID)
	// Resolve symlinks for comparison (macOS temp dirs have /private prefix)
	expectedSecondResolved, _ := filepath.EvalSymlinks(expectedSecondWorktreePath)
	actualSecondResolved, _ := filepath.EvalSymlinks(secondStartResult.WorktreePath)
	assert.Equal(t, expectedSecondResolved, actualSecondResolved)
	assert.DirExists(t, expectedSecondWorktreePath)

	// Verify it's NOT nested under the first worktree
	nestedPath := filepath.Join(expectedFirstWorktreePath, ".worktrees", secondTicket.ID)
	assert.NoDirExists(t, nestedPath, "Second worktree should not be nested under first worktree")

	// Verify both worktrees are siblings
	worktrees, err := gitCmd.ListWorktrees(context.Background())
	require.NoError(t, err)
	assert.Len(t, worktrees, 3) // Main + 2 worktrees

	// Find our worktrees
	var firstFound, secondFound bool
	for _, wt := range worktrees {
		if wt.Branch == firstTicket.ID {
			firstFound = true
			assert.Contains(t, wt.Path, ".worktrees/"+firstTicket.ID)
		}
		if wt.Branch == secondTicket.ID {
			secondFound = true
			assert.Contains(t, wt.Path, ".worktrees/"+secondTicket.ID)
		}
	}
	assert.True(t, firstFound, "First worktree should exist")
	assert.True(t, secondFound, "Second worktree should exist")

	// Verify ticket statuses
	// Load tickets from the worktree to verify they're in doing status
	secondTicketLoaded, err := worktreeApp.Manager.Get(context.Background(), secondTicket.ID)
	require.NoError(t, err)
	assert.Equal(t, ticket.StatusDoing, secondTicketLoaded.Status())
}
