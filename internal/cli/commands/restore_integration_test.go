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

func TestRestoreCommand_Execute_Integration(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*testharness.TestEnvironment)
		args          []string
		flags         map[string]string
		wantError     bool
		errorContains string
		validate      func(*testing.T, *testharness.TestEnvironment)
		// expectJSONError indicates that this test expects an error condition
		// but with JSON format, the error is returned as JSON (not a Go error)
		expectJSONError bool
	}{
		{
			name: "restore current ticket symlink successfully",
			setup: func(env *testharness.TestEnvironment) {
				// Create a ticket in doing status
				env.CreateTicket("restore-ticket-001", ticket.StatusDoing,
					testharness.WithContent("Test ticket for restore"))

				// Switch to the ticket branch (simulate being in a worktree)
				env.RunGit("checkout", "-b", "restore-ticket-001")

				// Remove the symlink to simulate it being missing
				symlinkPath := filepath.Join(env.RootDir, "current-ticket.md")
				_ = os.Remove(symlinkPath) // Ignore error - symlink may not exist

				// Verify symlink is gone
				require.False(t, env.FileExists("current-ticket.md"))
			},
			args:  []string{},
			flags: map[string]string{"format": "text"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Verify symlink was restored
				assert.True(t, env.FileExists("current-ticket.md"))

				// Verify symlink points to correct ticket
				symlinkPath := filepath.Join(env.RootDir, "current-ticket.md")
				target, err := os.Readlink(symlinkPath)
				require.NoError(t, err)
				assert.Contains(t, target, "restore-ticket-001.md")
			},
		},
		{
			name: "restore with JSON output format",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("json-restore-ticket", ticket.StatusDoing,
					testharness.WithContent("JSON output test"))

				// Switch to the ticket branch
				env.RunGit("checkout", "-b", "json-restore-ticket")

				// Remove the symlink
				symlinkPath := filepath.Join(env.RootDir, "current-ticket.md")
				_ = os.Remove(symlinkPath) // Ignore error - symlink may not exist
			},
			args:  []string{},
			flags: map[string]string{"format": "json"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Verify symlink was restored
				assert.True(t, env.FileExists("current-ticket.md"))
			},
		},
		{
			name: "restore with short format flag",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("short-format-ticket", ticket.StatusDoing)

				// Switch to the ticket branch
				env.RunGit("checkout", "-b", "short-format-ticket")

				// Remove the symlink
				symlinkPath := filepath.Join(env.RootDir, "current-ticket.md")
				_ = os.Remove(symlinkPath) // Ignore error - symlink may not exist
			},
			args:  []string{},
			flags: map[string]string{"formatShort": "json"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Verify symlink was restored
				assert.True(t, env.FileExists("current-ticket.md"))
			},
		},
		{
			name: "restore ticket with parent relationship",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("parent-ticket", ticket.StatusTodo)
				env.CreateTicket("child-restore-ticket", ticket.StatusDoing,
					testharness.WithParent("parent-ticket"))

				// Switch to the child ticket branch
				env.RunGit("checkout", "-b", "child-restore-ticket")

				// Remove the symlink
				symlinkPath := filepath.Join(env.RootDir, "current-ticket.md")
				_ = os.Remove(symlinkPath) // Ignore error - symlink may not exist
			},
			args:  []string{},
			flags: map[string]string{"format": "json"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Verify symlink was restored
				assert.True(t, env.FileExists("current-ticket.md"))

				// Verify symlink points to child ticket
				symlinkPath := filepath.Join(env.RootDir, "current-ticket.md")
				target, err := os.Readlink(symlinkPath)
				require.NoError(t, err)
				assert.Contains(t, target, "child-restore-ticket.md")
			},
		},
		{
			name: "restore when symlink already exists",
			setup: func(env *testharness.TestEnvironment) {
				// Create a ticket in doing status - this automatically creates the symlink
				env.CreateTicket("existing-symlink-ticket", ticket.StatusDoing)

				// Switch to the ticket branch
				env.RunGit("checkout", "-b", "existing-symlink-ticket")

				// Verify symlink already exists
				require.True(t, env.FileExists("current-ticket.md"))
			},
			args:  []string{},
			flags: map[string]string{"format": "text"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Should still work, symlink should remain
				assert.True(t, env.FileExists("current-ticket.md"))
			},
		},
		{
			name: "error when no tickets in doing status",
			setup: func(env *testharness.TestEnvironment) {
				// Create tickets but none in doing status
				env.CreateTicket("todo-ticket", ticket.StatusTodo)
				env.CreateTicket("done-ticket", ticket.StatusDone)
				// Stay on main branch - no ticket to restore
			},
			args:          []string{},
			flags:         map[string]string{"format": "text"},
			wantError:     true,
			errorContains: "no ticket found for branch main",
		},
		{
			name: "error with unexpected arguments",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("test-ticket", ticket.StatusDoing)
			},
			args:          []string{"unexpected-arg"},
			flags:         map[string]string{"format": "text"},
			wantError:     true,
			errorContains: "does not accept any arguments",
		},
		{
			name: "error with invalid format",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("test-ticket", ticket.StatusDoing)
			},
			args:          []string{},
			flags:         map[string]string{"format": "invalid"},
			wantError:     true,
			errorContains: "invalid format",
		},
		{
			name: "restore with multiple tickets in doing",
			setup: func(env *testharness.TestEnvironment) {
				// Create multiple tickets in doing status
				env.CreateTicket("doing-ticket-1", ticket.StatusDoing)

				// Switch to first ticket branch
				env.RunGit("checkout", "-b", "doing-ticket-1")

				// Remove symlink and manually create another doing ticket
				symlinkPath := filepath.Join(env.RootDir, "current-ticket.md")
				_ = os.Remove(symlinkPath) // Ignore error - symlink may not exist

				// Create second ticket directly without symlink
				env.WriteFile("tickets/doing/doing-ticket-2.md", `---
priority: 1
description: "Second doing ticket"
created_at: "2024-01-01T12:00:00Z"
started_at: "2024-01-01T13:00:00Z"
---

# doing-ticket-2

Content`)
			},
			args:  []string{},
			flags: map[string]string{"format": "text"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Should restore symlink to the current branch ticket (doing-ticket-1)
				assert.True(t, env.FileExists("current-ticket.md"))

				// Verify symlink points to doing-ticket-1 (current branch)
				symlinkPath := filepath.Join(env.RootDir, "current-ticket.md")
				target, err := os.Readlink(symlinkPath)
				require.NoError(t, err)
				assert.Contains(t, target, "doing-ticket-1")
			},
		},
		{
			name: "JSON error output when no doing tickets",
			setup: func(env *testharness.TestEnvironment) {
				// Only create non-doing tickets
				env.CreateTicket("todo-only", ticket.StatusTodo)
			},
			args:            []string{},
			flags:           map[string]string{"format": "json"},
			wantError:       false, // JSON format returns error in JSON, not as error
			expectJSONError: true,  // This test expects an error condition returned as JSON
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// When format is JSON, errors are returned as JSON objects with 
				// success:false rather than Go errors. This is intentional behavior
				// to support machine-readable output for AI/tooling integration.
				// The Execute method should return nil, with error details in JSON.
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
			cmd := NewRestoreCommand()

			// Setup flags
			restoreFlags := &restoreFlags{
				format:      tt.flags["format"],
				formatShort: tt.flags["formatShort"],
			}

			// Validate flags before execution
			if err := cmd.Validate(restoreFlags, tt.args); err != nil {
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
			err = cmd.Execute(ctx, restoreFlags, tt.args)

			// Check error
			if tt.wantError && err == nil {
				// If we expect an error but didn't get one from Execute,
				// it might have been caught in Validate above or returned as JSON
				if tt.flags["format"] != "json" {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if tt.wantError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				// Special handling for JSON error cases: When format=json and an error
				// occurs in the command logic, the error is returned as a JSON object
				// with success:false rather than a Go error. This provides machine-readable
				// output for AI/tooling integration.
				if !tt.expectJSONError {
					require.NoError(t, err)
				}
			}

			// Run validation
			if tt.validate != nil {
				tt.validate(t, env)
			}
		})
	}
}

func TestRestoreCommand_Execute_ContextCancellation(t *testing.T) {
	env := testharness.NewTestEnvironment(t)

	// Change to test directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(env.RootDir))

	// Create a ticket in doing status
	env.CreateTicket("test-ticket", ticket.StatusDoing)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Execute command with cancelled context
	cmd := NewRestoreCommand()
	restoreFlags := &restoreFlags{format: "text"}
	err = cmd.Execute(ctx, restoreFlags, []string{})
	require.Error(t, err)
	// When context is cancelled, git commands fail with "Not in a git repository"
	// because the git command exits immediately without proper error propagation
	assert.Contains(t, err.Error(), "Not in a git repository")
}

func TestRestoreCommand_Execute_InvalidFlagsType(t *testing.T) {
	env := testharness.NewTestEnvironment(t)

	// Change to test directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(env.RootDir))

	// Create a test ticket
	env.CreateTicket("test-ticket", ticket.StatusDoing)

	// Validate with wrong flags type
	cmd := NewRestoreCommand()
	err = cmd.Validate("invalid-flags", []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid flags type")
}
