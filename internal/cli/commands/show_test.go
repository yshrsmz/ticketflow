package commands

import (
	flag "github.com/spf13/pflag"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShowCommand_Name(t *testing.T) {
	cmd := NewShowCommand()
	assert.Equal(t, "show", cmd.Name())
}

func TestShowCommand_Aliases(t *testing.T) {
	cmd := NewShowCommand()
	assert.Nil(t, cmd.Aliases())
}

func TestShowCommand_Description(t *testing.T) {
	cmd := NewShowCommand()
	assert.Equal(t, "Show ticket details", cmd.Description())
}

func TestShowCommand_Usage(t *testing.T) {
	cmd := NewShowCommand()
	assert.Equal(t, "show <ticket-id> [--format text|json]", cmd.Usage())
}

func TestShowCommand_SetupFlags(t *testing.T) {
	cmd := &ShowCommand{}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	flags := cmd.SetupFlags(fs)

	assert.NotNil(t, flags)
	showFlags, ok := flags.(*showFlags)
	assert.True(t, ok)
	assert.Equal(t, FormatText, showFlags.format) // Default value

	// Test that flags are registered
	formatFlag := fs.Lookup("format")
	assert.NotNil(t, formatFlag)
	assert.Equal(t, FormatText, formatFlag.DefValue)
}

func TestShowCommand_Validate(t *testing.T) {
	tests := []struct {
		name      string
		flags     interface{} // Changed to interface{} to test type assertions
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid with ticket ID and default format",
			flags:     &showFlags{format: FormatText},
			args:      []string{"123456"},
			expectErr: false,
		},
		{
			name:      "valid with ticket ID and json format",
			flags:     &showFlags{format: FormatJSON},
			args:      []string{"test-ticket"},
			expectErr: false,
		},
		{
			name:      "valid with partial ticket ID",
			flags:     &showFlags{format: FormatText},
			args:      []string{"250813"},
			expectErr: false,
		},
		{
			name:      "missing ticket ID",
			flags:     &showFlags{format: FormatText},
			args:      []string{},
			expectErr: true,
			errMsg:    "missing ticket ID argument",
		},
		{
			name:      "invalid format",
			flags:     &showFlags{format: "yaml"},
			args:      []string{"123456"},
			expectErr: true,
			errMsg:    `invalid format: "yaml" (must be "text" or "json")`,
		},
		{
			name:      "empty format defaults to text",
			flags:     &showFlags{format: ""},
			args:      []string{"123456"},
			expectErr: false, // Now valid - defaults to "text"
		},
		{
			name:      "too many arguments",
			flags:     &showFlags{format: FormatText},
			args:      []string{"123456", "extra", "args"},
			expectErr: true,
			errMsg:    `unexpected arguments after ticket ID: [extra args]`,
		},
		{
			name:      "invalid flags type",
			flags:     "not a showFlags",
			args:      []string{"123456"},
			expectErr: true,
			errMsg:    `invalid flags type: expected *commands.showFlags, got string`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &ShowCommand{}
			err := cmd.Validate(tt.flags, tt.args)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Equal(t, tt.errMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
				// Check if empty format was defaulted to text
				if sf, ok := tt.flags.(*showFlags); ok && tt.name == "empty format defaults to text" {
					assert.Equal(t, FormatText, sf.format, "empty format should be defaulted to text")
				}
			}
		})
	}
}
