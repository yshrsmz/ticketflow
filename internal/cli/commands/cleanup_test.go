package commands

import (
	"context"
	flag "github.com/spf13/pflag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanupCommand_Interface(t *testing.T) {
	cmd := NewCleanupCommand()

	assert.Equal(t, "cleanup", cmd.Name())
	assert.Nil(t, cmd.Aliases())
	assert.Equal(t, "Clean up worktrees and branches", cmd.Description())
	assert.Equal(t, "cleanup [--dry-run] [--force] [--format text|json] [<ticket-id>]", cmd.Usage())
}

func TestCleanupCommand_SetupFlags(t *testing.T) {
	cmd := &CleanupCommand{}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	flags := cmd.SetupFlags(fs)

	// Verify flags is of correct type
	_, ok := flags.(*cleanupFlags)
	assert.True(t, ok, "SetupFlags should return *cleanupFlags")

	// Verify flags are registered
	assert.NotNil(t, fs.Lookup("dry-run"))
	assert.NotNil(t, fs.Lookup("force"))
	// Phase 1: With pflag, shorthand is not a separate flag - use ShorthandLookup
	assert.NotNil(t, fs.ShorthandLookup("f"))
	assert.NotNil(t, fs.Lookup("format"))
	// Phase 1: With pflag, shorthand is not a separate flag - use ShorthandLookup
	assert.NotNil(t, fs.ShorthandLookup("o"))
}

func TestCleanupCommand_Validate(t *testing.T) {
	tests := []struct {
		name        string
		flags       interface{}
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid auto-cleanup no arguments",
			flags:   &cleanupFlags{format: StringFlag{Long: "text"}},
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "valid ticket cleanup with ID",
			flags:   &cleanupFlags{format: StringFlag{Long: "text"}},
			args:    []string{"ticket-123"},
			wantErr: false,
		},
		{
			name:    "valid with dry-run flag",
			flags:   &cleanupFlags{dryRun: true, format: StringFlag{Long: "text"}},
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "valid with force flag",
			flags:   &cleanupFlags{force: BoolFlag{Long: true}, format: StringFlag{Long: "text"}},
			args:    []string{"ticket-123"},
			wantErr: false,
		},
		{
			name:    "valid with json format",
			flags:   &cleanupFlags{format: StringFlag{Long: "json"}},
			args:    []string{},
			wantErr: false,
		},
		{
			name:        "invalid format",
			flags:       &cleanupFlags{format: StringFlag{Long: "yaml"}},
			args:        []string{},
			wantErr:     true,
			errContains: "invalid format",
		},
		{
			name:        "dry-run with ticket ID not allowed",
			flags:       &cleanupFlags{dryRun: true, format: StringFlag{Long: "text"}},
			args:        []string{"ticket-123"},
			wantErr:     true,
			errContains: "--dry-run cannot be used when cleaning up a specific ticket",
		},
		{
			name:        "too many arguments",
			flags:       &cleanupFlags{format: StringFlag{Long: "text"}},
			args:        []string{"ticket-123", "extra"},
			wantErr:     true,
			errContains: "unexpected arguments after ticket ID",
		},
		{
			name:        "wrong flags type",
			flags:       "invalid",
			args:        []string{},
			wantErr:     true,
			errContains: "invalid flags type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &CleanupCommand{}
			err := cmd.Validate(tt.flags, tt.args)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)

				// Verify args were stored if valid
				if f, ok := tt.flags.(*cleanupFlags); ok {
					assert.Equal(t, tt.args, f.args)
				}
			}
		})
	}
}

func TestCleanupCommand_Execute_Errors(t *testing.T) {
	tests := []struct {
		name        string
		setupCtx    func() context.Context
		flags       interface{}
		args        []string
		errContains string
	}{
		{
			name: "context cancelled",
			setupCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			flags:       &cleanupFlags{format: StringFlag{Long: "text"}},
			args:        []string{},
			errContains: "context canceled",
		},
		{
			name:        "invalid flags type",
			setupCtx:    context.Background,
			flags:       "invalid",
			args:        []string{},
			errContains: "invalid flags type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &CleanupCommand{}
			ctx := tt.setupCtx()
			err := cmd.Execute(ctx, tt.flags, tt.args)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}
