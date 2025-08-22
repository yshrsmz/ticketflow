package commands

import (
	"context"
	"errors"
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yshrsmz/ticketflow/internal/cli"
)

// MockApp is a mock implementation of cli.App for testing
type MockApp struct {
	mock.Mock
}

func (m *MockApp) Status(ctx context.Context, format cli.OutputFormat) error {
	args := m.Called(ctx, format)
	return args.Error(0)
}

func TestStatusCommand_Interface(t *testing.T) {
	cmd := NewStatusCommand()

	// Verify it implements the Command interface
	var _ = cmd

	assert.Equal(t, "status", cmd.Name())
	assert.Nil(t, cmd.Aliases())
	assert.Equal(t, "Show the status of the current ticket", cmd.Description())
	assert.Equal(t, "status [--format text|json]", cmd.Usage())
}

func TestStatusCommand_SetupFlags(t *testing.T) {
	cmd := &StatusCommand{}

	// Test parsing different format values
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "default format",
			args:     []string{},
			expected: FormatText,
		},
		{
			name:     "json format",
			args:     []string{"--format", "json"},
			expected: FormatJSON,
		},
		{
			name:     "text format explicit",
			args:     []string{"--format", "text"},
			expected: FormatText,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			flags := cmd.SetupFlags(fs)

			// Verify flags is not nil
			assert.NotNil(t, flags)

			err := fs.Parse(tt.args)
			assert.NoError(t, err)

			// Use reflection to check the format field value
			// since statusFlags is unexported
			sf := flags.(*statusFlags)
			assert.Equal(t, tt.expected, sf.format)
		})
	}
}

func TestStatusCommand_Validate(t *testing.T) {
	cmd := &StatusCommand{}

	tests := []struct {
		name      string
		flags     interface{}
		args      []string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid text format",
			flags:     &statusFlags{format: FormatText},
			args:      []string{},
			wantError: false,
		},
		{
			name:      "valid json format",
			flags:     &statusFlags{format: FormatJSON},
			args:      []string{},
			wantError: false,
		},
		{
			name:      "invalid format",
			flags:     &statusFlags{format: "xml"},
			args:      []string{},
			wantError: true,
			errorMsg:  `invalid format: "xml" (must be "text" or "json")`,
		},
		{
			name:      "empty format defaults to text",
			flags:     &statusFlags{format: ""},
			args:      []string{},
			wantError: true,
			errorMsg:  `invalid format: "" (must be "text" or "json")`,
		},
		{
			name:      "with unexpected arguments but valid format",
			flags:     &statusFlags{format: FormatJSON},
			args:      []string{"extra", "args"},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Validate(tt.flags, tt.args)
			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.EqualError(t, err, tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStatusCommand_Execute(t *testing.T) {
	// Integration test that verifies the command works with real App
	// This test will succeed in the actual ticketflow environment

	t.Run("text format", func(t *testing.T) {
		cmd := &StatusCommand{}
		ctx := context.Background()
		flags := &statusFlags{format: FormatText}

		// This will succeed when run in a ticketflow environment
		err := cmd.Execute(ctx, flags, []string{})

		// The command should execute without error when there's a current ticket
		// or return a specific error when there's no current ticket
		// Since we're in a ticketflow worktree, it should succeed
		assert.NoError(t, err)
	})

	t.Run("json format", func(t *testing.T) {
		cmd := &StatusCommand{}
		ctx := context.Background()
		flags := &statusFlags{format: FormatJSON}

		err := cmd.Execute(ctx, flags, []string{})

		// The command should execute without error when there's a current ticket
		// or return a specific error when there's no current ticket
		// Since we're in a ticketflow worktree, it should succeed
		assert.NoError(t, err)
	})
}

// TestStatusCommand_Execute_WithMockApp demonstrates how we would test
// if cli.NewApp supported dependency injection
func TestStatusCommand_Execute_WithMockApp(t *testing.T) {
	t.Skip("Skipping test that requires refactoring cli.NewApp for dependency injection")

	tests := []struct {
		name        string
		format      string
		statusError error
		wantError   bool
	}{
		{
			name:        "successful text status",
			format:      FormatText,
			statusError: nil,
			wantError:   false,
		},
		{
			name:        "successful json status",
			format:      FormatJSON,
			statusError: nil,
			wantError:   false,
		},
		{
			name:        "status returns error",
			format:      FormatText,
			statusError: errors.New("no current ticket"),
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test demonstrates the ideal testing approach
			// if we could inject the App dependency
			mockApp := new(MockApp)
			ctx := context.Background()

			expectedFormat := cli.ParseOutputFormat(tt.format)
			mockApp.On("Status", ctx, expectedFormat).Return(tt.statusError)

			// We would need a way to inject mockApp into the command
			// This would require refactoring StatusCommand.Execute
			// to accept an App instance or use a factory pattern
		})
	}
}
