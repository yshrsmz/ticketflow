package testutil

import (
	"os/exec"
	"strings"
	"testing"
)

func TestSetupGitRepo_UsesMainAndDisablesSigning(t *testing.T) {
	dir := t.TempDir()

	repo := SetupGitRepo(t, dir)
	if repo == nil {
		t.Fatal("expected repo")
	}

	// Ensure current branch is main
	branchCmd := exec.Command("git", "branch", "--show-current")
	branchCmd.Dir = dir
	branch, err := branchCmd.Output()
	if err != nil {
		t.Fatalf("failed to get current branch: %v", err)
	}
	if got := strings.TrimSpace(string(branch)); got != "main" {
		t.Fatalf("expected branch main, got %q", got)
	}

	// Ensure commit signing disabled
	signCmd := exec.Command("git", "config", "--get", "commit.gpgSign")
	signCmd.Dir = dir
	signing, err := signCmd.Output()
	if err != nil {
		t.Fatalf("failed to read commit signing config: %v", err)
	}
	if got := strings.TrimSpace(string(signing)); got != "false" {
		t.Fatalf("expected commit signing disabled, got %q", got)
	}

	// Ensure init.defaultBranch is set so future repos use main
	initCmd := exec.Command("git", "config", "--get", "init.defaultBranch")
	initCmd.Dir = dir
	initBranch, err := initCmd.Output()
	if err != nil {
		t.Fatalf("failed to read init.defaultBranch: %v", err)
	}
	if got := strings.TrimSpace(string(initBranch)); got != "main" {
		t.Fatalf("expected init.defaultBranch main, got %q", got)
	}
}

func TestSetupGitRepoWithOptions_SkipInitialCommit(t *testing.T) {
	dir := t.TempDir()

	opts := DefaultGitOptions()
	opts.InitialCommit = false

	repo := SetupGitRepoWithOptions(t, dir, opts)
	if repo == nil {
		t.Fatal("expected repo")
	}

	// `git rev-parse --verify HEAD` should fail because no commit exists
	revParse := exec.Command("git", "rev-parse", "--verify", "HEAD")
	revParse.Dir = dir
	if err := revParse.Run(); err == nil {
		t.Fatalf("expected rev-parse to fail when no commits exist")
	}

	// Ensure HEAD still points to main so the first commit lands there
	symbolic := exec.Command("git", "symbolic-ref", "HEAD")
	symbolic.Dir = dir
	headRef, err := symbolic.Output()
	if err != nil {
		t.Fatalf("failed to read HEAD symbolic ref: %v", err)
	}
	if got := strings.TrimSpace(string(headRef)); got != "refs/heads/main" {
		t.Fatalf("expected HEAD to reference main, got %q", got)
	}
}
