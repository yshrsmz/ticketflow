package command

import (
	"context"
	"errors"
	flag "github.com/spf13/pflag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testCommand is a full implementation for testing interface contracts
type testCommand struct {
	name        string
	aliases     []string
	description string
	usage       string
	validateErr error
	executeErr  error
	flags       *testFlags
}

type testFlags struct {
	verbose bool
	output  string
}

func (c *testCommand) Name() string        { return c.name }
func (c *testCommand) Aliases() []string   { return c.aliases }
func (c *testCommand) Description() string { return c.description }
func (c *testCommand) Usage() string       { return c.usage }

func (c *testCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	flags := &testFlags{}
	fs.BoolVar(&flags.verbose, "verbose", false, "verbose output")
	fs.StringVar(&flags.output, "output", "", "output file")
	c.flags = flags
	return flags
}

func (c *testCommand) Validate(flags interface{}, args []string) error {
	if c.validateErr != nil {
		return c.validateErr
	}

	f, ok := flags.(*testFlags)
	if !ok {
		return errors.New("invalid flags type")
	}

	if f.output == "" && len(args) == 0 {
		return errors.New("either -output or arguments required")
	}

	return nil
}

func (c *testCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	if c.executeErr != nil {
		return c.executeErr
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}

func TestCommand_Interface(t *testing.T) {
	cmd := &testCommand{
		name:        "test",
		description: "A test command",
		usage:       "test [flags] [args...]",
	}

	// Test basic metadata methods
	assert.Equal(t, "test", cmd.Name())
	assert.Equal(t, "A test command", cmd.Description())
	assert.Equal(t, "test [flags] [args...]", cmd.Usage())

	// Test SetupFlags
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	flags := cmd.SetupFlags(fs)
	require.NotNil(t, flags)

	// Parse some flags
	err := fs.Parse([]string{"--verbose", "--output", "test.txt"})
	require.NoError(t, err)

	f, ok := flags.(*testFlags)
	require.True(t, ok)
	assert.True(t, f.verbose)
	assert.Equal(t, "test.txt", f.output)
}

func TestCommand_Validate(t *testing.T) {
	tests := []struct {
		name        string
		flags       *testFlags
		args        []string
		validateErr error
		wantErr     bool
	}{
		{
			name:  "valid with output flag",
			flags: &testFlags{output: "test.txt"},
			args:  []string{},
		},
		{
			name:  "valid with arguments",
			flags: &testFlags{},
			args:  []string{"arg1"},
		},
		{
			name:    "invalid without output or args",
			flags:   &testFlags{},
			args:    []string{},
			wantErr: true,
		},
		{
			name:        "custom validation error",
			flags:       &testFlags{output: "test.txt"},
			validateErr: errors.New("custom error"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &testCommand{
				validateErr: tt.validateErr,
			}

			err := cmd.Validate(tt.flags, tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCommand_Execute(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		cmd := &testCommand{}
		ctx := context.Background()
		flags := &testFlags{output: "test.txt"}

		err := cmd.Execute(ctx, flags, []string{"arg1"})
		assert.NoError(t, err)
	})

	t.Run("execution with error", func(t *testing.T) {
		cmd := &testCommand{
			executeErr: errors.New("execution failed"),
		}
		ctx := context.Background()
		flags := &testFlags{}

		err := cmd.Execute(ctx, flags, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "execution failed")
	})

	t.Run("execution with cancelled context", func(t *testing.T) {
		cmd := &testCommand{}
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		flags := &testFlags{}
		err := cmd.Execute(ctx, flags, []string{})
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})
}

func TestMetadata(t *testing.T) {
	meta := Metadata{
		Name:        "example",
		Description: "An example command",
		Usage:       "example [options]",
		Examples: []Example{
			{
				Description: "Basic usage",
				Command:     "ticketflow example",
			},
			{
				Description: "With options",
				Command:     "ticketflow example -verbose",
			},
		},
	}

	assert.Equal(t, "example", meta.Name)
	assert.Equal(t, "An example command", meta.Description)
	assert.Equal(t, "example [options]", meta.Usage)
	assert.Len(t, meta.Examples, 2)
	assert.Equal(t, "Basic usage", meta.Examples[0].Description)
}

func TestResult(t *testing.T) {
	t.Run("success result", func(t *testing.T) {
		result := Result{
			Success: true,
			Message: "Operation completed",
			Data:    map[string]string{"id": "123"},
		}

		assert.True(t, result.Success)
		assert.Equal(t, "Operation completed", result.Message)
		assert.NotNil(t, result.Data)
	})

	t.Run("failure result", func(t *testing.T) {
		result := Result{
			Success: false,
			Message: "Operation failed",
			Data:    nil,
		}

		assert.False(t, result.Success)
		assert.Equal(t, "Operation failed", result.Message)
		assert.Nil(t, result.Data)
	})
}
