package testutil

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// GitRepo represents a test git repository
type GitRepo struct {
	Dir        string
	LastStderr string // Capture stderr for debugging
}

// execCommand executes a git command and captures both stdout and stderr
func (r *GitRepo) execCommand(t *testing.T, args ...string) error {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = r.Dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	r.LastStderr = stderr.String()
	if err != nil {
		return fmt.Errorf("git %v failed: %w\nstderr: %s", args, err, r.LastStderr)
	}
	return nil
}

// SetupGitRepo creates and initializes a git repository for testing
// IMPORTANT: This configures git locally in the test directory, never globally
func SetupGitRepo(t *testing.T, dir string) *GitRepo {
	t.Helper()

	return SetupGitRepoWithOptions(t, dir, DefaultGitOptions())
}

// ConfigureGitLocally configures git user in the specified directory
// WARNING: NEVER use git config --global in tests - it modifies the user's git config!
func ConfigureGitLocally(t *testing.T, dir, name, email string) {
	t.Helper()

	// Configure user name locally
	cmd := exec.Command("git", "config", "user.name", name)
	cmd.Dir = dir // Critical: sets the working directory for local config
	err := cmd.Run()
	require.NoError(t, err, "Failed to configure git user.name")

	// Configure user email locally
	cmd = exec.Command("git", "config", "user.email", email)
	cmd.Dir = dir // Critical: sets the working directory for local config
	err = cmd.Run()
	require.NoError(t, err, "Failed to configure git user.email")
}

// AddCommit adds a file and commits it
func (r *GitRepo) AddCommit(t *testing.T, filename, content, message string) {
	t.Helper()

	// If filename is ".", we're adding all files
	if filename != "." {
		// Write file
		filePath := filepath.Join(r.Dir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err, "Failed to write file")
	}

	// Add file to git
	err := r.execCommand(t, "add", filename)
	require.NoError(t, err, "Failed to add file to git")

	// Commit
	err = r.execCommand(t, "commit", "-m", message)
	require.NoError(t, err, "Failed to commit")
}

// CreateBranch creates and checks out a new branch
func (r *GitRepo) CreateBranch(t *testing.T, branchName string) {
	t.Helper()
	err := r.execCommand(t, "checkout", "-b", branchName)
	require.NoError(t, err, "Failed to create branch")
}

// CheckoutBranch checks out an existing branch
func (r *GitRepo) CheckoutBranch(t *testing.T, branchName string) {
	t.Helper()
	err := r.execCommand(t, "checkout", branchName)
	require.NoError(t, err, "Failed to checkout branch")
}

// CurrentBranch returns the current branch name
func (r *GitRepo) CurrentBranch(t *testing.T) string {
	t.Helper()

	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = r.Dir
	output, err := cmd.Output()
	require.NoError(t, err, "Failed to get current branch")

	return strings.TrimSpace(string(output))
}

// AddRemote adds a remote repository
func (r *GitRepo) AddRemote(t *testing.T, name, url string) {
	t.Helper()
	err := r.execCommand(t, "remote", "add", name, url)
	require.NoError(t, err, "Failed to add remote")
}

// Tag creates a git tag
func (r *GitRepo) Tag(t *testing.T, tagName string) {
	t.Helper()
	err := r.execCommand(t, "tag", tagName)
	require.NoError(t, err, "Failed to create tag")
}

// GitOptions for customizing git setup
type GitOptions struct {
	UserName          string
	UserEmail         string
	DefaultBranch     string
	InitialCommit     bool
	AddRemote         bool
	RemoteName        string
	RemoteURL         string
	DisableSigning    bool
	InitDefaultBranch bool
}

// DefaultGitOptions returns default options for git setup
func DefaultGitOptions() GitOptions {
	return GitOptions{
		UserName:          "Test User",
		UserEmail:         "test@example.com",
		DefaultBranch:     "main",
		InitialCommit:     true,
		AddRemote:         false,
		DisableSigning:    true,
		InitDefaultBranch: true,
	}
}

// execGitCommand is a standalone helper for executing git commands with stderr capture
// This is used by SetupGitRepoWithOptions and other functions that don't have access to GitRepo methods
func execGitCommand(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stderr.String(), fmt.Errorf("git %v failed: %w\nstderr: %s", args, err, stderr.String())
	}
	return stderr.String(), nil
}

// SetupGitRepoWithOptions creates a git repository with custom options
func SetupGitRepoWithOptions(t *testing.T, dir string, opts GitOptions) *GitRepo {
	t.Helper()

	repo := &GitRepo{Dir: dir}

	// Initialize git repository
	args := []string{"init"}
	if opts.DefaultBranch != "" {
		args = append(args, "-b", opts.DefaultBranch)
	}
	stderr, err := execGitCommand(t, dir, args...)
	repo.LastStderr = stderr
	if err != nil {
		// Retry without -b for older git versions
		stderr, err = execGitCommand(t, dir, "init")
		repo.LastStderr = stderr
		require.NoError(t, err, "Failed to initialize git repository")
		if opts.DefaultBranch != "" {
			stderr, err = execGitCommand(t, dir, "symbolic-ref", "HEAD", "refs/heads/"+opts.DefaultBranch)
			repo.LastStderr = stderr
			require.NoError(t, err, "Failed to configure HEAD for default branch")
		}
	}

	// Configure git locally
	ConfigureGitLocally(t, dir, opts.UserName, opts.UserEmail)
	if opts.DisableSigning {
		stderr, err = execGitCommand(t, dir, "config", "commit.gpgSign", "false")
		repo.LastStderr = stderr
		require.NoError(t, err, "Failed to disable commit signing")
	}
	if opts.InitDefaultBranch && opts.DefaultBranch != "" {
		stderr, err = execGitCommand(t, dir, "config", "init.defaultBranch", opts.DefaultBranch)
		repo.LastStderr = stderr
		require.NoError(t, err, "Failed to set init.defaultBranch")
	}

	// Create initial commit if requested
	if opts.InitialCommit {
		stderr, err = execGitCommand(t, dir, "commit", "--allow-empty", "-m", "Initial commit")
		repo.LastStderr = stderr
		require.NoError(t, err, "Failed to create initial commit")
	}

	// Ensure the repository is on the requested branch when an initial commit exists
	if opts.DefaultBranch != "" && opts.InitialCommit {
		stderr, err = execGitCommand(t, dir, "checkout", opts.DefaultBranch)
		repo.LastStderr = stderr
		require.NoError(t, err, "Failed to checkout default branch")
	}

	// Add remote if requested
	if opts.AddRemote && opts.RemoteName != "" && opts.RemoteURL != "" {
		stderr, err = execGitCommand(t, dir, "remote", "add", opts.RemoteName, opts.RemoteURL)
		repo.LastStderr = stderr
		require.NoError(t, err, "Failed to add remote")
	}

	return repo
}
