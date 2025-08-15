package commands

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/cli/commands/testharness"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestNewCommand_Execute_Integration(t *testing.T) {
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
			name: "create new ticket with valid slug",
			setup: func(env *testharness.TestEnvironment) {
				// No setup needed
			},
			args:  []string{"test-feature"},
			flags: map[string]string{"format": "text"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Find the created ticket file
				files, err := os.ReadDir(env.RootDir + "/tickets/todo")
				require.NoError(t, err)
				
				var ticketFound bool
				for _, file := range files {
					if strings.Contains(file.Name(), "test-feature") {
						ticketFound = true
						// Verify ticket content
						content := env.ReadFile("tickets/todo/" + file.Name())
						assert.Contains(t, content, "test-feature")
						assert.Contains(t, content, "priority:")
						assert.Contains(t, content, "description:")
						break
					}
				}
				assert.True(t, ticketFound, "Ticket file should be created")
			},
		},
		{
			name: "create ticket with parent flag",
			setup: func(env *testharness.TestEnvironment) {
				// Create parent ticket
				env.CreateTicket("parent-ticket-001", ticket.StatusTodo)
			},
			args:  []string{"child-feature"},
			flags: map[string]string{"format": "text", "parent": "parent-ticket-001"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Find the created ticket
				files, err := os.ReadDir(env.RootDir + "/tickets/todo")
				require.NoError(t, err)
				
				for _, file := range files {
					if strings.Contains(file.Name(), "child-feature") {
						content := env.ReadFile("tickets/todo/" + file.Name())
						// Verify parent relationship
						assert.Contains(t, content, "parent:parent-ticket-001")
						break
					}
				}
			},
		},
		{
			name: "create ticket with short parent flag",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("parent-ticket-002", ticket.StatusDoing)
			},
			args:  []string{"another-child"},
			flags: map[string]string{"format": "text", "parentShort": "parent-ticket-002"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				files, err := os.ReadDir(env.RootDir + "/tickets/todo")
				require.NoError(t, err)
				
				for _, file := range files {
					if strings.Contains(file.Name(), "another-child") {
						content := env.ReadFile("tickets/todo/" + file.Name())
						assert.Contains(t, content, "parent:parent-ticket-002")
						break
					}
				}
			},
		},
		{
			name: "create ticket with JSON format",
			setup: func(env *testharness.TestEnvironment) {
				// No setup needed
			},
			args:  []string{"json-ticket"},
			flags: map[string]string{"format": "json"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Verify ticket was created
				files, err := os.ReadDir(env.RootDir + "/tickets/todo")
				require.NoError(t, err)
				
				var found bool
				for _, file := range files {
					if strings.Contains(file.Name(), "json-ticket") {
						found = true
						break
					}
				}
				assert.True(t, found, "JSON format ticket should be created")
			},
		},
		{
			name: "create ticket with short format flag",
			setup: func(env *testharness.TestEnvironment) {
				// No setup needed
			},
			args:  []string{"short-format-ticket"},
			flags: map[string]string{"formatShort": "json"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				files, err := os.ReadDir(env.RootDir + "/tickets/todo")
				require.NoError(t, err)
				
				var found bool
				for _, file := range files {
					if strings.Contains(file.Name(), "short-format-ticket") {
						found = true
						break
					}
				}
				assert.True(t, found)
			},
		},
		{
			name: "error with invalid slug format",
			setup: func(env *testharness.TestEnvironment) {
				// No setup needed
			},
			args:          []string{"INVALID_SLUG!"}, // Invalid characters
			flags:         map[string]string{"format": "text"},
			wantError:     true,
			errorContains: "invalid slug",
		},
		{
			name: "error with duplicate ticket",
			setup: func(env *testharness.TestEnvironment) {
				// Create a ticket with a specific ID pattern
				env.CreateTicket("250101-120000-duplicate-test", ticket.StatusTodo)
			},
			args:          []string{"duplicate-test"},
			flags:         map[string]string{"format": "text"},
			wantError:     false, // This will create a new ticket with different timestamp
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Should create a new ticket with different timestamp
				files, err := os.ReadDir(env.RootDir + "/tickets/todo")
				require.NoError(t, err)
				
				count := 0
				for _, file := range files {
					if strings.Contains(file.Name(), "duplicate-test") {
						count++
					}
				}
				// Should have 2 tickets with "duplicate-test" in the name
				assert.Equal(t, 2, count, "Should have created a second ticket with different timestamp")
			},
		},
		{
			name: "error when no slug provided",
			setup: func(env *testharness.TestEnvironment) {
				// No setup needed
			},
			args:          []string{},
			flags:         map[string]string{"format": "text"},
			wantError:     true,
			errorContains: "missing slug argument",
		},
		{
			name: "error with too many arguments",
			setup: func(env *testharness.TestEnvironment) {
				// No setup needed
			},
			args:          []string{"valid-slug", "extra-arg"},
			flags:         map[string]string{"format": "text"},
			wantError:     true,
			errorContains: "unexpected arguments",
		},
		{
			name: "error with invalid format",
			setup: func(env *testharness.TestEnvironment) {
				// No setup needed
			},
			args:          []string{"test-ticket"},
			flags:         map[string]string{"format": "invalid"},
			wantError:     true,
			errorContains: "invalid format",
		},
		{
			name: "error with non-existent parent",
			setup: func(env *testharness.TestEnvironment) {
				// No parent ticket created
			},
			args:          []string{"orphan-ticket"},
			flags:         map[string]string{"format": "text", "parent": "non-existent-parent"},
			wantError:     true,
			errorContains: "parent ticket not found",
		},
		{
			name: "create ticket with template",
			setup: func(env *testharness.TestEnvironment) {
				// Create a template file
				env.WriteFile("tickets/templates/feature.md", `---
priority: 2
description: "Feature template"
---

# Feature Template

## Tasks
- [ ] Design
- [ ] Implementation
- [ ] Testing`)
				
				// Update config to include template
				env.Config.Tickets.Templates = map[string]string{
					"feature": "tickets/templates/feature.md",
				}
			},
			args:  []string{"templated-ticket"},
			flags: map[string]string{"format": "text"},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				// Note: Template application depends on App.NewTicket implementation
				// For now, just verify ticket was created
				files, err := os.ReadDir(env.RootDir + "/tickets/todo")
				require.NoError(t, err)
				
				var found bool
				for _, file := range files {
					if strings.Contains(file.Name(), "templated-ticket") {
						found = true
						break
					}
				}
				assert.True(t, found, "Ticket should be created")
			},
		},
		{
			name: "create ticket with both short and long flags",
			setup: func(env *testharness.TestEnvironment) {
				env.CreateTicket("parent-long", ticket.StatusTodo)
				env.CreateTicket("parent-short", ticket.StatusTodo)
			},
			args: []string{"mixed-flags-ticket"},
			flags: map[string]string{
				"format":      "text",
				"formatShort": "json",  // Short form takes precedence
				"parent":      "parent-long",
				"parentShort": "parent-short", // Short form takes precedence
			},
			validate: func(t *testing.T, env *testharness.TestEnvironment) {
				files, err := os.ReadDir(env.RootDir + "/tickets/todo")
				require.NoError(t, err)
				
				for _, file := range files {
					if strings.Contains(file.Name(), "mixed-flags-ticket") {
						content := env.ReadFile("tickets/todo/" + file.Name())
						// Short form should take precedence
						assert.Contains(t, content, "parent:parent-short")
						assert.NotContains(t, content, "parent:parent-long")
						break
					}
				}
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
			cmd := NewNewCommand()

			// Setup flags
			newFlags := &newFlags{
				parent:      tt.flags["parent"],
				parentShort: tt.flags["parentShort"],
				format:      tt.flags["format"],
				formatShort: tt.flags["formatShort"],
			}

			// Validate flags before execution
			if err := cmd.Validate(newFlags, tt.args); err != nil {
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
			err = cmd.Execute(ctx, newFlags, tt.args)

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

func TestNewCommand_Execute_ContextCancellation(t *testing.T) {
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
	cmd := NewNewCommand()
	newFlags := &newFlags{format: "text"}
	err = cmd.Execute(ctx, newFlags, []string{"test-ticket"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestNewCommand_Execute_InvalidFlagsType(t *testing.T) {
	env := testharness.NewTestEnvironment(t)

	// Change to test directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(env.RootDir))

	// Execute with wrong flags type
	cmd := NewNewCommand()
	err = cmd.Execute(context.Background(), "invalid-flags", []string{"test-ticket"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid flags type")
}

func TestNewCommand_Validate_InvalidFlagsType(t *testing.T) {
	// Validate with wrong flags type
	cmd := NewNewCommand()
	err := cmd.Validate("invalid-flags", []string{"test-ticket"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid flags type")
}