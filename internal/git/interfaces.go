package git

import "context"

// BasicGitClient defines the interface for basic git operations
type BasicGitClient interface {
	// Core git operations
	Exec(ctx context.Context, args ...string) (string, error)
	CurrentBranch(ctx context.Context) (string, error)
	CreateBranch(ctx context.Context, name string) error
	BranchExists(ctx context.Context, branch string) (bool, error)
	HasUncommittedChanges(ctx context.Context) (bool, error)
	Add(ctx context.Context, files ...string) error
	Commit(ctx context.Context, message string) error
	Checkout(ctx context.Context, branch string) error
	MergeSquash(ctx context.Context, branch string) error
	Push(ctx context.Context, remote, branch string, setUpstream bool) error
	RootPath() (string, error)
}

// WorktreeClient defines the interface for git worktree operations
type WorktreeClient interface {
	BasicGitClient

	// Worktree-specific operations
	ListWorktrees(ctx context.Context) ([]WorktreeInfo, error)
	AddWorktree(ctx context.Context, path, branch string) error
	RemoveWorktree(ctx context.Context, path string) error
	PruneWorktrees(ctx context.Context) error
	FindWorktreeByBranch(ctx context.Context, branch string) (*WorktreeInfo, error)
	HasWorktree(ctx context.Context, branch string) (bool, error)
	RunInWorktree(ctx context.Context, worktreePath string, args ...string) (string, error)
}

// GitClient is the complete interface combining basic and worktree operations
// This maintains backward compatibility
type GitClient interface {
	WorktreeClient

	// Branch divergence operations
	GetDefaultBranch(ctx context.Context) (string, error)
	BranchDivergedFrom(ctx context.Context, branch, baseBranch string) (bool, error)
	GetBranchCommit(ctx context.Context, branch string) (string, error)
	GetBranchDivergenceInfo(ctx context.Context, branch, baseBranch string) (ahead, behind int, err error)
	IsBranchMerged(ctx context.Context, branch, targetBranch string) (bool, error)
}
