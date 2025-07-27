package ticket

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yshrsmz/ticketflow/internal/config"
)

// Manager manages ticket operations
type Manager struct {
	config      *config.Config
	projectRoot string
}

// NewManager creates a new ticket manager
func NewManager(cfg *config.Config, projectRoot string) *Manager {
	return &Manager{
		config:      cfg,
		projectRoot: projectRoot,
	}
}

// Create creates a new ticket in the todo directory
func (m *Manager) Create(slug string) (*Ticket, error) {
	// Validate slug
	if !IsValidSlug(slug) {
		return nil, fmt.Errorf("invalid slug format: %s", slug)
	}

	// Generate ID
	id := GenerateID(slug)

	// Check if ticket already exists in any directory
	if _, err := m.FindTicket(id); err == nil {
		return nil, fmt.Errorf("ticket already exists: %s", id)
	}

	// Create ticket in todo directory
	todoPath := m.config.GetTodoPath(m.projectRoot)
	ticketPath := filepath.Join(todoPath, id+".md")

	// Create ticket
	ticket := New(slug, "")
	ticket.ID = id
	ticket.Slug = slug
	ticket.Path = ticketPath
	ticket.Content = m.config.Tickets.Template

	// Ensure todo directory exists
	if err := os.MkdirAll(todoPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create todo directory: %w", err)
	}

	// Write ticket file
	data, err := ticket.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize ticket: %w", err)
	}

	if err := os.WriteFile(ticketPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write ticket file: %w", err)
	}

	return ticket, nil
}

// Get retrieves a ticket by ID from any directory
func (m *Manager) Get(id string) (*Ticket, error) {
	ticketPath, err := m.FindTicket(id)
	if err != nil {
		return nil, err
	}
	return m.loadTicket(ticketPath)
}

// List lists tickets with optional status filter
func (m *Manager) List(statusFilter string) ([]Ticket, error) {
	// Determine which directories to search
	dirs := m.getDirectoriesForStatus(statusFilter)

	var tickets []Ticket
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue // Directory doesn't exist yet
			}
			return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			ticketPath := filepath.Join(dir, entry.Name())
			ticket, err := m.loadTicket(ticketPath)
			if err != nil {
				// Skip invalid tickets
				continue
			}

			tickets = append(tickets, *ticket)
		}
	}

	// Sort by priority first, then by creation time (newest first)
	sort.Slice(tickets, func(i, j int) bool {
		if tickets[i].Priority != tickets[j].Priority {
			return tickets[i].Priority < tickets[j].Priority
		}
		return tickets[i].CreatedAt.Time.After(tickets[j].CreatedAt.Time)
	})

	return tickets, nil
}

// getDirectoriesForStatus returns the directories to search based on status filter
func (m *Manager) getDirectoriesForStatus(statusFilter string) []string {
	switch statusFilter {
	case "todo":
		return []string{m.config.GetTodoPath(m.projectRoot)}
	case "doing":
		return []string{m.config.GetDoingPath(m.projectRoot)}
	case "done":
		return []string{m.config.GetDonePath(m.projectRoot)}
	case "": // All active tickets (todo and doing)
		return []string{
			m.config.GetTodoPath(m.projectRoot),
			m.config.GetDoingPath(m.projectRoot),
		}
	default: // All tickets
		return []string{
			m.config.GetTodoPath(m.projectRoot),
			m.config.GetDoingPath(m.projectRoot),
			m.config.GetDonePath(m.projectRoot),
		}
	}
}

// Update updates a ticket
func (m *Manager) Update(ticket *Ticket) error {
	if ticket.Path == "" {
		return fmt.Errorf("ticket path not set")
	}

	data, err := ticket.ToBytes()
	if err != nil {
		return fmt.Errorf("failed to serialize ticket: %w", err)
	}

	if err := os.WriteFile(ticket.Path, data, 0644); err != nil {
		return fmt.Errorf("failed to write ticket file: %w", err)
	}

	return nil
}

// GetCurrentTicket gets the currently active ticket (if any)
func (m *Manager) GetCurrentTicket() (*Ticket, error) {
	linkPath := filepath.Join(m.projectRoot, "current-ticket.md")

	// Check if symlink exists
	target, err := os.Readlink(linkPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No current ticket
		}
		return nil, fmt.Errorf("failed to read current ticket link: %w", err)
	}

	// Load the target ticket
	ticketPath := filepath.Join(m.projectRoot, target)
	return m.loadTicket(ticketPath)
}

// SetCurrentTicket sets the current ticket symlink
func (m *Manager) SetCurrentTicket(ticket *Ticket) error {
	linkPath := filepath.Join(m.projectRoot, "current-ticket.md")

	// Remove existing link if any
	os.Remove(linkPath)

	if ticket == nil {
		return nil
	}

	// Create relative path for symlink
	relPath, err := filepath.Rel(m.projectRoot, ticket.Path)
	if err != nil {
		return fmt.Errorf("failed to create relative path: %w", err)
	}

	// Create symlink
	if err := os.Symlink(relPath, linkPath); err != nil {
		return fmt.Errorf("failed to create current ticket link: %w", err)
	}

	return nil
}

// loadTicket loads a ticket from file
func (m *Manager) loadTicket(path string) (*Ticket, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read ticket file: %w", err)
	}

	ticket, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ticket: %w", err)
	}

	// Set computed fields
	filename := filepath.Base(path)
	ticket.ID = ExtractIDFromFilename(filename)
	ticket.Path = path

	// Extract slug from ID
	_, slug, err := ParseID(ticket.ID)
	if err == nil {
		ticket.Slug = slug
	}

	return ticket, nil
}

// ReadContent reads the content portion of a ticket (without frontmatter)
func (m *Manager) ReadContent(id string) (string, error) {
	ticket, err := m.Get(id)
	if err != nil {
		return "", err
	}
	return ticket.Content, nil
}

// WriteContent writes the content portion of a ticket
func (m *Manager) WriteContent(id string, content string) error {
	ticket, err := m.Get(id)
	if err != nil {
		return err
	}

	ticket.Content = content
	return m.Update(ticket)
}

// findTicketInDir searches for a ticket in a specific directory
func (m *Manager) findTicketInDir(ticketID, dir string) (string, error) {
	// Try exact match first
	ticketPath := filepath.Join(dir, ticketID+".md")
	if _, err := os.Stat(ticketPath); err == nil {
		return ticketPath, nil
	}

	// Try prefix match
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("directory not found")
		}
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	var matches []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		entryID := ExtractIDFromFilename(entry.Name())
		if strings.HasPrefix(entryID, ticketID) {
			matches = append(matches, filepath.Join(dir, entry.Name()))
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("ticket not found")
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("ambiguous ticket ID, multiple matches found")
	}

	return matches[0], nil
}

// FindTicket searches for a ticket across all directories
func (m *Manager) FindTicket(ticketID string) (string, error) {
	// Search in todo -> doing -> done order
	dirs := []string{
		m.config.GetTodoPath(m.projectRoot),
		m.config.GetDoingPath(m.projectRoot),
		m.config.GetDonePath(m.projectRoot),
	}

	var lastErr error
	for _, dir := range dirs {
		path, err := m.findTicketInDir(ticketID, dir)
		if err == nil {
			return path, nil
		}
		// Keep track of errors other than "not found"
		if err != nil && !strings.Contains(err.Error(), "not found") {
			lastErr = err
		}
	}

	// If we have a specific error (like ambiguous), return that
	if lastErr != nil {
		return "", lastErr
	}

	return "", fmt.Errorf("ticket not found: %s", ticketID)
}
