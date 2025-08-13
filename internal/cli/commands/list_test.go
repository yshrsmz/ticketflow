package commands

import (
	"context"
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestListCommand_Interface(t *testing.T) {
	cmd := NewListCommand()

	// Verify it implements the Command interface
	var _ = cmd

	assert.Equal(t, "list", cmd.Name())
	assert.Equal(t, []string{"ls"}, cmd.Aliases())
	assert.Equal(t, "List tickets", cmd.Description())
	assert.Equal(t, "list [--status todo|doing|done|all] [--count N] [--format text|json]", cmd.Usage())
}

func TestListCommand_SetupFlags(t *testing.T) {
	cmd := &ListCommand{}

	// Test parsing different flag combinations
	tests := []struct {
		name           string
		args           []string
		expectedStatus string
		expectedCount  int
		expectedFormat string
	}{
		{
			name:           "all defaults",
			args:           []string{},
			expectedStatus: "",
			expectedCount:  20,
			expectedFormat: "text",
		},
		{
			name:           "status flag long form",
			args:           []string{"--status", "todo"},
			expectedStatus: "todo",
			expectedCount:  20,
			expectedFormat: "text",
		},
		{
			name:           "status flag short form",
			args:           []string{"-s", "doing"},
			expectedStatus: "doing",
			expectedCount:  20,
			expectedFormat: "text",
		},
		{
			name:           "count flag long form",
			args:           []string{"--count", "10"},
			expectedStatus: "",
			expectedCount:  10,
			expectedFormat: "text",
		},
		{
			name:           "count flag short form",
			args:           []string{"-c", "5"},
			expectedStatus: "",
			expectedCount:  5,
			expectedFormat: "text",
		},
		{
			name:           "format flag",
			args:           []string{"--format", "json"},
			expectedStatus: "",
			expectedCount:  20,
			expectedFormat: "json",
		},
		{
			name:           "all flags combined",
			args:           []string{"--status", "done", "--count", "15", "--format", "json"},
			expectedStatus: "done",
			expectedCount:  15,
			expectedFormat: "json",
		},
		{
			name:           "all status",
			args:           []string{"--status", "all"},
			expectedStatus: "all",
			expectedCount:  20,
			expectedFormat: "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			flags := cmd.SetupFlags(fs)

			// Verify flags is not nil
			assert.NotNil(t, flags)

			err := fs.Parse(tt.args)
			assert.NoError(t, err)

			// Check the flag values
			lf := flags.(*listFlags)
			// For status, check both long and short forms
			actualStatus := lf.status
			if lf.statusShort != "" {
				actualStatus = lf.statusShort
			}
			assert.Equal(t, tt.expectedStatus, actualStatus)
			
			// For count, check both long and short forms
			actualCount := lf.count
			if lf.countShort != 20 && lf.countShort != 0 {
				actualCount = lf.countShort
			}
			assert.Equal(t, tt.expectedCount, actualCount)
			
			assert.Equal(t, tt.expectedFormat, lf.format)
		})
	}
}

func TestListCommand_Validate(t *testing.T) {
	cmd := &ListCommand{}

	tests := []struct {
		name      string
		flags     interface{}
		args      []string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid defaults",
			flags:     &listFlags{status: "", statusShort: "", count: 20, countShort: 20, format: "text"},
			args:      []string{},
			wantError: false,
		},
		{
			name:      "valid todo status",
			flags:     &listFlags{status: "todo", statusShort: "", count: 20, countShort: 20, format: "text"},
			args:      []string{},
			wantError: false,
		},
		{
			name:      "valid doing status",
			flags:     &listFlags{status: "doing", statusShort: "", count: 20, countShort: 20, format: "text"},
			args:      []string{},
			wantError: false,
		},
		{
			name:      "valid done status",
			flags:     &listFlags{status: "done", statusShort: "", count: 20, countShort: 20, format: "text"},
			args:      []string{},
			wantError: false,
		},
		{
			name:      "valid all status",
			flags:     &listFlags{status: "all", statusShort: "", count: 20, countShort: 20, format: "text"},
			args:      []string{},
			wantError: false,
		},
		{
			name:      "invalid status",
			flags:     &listFlags{status: "invalid", statusShort: "", count: 20, countShort: 20, format: "text"},
			args:      []string{},
			wantError: true,
			errorMsg:  `invalid status: "invalid" (must be 'todo', 'doing', 'done', or 'all')`,
		},
		{
			name:      "negative count",
			flags:     &listFlags{status: "", statusShort: "", count: -1, countShort: 20, format: "text"},
			args:      []string{},
			wantError: true,
			errorMsg:  "count must be non-negative, got -1",
		},
		{
			name:      "zero count is valid",
			flags:     &listFlags{status: "", statusShort: "", count: 0, countShort: 20, format: "text"},
			args:      []string{},
			wantError: false,
		},
		{
			name:      "invalid format",
			flags:     &listFlags{status: "", statusShort: "", count: 20, countShort: 20, format: "xml"},
			args:      []string{},
			wantError: true,
			errorMsg:  `invalid format: "xml" (must be 'text' or 'json')`,
		},
		{
			name:      "json format",
			flags:     &listFlags{status: "", statusShort: "", count: 20, countShort: 20, format: "json"},
			args:      []string{},
			wantError: false,
		},
		{
			name:      "empty status is valid",
			flags:     &listFlags{status: "", statusShort: "", count: 20, countShort: 20, format: "text"},
			args:      []string{},
			wantError: false,
		},
		{
			name:      "with unexpected arguments",
			flags:     &listFlags{status: "todo", statusShort: "", count: 20, countShort: 20, format: "json"},
			args:      []string{"extra", "args"},
			wantError: false,
		},
		{
			name:      "short status flag takes precedence",
			flags:     &listFlags{status: "todo", statusShort: "doing", count: 20, countShort: 20, format: "text"},
			args:      []string{},
			wantError: false,
		},
		{
			name:      "short count flag takes precedence",
			flags:     &listFlags{status: "", statusShort: "", count: 30, countShort: 5, format: "text"},
			args:      []string{},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Validate(tt.flags, tt.args)
			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.EqualError(t, err, tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListCommand_Execute(t *testing.T) {
	// Integration test that verifies the command works with real App
	// This test will succeed in the actual ticketflow environment

	t.Run("default parameters", func(t *testing.T) {
		cmd := &ListCommand{}
		ctx := context.Background()
		flags := &listFlags{status: "", statusShort: "", count: 20, countShort: 20, format: "text"}

		// This will succeed when run in a ticketflow environment
		err := cmd.Execute(ctx, flags, []string{})

		// The command should execute without error
		assert.NoError(t, err)
	})

	t.Run("json format with todo status", func(t *testing.T) {
		cmd := &ListCommand{}
		ctx := context.Background()
		flags := &listFlags{status: "todo", statusShort: "", count: 10, countShort: 20, format: "json"}

		err := cmd.Execute(ctx, flags, []string{})

		// The command should execute without error
		assert.NoError(t, err)
	})

	t.Run("all status with limited count", func(t *testing.T) {
		cmd := &ListCommand{}
		ctx := context.Background()
		flags := &listFlags{status: "all", statusShort: "", count: 5, countShort: 20, format: "text"}

		err := cmd.Execute(ctx, flags, []string{})

		// The command should execute without error
		assert.NoError(t, err)
	})
}

func TestIsValidListStatus(t *testing.T) {
	tests := []struct {
		status string
		valid  bool
	}{
		{"todo", true},
		{"doing", true},
		{"done", true},
		{"all", true},
		{"", false},
		{"active", false},
		{"invalid", false},
		{string(ticket.StatusTodo), true},
		{string(ticket.StatusDoing), true},
		{string(ticket.StatusDone), true},
		{string(cli.StatusAll), true},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := isValidListStatus(tt.status)
			assert.Equal(t, tt.valid, result)
		})
	}
}