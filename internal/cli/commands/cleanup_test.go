package commands

import (
	"context"
	"flag"
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
	assert.NotNil(t, fs.Lookup("f"))
	assert.NotNil(t, fs.Lookup("format"))
	assert.NotNil(t, fs.Lookup("o"))
}

func TestCleanupCommand_Validate(t *testing.T) {
	tests := []struct {
		name      string
		flags     interface{}
		args      []string
		wantErr   bool
		errContains string
	}{
		{
			name:    "valid auto-cleanup no arguments",
			flags:   &cleanupFlags{format: "text"},
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "valid ticket cleanup with ID",
			flags:   &cleanupFlags{format: "text"},
			args:    []string{"ticket-123"},
			wantErr: false,
		},
		{
			name:    "valid with dry-run flag",
			flags:   &cleanupFlags{dryRun: true, format: "text"},
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "valid with force flag",
			flags:   &cleanupFlags{force: true, format: "text"},
			args:    []string{"ticket-123"},
			wantErr: false,
		},
		{
			name:    "valid with json format",
			flags:   &cleanupFlags{format: "json"},
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "invalid format",
			flags:   &cleanupFlags{format: "yaml"},
			args:    []string{},
			wantErr: true,
			errContains: "invalid format",
		},
		{
			name:    "dry-run with ticket ID not allowed",
			flags:   &cleanupFlags{dryRun: true, format: "text"},
			args:    []string{"ticket-123"},
			wantErr: true,
			errContains: "--dry-run cannot be used when cleaning up a specific ticket",
		},
		{
			name:    "too many arguments",
			flags:   &cleanupFlags{format: "text"},
			args:    []string{"ticket-123", "extra"},
			wantErr: true,
			errContains: "unexpected arguments after ticket ID",
		},
		{
			name:    "wrong flags type",
			flags:   "invalid",
			args:    []string{},
			wantErr: true,
			errContains: "invalid flags type",
		},
		{
			name:    "force short form takes precedence",
			flags:   &cleanupFlags{force: false, forceShort: true, format: "text"},
			args:    []string{"ticket-123"},
			wantErr: false,
		},
		{
			name:    "format short form takes precedence",
			flags:   &cleanupFlags{format: "text", formatShort: "json"},
			args:    []string{},
			wantErr: false,
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
				
				// Verify args were stored and flags normalized if valid
				if f, ok := tt.flags.(*cleanupFlags); ok {
					assert.Equal(t, tt.args, f.args)
					// Check normalization worked
					if tt.name == "force short form takes precedence" {
						assert.True(t, f.force)
					}
					if tt.name == "format short form takes precedence" {
						assert.Equal(t, "json", f.format)
					}
				}
			}
		})
	}
}

func TestCleanupFlags_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		flags    *cleanupFlags
		expected *cleanupFlags
	}{
		{
			name: "no short forms",
			flags: &cleanupFlags{
				force:  true,
				format: "json",
			},
			expected: &cleanupFlags{
				force:  true,
				format: "json",
			},
		},
		{
			name: "force short form overrides via OR",
			flags: &cleanupFlags{
				force:      false,
				forceShort: true,
				format:     "text",
			},
			expected: &cleanupFlags{
				force:      true,
				forceShort: true,
				format:     "text",
			},
		},
		{
			name: "format short form overrides",
			flags: &cleanupFlags{
				format:      "text",
				formatShort: "json",
			},
			expected: &cleanupFlags{
				format:      "json",
				formatShort: "json",
			},
		},
		{
			name: "both short forms override",
			flags: &cleanupFlags{
				force:       false,
				forceShort:  true,
				format:      "text",
				formatShort: "json",
			},
			expected: &cleanupFlags{
				force:       true,
				forceShort:  true,
				format:      "json",
				formatShort: "json",
			},
		},
		{
			name: "both force flags true",
			flags: &cleanupFlags{
				force:      true,
				forceShort: true,
				format:     "text",
			},
			expected: &cleanupFlags{
				force:      true,
				forceShort: true,
				format:     "text",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.flags.normalize()
			assert.Equal(t, tt.expected, tt.flags)
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
			flags:       &cleanupFlags{format: "text"},
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