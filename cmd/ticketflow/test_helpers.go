package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/config"
)

// Test constants
const (
	testTicketID      = "250101-120000-test-feature"
	testDefaultBranch = "main"
	testTicketSlug    = "test-feature"
	testTimeout       = 5 // seconds for test operations

	// Test ticket IDs
	testTicket1ID = "250101-120000-test-1"
	testTicket2ID = "250102-120000-test-2"
	testTicket3ID = "250103-120000-test-3"
	testTicket4ID = "250104-120000-test-4"

	// Test timestamps
	testCreatedTime = "2025-01-01T12:00:00Z"
	testStartedTime = "2025-01-01T13:00:00Z"
	testClosedTime  = "2025-01-01T14:00:00Z"
)

// setupTestRepo creates a basic test repository with config and directories
// IMPORTANT: This function configures git locally within the test directory only.
// Never use --global flag in tests as it modifies the user's git configuration.
func setupTestRepo(t *testing.T, tmpDir string) {
	// Initialize git repo first
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to init git repo: %s", string(output))

	// Configure git locally in the test repo (not globally)
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	err = cmd.Run()
	require.NoError(t, err)
	
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	err = cmd.Run()
	require.NoError(t, err)
	
	cmd = exec.Command("git", "config", "init.defaultBranch", "main")
	cmd.Dir = tmpDir
	err = cmd.Run()
	require.NoError(t, err)

	// Ensure we're on the main branch
	// First check what branch we're on
	cmd = exec.Command("git", "branch", "--show-current")
	cmd.Dir = tmpDir
	currentBranch, err := cmd.Output()
	require.NoError(t, err)
	
	// If we're not on main, create and switch to main branch
	if strings.TrimSpace(string(currentBranch)) != "main" {
		cmd = exec.Command("git", "checkout", "-b", "main")
		cmd.Dir = tmpDir
		output, err = cmd.CombinedOutput()
		require.NoError(t, err, "Failed to create main branch: %s", string(output))
	}

	// Create config file
	cfg := config.Default()
	cfg.Git.DefaultBranch = "main"
	cfg.Worktree.Enabled = false

	// Create config YAML content
	configContent := `git:
  default_branch: main
worktree:
  enabled: false
  base_dir: ../ticketflow.worktrees
tickets:
  dir: tickets
  todo_dir: todo
  doing_dir: doing
  done_dir: done
output:
  default_format: text
`

	err = os.WriteFile(filepath.Join(tmpDir, ".ticketflow.yaml"), []byte(configContent), 0644)
	require.NoError(t, err)

	// Create directories
	for _, dir := range []string{"tickets/todo", "tickets/doing", "tickets/done"} {
		err = os.MkdirAll(filepath.Join(tmpDir, dir), 0755)
		require.NoError(t, err)
	}

	// Create empty .current file
	err = os.WriteFile(filepath.Join(tmpDir, "tickets", ".current"), []byte(""), 0644)
	require.NoError(t, err)

	// Create initial commit
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	err = cmd.Run()
	require.NoError(t, err)

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	err = cmd.Run()
	require.NoError(t, err)
}

// setupTestRepoWithTickets creates a test repo with sample tickets
func setupTestRepoWithTickets(t *testing.T, tmpDir string) {
	setupTestRepo(t, tmpDir)

	// Create some test tickets
	tickets := []struct {
		id     string
		status string
	}{
		{testTicket1ID, "todo"},
		{testTicket2ID, "todo"},
		{testTicket3ID, "doing"},
		{testTicket4ID, "done"},
	}

	for _, tc := range tickets {
		var content string
		switch tc.status {
		case "todo":
			content = fmt.Sprintf(`---
priority: 2
created_at: %s
---

# Test Ticket %s

This is a test ticket.
`, testCreatedTime, tc.id)
		case "doing":
			content = fmt.Sprintf(`---
priority: 2
created_at: %s
started_at: %s
---

# Test Ticket %s

This is a test ticket in progress.
`, testCreatedTime, testStartedTime, tc.id)
		case "done":
			content = fmt.Sprintf(`---
priority: 2
created_at: %s
started_at: %s
closed_at: %s
---

# Test Ticket %s

This is a completed test ticket.
`, testCreatedTime, testStartedTime, testClosedTime, tc.id)
		}

		path := filepath.Join(tmpDir, "tickets", tc.status, tc.id+".md")
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Commit all tickets
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	err := cmd.Run()
	require.NoError(t, err)

	cmd = exec.Command("git", "commit", "-m", "Add test tickets")
	cmd.Dir = tmpDir
	err = cmd.Run()
	require.NoError(t, err)
}

// setupTestRepoWithTicket creates a test repo with a single ticket
func setupTestRepoWithTicket(t *testing.T, tmpDir string) string {
	setupTestRepo(t, tmpDir)

	ticketID := testTicketID
	content := fmt.Sprintf(`---
priority: 2
created_at: %s
---

# Test Feature

This is a test ticket for testing show command.
`, testCreatedTime)

	path := filepath.Join(tmpDir, "tickets", "todo", ticketID+".md")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	// Commit the ticket
	cmd := exec.Command("git", "add", path)
	cmd.Dir = tmpDir
	err = cmd.Run()
	require.NoError(t, err)

	cmd = exec.Command("git", "commit", "-m", "Add test ticket")
	cmd.Dir = tmpDir
	err = cmd.Run()
	require.NoError(t, err)

	return ticketID
}

// setupTestRepoWithStartedTicket creates a test repo with a ticket in doing status
func setupTestRepoWithStartedTicket(t *testing.T, tmpDir string) string {
	// Start fresh without calling setupTestRepoWithTicket to avoid conflicts
	setupTestRepo(t, tmpDir)

	ticketID := testTicketID

	// Create and checkout the feature branch
	cmd := exec.Command("git", "checkout", "-b", ticketID)
	cmd.Dir = tmpDir
	err := cmd.Run()
	require.NoError(t, err)

	// Create ticket directly in doing status
	content := fmt.Sprintf(`---
priority: 2
created_at: %s
started_at: %s
---

# Test Feature

This is a test ticket that has been started.
`, testCreatedTime, testStartedTime)

	path := filepath.Join(tmpDir, "tickets", "doing", ticketID+".md")
	err = os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	// Set as current ticket - create symlink
	linkPath := filepath.Join(tmpDir, "current-ticket.md")
	targetPath := filepath.Join("tickets", "doing", ticketID+".md")
	err = os.Symlink(targetPath, linkPath)
	require.NoError(t, err)

	// Commit the changes
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	err = cmd.Run()
	require.NoError(t, err)

	cmd = exec.Command("git", "commit", "-m", "Add started ticket")
	cmd.Dir = tmpDir
	err = cmd.Run()
	require.NoError(t, err)

	return ticketID
}
