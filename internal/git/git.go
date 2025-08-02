package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	ticketerrors "github.com/yshrsmz/ticketflow/internal/errors"
)

// Git provides git operations
type Git struct {
	repoPath string
	Root     string // Git repository root path
}

// New creates a new Git instance
func New(repoPath string) *Git {
	root, _ := FindProjectRoot(repoPath)
	return &Git{
		repoPath: repoPath,
		Root:     root,
	}
}

// Exec executes a git command
func (g *Git) Exec(ctx context.Context, args ...string) (string, error) {
	// Check context
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("operation cancelled: %w", err)
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
func IsGitRepo(path string) bool {
	cmd := exec.Command(GitCmd, SubcmdRevParse, FlagGitDir)
	cmd.Dir = path
	return cmd.Run() == nil
}

// FindProjectRoot finds the git project root from current directory
func FindProjectRoot(startPath string) (string, error) {
	cmd := exec.Command(GitCmd, SubcmdRevParse, FlagShowToplevel)
	cmd.Dir = startPath

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return "", ticketerrors.ErrNotGitRepo
	}

	return strings.TrimSpace(stdout.String()), nil
}

// RootPath returns the git repository root path
func (g *Git) RootPath() (string, error) {
	if g.Root == "" {
		root, err := FindProjectRoot(g.repoPath)
		if err != nil {
			return "", err
		}
		g.Root = root
	}
	return g.Root, nil
}
