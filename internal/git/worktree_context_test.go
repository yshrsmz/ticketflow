package git

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWorktreeOperationsWithCancelledContext tests all worktree operations with cancelled context
func TestWorktreeOperationsWithCancelledContext(t *testing.T) {
	git, tmpDir := setupTestGitRepo(t)

	// Setup a worktree for testing operations that need an existing worktree
	ctx := context.Background()
	existingWorktreePath := filepath.Join(tmpDir, ".worktrees", "existing")
	err := git.AddWorktree(ctx, existingWorktreePath, "existing-branch")
	require.NoError(t, err)

	tests := []struct {
		name      string
		setup     func() (context.Context, context.CancelFunc)
		operation func(context.Context) error
	}{
		{
			name: "AddWorktree",
			setup: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			operation: func(ctx context.Context) error {
				worktreePath := filepath.Join(tmpDir, ".worktrees", "cancelled-branch")
				return git.AddWorktree(ctx, worktreePath, "cancelled-branch")
			},
		},
		{
			name: "ListWorktrees",
			setup: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			operation: func(ctx context.Context) error {
				_, err := git.ListWorktrees(ctx)
				return err
			},
		},
		{
			name: "RemoveWorktree",
			setup: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			operation: func(ctx context.Context) error {
				return git.RemoveWorktree(ctx, existingWorktreePath)
			},
		},
		{
			name: "FindWorktreeByBranch",
			setup: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			operation: func(ctx context.Context) error {
				_, err := git.FindWorktreeByBranch(ctx, "existing-branch")
				return err
			},
		},
		{
			name: "HasWorktree",
			setup: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			operation: func(ctx context.Context) error {
				_, err := git.HasWorktree(ctx, "existing-branch")
				return err
			},
		},
		{
			name: "RunInWorktree",
			setup: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			operation: func(ctx context.Context) error {
				_, err := git.RunInWorktree(ctx, existingWorktreePath, "status")
				return err
			},
		},
		{
			name: "PruneWorktrees",
			setup: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			operation: func(ctx context.Context) error {
				return git.PruneWorktrees(ctx)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.setup()
			cancel() // Cancel immediately

			err := tt.operation(ctx)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "operation cancelled")
		})
	}
}

// TestWorktreeOperationsWithTimeout tests worktree operations with timeout
func TestWorktreeOperationsWithTimeout(t *testing.T) {
	git, _ := setupTestGitRepo(t)

	tests := []struct {
		name      string
		timeout   time.Duration
		operation func(context.Context) error
		wantErr   bool
	}{
		{
			name:    "ListWorktrees with short timeout",
			timeout: 1 * time.Microsecond,
			operation: func(ctx context.Context) error {
				time.Sleep(5 * time.Millisecond) // Ensure timeout
				_, err := git.ListWorktrees(ctx)
				return err
			},
			wantErr: true,
		},
		{
			name:    "ListWorktrees with sufficient timeout",
			timeout: 1 * time.Second,
			operation: func(ctx context.Context) error {
				_, err := git.ListWorktrees(ctx)
				return err
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			err := tt.operation(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestConcurrentWorktreeOperations tests concurrent worktree operations
func TestConcurrentWorktreeOperations(t *testing.T) {
	git, tmpDir := setupTestGitRepo(t)

	// Create some worktrees first
	for i := 0; i < 3; i++ {
		branch := fmt.Sprintf("test-branch-%d", i)
		path := filepath.Join(tmpDir, ".worktrees", branch)
		err := git.AddWorktree(context.Background(), path, branch)
		require.NoError(t, err)
	}

	// Create coordination channel
	startChan := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	errChan := make(chan error, 30)

	// Start multiple concurrent operations
	for i := 0; i < 10; i++ {
		wg.Add(3)

		// List operation
		go func() {
			defer wg.Done()
			<-startChan
			_, err := git.ListWorktrees(ctx)
			if err != nil {
				errChan <- err
			}
		}()

		// Find operation
		go func() {
			defer wg.Done()
			<-startChan
			_, err := git.FindWorktreeByBranch(ctx, "main")
			if err != nil {
				errChan <- err
			}
		}()

		// Has operation
		go func() {
			defer wg.Done()
			<-startChan
			_, err := git.HasWorktree(ctx, "main")
			if err != nil {
				errChan <- err
			}
		}()
	}

	// Give goroutines time to block
	time.Sleep(10 * time.Millisecond)

	// Cancel context first
	cancel()

	// Then start all operations
	close(startChan)

	// Wait for all operations to complete
	wg.Wait()
	close(errChan)

	// Count cancelled operations
	errorCount := 0
	for err := range errChan {
		if err != nil {
			errorCount++
			assert.Contains(t, err.Error(), "operation cancelled")
		}
	}

	// All operations should have been cancelled
	assert.Equal(t, 30, errorCount, "All operations should fail with cancelled context")
}

// TestWorktreeStateConsistency tests worktree state consistency after cancellation
func TestWorktreeStateConsistency(t *testing.T) {
	git, tmpDir := setupTestGitRepo(t)

	// Create a context that we'll cancel during AddWorktree
	ctx, cancel := context.WithCancel(context.Background())

	// Path for the new worktree
	worktreePath := filepath.Join(tmpDir, ".worktrees", "consistency-test")

	// Start AddWorktree in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- git.AddWorktree(ctx, worktreePath, "consistency-branch")
	}()

	// Cancel quickly
	time.Sleep(1 * time.Millisecond)
	cancel()

	// Get the result
	err := <-errChan

	// If operation was cancelled, verify state
	if err != nil && strings.Contains(err.Error(), "operation cancelled") {
		// Verify worktree list is consistent
		worktrees, listErr := git.ListWorktrees(context.Background())
		assert.NoError(t, listErr)

		// The cancelled worktree should not be in the list
		for _, wt := range worktrees {
			assert.NotEqual(t, "consistency-branch", wt.Branch)
		}
	}
}

// TestWorktreeContextInheritance tests context inheritance in worktree operations
func TestWorktreeContextInheritance(t *testing.T) {
	git, _ := setupTestGitRepo(t)

	// Create parent context with timeout
	parentCtx, parentCancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer parentCancel()

	// Create child context
	childCtx, childCancel := context.WithCancel(parentCtx)
	defer childCancel()

	// Wait for parent timeout
	time.Sleep(100 * time.Millisecond)

	// Child context should also be cancelled
	_, err := git.ListWorktrees(childCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation cancelled")
}
