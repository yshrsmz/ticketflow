package git

// BasicGitClient defines the interface for basic git operations
type BasicGitClient interface {
	// Core git operations
	Exec(args ...string) (string, error)
	CurrentBranch() (string, error)
	CreateBranch(name string) error
	HasUncommittedChanges() (bool, error)
	Add(files ...string) error
	Commit(message string) error
	Checkout(branch string) error
	MergeSquash(branch string) error
	Push(remote, branch string, setUpstream bool) error
	RootPath() (string, error)
}

// WorktreeClient defines the interface for git worktree operations
type WorktreeClient interface {
	BasicGitClient
	
	// Worktree-specific operations
	ListWorktrees() ([]WorktreeInfo, error)
	AddWorktree(path, branch string) error
	RemoveWorktree(path string) error
	PruneWorktrees() error
	FindWorktreeByBranch(branch string) (*WorktreeInfo, error)
	HasWorktree(branch string) (bool, error)
	RunInWorktree(worktreePath string, args ...string) (string, error)
}

// GitClient is the complete interface combining basic and worktree operations
// This maintains backward compatibility
type GitClient interface {
	WorktreeClient
}