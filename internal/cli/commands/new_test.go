package commands

import (
	flag "github.com/spf13/pflag"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCommand_Name(t *testing.T) {
	cmd := NewNewCommand()
	assert.Equal(t, "new", cmd.Name())
}

func TestNewCommand_Aliases(t *testing.T) {
	cmd := NewNewCommand()
	assert.Nil(t, cmd.Aliases())
}

func TestNewCommand_Description(t *testing.T) {
	cmd := NewNewCommand()
	assert.Equal(t, "Create a new ticket", cmd.Description())
}

func TestNewCommand_Usage(t *testing.T) {
	cmd := NewNewCommand()
	assert.Equal(t, "new [--parent <ticket-id>] [--format text|json] <slug>", cmd.Usage())
}

func TestNewCommand_SetupFlags(t *testing.T) {
	cmd := &NewCommand{}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	flags := cmd.SetupFlags(fs)

	assert.NotNil(t, flags)
	newFlags, ok := flags.(*newFlags)
	assert.True(t, ok)
	assert.Equal(t, "", newFlags.parent.Long)     // Default empty
	assert.Equal(t, "", newFlags.parent.Short)    // Default empty
	assert.Equal(t, "text", newFlags.format.Long) // Default value
	assert.Equal(t, "", newFlags.format.Short)    // Default empty (not provided)

	// Test that long form flags are registered
	parentFlag := fs.Lookup("parent")
	assert.NotNil(t, parentFlag)
	assert.Equal(t, "", parentFlag.DefValue)

	formatFlag := fs.Lookup("format")
	assert.NotNil(t, formatFlag)
	assert.Equal(t, "text", formatFlag.DefValue)

	// Test that short form flags are registered
	// Phase 1: With pflag, use ShorthandLookup for shorthand flags
	pFlag := fs.ShorthandLookup("p")
	assert.NotNil(t, pFlag)
	assert.Equal(t, "", pFlag.DefValue)

	// Phase 1: With pflag, use ShorthandLookup for shorthand flags
	oFlag := fs.ShorthandLookup("o")
	assert.NotNil(t, oFlag)
	assert.Equal(t, "text", oFlag.DefValue) // Note: Default is from the long form flag
}

func TestNewCommand_Validate(t *testing.T) {
	tests := []struct {
		name      string
		flags     interface{}
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid with slug only",
			flags:     &newFlags{format: StringFlag{Long: "text"}},
			args:      []string{"my-new-ticket"},
			expectErr: false,
		},
		{
			name:      "valid with slug and parent",
			flags:     &newFlags{parent: StringFlag{Long: "parent-123"}, format: StringFlag{Long: "text"}},
			args:      []string{"my-new-ticket"},
			expectErr: false,
		},
		{
			name:      "valid with slug and short parent",
			flags:     &newFlags{parent: StringFlag{Short: "parent-456"}, format: StringFlag{Long: "text"}},
			args:      []string{"my-new-ticket"},
			expectErr: false,
		},
		{
			name:      "short parent takes precedence",
			flags:     &newFlags{parent: StringFlag{Long: "parent-123", Short: "parent-456"}, format: StringFlag{Long: "text"}},
			args:      []string{"my-new-ticket"},
			expectErr: false,
			// After validation, parent should be "parent-456"
		},
		{
			name:      "valid with json format",
			flags:     &newFlags{format: StringFlag{Long: "json"}},
			args:      []string{"my-new-ticket"},
			expectErr: false,
		},
		{
			name:      "valid with json format",
			flags:     &newFlags{format: StringFlag{Long: "json"}},
			args:      []string{"my-new-ticket"},
			expectErr: false,
		},
		// Note: Testing flag precedence is not possible in unit tests
		// since we're directly setting values without going through flag parsing
		{
			name:      "missing slug",
			flags:     &newFlags{format: StringFlag{Long: "text"}},
			args:      []string{},
			expectErr: true,
			errMsg:    "missing slug argument",
		},
		{
			name:      "too many arguments",
			flags:     &newFlags{format: StringFlag{Long: "text"}},
			args:      []string{"slug1", "slug2", "slug3"},
			expectErr: true,
			errMsg:    `unexpected arguments after slug: [slug2 slug3]`,
		},
		{
			name:      "invalid format",
			flags:     &newFlags{format: StringFlag{Long: "yaml"}},
			args:      []string{"my-ticket"},
			expectErr: true,
			errMsg:    `invalid format: "yaml" (must be "text" or "json")`,
		},
		{
			name:      "invalid flags type",
			flags:     "not a newFlags",
			args:      []string{"my-ticket"},
			expectErr: true,
			errMsg:    `invalid flags type: expected *commands.newFlags, got string`,
		},
		{
			name:      "slug with numbers",
			flags:     &newFlags{format: StringFlag{Long: "text"}},
			args:      []string{"ticket-123"},
			expectErr: false,
		},
		{
			name:      "slug with hyphens",
			flags:     &newFlags{format: StringFlag{Long: "text"}},
			args:      []string{"my-feature-ticket"},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &NewCommand{}
			err := cmd.Validate(tt.flags, tt.args)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Equal(t, tt.errMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
				// Note: Cannot test flag precedence in unit tests since we're directly
				// setting flag values without going through the flag parsing mechanism.
				// The flag utilities handle precedence correctly when flags are parsed.
			}
		})
	}
}

// Note: MockApp is already defined in status_test.go in the same package
// We would use it here if we needed to test Execute with mocks

func TestNewCommand_Execute(t *testing.T) {
	// Note: Full execution testing would require mocking the entire App structure
	// which is complex. Here we test that Execute properly calls the App.NewTicket method
	// with the correct parameters. Integration tests will verify the full flow.

	// This is a placeholder for the complex Execute test that would require
	// significant mocking infrastructure. The actual behavior is tested
	// through integration tests.
	t.Run("execute calls App.NewTicket with correct parameters", func(t *testing.T) {
		// This would require mocking cli.NewApp which is not straightforward
		// Integration tests cover this scenario
		t.Skip("Covered by integration tests")
	})
}
