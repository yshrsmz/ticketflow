package commands

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/cli/commands/testharness"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// Ensures starting a ticket from within an existing worktree creates the new
// worktree at the main repository's configured base directory (not nested).
func TestStartCommand_FromWithinWorktree_CreatesAtMainBase(t *testing.T) {
	// Setup test environment
	env := testharness.NewTestEnvironment(t)

	// Create two tickets and commit
	env.CreateTicket("first", ticket.StatusTodo)
	env.RunGit("add", ".")
	env.RunGit("commit", "-m", "Add first ticket")

	env.CreateTicket("second", ticket.StatusTodo)
	env.RunGit("add", ".")
	env.RunGit("commit", "-m", "Add second ticket")

	// Create a worktree for the first ticket
	env.CreateWorktree("first")

	// Change working directory to the first ticket's worktree
	worktreePath := filepath.Join(filepath.Dir(env.RootDir), "test-worktrees", "first")
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldWd) }()
	require.NoError(t, os.Chdir(worktreePath))

	// Start the second ticket from within the worktree
	cmd := NewStartCommand()
	flags := &startFlags{format: StringFlag{Long: string(FormatText)}}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = cmd.Execute(ctx, flags, []string{"second"})
	require.NoError(t, err)

	// Verify the new worktree is created at the expected base (main root ../test-worktrees)
	expectedPath := filepath.Join(filepath.Dir(env.RootDir), "test-worktrees", "second")

	// Parse `git worktree list --porcelain` to find the exact path
	output := env.RunGit("worktree", "list", "--porcelain")
	lines := strings.Split(strings.TrimSpace(output), "\n")
	found := false
	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			path := strings.TrimPrefix(line, "worktree ")
			// Normalize symlinks for comparison on macOS
			got, err := filepath.EvalSymlinks(path)
			require.NoError(t, err)
			want, err := filepath.EvalSymlinks(expectedPath)
			require.NoError(t, err)
			if got == want {
				found = true
				break
			}
		}
	}

	assert.True(t, found, "expected worktree at %s not found in list", expectedPath)
}
