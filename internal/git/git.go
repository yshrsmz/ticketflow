package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	ticketerrors "github.com/yshrsmz/ticketflow/internal/errors"
)

// Git provides git operations
type Git struct {
	repoPath string
	root     string        // Git repository root path (private)
	rootOnce sync.Once     // Ensures root is initialized only once
	rootErr  error         // Error from root initialization
	timeout  time.Duration // Timeout for git operations
}

// New creates a new Git instance with default timeout
func New(repoPath string) *Git {
	return NewWithTimeout(repoPath, 30*time.Second) // 30 seconds default
}

// NewWithTimeout creates a new Git instance with custom timeout
func NewWithTimeout(repoPath string, timeout time.Duration) *Git {
	// Use background context for initialization
	root, err := FindProjectRoot(context.Background(), repoPath)
	if err != nil {
		// Not in a git repo or other error - Root will be empty
		// and lazy initialization will be attempted in RootPath()
		root = ""
	}
	return &Git{
		repoPath: repoPath,
		root:     root,
		timeout:  timeout,
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
		// Only initialize if not already set during construction
		if g.root == "" {
			// Use background context for lazy initialization
			root, err := FindProjectRoot(context.Background(), g.repoPath)
			if err != nil {
				g.rootErr = err
				return
			}
			g.root = root
		}
	})

	if g.rootErr != nil {
		return "", g.rootErr
	}
	return g.root, nil
}
