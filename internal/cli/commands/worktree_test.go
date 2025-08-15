package commands

import (
	"context"
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorktreeCommand_Interface(t *testing.T) {
	cmd := NewWorktreeCommand()

	assert.Equal(t, "worktree", cmd.Name())
	assert.Nil(t, cmd.Aliases())
	assert.Equal(t, "Manage git worktrees associated with tickets", cmd.Description())
	assert.Equal(t, "worktree <subcommand> [options]", cmd.Usage())
}

func TestWorktreeCommand_SetupFlags(t *testing.T) {
	cmd := &WorktreeCommand{}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	flags := cmd.SetupFlags(fs)

	// No flags for parent command
	assert.Nil(t, flags)
}

func TestWorktreeCommand_Validate(t *testing.T) {
	cmd := &WorktreeCommand{}

	// Validate doesn't perform validation for parent command
	err := cmd.Validate(nil, []string{})
	assert.NoError(t, err)

	err = cmd.Validate(nil, []string{"list"})
	assert.NoError(t, err)

	err = cmd.Validate(nil, []string{"unknown"})
	assert.NoError(t, err)
}

func TestWorktreeCommand_Execute(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "no subcommand shows usage",
			args:    []string{},
			wantErr: false,
		},
		{
			name:        "unknown subcommand",
			args:        []string{"unknown"},
			wantErr:     true,
			errContains: "unknown worktree subcommand",
		},
		{
			name:        "invalid subcommand",
			args:        []string{"delete"},
			wantErr:     true,
			errContains: "unknown worktree subcommand",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewWorktreeCommand()
			err := cmd.Execute(context.Background(), nil, tt.args)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
