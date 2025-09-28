package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid text format",
			format:  FormatText,
			wantErr: false,
		},
		{
			name:    "valid json format",
			format:  FormatJSON,
			wantErr: false,
		},
		{
			name:    "invalid yaml format",
			format:  "yaml",
			wantErr: true,
			errMsg:  `invalid format: "yaml" (must be "text" or "json")`,
		},
		{
			name:    "invalid xml format",
			format:  "xml",
			wantErr: true,
			errMsg:  `invalid format: "xml" (must be "text" or "json")`,
		},
		{
			name:    "empty format",
			format:  "",
			wantErr: true,
			errMsg:  `invalid format: "" (must be "text" or "json")`,
		},
		{
			name:    "invalid format with special characters",
			format:  "text/plain",
			wantErr: true,
			errMsg:  `invalid format: "text/plain" (must be "text" or "json")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFormat(tt.format)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.EqualError(t, err, tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExtractParentFromTicket(t *testing.T) {
	tests := []struct {
		name     string
		ticket   *ticket.Ticket
		expected string
	}{
		{
			name:     "nil ticket",
			ticket:   nil,
			expected: "",
		},
		{
			name:     "ticket with no related items",
			ticket:   &ticket.Ticket{},
			expected: "",
		},
		{
			name: "ticket with empty related slice",
			ticket: &ticket.Ticket{
				Related: []string{},
			},
			expected: "",
		},
		{
			name: "ticket with parent relationship",
			ticket: &ticket.Ticket{
				Related: []string{"parent:parent-ticket-123"},
			},
			expected: "parent-ticket-123",
		},
		{
			name: "ticket with multiple relationships including parent",
			ticket: &ticket.Ticket{
				Related: []string{
					"blocks:other-ticket",
					"parent:main-parent",
					"related:sibling-ticket",
				},
			},
			expected: "main-parent",
		},
		{
			name: "ticket with only non-parent relationships",
			ticket: &ticket.Ticket{
				Related: []string{
					"blocks:ticket-1",
					"blocked-by:ticket-2",
					"related:ticket-3",
				},
			},
			expected: "",
		},
		{
			name: "ticket with parent at the end",
			ticket: &ticket.Ticket{
				Related: []string{
					"related:ticket-1",
					"blocks:ticket-2",
					"parent:final-parent",
				},
			},
			expected: "final-parent",
		},
		{
			name: "ticket with empty parent value",
			ticket: &ticket.Ticket{
				Related: []string{"parent:"},
			},
			expected: "",
		},
		{
			name: "ticket with parent containing special characters",
			ticket: &ticket.Ticket{
				Related: []string{"parent:250815-171527-feature-branch"},
			},
			expected: "250815-171527-feature-branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractParentFromTicket(tt.ticket)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test types for AssertFlags testing
type testFlags struct {
	Value string
}

type otherFlags struct {
	Number int
}

func TestAssertFlags(t *testing.T) {
	t.Run("successful type assertion", func(t *testing.T) {
		flags := &testFlags{Value: "test"}
		result, err := AssertFlags[testFlags](flags)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test", result.Value)
	})

	t.Run("failed type assertion - wrong type", func(t *testing.T) {
		flags := &otherFlags{Number: 42}
		result, err := AssertFlags[testFlags](flags)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid flags type")
		assert.Contains(t, err.Error(), "expected *commands.testFlags")
		assert.Contains(t, err.Error(), "got *commands.otherFlags")
	})

	t.Run("failed type assertion - non-pointer", func(t *testing.T) {
		flags := testFlags{Value: "test"} // not a pointer
		result, err := AssertFlags[testFlags](&flags)

		// This actually succeeds because &flags creates a pointer
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("failed type assertion - interface{} containing wrong type", func(t *testing.T) {
		var flags interface{} = &otherFlags{Number: 42}
		result, err := AssertFlags[testFlags](flags)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid flags type")
	})

	t.Run("failed type assertion - nil interface", func(t *testing.T) {
		var flags interface{}
		result, err := AssertFlags[testFlags](flags)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid flags type")
	})

	t.Run("successful assertion with complex type", func(t *testing.T) {
		type complexFlags struct {
			format string
			force  bool
			count  int
		}

		flags := &complexFlags{
			format: "json",
			force:  true,
			count:  10,
		}

		result, err := AssertFlags[complexFlags](flags)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "json", result.format)
		assert.True(t, result.force)
		assert.Equal(t, 10, result.count)
	})
}
