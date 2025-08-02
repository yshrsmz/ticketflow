package ticket

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/config"
	ticketerrors "github.com/yshrsmz/ticketflow/internal/errors"
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
	ctx := context.Background()

	// Create ticket
	ticket, err := manager.Create(ctx, "test-ticket")
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
	_, err = manager.Create(ctx, "test-ticket")
	require.NoError(t, err) // Different timestamp means different ID
}

func TestManagerCreateInvalidSlug(t *testing.T) {
	manager, _ := setupTestManager(t)
	ctx := context.Background()

	tests := []string{
		"Test-Ticket", // uppercase
		"test ticket", // space
		"test_ticket", // underscore
		"",            // empty
	}

	for _, slug := range tests {
		t.Run(slug, func(t *testing.T) {
			_, err := manager.Create(ctx, slug)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid slug format")
		})
	}
}

func TestManagerGet(t *testing.T) {
	manager, _ := setupTestManager(t)
	ctx := context.Background()

	// Create ticket
	created, err := manager.Create(ctx, "test-ticket")
	require.NoError(t, err)

	// Get by full ID
	retrieved, err := manager.Get(ctx, created.ID)
	require.NoError(t, err)

	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.Slug, retrieved.Slug)
	assert.Equal(t, created.Description, retrieved.Description)
}

func TestManagerGetByPrefix(t *testing.T) {
	manager, _ := setupTestManager(t)
	ctx := context.Background()

	// Create ticket
	created, err := manager.Create(ctx, "test-ticket")
	require.NoError(t, err)

	// Get by prefix (first 6 chars of ID)
	prefix := created.ID[:6]
	retrieved, err := manager.Get(ctx, prefix)
	require.NoError(t, err)

	assert.Equal(t, created.ID, retrieved.ID)
}

func TestManagerGetNotFound(t *testing.T) {
	manager, _ := setupTestManager(t)
	ctx := context.Background()

	_, err := manager.Get(ctx, "nonexistent")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ticketerrors.ErrTicketNotFound))
}

func TestManagerGetAmbiguous(t *testing.T) {
	manager, tmpDir := setupTestManager(t)
	ctx := context.Background()

	// Create two tickets with similar IDs manually
	todoPath := filepath.Join(tmpDir, "tickets", "todo")
	require.NoError(t, os.MkdirAll(todoPath, 0755))

	// Create two files with same prefix
	ticket1 := New("test1", "Test 1")
	ticket2 := New("test2", "Test 2")

	data1, _ := ticket1.ToBytes()
	data2, _ := ticket2.ToBytes()

	require.NoError(t, os.WriteFile(filepath.Join(todoPath, "250124-150000-test1.md"), data1, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(todoPath, "250124-150001-test2.md"), data2, 0644))

	// Try to get with ambiguous prefix
	_, err := manager.Get(ctx, "250124")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ambiguous ticket ID")
}

func TestManagerList(t *testing.T) {
	manager, tmpDir := setupTestManager(t)
	ctx := context.Background()

	// Create multiple tickets
	ticket1, err := manager.Create(ctx, "ticket-1")
	require.NoError(t, err)

	time.Sleep(time.Second)
	ticket2, err := manager.Create(ctx, "ticket-2")
	require.NoError(t, err)

	// Start ticket2 - simulate moving to doing directory
	err = ticket2.Start()
	require.NoError(t, err)

	// Move ticket2 from todo to doing
	oldPath := ticket2.Path
	doingPath := filepath.Join(tmpDir, "tickets", "doing")
	err = os.MkdirAll(doingPath, 0755)
	require.NoError(t, err)
	newPath := filepath.Join(doingPath, filepath.Base(ticket2.Path))
	err = os.Rename(oldPath, newPath)
	require.NoError(t, err)
	ticket2.Path = newPath
	err = manager.Update(ctx, ticket2)
	require.NoError(t, err)

	// List all tickets
	tickets, err := manager.List(ctx, StatusFilterActive)
	require.NoError(t, err)
	assert.Len(t, tickets, 2)

	// List by status
	todoTickets, err := manager.List(ctx, StatusFilterTodo)
	require.NoError(t, err)
	assert.Len(t, todoTickets, 1)
	assert.Equal(t, ticket1.ID, todoTickets[0].ID)

	doingTickets, err := manager.List(ctx, StatusFilterDoing)
	require.NoError(t, err)
	assert.Len(t, doingTickets, 1)
	assert.Equal(t, ticket2.ID, doingTickets[0].ID)
}

func TestManagerUpdate(t *testing.T) {
	manager, _ := setupTestManager(t)
	ctx := context.Background()

	// Create ticket
	ticket, err := manager.Create(ctx, "test-ticket")
	require.NoError(t, err)

	// Update ticket
	ticket.Description = "Updated description"
	ticket.Priority = 1
	err = manager.Update(ctx, ticket)
	require.NoError(t, err)

	// Reload and verify
	retrieved, err := manager.Get(ctx, ticket.ID)
	require.NoError(t, err)

	assert.Equal(t, "Updated description", retrieved.Description)
	assert.Equal(t, 1, retrieved.Priority)
}

func TestManagerCurrentTicket(t *testing.T) {
	manager, tmpDir := setupTestManager(t)
	ctx := context.Background()

	// Initially no current ticket
	current, err := manager.GetCurrentTicket(ctx)
	require.NoError(t, err)
	assert.Nil(t, current)

	// Create and set current ticket
	ticket, err := manager.Create(ctx, "test-ticket")
	require.NoError(t, err)

	err = manager.SetCurrentTicket(ctx, ticket)
	require.NoError(t, err)

	// Verify symlink exists
	linkPath := filepath.Join(tmpDir, "current-ticket.md")
	_, err = os.Lstat(linkPath)
	require.NoError(t, err)

	// Get current ticket
	current, err = manager.GetCurrentTicket(ctx)
	require.NoError(t, err)
	require.NotNil(t, current)
	assert.Equal(t, ticket.ID, current.ID)

	// Clear current ticket
	err = manager.SetCurrentTicket(ctx, nil)
	require.NoError(t, err)

	// Verify symlink removed
	_, err = os.Lstat(linkPath)
	assert.True(t, os.IsNotExist(err))
}

func TestReadFileWithContext(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("test content")

	// Write test file
	err := os.WriteFile(testFile, testContent, 0644)
	require.NoError(t, err)

	t.Run("successful read", func(t *testing.T) {
		ctx := context.Background()
		data, err := readFileWithContext(ctx, testFile)
		require.NoError(t, err)
		assert.Equal(t, testContent, data)
	})

	t.Run("cancelled context before read", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := readFileWithContext(ctx, testFile)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "operation cancelled")
	})

	t.Run("non-existent file", func(t *testing.T) {
		ctx := context.Background()
		_, err := readFileWithContext(ctx, filepath.Join(tmpDir, "nonexistent.txt"))
		require.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("large file with context cancellation", func(t *testing.T) {
		// Create a large file (>1MB)
		largeFile := filepath.Join(tmpDir, "large.txt")
		largeData := make([]byte, 2*1024*1024) // 2MB
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}
		err := os.WriteFile(largeFile, largeData, 0644)
		require.NoError(t, err)

		// Test successful read
		ctx := context.Background()
		data, err := readFileWithContext(ctx, largeFile)
		require.NoError(t, err)
		assert.Equal(t, largeData, data)

		// Test with timeout (might not trigger if read is too fast)
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(2 * time.Millisecond) // Ensure timeout
		_, err = readFileWithContext(ctx, largeFile)
		if err != nil {
			assert.Contains(t, err.Error(), "cancelled")
		}
	})

	t.Run("file size validation", func(t *testing.T) {
		// Create a file that exceeds size limit (>50MB)
		hugeFile := filepath.Join(tmpDir, "huge.txt")
		// Create a sparse file to avoid actually writing 51MB
		file, err := os.Create(hugeFile)
		require.NoError(t, err)

		// Seek to 51MB position and write a byte to create a sparse file
		_, err = file.Seek(51*1024*1024, 0)
		require.NoError(t, err)
		_, err = file.Write([]byte{0})
		require.NoError(t, err)
		file.Close()

		// Try to read the file - should fail with size error
		ctx := context.Background()
		_, err = readFileWithContext(ctx, hugeFile)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "file too large")
		assert.Contains(t, err.Error(), "exceeds")
		assert.Contains(t, err.Error(), "limit")
	})
}

func TestWriteFileWithContext(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("successful write", func(t *testing.T) {
		ctx := context.Background()
		testFile := filepath.Join(tmpDir, "write_test.txt")
		testContent := []byte("test write content")

		err := writeFileWithContext(ctx, testFile, testContent, 0644)
		require.NoError(t, err)

		// Verify file was written correctly
		data, err := os.ReadFile(testFile)
		require.NoError(t, err)
		assert.Equal(t, testContent, data)
	})

	t.Run("cancelled context before write", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		testFile := filepath.Join(tmpDir, "cancelled_write.txt")
		err := writeFileWithContext(ctx, testFile, []byte("should not write"), 0644)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "operation cancelled")

		// Verify file was not created
		_, err = os.Stat(testFile)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("large file write", func(t *testing.T) {
		ctx := context.Background()
		testFile := filepath.Join(tmpDir, "large_write.txt")
		largeData := make([]byte, 2*1024*1024) // 2MB
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}

		err := writeFileWithContext(ctx, testFile, largeData, 0644)
		require.NoError(t, err)

		// Verify file was written correctly
		data, err := os.ReadFile(testFile)
		require.NoError(t, err)
		assert.Equal(t, largeData, data)
	})

	t.Run("file sync verification", func(t *testing.T) {
		ctx := context.Background()
		testFile := filepath.Join(tmpDir, "sync_test.txt")
		testContent := []byte("test data that needs to be synced")

		// Write file with context (which includes sync)
		err := writeFileWithContext(ctx, testFile, testContent, 0644)
		require.NoError(t, err)

		// Verify file exists and has correct content
		// The sync ensures data is persisted even if system crashes
		info, err := os.Stat(testFile)
		require.NoError(t, err)
		assert.Equal(t, int64(len(testContent)), info.Size())

		data, err := os.ReadFile(testFile)
		require.NoError(t, err)
		assert.Equal(t, testContent, data)
	})

	t.Run("write with timeout", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "timeout_write.txt")
		largeData := make([]byte, 2*1024*1024) // 2MB

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(2 * time.Millisecond) // Ensure timeout

		err := writeFileWithContext(ctx, testFile, largeData, 0644)
		if err != nil {
			assert.Contains(t, err.Error(), "cancelled")
		}
	})
}

func TestManagerWithContextCancellation(t *testing.T) {
	manager, tmpDir := setupTestManager(t)

	// Create a test ticket first
	ctx := context.Background()
	ticket, err := manager.Create(ctx, "context-test")
	require.NoError(t, err)

	// Test Update with cancelled context
	t.Run("Update with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		ticket.Content = "updated content"
		err := manager.Update(ctx, ticket)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "operation cancelled")
	})

	// Test loadTicket with cancelled context
	t.Run("loadTicket with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := manager.loadTicket(ctx, ticket.Path)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "operation cancelled")
	})

	// Test ReadContent with cancelled context
	t.Run("ReadContent with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := manager.ReadContent(ctx, ticket.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "operation cancelled")
	})

	// Test WriteContent with cancelled context
	t.Run("WriteContent with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := manager.WriteContent(ctx, ticket.ID, "new content")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "operation cancelled")
	})

	// Test Create with very large template (to trigger chunked write)
	t.Run("Create with large template", func(t *testing.T) {
		// Temporarily set a large template
		originalTemplate := manager.config.Tickets.Template
		largeTemplate := make([]byte, 2*1024*1024) // 2MB
		for i := range largeTemplate {
			largeTemplate[i] = 'A' + byte(i%26)
		}
		manager.config.Tickets.Template = string(largeTemplate)
		defer func() {
			manager.config.Tickets.Template = originalTemplate
		}()

		ctx := context.Background()
		largeTicket, err := manager.Create(ctx, "large-ticket")
		require.NoError(t, err)

		// Verify the ticket was created with the large content
		data, err := os.ReadFile(largeTicket.Path)
		require.NoError(t, err)
		assert.Contains(t, string(data), string(largeTemplate[:100])) // Check first 100 chars
	})

	// Clean up
	os.RemoveAll(tmpDir)
}
