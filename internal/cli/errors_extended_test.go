package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		outputFormat   OutputFormat
		expectedOutput string
		expectJSON     bool
	}{
		{
			name:           "nil error",
			err:            nil,
			outputFormat:   FormatText,
			expectedOutput: "",
		},
		{
			name: "CLI error text format",
			err: NewError(
				ErrTicketNotFound,
				"Ticket not found",
				"The ticket with ID 'test-123' does not exist",
				[]string{"Check ticket ID", "Use 'ticketflow list' to see available tickets"},
			),
			outputFormat:   FormatText,
			expectedOutput: "Error: Ticket not found\nDetails: The ticket with ID 'test-123' does not exist\n\nSuggestions:\n  - Check ticket ID\n  - Use 'ticketflow list' to see available tickets\n",
		},
		{
			name: "CLI error JSON format",
			err: NewError(
				ErrTicketNotFound,
				"Ticket not found",
				"The ticket with ID 'test-123' does not exist",
				[]string{"Check ticket ID"},
			),
			outputFormat: FormatJSON,
			expectJSON:   true,
		},
		{
			name:           "generic error",
			err:            fmt.Errorf("something went wrong"),
			outputFormat:   FormatText,
			expectedOutput: "Error: something went wrong\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			// Set global output format
			oldFormat := GlobalOutputFormat
			GlobalOutputFormat = tt.outputFormat
			defer func() { GlobalOutputFormat = oldFormat }()

			HandleError(tt.err)

			w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := buf.String()

			if tt.expectJSON {
				// Verify JSON structure
				var result map[string]interface{}
				err := json.Unmarshal([]byte(output), &result)
				assert.NoError(t, err)
				assert.Contains(t, result, "error")
			} else {
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}

func TestHandleCLIError_EnvironmentVariable(t *testing.T) {
	// Test that environment variable overrides format
	testErr := NewError(
		ErrConfigNotFound,
		"Configuration not found",
		"",
		nil,
	)

	// Set environment variable using t.Setenv for automatic cleanup
	t.Setenv("TICKETFLOW_OUTPUT_FORMAT", "json")

	// Capture stderr
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer func() {
		os.Stderr = oldStderr
		r.Close()
	}()
	os.Stderr = w

	// Ensure global format is text (to test env override)
	oldFormat := GlobalOutputFormat
	GlobalOutputFormat = FormatText
	defer func() { GlobalOutputFormat = oldFormat }()

	handleCLIError(testErr)

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Should be JSON despite text format
	var result map[string]interface{}
	err2 := json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err2)
	assert.Contains(t, result, "error")
}

func TestOutputJSONError(t *testing.T) {
	tests := []struct {
		name     string
		err      *CLIError
		expected map[string]interface{}
	}{
		{
			name: "error with all fields",
			err: &CLIError{
				Code:        ErrTicketExists,
				Message:     "Ticket already exists",
				Details:     "A ticket with slug 'test-feature' already exists",
				Suggestions: []string{"Use a different slug", "Check existing tickets"},
			},
			expected: map[string]interface{}{
				"error": map[string]interface{}{
					"code":    ErrTicketExists,
					"message": "Ticket already exists",
					"details": "A ticket with slug 'test-feature' already exists",
					"suggestions": []interface{}{
						"Use a different slug",
						"Check existing tickets",
					},
				},
			},
		},
		{
			name: "error without optional fields",
			err: &CLIError{
				Code:    ErrNotGitRepo,
				Message: "Not in a git repository",
			},
			expected: map[string]interface{}{
				"error": map[string]interface{}{
					"code":    ErrNotGitRepo,
					"message": "Not in a git repository",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			OutputJSONError(tt.err)

			w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := buf.String()

			var result map[string]interface{}
			err := json.Unmarshal([]byte(output), &result)
			require.NoError(t, err)

			// Compare JSON structures
			expectedJSON, _ := json.Marshal(tt.expected)
			actualJSON, _ := json.Marshal(result)
			assert.JSONEq(t, string(expectedJSON), string(actualJSON))
		})
	}
}

func TestCLIError_ErrorMethod(t *testing.T) {
	err := NewError(
		ErrGitDirtyWorkspace,
		"Workspace has uncommitted changes",
		"Please commit or stash your changes",
		nil,
	)

	assert.Equal(t, "Workspace has uncommitted changes", err.Error())
}

func TestNewError_Creation(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		message     string
		details     string
		suggestions []string
	}{
		{
			name:        "full error",
			code:        ErrWorktreeExists,
			message:     "Worktree already exists",
			details:     "A worktree for ticket 'test-123' already exists",
			suggestions: []string{"Use existing worktree", "Remove old worktree first"},
		},
		{
			name:        "minimal error",
			code:        ErrPermissionDenied,
			message:     "Permission denied",
			details:     "",
			suggestions: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewError(tt.code, tt.message, tt.details, tt.suggestions)

			assert.Equal(t, tt.code, err.Code)
			assert.Equal(t, tt.message, err.Message)
			assert.Equal(t, tt.details, err.Details)
			assert.Equal(t, tt.suggestions, err.Suggestions)
		})
	}
}

// Test all error codes are defined properly
func TestErrorCodes(t *testing.T) {
	// This test ensures all error codes are unique strings
	errorCodes := []string{
		ErrNotGitRepo,
		ErrConfigNotFound,
		ErrConfigInvalid,
		ErrPermissionDenied,
		ErrTicketNotFound,
		ErrTicketExists,
		ErrTicketInvalid,
		ErrTicketNotStarted,
		ErrTicketAlreadyStarted,
		ErrTicketAlreadyClosed,
		ErrTicketNotDone,
		ErrGitDirtyWorkspace,
		ErrGitBranchExists,
		ErrGitMergeFailed,
		ErrGitPushFailed,
		ErrWorktreeExists,
		ErrWorktreeNotFound,
		ErrWorktreeCreateFailed,
		ErrWorktreeRemoveFailed,
		ErrInvalidContext,
	}

	// Check for duplicates
	seen := make(map[string]bool)
	for _, code := range errorCodes {
		assert.False(t, seen[code], "Duplicate error code: %s", code)
		seen[code] = true
		assert.NotEmpty(t, code, "Error code should not be empty")
	}
}

func TestHandleError_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func()
		err          error
		expectOutput string
	}{
		{
			name: "CLI error without details",
			err: NewError(
				ErrConfigInvalid,
				"Invalid configuration",
				"",
				nil,
			),
			expectOutput: "Error: Invalid configuration\n",
		},
		{
			name: "CLI error with single suggestion",
			err: NewError(
				ErrGitBranchExists,
				"Branch already exists",
				"",
				[]string{"Use a different branch name"},
			),
			expectOutput: "Error: Branch already exists\n\nSuggestions:\n  - Use a different branch name\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			// Ensure text format
			oldFormat := GlobalOutputFormat
			GlobalOutputFormat = FormatText
			defer func() { GlobalOutputFormat = oldFormat }()

			HandleError(tt.err)

			w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := buf.String()

			assert.Equal(t, tt.expectOutput, output)
		})
	}
}
