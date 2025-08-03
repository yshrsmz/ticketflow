package ticket

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTicketStatus(t *testing.T) {
	t.Parallel()
	now := time.Now()

	tests := []struct {
		name     string
		ticket   Ticket
		expected Status
	}{
		{
			name: "todo ticket",
			ticket: Ticket{
				StartedAt: RFC3339TimePtr{},
				ClosedAt:  RFC3339TimePtr{},
			},
			expected: StatusTodo,
		},
		{
			name: "doing ticket",
			ticket: Ticket{
				StartedAt: NewRFC3339TimePtr(&now),
				ClosedAt:  RFC3339TimePtr{},
			},
			expected: StatusDoing,
		},
		{
			name: "done ticket",
			ticket: Ticket{
				StartedAt: NewRFC3339TimePtr(&now),
				ClosedAt:  NewRFC3339TimePtr(&now),
			},
			expected: StatusDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.ticket.Status())
		})
	}
}

func TestTicketHasWorktree(t *testing.T) {
	t.Parallel()
	now := time.Now()

	tests := []struct {
		name     string
		ticket   Ticket
		expected bool
	}{
		{
			name: "todo ticket has no worktree",
			ticket: Ticket{
				StartedAt: RFC3339TimePtr{},
				ClosedAt:  RFC3339TimePtr{},
			},
			expected: false,
		},
		{
			name: "doing ticket has worktree",
			ticket: Ticket{
				StartedAt: NewRFC3339TimePtr(&now),
				ClosedAt:  RFC3339TimePtr{},
			},
			expected: true,
		},
		{
			name: "done ticket has no worktree",
			ticket: Ticket{
				StartedAt: NewRFC3339TimePtr(&now),
				ClosedAt:  NewRFC3339TimePtr(&now),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.ticket.HasWorktree())
		})
	}
}

func TestParse(t *testing.T) {
	t.Parallel()
	content := `---
priority: 1
description: "Test ticket"
created_at: 2025-01-24T10:00:00Z
started_at: null
closed_at: null
related: ["123456-789012-related"]
---

# Test Content

This is the ticket content.

## Task List
- [ ] Task 1
- [ ] Task 2`

	ticket, err := Parse([]byte(content))
	require.NoError(t, err)

	assert.Equal(t, 1, ticket.Priority)
	assert.Equal(t, "Test ticket", ticket.Description)
	assert.Equal(t, []string{"123456-789012-related"}, ticket.Related)
	assert.Contains(t, ticket.Content, "# Test Content")
	assert.Contains(t, ticket.Content, "- [ ] Task 1")
}

func TestParseInvalidFormat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		content string
		errMsg  string
	}{
		{
			name:    "missing frontmatter",
			content: "# Just content",
			errMsg:  "invalid ticket format: missing frontmatter",
		},
		{
			name: "invalid yaml",
			content: `---
invalid: yaml: content:
---

Content`,
			errMsg: "failed to parse frontmatter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.content))
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestToBytes(t *testing.T) {
	t.Parallel()
	now := time.Now()
	ticket := &Ticket{
		Priority:    1,
		Description: "Test ticket",
		CreatedAt:   NewRFC3339Time(now),
		StartedAt:   RFC3339TimePtr{},
		ClosedAt:    RFC3339TimePtr{},
		Related:     []string{"related-1", "related-2"},
		Content:     "# Test Content\n\nThis is the content.",
	}

	data, err := ticket.ToBytes()
	require.NoError(t, err)

	content := string(data)
	assert.True(t, strings.HasPrefix(content, "---\n"))
	assert.Contains(t, content, "priority: 1")
	assert.Contains(t, content, `description: Test ticket`)
	assert.Contains(t, content, "# Test Content")

	// Test round trip
	parsed, err := Parse(data)
	require.NoError(t, err)
	assert.Equal(t, ticket.Priority, parsed.Priority)
	assert.Equal(t, ticket.Description, parsed.Description)
	assert.Equal(t, ticket.Related, parsed.Related)
	assert.Equal(t, ticket.Content, parsed.Content)
}

func TestGenerateID(t *testing.T) {
	t.Parallel()
	id := GenerateID("test-slug")

	// ID should have format YYMMDD-HHMMSS-test-slug
	parts := strings.SplitN(id, "-", 3)
	require.Len(t, parts, 3)

	// Verify date format
	assert.Len(t, parts[0], 6) // YYMMDD
	assert.Len(t, parts[1], 6) // HHMMSS
	assert.Equal(t, "test-slug", parts[2])
}

func TestParseID(t *testing.T) {
	t.Parallel()
	// Generate a known ID
	id := "250124-150000-test-slug"

	timestamp, slug, err := ParseID(id)
	require.NoError(t, err)

	assert.Equal(t, "test-slug", slug)
	assert.Equal(t, 2025, timestamp.Year())
	assert.Equal(t, time.January, timestamp.Month())
	assert.Equal(t, 24, timestamp.Day())
	assert.Equal(t, 15, timestamp.Hour())
	assert.Equal(t, 0, timestamp.Minute())
	assert.Equal(t, 0, timestamp.Second())
}

func TestParseIDInvalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		id   string
	}{
		{
			name: "too few parts",
			id:   "250124-test",
		},
		{
			name: "invalid timestamp",
			id:   "invalid-123456-test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ParseID(tt.id)
			assert.Error(t, err)
		})
	}
}

func TestIsValidSlug(t *testing.T) {
	t.Parallel()
	tests := []struct {
		slug  string
		valid bool
	}{
		{"valid-slug", true},
		{"test123", true},
		{"123test", true},
		{"test-123-slug", true},
		{"", false},
		{"Test-Slug", false}, // uppercase
		{"test_slug", false}, // underscore
		{"test slug", false}, // space
		{"test.slug", false}, // dot
	}

	for _, tt := range tests {
		t.Run(tt.slug, func(t *testing.T) {
			assert.Equal(t, tt.valid, IsValidSlug(tt.slug))
		})
	}
}

func TestExtractIDFromFilename(t *testing.T) {
	t.Parallel()
	tests := []struct {
		filename string
		expected string
	}{
		{"250124-150000-test-slug.md", "250124-150000-test-slug"},
		{"250124-150000-test-slug", "250124-150000-test-slug"},
		{"test.md", "test"},
		{"test", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			assert.Equal(t, tt.expected, ExtractIDFromFilename(tt.filename))
		})
	}
}

func TestTicketStartClose(t *testing.T) {
	t.Parallel()
	ticket := New("test", "Test ticket")

	// Test start
	err := ticket.Start()
	require.NoError(t, err)
	assert.NotNil(t, ticket.StartedAt)

	// Test already started
	err = ticket.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already started")

	// Test close
	err = ticket.Close()
	require.NoError(t, err)
	assert.NotNil(t, ticket.ClosedAt)

	// Test already closed
	err = ticket.Close()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already closed")
}

func TestTicketCloseNotStarted(t *testing.T) {
	t.Parallel()
	ticket := New("test", "Test ticket")

	err := ticket.Close()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not started")
}
