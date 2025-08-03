package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/cli"
)

// This file shows how to refactor the existing handlers_test.go to eliminate race conditions

// TestHandleNewRefactored demonstrates the refactored version with proper output capture
func TestHandleNewRefactored(t *testing.T) {
	t.Parallel() // Safe to run in parallel now!

	tests := []struct {
		name          string
		slug          string
		format        string
		setupFunc     func(t *testing.T, tmpDir string)
		expectedError bool
		errorContains string
		checkOutput   func(t *testing.T, stdout, stderr string)
	}{
		{
			name:   "successful text format",
			slug:   "test-feature",
			format: "text",
			setupFunc: func(t *testing.T, tmpDir string) {
				setupTestRepo(t, tmpDir)
			},
			expectedError: false,
			checkOutput: func(t *testing.T, stdout, stderr string) {
				// For text format, we might see status messages
				// but the exact format depends on implementation
			},
		},
		{
			name:   "successful json format",
			slug:   "test-feature",
			format: "json",
			setupFunc: func(t *testing.T, tmpDir string) {
				setupTestRepo(t, tmpDir)
			},
			expectedError: false,
			checkOutput: func(t *testing.T, stdout, stderr string) {
				var result map[string]interface{}
				err := json.Unmarshal([]byte(stdout), &result)
				assert.NoError(t, err)
				assert.Contains(t, result, "ticket")
			},
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
			t.Parallel() // Test cases can also run in parallel!

			tmpDir := t.TempDir()
			tt.setupFunc(t, tmpDir)

			ctx := context.Background()

			// Create buffers to capture output
			var stdout, stderr bytes.Buffer
			outputFormat := cli.ParseOutputFormat(tt.format)

			// Create test-specific output writer
			outputWriter := cli.NewOutputWriter(&stdout, &stderr, outputFormat)

			// Create app with test output writer
			app, err := cli.NewAppWithOptions(ctx,
				cli.WithWorkingDirectory(tmpDir),
				cli.WithOutputWriter(outputWriter),
			)

			var cmdErr error
			if err != nil {
				// If app creation failed, write error using output writer
				outputWriter.Error(err)
				cmdErr = err
			} else {
				// Note: NewTicket method needs to be updated to use app.Output
				// instead of directly writing to os.Stdout
				cmdErr = app.NewTicket(ctx, tt.slug, outputFormat)
			}

			if tt.expectedError {
				assert.Error(t, cmdErr)
				if tt.errorContains != "" {
					// Check both the error and stderr for the error message
					errorFound := false
					if cmdErr != nil && contains(cmdErr.Error(), tt.errorContains) {
						errorFound = true
					}
					if !errorFound && contains(stderr.String(), tt.errorContains) {
						errorFound = true
					}
					assert.True(t, errorFound, "Expected error containing '%s', got error: %v, stderr: %s",
						tt.errorContains, cmdErr, stderr.String())
				}
			} else {
				assert.NoError(t, cmdErr)

				// Check output if provided
				if tt.checkOutput != nil {
					tt.checkOutput(t, stdout.String(), stderr.String())
				}

				// Verify ticket was created
				files, err := filepath.Glob(filepath.Join(tmpDir, "tickets", "todo", "*"))
				require.NoError(t, err)
				assert.NotEmpty(t, files)
			}
		})
	}
}

// TestHandleShowRefactored demonstrates output capture for show command
func TestHandleShowRefactored(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		ticketID      string
		format        string
		setupFunc     func(t *testing.T, tmpDir string) string // returns ticket ID
		expectedError bool
		errorContains string
		checkOutput   func(t *testing.T, stdout, stderr string, ticketID string)
	}{
		{
			name:      "show existing ticket text format",
			format:    "text",
			setupFunc: setupTestRepoWithTicket,
			checkOutput: func(t *testing.T, stdout, stderr string, ticketID string) {
				// For text format, verify expected fields
				assert.Contains(t, stdout, "ID:")
				assert.Contains(t, stdout, "Status:")
				assert.Contains(t, stdout, "Priority:")
				assert.Contains(t, stdout, ticketID)
			},
		},
		{
			name:      "show existing ticket json format",
			format:    "json",
			setupFunc: setupTestRepoWithTicket,
			checkOutput: func(t *testing.T, stdout, stderr string, ticketID string) {
				var result map[string]interface{}
				err := json.Unmarshal([]byte(stdout), &result)
				assert.NoError(t, err)
				assert.Contains(t, result, "ticket")

				ticketData, ok := result["ticket"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, ticketID, ticketData["id"])
			},
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

			// Create app with test output writer
			app, err := cli.NewAppWithOptions(ctx,
				cli.WithWorkingDirectory(tmpDir),
				cli.WithOutputWriter(outputWriter),
			)

			var cmdErr error
			if err != nil {
				outputWriter.Error(err)
				cmdErr = err
			} else {
				// This is a simplified version - the actual implementation
				// would need to be updated to use app.Output
				ticketObj, err := app.Manager.Get(ctx, ticketID)
				if err != nil {
					outputWriter.Error(err)
					cmdErr = err
				} else {
					// Output the ticket data
					if outputFormat == cli.FormatJSON {
						cmdErr = outputWriter.PrintJSON(map[string]interface{}{
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
						})
					} else {
						// Text format output
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
				}
			}

			if tt.expectedError {
				assert.Error(t, cmdErr)
				if tt.errorContains != "" {
					errorFound := cmdErr != nil && contains(cmdErr.Error(), tt.errorContains) ||
						contains(stderr.String(), tt.errorContains)
					assert.True(t, errorFound, "Expected error containing '%s'", tt.errorContains)
				}
			} else {
				assert.NoError(t, cmdErr)
				if tt.checkOutput != nil {
					tt.checkOutput(t, stdout.String(), stderr.String(), ticketID)
				}
			}
		})
	}
}

// TestErrorsExtendedRefactored shows how to test error handling without global state
func TestErrorsExtendedRefactored(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		err            error
		outputFormat   cli.OutputFormat
		expectedStderr string
		expectJSON     bool
	}{
		{
			name: "CLI error text format",
			err: cli.NewError(
				cli.ErrTicketNotFound,
				"Ticket not found",
				"The ticket with ID 'test-123' does not exist",
				[]string{"Check ticket ID", "Use 'ticketflow list' to see available tickets"},
			),
			outputFormat:   cli.FormatText,
			expectedStderr: "Error: Ticket not found\nDetails: The ticket with ID 'test-123' does not exist\n\nSuggestions:\n  - Check ticket ID\n  - Use 'ticketflow list' to see available tickets\n",
		},
		{
			name: "CLI error JSON format",
			err: cli.NewError(
				cli.ErrTicketNotFound,
				"Ticket not found",
				"The ticket with ID 'test-123' does not exist",
				[]string{"Check ticket ID"},
			),
			outputFormat: cli.FormatJSON,
			expectJSON:   true,
		},
		{
			name:           "generic error",
			err:            fmt.Errorf("something went wrong"),
			outputFormat:   cli.FormatText,
			expectedStderr: "Error: something went wrong\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create buffers to capture output
			var stdout, stderr bytes.Buffer

			// Create test-specific output writer
			outputWriter := cli.NewOutputWriter(&stdout, &stderr, tt.outputFormat)

			// Handle the error using the output writer
			outputWriter.Error(tt.err)

			if tt.expectJSON {
				// Verify JSON structure
				var result map[string]interface{}
				err := json.Unmarshal(stderr.Bytes(), &result)
				assert.NoError(t, err)
				assert.Contains(t, result, "error")

				// Verify error fields
				errorData, ok := result["error"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, "TICKET_NOT_FOUND", errorData["code"])
				assert.Equal(t, "Ticket not found", errorData["message"])
			} else {
				assert.Equal(t, tt.expectedStderr, stderr.String())
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
