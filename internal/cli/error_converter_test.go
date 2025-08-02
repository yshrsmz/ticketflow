package cli

import (
	"errors"
	"fmt"
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