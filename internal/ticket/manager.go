package ticket

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/yshrsmz/ticketflow/internal/config"
	ticketerrors "github.com/yshrsmz/ticketflow/internal/errors"
)

// StatusFilter represents the filter type for listing tickets
type StatusFilter string

// Status filter constants for List method
const (
	StatusFilterAll    StatusFilter = "all"    // Include all tickets (todo, doing, done)
	StatusFilterActive StatusFilter = "active" // Include only active tickets (todo, doing)
	StatusFilterTodo   StatusFilter = "todo"   // Include only todo tickets
	StatusFilterDoing  StatusFilter = "doing"  // Include only doing tickets
	StatusFilterDone   StatusFilter = "done"   // Include only done tickets
)

const (
	// initialTicketCapacity is the initial capacity for ticket slices
	// Most projects have 10-50 active tickets, so 50 is a good starting capacity
	initialTicketCapacity = 50

	// initialMatchCapacity is the initial capacity for search match slices
	// Most searches result in 0-1 matches, rarely more than 5
	initialMatchCapacity = 5
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
func (m *Manager) Create(ctx context.Context, slug string) (*Ticket, error) {
	// Check context
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("operation cancelled: %w", err)
	}
	// Validate slug
	if !IsValidSlug(slug) {
		return nil, fmt.Errorf("invalid slug format: %s", slug)
	}

	// Generate ID
	id := GenerateID(slug)

	// Check if ticket already exists in any directory
	if _, err := m.FindTicket(ctx, id); err == nil {
		return nil, ticketerrors.NewTicketError("create", id, ticketerrors.ErrTicketExists)
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
		return nil, ticketerrors.NewTicketError("create", id, fmt.Errorf("failed to serialize ticket: %w", err))
	}

	if err := writeFileWithContext(ctx, ticketPath, data, 0644); err != nil {
		return nil, ticketerrors.NewTicketError("create", id, fmt.Errorf("failed to write ticket file: %w", err))
	}

	return ticket, nil
}

// Get retrieves a ticket by ID from any directory
func (m *Manager) Get(ctx context.Context, id string) (*Ticket, error) {
	// Check context
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("operation cancelled: %w", err)
	}

	ticketPath, err := m.FindTicket(ctx, id)
	if err != nil {
		return nil, err
	}
	return m.loadTicket(ctx, ticketPath)
}

// List lists tickets with optional status filter
func (m *Manager) List(ctx context.Context, statusFilter StatusFilter) ([]Ticket, error) {
	// Check context
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("operation cancelled: %w", err)
	}
	// Determine which directories to search
	dirs := m.getDirectoriesForStatus(statusFilter)
	if dirs == nil {
		return nil, fmt.Errorf("invalid status filter: %s", statusFilter)
	}

	// Count total files first to better estimate capacity
	totalFiles := 0
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if !os.IsNotExist(err) {
				// Log error but continue with other directories
				continue
			}
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				totalFiles++
			}
		}
	}

	// Use concurrent loading if we have enough files to benefit from it
	// Threshold is set to 10 files based on typical overhead of goroutines
	if totalFiles >= 10 {
		return m.listConcurrent(ctx, dirs)
	}

	// Fall back to sequential for small numbers of tickets
	return m.listSequential(ctx, dirs)
}

// listSequential lists tickets sequentially (original implementation)
func (m *Manager) listSequential(ctx context.Context, dirs []string) ([]Ticket, error) {
	// Pre-allocate tickets slice with reasonable capacity based on typical usage
	// This avoids multiple reallocations during append operations without double-reading directories
	tickets := make([]Ticket, 0, initialTicketCapacity)
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
			// Check context in loop
			if err := ctx.Err(); err != nil {
				return nil, fmt.Errorf("operation cancelled: %w", err)
			}
			ticket, err := m.loadTicket(ctx, ticketPath)
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
		return tickets[i].CreatedAt.After(tickets[j].CreatedAt.Time)
	})

	return tickets, nil
}

// listConcurrent lists tickets using concurrent file operations
func (m *Manager) listConcurrent(ctx context.Context, dirs []string) ([]Ticket, error) {
	// Collect all ticket files to process
	var ticketPaths []string
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
		}

		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				ticketPaths = append(ticketPaths, filepath.Join(dir, entry.Name()))
			}
		}
	}

	if len(ticketPaths) == 0 {
		return []Ticket{}, nil
	}

	// Determine optimal number of workers
	numWorkers := runtime.NumCPU()
	if numWorkers > len(ticketPaths) {
		numWorkers = len(ticketPaths)
	}
	// Cap at 8 workers to avoid excessive file handles
	if numWorkers > 8 {
		numWorkers = 8
	}

	// Create semaphore to limit concurrent file operations
	sem := semaphore.NewWeighted(int64(numWorkers))

	// Pre-allocate result slice with exact capacity
	tickets := make([]Ticket, 0, len(ticketPaths))
	var mu sync.Mutex

	// Use errgroup for structured concurrency
	g, ctx := errgroup.WithContext(ctx)

	for _, path := range ticketPaths {
		path := path // Capture for goroutine

		g.Go(func() error {
			// Acquire semaphore
			if err := sem.Acquire(ctx, 1); err != nil {
				return fmt.Errorf("failed to acquire semaphore: %w", err)
			}
			defer sem.Release(1)

			// Check context before loading
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("operation cancelled: %w", err)
			}

			// Load ticket
			ticket, err := m.loadTicket(ctx, path)
			if err != nil {
				// Skip invalid tickets - don't fail the entire operation
				return nil
			}

			// Add to results with mutex protection
			mu.Lock()
			tickets = append(tickets, *ticket)
			mu.Unlock()

			return nil
		})
	}

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Sort by priority first, then by creation time (newest first)
	sort.Slice(tickets, func(i, j int) bool {
		if tickets[i].Priority != tickets[j].Priority {
			return tickets[i].Priority < tickets[j].Priority
		}
		return tickets[i].CreatedAt.After(tickets[j].CreatedAt.Time)
	})

	return tickets, nil
}

// getDirectoriesForStatus returns the directories to search based on status filter
func (m *Manager) getDirectoriesForStatus(statusFilter StatusFilter) []string {
	switch statusFilter {
	case StatusFilterTodo:
		return []string{m.config.GetTodoPath(m.projectRoot)}
	case StatusFilterDoing:
		return []string{m.config.GetDoingPath(m.projectRoot)}
	case StatusFilterDone:
		return []string{m.config.GetDonePath(m.projectRoot)}
	case StatusFilterActive, "": // Active tickets (todo and doing)
		return []string{
			m.config.GetTodoPath(m.projectRoot),
			m.config.GetDoingPath(m.projectRoot),
		}
	case StatusFilterAll:
		return []string{
			m.config.GetTodoPath(m.projectRoot),
			m.config.GetDoingPath(m.projectRoot),
			m.config.GetDonePath(m.projectRoot),
		}
	default:
		// Return nil to indicate invalid filter
		return nil
	}
}

// Update updates a ticket
func (m *Manager) Update(ctx context.Context, ticket *Ticket) error {
	// Check context
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operation cancelled: %w", err)
	}
	if ticket.Path == "" {
		return ticketerrors.NewTicketError("update", ticket.ID, ticketerrors.ErrTicketInvalid)
	}

	data, err := ticket.ToBytes()
	if err != nil {
		return ticketerrors.NewTicketError("update", ticket.ID, fmt.Errorf("failed to serialize ticket: %w", err))
	}

	if err := writeFileWithContext(ctx, ticket.Path, data, 0644); err != nil {
		return ticketerrors.NewTicketError("update", ticket.ID, fmt.Errorf("failed to write ticket file: %w", err))
	}

	return nil
}

// GetCurrentTicket gets the currently active ticket (if any)
func (m *Manager) GetCurrentTicket(ctx context.Context) (*Ticket, error) {
	// Check context
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("operation cancelled: %w", err)
	}
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
	return m.loadTicket(ctx, ticketPath)
}

// SetCurrentTicket sets the current ticket symlink
func (m *Manager) SetCurrentTicket(ctx context.Context, ticket *Ticket) error {
	// Check context
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operation cancelled: %w", err)
	}
	linkPath := filepath.Join(m.projectRoot, "current-ticket.md")

	// Remove existing link if any
	_ = os.Remove(linkPath)

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
func (m *Manager) loadTicket(ctx context.Context, path string) (*Ticket, error) {
	// Check context
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("operation cancelled: %w", err)
	}
	data, err := readFileWithContext(ctx, path)
	if err != nil {
		return nil, ticketerrors.NewTicketError("read", filepath.Base(path), fmt.Errorf("failed to read ticket file: %w", err))
	}

	ticket, err := Parse(data)
	if err != nil {
		return nil, ticketerrors.NewTicketError("parse", filepath.Base(path), fmt.Errorf("failed to parse ticket: %w", err))
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
func (m *Manager) ReadContent(ctx context.Context, id string) (string, error) {
	// Check context
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("operation cancelled: %w", err)
	}

	ticket, err := m.Get(ctx, id)
	if err != nil {
		return "", err
	}
	return ticket.Content, nil
}

// WriteContent writes the content portion of a ticket
func (m *Manager) WriteContent(ctx context.Context, id string, content string) error {
	// Check context
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operation cancelled: %w", err)
	}

	ticket, err := m.Get(ctx, id)
	if err != nil {
		return err
	}

	ticket.Content = content
	return m.Update(ctx, ticket)
}

// findTicketInDir searches for a ticket in a specific directory
func (m *Manager) findTicketInDir(ticketID, dir string) (string, error) {
	// Try exact match first
	ticketPath := filepath.Join(dir, ticketID+FileExtension)
	if _, err := os.Stat(ticketPath); err == nil {
		return ticketPath, nil
	}

	// Try prefix match
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ticketerrors.ErrTicketNotFound
		}
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	// Pre-allocate matches slice with small initial capacity
	matches := make([]string, 0, initialMatchCapacity)
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
		return "", ticketerrors.ErrTicketNotFound
	}
	if len(matches) > 1 {
		return "", ticketerrors.NewTicketError("find", ticketID, fmt.Errorf("ambiguous ticket ID, multiple matches found"))
	}

	return matches[0], nil
}

// FindTicket searches for a ticket across all directories
func (m *Manager) FindTicket(ctx context.Context, ticketID string) (string, error) {
	// Check context
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("operation cancelled: %w", err)
	}
	// Search in todo -> doing -> done order
	dirs := []string{
		m.config.GetTodoPath(m.projectRoot),
		m.config.GetDoingPath(m.projectRoot),
		m.config.GetDonePath(m.projectRoot),
	}

	var lastErr error
	for _, dir := range dirs {
		// Check context in loop
		if err := ctx.Err(); err != nil {
			return "", fmt.Errorf("operation cancelled: %w", err)
		}
		path, err := m.findTicketInDir(ticketID, dir)
		if err == nil {
			return path, nil
		}
		// Keep track of errors other than "not found"
		if err != nil && !errors.Is(err, ticketerrors.ErrTicketNotFound) {
			lastErr = err
		}
	}

	// If we have a specific error (like ambiguous), return that
	if lastErr != nil {
		return "", lastErr
	}

	return "", ticketerrors.NewTicketError("find", ticketID, ticketerrors.ErrTicketNotFound)
}

// readFileWithContext reads a file with context support
func readFileWithContext(ctx context.Context, path string) ([]byte, error) {
	// Check context before starting
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("operation cancelled: %w", err)
	}

	// Open file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close() // Ignore close error for read operations
	}()

	// Get file info
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Validate file size (50MB limit for ticket files)
	const maxTicketSize = 50 * 1024 * 1024 // 50MB
	if info.Size() > maxTicketSize {
		return nil, fmt.Errorf("file too large: %d bytes exceeds %d bytes limit", info.Size(), maxTicketSize)
	}

	// For small files (< 1MB), read all at once
	if info.Size() < 1024*1024 {
		// Check context one more time before reading
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("operation cancelled: %w", err)
		}
		return os.ReadFile(path)
	}

	// For larger files, read in chunks with context checks
	const chunkSize = 64 * 1024 // 64KB chunks
	// Pre-allocate result slice based on file size to avoid multiple reallocations
	result := make([]byte, 0, info.Size())
	buffer := make([]byte, chunkSize)

	for {
		// Check context before each chunk
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("operation cancelled during read: %w", err)
		}

		n, err := file.Read(buffer)
		if n > 0 {
			result = append(result, buffer[:n]...)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}

	return result, nil
}

// writeFileWithContext writes a file with context support
func writeFileWithContext(ctx context.Context, path string, data []byte, perm os.FileMode) (err error) {
	// Check context before starting
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operation cancelled: %w", err)
	}

	// For small files (< 1MB), write all at once
	if len(data) < 1024*1024 {
		// Check context one more time before writing
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("operation cancelled: %w", err)
		}
		return os.WriteFile(path, data, perm)
	}

	// For larger files, write in chunks with context checks
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close file: %w", cerr)
		}
	}()

	const chunkSize = 64 * 1024 // 64KB chunks
	for i := 0; i < len(data); i += chunkSize {
		// Check context before each chunk
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("operation cancelled during write: %w", err)
		}

		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}

		_, err := file.Write(data[i:end])
		if err != nil {
			return err
		}
	}

	// Ensure data is persisted to disk
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	return nil
}
