package testutil

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// TicketOption modifies a ticket fixture
type TicketOption func(*ticket.Ticket)

// TicketFixture creates a test ticket with sensible defaults
func TicketFixture(opts ...TicketOption) *ticket.Ticket {
	now := time.Now()
	t := &ticket.Ticket{
		ID:          "test-ticket-123",
		Path:        "/test/tickets/todo/test-ticket-123.md",
		Description: "Test ticket",
		Priority:    1,
		CreatedAt:   ticket.RFC3339Time{Time: now},
		Content:     "# Test Ticket\n\nThis is a test ticket.",
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// WithID sets the ticket ID
func WithID(id string) TicketOption {
	return func(t *ticket.Ticket) {
		t.ID = id
		// Update path to match ID
		if t.Path != "" {
			t.Path = "/test/tickets/todo/" + id + ".md"
		}
	}
}

// WithPath sets the ticket path
func WithPath(path string) TicketOption {
	return func(t *ticket.Ticket) {
		t.Path = path
	}
}

// WithDescription sets the ticket description
func WithDescription(desc string) TicketOption {
	return func(t *ticket.Ticket) {
		t.Description = desc
	}
}

// WithPriority sets the ticket priority
func WithPriority(priority int) TicketOption {
	return func(t *ticket.Ticket) {
		t.Priority = priority
	}
}

// WithContent sets the ticket content
func WithContent(content string) TicketOption {
	return func(t *ticket.Ticket) {
		t.Content = content
	}
}

// WithStatus sets the ticket status by adjusting timestamps
func WithStatus(status ticket.Status) TicketOption {
	return func(t *ticket.Ticket) {
		now := time.Now()
		switch status {
		case ticket.StatusDoing:
			t.StartedAt = ticket.RFC3339TimePtr{Time: &now}
			// Update path to doing directory
			if t.Path != "" {
				t.Path = "/test/tickets/doing/" + t.ID + ".md"
			}
		case ticket.StatusDone:
			startTime := now.Add(-1 * time.Hour)
			t.StartedAt = ticket.RFC3339TimePtr{Time: &startTime}
			t.ClosedAt = ticket.RFC3339TimePtr{Time: &now}
			// Update path to done directory
			if t.Path != "" {
				t.Path = "/test/tickets/done/" + t.ID + ".md"
			}
		}
	}
}

// WithRelated sets the ticket related items
func WithRelated(related []string) TicketOption {
	return func(t *ticket.Ticket) {
		t.Related = related
	}
}

// WithCreatedAt sets the ticket creation time
func WithCreatedAt(createdAt time.Time) TicketOption {
	return func(t *ticket.Ticket) {
		t.CreatedAt = ticket.RFC3339Time{Time: createdAt}
	}
}

// ConfigFixture creates a test config with sensible defaults
func ConfigFixture(opts ...ConfigOption) *config.Config {
	cfg := config.Default()
	cfg.Git.DefaultBranch = "main"
	cfg.Worktree.Enabled = true
	cfg.Worktree.BaseDir = "../test.worktrees"

	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// ConfigOption modifies a config fixture
type ConfigOption func(*config.Config)

// WithWorktreeDisabled disables worktree functionality
func WithWorktreeDisabled() ConfigOption {
	return func(c *config.Config) {
		c.Worktree.Enabled = false
	}
}

// WithWorktreeBaseDir sets the worktree base directory
func WithWorktreeBaseDir(dir string) ConfigOption {
	return func(c *config.Config) {
		c.Worktree.BaseDir = dir
	}
}

// WithDefaultBranch sets the default branch
func WithDefaultBranch(branch string) ConfigOption {
	return func(c *config.Config) {
		c.Git.DefaultBranch = branch
	}
}

// WithOutputFormat sets the output format
func WithOutputFormat(format string) ConfigOption {
	return func(c *config.Config) {
		c.Output.DefaultFormat = format
	}
}

// GenerateTicketID generates a test ticket ID with timestamp
func GenerateTicketID(t *testing.T, suffix string) string {
	t.Helper()
	timestamp := time.Now().Format("060102-150405")
	return timestamp + "-" + suffix
}

// TicketContent generates standard ticket content with frontmatter
func TicketContent(priority int, description string, createdAt time.Time, extra map[string]interface{}) string {
	content := "---\n"
	content += "priority: " + strconv.Itoa(priority) + "\n"
	content += "description: \"" + description + "\"\n"
	content += "created_at: \"" + createdAt.Format(time.RFC3339) + "\"\n"

	for key, value := range extra {
		if value == nil {
			continue
		}
		switch v := value.(type) {
		case string:
			content += fmt.Sprintf("%s: %q\n", key, v)
		case []string:
			content += key + ":\n"
			for _, item := range v {
				content += "  - " + item + "\n"
			}
		case time.Time:
			if !v.IsZero() {
				content += fmt.Sprintf("%s: %q\n", key, v.Format(time.RFC3339))
			}
		case *time.Time:
			if v != nil && !v.IsZero() {
				content += fmt.Sprintf("%s: %q\n", key, v.Format(time.RFC3339))
			}
		default:
			content += fmt.Sprintf("%s: %v\n", key, value)
		}
	}

	content += "---\n\n"
	content += "# " + description + "\n\n"
	content += "Test ticket content"

	return content
}
