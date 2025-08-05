package integration

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/config"
	ticketerrors "github.com/yshrsmz/ticketflow/internal/errors"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestStartTicketWithDivergedBranch(t *testing.T) {
	// Cannot run in parallel due to os.Chdir
	
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
	err = cfg.Save(filepath.Join(repoPath, ".ticketflow.yaml"))
	require.NoError(t, err)

	// Commit config
	gitCmd := git.New(repoPath)
	ctx := context.Background()
	_, err = gitCmd.Exec(ctx, "add", ".ticketflow.yaml", ".gitignore")
	require.NoError(t, err)
	_, err = gitCmd.Exec(ctx, "commit", "-m", "Initialize ticketflow")
	require.NoError(t, err)

	// Create app instance
	app, err := cli.NewAppWithWorkingDir(ctx, t, repoPath)
	require.NoError(t, err)

	// Create a test ticket
	err = app.NewTicket(ctx, "diverged-branch-test", "", cli.FormatText)
	require.NoError(t, err)

	// Commit the ticket
	_, err = gitCmd.Exec(ctx, "add", "tickets/")
	require.NoError(t, err)
	_, err = gitCmd.Exec(ctx, "commit", "-m", "Add test ticket")
	require.NoError(t, err)

	// Get the ticket ID
	tickets, err := app.Manager.List(ctx, ticket.StatusFilterActive)
	require.NoError(t, err)
	require.Len(t, tickets, 1)
	ticketID := tickets[0].ID

	// Create the branch manually with a commit
	err = gitCmd.CreateBranch(ctx, ticketID)
	require.NoError(t, err)
	err = gitCmd.Checkout(ctx, ticketID)
	require.NoError(t, err)

	// Make a commit on the branch
	testFile := filepath.Join(repoPath, "branch-file.txt")
	err = os.WriteFile(testFile, []byte("branch content"), 0644)
	require.NoError(t, err)
	_, err = gitCmd.Exec(ctx, "add", "branch-file.txt")
	require.NoError(t, err)
	_, err = gitCmd.Exec(ctx, "commit", "-m", "Branch commit")
	require.NoError(t, err)

	// Switch back to main and make another commit
	err = gitCmd.Checkout(ctx, "main")
	require.NoError(t, err)
	mainFile := filepath.Join(repoPath, "main-file.txt")
	err = os.WriteFile(mainFile, []byte("main content"), 0644)
	require.NoError(t, err)
	_, err = gitCmd.Exec(ctx, "add", "main-file.txt")
	require.NoError(t, err)
	_, err = gitCmd.Exec(ctx, "commit", "-m", "Main commit")
	require.NoError(t, err)

	// Verify branch has diverged
	diverged, err := gitCmd.BranchDivergedFrom(ctx, ticketID, "main")
	require.NoError(t, err)
	assert.True(t, diverged, "Branch should have diverged")

	ahead, behind, err := gitCmd.GetBranchDivergenceInfo(ctx, ticketID, "main")
	require.NoError(t, err)
	assert.Equal(t, 1, ahead, "Branch should be 1 commit ahead")
	assert.Equal(t, 1, behind, "Branch should be 1 commit behind")

	// Now try to start the ticket
	// In CI/non-interactive mode, it should automatically use the default option (recreate branch)
	err = app.StartTicket(ctx, ticketID, false)
	require.NoError(t, err, "StartTicket should succeed in non-interactive mode by using default option")

	// Verify worktree was created successfully
	hasWorktree, err := gitCmd.HasWorktree(ctx, ticketID)
	require.NoError(t, err)
	assert.True(t, hasWorktree, "Worktree should exist after successful start")
	
	// Verify branch was recreated at main's HEAD
	// The branch should no longer be diverged
	diverged, err = gitCmd.BranchDivergedFrom(ctx, ticketID, "main")
	require.NoError(t, err)
	assert.False(t, diverged, "Branch should not be diverged after recreation")
}

func TestBranchDivergenceWithSameCommit(t *testing.T) {
	// Cannot run in parallel due to os.Chdir
	
	// This test verifies that when a branch exists at the same commit as main,
	// but StartTicket creates a new commit (status change), it will detect divergence
	// and prompt the user. This is expected behavior.
	
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
	err = cfg.Save(filepath.Join(repoPath, ".ticketflow.yaml"))
	require.NoError(t, err)

	// Commit config
	gitCmd := git.New(repoPath)
	ctx := context.Background()
	_, err = gitCmd.Exec(ctx, "add", ".ticketflow.yaml", ".gitignore")
	require.NoError(t, err)
	_, err = gitCmd.Exec(ctx, "commit", "-m", "Initialize ticketflow")
	require.NoError(t, err)

	// Create app instance
	app, err := cli.NewAppWithWorkingDir(ctx, t, repoPath)
	require.NoError(t, err)

	// Create a test ticket
	err = app.NewTicket(ctx, "same-commit-test", "", cli.FormatText)
	require.NoError(t, err)

	// Get the ticket ID first (before committing)
	tickets, err := app.Manager.List(ctx, ticket.StatusFilterActive)
	require.NoError(t, err)
	require.Len(t, tickets, 1)
	ticketID := tickets[0].ID

	// Commit the ticket
	_, err = gitCmd.Exec(ctx, "add", "tickets/")
	require.NoError(t, err)
	_, err = gitCmd.Exec(ctx, "commit", "-m", "Add test ticket")
	require.NoError(t, err)

	// Create the branch manually at current HEAD
	// Use git branch instead of CreateBranch to avoid checking out the branch
	_, err = gitCmd.Exec(ctx, "branch", ticketID)
	require.NoError(t, err)

	// At this point, branch and main are at the same commit
	diverged, err := gitCmd.BranchDivergedFrom(ctx, ticketID, "main")
	require.NoError(t, err)
	assert.False(t, diverged, "Branch should not have diverged initially")

	// Now try to start the ticket
	// This will:
	// 1. Change ticket status to "doing" and commit (moving main forward)
	// 2. Try to create worktree, which will detect the branch is now behind
	// 3. In non-interactive mode, automatically choose to recreate the branch
	err = app.StartTicket(ctx, ticketID, false)
	require.NoError(t, err, "StartTicket should succeed in non-interactive mode")

	// Verify worktree was created successfully
	hasWorktree, err := gitCmd.HasWorktree(ctx, ticketID)
	require.NoError(t, err)
	assert.True(t, hasWorktree, "Worktree should exist after successful start")
}

func TestBranchDivergenceErrorMessage(t *testing.T) {
	// Test the error message formatting
	err := ticketerrors.NewBranchDivergenceError("feature-123", "main", 3, 5)
	assert.Equal(t, "branch feature-123 has diverged from main (3 commits ahead, 5 behind)", err.Error())
	
	// Test error matching
	assert.True(t, errors.Is(err, ticketerrors.ErrBranchDiverged))
}