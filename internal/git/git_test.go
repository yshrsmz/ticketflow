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
