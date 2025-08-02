package config

// Configuration file names
const (
	ConfigFileName = ".ticketflow.yaml"
)

// Default configuration values
const (
	DefaultBranch       = "main"
	DefaultWorktreeBase = "../.worktrees"
	DefaultTicketsDir   = "tickets"
	DefaultTodoDir      = "todo"
	DefaultDoingDir     = "doing"
	DefaultDoneDir      = "done"
	DefaultOutputFormat = "text"
)

// Output format types
const (
	FormatText = "text"
	FormatJSON = "json"
)

// Default permissions
const (
	DirPermission  = 0755
	FilePermission = 0644
)
