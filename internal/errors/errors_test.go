package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTicketError(t *testing.T) {
	underlyingErr := errors.New("underlying error")
	err := NewTicketError("create", "ticket-123", underlyingErr)

	ticketErr, ok := err.(*TicketError)
	assert.True(t, ok)
	assert.Equal(t, "create", ticketErr.Op)
	assert.Equal(t, "ticket-123", ticketErr.TicketID)
	assert.Equal(t, underlyingErr, ticketErr.Err)
	assert.Contains(t, err.Error(), "create ticket ticket-123")
	assert.Equal(t, underlyingErr, ticketErr.Unwrap())
}

func TestTicketErrorWithContext(t *testing.T) {
	underlyingErr := errors.New("underlying error")
	err := NewTicketErrorWithContext("create", "ticket-123", underlyingErr, "worktree", "init")

	ticketErr, ok := err.(*TicketError)
	assert.True(t, ok)
	assert.Equal(t, "create", ticketErr.Op)
	assert.Equal(t, "ticket-123", ticketErr.TicketID)
	assert.Equal(t, underlyingErr, ticketErr.Err)
	assert.Equal(t, []string{"worktree", "init"}, ticketErr.Context)
	assert.Contains(t, err.Error(), "worktree > init > create ticket ticket-123")
	assert.Equal(t, underlyingErr, ticketErr.Unwrap())
}

func TestGitError(t *testing.T) {
	underlyingErr := errors.New("git command failed")
	err := NewGitError("push", "feature-branch", underlyingErr)

	gitErr, ok := err.(*GitError)
	assert.True(t, ok)
	assert.Equal(t, "push", gitErr.Op)
	assert.Equal(t, "feature-branch", gitErr.Branch)
	assert.Equal(t, underlyingErr, gitErr.Err)
	assert.Contains(t, err.Error(), "git push on branch feature-branch")
	assert.Equal(t, underlyingErr, gitErr.Unwrap())
}

func TestWorktreeError(t *testing.T) {
	underlyingErr := errors.New("directory exists")
	err := NewWorktreeError("create", "/path/to/worktree", underlyingErr)

	worktreeErr, ok := err.(*WorktreeError)
	assert.True(t, ok)
	assert.Equal(t, "create", worktreeErr.Op)
	assert.Equal(t, "/path/to/worktree", worktreeErr.Path)
	assert.Equal(t, underlyingErr, worktreeErr.Err)
	assert.Contains(t, err.Error(), "worktree create at /path/to/worktree")
	assert.Equal(t, underlyingErr, worktreeErr.Unwrap())
}

func TestConfigError(t *testing.T) {
	underlyingErr := errors.New("invalid value")
	err := NewConfigError("output.format", "xml", underlyingErr)

	configErr, ok := err.(*ConfigError)
	assert.True(t, ok)
	assert.Equal(t, "output.format", configErr.Field)
	assert.Equal(t, "xml", configErr.Value)
	assert.Equal(t, underlyingErr, configErr.Err)
	assert.Contains(t, err.Error(), "config field output.format with value \"xml\"")
	assert.Equal(t, underlyingErr, configErr.Unwrap())
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "ticket not found",
			err:  ErrTicketNotFound,
			want: true,
		},
		{
			name: "worktree not found",
			err:  ErrWorktreeNotFound,
			want: true,
		},
		{
			name: "branch not found",
			err:  ErrBranchNotFound,
			want: true,
		},
		{
			name: "config not found",
			err:  ErrConfigNotFound,
			want: true,
		},
		{
			name: "other error",
			err:  errors.New("other error"),
			want: false,
		},
		{
			name: "wrapped ticket not found",
			err:  fmt.Errorf("failed to get ticket: %w", ErrTicketNotFound),
			want: true,
		},
		{
			name: "nested wrapped error",
			err:  fmt.Errorf("operation failed: %w", fmt.Errorf("ticket lookup: %w", ErrTicketNotFound)),
			want: true,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "ticket exists error",
			err:  ErrTicketExists,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsNotFound(tt.err))
		})
	}
}

func TestIsAlreadyExists(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "ticket exists",
			err:  ErrTicketExists,
			want: true,
		},
		{
			name: "branch exists",
			err:  ErrBranchExists,
			want: true,
		},
		{
			name: "worktree exists",
			err:  ErrWorktreeExists,
			want: true,
		},
		{
			name: "ticket already started",
			err:  ErrTicketAlreadyStarted,
			want: true,
		},
		{
			name: "ticket already closed",
			err:  ErrTicketAlreadyClosed,
			want: true,
		},
		{
			name: "other error",
			err:  errors.New("other error"),
			want: false,
		},
		{
			name: "wrapped exists error",
			err:  fmt.Errorf("creation failed: %w", ErrTicketExists),
			want: true,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "not found error",
			err:  ErrTicketNotFound,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsAlreadyExists(tt.err))
		})
	}
}

func TestConstructorValidation(t *testing.T) {
	t.Run("NewTicketError validation", func(t *testing.T) {
		// Empty operation
		err := NewTicketError("", "ticket-123", errors.New("test"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "operation cannot be empty")

		// Nil error
		err = NewTicketError("create", "ticket-123", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "underlying error cannot be nil")

		// Valid
		err = NewTicketError("create", "ticket-123", errors.New("test"))
		assert.NotNil(t, err)
		_, ok := err.(*TicketError)
		assert.True(t, ok)
	})

	t.Run("NewGitError validation", func(t *testing.T) {
		// Empty operation
		err := NewGitError("", "main", errors.New("test"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "operation cannot be empty")

		// Nil error
		err = NewGitError("push", "main", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "underlying error cannot be nil")

		// Valid
		err = NewGitError("push", "main", errors.New("test"))
		assert.NotNil(t, err)
		_, ok := err.(*GitError)
		assert.True(t, ok)
	})

	t.Run("NewWorktreeError validation", func(t *testing.T) {
		// Empty operation
		err := NewWorktreeError("", "/path", errors.New("test"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "operation cannot be empty")

		// Nil error
		err = NewWorktreeError("create", "/path", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "underlying error cannot be nil")

		// Valid
		err = NewWorktreeError("create", "/path", errors.New("test"))
		assert.NotNil(t, err)
		_, ok := err.(*WorktreeError)
		assert.True(t, ok)
	})

	t.Run("NewConfigError validation", func(t *testing.T) {
		// Nil error
		err := NewConfigError("field", "value", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "underlying error cannot be nil")

		// Valid
		err = NewConfigError("field", "value", errors.New("test"))
		assert.NotNil(t, err)
		_, ok := err.(*ConfigError)
		assert.True(t, ok)

		// Empty field is allowed
		err = NewConfigError("", "value", errors.New("test"))
		assert.NotNil(t, err)
		_, ok = err.(*ConfigError)
		assert.True(t, ok)
	})
}

func TestErrorFormatting(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		contains string
	}{
		{
			name:     "ticket error with ID",
			err:      NewTicketError("update", "123-test", errors.New("failed")),
			contains: "update ticket 123-test: failed",
		},
		{
			name:     "ticket error without ID",
			err:      NewTicketError("list", "", errors.New("failed")),
			contains: "list ticket: failed",
		},
		{
			name:     "git error with branch",
			err:      NewGitError("checkout", "main", errors.New("failed")),
			contains: "git checkout on branch main: failed",
		},
		{
			name:     "git error without branch",
			err:      NewGitError("status", "", errors.New("failed")),
			contains: "git status: failed",
		},
		{
			name:     "worktree error with path",
			err:      NewWorktreeError("remove", "/tmp/wt", errors.New("failed")),
			contains: "worktree remove at /tmp/wt: failed",
		},
		{
			name:     "worktree error without path",
			err:      NewWorktreeError("list", "", errors.New("failed")),
			contains: "worktree list: failed",
		},
		{
			name:     "config error with field and value",
			err:      NewConfigError("tickets.dir", "/invalid", errors.New("failed")),
			contains: "config field tickets.dir with value \"/invalid\": failed",
		},
		{
			name:     "config error with field only",
			err:      NewConfigError("git.branch", "", errors.New("failed")),
			contains: "config field git.branch: failed",
		},
		{
			name:     "config error without field",
			err:      NewConfigError("", "", errors.New("failed")),
			contains: "config: failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, tt.err.Error(), tt.contains)
		})
	}
}

