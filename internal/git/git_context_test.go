package git

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExecWithCancelledContext tests that Exec returns immediately with cancelled context
func TestExecWithCancelledContext(t *testing.T) {
	git, _ := setupTestGitRepo(t)

	// Create an already cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Try to execute a command with cancelled context
	output, err := git.Exec(ctx, "status")

	// Should fail with context error
	assert.Error(t, err)
	assert.Empty(t, output)
	assert.Contains(t, err.Error(), "operation cancelled")
}

// TestExecWithTimeout tests that Exec respects context timeout
func TestExecWithTimeout(t *testing.T) {
	git, _ := setupTestGitRepo(t)

	// Create a context with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Sleep to ensure timeout occurs
	time.Sleep(5 * time.Millisecond)

	// Try to execute a command
	output, err := git.Exec(ctx, "status")

	// Should fail with context error
	assert.Error(t, err)
	assert.Empty(t, output)
	assert.Contains(t, err.Error(), "operation cancelled")
}

// TestCurrentBranchWithCancelledContext tests CurrentBranch with cancelled context
func TestCurrentBranchWithCancelledContext(t *testing.T) {
	git, _ := setupTestGitRepo(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	branch, err := git.CurrentBranch(ctx)
	assert.Error(t, err)
	assert.Empty(t, branch)
}

// TestCreateBranchWithCancelledContext tests CreateBranch with cancelled context
func TestCreateBranchWithCancelledContext(t *testing.T) {
	git, _ := setupTestGitRepo(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := git.CreateBranch(ctx, "test-branch")
	assert.Error(t, err)
}

// TestHasUncommittedChangesWithCancelledContext tests HasUncommittedChanges with cancelled context
func TestHasUncommittedChangesWithCancelledContext(t *testing.T) {
	git, _ := setupTestGitRepo(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	hasChanges, err := git.HasUncommittedChanges(ctx)
	assert.Error(t, err)
	assert.False(t, hasChanges)
}

// TestAddWithCancelledContext tests Add with cancelled context
func TestAddWithCancelledContext(t *testing.T) {
	git, _ := setupTestGitRepo(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := git.Add(ctx, "test.txt")
	assert.Error(t, err)
}

// TestCommitWithCancelledContext tests Commit with cancelled context
func TestCommitWithCancelledContext(t *testing.T) {
	git, _ := setupTestGitRepo(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := git.Commit(ctx, "Test commit")
	assert.Error(t, err)
}

// TestCheckoutWithCancelledContext tests Checkout with cancelled context
func TestCheckoutWithCancelledContext(t *testing.T) {
	git, _ := setupTestGitRepo(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := git.Checkout(ctx, "main")
	assert.Error(t, err)
}

// TestMergeSquashWithCancelledContext tests MergeSquash with cancelled context
func TestMergeSquashWithCancelledContext(t *testing.T) {
	git, _ := setupTestGitRepo(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := git.MergeSquash(ctx, "feature-branch")
	assert.Error(t, err)
}

// TestPushWithCancelledContext tests Push with cancelled context
func TestPushWithCancelledContext(t *testing.T) {
	git, _ := setupTestGitRepo(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := git.Push(ctx, "origin", "main", false)
	assert.Error(t, err)
}

// TestIsGitRepoWithCancelledContext tests IsGitRepo with cancelled context
func TestIsGitRepoWithCancelledContext(t *testing.T) {
	_, tmpDir := setupTestGitRepo(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should return false when context is cancelled
	isRepo := IsGitRepo(ctx, tmpDir)
	assert.False(t, isRepo)
}

// TestFindProjectRootWithCancelledContext tests FindProjectRoot with cancelled context
func TestFindProjectRootWithCancelledContext(t *testing.T) {
	_, tmpDir := setupTestGitRepo(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	root, err := FindProjectRoot(ctx, tmpDir)
	assert.Error(t, err)
	assert.Empty(t, root)
}

// TestLongRunningOperationCancellation tests cancelling a long-running operation
func TestLongRunningOperationCancellation(t *testing.T) {
	git, tmpDir := setupTestGitRepo(t)

	// Create a large file to make operations potentially slower
	largePath := tmpDir + "/large.txt"
	data := make([]byte, 10*1024*1024) // 10MB
	for i := range data {
		data[i] = byte(i % 256)
	}
	err := os.WriteFile(largePath, data, 0644)
	require.NoError(t, err)

	// Add the large file
	ctx := context.Background()
	err = git.Add(ctx, "large.txt")
	require.NoError(t, err)

	// Create a context that will be cancelled during operation
	ctx, cancel := context.WithCancel(context.Background())

	// Start a goroutine that cancels after a short delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	// Try to commit - this might be cancelled mid-operation
	err = git.Commit(ctx, "Add large file")

	// The operation may succeed if it completes before cancellation,
	// or fail if cancelled in time
	if err != nil {
		// When context is cancelled, the command may be killed with SIGKILL
		// which results in "signal: killed" error message, or it may show
		// "operation cancelled" if caught early
		errStr := err.Error()
		assert.True(t, strings.Contains(errStr, "operation cancelled") ||
			strings.Contains(errStr, "signal: killed"),
			"Expected error to contain 'operation cancelled' or 'signal: killed', got: %s", errStr)
	}
}

// TestContextPropagationInExec tests that context is properly propagated to exec.CommandContext
func TestContextPropagationInExec(t *testing.T) {
	git, _ := setupTestGitRepo(t)

	// Test with a deadline context
	deadline := time.Now().Add(100 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	// Execute a simple command
	output, err := git.Exec(ctx, "status", "--porcelain")

	// Should succeed as the command is fast
	assert.NoError(t, err)
	assert.NotNil(t, output)

	// Wait for deadline to pass
	time.Sleep(150 * time.Millisecond)

	// Now the context should be expired
	_, err = git.Exec(ctx, "status")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation cancelled")
}

// BenchmarkContextCheckOverhead benchmarks the overhead of context checking
func BenchmarkContextCheckOverhead(b *testing.B) {
	// Setup outside the benchmark loop
	tmpDir := b.TempDir()
	git := New(tmpDir)
	ctx := context.Background()

	// Initialize repo
	_, err := git.Exec(ctx, "init")
	require.NoError(b, err)
	_, err = git.Exec(ctx, "config", "user.name", "Test User")
	require.NoError(b, err)
	_, err = git.Exec(ctx, "config", "user.email", "test@example.com")
	require.NoError(b, err)

	// Create initial commit
	readmePath := filepath.Join(tmpDir, "README.md")
	err = os.WriteFile(readmePath, []byte("# Test\n"), 0644)
	require.NoError(b, err)
	_, err = git.Exec(ctx, "add", "README.md")
	require.NoError(b, err)
	_, err = git.Exec(ctx, "commit", "-m", "Initial commit")
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = git.CurrentBranch(ctx)
	}
}

// BenchmarkContextCheckWithCancellation benchmarks operations with cancelled contexts
func BenchmarkContextCheckWithCancellation(b *testing.B) {
	// Setup outside the benchmark loop
	tmpDir := b.TempDir()
	git := New(tmpDir)
	ctx := context.Background()

	// Initialize repo
	_, err := git.Exec(ctx, "init")
	require.NoError(b, err)
	_, err = git.Exec(ctx, "config", "user.name", "Test User")
	require.NoError(b, err)
	_, err = git.Exec(ctx, "config", "user.email", "test@example.com")
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, _ = git.CurrentBranch(ctx)
	}
}
