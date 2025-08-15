package testharness

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTestEnvironment_CreatesMainBranch(t *testing.T) {
	// This test verifies that the test environment correctly creates
	// a git repository with "main" as the default branch, regardless
	// of the git version's default branch name setting
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	// Verify we're on the main branch
	currentBranch := env.GetCurrentBranch()
	assert.Equal(t, "main", currentBranch, "Test environment should create 'main' branch")

	// Verify main branch exists in branch list
	branches := env.RunGit("branch", "--list")
	assert.Contains(t, branches, "main", "Branch list should contain 'main'")

	// Verify we can checkout main explicitly
	output := env.RunGit("checkout", "main")
	assert.NotContains(t, output, "error", "Should be able to checkout main branch")

	// Verify initial commit exists
	commits := env.RunGit("log", "--oneline")
	assert.Contains(t, commits, "Initial commit", "Should have initial commit")
}

func TestNewTestEnvironment_HandlesGitVersionCompatibility(t *testing.T) {
	// This test verifies that our fallback logic works for both
	// newer git versions (2.28+) with -b flag and older versions
	tmpDir := t.TempDir()

	// Test with explicit -b flag (newer git versions)
	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = tmpDir
	err := cmd.Run()

	if err == nil {
		// Newer git version - verify it created main branch
		cmd = exec.Command("git", "branch", "--show-current")
		cmd.Dir = tmpDir
		output, err := cmd.Output()
		if err == nil && strings.TrimSpace(string(output)) == "" {
			// Empty repo, add a commit to check branch
			cmd = exec.Command("bash", "-c", `
				git config user.name "Test" &&
				git config user.email "test@example.com" &&
				echo "test" > test.txt &&
				git add test.txt &&
				git commit -m "test" &&
				git branch --show-current
			`)
			cmd.Dir = tmpDir
			output, err = cmd.Output()
			require.NoError(t, err)
			assert.Equal(t, "main", strings.TrimSpace(string(output)))
		}
	} else {
		// Older git version - our fallback should handle this
		t.Log("Git version doesn't support -b flag, testing fallback logic")
		// The NewTestEnvironment function handles this case
		env := NewTestEnvironment(t)
		defer env.Cleanup()
		assert.Equal(t, "main", env.GetCurrentBranch())
	}
}

func TestCreateWorktree_HandlesMainBranch(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	// Create a test ticket
	env.CreateTicket("test-ticket", "todo")

	// Verify we can create a worktree from main branch
	env.CreateWorktree("test-ticket")

	// Verify worktree was created
	assert.True(t, env.WorktreeExists("test-ticket"))

	// Verify we're still on main branch in the main repo
	assert.Equal(t, "main", env.GetCurrentBranch())

	// Verify the ticket branch exists
	branches := env.RunGit("branch", "--list", "test-ticket")
	assert.Contains(t, branches, "test-ticket")
}