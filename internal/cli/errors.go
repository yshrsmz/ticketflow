package cli

import (
	"encoding/json"
	"fmt"
	"os"
)

// GlobalOutputFormat is set by command parsing to control error output format
var GlobalOutputFormat OutputFormat = FormatText

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
func HandleError(err error) {
	if err == nil {
		return
	}

	// Check if it's a CLI error
	if cliErr, ok := err.(*CLIError); ok {
		handleCLIError(cliErr)
		return
	}

	// Generic error
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}

func handleCLIError(err *CLIError) {
	// Check global format first (set by command line parsing)
	if GlobalOutputFormat == FormatJSON {
		OutputJSONError(err)
		return
	}
	
	// Check if JSON output is requested via environment variable
	// This allows error formatting even before app initialization
	if os.Getenv("TICKETFLOW_OUTPUT_FORMAT") == "json" {
		OutputJSONError(err)
		return
	}
	
	// Default text format
	fmt.Fprintf(os.Stderr, "Error: %s\n", err.Message)
	
	if err.Details != "" {
		fmt.Fprintf(os.Stderr, "Details: %s\n", err.Details)
	}
	
	if len(err.Suggestions) > 0 {
		fmt.Fprintf(os.Stderr, "\nSuggestions:\n")
		for _, suggestion := range err.Suggestions {
			fmt.Fprintf(os.Stderr, "  - %s\n", suggestion)
		}
	}
}

// OutputJSONError outputs error in JSON format
func OutputJSONError(err *CLIError) {
	output := map[string]interface{}{
		"error": err,
	}
	
	encoder := json.NewEncoder(os.Stderr)
	encoder.SetIndent("", "  ")
	encoder.Encode(output)
}