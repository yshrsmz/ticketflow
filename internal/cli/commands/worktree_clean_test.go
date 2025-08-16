package commands

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorktreeCleanCommand_Interface(t *testing.T) {
	t.Parallel()
	cmd := NewWorktreeCleanCommand()

	assert.Equal(t, "clean", cmd.Name())
	assert.Nil(t, cmd.Aliases())
	assert.Equal(t, "Remove orphaned worktrees", cmd.Description())
	assert.Equal(t, "worktree clean [--format text|json]", cmd.Usage())
}

func TestWorktreeCleanCommand_SetupFlags(t *testing.T) {
	t.Parallel()
	cmd := &WorktreeCleanCommand{}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	flags := cmd.SetupFlags(fs)

	// Now returns worktreeCleanFlags with format support
	assert.NotNil(t, flags)
	assert.IsType(t, &worktreeCleanFlags{}, flags)
	
	// Verify format flags were registered
	assert.NotNil(t, fs.Lookup("format"))
	assert.NotNil(t, fs.Lookup("o"))
}

func TestWorktreeCleanCommand_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		flags       interface{}
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid no arguments with text format",
			flags:   &worktreeCleanFlags{format: "text"},
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "valid no arguments with json format",
			flags:   &worktreeCleanFlags{format: "json"},
			args:    []string{},
			wantErr: false,
		},
		{
			name:        "invalid format",
			flags:       &worktreeCleanFlags{format: "yaml"},
			args:        []string{},
			wantErr:     true,
			errContains: "invalid format",
		},
		{
			name:        "unexpected arguments",
			flags:       &worktreeCleanFlags{format: "text"},
			args:        []string{"extra"},
			wantErr:     true,
			errContains: "takes no arguments",
		},
		{
			name:        "multiple unexpected arguments",
			flags:       &worktreeCleanFlags{format: "text"},
			args:        []string{"extra", "args"},
			wantErr:     true,
			errContains: "takes no arguments",
		},
		{
			name:        "invalid flags type",
			flags:       "invalid",
			args:        []string{},
			wantErr:     true,
			errContains: "invalid flags type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &WorktreeCleanCommand{}
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
