package git

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAddWorktreeWithCancelledContext tests AddWorktree with cancelled context
func TestAddWorktreeWithCancelledContext(t *testing.T) {
	git, tmpDir := setupTestGitRepo(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	worktreePath := filepath.Join(tmpDir, ".worktrees", "cancelled-branch")
	err := git.AddWorktree(ctx, worktreePath, "cancelled-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation cancelled")
}

// TestListWorktreesWithCancelledContext tests ListWorktrees with cancelled context
func TestListWorktreesWithCancelledContext(t *testing.T) {
	git, _ := setupTestGitRepo(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	worktrees, err := git.ListWorktrees(ctx)
	assert.Error(t, err)
	assert.Nil(t, worktrees)
}

// TestRemoveWorktreeWithCancelledContext tests RemoveWorktree with cancelled context
func TestRemoveWorktreeWithCancelledContext(t *testing.T) {
	git, tmpDir := setupTestGitRepo(t)

	// First add a worktree
	ctx := context.Background()
	worktreePath := filepath.Join(tmpDir, ".worktrees", "to-remove")
	err := git.AddWorktree(ctx, worktreePath, "to-remove")
	assert.NoError(t, err)

	// Now try to remove with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = git.RemoveWorktree(ctx, worktreePath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation cancelled")
}

// TestFindWorktreeByBranchWithCancelledContext tests FindWorktreeByBranch with cancelled context
func TestFindWorktreeByBranchWithCancelledContext(t *testing.T) {
	git, tmpDir := setupTestGitRepo(t)

	// First add a worktree
	ctx := context.Background()
	branch := "find-me-cancelled"
	worktreePath := filepath.Join(tmpDir, ".worktrees", branch)
	err := git.AddWorktree(ctx, worktreePath, branch)
	assert.NoError(t, err)

	// Now try to find with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	wt, err := git.FindWorktreeByBranch(ctx, branch)
	assert.Error(t, err)
	assert.Nil(t, wt)
}

// TestHasWorktreeWithCancelledContext tests HasWorktree with cancelled context
func TestHasWorktreeWithCancelledContext(t *testing.T) {
	git, tmpDir := setupTestGitRepo(t)

	// First add a worktree
	ctx := context.Background()
	branch := "has-worktree-test"
	worktreePath := filepath.Join(tmpDir, ".worktrees", branch)
	err := git.AddWorktree(ctx, worktreePath, branch)
	assert.NoError(t, err)

	// Now check with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	has, err := git.HasWorktree(ctx, branch)
	assert.Error(t, err)
	assert.False(t, has)
}

// TestRunInWorktreeWithCancelledContext tests RunInWorktree with cancelled context
func TestRunInWorktreeWithCancelledContext(t *testing.T) {
	git, tmpDir := setupTestGitRepo(t)

	// First add a worktree
	ctx := context.Background()
	branch := "run-in-worktree"
	worktreePath := filepath.Join(tmpDir, ".worktrees", branch)
	err := git.AddWorktree(ctx, worktreePath, branch)
	assert.NoError(t, err)

	// Now run command with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	output, err := git.RunInWorktree(ctx, worktreePath, "status")
	assert.Error(t, err)
	assert.Empty(t, output)
	assert.Contains(t, err.Error(), "operation cancelled")
}

// TestPruneWorktreesWithCancelledContext tests PruneWorktrees with cancelled context
func TestPruneWorktreesWithCancelledContext(t *testing.T) {
	git, _ := setupTestGitRepo(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := git.PruneWorktrees(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation cancelled")
}
