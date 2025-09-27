package commands

import (
	"context"
	flag "github.com/spf13/pflag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/command"
)

func TestCloseCommand_Interface(t *testing.T) {
	t.Parallel()

	cmd := NewCloseCommand()

	// Verify it implements the Command interface
	var _ = command.Command(cmd)

	// Test Name
	assert.Equal(t, "close", cmd.Name())

	// Test Aliases
	assert.Nil(t, cmd.Aliases())

	// Test Description
	assert.Equal(t, "Close a ticket", cmd.Description())

	// Test Usage
	assert.Equal(t, "close [--force] [--reason <message>] [--format text|json] [<ticket-id>]", cmd.Usage())
}

func TestCloseCommand_SetupFlags(t *testing.T) {
	t.Parallel()

	cmd := &CloseCommand{}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	flags := cmd.SetupFlags(fs)

	// Verify flags is of correct type
	closeFlags, ok := flags.(*closeFlags)
	require.True(t, ok, "flags should be *closeFlags")

	// Test default values
	assert.False(t, closeFlags.force.Long)
	assert.False(t, closeFlags.force.Short)
	assert.Equal(t, "", closeFlags.reason)
	assert.Equal(t, FormatText, closeFlags.format.Long)
	assert.Equal(t, "", closeFlags.format.Short)

	// Test that flags are registered
	forceFlag := fs.Lookup("force")
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "false", forceFlag.DefValue)

	forceShortFlag := fs.Lookup("f")
	assert.NotNil(t, forceShortFlag)

	reasonFlag := fs.Lookup("reason")
	assert.NotNil(t, reasonFlag)
	assert.Equal(t, "", reasonFlag.DefValue)

	formatFlag := fs.Lookup("format")
	assert.NotNil(t, formatFlag)
	assert.Equal(t, FormatText, formatFlag.DefValue)

	formatShortFlag := fs.Lookup("o")
	assert.NotNil(t, formatShortFlag)
	assert.Equal(t, "", formatShortFlag.DefValue)
}

func TestCloseCommand_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		flags       interface{}
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid no arguments",
			flags: &closeFlags{
				format: StringFlag{Long: FormatText},
			},
			args:        []string{},
			expectError: false,
		},
		{
			name: "valid with ticket ID",
			flags: &closeFlags{
				format: StringFlag{Long: FormatText},
			},
			args:        []string{"ticket-123"},
			expectError: false,
		},
		{
			name: "too many arguments",
			flags: &closeFlags{
				format: StringFlag{Long: FormatText},
			},
			args:        []string{"ticket-123", "extra"},
			expectError: true,
			errorMsg:    "unexpected arguments after ticket ID",
		},
		{
			name: "invalid format",
			flags: &closeFlags{
				format: StringFlag{Long: "invalid"},
			},
			args:        []string{},
			expectError: true,
			errorMsg:    "invalid format",
		},
		{
			name:        "wrong flags type",
			flags:       struct{}{},
			args:        []string{},
			expectError: true,
			errorMsg:    "invalid flags type",
		},
		// Note: Testing flag precedence is not possible in unit tests
		// since we're directly setting values without going through flag parsing
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := &CloseCommand{}
			err := cmd.Validate(tt.flags, tt.args)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)

				// Verify args are stored
				if f, ok := tt.flags.(*closeFlags); ok {
					assert.Equal(t, tt.args, f.args)
				}
			}
		})
	}
}

func TestCloseCommand_Execute_Errors(t *testing.T) {
	// Cannot use t.Parallel() - cli.NewApp modifies global state

	tests := []struct {
		name        string
		flags       *closeFlags
		args        []string
		setupApp    func() error
		expectError bool
		errorMsg    string
	}{
		{
			name: "context cancelled",
			flags: &closeFlags{
				format: StringFlag{Long: FormatText},
				args:   []string{},
			},
			args: []string{},
			setupApp: func() error {
				return nil
			},
			expectError: true,
		},
		{
			name:  "invalid flags type",
			flags: nil, // Will pass wrong type
			args:  []string{},
			setupApp: func() error {
				return nil
			},
			expectError: true,
			errorMsg:    "invalid flags type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup any required app state
			if tt.setupApp != nil {
				err := tt.setupApp()
				require.NoError(t, err, "setup should not fail")
			}

			cmd := &CloseCommand{}

			var ctx context.Context
			if tt.name == "context cancelled" {
				cancelCtx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				ctx = cancelCtx
			} else {
				ctx = context.Background()
			}

			var err error
			if tt.flags == nil {
				// Test with wrong type
				err = cmd.Execute(ctx, struct{}{}, tt.args)
			} else {
				err = cmd.Execute(ctx, tt.flags, tt.args)
			}

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCloseCommand_Execute_Mock(t *testing.T) {
	// Cannot use t.Parallel() - would need mock App setup

	t.Run("text output calls correct app methods", func(t *testing.T) {
		// This test would require mocking the App and its methods
		// For now, we're testing the basic structure
		// Full mock testing would be done in integration tests
		t.Skip("Requires mock App implementation")
	})

	t.Run("JSON output format", func(t *testing.T) {
		// This test would require mocking the App and capturing JSON output
		t.Skip("Requires mock App implementation and output capture")
	})
}

// Test helper functions
func TestCloseCommand_OutputHelpers_Legacy(t *testing.T) {
	t.Parallel()

	t.Run("outputCloseErrorJSON formats error correctly", func(t *testing.T) {
		// This would test the error output format
		// Requires mocking app.Output.PrintJSON
		t.Skip("Requires mock Output implementation")
	})

	t.Run("outputCloseSuccessJSON includes all fields", func(t *testing.T) {
		// This would test that all expected fields are included
		// Requires mocking app.Output.PrintJSON and app.Manager.GetTicket
		t.Skip("Requires mock App and Manager implementation")
	})
}

// Integration test placeholder
func TestCloseCommand_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("close current ticket", func(t *testing.T) {
		// This would be an integration test with a real ticket
		t.Skip("Requires full integration test setup")
	})

	t.Run("close ticket by ID", func(t *testing.T) {
		// This would be an integration test with a real ticket
		t.Skip("Requires full integration test setup")
	})

	t.Run("force close with uncommitted changes", func(t *testing.T) {
		// This would test the force flag behavior
		t.Skip("Requires full integration test setup")
	})

	t.Run("close with reason", func(t *testing.T) {
		// This would test the reason flag behavior
		t.Skip("Requires full integration test setup")
	})
}
