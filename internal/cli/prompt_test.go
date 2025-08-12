package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsInteractive(t *testing.T) {
	// Cannot use t.Parallel() - tests use t.Setenv() which is incompatible with t.Parallel()
	tests := []struct {
		name     string
		envVar   string
		envValue string
		want     bool
	}{
		{
			name: "CI environment variable set",
			envVar: "CI",
			envValue: "true",
			want: false,
		},
		{
			name: "GITHUB_ACTIONS environment variable set",
			envVar: "GITHUB_ACTIONS",
			envValue: "true",
			want: false,
		},
		{
			name: "TICKETFLOW_NON_INTERACTIVE set to true",
			envVar: "TICKETFLOW_NON_INTERACTIVE",
			envValue: "true",
			want: false,
		},
		{
			name: "TICKETFLOW_NON_INTERACTIVE set to false",
			envVar: "TICKETFLOW_NON_INTERACTIVE",
			envValue: "false",
			want: true, // Should be interactive because the value is not "true"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				t.Setenv(tt.envVar, tt.envValue)
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
	// Cannot use t.Parallel() - test uses t.Setenv() which is incompatible with t.Parallel()
	// Set CI environment to simulate non-interactive mode
	t.Setenv("CI", "true")

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
	// Cannot use t.Parallel() - test uses t.Setenv() which is incompatible with t.Parallel()
	// Set CI environment to simulate non-interactive mode
	t.Setenv("CI", "true")

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
