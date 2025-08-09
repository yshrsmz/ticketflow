package cli

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	ticketerrors "github.com/yshrsmz/ticketflow/internal/errors"
)

func TestConvertError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedCode   string
		expectedMsg    string
		checkDetails   bool
		expectedDetail string
		hasSuggestions bool
	}{
		{
			name:         "nil error",
			err:          nil,
			expectedCode: "",
		},
		{
			name:         "already CLI error",
			err:          NewError(ErrTicketNotFound, "Not found", "details", nil),
			expectedCode: ErrTicketNotFound,
			expectedMsg:  "Not found",
		},
		{
			name:           "ticket not found",
			err:            ticketerrors.ErrTicketNotFound,
			expectedCode:   ErrTicketNotFound,
			expectedMsg:    "Ticket not found",
			hasSuggestions: true,
		},
		{
			name:           "ticket exists",
			err:            ticketerrors.ErrTicketExists,
			expectedCode:   ErrTicketExists,
			expectedMsg:    "Ticket already exists",
			hasSuggestions: true,
		},
		{
			name:           "ticket already started",
			err:            ticketerrors.ErrTicketAlreadyStarted,
			expectedCode:   ErrTicketAlreadyStarted,
			expectedMsg:    "Ticket already started",
			hasSuggestions: true,
		},
		{
			name:           "ticket already closed",
			err:            ticketerrors.ErrTicketAlreadyClosed,
			expectedCode:   ErrTicketAlreadyClosed,
			expectedMsg:    "Ticket already closed",
			hasSuggestions: true,
		},
		{
			name:           "ticket not started",
			err:            ticketerrors.ErrTicketNotStarted,
			expectedCode:   ErrTicketNotStarted,
			expectedMsg:    "Ticket not started",
			hasSuggestions: true,
		},
		{
			name:           "not git repo",
			err:            ticketerrors.ErrNotGitRepo,
			expectedCode:   ErrNotGitRepo,
			expectedMsg:    "Not in a git repository",
			hasSuggestions: true,
		},
		{
			name:           "worktree exists",
			err:            ticketerrors.ErrWorktreeExists,
			expectedCode:   ErrWorktreeExists,
			expectedMsg:    "Worktree already exists",
			hasSuggestions: true,
		},
		{
			name:           "worktree not found",
			err:            ticketerrors.ErrWorktreeNotFound,
			expectedCode:   ErrWorktreeNotFound,
			expectedMsg:    "Worktree not found",
			hasSuggestions: true,
		},
		{
			name:           "config not found",
			err:            ticketerrors.ErrConfigNotFound,
			expectedCode:   ErrConfigNotFound,
			expectedMsg:    "Configuration not found",
			hasSuggestions: true,
		},
		{
			name:           "config invalid",
			err:            ticketerrors.ErrConfigInvalid,
			expectedCode:   ErrConfigInvalid,
			expectedMsg:    "Invalid configuration",
			hasSuggestions: true,
		},
		{
			name:         "ticket error type",
			err:          ticketerrors.NewTicketError("create", "test-123", fmt.Errorf("failed")),
			expectedCode: ErrTicketInvalid,
			expectedMsg:  "Ticket operation failed: create",
		},
		{
			name:         "git error type",
			err:          ticketerrors.NewGitError("checkout", "main", fmt.Errorf("failed")),
			expectedCode: ErrGitMergeFailed,
			expectedMsg:  "Git operation failed: checkout",
		},
		{
			name:         "worktree error create",
			err:          ticketerrors.NewWorktreeError("create", "test-123", fmt.Errorf("failed")),
			expectedCode: ErrWorktreeCreateFailed,
			expectedMsg:  "Worktree operation failed: create",
		},
		{
			name:         "worktree error remove",
			err:          ticketerrors.NewWorktreeError("remove", "test-123", fmt.Errorf("failed")),
			expectedCode: ErrWorktreeRemoveFailed,
			expectedMsg:  "Worktree operation failed: remove",
		},
		{
			name:           "config error type",
			err:            ticketerrors.NewConfigError("worktree.baseDir", "/invalid/path", fmt.Errorf("invalid path")),
			expectedCode:   ErrConfigInvalid,
			expectedMsg:    "Configuration error",
			hasSuggestions: true,
		},
		{
			name: "generic error",
			err:  fmt.Errorf("something went wrong"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertError(tt.err)

			if tt.err == nil {
				assert.Nil(t, result)
				return
			}

			if tt.expectedCode == "" {
				// Generic error case
				assert.Equal(t, tt.err, result)
				return
			}

			cliErr, ok := result.(*CLIError)
			assert.True(t, ok, "Expected CLIError type")

			assert.Equal(t, tt.expectedCode, cliErr.Code)
			assert.Equal(t, tt.expectedMsg, cliErr.Message)

			if tt.checkDetails {
				assert.Equal(t, tt.expectedDetail, cliErr.Details)
			}

			if tt.hasSuggestions {
				assert.NotEmpty(t, cliErr.Suggestions)
			}
		})
	}
}

func TestConvertError_WrappedErrors(t *testing.T) {
	// Test that wrapped errors are properly detected
	tests := []struct {
		name         string
		err          error
		expectedCode string
	}{
		{
			name:         "wrapped ticket not found",
			err:          fmt.Errorf("operation failed: %w", ticketerrors.ErrTicketNotFound),
			expectedCode: ErrTicketNotFound,
		},
		{
			name:         "deeply wrapped config error",
			err:          fmt.Errorf("initialization: %w", fmt.Errorf("loading config: %w", ticketerrors.ErrConfigNotFound)),
			expectedCode: ErrConfigNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertError(tt.err)

			cliErr, ok := result.(*CLIError)
			assert.True(t, ok)
			assert.Equal(t, tt.expectedCode, cliErr.Code)
		})
	}
}

func TestConvertError_PreservesContext(t *testing.T) {
	// Test that error context is preserved in details
	originalErr := ticketerrors.NewTicketErrorWithContext(
		"update",
		"test-123",
		fmt.Errorf("file not found"),
		"worktree", "sync",
	)

	result := ConvertError(originalErr)

	cliErr, ok := result.(*CLIError)
	assert.True(t, ok)
	assert.Equal(t, ErrTicketInvalid, cliErr.Code)
	assert.Contains(t, cliErr.Details, "file not found")
}

func TestConvertError_ConfigFieldSuggestion(t *testing.T) {
	// Test that config errors include field-specific suggestions
	configErr := ticketerrors.NewConfigError("git.timeout", "invalid-duration", errors.New("invalid duration"))

	result := ConvertError(configErr)

	cliErr, ok := result.(*CLIError)
	assert.True(t, ok)
	assert.Equal(t, ErrConfigInvalid, cliErr.Code)
	assert.Len(t, cliErr.Suggestions, 1)
	assert.Contains(t, cliErr.Suggestions[0], "git.timeout")
}

func TestEnhanceWorktreeGitError(t *testing.T) {
	tests := []struct {
		name               string
		gitErr             *ticketerrors.GitError
		expectedCode       string
		expectedMsg        string
		expectedSuggestion string
		shouldEnhance      bool
	}{
		{
			name:               "corrupted worktree",
			gitErr:             &ticketerrors.GitError{Op: "worktree", Err: fmt.Errorf("fatal: '/path/to/worktree' is not a working tree")},
			expectedCode:       ErrWorktreeRemoveFailed,
			expectedMsg:        "Worktree appears to be corrupted",
			expectedSuggestion: "Run 'git worktree prune'",
			shouldEnhance:      true,
		},
		{
			name:               "directory already exists",
			gitErr:             &ticketerrors.GitError{Op: "worktree", Err: fmt.Errorf("fatal: '/path/to/worktree' already exists")},
			expectedCode:       ErrWorktreeExists,
			expectedMsg:        "Worktree directory already exists",
			expectedSuggestion: "Remove the directory manually if it's no longer needed",
			shouldEnhance:      true,
		},
		{
			name:               "branch already checked out",
			gitErr:             &ticketerrors.GitError{Op: "worktree", Err: fmt.Errorf("fatal: 'feature-branch' is already checked out")},
			expectedCode:       ErrWorktreeExists,
			expectedMsg:        "Branch is already checked out in another worktree",
			expectedSuggestion: "git worktree list",
			shouldEnhance:      true,
		},
		{
			name:               "cannot create work tree dir",
			gitErr:             &ticketerrors.GitError{Op: "worktree", Err: fmt.Errorf("fatal: could not create work tree dir '/path/to/worktree'")},
			expectedCode:       ErrWorktreeCreateFailed,
			expectedMsg:        "Cannot create worktree directory",
			expectedSuggestion: "Check directory permissions",
			shouldEnhance:      true,
		},
		{
			name:               "invalid reference",
			gitErr:             &ticketerrors.GitError{Op: "worktree", Err: fmt.Errorf("fatal: invalid reference: refs/heads/invalid-branch")},
			expectedCode:       ErrWorktreeCreateFailed,
			expectedMsg:        "Invalid git reference for worktree",
			expectedSuggestion: "Ensure the branch name is valid",
			shouldEnhance:      true,
		},
		{
			name:               "permission denied",
			gitErr:             &ticketerrors.GitError{Op: "worktree", Err: fmt.Errorf("fatal: could not create directory: permission denied")},
			expectedCode:       ErrPermissionDenied,
			expectedMsg:        "Permission denied for worktree operation",
			expectedSuggestion: "Check file and directory permissions",
			shouldEnhance:      true,
		},
		{
			name:               "locked worktree",
			gitErr:             &ticketerrors.GitError{Op: "worktree", Err: fmt.Errorf("fatal: '/path/to/worktree' is locked")},
			expectedCode:       ErrWorktreeRemoveFailed,
			expectedMsg:        "Worktree is locked",
			expectedSuggestion: "Check if another process is using the worktree",
			shouldEnhance:      true,
		},
		{
			name:          "unrelated error",
			gitErr:        &ticketerrors.GitError{Op: "worktree", Err: fmt.Errorf("some other error")},
			shouldEnhance: false,
		},
		{
			name:          "nil error",
			gitErr:        nil,
			shouldEnhance: false,
		},
		{
			name:          "nil wrapped error",
			gitErr:        &ticketerrors.GitError{Op: "worktree", Err: nil},
			shouldEnhance: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := enhanceWorktreeGitError(tt.gitErr)

			if !tt.shouldEnhance {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedCode, result.Code)
			assert.Equal(t, tt.expectedMsg, result.Message)
			assert.NotEmpty(t, result.Suggestions)

			// Check that at least one suggestion contains the expected text
			found := false
			for _, suggestion := range result.Suggestions {
				if strings.Contains(suggestion, tt.expectedSuggestion) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected suggestion containing '%s' not found in %v", tt.expectedSuggestion, result.Suggestions)
			}
		})
	}
}

func TestEnhanceWorktreeError(t *testing.T) {
	tests := []struct {
		name               string
		worktreeErr        *ticketerrors.WorktreeError
		expectedCode       string
		expectedMsg        string
		expectedSuggestion string
		shouldEnhance      bool
	}{
		{
			name:               "corrupted worktree",
			worktreeErr:        &ticketerrors.WorktreeError{Op: "remove", Path: "test-123", Err: fmt.Errorf("fatal: '/path/to/worktree' is not a working tree")},
			expectedCode:       ErrWorktreeRemoveFailed,
			expectedMsg:        "Worktree appears to be corrupted",
			expectedSuggestion: "Run 'git worktree prune'",
			shouldEnhance:      true,
		},
		{
			name:               "directory already exists",
			worktreeErr:        &ticketerrors.WorktreeError{Op: "create", Path: "test-123", Err: fmt.Errorf("fatal: '/path/to/worktree' already exists")},
			expectedCode:       ErrWorktreeExists,
			expectedMsg:        "Worktree directory already exists",
			expectedSuggestion: "Remove the directory manually if it's no longer needed",
			shouldEnhance:      true,
		},
		{
			name:               "branch already checked out",
			worktreeErr:        &ticketerrors.WorktreeError{Op: "create", Path: "test-123", Err: fmt.Errorf("fatal: 'feature-branch' is already checked out")},
			expectedCode:       ErrWorktreeExists,
			expectedMsg:        "Branch is already checked out in another worktree",
			expectedSuggestion: "git worktree list",
			shouldEnhance:      true,
		},
		{
			name:               "cannot create work tree dir",
			worktreeErr:        &ticketerrors.WorktreeError{Op: "create", Path: "test-123", Err: fmt.Errorf("fatal: could not create work tree dir '/path/to/worktree'")},
			expectedCode:       ErrWorktreeCreateFailed,
			expectedMsg:        "Cannot create worktree directory",
			expectedSuggestion: "Ensure you have enough disk space",
			shouldEnhance:      true,
		},
		{
			name:               "permission denied",
			worktreeErr:        &ticketerrors.WorktreeError{Op: "create", Path: "test-123", Err: fmt.Errorf("fatal: could not create directory: permission denied")},
			expectedCode:       ErrPermissionDenied,
			expectedMsg:        "Permission denied for worktree operation",
			expectedSuggestion: "Check file and directory permissions",
			shouldEnhance:      true,
		},
		{
			name:          "unrelated error",
			worktreeErr:   &ticketerrors.WorktreeError{Op: "create", Path: "test-123", Err: fmt.Errorf("some other error")},
			shouldEnhance: false,
		},
		{
			name:          "nil error",
			worktreeErr:   nil,
			shouldEnhance: false,
		},
		{
			name:          "nil wrapped error",
			worktreeErr:   &ticketerrors.WorktreeError{Op: "create", Path: "test-123", Err: nil},
			shouldEnhance: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := enhanceWorktreeError(tt.worktreeErr)

			if !tt.shouldEnhance {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedCode, result.Code)
			assert.Equal(t, tt.expectedMsg, result.Message)
			assert.NotEmpty(t, result.Suggestions)

			// Check that at least one suggestion contains the expected text
			found := false
			for _, suggestion := range result.Suggestions {
				if strings.Contains(suggestion, tt.expectedSuggestion) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected suggestion containing '%s' not found in %v", tt.expectedSuggestion, result.Suggestions)
			}
		})
	}
}

func TestConvertError_WorktreeEnhancement(t *testing.T) {
	// Test that worktree errors are enhanced through ConvertError
	tests := []struct {
		name          string
		err           error
		expectedCode  string
		expectedMsg   string
		shouldEnhance bool
	}{
		{
			name:          "git error with worktree operation",
			err:           ticketerrors.NewGitError("worktree", "", fmt.Errorf("fatal: '/path' is not a working tree")),
			expectedCode:  ErrWorktreeRemoveFailed,
			expectedMsg:   "Worktree appears to be corrupted",
			shouldEnhance: true,
		},
		{
			name:          "worktree error with git pattern",
			err:           ticketerrors.NewWorktreeError("create", "test-123", fmt.Errorf("fatal: 'branch' is already checked out")),
			expectedCode:  ErrWorktreeExists,
			expectedMsg:   "Branch is already checked out in another worktree",
			shouldEnhance: true,
		},
		{
			name:          "git error without worktree operation",
			err:           ticketerrors.NewGitError("checkout", "main", fmt.Errorf("failed")),
			expectedCode:  ErrGitMergeFailed,
			expectedMsg:   "Git operation failed: checkout",
			shouldEnhance: false,
		},
		{
			name:          "worktree error without known pattern",
			err:           ticketerrors.NewWorktreeError("create", "test-123", fmt.Errorf("unknown error")),
			expectedCode:  ErrWorktreeCreateFailed,
			expectedMsg:   "Worktree operation failed: create",
			shouldEnhance: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertError(tt.err)

			cliErr, ok := result.(*CLIError)
			assert.True(t, ok)
			assert.Equal(t, tt.expectedCode, cliErr.Code)
			assert.Equal(t, tt.expectedMsg, cliErr.Message)

			if tt.shouldEnhance {
				assert.NotEmpty(t, cliErr.Suggestions, "Enhanced errors should have suggestions")
			}
		})
	}
}
