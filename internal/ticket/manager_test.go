package ticket

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/config"
)

func setupTestManager(t *testing.T) (*Manager, string) {
	tmpDir := t.TempDir()

	cfg := config.Default()
	cfg.Tickets.Dir = "tickets"

	manager := NewManager(cfg, tmpDir)

	return manager, tmpDir
}

func TestManagerCreate(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create ticket
	ticket, err := manager.Create("test-ticket")
	require.NoError(t, err)

	assert.NotEmpty(t, ticket.ID)
	assert.Equal(t, "test-ticket", ticket.Slug)
	assert.Contains(t, ticket.Path, "test-ticket.md")
	assert.Equal(t, manager.config.Tickets.Template, ticket.Content)

	// Verify file exists
	_, err = os.Stat(ticket.Path)
	require.NoError(t, err)

	// Test duplicate creation (should fail due to timing)
	time.Sleep(time.Second) // Ensure different timestamp
	_, err = manager.Create("test-ticket")
	require.NoError(t, err) // Different timestamp means different ID
}

func TestManagerCreateInvalidSlug(t *testing.T) {
	manager, _ := setupTestManager(t)

	tests := []string{
		"Test-Ticket", // uppercase
		"test ticket", // space
		"test_ticket", // underscore
		"",            // empty
	}

	for _, slug := range tests {
		t.Run(slug, func(t *testing.T) {
			_, err := manager.Create(slug)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid slug format")
		})
	}
}

func TestManagerGet(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create ticket
	created, err := manager.Create("test-ticket")
	require.NoError(t, err)

	// Get by full ID
	retrieved, err := manager.Get(created.ID)
	require.NoError(t, err)

	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.Slug, retrieved.Slug)
	assert.Equal(t, created.Description, retrieved.Description)
}

func TestManagerGetByPrefix(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create ticket
	created, err := manager.Create("test-ticket")
	require.NoError(t, err)

	// Get by prefix (first 6 chars of ID)
	prefix := created.ID[:6]
	retrieved, err := manager.Get(prefix)
	require.NoError(t, err)

	assert.Equal(t, created.ID, retrieved.ID)
}

func TestManagerGetNotFound(t *testing.T) {
	manager, _ := setupTestManager(t)

	_, err := manager.Get("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ticket not found")
}

func TestManagerGetAmbiguous(t *testing.T) {
	manager, tmpDir := setupTestManager(t)

	// Create two tickets with similar IDs manually
	todoPath := filepath.Join(tmpDir, "tickets", "todo")
	os.MkdirAll(todoPath, 0755)

	// Create two files with same prefix
	ticket1 := New("test1", "Test 1")
	ticket2 := New("test2", "Test 2")

	data1, _ := ticket1.ToBytes()
	data2, _ := ticket2.ToBytes()

	os.WriteFile(filepath.Join(todoPath, "250124-150000-test1.md"), data1, 0644)
	os.WriteFile(filepath.Join(todoPath, "250124-150001-test2.md"), data2, 0644)

	// Try to get with ambiguous prefix
	_, err := manager.Get("250124")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ambiguous ticket ID")
}

func TestManagerList(t *testing.T) {
	manager, tmpDir := setupTestManager(t)

	// Create multiple tickets
	ticket1, err := manager.Create("ticket-1")
	require.NoError(t, err)

	time.Sleep(time.Second)
	ticket2, err := manager.Create("ticket-2")
	require.NoError(t, err)

	// Start ticket2 - simulate moving to doing directory
	ticket2.Start()

	// Move ticket2 from todo to doing
	oldPath := ticket2.Path
	doingPath := filepath.Join(tmpDir, "tickets", "doing")
	os.MkdirAll(doingPath, 0755)
	newPath := filepath.Join(doingPath, filepath.Base(ticket2.Path))
	os.Rename(oldPath, newPath)
	ticket2.Path = newPath
	err = manager.Update(ticket2)
	require.NoError(t, err)

	// List all tickets
	tickets, err := manager.List("")
	require.NoError(t, err)
	assert.Len(t, tickets, 2)

	// List by status
	todoTickets, err := manager.List(string(StatusTodo))
	require.NoError(t, err)
	assert.Len(t, todoTickets, 1)
	assert.Equal(t, ticket1.ID, todoTickets[0].ID)

	doingTickets, err := manager.List(string(StatusDoing))
	require.NoError(t, err)
	assert.Len(t, doingTickets, 1)
	assert.Equal(t, ticket2.ID, doingTickets[0].ID)
}

func TestManagerUpdate(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create ticket
	ticket, err := manager.Create("test-ticket")
	require.NoError(t, err)

	// Update ticket
	ticket.Description = "Updated description"
	ticket.Priority = 1
	err = manager.Update(ticket)
	require.NoError(t, err)

	// Reload and verify
	retrieved, err := manager.Get(ticket.ID)
	require.NoError(t, err)

	assert.Equal(t, "Updated description", retrieved.Description)
	assert.Equal(t, 1, retrieved.Priority)
}

func TestManagerCurrentTicket(t *testing.T) {
	manager, tmpDir := setupTestManager(t)

	// Initially no current ticket
	current, err := manager.GetCurrentTicket()
	require.NoError(t, err)
	assert.Nil(t, current)

	// Create and set current ticket
	ticket, err := manager.Create("test-ticket")
	require.NoError(t, err)

	err = manager.SetCurrentTicket(ticket)
	require.NoError(t, err)

	// Verify symlink exists
	linkPath := filepath.Join(tmpDir, "current-ticket.md")
	_, err = os.Lstat(linkPath)
	require.NoError(t, err)

	// Get current ticket
	current, err = manager.GetCurrentTicket()
	require.NoError(t, err)
	require.NotNil(t, current)
	assert.Equal(t, ticket.ID, current.ID)

	// Clear current ticket
	err = manager.SetCurrentTicket(nil)
	require.NoError(t, err)

	// Verify symlink removed
	_, err = os.Lstat(linkPath)
	assert.True(t, os.IsNotExist(err))
}
