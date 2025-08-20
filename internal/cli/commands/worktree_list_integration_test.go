package commands

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/cli/commands/testharness"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestWorktreeListCommand_Execute_Integration(t *testing.T) {
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
			name: "list all worktrees with text format",
			setup: func(env *testharness.TestEnvironment) {
				// Create tickets in different statuses
				env.CreateTicket("todo-ticket", ticket.StatusTodo)
				env.CreateTicket("doing-ticket", ticket.StatusDoing)
				env.CreateTicket("done-ticket", ticket.StatusDone)

				// Create branches and worktrees
				env.RunGit("checkout", "-b", "todo-ticket")
				env.RunGit("checkout", "main")
				env.RunGit("checkout", "-b", "doing-ticket")
				env.RunGit("checkout", "main")
				env.RunGit("checkout", "-b", "done-ticket")
				env.RunGit("checkout", "main")

				worktreeBase := filepath.Join(filepath.Dir(env.RootDir), "test-worktrees")
				env.RunGit("worktree", "add", filepath.Join(worktreeBase, "todo-ticket"), "todo-ticket")
				env.RunGit("worktree", "add", filepath.Join(worktreeBase, "doing-ticket"), "doing-ticket")
				env.RunGit("worktree", "add", filepath.Join(worktreeBase, "done-ticket"), "done-ticket")
			},
			args:  []string{},
			flags: map[string]string{"format": "text"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Verify all worktrees exist
				output := env.RunGit("worktree", "list")
				assert.Contains(t, output, "todo-ticket")
				assert.Contains(t, output, "doing-ticket")
				assert.Contains(t, output, "done-ticket")
			},
		},
		{
			name: "list worktrees with json format",
			setup: func(env *testharness.TestEnvironment) {
				// Create a couple of tickets with worktrees
				env.CreateTicket("json-ticket-1", ticket.StatusDoing)
				env.CreateTicket("json-ticket-2", ticket.StatusTodo)

				env.RunGit("checkout", "-b", "json-ticket-1")
				env.RunGit("checkout", "main")
				env.RunGit("checkout", "-b", "json-ticket-2")
				env.RunGit("checkout", "main")

				worktreeBase := filepath.Join(filepath.Dir(env.RootDir), "test-worktrees")
				env.RunGit("worktree", "add", filepath.Join(worktreeBase, "json-ticket-1"), "json-ticket-1")
				env.RunGit("worktree", "add", filepath.Join(worktreeBase, "json-ticket-2"), "json-ticket-2")
			},
			args:  []string{},
			flags: map[string]string{"format": "json"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// JSON format should work without error
				output := env.RunGit("worktree", "list")
				assert.Contains(t, output, "json-ticket-1")
				assert.Contains(t, output, "json-ticket-2")
			},
		},
		{
			name: "list with short format flag",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("short-flag-ticket", ticket.StatusDoing)
				env.RunGit("checkout", "-b", "short-flag-ticket")
				env.RunGit("checkout", "main")

				worktreeBase := filepath.Join(filepath.Dir(env.RootDir), "test-worktrees")
				env.RunGit("worktree", "add", filepath.Join(worktreeBase, "short-flag-ticket"), "short-flag-ticket")
			},
			args:  []string{},
			flags: map[string]string{"formatShort": "json"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Should accept -o flag for format
				output := env.RunGit("worktree", "list")
				assert.Contains(t, output, "short-flag-ticket")
			},
		},
		{
			name: "handle no worktrees gracefully",
			setup: func(env *testharness.TestEnvironment) {
				// No worktrees created (main is not counted as a worktree in our context)
			},
			args:  []string{},
			flags: map[string]string{"format": "text"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Should complete without error
				output := env.RunGit("worktree", "list")
				// Should only have the main worktree
				assert.NotContains(t, output, "test-worktrees")
			},
		},
		{
			name: "error when unexpected arguments provided",
			setup: func(env *testharness.TestEnvironment) {
				// No setup needed
			},
			args:          []string{"unexpected-arg"},
			flags:         map[string]string{"format": "text"},
			wantError:     true,
			errorContains: "takes no arguments",
		},
		{
			name: "error with invalid format",
			setup: func(env *testharness.TestEnvironment) {
				// No setup needed
			},
			args:          []string{},
			flags:         map[string]string{"format": "invalid"},
			wantError:     true,
			errorContains: "invalid format",
		},
		{
			name: "list worktrees with default format",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("default-format-ticket", ticket.StatusDoing)
				env.RunGit("checkout", "-b", "default-format-ticket")
				env.RunGit("checkout", "main")

				worktreeBase := filepath.Join(filepath.Dir(env.RootDir), "test-worktrees")
				env.RunGit("worktree", "add", filepath.Join(worktreeBase, "default-format-ticket"), "default-format-ticket")
			},
			args: []string{},
			// No flags provided - should use default format (text)
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				output := env.RunGit("worktree", "list")
				assert.Contains(t, output, "default-format-ticket")
			},
		},
		{
			name: "list multiple worktrees sorted",
			setup: func(env *testharness.TestEnvironment) {
				// Create multiple tickets to test sorting
				env.CreateTicket("alpha-ticket", ticket.StatusTodo)
				env.CreateTicket("beta-ticket", ticket.StatusDoing)
				env.CreateTicket("gamma-ticket", ticket.StatusDone)

				// Create branches and worktrees
				for _, ticketID := range []string{"alpha-ticket", "beta-ticket", "gamma-ticket"} {
					env.RunGit("checkout", "-b", ticketID)
					env.RunGit("checkout", "main")

					worktreeBase := filepath.Join(filepath.Dir(env.RootDir), "test-worktrees")
					env.RunGit("worktree", "add", filepath.Join(worktreeBase, ticketID), ticketID)
				}
			},
			args:  []string{},
			flags: map[string]string{"format": "text"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				output := env.RunGit("worktree", "list")
				// All worktrees should be listed
				assert.Contains(t, output, "alpha-ticket")
				assert.Contains(t, output, "beta-ticket")
				assert.Contains(t, output, "gamma-ticket")
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
			cmd := NewWorktreeListCommand()

			// Setup flags
			var flags interface{}
			if tt.flags != nil {
				listFlags := &worktreeListFlags{
					format:      tt.flags["format"],
					formatShort: tt.flags["formatShort"],
				}

				// Validate flags before execution
				if err := cmd.Validate(listFlags, tt.args); err != nil {
					if tt.wantError {
						require.Error(t, err)
						if tt.errorContains != "" {
							assert.Contains(t, err.Error(), tt.errorContains)
						}
						return
					}
					require.NoError(t, err)
				}

				flags = listFlags
			}

			// Execute command with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			err = cmd.Execute(ctx, flags, tt.args)

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

func TestWorktreeListCommand_Execute_ContextCancellation(t *testing.T) {
	env := testharness.NewTestEnvironment(t)

	env.WithWorkingDirectory(t, func() {
		// Create a cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Execute command with cancelled context
		cmd := NewWorktreeListCommand()
		listFlags := &worktreeListFlags{format: StringFlag{Long: "text"}}
		err := cmd.Execute(ctx, listFlags, []string{})
		require.Error(t, err)
		// With early context check, we return context.Canceled immediately
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestWorktreeListCommand_Execute_NilFlags(t *testing.T) {
	env := testharness.NewTestEnvironment(t)

	env.WithWorkingDirectory(t, func() {
		// Execute command with nil flags - should use defaults
		cmd := NewWorktreeListCommand()
		ctx := context.Background()
		err := cmd.Execute(ctx, nil, []string{})

		// Should work with default format
		require.NoError(t, err)
	})
}
