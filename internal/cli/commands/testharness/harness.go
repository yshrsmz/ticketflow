// Package testharness provides integration testing utilities for CLI commands
package testharness

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/ticket"
	"gopkg.in/yaml.v3"
)

// TestEnvironment provides a complete test environment for CLI integration tests
type TestEnvironment struct {
	t          *testing.T
	RootDir    string
	ConfigPath string
	Config     *config.Config
	ctx        context.Context
}

// NewTestEnvironment creates a new test environment with git repo and ticketflow setup
func NewTestEnvironment(t *testing.T) *TestEnvironment {
	t.Helper()

	tmpDir := t.TempDir()
	env := &TestEnvironment{
		t:          t,
		RootDir:    tmpDir,
		ConfigPath: filepath.Join(tmpDir, ".ticketflow.yaml"),
		ctx:        context.Background(),
	}

	// Initialize git repo with explicit main branch
	// Use -b flag to set initial branch name (git 2.28+) or fall back to renaming
	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		// Fallback for older git versions: init then rename branch
		env.RunGit("init")
		// Create initial commit first (can't rename empty branch)
		env.WriteFile("README.md", "# Test Repository")
		env.RunGit("add", "README.md")
		env.RunGit("config", "user.name", "Test User")
		env.RunGit("config", "user.email", "test@example.com")
		env.RunGit("commit", "-m", "Initial commit")
		// Now rename the branch to main
		env.RunGit("branch", "-M", "main")
	} else {
		// Successfully created with main branch, now configure
		env.RunGit("config", "user.name", "Test User")
		env.RunGit("config", "user.email", "test@example.com")
		// Create initial commit to have a valid HEAD
		env.WriteFile("README.md", "# Test Repository")
		env.RunGit("add", "README.md")
		env.RunGit("commit", "-m", "Initial commit")
	}

	// Create ticket directories
	ticketsDir := filepath.Join(tmpDir, "tickets")
	require.NoError(t, os.MkdirAll(filepath.Join(ticketsDir, "todo"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(ticketsDir, "doing"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(ticketsDir, "done"), 0755))

	// Create default config
	cfg := config.Default()
	cfg.Git.DefaultBranch = "main"
	cfg.Worktree.Enabled = true
	cfg.Worktree.BaseDir = "../test-worktrees"
	cfg.Tickets.Dir = "tickets"
	cfg.Tickets.TodoDir = "todo"
	cfg.Tickets.DoingDir = "doing"
	cfg.Tickets.DoneDir = "done"
	env.Config = cfg

	// Write config file
	data, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(env.ConfigPath, data, 0644))

	// Create worktree base directory
	worktreeDir := filepath.Join(filepath.Dir(tmpDir), "test-worktrees")
	require.NoError(t, os.MkdirAll(worktreeDir, 0755))

	return env
}

// Context returns the test context
func (e *TestEnvironment) Context() context.Context {
	return e.ctx
}

// RunGit runs a git command in the test repository
func (e *TestEnvironment) RunGit(args ...string) string {
	e.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = e.RootDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		e.t.Fatalf("git %s failed: %v\nOutput: %s", strings.Join(args, " "), err, output)
	}
	return string(output)
}

// WriteFile writes a file in the test repository
func (e *TestEnvironment) WriteFile(path, content string) {
	e.t.Helper()
	fullPath := filepath.Join(e.RootDir, path)

	// Ensure the resolved path is within RootDir to prevent directory traversal
	cleanPath, err := filepath.Abs(fullPath)
	require.NoError(e.t, err)
	if !strings.HasPrefix(cleanPath, e.RootDir) {
		e.t.Fatalf("path %q escapes test directory", path)
	}

	dir := filepath.Dir(fullPath)
	require.NoError(e.t, os.MkdirAll(dir, 0755))
	require.NoError(e.t, os.WriteFile(fullPath, []byte(content), 0644))
}

// ReadFile reads a file from the test repository
func (e *TestEnvironment) ReadFile(path string) string {
	e.t.Helper()
	fullPath := filepath.Join(e.RootDir, path)
	data, err := os.ReadFile(fullPath)
	require.NoError(e.t, err)
	return string(data)
}

// FileExists checks if a file exists in the test repository
func (e *TestEnvironment) FileExists(path string) bool {
	fullPath := filepath.Join(e.RootDir, path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// CreateTicket creates a test ticket in the specified status
func (e *TestEnvironment) CreateTicket(id string, status ticket.Status, options ...TicketOption) *ticket.Ticket {
	e.t.Helper()

	now := time.Now()
	t := &ticket.Ticket{
		ID:          id,
		Priority:    1,
		Description: "Test ticket",
		CreatedAt:   ticket.RFC3339Time{Time: now},
	}

	// Apply options
	for _, opt := range options {
		opt(t)
	}

	// Set status-specific fields
	switch status {
	case ticket.StatusDoing:
		t.StartedAt = ticket.RFC3339TimePtr{Time: &now}
	case ticket.StatusDone:
		startTime := now.Add(-time.Hour)
		t.StartedAt = ticket.RFC3339TimePtr{Time: &startTime}
		t.ClosedAt = ticket.RFC3339TimePtr{Time: &now}
	}

	// Marshal ticket to YAML with frontmatter
	var content strings.Builder
	content.WriteString("---\n")

	frontmatter := map[string]interface{}{
		"priority":    t.Priority,
		"description": t.Description,
		"created_at":  t.CreatedAt.Format(time.RFC3339),
	}

	if t.StartedAt.Time != nil {
		frontmatter["started_at"] = t.StartedAt.Time.Format(time.RFC3339)
	}
	if t.ClosedAt.Time != nil {
		frontmatter["closed_at"] = t.ClosedAt.Time.Format(time.RFC3339)
	}
	if len(t.Related) > 0 {
		frontmatter["related"] = t.Related
	}

	data, err := yaml.Marshal(frontmatter)
	require.NoError(e.t, err)
	content.WriteString(string(data))
	content.WriteString("---\n\n")
	content.WriteString(fmt.Sprintf("# %s\n\n%s\n", t.ID, t.Content))

	// Write ticket file
	ticketPath := e.TicketPath(string(status), id+".md")
	e.WriteFile(ticketPath, content.String())

	// If in doing status, create symlink
	if status == ticket.StatusDoing {
		symlinkPath := filepath.Join(e.RootDir, "current-ticket.md")
		// Properly handle symlink removal with error checking
		if err := os.Remove(symlinkPath); err != nil && !os.IsNotExist(err) {
			e.t.Fatalf("failed to remove existing symlink: %v", err)
		}
		target := ticketPath
		require.NoError(e.t, os.Symlink(target, symlinkPath))
	}

	return t
}

// TicketOption allows customizing ticket creation
type TicketOption func(*ticket.Ticket)

// WithParent sets a parent ticket relationship
func WithParent(parentID string) TicketOption {
	return func(t *ticket.Ticket) {
		t.Related = append(t.Related, fmt.Sprintf("parent:%s", parentID))
	}
}

// WithContent sets the ticket content
func WithContent(content string) TicketOption {
	return func(t *ticket.Ticket) {
		t.Content = content
	}
}

// WithDescription sets the ticket description
func WithDescription(description string) TicketOption {
	return func(t *ticket.Ticket) {
		t.Description = description
	}
}

// TicketPath returns the path to a ticket file
func (e *TestEnvironment) TicketPath(status, filename string) string {
	return filepath.Join("tickets", status, filename)
}

// CreateWorktree creates a git worktree for a ticket
func (e *TestEnvironment) CreateWorktree(ticketID string) {
	e.t.Helper()

	// Check if branch already exists
	branches := e.RunGit("branch", "--list", ticketID)
	if !strings.Contains(branches, ticketID) {
		// Create branch only if it doesn't exist
		// First ensure we're on main branch
		currentBranch := e.GetCurrentBranch()
		if currentBranch != "main" {
			e.RunGit("checkout", "main")
		}
		e.RunGit("checkout", "-b", ticketID)
		e.RunGit("checkout", "main")
	}

	// Create worktree
	worktreePath := filepath.Join(filepath.Dir(e.RootDir), "test-worktrees", ticketID)
	e.RunGit("worktree", "add", worktreePath, ticketID)
}

// WorktreeExists checks if a worktree exists for a ticket
func (e *TestEnvironment) WorktreeExists(ticketID string) bool {
	output := e.RunGit("worktree", "list")
	worktreePath := filepath.Join("test-worktrees", ticketID)
	return strings.Contains(output, worktreePath)
}

// GetCurrentBranch returns the current git branch
func (e *TestEnvironment) GetCurrentBranch() string {
	output := e.RunGit("branch", "--show-current")
	return strings.TrimSpace(output)
}

// LastCommitMessage returns the last commit message
func (e *TestEnvironment) LastCommitMessage() string {
	output := e.RunGit("log", "-1", "--pretty=%B")
	return strings.TrimSpace(output)
}

// HasUncommittedChanges checks if there are uncommitted changes
func (e *TestEnvironment) HasUncommittedChanges() bool {
	output := e.RunGit("status", "--porcelain")
	return len(strings.TrimSpace(output)) > 0
}

// WithWorkingDirectory changes to the test root directory and restores the original
// working directory when the function completes. This is useful for tests that need
// to run commands that expect to be in the project root.
func (e *TestEnvironment) WithWorkingDirectory(t *testing.T, fn func()) {
	t.Helper()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(e.RootDir))
	fn()
}

// Cleanup performs any necessary cleanup
func (e *TestEnvironment) Cleanup() {
	// Cleanup is handled by t.TempDir()
}

// CaptureOutput captures stdout during function execution and returns it as a string
func CaptureOutput(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Use a channel to safely read output
	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	// Execute the function
	fn()

	// Restore stdout and close writer
	_ = w.Close() // Ignore close error as we've already captured the output
	os.Stdout = old

	// Wait for reader to finish
	return <-done
}
