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
	assert.Equal(t, "worktree clean", cmd.Usage())
}

func TestWorktreeCleanCommand_SetupFlags(t *testing.T) {
	t.Parallel()
	cmd := &WorktreeCleanCommand{}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	flags := cmd.SetupFlags(fs)

	// No flags for clean command
	assert.Nil(t, flags)
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
			name:    "valid no arguments",
			flags:   nil,
			args:    []string{},
			wantErr: false,
		},
		{
			name:        "unexpected arguments",
			flags:       nil,
			args:        []string{"extra"},
			wantErr:     true,
			errContains: "takes no arguments",
		},
		{
			name:        "multiple unexpected arguments",
			flags:       nil,
			args:        []string{"extra", "args"},
			wantErr:     true,
			errContains: "takes no arguments",
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
