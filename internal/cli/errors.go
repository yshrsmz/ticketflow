package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// globalOutputFormat is set by command parsing to control error output format
var (
	globalOutputFormat OutputFormat = FormatText
	formatMutex        sync.RWMutex
)

// SetGlobalOutputFormat sets the global output format in a thread-safe manner
func SetGlobalOutputFormat(format OutputFormat) {
	formatMutex.Lock()
	defer formatMutex.Unlock()
	globalOutputFormat = format
}

// GetGlobalOutputFormat gets the global output format in a thread-safe manner
func GetGlobalOutputFormat() OutputFormat {
	formatMutex.RLock()
	defer formatMutex.RUnlock()
	return globalOutputFormat
}

// Error codes
const (
	// System errors
	ErrNotGitRepo       = "NOT_GIT_REPO"
	ErrConfigNotFound   = "CONFIG_NOT_FOUND"
	ErrConfigInvalid    = "CONFIG_INVALID"
	ErrPermissionDenied = "PERMISSION_DENIED"

	// Ticket errors
	ErrTicketNotFound       = "TICKET_NOT_FOUND"
	ErrTicketExists         = "TICKET_EXISTS"
	ErrTicketInvalid        = "TICKET_INVALID"
	ErrTicketNotStarted     = "TICKET_NOT_STARTED"
	ErrTicketAlreadyStarted = "TICKET_ALREADY_STARTED"
	ErrTicketAlreadyClosed  = "TICKET_ALREADY_CLOSED"
	ErrTicketNotDone        = "TICKET_NOT_DONE"

	// Git errors
	ErrGitDirtyWorkspace = "GIT_DIRTY_WORKSPACE"
	ErrGitBranchExists   = "GIT_BRANCH_EXISTS"
	ErrGitMergeFailed    = "GIT_MERGE_FAILED"
	ErrGitPushFailed     = "GIT_PUSH_FAILED"

	// Worktree errors
	ErrWorktreeExists       = "WORKTREE_EXISTS"
	ErrWorktreeNotFound     = "WORKTREE_NOT_FOUND"
	ErrWorktreeCreateFailed = "WORKTREE_CREATE_FAILED"
	ErrWorktreeRemoveFailed = "WORKTREE_REMOVE_FAILED"
	ErrInvalidContext       = "INVALID_CONTEXT"
)

// CLIError represents a structured error for CLI output
type CLIError struct {
	Code        string   `json:"code"`
	Message     string   `json:"message"`
	Details     string   `json:"details,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// Error implements the error interface
func (e *CLIError) Error() string {
	return e.Message
}

// NewError creates a new CLI error
func NewError(code, message, details string, suggestions []string) *CLIError {
	return &CLIError{
		Code:        code,
		Message:     message,
		Details:     details,
		Suggestions: suggestions,
	}
}

// HandleError handles errors appropriately based on output format
// Deprecated: Use OutputWriter.Error instead for better testability and thread safety
func HandleError(err error) {
	if err == nil {
		return
	}

	// Determine output format
	format := GetGlobalOutputFormat()

	// Check if JSON output is requested via environment variable
	// This allows error formatting even before app initialization
	if format == FormatText && os.Getenv("TICKETFLOW_OUTPUT_FORMAT") == "json" {
		format = FormatJSON
	}

	// Create default writer and delegate
	writer := NewOutputWriter(nil, nil, format)
	writer.Error(err)
}

// Error writes an error to stderr based on the output format
func (w *OutputWriter) Error(err error) {
	if err == nil {
		return
	}

	// Check if it's a CLI error
	if cliErr, ok := err.(*CLIError); ok {
		w.handleCLIError(cliErr)
		return
	}

	// Generic error
	_, _ = fmt.Fprintf(w.stderr, "Error: %v\n", err)
}

func (w *OutputWriter) handleCLIError(err *CLIError) {
	if w.format == FormatJSON {
		w.outputJSONError(err)
		return
	}

	// Default text format
	_, _ = fmt.Fprintf(w.stderr, "Error: %s\n", err.Message)

	if err.Details != "" {
		_, _ = fmt.Fprintf(w.stderr, "Details: %s\n", err.Details)
	}

	if len(err.Suggestions) > 0 {
		_, _ = fmt.Fprintf(w.stderr, "\nSuggestions:\n")
		for _, suggestion := range err.Suggestions {
			_, _ = fmt.Fprintf(w.stderr, "  - %s\n", suggestion)
		}
	}
}

func (w *OutputWriter) outputJSONError(err *CLIError) {
	output := map[string]interface{}{
		"error": err,
	}

	encoder := json.NewEncoder(w.stderr)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(output)
}

// OutputJSONError outputs error in JSON format
func OutputJSONError(err *CLIError) {
	output := map[string]interface{}{
		"error": err,
	}

	encoder := json.NewEncoder(os.Stderr)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(output)
}
