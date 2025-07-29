package git

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestGitRepo(t *testing.T) (*Git, string) {
	tmpDir := t.TempDir()
	git := New(tmpDir)

	// Initialize repo
	_, err := git.Exec("init")
	require.NoError(t, err)

	// Set git config
	_, err = git.Exec("config", "user.name", "Test User")
	require.NoError(t, err)
	_, err = git.Exec("config", "user.email", "test@example.com")
	require.NoError(t, err)

	// Create initial commit
	readmePath := filepath.Join(tmpDir, "README.md")
	err = os.WriteFile(readmePath, []byte("# Test\n"), 0644)
	require.NoError(t, err)

	_, err = git.Exec("add", "README.md")
	require.NoError(t, err)
	_, err = git.Exec("commit", "-m", "Initial commit")
	require.NoError(t, err)

	return git, tmpDir
}

func TestAddWorktree(t *testing.T) {
	git, tmpDir := setupTestGitRepo(t)

	// Add a worktree
	worktreePath := filepath.Join(tmpDir, ".worktrees", "test-branch")
	err := git.AddWorktree(worktreePath, "test-branch")
	require.NoError(t, err)

	// Verify worktree exists
	_, err = os.Stat(worktreePath)
	require.NoError(t, err)

	// Verify worktree is in list
	worktrees, err := git.ListWorktrees()
	require.NoError(t, err)
	assert.Len(t, worktrees, 2) // main + new worktree

	// Find the new worktree
	found := false
	for _, wt := range worktrees {
		if wt.Branch == "test-branch" {
			found = true
			// Resolve symlinks before comparing paths (macOS compatibility)
			expectedPath, err := filepath.EvalSymlinks(worktreePath)
			require.NoError(t, err)
			actualPath, err := filepath.EvalSymlinks(wt.Path)
			require.NoError(t, err)
			assert.Equal(t, expectedPath, actualPath)
			break
		}
	}
	assert.True(t, found, "Worktree not found in list")
}

func TestListWorktrees(t *testing.T) {
	git, tmpDir := setupTestGitRepo(t)

	// Initially should have just the main worktree
	worktrees, err := git.ListWorktrees()
	require.NoError(t, err)
	assert.Len(t, worktrees, 1)
	// Resolve symlinks before comparing paths (macOS compatibility)
	expectedPath, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)
	actualPath, err := filepath.EvalSymlinks(worktrees[0].Path)
	require.NoError(t, err)
	assert.Equal(t, expectedPath, actualPath)

	// Add multiple worktrees
	for i := 1; i <= 3; i++ {
		branch := fmt.Sprintf("feature-%d", i)
		path := filepath.Join(tmpDir, ".worktrees", branch)
		err := git.AddWorktree(path, branch)
		require.NoError(t, err)
	}

	// Should now have 4 worktrees
	worktrees, err = git.ListWorktrees()
	require.NoError(t, err)
	assert.Len(t, worktrees, 4)
}

func TestRemoveWorktree(t *testing.T) {
	git, tmpDir := setupTestGitRepo(t)

	// Add a worktree
	worktreePath := filepath.Join(tmpDir, ".worktrees", "temp-branch")
	err := git.AddWorktree(worktreePath, "temp-branch")
	require.NoError(t, err)

	// Verify it exists
	worktrees, err := git.ListWorktrees()
	require.NoError(t, err)
	assert.Len(t, worktrees, 2)

	// Remove the worktree
	err = git.RemoveWorktree(worktreePath)
	require.NoError(t, err)

	// Verify it's gone
	worktrees, err = git.ListWorktrees()
	require.NoError(t, err)
	assert.Len(t, worktrees, 1)

	// Directory should be removed
	_, err = os.Stat(worktreePath)
	assert.True(t, os.IsNotExist(err))
}

func TestFindWorktreeByBranch(t *testing.T) {
	git, tmpDir := setupTestGitRepo(t)

	// Add a worktree
	branch := "find-me"
	worktreePath := filepath.Join(tmpDir, ".worktrees", branch)
	err := git.AddWorktree(worktreePath, branch)
	require.NoError(t, err)

	// Find it by branch
	wt, err := git.FindWorktreeByBranch(branch)
	require.NoError(t, err)
	require.NotNil(t, wt)
	assert.Equal(t, branch, wt.Branch)
	// Resolve symlinks before comparing paths (macOS compatibility)
	expectedPath, err := filepath.EvalSymlinks(worktreePath)
	require.NoError(t, err)
	actualPath, err := filepath.EvalSymlinks(wt.Path)
	require.NoError(t, err)
	assert.Equal(t, expectedPath, actualPath)

	// Try to find non-existent branch
	wt, err = git.FindWorktreeByBranch("non-existent")
	require.NoError(t, err)
	assert.Nil(t, wt)
}

func TestHasWorktree(t *testing.T) {
	git, tmpDir := setupTestGitRepo(t)

	// Initially no worktree for branch
	has, err := git.HasWorktree("test-branch")
	require.NoError(t, err)
	assert.False(t, has)

	// Add worktree
	worktreePath := filepath.Join(tmpDir, ".worktrees", "test-branch")
	err = git.AddWorktree(worktreePath, "test-branch")
	require.NoError(t, err)

	// Now should have worktree
	has, err = git.HasWorktree("test-branch")
	require.NoError(t, err)
	assert.True(t, has)
}

func TestRunInWorktree(t *testing.T) {
	git, tmpDir := setupTestGitRepo(t)

	// Add a worktree
	branch := "work-branch"
	worktreePath := filepath.Join(tmpDir, ".worktrees", branch)
	err := git.AddWorktree(worktreePath, branch)
	require.NoError(t, err)

	// Create a file in the worktree
	testFile := filepath.Join(worktreePath, "test.txt")
	err = os.WriteFile(testFile, []byte("test content\n"), 0644)
	require.NoError(t, err)

	// Run git status in the worktree
	output, err := git.RunInWorktree(worktreePath, "status", "--porcelain")
	require.NoError(t, err)
	assert.Contains(t, output, "test.txt")

	// Add and commit in the worktree
	_, err = git.RunInWorktree(worktreePath, "add", "test.txt")
	require.NoError(t, err)
	_, err = git.RunInWorktree(worktreePath, "commit", "-m", "Add test file")
	require.NoError(t, err)

	// Verify the commit
	output, err = git.RunInWorktree(worktreePath, "log", "--oneline", "-1")
	require.NoError(t, err)
	assert.Contains(t, output, "Add test file")
}

func TestPruneWorktrees(t *testing.T) {
	git, tmpDir := setupTestGitRepo(t)

	// Add a worktree
	worktreePath := filepath.Join(tmpDir, ".worktrees", "prune-test")
	err := git.AddWorktree(worktreePath, "prune-test")
	require.NoError(t, err)

	// Manually delete the worktree directory (simulating external deletion)
	err = os.RemoveAll(worktreePath)
	require.NoError(t, err)

	// Prune should clean up the stale reference
	err = git.PruneWorktrees()
	require.NoError(t, err)

	// List should not include the deleted worktree
	worktrees, err := git.ListWorktrees()
	require.NoError(t, err)
	for _, wt := range worktrees {
		assert.NotEqual(t, "prune-test", wt.Branch)
	}
}
