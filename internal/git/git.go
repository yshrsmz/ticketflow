package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	ticketerrors "github.com/yshrsmz/ticketflow/internal/errors"
)

const (
	// DefaultGitTimeout is the default timeout for git operations
	DefaultGitTimeout = 30 * time.Second

	// TestGitTimeout is a shorter timeout suitable for tests
	TestGitTimeout = 10 * time.Second
)

// Git provides git operations
type Git struct {
	repoPath string
	root     string        // Git repository root path (private)
	rootOnce sync.Once     // Ensures root is initialized only once
	rootErr  error         // Error from root initialization
	timeout  time.Duration // Timeout for git operations
}

// isValidBranchName validates a git branch name according to git-check-ref-format rules
// This function validates branch names based on Git's ref naming rules:
// https://git-scm.com/docs/git-check-ref-format
func isValidBranchName(name string) bool {
	// Rule 1: Branch name cannot be empty
	if name == "" {
		return false
	}

	// Rule 2: Check for invalid multi-character patterns
	// - Cannot contain ".." (directory traversal)
	// - Cannot contain "@{" (reflog syntax)
	// - Cannot contain "//" (consecutive slashes)
	if strings.Contains(name, "..") || strings.Contains(name, "@{") || strings.Contains(name, "//") {
		return false
	}

	// Rule 3: Cannot begin or end with certain characters
	// - Cannot start with: slash (/), dot (.), or whitespace
	// - Cannot end with: slash (/), dot (.), or whitespace
	if len(name) > 0 {
		firstChar := name[0]
		lastChar := name[len(name)-1]

		// Check first character
		if firstChar == '/' || firstChar == '.' || isWhitespace(firstChar) {
			return false
		}

		// Check last character
		if lastChar == '/' || lastChar == '.' || isWhitespace(lastChar) {
			return false
		}
	}

	// Rule 4: Validate each character in the branch name
	// Invalid characters anywhere in the name:
	// - Control characters (ASCII 0-31, 127)
	// - Whitespace characters (space, tab, etc.)
	// - Special git characters: colon (:), question mark (?), asterisk (*),
	//   open bracket ([), backslash (\)
	for _, ch := range name {
		if !isValidBranchChar(ch) {
			return false
		}
	}

	return true
}

// isValidBranchChar checks if a character is valid in a git branch name
func isValidBranchChar(ch rune) bool {
	// Check for control characters (ASCII 0-31 and 127)
	if ch <= 0x1f || ch == 0x7f {
		return false
	}

	// Check for whitespace characters
	// Only check ASCII whitespace since git branch names should not contain Unicode whitespace
	if ch <= 0x7F && isWhitespace(byte(ch)) {
		return false
	}

	// Check for forbidden special characters
	switch ch {
	case ':', '?', '*', '[', '\\':
		return false
	}

	return true
}

// isWhitespace checks if a byte is a whitespace character
func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\f' || b == '\v'
}

// validateTimeout ensures the timeout is positive, defaulting to DefaultGitTimeout if not
func validateTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return DefaultGitTimeout
	}
	return timeout
}

// New creates a new Git instance with default timeout
func New(repoPath string) *Git {
	if repoPath == "" {
		repoPath = "."
	}

	// Note: We don't validate path existence here - git commands will fail with appropriate errors
	// This allows for cases where the directory will be created later

	return NewWithTimeout(repoPath, DefaultGitTimeout)
}

// NewWithTimeout creates a new Git instance with custom timeout
func NewWithTimeout(repoPath string, timeout time.Duration) *Git {
	// Validate inputs
	if repoPath == "" {
		// Default to current directory if empty
		repoPath = "."
	}

	return &Git{
		repoPath: repoPath,
		timeout:  validateTimeout(timeout),
	}
}

// Exec executes a git command
func (g *Git) Exec(ctx context.Context, args ...string) (string, error) {
	// Check context
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("operation cancelled: %w", err)
	}

	// Apply timeout if not already set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && g.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, GitCmd, args...)
	cmd.Dir = g.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Extract the git subcommand and branch if applicable
		subcommand := ""
		branch := ""
		if len(args) > 0 {
			subcommand = args[0]
		}
		// For branch-related commands, try to extract branch name
		if len(args) > 1 && (subcommand == SubcmdCheckout || subcommand == SubcmdPush || subcommand == SubcmdPull || subcommand == SubcmdMerge) {
			branch = args[len(args)-1]
		}

		// Check if error is due to timeout
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			gitErr := ticketerrors.NewGitError(subcommand, branch,
				fmt.Errorf("operation timed out after %v: %w", g.timeout, err))
			return "", gitErr
		}

		gitErr := ticketerrors.NewGitError(subcommand, branch,
			fmt.Errorf("command failed: %w\n%s", err, stderr.String()))
		return "", gitErr
	}

	return strings.TrimSpace(stdout.String()), nil
}

// CurrentBranch returns the current branch name
func (g *Git) CurrentBranch(ctx context.Context) (string, error) {
	return g.Exec(ctx, SubcmdRevParse, FlagAbbrevRef, RefHEAD)
}

// CreateBranch creates and checks out a new branch
func (g *Git) CreateBranch(ctx context.Context, name string) error {
	_, err := g.Exec(ctx, SubcmdCheckout, FlagBranch, name)
	return err
}

// HasUncommittedChanges checks if there are uncommitted changes
func (g *Git) HasUncommittedChanges(ctx context.Context) (bool, error) {
	output, err := g.Exec(ctx, SubcmdStatus, FlagPorcelain)
	if err != nil {
		return false, err
	}
	return output != "", nil
}

// Add stages files
func (g *Git) Add(ctx context.Context, files ...string) error {
	args := append([]string{SubcmdAdd}, files...)
	_, err := g.Exec(ctx, args...)
	return err
}

// Commit creates a commit
func (g *Git) Commit(ctx context.Context, message string) error {
	_, err := g.Exec(ctx, SubcmdCommit, FlagMessage, message)
	return err
}

// Checkout switches to a branch
func (g *Git) Checkout(ctx context.Context, branch string) error {
	_, err := g.Exec(ctx, SubcmdCheckout, branch)
	return err
}

// BranchExists checks if a branch exists locally
func (g *Git) BranchExists(ctx context.Context, branch string) (bool, error) {
	// Validate branch name to prevent command injection
	if !isValidBranchName(branch) {
		return false, fmt.Errorf("invalid branch name: %s", branch)
	}

	// Use git show-ref --verify --quiet refs/heads/<branch>
	// This command returns exit code 0 if branch exists, non-zero otherwise
	_, err := g.Exec(ctx, SubcmdShowRef, FlagVerify, FlagQuiet, fmt.Sprintf("refs/heads/%s", branch))
	if err != nil {
		// Check if this is a git error (branch doesn't exist) vs actual error
		if _, ok := err.(*ticketerrors.GitError); ok {
			// The command returns non-zero when branch doesn't exist, which is expected
			return false, nil
		}
		// Some other error occurred
		return false, err
	}
	// Branch exists
	return true, nil
}

// MergeSquash performs a squash merge
func (g *Git) MergeSquash(ctx context.Context, branch string) error {
	_, err := g.Exec(ctx, SubcmdMerge, FlagSquash, branch)
	return err
}

// Push pushes a branch to remote
func (g *Git) Push(ctx context.Context, remote, branch string, setUpstream bool) error {
	args := []string{SubcmdPush}
	if setUpstream {
		args = append(args, FlagUpstream)
	}
	args = append(args, remote, branch)
	_, err := g.Exec(ctx, args...)
	return err
}

// GetDefaultBranch returns the configured default branch (main/master)
func (g *Git) GetDefaultBranch(ctx context.Context) (string, error) {
	// Try to get from remote HEAD
	output, err := g.Exec(ctx, SubcmdRevParse, FlagAbbrevRef, "origin/HEAD")
	if err == nil {
		// Remove "origin/" prefix
		branch := strings.TrimPrefix(strings.TrimSpace(output), "origin/")
		if branch != "" && branch != "HEAD" {
			return branch, nil
		}
	}

	// Fallback to checking common default branch names
	for _, branch := range []string{"main", "master"} {
		if exists, _ := g.BranchExists(ctx, branch); exists {
			return branch, nil
		}
	}

	return "", fmt.Errorf("could not determine default branch")
}

// GetBranchCommit gets the commit hash a branch points to
func (g *Git) GetBranchCommit(ctx context.Context, branch string) (string, error) {
	// Validate branch name to prevent command injection
	if !isValidBranchName(branch) {
		return "", fmt.Errorf("invalid branch name: %s", branch)
	}

	output, err := g.Exec(ctx, SubcmdRevParse, branch)
	if err != nil {
		return "", fmt.Errorf("failed to get commit for branch %s: %w", branch, err)
	}

	return strings.TrimSpace(output), nil
}

// BranchDivergedFrom checks if a branch has diverged from a base branch
func (g *Git) BranchDivergedFrom(ctx context.Context, branch, baseBranch string) (bool, error) {
	branchCommit, err := g.GetBranchCommit(ctx, branch)
	if err != nil {
		return false, fmt.Errorf("failed to get commit for branch %s: %w", branch, err)
	}

	baseCommit, err := g.GetBranchCommit(ctx, baseBranch)
	if err != nil {
		return false, fmt.Errorf("failed to get commit for base branch %s: %w", baseBranch, err)
	}

	return branchCommit != baseCommit, nil
}

// GetBranchDivergenceInfo returns commits ahead/behind between branches
func (g *Git) GetBranchDivergenceInfo(ctx context.Context, branch, baseBranch string) (ahead, behind int, err error) {
	// Validate branch names
	if !isValidBranchName(branch) || !isValidBranchName(baseBranch) {
		return 0, 0, fmt.Errorf("invalid branch name")
	}

	// Get commits ahead (in branch but not in baseBranch)
	aheadOutput, err := g.Exec(ctx, SubcmdRevList, FlagCount, fmt.Sprintf("%s..%s", baseBranch, branch))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count commits ahead: %w", err)
	}

	// Get commits behind (in baseBranch but not in branch)
	behindOutput, err := g.Exec(ctx, SubcmdRevList, FlagCount, fmt.Sprintf("%s..%s", branch, baseBranch))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count commits behind: %w", err)
	}

	// Parse the counts (ignore error, will be 0 if parse fails)
	ahead, _ = strconv.Atoi(strings.TrimSpace(aheadOutput))
	behind, _ = strconv.Atoi(strings.TrimSpace(behindOutput))

	return ahead, behind, nil
}

// IsGitRepo checks if the path is a git repository
func IsGitRepo(ctx context.Context, path string) bool {
	cmd := exec.CommandContext(ctx, GitCmd, SubcmdRevParse, FlagGitDir)
	cmd.Dir = path
	return cmd.Run() == nil
}

// FindProjectRoot finds the git project root from current directory
func FindProjectRoot(ctx context.Context, startPath string) (string, error) {
	cmd := exec.CommandContext(ctx, GitCmd, SubcmdRevParse, FlagShowToplevel)
	cmd.Dir = startPath

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return "", ticketerrors.ErrNotGitRepo
	}

	return strings.TrimSpace(stdout.String()), nil
}

// RootPath returns the git repository root path (thread-safe)
func (g *Git) RootPath() (string, error) {
	g.rootOnce.Do(func() {
		// Use background context for lazy initialization
		root, err := FindProjectRoot(context.Background(), g.repoPath)
		if err != nil {
			g.rootErr = err
			return
		}
		g.root = root
	})

	if g.rootErr != nil {
		return "", g.rootErr
	}
	return g.root, nil
}
