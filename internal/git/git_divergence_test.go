package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/testutil"
)

// initGitRepo initializes a git repository in the given directory
func initGitRepo(t *testing.T, ctx context.Context, dir string) error {
	g := New(dir)

	// Initialize repo with main as default branch
	// Note: --initial-branch requires git 2.28+, we have a fallback for older versions
	if _, err := g.Exec(ctx, "init", "--initial-branch=main"); err != nil {
		// Fallback for older git versions (< 2.28)
		if _, err := g.Exec(ctx, "init"); err != nil {
			return err
		}
		// Try to rename branch to main
		_, _ = g.Exec(ctx, "branch", "-m", "main")
	}

	// Set git config
	testutil.GitConfigApply(t, g)

	return nil
}

func TestGetDefaultBranch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupFunc     func(t *testing.T, g *Git)
		wantBranch    string
		wantErr       bool
		skipRemoteRef bool
	}{
		{
			name: "detects_from_origin_HEAD",
			setupFunc: func(t *testing.T, g *Git) {
				// Setup a remote with HEAD pointing to main
				ctx := context.Background()
				_, err := g.Exec(ctx, "remote", "add", "origin", ".")
				require.NoError(t, err)
				_, err = g.Exec(ctx, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
				require.NoError(t, err)
			},
			wantBranch: "main",
		},
		{
			name: "fallback_to_main_branch",
			setupFunc: func(t *testing.T, g *Git) {
				// main branch already exists from setup
			},
			wantBranch:    "main",
			skipRemoteRef: true,
		},
		{
			name: "fallback_to_master_branch",
			setupFunc: func(t *testing.T, g *Git) {
				ctx := context.Background()
				// Delete main branch and create master
				_, err := g.Exec(ctx, "checkout", "-b", "master")
				require.NoError(t, err)
				_, err = g.Exec(ctx, "branch", "-d", "main")
				require.NoError(t, err)
			},
			wantBranch:    "master",
			skipRemoteRef: true,
		},
		{
			name: "error_when_no_default_branch",
			setupFunc: func(t *testing.T, g *Git) {
				ctx := context.Background()
				// Create and switch to a non-standard branch
				_, err := g.Exec(ctx, "checkout", "-b", "develop")
				require.NoError(t, err)
				// Delete both main and master
				_, err = g.Exec(ctx, "branch", "-d", "main")
				require.NoError(t, err)
			},
			wantErr:       true,
			skipRemoteRef: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temporary directory
			tmpDir := t.TempDir()

			// Initialize git repo
			ctx := context.Background()
			g := New(tmpDir)
			require.NoError(t, initGitRepo(t, ctx, tmpDir))

			// Create initial commit on main
			testFile := filepath.Join(tmpDir, "test.txt")
			require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))
			_, err := g.Exec(ctx, "add", ".")
			require.NoError(t, err)
			_, err = g.Exec(ctx, "commit", "-m", "Initial commit")
			require.NoError(t, err)

			// Run setup function
			if tt.setupFunc != nil {
				tt.setupFunc(t, g)
			}

			// Test GetDefaultBranch
			branch, err := g.GetDefaultBranch(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantBranch, branch)
		})
	}
}

func TestBranchDivergence(t *testing.T) {
	t.Parallel()

	// Create temporary directory
	tmpDir := t.TempDir()

	// Initialize git repo
	ctx := context.Background()
	g := New(tmpDir)
	require.NoError(t, initGitRepo(t, ctx, tmpDir))

	// Create initial commit on main
	testFile := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("initial"), 0644))
	_, err := g.Exec(ctx, "add", ".")
	require.NoError(t, err)
	_, err = g.Exec(ctx, "commit", "-m", "Initial commit")
	require.NoError(t, err)

	// Get main branch commit
	mainCommit, err := g.GetBranchCommit(ctx, "main")
	require.NoError(t, err)
	assert.NotEmpty(t, mainCommit)

	// Create feature branch
	err = g.CreateBranch(ctx, "feature")
	require.NoError(t, err)

	// Test no divergence initially
	diverged, err := g.BranchDivergedFrom(ctx, "feature", "main")
	require.NoError(t, err)
	assert.False(t, diverged)

	// Test divergence info - should be 0/0
	ahead, behind, err := g.GetBranchDivergenceInfo(ctx, "feature", "main")
	require.NoError(t, err)
	assert.Equal(t, 0, ahead)
	assert.Equal(t, 0, behind)

	// Make commit on feature branch
	err = g.Checkout(ctx, "feature")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(testFile, []byte("feature change"), 0644))
	_, err = g.Exec(ctx, "add", ".")
	require.NoError(t, err)
	_, err = g.Exec(ctx, "commit", "-m", "Feature commit")
	require.NoError(t, err)

	// Test divergence - feature is ahead
	diverged, err = g.BranchDivergedFrom(ctx, "feature", "main")
	require.NoError(t, err)
	assert.True(t, diverged)

	// Test divergence info - should be 1 ahead, 0 behind
	ahead, behind, err = g.GetBranchDivergenceInfo(ctx, "feature", "main")
	require.NoError(t, err)
	assert.Equal(t, 1, ahead)
	assert.Equal(t, 0, behind)

	// Make commit on main branch
	err = g.Checkout(ctx, "main")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(testFile, []byte("main change"), 0644))
	_, err = g.Exec(ctx, "add", ".")
	require.NoError(t, err)
	_, err = g.Exec(ctx, "commit", "-m", "Main commit")
	require.NoError(t, err)

	// Test divergence info - should be 1 ahead, 1 behind
	ahead, behind, err = g.GetBranchDivergenceInfo(ctx, "feature", "main")
	require.NoError(t, err)
	assert.Equal(t, 1, ahead)
	assert.Equal(t, 1, behind)
}

func TestGetBranchCommit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		branch  string
		wantErr bool
		errMsg  string
	}{
		{
			name:   "valid_branch",
			branch: "main",
		},
		{
			name:    "invalid_branch_name",
			branch:  "branch with spaces",
			wantErr: true,
			errMsg:  "invalid branch name",
		},
		{
			name:    "non_existent_branch",
			branch:  "non-existent",
			wantErr: true,
			errMsg:  "failed to get commit",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temporary directory
			tmpDir := t.TempDir()

			// Initialize git repo
			ctx := context.Background()
			g := New(tmpDir)
			require.NoError(t, initGitRepo(t, ctx, tmpDir))

			// Create initial commit
			testFile := filepath.Join(tmpDir, "test.txt")
			require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))
			_, err := g.Exec(ctx, "add", ".")
			require.NoError(t, err)
			_, err = g.Exec(ctx, "commit", "-m", "Initial commit")
			require.NoError(t, err)

			// Test GetBranchCommit
			commit, err := g.GetBranchCommit(ctx, tt.branch)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, commit)
			// Commit hash should be 40 characters
			assert.Len(t, commit, 40)
		})
	}
}

func TestIsBranchMerged(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tmpDir := t.TempDir()

	// Initialize git repo
	require.NoError(t, initGitRepo(t, ctx, tmpDir))

	g := New(tmpDir)

	// Create initial commit on main
	testFile := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("initial"), 0644))
	_, err := g.Exec(ctx, "add", "test.txt")
	require.NoError(t, err)
	_, err = g.Exec(ctx, "commit", "-m", "Initial commit")
	require.NoError(t, err)

	t.Run("unmerged branch", func(t *testing.T) {
		// Create and checkout new branch
		_, err := g.Exec(ctx, "checkout", "-b", "feature-unmerged")
		require.NoError(t, err)

		// Add a commit
		require.NoError(t, os.WriteFile(testFile, []byte("feature change"), 0644))
		_, err = g.Exec(ctx, "add", "test.txt")
		require.NoError(t, err)
		_, err = g.Exec(ctx, "commit", "-m", "Feature commit")
		require.NoError(t, err)

		// Go back to main
		_, err = g.Exec(ctx, "checkout", "main")
		require.NoError(t, err)

		// Check if feature branch is merged (should be false)
		merged, err := g.IsBranchMerged(ctx, "feature-unmerged", "main")
		require.NoError(t, err)
		assert.False(t, merged, "unmerged branch should not be marked as merged")
	})

	t.Run("merged branch", func(t *testing.T) {
		// Create and checkout new branch
		_, err := g.Exec(ctx, "checkout", "-b", "feature-merged")
		require.NoError(t, err)

		// Add a commit
		mergeFile := filepath.Join(tmpDir, "merge.txt")
		require.NoError(t, os.WriteFile(mergeFile, []byte("merge content"), 0644))
		_, err = g.Exec(ctx, "add", "merge.txt")
		require.NoError(t, err)
		_, err = g.Exec(ctx, "commit", "-m", "Merge feature commit")
		require.NoError(t, err)

		// Go back to main and merge
		_, err = g.Exec(ctx, "checkout", "main")
		require.NoError(t, err)
		_, err = g.Exec(ctx, "merge", "feature-merged")
		require.NoError(t, err)

		// Check if feature branch is merged (should be true)
		merged, err := g.IsBranchMerged(ctx, "feature-merged", "main")
		require.NoError(t, err)
		assert.True(t, merged, "merged branch should be marked as merged")
	})

	t.Run("current branch", func(t *testing.T) {
		// Check if main is merged into itself (should be true)
		merged, err := g.IsBranchMerged(ctx, "main", "main")
		require.NoError(t, err)
		assert.True(t, merged, "current branch should be marked as merged into itself")
	})

	t.Run("nonexistent branch", func(t *testing.T) {
		// Check if nonexistent branch is merged (should be false, no error)
		merged, err := g.IsBranchMerged(ctx, "nonexistent", "main")
		require.NoError(t, err)
		assert.False(t, merged, "nonexistent branch should not be marked as merged")
	})

	t.Run("nonexistent target branch", func(t *testing.T) {
		// Check against nonexistent target branch (should return false, no error per implementation)
		merged, err := g.IsBranchMerged(ctx, "main", "nonexistent-target")
		// The implementation returns false when the branch doesn't exist
		require.NoError(t, err)
		assert.False(t, merged, "should return false for nonexistent target branch")
	})
}
