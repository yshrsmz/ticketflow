package commands

import (
	"context"
	"errors"
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yshrsmz/ticketflow/internal/cli"
)

func TestStartCommand_Name(t *testing.T) {
	cmd := NewStartCommand()
	assert.Equal(t, "start", cmd.Name())
}

func TestStartCommand_Aliases(t *testing.T) {
	cmd := NewStartCommand()
	assert.Nil(t, cmd.Aliases())
}

func TestStartCommand_Description(t *testing.T) {
	cmd := NewStartCommand()
	assert.Equal(t, "Start work on a ticket", cmd.Description())
}

func TestStartCommand_Usage(t *testing.T) {
	cmd := NewStartCommand()
	assert.Equal(t, "start [--force] [--format text|json] <ticket-id>", cmd.Usage())
}

func TestStartCommand_SetupFlags(t *testing.T) {
	cmd := NewStartCommand()
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	flags := cmd.SetupFlags(fs)

	assert.NotNil(t, flags)
	sf, ok := flags.(*startFlags)
	assert.True(t, ok)
	assert.NotNil(t, sf)

	// Check that flags are registered
	forceFlag := fs.Lookup("force")
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "false", forceFlag.DefValue)

	formatFlag := fs.Lookup("format")
	assert.NotNil(t, formatFlag)
	assert.Equal(t, "text", formatFlag.DefValue)

	// Check short forms
	fFlag := fs.Lookup("f")
	assert.NotNil(t, fFlag)
	assert.Equal(t, "false", fFlag.DefValue)

	oFlag := fs.Lookup("o")
	assert.NotNil(t, oFlag)
	assert.Equal(t, "", oFlag.DefValue)
}

func TestStartCommand_Validate(t *testing.T) {
	tests := []struct {
		name          string
		flags         interface{}
		args          []string
		setupFlags    func(*startFlags)
		expectedError string
	}{
		{
			name: "valid with ticket ID",
			flags: &startFlags{
				format: StringFlag{Long: "text"},
			},
			args:          []string{"250813-123456-test"},
			expectedError: "",
		},
		{
			name: "valid with force flag",
			flags: &startFlags{
				force:  BoolFlag{Long: true},
				format: StringFlag{Long: "text"},
			},
			args:          []string{"250813-123456-test"},
			expectedError: "",
		},
		{
			name: "valid with json format",
			flags: &startFlags{
				format: StringFlag{Long: "json"},
			},
			args:          []string{"250813-123456-test"},
			expectedError: "",
		},
		// Note: Testing flag precedence is not possible in unit tests
		// since we're directly setting values without going through flag parsing
		{
			name:          "missing ticket ID",
			flags:         &startFlags{},
			args:          []string{},
			expectedError: "missing ticket argument",
		},
		{
			name: "too many arguments",
			flags: &startFlags{
				format: StringFlag{Long: "text"},
			},
			args:          []string{"ticket1", "ticket2"},
			expectedError: "unexpected arguments after ticket ID: [ticket2]",
		},
		{
			name: "invalid format",
			flags: &startFlags{
				format: StringFlag{Long: "yaml"},
			},
			args:          []string{"250813-123456-test"},
			expectedError: "invalid format: \"yaml\"",
		},
		{
			name:          "wrong flags type",
			flags:         struct{}{},
			args:          []string{"250813-123456-test"},
			expectedError: "invalid flags type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewStartCommand()

			err := cmd.Validate(tt.flags, tt.args)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				// No need to verify normalization with new flag utilities
			}
		})
	}
}

func TestStartCommand_Execute(t *testing.T) {
	tests := []struct {
		name          string
		flags         interface{}
		args          []string
		mockStartFunc func(ctx context.Context, ticketID string, force bool, format cli.OutputFormat) error
		expectedError string
		ctxCancelled  bool
		appError      error
	}{
		{
			name: "successful start with text format",
			flags: &startFlags{
				force:  BoolFlag{Long: false},
				format: StringFlag{Long: "text"},
			},
			args: []string{"250813-123456-test"},
			mockStartFunc: func(ctx context.Context, ticketID string, force bool, format cli.OutputFormat) error {
				assert.Equal(t, "250813-123456-test", ticketID)
				assert.False(t, force)
				assert.Equal(t, cli.FormatText, format)
				return nil
			},
			expectedError: "",
		},
		{
			name: "successful start with json format",
			flags: &startFlags{
				force:  BoolFlag{Long: false},
				format: StringFlag{Long: "json"},
			},
			args: []string{"250813-123456-test"},
			mockStartFunc: func(ctx context.Context, ticketID string, force bool, format cli.OutputFormat) error {
				assert.Equal(t, "250813-123456-test", ticketID)
				assert.False(t, force)
				assert.Equal(t, cli.FormatJSON, format)
				return nil
			},
			expectedError: "",
		},
		{
			name: "successful start with force",
			flags: &startFlags{
				force:  BoolFlag{Long: true},
				format: StringFlag{Long: "text"},
			},
			args: []string{"250813-123456-test"},
			mockStartFunc: func(ctx context.Context, ticketID string, force bool, format cli.OutputFormat) error {
				assert.Equal(t, "250813-123456-test", ticketID)
				assert.True(t, force)
				assert.Equal(t, cli.FormatText, format)
				return nil
			},
			expectedError: "",
		},
		{
			name: "error from StartTicket",
			flags: &startFlags{
				force:  BoolFlag{Long: false},
				format: StringFlag{Long: "text"},
			},
			args: []string{"250813-123456-test"},
			mockStartFunc: func(ctx context.Context, ticketID string, force bool, format cli.OutputFormat) error {
				return errors.New("ticket not found")
			},
			expectedError: "ticket not found",
		},
		{
			name:          "wrong flags type",
			flags:         struct{}{},
			args:          []string{"250813-123456-test"},
			expectedError: "invalid flags type",
		},
		{
			name: "context cancelled",
			flags: &startFlags{
				force:  BoolFlag{Long: false},
				format: StringFlag{Long: "text"},
			},
			args:          []string{"250813-123456-test"},
			expectedError: "context canceled",
			ctxCancelled:  true,
		},
		{
			name: "app creation error",
			flags: &startFlags{
				force:  BoolFlag{Long: false},
				format: StringFlag{Long: "text"},
			},
			args:          []string{"250813-123456-test"},
			appError:      errors.New("failed to create app"),
			expectedError: "failed to create app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test context
			ctx := context.Background()
			if tt.ctxCancelled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel() // Cancel immediately
			}

			// Skip tests that require actual App interaction (not context cancellation or type validation)
			// These scenarios are better covered by integration tests
			// The app creation error test also needs to be skipped as we can't mock cli.NewApp
			if !tt.ctxCancelled && tt.expectedError != "invalid flags type" {
				t.Skipf("Skipping '%s': requires actual App interaction - covered by integration tests", tt.name)
			}

			cmd := NewStartCommand()
			err := cmd.Execute(ctx, tt.flags, tt.args)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestStartFlags_Normalize has been removed since normalize() is no longer needed.
// The flag utilities now handle precedence automatically through the Value() methods.
