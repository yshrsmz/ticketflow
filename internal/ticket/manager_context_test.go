package ticket

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/config"
)

// TestCreateWithCancelledContext tests Create with cancelled context
func TestCreateWithCancelledContext(t *testing.T) {
	manager, _ := setupTestManager(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ticket, err := manager.Create(ctx, "cancelled-ticket")
	assert.Error(t, err)
	assert.Nil(t, ticket)
	assert.Contains(t, err.Error(), "operation cancelled")
}

// TestGetWithCancelledContext tests Get with cancelled context
func TestGetWithCancelledContext(t *testing.T) {
	manager, _ := setupTestManager(t)

	// First create a ticket
	ctx := context.Background()
	ticket, err := manager.Create(ctx, "get-test")
	require.NoError(t, err)

	// Now try to get with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	retrievedTicket, err := manager.Get(ctx, ticket.ID)
	assert.Error(t, err)
	assert.Nil(t, retrievedTicket)
	assert.Contains(t, err.Error(), "operation cancelled")
}

// TestListWithCancelledContext tests List with cancelled context
func TestListWithCancelledContext(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create a ticket first
	ctx := context.Background()
	_, err := manager.Create(ctx, "list-test")
	require.NoError(t, err)

	// Now try to list with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tickets, err := manager.List(ctx, StatusFilterAll)
	assert.Error(t, err)
	assert.Nil(t, tickets)
	assert.Contains(t, err.Error(), "operation cancelled")
}

// TestUpdateWithCancelledContext tests Update with cancelled context
func TestUpdateWithCancelledContext(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create a ticket first
	ctx := context.Background()
	ticket, err := manager.Create(ctx, "update-test")
	require.NoError(t, err)

	// Modify the ticket
	ticket.Content = "Updated content"

	// Now try to update with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = manager.Update(ctx, ticket)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation cancelled")
}

// TestGetCurrentTicketWithCancelledContext tests GetCurrentTicket with cancelled context
func TestGetCurrentTicketWithCancelledContext(t *testing.T) {
	manager, _ := setupTestManager(t)

	// First set a current ticket
	ctx := context.Background()
	ticket, err := manager.Create(ctx, "current-test")
	require.NoError(t, err)
	err = manager.SetCurrentTicket(ctx, ticket)
	require.NoError(t, err)

	// Now try to get current with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	currentTicket, err := manager.GetCurrentTicket(ctx)
	assert.Error(t, err)
	assert.Nil(t, currentTicket)
	assert.Contains(t, err.Error(), "operation cancelled")
}

// TestSetCurrentTicketWithCancelledContext tests SetCurrentTicket with cancelled context
func TestSetCurrentTicketWithCancelledContext(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create a ticket first
	ctx := context.Background()
	ticket, err := manager.Create(ctx, "set-current-test")
	require.NoError(t, err)

	// Now try to set current with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = manager.SetCurrentTicket(ctx, ticket)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation cancelled")
}

// TestReadContentWithCancelledContext tests ReadContent with cancelled context
func TestReadContentWithCancelledContext(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create a ticket first
	ctx := context.Background()
	ticket, err := manager.Create(ctx, "read-content-test")
	require.NoError(t, err)

	// Now try to read content with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	content, err := manager.ReadContent(ctx, ticket.ID)
	assert.Error(t, err)
	assert.Empty(t, content)
	assert.Contains(t, err.Error(), "operation cancelled")
}

// TestWriteContentWithCancelledContext tests WriteContent with cancelled context
func TestWriteContentWithCancelledContext(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create a ticket first
	ctx := context.Background()
	ticket, err := manager.Create(ctx, "write-content-test")
	require.NoError(t, err)

	// Now try to write content with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = manager.WriteContent(ctx, ticket.ID, "New content")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation cancelled")
}

// TestFindTicketWithCancelledContext tests FindTicket with cancelled context
func TestFindTicketWithCancelledContext(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create a ticket first
	ctx := context.Background()
	ticket, err := manager.Create(ctx, "find-test")
	require.NoError(t, err)

	// Now try to find with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	path, err := manager.FindTicket(ctx, ticket.ID)
	assert.Error(t, err)
	assert.Empty(t, path)
	assert.Contains(t, err.Error(), "operation cancelled")
}

// TestContextTimeoutDuringFileOperation tests timeout during file operations
func TestContextTimeoutDuringFileOperation(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create a ticket first with no timeout
	ctx := context.Background()
	ticket, err := manager.Create(ctx, "timeout-test")
	require.NoError(t, err)

	// Create a context with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for timeout to trigger
	time.Sleep(5 * time.Millisecond)

	// Try to write content with expired context
	err = manager.WriteContent(ctx, ticket.ID, "Some content")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation cancelled")
}

// TestReadFileWithContextHelper tests the readFileWithContext helper
func TestReadFileWithContextHelper(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("test content")

	err := os.WriteFile(testFile, testContent, 0644)
	require.NoError(t, err)

	// Test normal read
	ctx := context.Background()
	data, err := readFileWithContext(ctx, testFile)
	assert.NoError(t, err)
	assert.Equal(t, testContent, data)

	// Test cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	data, err = readFileWithContext(ctx, testFile)
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "operation cancelled")
}

// TestWriteFileWithContextHelper tests the writeFileWithContext helper
func TestWriteFileWithContextHelper(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-write.txt")
	testContent := []byte("test write content")

	// Test normal write
	ctx := context.Background()
	err := writeFileWithContext(ctx, testFile, testContent, 0644)
	assert.NoError(t, err)

	// Verify file was written
	data, err := os.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Equal(t, testContent, data)

	// Test cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	testFile2 := filepath.Join(tmpDir, "test-write-cancelled.txt")
	err = writeFileWithContext(ctx, testFile2, testContent, 0644)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation cancelled")

	// File should not exist
	_, err = os.Stat(testFile2)
	assert.True(t, os.IsNotExist(err))
}

// BenchmarkContextCheckInCreate benchmarks context checking overhead in Create
func BenchmarkContextCheckInCreate(b *testing.B) {
	// Setup
	tmpDir := b.TempDir()
	cfg := config.Default()
	cfg.Tickets.Dir = "tickets"
	manager := NewManager(cfg, tmpDir)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		slug := fmt.Sprintf("bench-ticket-%d", i)
		_, _ = manager.Create(ctx, slug)
	}
}

// BenchmarkCancelledContextInList benchmarks List with cancelled context
func BenchmarkCancelledContextInList(b *testing.B) {
	// Setup
	tmpDir := b.TempDir()
	cfg := config.Default()
	cfg.Tickets.Dir = "tickets"
	manager := NewManager(cfg, tmpDir)

	// Create some tickets
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		_, err := manager.Create(ctx, fmt.Sprintf("bench-ticket-%d", i))
		require.NoError(b, err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, _ = manager.List(ctx, StatusFilterAll)
	}
}
