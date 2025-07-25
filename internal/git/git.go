package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
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
func (g *Git) Exec(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s failed: %w\n%s",
			strings.Join(args, " "), err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// CurrentBranch returns the current branch name
func (g *Git) CurrentBranch() (string, error) {
	return g.Exec("rev-parse", "--abbrev-ref", "HEAD")
}

// CreateBranch creates and checks out a new branch
func (g *Git) CreateBranch(name string) error {
	_, err := g.Exec("checkout", "-b", name)
	return err
}

// HasUncommittedChanges checks if there are uncommitted changes
func (g *Git) HasUncommittedChanges() (bool, error) {
	output, err := g.Exec("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return output != "", nil
}

// Add stages files
func (g *Git) Add(files ...string) error {
	args := append([]string{"add"}, files...)
	_, err := g.Exec(args...)
	return err
}

// Commit creates a commit
func (g *Git) Commit(message string) error {
	_, err := g.Exec("commit", "-m", message)
	return err
}

// Checkout switches to a branch
func (g *Git) Checkout(branch string) error {
	_, err := g.Exec("checkout", branch)
	return err
}

// MergeSquash performs a squash merge
func (g *Git) MergeSquash(branch string) error {
	_, err := g.Exec("merge", "--squash", branch)
	return err
}

// Push pushes a branch to remote
func (g *Git) Push(remote, branch string, setUpstream bool) error {
	args := []string{"push"}
	if setUpstream {
		args = append(args, "-u")
	}
	args = append(args, remote, branch)
	_, err := g.Exec(args...)
	return err
}

// IsGitRepo checks if the path is a git repository
func IsGitRepo(path string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = path
	return cmd.Run() == nil
}

// FindProjectRoot finds the git project root from current directory
func FindProjectRoot(startPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = startPath
	
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("not in a git repository")
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