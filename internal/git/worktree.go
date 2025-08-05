package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ticketerrors "github.com/yshrsmz/ticketflow/internal/errors"
)

const (
	// worktreeLinesPerEntry is the average number of output lines per worktree
	// in the porcelain format (typically 3-4 lines per worktree)
	worktreeLinesPerEntry = 4
)

// WorktreeInfo represents worktree information
type WorktreeInfo struct {
	Path   string
	Branch string
	HEAD   string
}

// ListWorktrees lists all worktrees
func (g *Git) ListWorktrees(ctx context.Context) ([]WorktreeInfo, error) {
	output, err := g.Exec(ctx, SubcmdWorktree, WorktreeList, FlagPorcelain)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(output, "\n")
	// Pre-allocate worktrees slice based on number of lines
	// Each worktree typically has 3-4 lines of output
	estimatedWorktrees := len(lines) / worktreeLinesPerEntry
	if estimatedWorktrees < 1 {
		estimatedWorktrees = 1
	}
	worktrees := make([]WorktreeInfo, 0, estimatedWorktrees)

	var current WorktreeInfo
	for _, line := range lines {
		if line == "" {
			if current.Path != "" {
				worktrees = append(worktrees, current)
				current = WorktreeInfo{}
			}
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}

		switch parts[0] {
		case "worktree":
			current.Path = parts[1]
		case "HEAD":
			current.HEAD = parts[1]
		case "branch":
			current.Branch = strings.TrimPrefix(parts[1], "refs/heads/")
		}
	}

	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees, nil
}

// AddWorktree creates a new worktree
func (g *Git) AddWorktree(ctx context.Context, path, branch string) error {
	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return ticketerrors.NewWorktreeError("create", path, fmt.Errorf("failed to create worktree directory: %w", err))
	}

	// Check if branch already exists
	branchExists, err := g.BranchExists(ctx, branch)
	if err != nil {
		return ticketerrors.NewWorktreeError("create", path, fmt.Errorf("failed to check if branch exists: %w", err))
	}

	// If branch exists, check if it has diverged from the default branch
	if branchExists {
		// Get the default branch
		defaultBranch, err := g.GetDefaultBranch(ctx)
		if err != nil {
			return ticketerrors.NewWorktreeError("create", path,
				fmt.Errorf("failed to get default branch: %w", err))
		}

		// Check if branch has diverged
		diverged, err := g.BranchDivergedFrom(ctx, branch, defaultBranch)
		if err != nil {
			return ticketerrors.NewWorktreeError("create", path,
				fmt.Errorf("failed to check branch divergence: %w", err))
		}

		if diverged {
			// Get divergence details
			ahead, behind, err := g.GetBranchDivergenceInfo(ctx, branch, defaultBranch)
			if err != nil {
				return ticketerrors.NewWorktreeError("create", path,
					fmt.Errorf("failed to get divergence info: %w", err))
			}

			// Log divergence detection for debugging
			// Note: This assumes a logger is available. If not, this line can be removed.
			// log.Printf("Branch %s has diverged from %s: %d commits ahead, %d behind", branch, defaultBranch, ahead, behind)

			// Return a specific error that the CLI can handle
			return ticketerrors.NewBranchDivergenceError(branch, defaultBranch, ahead, behind)
		}

		// Branch exists and hasn't diverged, use it
		_, err = g.Exec(ctx, SubcmdWorktree, WorktreeAdd, path, branch)
	} else {
		// Branch doesn't exist, create it with -b flag
		_, err = g.Exec(ctx, SubcmdWorktree, WorktreeAdd, path, FlagBranch, branch)
	}

	return err
}

// RemoveWorktree removes a worktree
func (g *Git) RemoveWorktree(ctx context.Context, path string) error {
	_, err := g.Exec(ctx, SubcmdWorktree, WorktreeRemove, path, FlagForce)
	return err
}

// PruneWorktrees removes worktree information for deleted directories
func (g *Git) PruneWorktrees(ctx context.Context) error {
	_, err := g.Exec(ctx, SubcmdWorktree, WorktreePrune)
	return err
}

// FindWorktreeByBranch finds a worktree by its branch name
func (g *Git) FindWorktreeByBranch(ctx context.Context, branch string) (*WorktreeInfo, error) {
	worktrees, err := g.ListWorktrees(ctx)
	if err != nil {
		return nil, err
	}

	for _, wt := range worktrees {
		if wt.Branch == branch {
			return &wt, nil
		}
	}

	return nil, nil
}

// HasWorktree checks if a worktree exists for the given branch
func (g *Git) HasWorktree(ctx context.Context, branch string) (bool, error) {
	wt, err := g.FindWorktreeByBranch(ctx, branch)
	if err != nil {
		return false, err
	}
	return wt != nil, nil
}

// RunInWorktree executes a command in a specific worktree
func (g *Git) RunInWorktree(ctx context.Context, worktreePath string, args ...string) (string, error) {
	// Create a new Git instance for the worktree with same timeout
	wtGit := NewWithTimeout(worktreePath, g.timeout)
	return wtGit.Exec(ctx, args...)
}
