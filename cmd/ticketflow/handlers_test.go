package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestHandleInit(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for the test
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	err := cmd.Run()
	require.NoError(t, err)

	// Configure git locally
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	err = cmd.Run()
	require.NoError(t, err)

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	err = cmd.Run()
	require.NoError(t, err)

	// Test InitCommandWithWorkingDir instead of handleInit
	ctx := context.Background()
	err = cli.InitCommandWithWorkingDir(ctx, tmpDir)
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
	t.Parallel()

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
			t.Parallel()

			tmpDir := t.TempDir()
			tt.setupFunc(t, tmpDir)

			ctx := context.Background()

			// Create app with working directory
			app, err := cli.NewAppWithWorkingDir(ctx, t, tmpDir)
			var cmdErr error

			if err != nil {
				cmdErr = err
			} else {
				// Capture output for JSON format
				if tt.format == "json" {
					oldStdout := os.Stdout
					r, w, err := os.Pipe()
					require.NoError(t, err)
					defer func() {
						os.Stdout = oldStdout
						if err := r.Close(); err != nil {
							t.Logf("Failed to close reader: %v", err)
						}
					}()
					os.Stdout = w

					// Run the command
					outputFormat := cli.ParseOutputFormat(tt.format)
					cli.GlobalOutputFormat = outputFormat
					cmdErr = app.NewTicket(ctx, tt.slug, outputFormat)

					// Close write end and read output
					if err := w.Close(); err != nil {
						t.Logf("Failed to close writer: %v", err)
					}

					var buf bytes.Buffer
					_, _ = io.Copy(&buf, r)

					// For JSON format, verify output structure
					if !tt.expectedError && cmdErr == nil {
						var result map[string]interface{}
						err2 := json.Unmarshal(buf.Bytes(), &result)
						assert.NoError(t, err2)
						assert.Contains(t, result, "ticket")
					}
				} else {
					outputFormat := cli.ParseOutputFormat(tt.format)
					cli.GlobalOutputFormat = outputFormat
					cmdErr = app.NewTicket(ctx, tt.slug, outputFormat)
				}
			}

			if tt.expectedError {
				assert.Error(t, cmdErr)
				if tt.errorContains != "" {
					assert.Contains(t, cmdErr.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, cmdErr)

				// Verify ticket was created
				files, err := os.ReadDir(filepath.Join(tmpDir, "tickets", "todo"))
				require.NoError(t, err)
				assert.NotEmpty(t, files)
			}
		})
	}
}

func TestHandleList(t *testing.T) {
	// Cannot run in parallel because handleList works in the current directory

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
			origDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				if err := os.Chdir(origDir); err != nil {
					t.Logf("Failed to restore directory: %v", err)
				}
			}()

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
	// Cannot run in parallel because handleShow works in the current directory

	tests := []struct {
		name          string
		ticketID      string
		format        string
		setupFunc     func(t *testing.T, tmpDir string) string // returns ticket ID
		expectedError bool
		errorContains string
	}{
		{
			name:          "show existing ticket text format",
			format:        "text",
			setupFunc:     setupTestRepoWithTicket,
			expectedError: false,
		},
		{
			name:          "show existing ticket json format",
			format:        "json",
			setupFunc:     setupTestRepoWithTicket,
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
			origDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				if err := os.Chdir(origDir); err != nil {
					t.Logf("Failed to restore directory: %v", err)
				}
			}()

			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			ticketID := tt.setupFunc(t, tmpDir)
			if tt.ticketID != "" {
				ticketID = tt.ticketID
			}

			// Capture output
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			require.NoError(t, err)
			defer func() {
				os.Stdout = oldStdout
				if err := r.Close(); err != nil {
					t.Logf("Failed to close reader: %v", err)
				}
			}()
			os.Stdout = w

			ctx := context.Background()
			err = handleShow(ctx, ticketID, tt.format)

			if err := w.Close(); err != nil {
				t.Logf("Warning: failed to close writer: %v", err)
			}

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
	// Cannot run in parallel because handleStart works in the current directory

	tests := []struct {
		name          string
		ticketID      string
		noPush        bool
		setupFunc     func(t *testing.T, tmpDir string) string // returns ticket ID
		expectedError bool
		errorContains string
	}{
		{
			name:          "start valid ticket",
			noPush:        true,
			setupFunc:     setupTestRepoWithTicket,
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
			origDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				if err := os.Chdir(origDir); err != nil {
					t.Logf("Failed to restore directory: %v", err)
				}
			}()

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
	// Cannot run in parallel because handleClose works in the current directory

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
			origDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				if err := os.Chdir(origDir); err != nil {
					t.Logf("Failed to restore directory: %v", err)
				}
			}()

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
	t.Parallel()

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
	t.Parallel()

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
			r, w, err := os.Pipe()
			require.NoError(t, err)
			defer func() {
				os.Stdout = oldStdout
				if err := r.Close(); err != nil {
					t.Logf("Failed to close reader: %v", err)
				}
			}()
			os.Stdout = w

			err = outputJSON(tt.data)

			if err := w.Close(); err != nil {
				t.Logf("Warning: failed to close writer: %v", err)
			}

			assert.NoError(t, err)

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := strings.TrimSpace(buf.String())

			assert.JSONEq(t, tt.expected, output)
		})
	}
}
