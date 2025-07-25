package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewError(t *testing.T) {
	err := NewError(
		ErrNotGitRepo,
		"Not in a git repository",
		"Current directory is not part of a git repository",
		[]string{"Navigate to project root", "Run git init"},
	)

	assert.Equal(t, ErrNotGitRepo, err.Code)
	assert.Equal(t, "Not in a git repository", err.Message)
	assert.Equal(t, "Current directory is not part of a git repository", err.Details)
	assert.Len(t, err.Suggestions, 2)
}

func TestCLIError_Error(t *testing.T) {
	err := &CLIError{
		Code:    ErrTicketNotFound,
		Message: "Ticket not found",
	}

	assert.Equal(t, "Ticket not found", err.Error())
}