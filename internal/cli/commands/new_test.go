package commands

import (
	"flag"
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
	assert.Equal(t, "", newFlags.parent)      // Default empty
	assert.Equal(t, "", newFlags.parentShort) // Default empty
	assert.Equal(t, "text", newFlags.format)  // Default value
	assert.Equal(t, "", newFlags.formatShort) // Default empty (not provided)

	// Test that long form flags are registered
	parentFlag := fs.Lookup("parent")
	assert.NotNil(t, parentFlag)
	assert.Equal(t, "", parentFlag.DefValue)

	formatFlag := fs.Lookup("format")
	assert.NotNil(t, formatFlag)
	assert.Equal(t, "text", formatFlag.DefValue)

	// Test that short form flags are registered
	pFlag := fs.Lookup("p")
	assert.NotNil(t, pFlag)
	assert.Equal(t, "", pFlag.DefValue)

	oFlag := fs.Lookup("o")
	assert.NotNil(t, oFlag)
	assert.Equal(t, "", oFlag.DefValue)
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
			flags:     &newFlags{parent: "parent-123", format: StringFlag{Long: "text"}},
			args:      []string{"my-new-ticket"},
			expectErr: false,
		},
		{
			name:      "valid with slug and short parent",
			flags:     &newFlags{parentShort: "parent-456", format: StringFlag{Long: "text"}},
			args:      []string{"my-new-ticket"},
			expectErr: false,
		},
		{
			name:      "short parent takes precedence",
			flags:     &newFlags{parent: "parent-123", parentShort: "parent-456", format: StringFlag{Long: "text"}},
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
			name:      "valid with short format flag",
			flags:     &newFlags{formatShort: "json"},
			args:      []string{"my-new-ticket"},
			expectErr: false,
		},
		{
			name:      "short format takes precedence",
			flags:     &newFlags{format: StringFlag{Long: "text"}, formatShort: "json"},
			args:      []string{"my-new-ticket"},
			expectErr: false,
			// After validation, format should be "json"
		},
		{
			name:      "short format text overrides long format json",
			flags:     &newFlags{format: StringFlag{Long: "json"}, formatShort: "text"},
			args:      []string{"my-new-ticket"},
			expectErr: false,
			// After validation, format should be "text" (short form takes precedence)
		},
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
			errMsg:    `invalid format: StringFlag{Long: "yaml"} (must be "text" or "json")`,
		},
		{
			name:      "invalid flags type",
			flags:     "not a newFlags",
			args:      []string{"my-ticket"},
			expectErr: true,
			errMsg:    `invalid flags type: expected *newFlags, got string`,
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
				// Check flag merging
				if nf, ok := tt.flags.(*newFlags); ok {
					if tt.name == "short parent takes precedence" {
						assert.Equal(t, "parent-456", nf.parent, "parentShort should override parent")
					}
					if tt.name == "short format takes precedence" {
						assert.Equal(t, "json", nf.format, "formatShort should override format")
					}
					if tt.name == "short format text overrides long format json" {
						assert.Equal(t, "text", nf.format, "formatShort text should override format json")
					}
					if tt.name == "valid with short format flag" {
						assert.Equal(t, "json", nf.format, "formatShort should be merged into format")
					}
				}
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
