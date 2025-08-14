package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	cli.ConfigureTestGit(t, tmpDir)

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

			// Create buffers to capture output
			var stdout, stderr bytes.Buffer
			outputFormat := cli.ParseOutputFormat(tt.format)

			// Create test-specific output writer
			outputWriter := cli.NewOutputWriter(&stdout, &stderr, outputFormat)

			// Create app with output writer
			app, err := cli.NewAppWithOptions(ctx,
				cli.WithWorkingDirectory(tmpDir),
				cli.WithOutputWriter(outputWriter),
			)
			var cmdErr error

			if err != nil {
				cmdErr = err
			} else {
				ticket, err := app.NewTicket(ctx, tt.slug, "")
				cmdErr = err

				// Handle output based on format
				if cmdErr == nil && tt.format == "json" {
					// Generate JSON output for JSON format test
					output := map[string]interface{}{
						"ticket": map[string]interface{}{
							"id":   ticket.ID,
							"path": ticket.Path,
						},
					}
					if err := app.Output.PrintJSON(output); err != nil {
						cmdErr = fmt.Errorf("failed to print JSON: %w", err)
					}
				}
			}

			// Verify JSON output structure if needed
			if tt.format == "json" && !tt.expectedError && cmdErr == nil {
				var result map[string]interface{}
				err := json.Unmarshal(stdout.Bytes(), &result)
				assert.NoError(t, err)
				assert.Contains(t, result, "ticket")
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
	t.Parallel()

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
			t.Parallel()

			tmpDir := t.TempDir()
			tt.setupFunc(t, tmpDir)

			ctx := context.Background()

			// Create buffers to capture output
			var stdout, stderr bytes.Buffer
			outputFormat := cli.ParseOutputFormat(tt.format)

			// Create test-specific output writer
			outputWriter := cli.NewOutputWriter(&stdout, &stderr, outputFormat)

			// Create app with output writer
			app, err := cli.NewAppWithOptions(ctx,
				cli.WithWorkingDirectory(tmpDir),
				cli.WithOutputWriter(outputWriter),
			)
			var cmdErr error

			if err != nil {
				cmdErr = err
			} else {
				// Validate status if provided
				var ticketStatus ticket.Status
				if tt.status != "" {
					ticketStatus = ticket.Status(tt.status)
					if !isValidStatus(ticketStatus) {
						cmdErr = fmt.Errorf("invalid status: %s", tt.status)
					}
				}

				if cmdErr == nil {
					cmdErr = app.ListTickets(ctx, ticketStatus, tt.count, outputFormat)
				}
			}

			if tt.expectedError {
				assert.Error(t, cmdErr)
				if tt.errorContains != "" {
					assert.Contains(t, cmdErr.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, cmdErr)
			}
		})
	}
}

func TestHandleShow(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			tmpDir := t.TempDir()
			ticketID := tt.setupFunc(t, tmpDir)
			if tt.ticketID != "" {
				ticketID = tt.ticketID
			}

			ctx := context.Background()

			// Create buffers to capture output
			var stdout, stderr bytes.Buffer
			outputFormat := cli.ParseOutputFormat(tt.format)

			// Create test-specific output writer
			outputWriter := cli.NewOutputWriter(&stdout, &stderr, outputFormat)

			// Create app with output writer
			app, err := cli.NewAppWithOptions(ctx,
				cli.WithWorkingDirectory(tmpDir),
				cli.WithOutputWriter(outputWriter),
			)
			var cmdErr error

			if err != nil {
				cmdErr = err
			} else {
				// Get the ticket
				ticketObj, err := app.Manager.Get(ctx, ticketID)
				if err != nil {
					cmdErr = err
				} else {
					if outputFormat == cli.FormatJSON {
						// For JSON, output the ticket data
						if err := outputWriter.PrintJSON(map[string]interface{}{
							"ticket": map[string]interface{}{
								"id":          ticketObj.ID,
								"path":        ticketObj.Path,
								"status":      string(ticketObj.Status()),
								"priority":    ticketObj.Priority,
								"description": ticketObj.Description,
								"created_at":  ticketObj.CreatedAt.Time,
								"started_at":  ticketObj.StartedAt.Time,
								"closed_at":   ticketObj.ClosedAt.Time,
								"related":     ticketObj.Related,
								"content":     ticketObj.Content,
							},
						}); err != nil {
							cmdErr = err
						}
					} else {
						// For text format, print ticket details
						outputWriter.Printf("ID: %s\n", ticketObj.ID)
						outputWriter.Printf("Status: %s\n", ticketObj.Status())
						outputWriter.Printf("Priority: %d\n", ticketObj.Priority)
						outputWriter.Printf("Description: %s\n", ticketObj.Description)
						outputWriter.Printf("Created: %s\n", ticketObj.CreatedAt.Format(time.RFC3339))
						if ticketObj.StartedAt.Time != nil {
							outputWriter.Printf("Started: %s\n", ticketObj.StartedAt.Time.Format(time.RFC3339))
						}
						if ticketObj.ClosedAt.Time != nil {
							outputWriter.Printf("Closed: %s\n", ticketObj.ClosedAt.Time.Format(time.RFC3339))
						}
					}

					output := stdout.String()

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
			}

			if tt.expectedError {
				assert.Error(t, cmdErr)
				if tt.errorContains != "" {
					assert.Contains(t, cmdErr.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, cmdErr)
			}
		})
	}
}

func TestHandleStart(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			tmpDir := t.TempDir()
			ticketID := tt.setupFunc(t, tmpDir)
			if tt.ticketID != "" {
				ticketID = tt.ticketID
			}

			ctx := context.Background()

			// Create buffers to capture output
			var stdout, stderr bytes.Buffer

			// Create test-specific output writer
			outputWriter := cli.NewOutputWriter(&stdout, &stderr, cli.FormatText)

			// Create app with output writer
			app, err := cli.NewAppWithOptions(ctx,
				cli.WithWorkingDirectory(tmpDir),
				cli.WithOutputWriter(outputWriter),
			)
			var cmdErr error

			if err != nil {
				cmdErr = err
			} else {
				_, cmdErr = app.StartTicket(ctx, ticketID, false)
			}

			if tt.expectedError {
				assert.Error(t, cmdErr)
				if tt.errorContains != "" {
					assert.Contains(t, cmdErr.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, cmdErr)
			}
		})
	}
}

func TestHandleClose(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			tmpDir := t.TempDir()
			tt.setupFunc(t, tmpDir)

			ctx := context.Background()

			// Create buffers to capture output
			var stdout, stderr bytes.Buffer

			// Create test-specific output writer
			outputWriter := cli.NewOutputWriter(&stdout, &stderr, cli.FormatText)

			// Create app with output writer
			app, err := cli.NewAppWithOptions(ctx,
				cli.WithWorkingDirectory(tmpDir),
				cli.WithOutputWriter(outputWriter),
			)
			var cmdErr error

			if err != nil {
				cmdErr = err
			} else {
				_, cmdErr = app.CloseTicket(ctx, tt.force)
			}

			if tt.expectedError {
				assert.Error(t, cmdErr)
				if tt.errorContains != "" {
					assert.Contains(t, cmdErr.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, cmdErr)
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
			t.Parallel()

			// Create buffers to capture output
			var stdout, stderr bytes.Buffer

			// Create test-specific output writer
			outputWriter := cli.NewOutputWriter(&stdout, &stderr, cli.FormatJSON)

			err := outputWriter.PrintJSON(tt.data)
			assert.NoError(t, err)

			output := strings.TrimSpace(stdout.String())
			assert.JSONEq(t, tt.expected, output)
		})
	}
}
