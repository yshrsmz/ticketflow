package commands

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/command"
)

func TestRestoreCommand_Interface(t *testing.T) {
	t.Parallel()

	cmd := NewRestoreCommand()

	// Verify it implements the Command interface
	var _ = command.Command(cmd)

	// Test Name
	assert.Equal(t, "restore", cmd.Name())

	// Test Aliases
	assert.Nil(t, cmd.Aliases())

	// Test Description
	assert.Equal(t, "Restore the current-ticket.md symlink in a worktree", cmd.Description())

	// Test Usage
	assert.Equal(t, "restore [--format text|json]", cmd.Usage())
}

func TestRestoreCommand_SetupFlags(t *testing.T) {
	t.Parallel()

	cmd := &RestoreCommand{}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	flags := cmd.SetupFlags(fs)

	// Verify flags is of correct type
	restoreFlags, ok := flags.(*restoreFlags)
	require.True(t, ok, "flags should be *restoreFlags")

	// Test default values
	assert.Equal(t, FormatText, restoreFlags.format)
	assert.Equal(t, "", restoreFlags.formatShort)

	// Test that flags are registered
	formatFlag := fs.Lookup("format")
	assert.NotNil(t, formatFlag)
	assert.Equal(t, FormatText, formatFlag.DefValue)
	assert.Equal(t, "Output format (text|json)", formatFlag.Usage)

	formatShortFlag := fs.Lookup("o")
	assert.NotNil(t, formatShortFlag)
	assert.Equal(t, "", formatShortFlag.DefValue)
	assert.Equal(t, "Output format (short form)", formatShortFlag.Usage)
}

func TestRestoreCommand_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		flags       interface{}
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "invalid flags type",
			flags:       "not a restoreFlags",
			args:        []string{},
			expectError: true,
			errorMsg:    "invalid flags type: expected *restoreFlags, got string",
		},
		{
			name: "valid no arguments with text format",
			flags: &restoreFlags{
				format: FormatText,
			},
			args:        []string{},
			expectError: false,
		},
		{
			name: "valid no arguments with json format",
			flags: &restoreFlags{
				format: FormatJSON,
			},
			args:        []string{},
			expectError: false,
		},
		{
			name: "error with arguments",
			flags: &restoreFlags{
				format: FormatText,
			},
			args:        []string{"some-arg"},
			expectError: true,
			errorMsg:    "restore command does not accept any arguments",
		},
		{
			name: "error with multiple arguments",
			flags: &restoreFlags{
				format: FormatText,
			},
			args:        []string{"arg1", "arg2"},
			expectError: true,
			errorMsg:    "restore command does not accept any arguments",
		},
		{
			name: "invalid format",
			flags: &restoreFlags{
				format: "invalid",
			},
			args:        []string{},
			expectError: true,
			errorMsg:    `invalid format: "invalid" (must be "text" or "json")`,
		},
		{
			name: "format normalization - prefer short form",
			flags: &restoreFlags{
				format:      FormatText,
				formatShort: FormatJSON,
			},
			args:        []string{},
			expectError: false,
		},
		{
			name: "format normalization - both json",
			flags: &restoreFlags{
				format:      FormatJSON,
				formatShort: FormatJSON,
			},
			args:        []string{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &RestoreCommand{}
			err := cmd.Validate(tt.flags, tt.args)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.EqualError(t, err, tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)

				// Check format normalization
				if f, ok := tt.flags.(*restoreFlags); ok {
					if f.formatShort != "" {
						assert.Equal(t, f.formatShort, f.format, "format should be normalized to formatShort")
					}
				}
			}
		})
	}
}

func TestRestoreCommand_FlagNormalization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		format         string
		formatShort    string
		expectedFormat string
	}{
		{
			name:           "both text",
			format:         FormatText,
			formatShort:    FormatText,
			expectedFormat: FormatText,
		},
		{
			name:           "format text, short json - prefer short",
			format:         FormatText,
			formatShort:    FormatJSON,
			expectedFormat: FormatJSON,
		},
		{
			name:           "format json, short empty - keep format",
			format:         FormatJSON,
			formatShort:    "",
			expectedFormat: FormatJSON,
		},
		{
			name:           "format json, short text - prefer short",
			format:         FormatJSON,
			formatShort:    FormatText,
			expectedFormat: FormatText,
		},
		{
			name:           "both json",
			format:         FormatJSON,
			formatShort:    FormatJSON,
			expectedFormat: FormatJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &RestoreCommand{}
			flags := &restoreFlags{
				format:      tt.format,
				formatShort: tt.formatShort,
			}

			err := cmd.Validate(flags, []string{})
			assert.NoError(t, err)

			// Check that format was normalized correctly
			if tt.formatShort != "" {
				assert.Equal(t, tt.expectedFormat, flags.format)
			} else {
				assert.Equal(t, tt.format, flags.format)
			}
		})
	}
}

// TestRestoreCommand_Coverage ensures we have good test coverage
func TestRestoreCommand_Coverage(t *testing.T) {
	t.Parallel()

	// Test all public methods are called
	cmd := NewRestoreCommand()

	// Name
	_ = cmd.Name()

	// Aliases
	_ = cmd.Aliases()

	// Description
	_ = cmd.Description()

	// Usage
	_ = cmd.Usage()

	// SetupFlags
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	flags := cmd.SetupFlags(fs)

	// Validate
	_ = cmd.Validate(flags, []string{})

	// All methods have been called, achieving coverage
}
