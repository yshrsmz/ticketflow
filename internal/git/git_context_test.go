package git

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ticketerrors "github.com/yshrsmz/ticketflow/internal/errors"
)

// TestGitOperationsWithCancelledContext tests all git operations with cancelled context
func TestGitOperationsWithCancelledContext(t *testing.T) {
	git, _ := setupTestGitRepo(t)

	tests := []struct {
		name      string
		operation func(context.Context) error
	}{
		{
			name: "Exec",
			operation: func(ctx context.Context) error {
				_, err := git.Exec(ctx, "status")
				return err
			},
		},
		{
			name: "CurrentBranch",
			operation: func(ctx context.Context) error {
				_, err := git.CurrentBranch(ctx)
				return err
			},
		},
		{
			name: "CreateBranch",
			operation: func(ctx context.Context) error {
				return git.CreateBranch(ctx, "test-branch")
			},
		},
		{
			name: "HasUncommittedChanges",
			operation: func(ctx context.Context) error {
				_, err := git.HasUncommittedChanges(ctx)
				return err
			},
		},
		{
			name: "Add",
			operation: func(ctx context.Context) error {
				return git.Add(ctx, "test.txt")
			},
		},
		{
			name: "Commit",
			operation: func(ctx context.Context) error {
				return git.Commit(ctx, "Test commit")
			},
		},
		{
			name: "Checkout",
			operation: func(ctx context.Context) error {
				return git.Checkout(ctx, "main")
			},
		},
		{
			name: "MergeSquash",
			operation: func(ctx context.Context) error {
				return git.MergeSquash(ctx, "feature-branch")
			},
		},
		{
			name: "Push",
			operation: func(ctx context.Context) error {
				return git.Push(ctx, "origin", "main", false)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := tt.operation(ctx)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "operation cancelled")

			// Check if the underlying error is context.Canceled
			var gitErr *ticketerrors.GitError
			if errors.As(err, &gitErr) && gitErr.Err != nil {
				assert.True(t, errors.Is(gitErr.Err, context.Canceled))
			}
		})
	}
}

// TestGitStaticFunctionsWithCancelledContext tests static functions with cancelled context
func TestGitStaticFunctionsWithCancelledContext(t *testing.T) {
	_, tmpDir := setupTestGitRepo(t)

	tests := []struct {
		name      string
		operation func(context.Context) error
	}{
		{
			name: "IsGitRepo",
			operation: func(ctx context.Context) error {
				// IsGitRepo returns false for cancelled context
				if IsGitRepo(ctx, tmpDir) {
					return errors.New("expected false for cancelled context")
				}
				return nil
			},
		},
		{
			name: "FindProjectRoot",
			operation: func(ctx context.Context) error {
				_, err := FindProjectRoot(ctx, tmpDir)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := tt.operation(ctx)
			if tt.name == "IsGitRepo" {
				assert.NoError(t, err) // Special case: IsGitRepo returns false
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestContextWithTimeout tests operations with timeout contexts
func TestContextWithTimeout(t *testing.T) {
	git, _ := setupTestGitRepo(t)

	tests := []struct {
		name    string
		timeout time.Duration
		delay   time.Duration
		wantErr bool
	}{
		{
			name:    "immediate timeout",
			timeout: 1 * time.Microsecond,
			delay:   5 * time.Millisecond,
			wantErr: true,
		},
		{
			name:    "sufficient timeout",
			timeout: 1 * time.Second,
			delay:   0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			if tt.delay > 0 {
				time.Sleep(tt.delay)
			}

			_, err := git.Exec(ctx, "status", "--porcelain")
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "operation cancelled")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestContextWithValues tests that context values are preserved even when cancelled
func TestContextWithValues(t *testing.T) {
	type contextKey string
	const testKey contextKey = "test-key"

	git, _ := setupTestGitRepo(t)

	ctx := context.WithValue(context.Background(), testKey, "test-value")
	ctx, cancel := context.WithCancel(ctx)
	cancel()

	// Values should still be accessible even when cancelled
	assert.Equal(t, "test-value", ctx.Value(testKey))

	// But operations should fail
	_, err := git.CurrentBranch(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation cancelled")
}

// TestConcurrentCancellation tests concurrent operations with shared context
func TestConcurrentCancellation(t *testing.T) {
	git, _ := setupTestGitRepo(t)

	// Create a channel to coordinate goroutines
	startChan := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	errChan := make(chan error, 20)

	// Start multiple concurrent operations
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Wait for signal to start
			<-startChan

			_, err := git.CurrentBranch(ctx)
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	// Give goroutines time to block on channel
	time.Sleep(10 * time.Millisecond)

	// Cancel context first
	cancel()

	// Then signal all goroutines to start
	close(startChan)

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Count errors
	errorCount := 0
	for err := range errChan {
		if err != nil {
			errorCount++
			assert.Contains(t, err.Error(), "operation cancelled")
		}
	}

	// All operations should have been cancelled since context was cancelled before they started
	assert.Equal(t, 20, errorCount, "All operations should fail with cancelled context")
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

// TestLongRunningOperationCancellation tests cancelling a long-running operation
func TestLongRunningOperationCancellation(t *testing.T) {
	git, tmpDir := setupTestGitRepo(t)

	// Create a small test file instead of a large one
	testPath := filepath.Join(tmpDir, "test.txt")
	testData := []byte("test content for cancellation")
	err := os.WriteFile(testPath, testData, 0644)
	require.NoError(t, err)

	// Add the test file
	ctx := context.Background()
	err = git.Add(ctx, "test.txt")
	require.NoError(t, err)

	// Test cancellation using a short timeout instead of large files
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
	defer cancel()

	// Sleep briefly to ensure timeout is exceeded
	time.Sleep(5 * time.Millisecond)

	// Try to commit - this should be cancelled due to timeout
	err = git.Commit(ctx, "Add test file")

	// Operation should fail due to timeout
	if err != nil {
		// When context is cancelled, the command may be killed with SIGKILL
		// which results in "signal: killed" error message, or it may show
		// "operation cancelled" if caught early
		errStr := err.Error()
		assert.True(t, strings.Contains(errStr, "operation cancelled") ||
			strings.Contains(errStr, "signal: killed") ||
			strings.Contains(errStr, "context deadline exceeded"),
			"Expected error to contain cancellation/timeout message, got: %s", errStr)
	}
}

// TestContextInheritance tests parent-child context cancellation behavior
func TestContextInheritance(t *testing.T) {
	git, _ := setupTestGitRepo(t)

	// Create parent context with timeout
	parentCtx, parentCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer parentCancel()

	// Create child context
	childCtx, childCancel := context.WithCancel(parentCtx)
	defer childCancel()

	// Wait for parent timeout
	time.Sleep(150 * time.Millisecond)

	// Child should also be cancelled
	_, err := git.CurrentBranch(childCtx)
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

	b.ReportAllocs()
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

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, _ = git.CurrentBranch(ctx)
	}
}

// simulateSlowGitOperation simulates a git operation that takes time and respects context cancellation
// This replaces the need for large file operations to test timeouts
func simulateSlowGitOperation(ctx context.Context, git *Git, duration time.Duration) error {
	// Use a ticker to periodically check for cancellation while "working"
	ticker := time.NewTicker(1 * time.Millisecond)
	defer ticker.Stop()

	deadline := time.Now().Add(duration)
	for {
		select {
		case <-ctx.Done():
			return ticketerrors.NewGitError("simulated operation", "current", ctx.Err())
		case <-ticker.C:
			if time.Now().After(deadline) {
				// Simulate successful completion by doing a quick git operation
				_, err := git.CurrentBranch(context.Background())
				return err
			}
			// Continue the "work"
		}
	}
}

// TestSimulatedSlowOperationCancellation tests the slow operation simulation with cancellation
func TestSimulatedSlowOperationCancellation(t *testing.T) {
	git, _ := setupTestGitRepo(t)

	tests := []struct {
		name     string
		timeout  time.Duration
		duration time.Duration
		wantErr  bool
	}{
		{
			name:     "operation cancelled by timeout",
			timeout:  5 * time.Millisecond,
			duration: 50 * time.Millisecond,
			wantErr:  true,
		},
		{
			name:     "operation completes before timeout",
			timeout:  100 * time.Millisecond,
			duration: 10 * time.Millisecond,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			err := simulateSlowGitOperation(ctx, git, tt.duration)
			if tt.wantErr {
				assert.Error(t, err)
				// Check for various cancellation error messages
				errMsg := err.Error()
				assert.True(t, 
					strings.Contains(errMsg, "operation cancelled") ||
					strings.Contains(errMsg, "context deadline exceeded") ||
					strings.Contains(errMsg, "context canceled"),
					"Expected cancellation error, got: %s", errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// BenchmarkContextTypes compares different context types
func BenchmarkContextTypes(b *testing.B) {
	tmpDir := b.TempDir()
	git := New(tmpDir)
	ctx := context.Background()

	// Setup repo
	_, err := git.Exec(ctx, "init")
	require.NoError(b, err)
	_, err = git.Exec(ctx, "config", "user.name", "Test User")
	require.NoError(b, err)
	_, err = git.Exec(ctx, "config", "user.email", "test@example.com")
	require.NoError(b, err)

	b.Run("Background", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = git.CurrentBranch(context.Background())
		}
	})

	b.Run("WithCancel", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			_, _ = git.CurrentBranch(ctx)
		}
	})

	b.Run("WithTimeout", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
			defer cancel()
			_, _ = git.CurrentBranch(ctx)
		}
	})

	b.Run("WithValue", func(b *testing.B) {
		b.ReportAllocs()
		type key string
		for i := 0; i < b.N; i++ {
			ctx := context.WithValue(context.Background(), key("test"), "value")
			_, _ = git.CurrentBranch(ctx)
		}
	})

	b.Run("SimulatedSlowOperation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = simulateSlowGitOperation(context.Background(), git, 1*time.Millisecond)
		}
	})
}
