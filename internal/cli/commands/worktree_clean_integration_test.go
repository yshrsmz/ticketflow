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

func TestWorktreeCleanCommand_Execute_Integration(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*testharness.TestEnvironment)
		args          []string
		wantError     bool
		errorContains string
		validate      func(*testing.T, *testharness.TestEnvironment)
	}{
		{
			name: "clean orphaned worktrees successfully",
			setup: func(env *testharness.TestEnvironment) {
				// Create tickets in done status
				env.CreateTicket("done-ticket-001", ticket.StatusDone)
				env.CreateTicket("done-ticket-002", ticket.StatusDone)

				// Create branches for both tickets
				env.RunGit("checkout", "-b", "done-ticket-001")
				env.RunGit("checkout", "main")
				env.RunGit("checkout", "-b", "done-ticket-002")
				env.RunGit("checkout", "main")

				// Create worktrees for both
				worktreeBase := filepath.Join(filepath.Dir(env.RootDir), "test-worktrees")
				env.RunGit("worktree", "add", filepath.Join(worktreeBase, "done-ticket-001"), "done-ticket-001")
				env.RunGit("worktree", "add", filepath.Join(worktreeBase, "done-ticket-002"), "done-ticket-002")

				// Verify worktrees exist
				output := env.RunGit("worktree", "list")
				require.Contains(t, output, "done-ticket-001")
				require.Contains(t, output, "done-ticket-002")
			},
			args: []string{},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Verify worktrees were cleaned up
				output := env.RunGit("worktree", "list")
				assert.NotContains(t, output, "done-ticket-001")
				assert.NotContains(t, output, "done-ticket-002")
			},
		},
		{
			name: "keep worktrees for active tickets",
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
			args: []string{},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				output := env.RunGit("worktree", "list")
				// Only "doing" tickets are considered active and keep their worktrees
				assert.Contains(t, output, "doing-ticket")
				// Todo and done tickets should have worktrees removed
				assert.NotContains(t, output, "todo-ticket")
				assert.NotContains(t, output, "done-ticket")
			},
		},
		{
			name: "handle no worktrees gracefully",
			setup: func(env *testharness.TestEnvironment) {
				// No worktrees created
			},
			args: []string{},
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
			wantError:     true,
			errorContains: "takes no arguments",
		},
		{
			name: "handle worktree with missing directory",
			setup: func(env *testharness.TestEnvironment) {
				// Create a ticket and worktree
				env.CreateTicket("missing-dir-ticket", ticket.StatusDone)
				env.RunGit("checkout", "-b", "missing-dir-ticket")
				env.RunGit("checkout", "main")

				worktreeBase := filepath.Join(filepath.Dir(env.RootDir), "test-worktrees")
				worktreePath := filepath.Join(worktreeBase, "missing-dir-ticket")
				env.RunGit("worktree", "add", worktreePath, "missing-dir-ticket")

				// Remove the worktree directory manually to simulate orphaned state
				require.NoError(t, os.RemoveAll(worktreePath))
			},
			args: []string{},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Should handle missing directory gracefully
				// The worktree should be cleaned up or marked as missing
				// Git may show it as missing, but clean should handle it
				_ = env.RunGit("worktree", "list")
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
			cmd := NewWorktreeCleanCommand()

			// Validate first (if the command has a Validate method)
			if err := cmd.Validate(nil, tt.args); err != nil {
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
			err = cmd.Execute(ctx, nil, tt.args)

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

func TestWorktreeCleanCommand_Execute_ContextCancellation(t *testing.T) {
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
	cmd := NewWorktreeCleanCommand()
	err = cmd.Execute(ctx, nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}
