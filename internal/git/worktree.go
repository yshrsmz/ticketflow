package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WorktreeInfo represents worktree information
type WorktreeInfo struct {
	Path   string
	Branch string
	HEAD   string
}

// ListWorktrees lists all worktrees
func (g *Git) ListWorktrees() ([]WorktreeInfo, error) {
	output, err := g.Exec("worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	var worktrees []WorktreeInfo
	lines := strings.Split(output, "\n")

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
func (g *Git) AddWorktree(path, branch string) error {
	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create worktree directory: %w", err)
	}

	_, err := g.Exec("worktree", "add", path, "-b", branch)
	return err
}

// RemoveWorktree removes a worktree
func (g *Git) RemoveWorktree(path string) error {
	_, err := g.Exec("worktree", "remove", path, "--force")
	return err
}

// PruneWorktrees removes worktree information for deleted directories
func (g *Git) PruneWorktrees() error {
	_, err := g.Exec("worktree", "prune")
	return err
}

// FindWorktreeByBranch finds a worktree by its branch name
func (g *Git) FindWorktreeByBranch(branch string) (*WorktreeInfo, error) {
	worktrees, err := g.ListWorktrees()
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
func (g *Git) HasWorktree(branch string) (bool, error) {
	wt, err := g.FindWorktreeByBranch(branch)
	if err != nil {
		return false, err
	}
	return wt != nil, nil
}

// RunInWorktree executes a command in a specific worktree
func (g *Git) RunInWorktree(worktreePath string, args ...string) (string, error) {
	// Create a new Git instance for the worktree
	wtGit := New(worktreePath)
	return wtGit.Exec(args...)
}
