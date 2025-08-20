package commands

import (
	"context"
	"flag"
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
	assert.False(t, closeFlags.force)
	assert.False(t, closeFlags.forceShort)
	assert.Equal(t, "", closeFlags.reason)
	assert.Equal(t, FormatText, closeFlags.format)
	assert.Equal(t, "", closeFlags.formatShort)

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
				format: FormatText,
			},
			args:        []string{},
			expectError: false,
		},
		{
			name: "valid with ticket ID",
			flags: &closeFlags{
				format: FormatText,
			},
			args:        []string{"ticket-123"},
			expectError: false,
		},
		{
			name: "too many arguments",
			flags: &closeFlags{
				format: FormatText,
			},
			args:        []string{"ticket-123", "extra"},
			expectError: true,
			errorMsg:    "unexpected arguments after ticket ID",
		},
		{
			name: "invalid format",
			flags: &closeFlags{
				format: "invalid",
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
		{
			name: "short form force takes precedence",
			flags: &closeFlags{
				force:      false,
				forceShort: true,
				format:     FormatText,
			},
			args:        []string{},
			expectError: false,
		},
		{
			name: "short form format takes precedence",
			flags: &closeFlags{
				format:      FormatText,
				formatShort: FormatJSON,
			},
			args:        []string{},
			expectError: false,
		},
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

				// Verify normalization happened
				if f, ok := tt.flags.(*closeFlags); ok {
					if f.forceShort {
						assert.True(t, f.force, "short form should set long form for force flag")
					}
					// Only check format normalization if formatShort is not empty
					if f.formatShort != "" {
						assert.Equal(t, f.formatShort, f.format, "short form should set long form when provided")
					}
					// Verify args are stored
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
				format: FormatText,
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

// Test flag normalization
func TestCloseFlags_Normalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flags    closeFlags
		expected closeFlags
	}{
		{
			name: "no short forms",
			flags: closeFlags{
				force:  true,
				format: FormatJSON,
			},
			expected: closeFlags{
				force:  true,
				format: FormatJSON,
			},
		},
		{
			name: "short force sets force via OR",
			flags: closeFlags{
				force:      false,
				forceShort: true,
				format:     FormatText,
			},
			expected: closeFlags{
				force:      true,
				forceShort: true,
				format:     FormatText,
			},
		},
		{
			name: "both force flags true",
			flags: closeFlags{
				force:      true,
				forceShort: true,
				format:     FormatText,
			},
			expected: closeFlags{
				force:      true,
				forceShort: true,
				format:     FormatText,
			},
		},
		{
			name: "short format overrides",
			flags: closeFlags{
				format:      FormatText,
				formatShort: FormatJSON,
			},
			expected: closeFlags{
				format:      FormatJSON,
				formatShort: FormatJSON,
			},
		},
		{
			name: "both short forms override",
			flags: closeFlags{
				force:       false,
				forceShort:  true,
				format:      FormatText,
				formatShort: FormatJSON,
			},
			expected: closeFlags{
				force:       true,
				forceShort:  true,
				format:      FormatJSON,
				formatShort: FormatJSON,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flags := tt.flags
			flags.normalize()

			assert.Equal(t, tt.expected.force, flags.force)
			assert.Equal(t, tt.expected.format, flags.format)
		})
	}
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
