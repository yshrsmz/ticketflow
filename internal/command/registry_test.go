package command

import (
	"context"
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCommand is a test implementation of Command
type mockCommand struct {
	name        string
	aliases     []string
	description string
	usage       string
	executed    bool
}

func (m *mockCommand) Name() string        { return m.name }
func (m *mockCommand) Aliases() []string  { return m.aliases }
func (m *mockCommand) Description() string { return m.description }
func (m *mockCommand) Usage() string       { return m.usage }

func (m *mockCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	return struct{}{}
}

func (m *mockCommand) Validate(flags interface{}, args []string) error {
	return nil
}

func (m *mockCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	m.executed = true
	return nil
}

func TestRegistry_Register(t *testing.T) {
	tests := []struct {
		name    string
		command Command
		wantErr bool
		errMsg  string
	}{
		{
			name: "register valid command",
			command: &mockCommand{
				name:        "test",
				description: "Test command",
			},
			wantErr: false,
		},
		{
			name:    "register nil command",
			command: nil,
			wantErr: true,
			errMsg:  "cannot register nil command",
		},
		{
			name: "register command with empty name",
			command: &mockCommand{
				name:        "",
				description: "Test command",
			},
			wantErr: true,
			errMsg:  "cannot register command with empty name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry()
			err := r.Register(tt.command)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	r := NewRegistry()

	cmd1 := &mockCommand{name: "test", description: "First"}
	cmd2 := &mockCommand{name: "test", description: "Second"}

	// First registration should succeed
	err := r.Register(cmd1)
	require.NoError(t, err)

	// Second registration with same name should fail
	err = r.Register(cmd2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry()

	cmd := &mockCommand{
		name:        "test",
		description: "Test command",
	}

	// Register command
	err := r.Register(cmd)
	require.NoError(t, err)

	// Get existing command
	got, exists := r.Get("test")
	assert.True(t, exists)
	assert.Equal(t, cmd, got)

	// Get non-existing command
	got, exists = r.Get("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, got)
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()

	cmd1 := &mockCommand{name: "cmd1", description: "Command 1"}
	cmd2 := &mockCommand{name: "cmd2", description: "Command 2"}
	cmd3 := &mockCommand{name: "cmd3", description: "Command 3"}

	// Register commands
	require.NoError(t, r.Register(cmd1))
	require.NoError(t, r.Register(cmd2))
	require.NoError(t, r.Register(cmd3))

	// List should return all commands
	commands := r.List()
	assert.Len(t, commands, 3)

	// Check that all commands are present
	names := make(map[string]bool)
	for _, cmd := range commands {
		names[cmd.Name()] = true
	}

	assert.True(t, names["cmd1"])
	assert.True(t, names["cmd2"])
	assert.True(t, names["cmd3"])
}

func TestRegistry_Aliases(t *testing.T) {
	r := NewRegistry()

	// Register command with aliases
	cmd := &mockCommand{
		name:        "version",
		aliases:     []string{"-v", "--version"},
		description: "Show version",
	}

	err := r.Register(cmd)
	require.NoError(t, err)

	// Should get command by name
	got, exists := r.Get("version")
	assert.True(t, exists)
	assert.Equal(t, cmd, got)

	// Should get command by first alias
	got, exists = r.Get("-v")
	assert.True(t, exists)
	assert.Equal(t, cmd, got)

	// Should get command by second alias
	got, exists = r.Get("--version")
	assert.True(t, exists)
	assert.Equal(t, cmd, got)

	// Should not find non-existent alias
	got, exists = r.Get("-version")
	assert.False(t, exists)
	assert.Nil(t, got)
}

func TestRegistry_DuplicateAlias(t *testing.T) {
	r := NewRegistry()

	// Register first command with alias
	cmd1 := &mockCommand{
		name:    "version",
		aliases: []string{"-v"},
	}
	err := r.Register(cmd1)
	require.NoError(t, err)

	// Try to register second command with same alias
	cmd2 := &mockCommand{
		name:    "verbose",
		aliases: []string{"-v"},
	}
	err = r.Register(cmd2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "alias \"-v\" already registered")
}

func TestRegistry_ThreadSafety(t *testing.T) {
	r := NewRegistry()

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			cmd := &mockCommand{
				name:        string(rune('a' + n)),
				description: "Concurrent command",
			}
			_ = r.Register(cmd)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			r.List()
			r.Get("a")
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all commands were registered
	commands := r.List()
	assert.Len(t, commands, 10)
}
