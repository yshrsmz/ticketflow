package config

import "time"

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

// Default timeout values
const (
	DefaultGitTimeoutSeconds          = 30
	DefaultInitCommandsTimeoutSeconds = 60
	DefaultGitTimeout                 = DefaultGitTimeoutSeconds * time.Second
	DefaultInitCommandsTimeout        = DefaultInitCommandsTimeoutSeconds * time.Second
	MaxTimeoutSeconds                 = 3600 // 1 hour maximum
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

// File size limits
const (
	// MaxConfigSize is the maximum allowed size for config files.
	// This prevents accidental loading of non-config files and protects
	// against malicious large files. 1MB should be more than sufficient
	// for any reasonable configuration.
	MaxConfigSize = 1024 * 1024 // 1MB
)
