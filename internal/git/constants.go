package git

// Git command names
const (
	GitCmd = "git"
)

// Git subcommands
const (
	SubcmdAdd      = "add"
	SubcmdCheckout = "checkout"
	SubcmdCommit   = "commit"
	SubcmdMerge    = "merge"
	SubcmdPull     = "pull"
	SubcmdPush     = "push"
	SubcmdRevParse = "rev-parse"
	SubcmdStatus   = "status"
	SubcmdWorktree = "worktree"
	SubcmdBranch   = "branch"
	SubcmdRemote   = "remote"
	SubcmdConfig   = "config"
	SubcmdLog      = "log"
)

// Git command flags and options
const (
	FlagAbbrevRef    = "--abbrev-ref"
	FlagPorcelain    = "--porcelain"
	FlagShowToplevel = "--show-toplevel"
	FlagSquash       = "--squash"
	FlagGitDir       = "--git-dir"
	FlagUpstream     = "-u"
	FlagBranch       = "-b"
	FlagMessage      = "-m"
	FlagVerbose      = "-v"
	FlagAll          = "-a"
	FlagDelete       = "-d"
	FlagForce        = "--force"
	FlagSet          = "--set"
	FlagUnset        = "--unset"
	FlagReplace      = "--replace-all"
)

// Git worktree subcommands
const (
	WorktreeAdd    = "add"
	WorktreeList   = "list"
	WorktreeRemove = "remove"
	WorktreePrune  = "prune"
)

// Git special references
const (
	RefHEAD = "HEAD"
)
