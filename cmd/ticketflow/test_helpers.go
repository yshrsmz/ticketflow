package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/testutil"
)

// setupTestRepo creates a ticketflow test repository with config and ticket directories
func setupTestRepo(t *testing.T, tmpDir string) *testutil.GitRepo {
	cfg := config.Default()
	cfg.Git.DefaultBranch = testutil.TestDefaultBranch
	cfg.Worktree.Enabled = false

	repo := testutil.SetupTicketflowRepo(t, tmpDir, testutil.WithCustomConfig(cfg))
	require.NotNil(t, repo)

	repo.AddCommit(t, ".", "", "Bootstrap ticketflow project")
	return repo
}

// setupTestRepoWithTickets creates a test repo with sample tickets
func setupTestRepoWithTickets(t *testing.T, tmpDir string) {
	repo := setupTestRepo(t, tmpDir)

	// Create some test tickets
	tickets := []struct {
		id     string
		status string
	}{
		{testutil.TestTicket1ID, "todo"},
		{testutil.TestTicket2ID, "todo"},
		{testutil.TestTicket3ID, "doing"},
		{testutil.TestTicket4ID, "done"},
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
`, testutil.TestCreatedTime, tc.id)
		case "doing":
			content = fmt.Sprintf(`---
priority: 2
created_at: %s
started_at: %s
---

# Test Ticket %s

This is a test ticket in progress.
`, testutil.TestCreatedTime, testutil.TestStartedTime, tc.id)
		case "done":
			content = fmt.Sprintf(`---
priority: 2
created_at: %s
started_at: %s
closed_at: %s
---

# Test Ticket %s

This is a completed test ticket.
`, testutil.TestCreatedTime, testutil.TestStartedTime, testutil.TestClosedTime, tc.id)
		}

		path := filepath.Join(tmpDir, "tickets", tc.status, tc.id+".md")
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	}

	// Commit all tickets
	repo.AddCommit(t, ".", "", "Add test tickets")
}

// setupTestRepoWithTicket creates a test repo with a single ticket
func setupTestRepoWithTicket(t *testing.T, tmpDir string) string {
	repo := setupTestRepo(t, tmpDir)

	ticketID := testutil.TestTicketID
	content := fmt.Sprintf(`---
priority: 2
created_at: %s
---

# Test Feature

This is a test ticket for testing show command.
`, testutil.TestCreatedTime)

	// Commit the ticket with generated content
	repo.AddCommit(t, filepath.Join("tickets", "todo", ticketID+".md"), content, "Add test ticket")

	return ticketID
}

// setupTestRepoWithStartedTicket creates a test repo with a ticket in doing status
func setupTestRepoWithStartedTicket(t *testing.T, tmpDir string) string {
	repo := setupTestRepo(t, tmpDir)

	ticketID := testutil.TestTicketID

	// Create and checkout the feature branch
	repo.CreateBranch(t, ticketID)

	// Create ticket directly in doing status
	content := fmt.Sprintf(`---
priority: 2
created_at: %s
started_at: %s
---

# Test Feature

This is a test ticket that has been started.
`, testutil.TestCreatedTime, testutil.TestStartedTime)

	path := filepath.Join(tmpDir, "tickets", "doing", ticketID+".md")
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	// Set as current ticket - create symlink
	linkPath := filepath.Join(tmpDir, "current-ticket.md")
	targetPath := filepath.Join("tickets", "doing", ticketID+".md")
	require.NoError(t, os.Symlink(targetPath, linkPath))

	// Commit the changes
	repo.AddCommit(t, ".", "", "Add started ticket")

	return ticketID
}
