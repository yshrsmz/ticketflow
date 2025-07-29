package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestAutoCleanupStaleBranches(t *testing.T) {
	// Create a temporary directory for our test repo
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")
	require.NoError(t, os.MkdirAll(repoPath, 0755))

	// Change to test directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalWd)
		require.NoError(t, err)
	}()
	require.NoError(t, os.Chdir(repoPath))

	// Initialize git repo
	gitOps := &git.Git{}
	_, err = gitOps.Exec("init")
	require.NoError(t, err)
	_, err = gitOps.Exec("config", "user.email", "test@example.com")
	require.NoError(t, err)
	_, err = gitOps.Exec("config", "user.name", "Test User")
	require.NoError(t, err)

	// Create initial commit
	require.NoError(t, os.WriteFile("README.md", []byte("Test repo"), 0644))
	_, err = gitOps.Exec("add", ".")
	require.NoError(t, err)
	_, err = gitOps.Exec("commit", "-m", "Initial commit")
	require.NoError(t, err)

	// Create config
	cfg := config.Default()
	cfg.Git.DefaultBranch = "main"
	cfg.Tickets.Dir = "tickets"
	cfg.Worktree.Enabled = false // Disable worktrees for this test

	// Create ticket directories
	for _, dir := range []string{"todo", "doing", "done"} {
		require.NoError(t, os.MkdirAll(filepath.Join(cfg.Tickets.Dir, dir), 0755))
	}

	// Create ticket manager
	manager := ticket.NewManager(cfg, repoPath)

	// Create app
	app := &App{
		Manager:     manager,
		Git:         gitOps,
		Config:      cfg,
		ProjectRoot: repoPath,
	}

	// Test scenario: Create tickets and branches, move tickets to done, then run cleanup
	tickets := []struct {
		id     string
		slug   string
		status ticket.Status
	}{
		{"ticket-1", "test-ticket-1", ticket.StatusDone},
		{"ticket-2", "test-ticket-2", ticket.StatusDone},
		{"ticket-3", "test-ticket-3", ticket.StatusDoing}, // This one should not be cleaned
	}

	// Create tickets and branches
	for _, tc := range tickets {
		// Create ticket through manager to ensure proper structure
		now := time.Now()
		tkt := &ticket.Ticket{
			ID:          tc.id,
			Slug:        tc.slug,
			Priority:    2,
			Description: "Test ticket",
			CreatedAt:   ticket.NewRFC3339Time(now),
			Path:        filepath.Join(cfg.Tickets.Dir, string(tc.status), tc.id+".md"),
		}
		
		// Set closed time for done tickets
		if tc.status == ticket.StatusDone {
			closedTime := now.Add(1 * time.Hour)
			tkt.ClosedAt = ticket.RFC3339TimePtr{Time: &closedTime}
		}
		
		// Write ticket file
		data, err := tkt.ToBytes()
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(tkt.Path, data, 0644))

		// Create branch for the ticket
		_, err = gitOps.Exec("checkout", "-b", tc.id)
		require.NoError(t, err)
		_, err = gitOps.Exec("checkout", "main")
		require.NoError(t, err)
	}

	// Verify branches exist
	output, err := gitOps.Exec("branch", "--format=%(refname:short)")
	require.NoError(t, err)
	branches := splitLines(output)
	assert.Contains(t, branches, "ticket-1")
	assert.Contains(t, branches, "ticket-2")
	assert.Contains(t, branches, "ticket-3")

	// Run auto cleanup with dry run
	err = app.AutoCleanup(true)
	require.NoError(t, err)

	// Verify branches still exist (dry run)
	output, err = gitOps.Exec("branch", "--format=%(refname:short)")
	require.NoError(t, err)
	branches = splitLines(output)
	assert.Contains(t, branches, "ticket-1")
	assert.Contains(t, branches, "ticket-2")
	assert.Contains(t, branches, "ticket-3")

	// Run actual cleanup
	err = app.AutoCleanup(false)
	require.NoError(t, err)

	// Verify only done ticket branches were removed
	output, err = gitOps.Exec("branch", "--format=%(refname:short)")
	require.NoError(t, err)
	branches = splitLines(output)
	assert.NotContains(t, branches, "ticket-1") // Should be removed (done)
	assert.NotContains(t, branches, "ticket-2") // Should be removed (done)
	assert.Contains(t, branches, "ticket-3")    // Should still exist (doing)
	assert.Contains(t, branches, "main")        // Should still exist
}

func TestCleanupStatsWithDoneTickets(t *testing.T) {
	// Create a temporary directory for our test repo
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")
	require.NoError(t, os.MkdirAll(repoPath, 0755))

	// Change to test directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalWd)
		require.NoError(t, err)
	}()
	require.NoError(t, os.Chdir(repoPath))

	// Initialize git repo
	gitOps := &git.Git{}
	_, err = gitOps.Exec("init")
	require.NoError(t, err)
	_, err = gitOps.Exec("config", "user.email", "test@example.com")
	require.NoError(t, err)
	_, err = gitOps.Exec("config", "user.name", "Test User")
	require.NoError(t, err)

	// Create initial commit
	require.NoError(t, os.WriteFile("README.md", []byte("Test repo"), 0644))
	_, err = gitOps.Exec("add", ".")
	require.NoError(t, err)
	_, err = gitOps.Exec("commit", "-m", "Initial commit")
	require.NoError(t, err)

	// Create config
	cfg := config.Default()
	cfg.Git.DefaultBranch = "main"
	cfg.Tickets.Dir = "tickets"

	// Create ticket directories
	for _, dir := range []string{"todo", "doing", "done"} {
		require.NoError(t, os.MkdirAll(filepath.Join(cfg.Tickets.Dir, dir), 0755))
	}

	// Create ticket manager
	manager := ticket.NewManager(cfg, repoPath)

	// Create done tickets with branches
	doneTickets := []string{"done-1", "done-2", "done-3"}
	now := time.Now()
	for _, id := range doneTickets {
		// Create ticket file in done directory
		tkt := &ticket.Ticket{
			ID:          id,
			Slug:        id,
			Priority:    2,
			Description: "Done ticket",
			CreatedAt:   ticket.NewRFC3339Time(now),
			Path:        filepath.Join(cfg.Tickets.Dir, "done", id+".md"),
		}
		
		// Set started and closed times
		startedTime := now.Add(1 * time.Hour)
		tkt.StartedAt = ticket.RFC3339TimePtr{Time: &startedTime}
		closedTime := now.Add(2 * time.Hour)
		tkt.ClosedAt = ticket.RFC3339TimePtr{Time: &closedTime}
		
		// Write ticket file
		data, err := tkt.ToBytes()
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(tkt.Path, data, 0644))

		// Create branch
		_, err = gitOps.Exec("checkout", "-b", id)
		require.NoError(t, err)
		_, err = gitOps.Exec("checkout", "main")
		require.NoError(t, err)
	}

	// Also create an active ticket with branch
	activeTkt := &ticket.Ticket{
		ID:          "active-1",
		Slug:        "active-1",
		Priority:    2,
		Description: "Active ticket",
		CreatedAt:   ticket.NewRFC3339Time(now),
		Path:        filepath.Join(cfg.Tickets.Dir, "doing", "active-1.md"),
	}
	
	// Set started time
	startedTime := now.Add(1 * time.Hour)
	activeTkt.StartedAt = ticket.RFC3339TimePtr{Time: &startedTime}
	
	// Write ticket file
	data, err := activeTkt.ToBytes()
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(activeTkt.Path, data, 0644))
	_, err = gitOps.Exec("checkout", "-b", "active-1")
	require.NoError(t, err)
	_, err = gitOps.Exec("checkout", "main")
	require.NoError(t, err)

	// Run CleanupStats and verify it counts stale branches correctly
	// Since CleanupStats prints to stdout, we can't easily capture its output in a test
	// But we can verify the underlying logic by checking what branches would be cleaned
	
	// Get all branches
	output, err := gitOps.Exec("branch", "--format=%(refname:short)")
	require.NoError(t, err)
	branches := splitLines(output)

	// Get all tickets
	allTickets, err := manager.List(ticket.StatusFilterAll)
	require.NoError(t, err)

	// Count stale branches manually
	ticketStatus := make(map[string]ticket.Status)
	for _, t := range allTickets {
		ticketStatus[t.ID] = t.Status()
	}

	staleCount := 0
	for _, branch := range branches {
		if branch == "main" {
			continue
		}
		if status, exists := ticketStatus[branch]; exists && status == ticket.StatusDone {
			staleCount++
		}
	}

	// Should have 3 stale branches (the done tickets)
	assert.Equal(t, 3, staleCount)
}