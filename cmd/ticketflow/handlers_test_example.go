package main

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yshrsmz/ticketflow/internal/cli"
)

// Example of how to properly test output without race conditions
func TestHandleNewWithProperOutputCapture(t *testing.T) {
	t.Parallel() // Can now run in parallel!

	tests := []struct {
		name          string
		slug          string
		format        string
		setupFunc     func(t *testing.T, tmpDir string)
		expectedError bool
		errorContains string
	}{
		{
			name:   "successful json format",
			slug:   "test-feature",
			format: "json",
			setupFunc: func(t *testing.T, tmpDir string) {
				setupTestRepo(t, tmpDir)
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Can now run test cases in parallel!

			tmpDir := t.TempDir()
			tt.setupFunc(t, tmpDir)

			ctx := context.Background()

			// Create buffers to capture output
			var stdout, stderr bytes.Buffer
			outputFormat := cli.ParseOutputFormat(tt.format)

			// Create a test-specific output writer
			outputWriter := cli.NewOutputWriter(&stdout, &stderr, outputFormat)

			// Create app with test output writer
			app, err := cli.NewAppWithOptions(ctx,
				cli.WithWorkingDirectory(tmpDir),
				cli.WithOutputWriter(outputWriter),
			)

			var cmdErr error
			if err != nil {
				cmdErr = err
			} else {
				_, cmdErr = app.NewTicket(ctx, tt.slug, "")
			}

			if tt.expectedError {
				assert.Error(t, cmdErr)
				if tt.errorContains != "" {
					assert.Contains(t, cmdErr.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, cmdErr)

				// For JSON format, verify output structure
				if tt.format == "json" {
					var result map[string]interface{}
					err := json.Unmarshal(stdout.Bytes(), &result)
					assert.NoError(t, err)
					assert.Contains(t, result, "ticket")
				}
			}
		})
	}
}

// Example of testing error output without race conditions
func TestErrorHandlingWithProperOutputCapture(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create buffers to capture output
			var stdout, stderr bytes.Buffer

			// Create test-specific output writer
			outputWriter := cli.NewOutputWriter(&stdout, &stderr, tt.outputFormat)

			// Use the output writer to handle errors
			outputWriter.Error(tt.err)

			if tt.expectJSON {
				// Verify JSON structure
				var result map[string]interface{}
				err := json.Unmarshal(stderr.Bytes(), &result)
				assert.NoError(t, err)
				assert.Contains(t, result, "error")
			} else {
				assert.Equal(t, tt.expectedStderr, stderr.String())
			}
		})
	}
}
