package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/yshrsmz/ticketflow/internal/git"
)

// MockGitClient is a mock implementation of git.GitClient
type MockGitClient struct {
	mock.Mock
}

// Exec executes a git command with the given arguments
func (m *MockGitClient) Exec(ctx context.Context, args ...string) (string, error) {
	varArgs := []interface{}{ctx}
	for _, arg := range args {
		varArgs = append(varArgs, arg)
	}
	mockArgs := m.Called(varArgs...)
	return mockArgs.String(0), mockArgs.Error(1)
}

// CurrentBranch returns the current branch name
func (m *MockGitClient) CurrentBranch(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

// CreateBranch creates a new git branch
func (m *MockGitClient) CreateBranch(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

// HasUncommittedChanges checks if there are uncommitted changes
func (m *MockGitClient) HasUncommittedChanges(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

// Add stages files for commit
func (m *MockGitClient) Add(ctx context.Context, files ...string) error {
	varArgs := []interface{}{ctx}
	for _, file := range files {
		varArgs = append(varArgs, file)
	}
	args := m.Called(varArgs...)
	return args.Error(0)
}

// Commit creates a commit with the given message
func (m *MockGitClient) Commit(ctx context.Context, message string) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

// Checkout switches to the specified branch
func (m *MockGitClient) Checkout(ctx context.Context, branch string) error {
	args := m.Called(ctx, branch)
	return args.Error(0)
}

// MergeSquash performs a squash merge of the specified branch
func (m *MockGitClient) MergeSquash(ctx context.Context, branch string) error {
	args := m.Called(ctx, branch)
	return args.Error(0)
}

// Push pushes a branch to remote
func (m *MockGitClient) Push(ctx context.Context, remote, branch string, setUpstream bool) error {
	args := m.Called(ctx, remote, branch, setUpstream)
	return args.Error(0)
}

// RootPath returns the root path of the git repository
func (m *MockGitClient) RootPath() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// BranchExists checks if a branch exists locally
func (m *MockGitClient) BranchExists(ctx context.Context, branch string) (bool, error) {
	args := m.Called(ctx, branch)
	return args.Bool(0), args.Error(1)
}

// ListWorktrees lists all worktrees
func (m *MockGitClient) ListWorktrees(ctx context.Context) ([]git.WorktreeInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]git.WorktreeInfo), args.Error(1)
}

// AddWorktree creates a new worktree
func (m *MockGitClient) AddWorktree(ctx context.Context, path, branch string) error {
	args := m.Called(ctx, path, branch)
	return args.Error(0)
}

// RemoveWorktree removes a worktree
func (m *MockGitClient) RemoveWorktree(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

// PruneWorktrees removes worktree information for deleted directories
func (m *MockGitClient) PruneWorktrees(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// FindWorktreeByBranch finds a worktree by its branch name
func (m *MockGitClient) FindWorktreeByBranch(ctx context.Context, branch string) (*git.WorktreeInfo, error) {
	args := m.Called(ctx, branch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*git.WorktreeInfo), args.Error(1)
}

// HasWorktree checks if a worktree exists for the given branch
func (m *MockGitClient) HasWorktree(ctx context.Context, branch string) (bool, error) {
	args := m.Called(ctx, branch)
	return args.Bool(0), args.Error(1)
}

// RunInWorktree executes a command in a specific worktree
func (m *MockGitClient) RunInWorktree(ctx context.Context, worktreePath string, cmdArgs ...string) (string, error) {
	varArgs := []interface{}{ctx, worktreePath}
	for _, arg := range cmdArgs {
		varArgs = append(varArgs, arg)
	}
	args := m.Called(varArgs...)
	return args.String(0), args.Error(1)
}

// GetDefaultBranch returns the configured default branch (main/master)
func (m *MockGitClient) GetDefaultBranch(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

// BranchDivergedFrom checks if a branch has diverged from a base branch
func (m *MockGitClient) BranchDivergedFrom(ctx context.Context, branch, baseBranch string) (bool, error) {
	args := m.Called(ctx, branch, baseBranch)
	return args.Bool(0), args.Error(1)
}

// GetBranchCommit gets the commit hash a branch points to
func (m *MockGitClient) GetBranchCommit(ctx context.Context, branch string) (string, error) {
	args := m.Called(ctx, branch)
	return args.String(0), args.Error(1)
}

// GetBranchDivergenceInfo returns commits ahead/behind between branches
func (m *MockGitClient) GetBranchDivergenceInfo(ctx context.Context, branch, baseBranch string) (ahead, behind int, err error) {
	args := m.Called(ctx, branch, baseBranch)
	return args.Int(0), args.Int(1), args.Error(2)
}