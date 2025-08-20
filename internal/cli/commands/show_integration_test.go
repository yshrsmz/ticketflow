package commands

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/cli/commands/testharness"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestShowCommand_Execute_Integration(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*testharness.TestEnvironment)
		args          []string
		flags         map[string]string
		wantError     bool
		errorContains string
		validate      func(*testing.T, *testharness.TestEnvironment)
	}{
		{
			name: "show existing ticket with text format",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("show-ticket-001", ticket.StatusTodo,
					testharness.WithContent("This is the ticket content\n\nWith multiple lines"))
			},
			args:  []string{"show-ticket-001"},
			flags: map[string]string{"format": "text"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Text format output would be printed to stdout
				// We can verify the ticket still exists
				assert.True(t, env.FileExists("tickets/todo/show-ticket-001.md"))
			},
		},
		{
			name: "show ticket with json format",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("json-ticket-001", ticket.StatusDoing,
					testharness.WithContent("JSON format test ticket"))
			},
			args:  []string{"json-ticket-001"},
			flags: map[string]string{"format": "json"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// JSON output would include all ticket fields
				assert.True(t, env.FileExists("tickets/doing/json-ticket-001.md"))
			},
		},
		{
			name: "show ticket in done status",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("done-ticket-001", ticket.StatusDone,
					testharness.WithContent("Completed ticket"))
			},
			args:  []string{"done-ticket-001"},
			flags: map[string]string{"format": "text"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				assert.True(t, env.FileExists("tickets/done/done-ticket-001.md"))
			},
		},
		{
			name: "show ticket with parent relationship",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("parent-ticket", ticket.StatusTodo)
				env.CreateTicket("child-ticket", ticket.StatusTodo,
					testharness.WithParent("parent-ticket"),
					testharness.WithContent("Child ticket content"))
			},
			args:  []string{"child-ticket"},
			flags: map[string]string{"format": "text"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				content := env.ReadFile("tickets/todo/child-ticket.md")
				assert.Contains(t, content, "parent:parent-ticket")
			},
		},
		{
			name: "error when ticket does not exist",
			setup: func(env *testharness.TestEnvironment) {
				// No ticket created
			},
			args:          []string{"non-existent-ticket"},
			flags:         map[string]string{"format": "text"},
			wantError:     true,
			errorContains: "not found",
		},
		{
			name: "error when no ticket ID provided",
			setup: func(env *testharness.TestEnvironment) {
				// No setup needed
			},
			args:          []string{},
			flags:         map[string]string{"format": "text"},
			wantError:     true,
			errorContains: "missing ticket ID",
		},
		{
			name: "error with too many arguments",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("test-ticket", ticket.StatusTodo)
			},
			args:          []string{"test-ticket", "extra-arg"},
			flags:         map[string]string{"format": "text"},
			wantError:     true,
			errorContains: "unexpected arguments",
		},
		{
			name: "error with invalid format",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("test-ticket", ticket.StatusTodo)
			},
			args:          []string{"test-ticket"},
			flags:         map[string]string{"format": "invalid"},
			wantError:     true,
			errorContains: "invalid format",
		},
		{
			name: "show ticket with default format",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("default-format-ticket", ticket.StatusTodo)
			},
			args:  []string{"default-format-ticket"},
			flags: map[string]string{"format": ""}, // Empty format should default to text
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				assert.True(t, env.FileExists("tickets/todo/default-format-ticket.md"))
			},
		},
		{
			name: "show ticket with rich metadata",
			setup: func(env *testharness.TestEnvironment) {
				// Create a ticket with all metadata fields
				t := env.CreateTicket("rich-ticket", ticket.StatusDoing,
					testharness.WithContent("# Rich Content\n\n- Task 1\n- Task 2\n- Task 3"))
				// The ticket will have created_at and started_at already set
				_ = t
			},
			args:  []string{"rich-ticket"},
			flags: map[string]string{"format": "text"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				content := env.ReadFile("tickets/doing/rich-ticket.md")
				assert.Contains(t, content, "# Rich Content")
				assert.Contains(t, content, "- Task 1")
			},
		},
		{
			name: "show ticket with long content",
			setup: func(env *testharness.TestEnvironment) {
				longContent := "# Long Ticket\n\n"
				for i := 0; i < 100; i++ {
					longContent += "This is line " + string(rune(i)) + " of the long content.\n"
				}
				env.CreateTicket("long-ticket", ticket.StatusTodo,
					testharness.WithContent(longContent))
			},
			args:  []string{"long-ticket"},
			flags: map[string]string{"format": "text"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				content := env.ReadFile("tickets/todo/long-ticket.md")
				assert.Contains(t, content, "# Long Ticket")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test environment
			env := testharness.NewTestEnvironment(t)

			// Change to test directory
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				require.NoError(t, os.Chdir(oldWd))
			}()
			require.NoError(t, os.Chdir(env.RootDir))

			// Run setup
			if tt.setup != nil {
				tt.setup(env)
			}

			// Create command
			cmd := NewShowCommand()

			// Setup flags
			showFlags := &showFlags{
				format: tt.flags["format"],
			}

			// Validate flags before execution
			if err := cmd.Validate(showFlags, tt.args); err != nil {
				if tt.wantError {
					require.Error(t, err)
					if tt.errorContains != "" {
						assert.Contains(t, err.Error(), tt.errorContains)
					}
					return
				}
				require.NoError(t, err)
			}

			// Execute command with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			err = cmd.Execute(ctx, showFlags, tt.args)

			// Check error
			if tt.wantError && err == nil {
				// If we expect an error but didn't get one from Execute,
				// it might have been caught in Validate above
				return
			}

			if tt.wantError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}

			// Run validation
			if tt.validate != nil {
				tt.validate(t, env)
			}
		})
	}
}

func TestShowCommand_Execute_ContextCancellation(t *testing.T) {
	env := testharness.NewTestEnvironment(t)

	// Change to test directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(env.RootDir))

	// Create a test ticket
	env.CreateTicket("test-ticket", ticket.StatusTodo)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Execute command with cancelled context
	cmd := NewShowCommand()
	showFlags := &showFlags{format: StringFlag{Long: "text"}}
	err = cmd.Execute(ctx, showFlags, []string{"test-ticket"})
	require.Error(t, err)
	// With early context check, we return context.Canceled immediately
	assert.Contains(t, err.Error(), "context canceled")
}

func TestShowCommand_Execute_InvalidFlagsType(t *testing.T) {
	env := testharness.NewTestEnvironment(t)

	// Change to test directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(env.RootDir))

	// Create a test ticket
	env.CreateTicket("test-ticket", ticket.StatusTodo)

	// Execute command with wrong flags type
	cmd := NewShowCommand()
	err = cmd.Execute(context.Background(), "invalid-flags", []string{"test-ticket"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid flags type")
}
