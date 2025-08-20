package commands

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorktreeListCommand_Interface(t *testing.T) {
	t.Parallel()
	cmd := NewWorktreeListCommand()

	assert.Equal(t, "list", cmd.Name())
	assert.Nil(t, cmd.Aliases())
	assert.Equal(t, "List all worktrees", cmd.Description())
	assert.Equal(t, "worktree list [--format json]", cmd.Usage())
}

func TestWorktreeListCommand_SetupFlags(t *testing.T) {
	t.Parallel()
	cmd := &WorktreeListCommand{}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	flags := cmd.SetupFlags(fs)

	// Verify flags is of correct type
	_, ok := flags.(*worktreeListFlags)
	assert.True(t, ok, "SetupFlags should return *worktreeListFlags")

	// Verify flags are registered
	assert.NotNil(t, fs.Lookup("format"))
	assert.NotNil(t, fs.Lookup("o"))
}

func TestWorktreeListCommand_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		flags       interface{}
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid text format",
			flags:   &worktreeListFlags{format: FormatText},
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "valid json format",
			flags:   &worktreeListFlags{format: FormatJSON},
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "valid short form json",
			flags:   &worktreeListFlags{format: FormatText, formatShort: FormatJSON},
			args:    []string{},
			wantErr: false,
		},
		{
			name:        "invalid format",
			flags:       &worktreeListFlags{format: StringFlag{Long: "yaml"}},
			args:        []string{},
			wantErr:     true,
			errContains: "invalid format",
		},
		{
			name:        "unexpected arguments",
			flags:       &worktreeListFlags{format: FormatText},
			args:        []string{"extra"},
			wantErr:     true,
			errContains: "takes no arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &WorktreeListCommand{}
			err := cmd.Validate(tt.flags, tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWorktreeListCommand_ValidateFormatOverride(t *testing.T) {
	t.Parallel()
	cmd := &WorktreeListCommand{}

	// Test that short form overrides long form
	flags := &worktreeListFlags{
		format:      FormatText,
		formatShort: FormatJSON,
	}

	err := cmd.Validate(flags, []string{})
	assert.NoError(t, err)
	assert.Equal(t, FormatJSON, flags.format)
}
