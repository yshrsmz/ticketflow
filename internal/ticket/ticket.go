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
	Priority    int        `yaml:"priority"`
	Description string     `yaml:"description"`
	CreatedAt   time.Time  `yaml:"created_at"`
	StartedAt   *time.Time `yaml:"started_at"`
	ClosedAt    *time.Time `yaml:"closed_at"`
	Related     []string   `yaml:"related,omitempty"`
	Progress    int        `yaml:"progress,omitempty"`    // Progress percentage (0-100)
	Tasks       []Task     `yaml:"tasks,omitempty"`       // Task list for tracking

	// Computed fields
	ID      string `yaml:"-"`
	Slug    string `yaml:"-"`
	Path    string `yaml:"-"`
	Content string `yaml:"-"`
}

// Task represents a subtask within a ticket
type Task struct {
	Description string     `yaml:"description"`
	Completed   bool       `yaml:"completed"`
	CompletedAt *time.Time `yaml:"completed_at,omitempty"`
}

// Status returns the current status of the ticket
func (t *Ticket) Status() Status {
	if t.ClosedAt != nil {
		return StatusDone
	}
	if t.StartedAt != nil {
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
	if strings.HasSuffix(filename, ".md") {
		filename = strings.TrimSuffix(filename, ".md")
	}
	return filename
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
		CreatedAt:   now,
		StartedAt:   nil,
		ClosedAt:    nil,
		Related:     []string{},
	}
}

// Start marks the ticket as started
func (t *Ticket) Start() error {
	if t.StartedAt != nil {
		return fmt.Errorf("ticket already started")
	}
	if t.ClosedAt != nil {
		return fmt.Errorf("ticket already closed")
	}

	now := time.Now()
	t.StartedAt = &now
	return nil
}

// Close marks the ticket as closed
func (t *Ticket) Close() error {
	if t.StartedAt == nil {
		return fmt.Errorf("ticket not started")
	}
	if t.ClosedAt != nil {
		return fmt.Errorf("ticket already closed")
	}

	now := time.Now()
	t.ClosedAt = &now
	return nil
}

// CalculateProgress calculates progress based on completed tasks
func (t *Ticket) CalculateProgress() int {
	if len(t.Tasks) == 0 {
		return t.Progress
	}

	completed := 0
	for _, task := range t.Tasks {
		if task.Completed {
			completed++
		}
	}

	return (completed * 100) / len(t.Tasks)
}

// UpdateProgress updates the progress percentage
func (t *Ticket) UpdateProgress(progress int) error {
	if progress < 0 || progress > 100 {
		return fmt.Errorf("progress must be between 0 and 100")
	}
	t.Progress = progress
	return nil
}

// AddTask adds a new task to the ticket
func (t *Ticket) AddTask(description string) {
	t.Tasks = append(t.Tasks, Task{
		Description: description,
		Completed:   false,
	})
}

// CompleteTask marks a task as completed
func (t *Ticket) CompleteTask(index int) error {
	if index < 0 || index >= len(t.Tasks) {
		return fmt.Errorf("task index out of range")
	}
	
	if !t.Tasks[index].Completed {
		now := time.Now()
		t.Tasks[index].Completed = true
		t.Tasks[index].CompletedAt = &now
	}
	
	// Update overall progress
	t.Progress = t.CalculateProgress()
	return nil
}

// GetCompletedTasksCount returns the number of completed tasks
func (t *Ticket) GetCompletedTasksCount() int {
	count := 0
	for _, task := range t.Tasks {
		if task.Completed {
			count++
		}
	}
	return count
}