package testutil

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/testsupport/gitconfig"
)

// GitRepo represents a test git repository
type GitRepo struct {
	Dir        string
	LastStderr string // Capture stderr for debugging
}

// GitExecutor exposes the minimal Exec behaviour required to apply shared git configuration.
type GitExecutor = gitconfig.Executor

// gitCommandExecutor adapts git CLI invocations to the shared gitconfig.Executor interface.
type gitCommandExecutor struct {
	dir string
}

func (e gitCommandExecutor) Exec(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = e.dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stdout.String(), fmt.Errorf("git %v failed: %w\nstderr: %s", args, err, stderr.String())
	}
	return stdout.String(), nil
}

// execCommand executes a git command and captures both stdout and stderr.
// The stderr is stored in r.LastStderr for debugging purposes.
func (r *GitRepo) execCommand(t *testing.T, args ...string) error {
	t.Helper()
	return r.execCommandWithOutput(t, r.Dir, args...)
}

// execCommandWithOutput is a shared implementation for executing git commands.
// It captures stderr in r.LastStderr for debugging when errors occur.
func (r *GitRepo) execCommandWithOutput(t *testing.T, dir string, args ...string) error {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
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

	ConfigureGitClient(t, gitCommandExecutor{dir: dir}, GitOptions{
		UserName:          name,
		UserEmail:         email,
		DisableSigning:    false,
		DefaultBranch:     "",
		InitialCommit:     false,
		AddRemote:         false,
		InitDefaultBranch: false,
	})
}

// ConfigureGitClient applies canonical test git configuration via an Exec-capable client.
func ConfigureGitClient(t *testing.T, exec GitExecutor, opts ...GitOptions) {
	t.Helper()

	options := DefaultGitOptions()
	if len(opts) > 0 {
		options = opts[0]
	}

	gitconfig.Apply(t, exec, gitconfig.Options{
		UserName:             options.UserName,
		UserEmail:            options.UserEmail,
		DisableSigning:       options.DisableSigning,
		DefaultBranch:        options.DefaultBranch,
		SetInitDefaultBranch: options.InitDefaultBranch && options.DefaultBranch != "",
	})
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

// execGitCommand executes a git command in a directory and returns any error.
// This is used during setup before GitRepo is created. For debugging failed
// commands, the error message includes stderr output.
// If repo is provided, its LastStderr field will be populated by execCommandWithOutput.
func execGitCommand(t *testing.T, dir string, repo *GitRepo, args ...string) error {
	t.Helper()
	if repo == nil {
		// Create temporary repo just for execution
		repo = &GitRepo{Dir: dir}
	}
	return repo.execCommandWithOutput(t, dir, args...)
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
	err := execGitCommand(t, dir, repo, args...)
	if err != nil {
		// Only retry without -b if the error is due to an unrecognized -b flag
		errStr := err.Error()
		if strings.Contains(errStr, "unknown option") ||
			strings.Contains(errStr, "unrecognized option") ||
			strings.Contains(errStr, "unknown switch") ||
			strings.Contains(errStr, "invalid option") {
			// Older git version doesn't support -b flag
			err = execGitCommand(t, dir, repo, "init")
			require.NoError(t, err, "Failed to initialize git repository")
			if opts.DefaultBranch != "" {
				err = execGitCommand(t, dir, repo, "symbolic-ref", "HEAD", "refs/heads/"+opts.DefaultBranch)
				require.NoError(t, err, "Failed to configure HEAD for default branch")
			}
		} else {
			// Other error (permissions, disk space, etc.)
			require.NoError(t, err, "Failed to initialize git repository")
		}
	}

	// Configure git locally
	ConfigureGitClient(t, gitCommandExecutor{dir: dir}, opts)

	// Create initial commit if requested
	if opts.InitialCommit {
		err = execGitCommand(t, dir, repo, "commit", "--allow-empty", "-m", "Initial commit")
		require.NoError(t, err, "Failed to create initial commit")
	}

	// Ensure the repository is on the requested branch when an initial commit exists
	if opts.DefaultBranch != "" && opts.InitialCommit {
		err = execGitCommand(t, dir, repo, "checkout", opts.DefaultBranch)
		require.NoError(t, err, "Failed to checkout default branch")
	}

	// Add remote if requested
	if opts.AddRemote && opts.RemoteName != "" && opts.RemoteURL != "" {
		err = execGitCommand(t, dir, repo, "remote", "add", opts.RemoteName, opts.RemoteURL)
		require.NoError(t, err, "Failed to add remote")
	}

	return repo
}
