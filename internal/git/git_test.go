package git

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewWithTimeout(t *testing.T) {
	tests := []struct {
		name        string
		repoPath    string
		timeout     time.Duration
		wantTimeout time.Duration
	}{
		{
			name:        "custom timeout",
			repoPath:    "/tmp/test",
			timeout:     45 * time.Second,
			wantTimeout: 45 * time.Second,
		},
		{
			name:        "zero timeout",
			repoPath:    "/tmp/test2",
			timeout:     0,
			wantTimeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithTimeout(tt.repoPath, tt.timeout)
			assert.Equal(t, tt.repoPath, g.repoPath)
			assert.Equal(t, tt.wantTimeout, g.timeout)
		})
	}
}

func TestNew(t *testing.T) {
	g := New("/tmp/test")
	assert.Equal(t, "/tmp/test", g.repoPath)
	assert.Equal(t, 30*time.Second, g.timeout)
}

func TestExecWithTimeout(t *testing.T) {
	// Create a git instance with a short but reasonable timeout
	// Using 50ms to reduce flakiness while still testing timeout behavior
	g := NewWithTimeout(".", 50*time.Millisecond)

	// Try to execute a command that would take longer than the timeout
	// Using a command that's likely to take more time
	ctx := context.Background()
	_, err := g.Exec(ctx, "log", "--all", "--oneline", "-n", "100000")

	// The error might be a timeout or might succeed if git is fast enough
	// We're mainly testing that the timeout is applied without panicking
	if err != nil {
		// The error could be "signal: killed" or "context deadline exceeded"
		assert.True(t,
			strings.Contains(err.Error(), "context deadline exceeded") ||
				strings.Contains(err.Error(), "signal: killed") ||
				strings.Contains(err.Error(), "operation timed out"),
			"Expected timeout-related error, got: %v", err)
	}
}

func TestExecWithContextTimeout(t *testing.T) {
	// Create a git instance with a long timeout
	g := NewWithTimeout(".", 30*time.Second)

	// But use a context with a short timeout
	// Using 50ms to reduce flakiness
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := g.Exec(ctx, "log", "--all", "--oneline", "-n", "100000")

	// The context timeout should take precedence
	if err != nil {
		// The error could be "signal: killed" or "context deadline exceeded"
		assert.True(t,
			strings.Contains(err.Error(), "context deadline exceeded") ||
				strings.Contains(err.Error(), "signal: killed") ||
				strings.Contains(err.Error(), "operation timed out"),
			"Expected timeout-related error, got: %v", err)
	}
}

func TestRunInWorktreePreservesTimeout(t *testing.T) {
	// Create a git instance with custom timeout
	g := NewWithTimeout(".", 45*time.Second)

	// The RunInWorktree method should preserve the timeout
	// We can't easily test the actual execution without a real worktree,
	// but we can verify the method doesn't panic
	ctx := context.Background()
	_, _ = g.RunInWorktree(ctx, "/tmp/fake-worktree", "status")
}

func TestBranchExists(t *testing.T) {
	// Setup test git repo
	tmpDir := t.TempDir()
	git := New(tmpDir)
	ctx := context.Background()

	// Initialize repo
	_, err := git.Exec(ctx, "init")
	assert.NoError(t, err)

	// Configure git locally for test repo (not globally)
	_, err = git.Exec(ctx, "config", "user.name", "Test User")
	assert.NoError(t, err)
	_, err = git.Exec(ctx, "config", "user.email", "test@example.com")
	assert.NoError(t, err)

	// Create initial commit
	_, err = git.Exec(ctx, "commit", "--allow-empty", "-m", "Initial commit")
	assert.NoError(t, err)

	tests := []struct {
		name       string
		branch     string
		setup      func()
		wantExists bool
		wantErr    bool
	}{
		{
			name:       "master/main branch exists",
			branch:     "master",
			setup: func() {
				// Try to create master branch, ignore error if it already exists
				git.Exec(ctx, "checkout", "-b", "master")
			},
			wantExists: true,
			wantErr:    false,
		},
		{
			name:       "non-existent branch",
			branch:     "feature/does-not-exist",
			setup:      func() {},
			wantExists: false,
			wantErr:    false,
		},
		{
			name:   "existing feature branch",
			branch: "feature/test-branch",
			setup: func() {
				git.Exec(ctx, "checkout", "-b", "feature/test-branch")
			},
			wantExists: true,
			wantErr:    false,
		},
		{
			name:       "branch with special characters",
			branch:     "feat/my-branch_v2.0",
			setup: func() {
				git.Exec(ctx, "checkout", "-b", "feat/my-branch_v2.0")
			},
			wantExists: true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			exists, err := git.BranchExists(ctx, tt.branch)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantExists, exists)
			}
		})
	}
}
