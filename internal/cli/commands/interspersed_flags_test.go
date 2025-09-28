package commands

import (
	"testing"

	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInterspersedFlags_ShowCommand verifies that flags can be placed after positional arguments
func TestInterspersedFlags_ShowCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedTicket string
		expectedFormat string
	}{
		{
			name:           "flags before positional args (traditional)",
			args:           []string{"--format", "json", "ticket-123"},
			expectedTicket: "ticket-123",
			expectedFormat: "json",
		},
		{
			name:           "flags after positional args (interspersed)",
			args:           []string{"ticket-123", "--format", "json"},
			expectedTicket: "ticket-123",
			expectedFormat: "json",
		},
		// Note: show command doesn't have short flags, only --format
		{
			name:           "mixed flag placement",
			args:           []string{"--format", "text", "ticket-789", "--format", "json"}, // Last value wins in pflag
			expectedTicket: "ticket-789",
			expectedFormat: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewShowCommand()
			fs := flag.NewFlagSet(cmd.Name(), flag.ContinueOnError)
			flags := cmd.SetupFlags(fs)

			// Parse the args
			err := fs.Parse(tt.args)
			require.NoError(t, err)

			// Get the positional arguments
			positionalArgs := fs.Args()
			require.Len(t, positionalArgs, 1, "should have exactly one positional arg")
			assert.Equal(t, tt.expectedTicket, positionalArgs[0])

			// Verify the flags were parsed correctly
			showFlags := flags.(*showFlags)
			assert.Equal(t, tt.expectedFormat, showFlags.format)
		})
	}
}

// TestInterspersedFlags_StartCommand verifies interspersed flags for start command
func TestInterspersedFlags_StartCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedTicket string
		expectedForce  bool
		expectedFormat string
	}{
		{
			name:           "force flag after positional arg",
			args:           []string{"ticket-123", "--force"},
			expectedTicket: "ticket-123",
			expectedForce:  true,
			expectedFormat: "text", // default
		},
		{
			name:           "multiple flags after positional arg",
			args:           []string{"ticket-456", "--force", "--format", "json"},
			expectedTicket: "ticket-456",
			expectedForce:  true,
			expectedFormat: "json",
		},
		{
			name:           "short form flags after positional arg",
			args:           []string{"ticket-789", "-f", "-o", "json"},
			expectedTicket: "ticket-789",
			expectedForce:  true,
			expectedFormat: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewStartCommand()
			fs := flag.NewFlagSet(cmd.Name(), flag.ContinueOnError)
			flags := cmd.SetupFlags(fs)

			// Parse the args
			err := fs.Parse(tt.args)
			require.NoError(t, err)

			// Get the positional arguments
			positionalArgs := fs.Args()
			require.Len(t, positionalArgs, 1, "should have exactly one positional arg")
			assert.Equal(t, tt.expectedTicket, positionalArgs[0])

			// Verify the flags were parsed correctly
			startFlags := flags.(*startFlags)
			assert.Equal(t, tt.expectedForce, startFlags.force)
			assert.Equal(t, tt.expectedFormat, startFlags.format)
		})
	}
}

// TestInterspersedFlags_DoubleDashTerminator verifies that -- stops flag parsing
func TestInterspersedFlags_DoubleDashTerminator(t *testing.T) {
	cmd := NewShowCommand()
	fs := flag.NewFlagSet(cmd.Name(), flag.ContinueOnError)
	flags := cmd.SetupFlags(fs)

	// Args with double dash terminator
	args := []string{"--format", "json", "--", "ticket-123", "--not-a-flag"}

	err := fs.Parse(args)
	require.NoError(t, err)

	// After --, everything should be treated as positional args
	positionalArgs := fs.Args()
	assert.Len(t, positionalArgs, 2)
	assert.Equal(t, "ticket-123", positionalArgs[0])
	assert.Equal(t, "--not-a-flag", positionalArgs[1])

	// Format flag should still be parsed
	showFlags := flags.(*showFlags)
	assert.Equal(t, "json", showFlags.format)
}

// TestInterspersedFlags_ValidationStillWorks verifies that validation of extra positional args still works
func TestInterspersedFlags_ValidationStillWorks(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid with flags after positional",
			args:        []string{"ticket-123", "--format", "json"},
			expectError: false,
		},
		{
			name:        "invalid with extra positional args",
			args:        []string{"ticket-123", "extra-arg"},
			expectError: true,
			errorMsg:    "unexpected arguments after ticket ID: [extra-arg]",
		},
		{
			name:        "invalid with multiple extra positional args",
			args:        []string{"ticket-123", "extra1", "extra2"},
			expectError: true,
			errorMsg:    "unexpected arguments after ticket ID: [extra1 extra2]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewShowCommand()
			fs := flag.NewFlagSet(cmd.Name(), flag.ContinueOnError)
			flags := cmd.SetupFlags(fs)

			// Parse the args
			err := fs.Parse(tt.args)
			require.NoError(t, err)

			// Validate should catch extra positional args but not flags
			err = cmd.Validate(flags, fs.Args())
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
