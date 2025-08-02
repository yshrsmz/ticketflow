package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestHandleInit(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldDir)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Initialize git repo
	cmd := exec.Command("git", "init")
	err = cmd.Run()
	require.NoError(t, err)

	// Test handleInit
	ctx := context.Background()
	err = handleInit(ctx)
	assert.NoError(t, err)

	// Verify config file was created
	configPath := filepath.Join(tmpDir, ".ticketflow.yaml")
	assert.FileExists(t, configPath)

	// Verify directories were created
	assert.DirExists(t, filepath.Join(tmpDir, "tickets"))
	assert.DirExists(t, filepath.Join(tmpDir, "tickets", "todo"))
	assert.DirExists(t, filepath.Join(tmpDir, "tickets", "doing"))
	assert.DirExists(t, filepath.Join(tmpDir, "tickets", "done"))
}

func TestHandleNew(t *testing.T) {
	tests := []struct {
		name          string
		slug          string
		format        string
		setupFunc     func(t *testing.T, tmpDir string)
		expectedError bool
		errorContains string
	}{
		{
			name:   "successful text format",
			slug:   "test-feature",
			format: "text",
			setupFunc: func(t *testing.T, tmpDir string) {
				setupTestRepo(t, tmpDir)
			},
			expectedError: false,
		},
		{
			name:   "successful json format",
			slug:   "test-feature",
			format: "json",
			setupFunc: func(t *testing.T, tmpDir string) {
				setupTestRepo(t, tmpDir)
			},
			expectedError: false,
		},
		{
			name:   "invalid slug",
			slug:   "invalid slug with spaces",
			format: "text",
			setupFunc: func(t *testing.T, tmpDir string) {
				setupTestRepo(t, tmpDir)
			},
			expectedError: true,
			errorContains: "Invalid slug format",
		},
		{
			name:          "no config",
			slug:          "test-feature",
			format:        "text",
			setupFunc:     func(t *testing.T, tmpDir string) {},
			expectedError: true,
			errorContains: "Not in a git repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			oldDir, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(oldDir)

			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			tt.setupFunc(t, tmpDir)

			ctx := context.Background()

			// Capture output for JSON format
			if tt.format == "json" {
				oldStdout := os.Stdout
				r, w, _ := os.Pipe()
				os.Stdout = w

				// Run the command
				err = handleNew(ctx, tt.slug, tt.format)
				
				// Restore stdout and read output
				w.Close()
				os.Stdout = oldStdout
				
				var buf bytes.Buffer
				_, _ = io.Copy(&buf, r)
				r.Close()

				// For JSON format, verify output structure
				if !tt.expectedError {
					var result map[string]interface{}
					err2 := json.Unmarshal(buf.Bytes(), &result)
					assert.NoError(t, err2)
					assert.Contains(t, result, "ticket")
				}
			} else {
				err = handleNew(ctx, tt.slug, tt.format)
			}

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)

				// Verify ticket was created
				files, err := os.ReadDir(filepath.Join(tmpDir, "tickets", "todo"))
				require.NoError(t, err)
				assert.NotEmpty(t, files)
			}
		})
	}
}

func TestHandleList(t *testing.T) {
	tests := []struct {
		name          string
		status        string
		count         int
		format        string
		setupFunc     func(t *testing.T, tmpDir string)
		expectedError bool
		errorContains string
	}{
		{
			name:   "list all tickets text format",
			status: "",
			count:  10,
			format: "text",
			setupFunc: func(t *testing.T, tmpDir string) {
				setupTestRepoWithTickets(t, tmpDir)
			},
			expectedError: false,
		},
		{
			name:   "list todo tickets json format",
			status: "todo",
			count:  10,
			format: "json",
			setupFunc: func(t *testing.T, tmpDir string) {
				setupTestRepoWithTickets(t, tmpDir)
			},
			expectedError: false,
		},
		{
			name:   "invalid status",
			status: "invalid",
			count:  10,
			format: "text",
			setupFunc: func(t *testing.T, tmpDir string) {
				setupTestRepo(t, tmpDir)
			},
			expectedError: true,
			errorContains: "invalid status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			oldDir, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(oldDir)

			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			tt.setupFunc(t, tmpDir)

			ctx := context.Background()
			err = handleList(ctx, tt.status, tt.count, tt.format)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleShow(t *testing.T) {
	tests := []struct {
		name          string
		ticketID      string
		format        string
		setupFunc     func(t *testing.T, tmpDir string) string // returns ticket ID
		expectedError bool
		errorContains string
	}{
		{
			name:     "show existing ticket text format",
			format:   "text",
			setupFunc: setupTestRepoWithTicket,
			expectedError: false,
		},
		{
			name:     "show existing ticket json format",
			format:   "json",
			setupFunc: setupTestRepoWithTicket,
			expectedError: false,
		},
		{
			name:     "ticket not found",
			ticketID: "nonexistent-ticket",
			format:   "text",
			setupFunc: func(t *testing.T, tmpDir string) string {
				setupTestRepo(t, tmpDir)
				return "nonexistent-ticket"
			},
			expectedError: true,
			errorContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			oldDir, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(oldDir)

			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			ticketID := tt.setupFunc(t, tmpDir)
			if tt.ticketID != "" {
				ticketID = tt.ticketID
			}

			// Capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			
			ctx := context.Background()
			err = handleShow(ctx, ticketID, tt.format)
			
			w.Close()
			os.Stdout = oldStdout
			
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := buf.String()

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				
				if tt.format == "json" {
					// Verify JSON structure
					var result map[string]interface{}
					err = json.Unmarshal([]byte(output), &result)
					assert.NoError(t, err)
					assert.Contains(t, result, "ticket")
				} else {
					// Verify text output contains expected fields
					assert.Contains(t, output, "ID:")
					assert.Contains(t, output, "Status:")
					assert.Contains(t, output, "Priority:")
				}
			}
		})
	}
}

func TestHandleStart(t *testing.T) {
	tests := []struct {
		name          string
		ticketID      string
		noPush        bool
		setupFunc     func(t *testing.T, tmpDir string) string // returns ticket ID
		expectedError bool
		errorContains string
	}{
		{
			name:     "start valid ticket",
			noPush:   true,
			setupFunc: setupTestRepoWithTicket,
			expectedError: false,
		},
		{
			name:     "ticket not found",
			ticketID: "nonexistent-ticket",
			noPush:   true,
			setupFunc: func(t *testing.T, tmpDir string) string {
				setupTestRepo(t, tmpDir)
				return "nonexistent-ticket"
			},
			expectedError: true,
			errorContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			oldDir, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(oldDir)

			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			ticketID := tt.setupFunc(t, tmpDir)
			if tt.ticketID != "" {
				ticketID = tt.ticketID
			}

			ctx := context.Background()
			err = handleStart(ctx, ticketID, tt.noPush)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleClose(t *testing.T) {
	tests := []struct {
		name          string
		noPush        bool
		force         bool
		setupFunc     func(t *testing.T, tmpDir string)
		expectedError bool
		errorContains string
	}{
		{
			name:   "close current ticket",
			noPush: true,
			force:  false,
			setupFunc: func(t *testing.T, tmpDir string) {
				setupTestRepoWithStartedTicket(t, tmpDir)
			},
			expectedError: false,
		},
		{
			name:   "no current ticket",
			noPush: true,
			force:  false,
			setupFunc: func(t *testing.T, tmpDir string) {
				setupTestRepo(t, tmpDir)
			},
			expectedError: true,
			errorContains: "No active ticket",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			oldDir, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(oldDir)

			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			tt.setupFunc(t, tmpDir)

			ctx := context.Background()
			err = handleClose(ctx, tt.noPush, tt.force)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		status   ticket.Status
		expected bool
	}{
		{ticket.StatusTodo, true},
		{ticket.StatusDoing, true},
		{ticket.StatusDone, true},
		{ticket.Status("invalid"), false},
		{ticket.Status(""), false},
		{ticket.Status("pending"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			result := isValidStatus(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOutputJSON(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		expected string
	}{
		{
			name: "simple map",
			data: map[string]interface{}{
				"key": "value",
			},
			expected: `{"key":"value"}`,
		},
		{
			name: "nested structure",
			data: map[string]interface{}{
				"ticket": map[string]interface{}{
					"id":     "test-123",
					"status": "todo",
				},
			},
			expected: `{"ticket":{"id":"test-123","status":"todo"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := outputJSON(tt.data)
			
			w.Close()
			os.Stdout = oldStdout
			
			assert.NoError(t, err)
			
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := strings.TrimSpace(buf.String())
			
			assert.JSONEq(t, tt.expected, output)
		})
	}
}

// Helper functions for test setup

func setupTestRepo(t *testing.T, tmpDir string) {
	// Initialize git repo
	cmd := exec.Command("git", "init")
	err := cmd.Run()
	require.NoError(t, err)

	// Create config file
	cfg := config.Default()
	cfg.Git.DefaultBranch = "main"
	cfg.Worktree.Enabled = false
	
	// Create config YAML content
	configContent := `git:
  default_branch: main
worktree:
  enabled: false
  base_dir: ../ticketflow.worktrees
tickets:
  dir: tickets
  todo_dir: todo
  doing_dir: doing
  done_dir: done
output:
  default_format: text
`
	
	err = os.WriteFile(filepath.Join(tmpDir, ".ticketflow.yaml"), []byte(configContent), 0644)
	require.NoError(t, err)

	// Create directories
	for _, dir := range []string{"tickets/todo", "tickets/doing", "tickets/done"} {
		err = os.MkdirAll(filepath.Join(tmpDir, dir), 0755)
		require.NoError(t, err)
	}

	// Create empty .current file
	err = os.WriteFile(filepath.Join(tmpDir, "tickets", ".current"), []byte(""), 0644)
	require.NoError(t, err)

	// Configure git
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Run()
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Run()

	// Create initial commit
	cmd = exec.Command("git", "add", ".")
	err = cmd.Run()
	require.NoError(t, err)

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	err = cmd.Run()
	require.NoError(t, err)
}

func setupTestRepoWithTickets(t *testing.T, tmpDir string) {
	setupTestRepo(t, tmpDir)

	// Create some test tickets
	tickets := []struct {
		id     string
		status string
	}{
		{"250101-120000-test-1", "todo"},
		{"250102-120000-test-2", "todo"},
		{"250103-120000-test-3", "doing"},
		{"250104-120000-test-4", "done"},
	}

	for _, tc := range tickets {
		var content string
		switch tc.status {
		case "todo":
			content = fmt.Sprintf(`---
priority: 2
created_at: 2025-01-01T12:00:00Z
---

# Test Ticket %s

This is a test ticket.
`, tc.id)
		case "doing":
			content = fmt.Sprintf(`---
priority: 2
created_at: 2025-01-01T12:00:00Z
started_at: 2025-01-01T13:00:00Z
---

# Test Ticket %s

This is a test ticket in progress.
`, tc.id)
		case "done":
			content = fmt.Sprintf(`---
priority: 2
created_at: 2025-01-01T12:00:00Z
started_at: 2025-01-01T13:00:00Z
closed_at: 2025-01-01T14:00:00Z
---

# Test Ticket %s

This is a completed test ticket.
`, tc.id)
		}

		path := filepath.Join(tmpDir, "tickets", tc.status, tc.id+".md")
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Commit all tickets
	cmd := exec.Command("git", "add", ".")
	err := cmd.Run()
	require.NoError(t, err)

	cmd = exec.Command("git", "commit", "-m", "Add test tickets")
	err = cmd.Run()
	require.NoError(t, err)
}

func setupTestRepoWithTicket(t *testing.T, tmpDir string) string {
	setupTestRepo(t, tmpDir)

	ticketID := "250101-120000-test-feature"
	content := `---
priority: 2
created_at: 2025-01-01T12:00:00Z
---

# Test Feature

This is a test ticket for testing show command.
`

	path := filepath.Join(tmpDir, "tickets", "todo", ticketID+".md")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	// Commit the ticket
	cmd := exec.Command("git", "add", path)
	err = cmd.Run()
	require.NoError(t, err)

	cmd = exec.Command("git", "commit", "-m", "Add test ticket")
	err = cmd.Run()
	require.NoError(t, err)

	return ticketID
}

func setupTestRepoWithStartedTicket(t *testing.T, tmpDir string) string {
	// Start fresh without calling setupTestRepoWithTicket to avoid conflicts
	setupTestRepo(t, tmpDir)

	ticketID := "250101-120000-test-feature"
	
	// Create and checkout the feature branch
	cmd := exec.Command("git", "checkout", "-b", ticketID)
	err := cmd.Run()
	require.NoError(t, err)
	
	// Create ticket directly in doing status
	content := `---
priority: 2
created_at: 2025-01-01T12:00:00Z
started_at: 2025-01-01T13:00:00Z
---

# Test Feature

This is a test ticket that has been started.
`

	path := filepath.Join(tmpDir, "tickets", "doing", ticketID+".md")
	err = os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	// Set as current ticket - create symlink
	linkPath := filepath.Join(tmpDir, "current-ticket.md")
	targetPath := filepath.Join("tickets", "doing", ticketID+".md")
	err = os.Symlink(targetPath, linkPath)
	require.NoError(t, err)

	// Commit the changes
	cmd = exec.Command("git", "add", ".")
	err = cmd.Run()
	require.NoError(t, err)

	cmd = exec.Command("git", "commit", "-m", "Add started ticket")
	err = cmd.Run()
	require.NoError(t, err)

	return ticketID
}