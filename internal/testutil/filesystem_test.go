package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yshrsmz/ticketflow/internal/config"
)

func TestSetupTicketflowProject_Defaults(t *testing.T) {
	dir := t.TempDir()

	repo := SetupTicketflowProject(t, dir)
	if repo == nil {
		t.Fatalf("expected repo to be initialized")
	}

	// Ensure ticket directories exist
	for _, sub := range []string{"todo", "doing", "done"} {
		path := filepath.Join(dir, "tickets", sub)
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			t.Fatalf("expected directory %s to exist", path)
		}
	}

	// Ensure tickets/.current marker exists
	marker := filepath.Join(dir, "tickets", ".current")
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("expected marker %s to exist: %v", marker, err)
	}

	// Ensure config matches defaults and uses canonical dirs
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg.Tickets.TodoDir != "todo" {
		t.Fatalf("expected todo dir 'todo', got %q", cfg.Tickets.TodoDir)
	}
	if cfg.Tickets.Dir != "tickets" {
		t.Fatalf("expected tickets dir 'tickets', got %q", cfg.Tickets.Dir)
	}
}

func TestSetupTicketflowProject_WithCustomConfig(t *testing.T) {
	dir := t.TempDir()

	custom := config.Default()
	custom.Worktree.Enabled = false
	custom.Git.DefaultBranch = "develop"

	repo := SetupTicketflowProject(t, dir, WithCustomConfig(custom))
	if repo == nil {
		t.Fatalf("expected repo to be initialized")
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Worktree.Enabled {
		t.Fatalf("expected worktrees disabled")
	}
	if cfg.Git.DefaultBranch != "develop" {
		t.Fatalf("expected default branch 'develop', got %q", cfg.Git.DefaultBranch)
	}
}

func TestSetupTicketflowRepo_OverridesWithoutGit(t *testing.T) {
	dir := t.TempDir()

	repo := SetupTicketflowRepo(t, dir, WithoutGit())
	if repo == nil {
		t.Fatalf("expected repo to be initialized even when WithoutGit is provided")
	}
}

func TestSetupTicketflowProject_WithoutGit(t *testing.T) {
	dir := t.TempDir()

	repo := SetupTicketflowProject(t, dir, WithoutGit())
	if repo != nil {
		t.Fatalf("expected repo to be nil when git init disabled")
	}
}
