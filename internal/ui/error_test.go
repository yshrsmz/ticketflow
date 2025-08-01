package ui

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitCommandError(t *testing.T) {
	tests := []struct {
		name     string
		commands []string
		want     string
	}{
		{
			name:     "single_failed_command",
			commands: []string{"git fetch (exit status 1)"},
			want:     "some initialization commands failed: git fetch (exit status 1)",
		},
		{
			name:     "multiple_failed_commands",
			commands: []string{"git fetch (exit status 1)", "npm install (timeout)"},
			want:     "some initialization commands failed: git fetch (exit status 1), npm install (timeout)",
		},
		{
			name:     "empty_commands",
			commands: []string{},
			want:     "some initialization commands failed: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewInitCommandError(tt.commands)
			assert.Equal(t, tt.want, err.Error())
			assert.Equal(t, tt.commands, err.FailedCommands)
		})
	}
}

func TestIsInitCommandError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "init_command_error",
			err:  NewInitCommandError([]string{"git fetch"}),
			want: true,
		},
		{
			name: "regular_error",
			err:  errors.New("some other error"),
			want: false,
		},
		{
			name: "nil_error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInitCommandError(tt.err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestInitCommandError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("underlying error")
	err := &InitCommandError{
		FailedCommands: []string{"cmd"},
		underlying:     underlyingErr,
	}

	assert.Equal(t, underlyingErr, err.Unwrap())
}
