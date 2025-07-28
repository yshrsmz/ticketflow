package ticket

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Status represents the ticket status
type Status string

const (
	StatusTodo  Status = "todo"
	StatusDoing Status = "doing"
	StatusDone  Status = "done"
)

// Ticket represents a ticket with metadata and content
type Ticket struct {
	// Metadata (YAML frontmatter)
	Priority    int            `yaml:"priority"`
	Description string         `yaml:"description"`
	CreatedAt   RFC3339Time    `yaml:"created_at"`
	StartedAt   RFC3339TimePtr `yaml:"started_at"`
	ClosedAt    RFC3339TimePtr `yaml:"closed_at"`
	Related     []string       `yaml:"related,omitempty"`

	// Computed fields
	ID      string `yaml:"-"`
	Slug    string `yaml:"-"`
	Path    string `yaml:"-"`
	Content string `yaml:"-"`
}

// Status returns the current status of the ticket
func (t *Ticket) Status() Status {
	if t.ClosedAt.Time != nil {
		return StatusDone
	}
	if t.StartedAt.Time != nil {
		return StatusDoing
	}
	return StatusTodo
}

// HasWorktree checks if the ticket has an associated worktree
func (t *Ticket) HasWorktree() bool {
	return t.Status() == StatusDoing
}

// Parse parses a ticket file content
func Parse(content []byte) (*Ticket, error) {
	parts := bytes.SplitN(content, []byte("---\n"), 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid ticket format: missing frontmatter")
	}

	// Parse YAML frontmatter
	var ticket Ticket
	if err := yaml.Unmarshal(parts[1], &ticket); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Set content (remove leading newline if present)
	ticket.Content = strings.TrimPrefix(string(parts[2]), "\n")

	return &ticket, nil
}

// ToBytes converts the ticket to file content
func (t *Ticket) ToBytes() ([]byte, error) {
	var buf bytes.Buffer

	// Write frontmatter
	buf.WriteString("---\n")

	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(0)
	if err := encoder.Encode(t); err != nil {
		return nil, fmt.Errorf("failed to encode frontmatter: %w", err)
	}

	buf.WriteString("---\n\n")
	buf.WriteString(t.Content)

	return buf.Bytes(), nil
}

// GenerateID generates a ticket ID with current timestamp
func GenerateID(slug string) string {
	now := time.Now()
	return fmt.Sprintf("%s-%s",
		now.Format("060102-150405"),
		slug,
	)
}

// ParseID extracts timestamp and slug from ticket ID
func ParseID(id string) (time.Time, string, error) {
	parts := strings.SplitN(id, "-", 3)
	if len(parts) < 3 {
		return time.Time{}, "", fmt.Errorf("invalid ticket ID format")
	}

	// Parse timestamp
	timestamp, err := time.Parse("060102-150405",
		fmt.Sprintf("%s-%s", parts[0], parts[1]))
	if err != nil {
		return time.Time{}, "", fmt.Errorf("invalid timestamp: %w", err)
	}

	slug := parts[2]
	return timestamp, slug, nil
}

// ExtractIDFromFilename extracts ticket ID from filename
func ExtractIDFromFilename(filename string) string {
	// Remove .md extension if present
	return strings.TrimSuffix(filename, ".md")
}

// IsValidSlug checks if a slug is valid
func IsValidSlug(slug string) bool {
	if slug == "" {
		return false
	}

	for _, r := range slug {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			return false
		}
	}

	return true
}

// New creates a new ticket with defaults
func New(slug, description string) *Ticket {
	now := time.Now()
	return &Ticket{
		Priority:    2,
		Description: description,
		CreatedAt:   NewRFC3339Time(now),
		StartedAt:   RFC3339TimePtr{},
		ClosedAt:    RFC3339TimePtr{},
		Related:     []string{},
	}
}

// Start marks the ticket as started
func (t *Ticket) Start() error {
	if t.StartedAt.Time != nil {
		return fmt.Errorf("ticket already started")
	}
	if t.ClosedAt.Time != nil {
		return fmt.Errorf("ticket already closed")
	}

	now := time.Now()
	t.StartedAt = NewRFC3339TimePtr(&now)
	return nil
}

// Close marks the ticket as closed
func (t *Ticket) Close() error {
	if t.StartedAt.Time == nil {
		return fmt.Errorf("ticket not started")
	}
	if t.ClosedAt.Time != nil {
		return fmt.Errorf("ticket already closed")
	}

	now := time.Now()
	t.ClosedAt = NewRFC3339TimePtr(&now)
	return nil
}
