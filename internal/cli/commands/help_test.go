package commands

import (
	"bytes"
	"context"
	"flag"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/command"
)

// mockCommand is a test implementation of the Command interface
type mockCommand struct {
	name        string
	aliases     []string
	description string
	usage       string
}

func (m *mockCommand) Name() string                                    { return m.name }
func (m *mockCommand) Aliases() []string                               { return m.aliases }
func (m *mockCommand) Description() string                             { return m.description }
func (m *mockCommand) Usage() string                                   { return m.usage }
func (m *mockCommand) SetupFlags(fs *flag.FlagSet) interface{}         { return nil }
func (m *mockCommand) Validate(flags interface{}, args []string) error { return nil }
func (m *mockCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	return nil
}

// mockRegistry is a test implementation of the Registry interface
type mockRegistry struct {
	commands map[string]command.Command
	aliases  map[string]string
}

func newMockRegistry() *mockRegistry {
	return &mockRegistry{
		commands: make(map[string]command.Command),
		aliases:  make(map[string]string),
	}
}

func (r *mockRegistry) Register(cmd command.Command) error {
	r.commands[cmd.Name()] = cmd
	for _, alias := range cmd.Aliases() {
		r.aliases[alias] = cmd.Name()
	}
	return nil
}

func (r *mockRegistry) Get(name string) (command.Command, bool) {
	if cmd, ok := r.commands[name]; ok {
		return cmd, true
	}
	if cmdName, ok := r.aliases[name]; ok {
		return r.commands[cmdName], true
	}
	return nil, false
}

func (r *mockRegistry) List() []command.Command {
	commands := make([]command.Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		commands = append(commands, cmd)
	}
	return commands
}

func TestHelpCommand_Name(t *testing.T) {
	cmd := NewHelpCommand(nil, "1.0.0")
	assert.Equal(t, "help", cmd.Name())
}

func TestHelpCommand_Aliases(t *testing.T) {
	cmd := NewHelpCommand(nil, "1.0.0")
	aliases := cmd.Aliases()
	assert.Contains(t, aliases, "-h")
	assert.Contains(t, aliases, "--help")
}

func TestHelpCommand_Description(t *testing.T) {
	cmd := NewHelpCommand(nil, "1.0.0")
	assert.Equal(t, "Show help information", cmd.Description())
}

func TestHelpCommand_Usage(t *testing.T) {
	cmd := NewHelpCommand(nil, "1.0.0")
	assert.Equal(t, "help [command]", cmd.Usage())
}

func TestHelpCommand_SetupFlags(t *testing.T) {
	cmd := NewHelpCommand(nil, "1.0.0")
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	result := cmd.SetupFlags(fs)
	assert.Nil(t, result)
}

func TestHelpCommand_Validate(t *testing.T) {
	cmd := NewHelpCommand(nil, "1.0.0")

	tests := []struct {
		name string
		args []string
	}{
		{"no args", []string{}},
		{"with command", []string{"version"}},
		{"with multiple args", []string{"version", "extra"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Validate(nil, tt.args)
			assert.NoError(t, err)
		})
	}
}

func TestHelpCommand_Execute_GeneralHelp(t *testing.T) {
	// Create a mock registry with some commands
	registry := newMockRegistry()
	err := registry.Register(&mockCommand{
		name:        "version",
		aliases:     []string{"-v", "--version"},
		description: "Show version information",
		usage:       "version",
	})
	require.NoError(t, err)
	err = registry.Register(&mockCommand{
		name:        "help",
		aliases:     []string{"-h", "--help"},
		description: "Show help information",
		usage:       "help [command]",
	})
	require.NoError(t, err)

	cmd := NewHelpCommand(registry, "1.0.0")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = cmd.Execute(context.Background(), nil, []string{})
	require.NoError(t, err)

	err = w.Close()
	require.NoError(t, err)
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	// Check that output contains expected content
	assert.Contains(t, output, "TicketFlow - Git worktree-based ticket management system")
	assert.Contains(t, output, "v1.0.0")
	assert.Contains(t, output, "USAGE:")
	assert.Contains(t, output, "COMMANDS:")
	assert.Contains(t, output, "OPTIONS:")
	assert.Contains(t, output, "EXAMPLES:")
	assert.Contains(t, output, "version")
	assert.Contains(t, output, "Show version information")
	assert.Contains(t, output, "help")
	assert.Contains(t, output, "Show help information")
}

func TestHelpCommand_Execute_SpecificCommand(t *testing.T) {
	// Create a mock registry with some commands
	registry := newMockRegistry()
	err := registry.Register(&mockCommand{
		name:        "version",
		aliases:     []string{"-v", "--version"},
		description: "Show version information",
		usage:       "version",
	})
	require.NoError(t, err)

	cmd := NewHelpCommand(registry, "1.0.0")

	// Test showing help for a migrated command
	t.Run("migrated command", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := cmd.Execute(context.Background(), nil, []string{"version"})
		require.NoError(t, err)

		err = w.Close()
		require.NoError(t, err)
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, err = io.Copy(&buf, r)
		require.NoError(t, err)
		output := buf.String()

		assert.Contains(t, output, "Command: version")
		assert.Contains(t, output, "Description: Show version information")
		assert.Contains(t, output, "Usage: ticketflow version")
		assert.Contains(t, output, "Aliases: -v, --version")
	})

	// Test showing help for an unmigrated command
	t.Run("unmigrated command", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := cmd.Execute(context.Background(), nil, []string{"init"})
		require.NoError(t, err)

		err = w.Close()
		require.NoError(t, err)
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, err = io.Copy(&buf, r)
		require.NoError(t, err)
		output := buf.String()

		assert.Contains(t, output, "Command: init")
		assert.Contains(t, output, "Use 'ticketflow help' to see available options")
	})

	// Test showing help for an unknown command
	t.Run("unknown command", func(t *testing.T) {
		err := cmd.Execute(context.Background(), nil, []string{"unknown"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown command: unknown")
	})
}

func TestHelpCommand_GetMigratedCommands(t *testing.T) {
	t.Run("with nil registry", func(t *testing.T) {
		cmd := &HelpCommand{registry: nil, version: "1.0.0"}
		commands := cmd.getMigratedCommands()
		assert.Empty(t, commands)
	})

	t.Run("with empty registry", func(t *testing.T) {
		registry := newMockRegistry()
		cmd := &HelpCommand{registry: registry, version: "1.0.0"}
		commands := cmd.getMigratedCommands()
		assert.Empty(t, commands)
	})

	t.Run("with commands", func(t *testing.T) {
		registry := newMockRegistry()
		require.NoError(t, registry.Register(&mockCommand{name: "zebra", description: "Z command"}))
		require.NoError(t, registry.Register(&mockCommand{name: "alpha", description: "A command"}))
		require.NoError(t, registry.Register(&mockCommand{name: "beta", description: "B command"}))

		cmd := &HelpCommand{registry: registry, version: "1.0.0"}
		commands := cmd.getMigratedCommands()

		assert.Len(t, commands, 3)
		// Check that commands are sorted alphabetically
		assert.Equal(t, "alpha", commands[0].Name())
		assert.Equal(t, "beta", commands[1].Name())
		assert.Equal(t, "zebra", commands[2].Name())
	})
}

func TestHelpCommand_ShowCommandHelp(t *testing.T) {
	registry := newMockRegistry()
	err := registry.Register(&mockCommand{
		name:        "test",
		aliases:     []string{"-t", "--test"},
		description: "Test command",
		usage:       "test [options]",
	})
	require.NoError(t, err)

	cmd := &HelpCommand{registry: registry, version: "1.0.0"}

	t.Run("existing command", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := cmd.showCommandHelp("test")
		require.NoError(t, err)

		err = w.Close()
		require.NoError(t, err)
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, err = io.Copy(&buf, r)
		require.NoError(t, err)
		output := buf.String()

		assert.Contains(t, output, "Command: test")
		assert.Contains(t, output, "Description: Test command")
		assert.Contains(t, output, "Usage: ticketflow test [options]")
		assert.Contains(t, output, "Aliases: -t, --test")
	})

	t.Run("command via alias", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := cmd.showCommandHelp("-t")
		require.NoError(t, err)

		err = w.Close()
		require.NoError(t, err)
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, err = io.Copy(&buf, r)
		require.NoError(t, err)
		output := buf.String()

		assert.Contains(t, output, "Command: test")
	})

	t.Run("unknown command", func(t *testing.T) {
		err := cmd.showCommandHelp("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown command: nonexistent")
	})
}

func TestHelpCommand_OutputFormat(t *testing.T) {
	registry := newMockRegistry()
	cmd := NewHelpCommand(registry, "2.0.0")

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Execute(context.Background(), nil, []string{})
	require.NoError(t, err)

	err = w.Close()
	require.NoError(t, err)
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	// Verify the output structure
	lines := strings.Split(output, "\n")

	// Check first line contains version
	assert.True(t, strings.HasPrefix(lines[0], "TicketFlow"))
	assert.Contains(t, lines[0], "v2.0.0")

	// Check major sections are present
	var hasUsage, hasCommands, hasOptions, hasExamples bool
	for _, line := range lines {
		if line == "USAGE:" {
			hasUsage = true
		}
		if line == "COMMANDS:" {
			hasCommands = true
		}
		if line == "OPTIONS:" {
			hasOptions = true
		}
		if line == "EXAMPLES:" {
			hasExamples = true
		}
	}

	assert.True(t, hasUsage, "Should have USAGE section")
	assert.True(t, hasCommands, "Should have COMMANDS section")
	assert.True(t, hasOptions, "Should have OPTIONS section")
	assert.True(t, hasExamples, "Should have EXAMPLES section")
}
