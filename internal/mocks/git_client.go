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

// Exec executes a git command
func (m *MockGitClient) Exec(ctx context.Context, args ...string) (string, error) {
	// Convert variadic args to interface slice for testify
	iArgs := make([]interface{}, len(args)+1)
	iArgs[0] = ctx
	for i, arg := range args {
		iArgs[i+1] = arg
	}
	mockArgs := m.Called(iArgs...)
	return mockArgs.String(0), mockArgs.Error(1)
}

// CurrentBranch returns the current branch name
func (m *MockGitClient) CurrentBranch(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

// CreateBranch creates and checks out a new branch
func (m *MockGitClient) CreateBranch(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

// HasUncommittedChanges checks if there are uncommitted changes
func (m *MockGitClient) HasUncommittedChanges(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

// Add adds files to the staging area
func (m *MockGitClient) Add(ctx context.Context, files ...string) error {
	// Convert variadic args to interface slice for testify
	iArgs := make([]interface{}, len(files)+1)
	iArgs[0] = ctx
	for i, file := range files {
		iArgs[i+1] = file
	}
	args := m.Called(iArgs...)
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

// Push pushes changes to the remote repository
func (m *MockGitClient) Push(ctx context.Context, remote, branch string, setUpstream bool) error {
	args := m.Called(ctx, remote, branch, setUpstream)
	return args.Error(0)
}

// RootPath returns the root path of the git repository
func (m *MockGitClient) RootPath() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// ListWorktrees returns a list of worktrees
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

// PruneWorktrees prunes worktree information
func (m *MockGitClient) PruneWorktrees(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// FindWorktreeByBranch finds a worktree by branch name
func (m *MockGitClient) FindWorktreeByBranch(ctx context.Context, branch string) (*git.WorktreeInfo, error) {
	args := m.Called(ctx, branch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*git.WorktreeInfo), args.Error(1)
}

// HasWorktree checks if a worktree exists for the branch
func (m *MockGitClient) HasWorktree(ctx context.Context, branch string) (bool, error) {
	args := m.Called(ctx, branch)
	return args.Bool(0), args.Error(1)
}

// RunInWorktree runs a command in a specific worktree
func (m *MockGitClient) RunInWorktree(ctx context.Context, worktreePath string, cmdArgs ...string) (string, error) {
	// Convert to interface slice for testify
	iArgs := make([]interface{}, len(cmdArgs)+2)
	iArgs[0] = ctx
	iArgs[1] = worktreePath
	for i, arg := range cmdArgs {
		iArgs[i+2] = arg
	}
	args := m.Called(iArgs...)
	return args.String(0), args.Error(1)
}
