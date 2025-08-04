package errors

import (
	"errors"
	"fmt"
	"strings"
)

// Sentinel errors for common error conditions
var (
	// Ticket errors
	ErrTicketNotFound       = errors.New("ticket not found")
	ErrTicketExists         = errors.New("ticket already exists")
	ErrTicketInvalid        = errors.New("invalid ticket")
	ErrTicketNotStarted     = errors.New("ticket not started")
	ErrTicketAlreadyStarted = errors.New("ticket already started")
	ErrTicketAlreadyClosed  = errors.New("ticket already closed")
	ErrTicketNotDone        = errors.New("ticket not in done status")

	// Git errors
	ErrNotGitRepo     = errors.New("not a git repository")
	ErrDirtyWorkspace = errors.New("workspace has uncommitted changes")
	ErrBranchExists   = errors.New("branch already exists")
	ErrBranchNotFound = errors.New("branch not found")
	ErrMergeFailed    = errors.New("merge failed")
	ErrPushFailed     = errors.New("push failed")

	// Worktree errors
	ErrWorktreeExists       = errors.New("worktree already exists")
	ErrWorktreeNotFound     = errors.New("worktree not found")
	ErrWorktreeCreateFailed = errors.New("failed to create worktree")
	ErrWorktreeRemoveFailed = errors.New("failed to remove worktree")

	// Config errors
	ErrConfigNotFound = errors.New("configuration file not found")
	ErrConfigInvalid  = errors.New("invalid configuration")

	// System errors
	ErrPermissionDenied = errors.New("permission denied")
	ErrInvalidContext   = errors.New("invalid context")
)

// TicketError represents an error related to ticket operations
type TicketError struct {
	Op       string   // Operation that failed (e.g., "create", "start", "close")
	TicketID string   // ID of the ticket involved
	Err      error    // Underlying error
	Context  []string // Optional context chain (e.g., ["worktree", "create", "ticket"])
}

func (e *TicketError) Error() string {
	// Handle nil error case
	if e.Err == nil {
		var sb strings.Builder
		sb.WriteString(e.Op)
		sb.WriteString(" ticket: <nil>")
		return sb.String()
	}

	// Use strings.Builder for efficient string concatenation
	var sb strings.Builder

	// Pre-allocate estimated capacity
	estimatedSize := len(e.Op) + len(e.Err.Error()) + 10 // base overhead
	for _, ctx := range e.Context {
		estimatedSize += len(ctx) + 3 // " > "
	}
	if e.TicketID != "" {
		estimatedSize += len(e.TicketID) + 8 // " ticket "
	} else {
		estimatedSize += 7 // " ticket"
	}
	sb.Grow(estimatedSize)

	// Build context chain
	if len(e.Context) > 0 {
		for i, ctx := range e.Context {
			if i > 0 {
				sb.WriteString(" > ")
			}
			sb.WriteString(ctx)
		}
		sb.WriteString(" > ")
	}
	sb.WriteString(e.Op)

	// Always include "ticket" in the message
	if e.TicketID != "" {
		sb.WriteString(" ticket ")
		sb.WriteString(e.TicketID)
	} else {
		sb.WriteString(" ticket")
	}

	sb.WriteString(": ")
	sb.WriteString(e.Err.Error())

	return sb.String()
}

func (e *TicketError) Unwrap() error {
	return e.Err
}

// NewTicketError creates a new TicketError
func NewTicketError(op, ticketID string, err error) error {
	if op == "" {
		return fmt.Errorf("ticket error: operation cannot be empty")
	}
	if err == nil {
		return fmt.Errorf("ticket error: underlying error cannot be nil")
	}
	return &TicketError{
		Op:       op,
		TicketID: ticketID,
		Err:      err,
	}
}

// NewTicketErrorWithContext creates a new TicketError with context chain
func NewTicketErrorWithContext(op, ticketID string, err error, context ...string) error {
	if op == "" {
		return fmt.Errorf("ticket error: operation cannot be empty")
	}
	if err == nil {
		return fmt.Errorf("ticket error: underlying error cannot be nil")
	}
	return &TicketError{
		Op:       op,
		TicketID: ticketID,
		Err:      err,
		Context:  context,
	}
}

// GitError represents an error related to git operations
type GitError struct {
	Op     string // Operation that failed (e.g., "branch", "commit", "push")
	Branch string // Branch name if applicable
	Err    error  // Underlying error
}

// Error returns the string representation of the GitError
func (e *GitError) Error() string {
	if e.Branch != "" {
		return fmt.Sprintf("git %s on branch %s: %v", e.Op, e.Branch, e.Err)
	}
	return fmt.Sprintf("git %s: %v", e.Op, e.Err)
}

// Unwrap returns the underlying error for error wrapping support
func (e *GitError) Unwrap() error {
	return e.Err
}

// NewGitError creates a new GitError
func NewGitError(op, branch string, err error) error {
	if op == "" {
		return fmt.Errorf("git error: operation cannot be empty")
	}
	if err == nil {
		return fmt.Errorf("git error: underlying error cannot be nil")
	}
	return &GitError{
		Op:     op,
		Branch: branch,
		Err:    err,
	}
}

// WorktreeError represents an error related to worktree operations
type WorktreeError struct {
	Op   string // Operation that failed (e.g., "create", "remove", "list")
	Path string // Path of the worktree
	Err  error  // Underlying error
}

// Error returns the string representation of the WorktreeError
func (e *WorktreeError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("worktree %s at %s: %v", e.Op, e.Path, e.Err)
	}
	return fmt.Sprintf("worktree %s: %v", e.Op, e.Err)
}

// Unwrap returns the underlying error for error wrapping support
func (e *WorktreeError) Unwrap() error {
	return e.Err
}

// NewWorktreeError creates a new WorktreeError
func NewWorktreeError(op, path string, err error) error {
	if op == "" {
		return fmt.Errorf("worktree error: operation cannot be empty")
	}
	if err == nil {
		return fmt.Errorf("worktree error: underlying error cannot be nil")
	}
	return &WorktreeError{
		Op:   op,
		Path: path,
		Err:  err,
	}
}

// ConfigError represents an error related to configuration
type ConfigError struct {
	Field string // Configuration field that has an error
	Value string // The invalid value if applicable
	Err   error  // Underlying error
}

// Error returns the string representation of the ConfigError
func (e *ConfigError) Error() string {
	if e.Field != "" && e.Value != "" {
		return fmt.Sprintf("config field %s with value %q: %v", e.Field, e.Value, e.Err)
	} else if e.Field != "" {
		return fmt.Sprintf("config field %s: %v", e.Field, e.Err)
	}
	return fmt.Sprintf("config: %v", e.Err)
}

// Unwrap returns the underlying error for error wrapping support
func (e *ConfigError) Unwrap() error {
	return e.Err
}

// NewConfigError creates a new ConfigError
func NewConfigError(field, value string, err error) error {
	if err == nil {
		return fmt.Errorf("config error: underlying error cannot be nil")
	}
	return &ConfigError{
		Field: field,
		Value: value,
		Err:   err,
	}
}

// IsNotFound returns true if the error is a "not found" type error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrTicketNotFound) ||
		errors.Is(err, ErrWorktreeNotFound) ||
		errors.Is(err, ErrBranchNotFound) ||
		errors.Is(err, ErrConfigNotFound)
}

// IsAlreadyExists returns true if the error is an "already exists" type error
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrTicketExists) ||
		errors.Is(err, ErrBranchExists) ||
		errors.Is(err, ErrWorktreeExists) ||
		errors.Is(err, ErrTicketAlreadyStarted) ||
		errors.Is(err, ErrTicketAlreadyClosed)
}
