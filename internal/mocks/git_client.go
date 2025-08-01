package mocks

import (
	"github.com/stretchr/testify/mock"
	"github.com/yshrsmz/ticketflow/internal/git"
)

// MockGitClient is a mock implementation of git.GitClient
type MockGitClient struct {
	mock.Mock
}

// Exec executes a git command
func (m *MockGitClient) Exec(args ...string) (string, error) {
	// Convert variadic args to interface slice for testify
	iArgs := make([]interface{}, len(args))
	for i, arg := range args {
		iArgs[i] = arg
	}
	mockArgs := m.Called(iArgs...)
	return mockArgs.String(0), mockArgs.Error(1)
}

// CurrentBranch returns the current branch name
func (m *MockGitClient) CurrentBranch() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// CreateBranch creates and checks out a new branch
func (m *MockGitClient) CreateBranch(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

// HasUncommittedChanges checks if there are uncommitted changes
func (m *MockGitClient) HasUncommittedChanges() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

// Add adds files to the staging area
func (m *MockGitClient) Add(files ...string) error {
	// Convert variadic args to interface slice for testify
	iArgs := make([]interface{}, len(files))
	for i, file := range files {
		iArgs[i] = file
	}
	args := m.Called(iArgs...)
	return args.Error(0)
}

// Commit creates a commit with the given message
func (m *MockGitClient) Commit(message string) error {
	args := m.Called(message)
	return args.Error(0)
}

// Checkout switches to the specified branch
func (m *MockGitClient) Checkout(branch string) error {
	args := m.Called(branch)
	return args.Error(0)
}

// MergeSquash performs a squash merge of the specified branch
func (m *MockGitClient) MergeSquash(branch string) error {
	args := m.Called(branch)
	return args.Error(0)
}

// Push pushes changes to the remote repository
func (m *MockGitClient) Push(remote, branch string, setUpstream bool) error {
	args := m.Called(remote, branch, setUpstream)
	return args.Error(0)
}

// RootPath returns the root path of the git repository
func (m *MockGitClient) RootPath() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// ListWorktrees returns a list of worktrees
func (m *MockGitClient) ListWorktrees() ([]git.WorktreeInfo, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]git.WorktreeInfo), args.Error(1)
}

// AddWorktree creates a new worktree
func (m *MockGitClient) AddWorktree(path, branch string) error {
	args := m.Called(path, branch)
	return args.Error(0)
}

// RemoveWorktree removes a worktree
func (m *MockGitClient) RemoveWorktree(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

// PruneWorktrees prunes worktree information
func (m *MockGitClient) PruneWorktrees() error {
	args := m.Called()
	return args.Error(0)
}

// FindWorktreeByBranch finds a worktree by branch name
func (m *MockGitClient) FindWorktreeByBranch(branch string) (*git.WorktreeInfo, error) {
	args := m.Called(branch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*git.WorktreeInfo), args.Error(1)
}

// HasWorktree checks if a worktree exists for the branch
func (m *MockGitClient) HasWorktree(branch string) (bool, error) {
	args := m.Called(branch)
	return args.Bool(0), args.Error(1)
}

// RunInWorktree runs a command in a specific worktree
func (m *MockGitClient) RunInWorktree(worktreePath string, cmdArgs ...string) (string, error) {
	// Convert to interface slice for testify
	iArgs := make([]interface{}, len(cmdArgs)+1)
	iArgs[0] = worktreePath
	for i, arg := range cmdArgs {
		iArgs[i+1] = arg
	}
	args := m.Called(iArgs...)
	return args.String(0), args.Error(1)
}
