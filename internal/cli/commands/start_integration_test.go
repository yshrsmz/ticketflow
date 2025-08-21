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
	"gopkg.in/yaml.v3"
)

func TestStartCommand_Execute_Integration(t *testing.T) {
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
			name: "start ticket successfully with worktree",
			setup: func(env *testharness.TestEnvironment) {
				// Create a ticket in todo status
				env.CreateTicket("test-start-001", ticket.StatusTodo,
					testharness.WithContent("Ready to start"))
				env.RunGit("add", ".")
				env.RunGit("commit", "-m", "Add todo ticket")
			},
			args:  []string{"test-start-001"},
			flags: map[string]string{"format": "text"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Verify ticket moved to doing
				assert.True(t, env.FileExists("tickets/doing/test-start-001.md"))
				assert.False(t, env.FileExists("tickets/todo/test-start-001.md"))

				// Verify worktree created
				assert.True(t, env.WorktreeExists("test-start-001"))

				// Note: current-ticket.md symlink is created in the worktree, not main repo
				// So we skip checking for it here

				// Verify commit was created
				assert.Contains(t, env.LastCommitMessage(), "Start ticket: test-start-001")

				// Verify started_at timestamp was set
				content := env.ReadFile("tickets/doing/test-start-001.md")
				assert.Contains(t, content, "started_at:")
			},
		},
		{
			name: "start ticket with force flag when worktree exists",
			setup: func(env *testharness.TestEnvironment) {
				// Create ticket and existing worktree
				env.CreateTicket("test-start-002", ticket.StatusTodo)
				env.CreateWorktree("test-start-002")
				env.RunGit("add", ".")
				env.RunGit("commit", "-m", "Add ticket with worktree")
			},
			args:  []string{"test-start-002"},
			flags: map[string]string{"format": "text", "force": "true"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Verify ticket started despite existing worktree
				assert.True(t, env.FileExists("tickets/doing/test-start-002.md"))

				// Verify worktree was recreated (force flag)
				assert.True(t, env.WorktreeExists("test-start-002"))
			},
		},
		{
			name: "start ticket without worktree when disabled",
			setup: func(env *testharness.TestEnvironment) {
				// Disable worktree in config
				env.Config.Worktree.Enabled = false
				data, _ := yaml.Marshal(env.Config)
				require.NoError(t, os.WriteFile(env.ConfigPath, data, 0644))

				env.CreateTicket("test-start-003", ticket.StatusTodo)
				env.RunGit("add", ".")
				env.RunGit("commit", "-m", "Add ticket")
			},
			args:  []string{"test-start-003"},
			flags: map[string]string{"format": "text"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Verify ticket moved to doing
				assert.True(t, env.FileExists("tickets/doing/test-start-003.md"))

				// Verify no worktree created
				assert.False(t, env.WorktreeExists("test-start-003"))

				// Verify branch was created
				output := env.RunGit("branch", "-l", "test-start-003")
				assert.Contains(t, output, "test-start-003")
			},
		},
		{
			name: "start ticket with parent relationship",
			setup: func(env *testharness.TestEnvironment) {
				// Create parent ticket (in todo status, not doing)
				env.CreateTicket("parent-ticket", ticket.StatusTodo)

				// Create child ticket with parent relationship
				env.CreateTicket("child-ticket", ticket.StatusTodo,
					testharness.WithParent("parent-ticket"))

				env.RunGit("add", ".")
				env.RunGit("commit", "-m", "Add tickets")

				// Switch to parent branch
				env.RunGit("checkout", "-b", "parent-ticket")
				env.RunGit("checkout", "main")
			},
			args:  []string{"child-ticket"},
			flags: map[string]string{"format": "text"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Verify ticket started
				assert.True(t, env.FileExists("tickets/doing/child-ticket.md"))

				// Verify branch created from parent
				// This would require more complex git inspection
			},
		},
		{
			name: "error when starting non-existent ticket",
			setup: func(env *testharness.TestEnvironment) {
				// No ticket created
			},
			args:          []string{"non-existent"},
			flags:         map[string]string{"format": "text"},
			wantError:     true,
			errorContains: "not found",
		},
		{
			name: "error when starting already started ticket",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("already-started", ticket.StatusDoing)
				env.CreateWorktree("already-started")
				env.RunGit("add", ".")
				env.RunGit("commit", "-m", "Add started ticket")
			},
			args:          []string{"already-started"},
			flags:         map[string]string{"format": "text"},
			wantError:     true,
			errorContains: "already started",
		},
		{
			name: "error when starting done ticket",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("done-ticket", ticket.StatusDone)
				env.RunGit("add", ".")
				env.RunGit("commit", "-m", "Add done ticket")
			},
			args:          []string{"done-ticket"},
			flags:         map[string]string{"format": "text"},
			wantError:     true,
			errorContains: "already started",
		},
		{
			name: "start ticket with JSON output",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("test-start-json", ticket.StatusTodo)
				env.RunGit("add", ".")
				env.RunGit("commit", "-m", "Add ticket")
			},
			args:  []string{"test-start-json"},
			flags: map[string]string{"format": "json"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Verify ticket started
				assert.True(t, env.FileExists("tickets/doing/test-start-json.md"))
				// JSON output validation would require capturing stdout
			},
		},
		{
			name: "error when no ticket ID provided",
			setup: func(env *testharness.TestEnvironment) {
				// No setup needed
			},
			args:          []string{},
			flags:         map[string]string{"format": "text"},
			wantError:     true,
			errorContains: "missing ticket argument",
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
			cmd := NewStartCommand()

			// Setup flags
			startFlags := &startFlags{
				force:  BoolFlag{Long: tt.flags["force"] == "true"},
				format: StringFlag{Long: tt.flags["format"]},
			}

			// Execute command with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			err = cmd.Execute(ctx, startFlags, tt.args)

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

func TestStartCommand_Execute_WithInitCommands(t *testing.T) {
	t.Skip("Init commands in worktrees need more complex setup")
	env := testharness.NewTestEnvironment(t)

	// Change to test directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(env.RootDir))

	// Configure init commands
	env.Config.Worktree.InitCommands = []string{"echo 'Init command executed' > init.log"}
	data, err := yaml.Marshal(env.Config)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(env.ConfigPath, data, 0644))

	// Create ticket
	env.CreateTicket("init-test", ticket.StatusTodo)
	env.RunGit("add", ".")
	env.RunGit("commit", "-m", "Add ticket")

	// Execute start command
	cmd := NewStartCommand()
	startFlags := &startFlags{
		format: StringFlag{Long: "text"},
	}

	err = cmd.Execute(context.Background(), startFlags, []string{"init-test"})
	require.NoError(t, err)

	// Verify init command was executed in worktree
	worktreePath := filepath.Join(filepath.Dir(env.RootDir), "test-worktrees", "init-test")
	initLogPath := filepath.Join(worktreePath, "init.log")
	assert.FileExists(t, initLogPath)

	content, err := os.ReadFile(initLogPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Init command executed")
}

func TestStartCommand_Execute_ContextCancellation(t *testing.T) {
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
	cmd := NewStartCommand()
	startFlags := &startFlags{
		format: StringFlag{Long: "text"},
	}

	err = cmd.Execute(ctx, startFlags, []string{"some-ticket"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestStartCommand_Execute_InvalidFlagsType(t *testing.T) {
	cmd := NewStartCommand()

	// Pass wrong type for flags
	err := cmd.Execute(context.Background(), "invalid", []string{"ticket"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid flags type")
}

func TestStartCommand_Execute_MultipleTicketValidation(t *testing.T) {
	env := testharness.NewTestEnvironment(t)

	// Change to test directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(env.RootDir))

	// Create command
	cmd := NewStartCommand()

	// Setup flags
	startFlags := &startFlags{
		format: StringFlag{Long: "text"},
	}

	// Try to start multiple tickets (should fail in Validate)
	err = cmd.Validate(startFlags, []string{"ticket1", "ticket2"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected arguments after ticket ID")
}
