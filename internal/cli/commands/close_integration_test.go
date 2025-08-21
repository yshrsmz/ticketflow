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

func TestCloseCommand_Execute_Integration(t *testing.T) {
	// Integration tests run sequentially to avoid conflicts

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
			name: "close current ticket successfully",
			setup: func(env *testharness.TestEnvironment) {
				// Create a ticket in doing status
				env.CreateTicket("test-ticket-001", ticket.StatusDoing,
					testharness.WithContent("Test ticket content"))
				// Create and switch to ticket branch
				env.RunGit("checkout", "-b", "test-ticket-001")
				// Stage ALL files including the symlink
				env.RunGit("add", ".")
				env.RunGit("commit", "-m", "Start ticket")
			},
			args:  []string{},
			flags: map[string]string{"format": "text"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Verify ticket moved to done
				assert.True(t, env.FileExists("tickets/done/test-ticket-001.md"))
				assert.False(t, env.FileExists("tickets/doing/test-ticket-001.md"))

				// Verify commit was created
				assert.Contains(t, env.LastCommitMessage(), "Close ticket: test-ticket-001")

				// Verify current-ticket.md symlink removed
				assert.False(t, env.FileExists("current-ticket.md"))
			},
		},
		{
			name: "close specific ticket by ID",
			setup: func(env *testharness.TestEnvironment) {
				// Create a ticket in doing status
				env.CreateTicket("test-ticket-002", ticket.StatusDoing)
				// Create and switch to ticket branch
				env.RunGit("checkout", "-b", "test-ticket-002")
				env.RunGit("add", ".")
				env.RunGit("commit", "-m", "Add ticket")
			},
			args:  []string{"test-ticket-002"},
			flags: map[string]string{"format": "text"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Verify ticket moved to done
				assert.True(t, env.FileExists("tickets/done/test-ticket-002.md"))
				assert.False(t, env.FileExists("tickets/doing/test-ticket-002.md"))
			},
		},
		{
			name: "close ticket with reason",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("test-ticket-003", ticket.StatusDoing)
				// Create and switch to ticket branch
				env.RunGit("checkout", "-b", "test-ticket-003")
				env.RunGit("add", ".")
				env.RunGit("commit", "-m", "Add ticket")
			},
			args:  []string{},
			flags: map[string]string{"format": "text", "reason": "Task completed successfully"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Verify ticket has closure reason
				content := env.ReadFile("tickets/done/test-ticket-003.md")
				assert.Contains(t, content, "closure_reason: Task completed successfully")

				// Verify commit message includes reason
				assert.Contains(t, env.LastCommitMessage(), "Task completed successfully")
			},
		},
		{
			name: "close ticket with force flag and uncommitted changes",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("test-ticket-004", ticket.StatusDoing)
				// Create and switch to ticket branch
				env.RunGit("checkout", "-b", "test-ticket-004")
				env.RunGit("add", ".")
				env.RunGit("commit", "-m", "Add ticket")

				// Make uncommitted changes
				env.WriteFile("test.txt", "uncommitted changes")
			},
			args:  []string{},
			flags: map[string]string{"format": "text", "force": "true"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Verify ticket moved despite uncommitted changes
				assert.True(t, env.FileExists("tickets/done/test-ticket-004.md"))

				// The force flag allows closing with uncommitted changes,
				// but doesn't automatically commit them
				assert.True(t, env.HasUncommittedChanges())
			},
		},
		{
			name: "error when closing non-existent ticket",
			setup: func(env *testharness.TestEnvironment) {
				// No ticket created
			},
			args:          []string{"non-existent-ticket"},
			flags:         map[string]string{"format": "text"},
			wantError:     true,
			errorContains: "not found",
		},
		{
			name: "error when closing ticket not in doing status",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("test-ticket-005", ticket.StatusTodo)
				env.RunGit("add", ".")
				env.RunGit("commit", "-m", "Add todo ticket")
			},
			args:          []string{"test-ticket-005"},
			flags:         map[string]string{"format": "text"},
			wantError:     true,
			errorContains: "Reason required",
		},
		{
			name: "error when no current ticket",
			setup: func(env *testharness.TestEnvironment) {
				// No current ticket
			},
			args:          []string{},
			flags:         map[string]string{"format": "text"},
			wantError:     true,
			errorContains: "No active ticket",
		},
		{
			name: "close ticket with JSON output",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("test-ticket-006", ticket.StatusDoing)
				// Create and switch to ticket branch
				env.RunGit("checkout", "-b", "test-ticket-006")
				env.RunGit("add", ".")
				env.RunGit("commit", "-m", "Add ticket")
			},
			args:  []string{},
			flags: map[string]string{"format": "json"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Verify ticket moved
				assert.True(t, env.FileExists("tickets/done/test-ticket-006.md"))
				// JSON output validation would require capturing stdout
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
			cmd := NewCloseCommand()

			// Setup flags
			closeFlags := &closeFlags{
				force:  BoolFlag{Long: tt.flags["force"] == "true"},
				reason: tt.flags["reason"],
				format: StringFlag{Long: tt.flags["format"]},
				args:   tt.args,
			}

			// Execute command with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			err = cmd.Execute(ctx, closeFlags, tt.args)

			// Check error
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

func TestCloseCommand_Execute_WithWorktree(t *testing.T) {
	env := testharness.NewTestEnvironment(t)

	// Change to test directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(env.RootDir))

	// Create ticket and worktree
	env.CreateTicket("worktree-ticket", ticket.StatusDoing)
	// Create branch and switch to it
	env.RunGit("checkout", "-b", "worktree-ticket")
	env.RunGit("add", ".")
	env.RunGit("commit", "-m", "Start ticket with worktree")
	// Note: We're already on the ticket branch, no worktree needed for this test
	// The close command should work from the main repo branch

	// Execute close command
	cmd := NewCloseCommand()
	closeFlags := &closeFlags{
		format: StringFlag{Long: "text"},
		args:   []string{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = cmd.Execute(ctx, closeFlags, []string{})
	require.NoError(t, err)

	// Verify ticket closed
	assert.True(t, env.FileExists("tickets/done/worktree-ticket.md"))
	assert.False(t, env.FileExists("tickets/doing/worktree-ticket.md"))
}

func TestCloseCommand_Execute_ContextCancellation(t *testing.T) {
	env := testharness.NewTestEnvironment(t)

	// Change to test directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(env.RootDir))

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Execute command with cancelled context
	cmd := NewCloseCommand()
	closeFlags := &closeFlags{
		format: StringFlag{Long: "text"},
		args:   []string{},
	}

	err = cmd.Execute(ctx, closeFlags, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestCloseCommand_Execute_InvalidFlagsType(t *testing.T) {
	cmd := NewCloseCommand()

	// Pass wrong type for flags
	err := cmd.Execute(context.Background(), "invalid", []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid flags type")
}
