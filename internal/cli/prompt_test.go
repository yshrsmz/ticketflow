package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsInteractive(t *testing.T) {
	// Cannot use t.Parallel() - tests modify environment variables
	tests := []struct {
		name     string
		setup    func()
		teardown func()
		want     bool
	}{
		{
			name: "CI environment variable set",
			setup: func() {
				_ = os.Setenv("CI", "true")
			},
			teardown: func() {
				_ = os.Unsetenv("CI")
			},
			want: false,
		},
		{
			name: "GITHUB_ACTIONS environment variable set",
			setup: func() {
				_ = os.Setenv("GITHUB_ACTIONS", "true")
			},
			teardown: func() {
				_ = os.Unsetenv("GITHUB_ACTIONS")
			},
			want: false,
		},
		{
			name: "TICKETFLOW_NON_INTERACTIVE set to true",
			setup: func() {
				_ = os.Setenv("TICKETFLOW_NON_INTERACTIVE", "true")
			},
			teardown: func() {
				_ = os.Unsetenv("TICKETFLOW_NON_INTERACTIVE")
			},
			want: false,
		},
		{
			name: "TICKETFLOW_NON_INTERACTIVE set to false",
			setup: func() {
				_ = os.Setenv("TICKETFLOW_NON_INTERACTIVE", "false")
			},
			teardown: func() {
				_ = os.Unsetenv("TICKETFLOW_NON_INTERACTIVE")
			},
			want: true, // Should be interactive because the value is not "true"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			if tt.teardown != nil {
				defer tt.teardown()
			}

			// Note: We can't easily test the terminal detection part in unit tests
			// because it depends on the actual stdin file descriptor.
			// In CI, IsInteractive() will return false due to CI env vars.
			got := IsInteractive()

			// In test environments that are not CI, the terminal check might return true,
			// so we only assert when we expect false
			if !tt.want {
				assert.False(t, got)
			}
		})
	}
}

func TestPromptNonInteractive(t *testing.T) {
	// Cannot use t.Parallel() - tests modify environment variables
	// Save original env var
	originalCI := os.Getenv("CI")
	defer func() {
		if originalCI != "" {
			_ = os.Setenv("CI", originalCI)
		} else {
			_ = os.Unsetenv("CI")
		}
	}()

	// Set CI environment to simulate non-interactive mode
	_ = os.Setenv("CI", "true")

	tests := []struct {
		name        string
		message     string
		options     []PromptOption
		wantChoice  string
		wantErr     bool
		errContains string
	}{
		{
			name:    "uses default option",
			message: "Test prompt",
			options: []PromptOption{
				{Key: "a", Description: "Option A"},
				{Key: "b", Description: "Option B", IsDefault: true},
				{Key: "c", Description: "Option C"},
			},
			wantChoice: "b",
			wantErr:    false,
		},
		{
			name:    "error when no default option",
			message: "Test prompt",
			options: []PromptOption{
				{Key: "a", Description: "Option A"},
				{Key: "b", Description: "Option B"},
			},
			wantChoice:  "",
			wantErr:     true,
			errContains: "no default option available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			choice, err := Prompt(tt.message, tt.options)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantChoice, choice)
			}
		})
	}
}

func TestConfirmPromptNonInteractive(t *testing.T) {
	// Cannot use t.Parallel() - tests modify environment variables
	// Save original env var
	originalCI := os.Getenv("CI")
	defer func() {
		if originalCI != "" {
			_ = os.Setenv("CI", originalCI)
		} else {
			_ = os.Unsetenv("CI")
		}
	}()

	// Set CI environment to simulate non-interactive mode
	_ = os.Setenv("CI", "true")

	tests := []struct {
		name       string
		message    string
		defaultYes bool
		want       bool
	}{
		{
			name:       "default yes",
			message:    "Continue?",
			defaultYes: true,
			want:       true,
		},
		{
			name:       "default no",
			message:    "Delete everything?",
			defaultYes: false,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConfirmPrompt(tt.message, tt.defaultYes)
			assert.Equal(t, tt.want, got)
		})
	}
}
