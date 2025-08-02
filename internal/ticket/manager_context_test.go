package ticket

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/config"
)

// TestManagerOperationsWithCancelledContext tests all manager operations with cancelled context
func TestManagerOperationsWithCancelledContext(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create a ticket for operations that need one
	ctx := context.Background()
	existingTicket, err := manager.Create(ctx, "existing-ticket")
	require.NoError(t, err)

	// Set current ticket for operations that need it
	err = manager.SetCurrentTicket(ctx, existingTicket)
	require.NoError(t, err)

	tests := []struct {
		name      string
		operation func(context.Context) error
	}{
		{
			name: "Create",
			operation: func(ctx context.Context) error {
				_, err := manager.Create(ctx, "new-ticket")
				return err
			},
		},
		{
			name: "Get",
			operation: func(ctx context.Context) error {
				_, err := manager.Get(ctx, existingTicket.ID)
				return err
			},
		},
		{
			name: "List",
			operation: func(ctx context.Context) error {
				_, err := manager.List(ctx, StatusFilterAll)
				return err
			},
		},
		{
			name: "Update",
			operation: func(ctx context.Context) error {
				existingTicket.Content = "Updated"
				return manager.Update(ctx, existingTicket)
			},
		},
		{
			name: "GetCurrentTicket",
			operation: func(ctx context.Context) error {
				_, err := manager.GetCurrentTicket(ctx)
				return err
			},
		},
		{
			name: "SetCurrentTicket",
			operation: func(ctx context.Context) error {
				return manager.SetCurrentTicket(ctx, existingTicket)
			},
		},
		{
			name: "ReadContent",
			operation: func(ctx context.Context) error {
				_, err := manager.ReadContent(ctx, existingTicket.ID)
				return err
			},
		},
		{
			name: "WriteContent",
			operation: func(ctx context.Context) error {
				return manager.WriteContent(ctx, existingTicket.ID, "New content")
			},
		},
		{
			name: "FindTicket",
			operation: func(ctx context.Context) error {
				_, err := manager.FindTicket(ctx, existingTicket.ID)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := tt.operation(ctx)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "operation cancelled")

			// Verify error is properly wrapped
			assert.True(t, errors.Is(err, context.Canceled))
		})
	}
}

// TestFileOperationsWithCancelledContext tests file operation helpers with cancelled context
func TestFileOperationsWithCancelledContext(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("test content")

	// Create a test file
	err := os.WriteFile(testFile, testContent, 0644)
	require.NoError(t, err)

	tests := []struct {
		name      string
		operation func(context.Context) error
	}{
		{
			name: "readFileWithContext",
			operation: func(ctx context.Context) error {
				_, err := readFileWithContext(ctx, testFile)
				return err
			},
		},
		{
			name: "writeFileWithContext",
			operation: func(ctx context.Context) error {
				return writeFileWithContext(ctx, testFile, []byte("new content"), 0644)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := tt.operation(ctx)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "operation cancelled")
			assert.True(t, errors.Is(err, context.Canceled))
		})
	}
}

// TestConcurrentManagerOperations tests concurrent manager operations with cancellation
func TestConcurrentManagerOperations(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create some tickets first
	for i := 0; i < 5; i++ {
		_, err := manager.Create(context.Background(), fmt.Sprintf("ticket-%d", i))
		require.NoError(t, err)
	}

	// Create coordination channel
	startChan := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	errChan := make(chan error, 50)

	// Start multiple concurrent operations
	for i := 0; i < 10; i++ {
		wg.Add(5)

		// List operation
		go func() {
			defer wg.Done()
			<-startChan
			_, err := manager.List(ctx, StatusFilterAll)
			if err != nil {
				errChan <- err
			}
		}()

		// Create operation
		go func(id int) {
			defer wg.Done()
			<-startChan
			_, err := manager.Create(ctx, fmt.Sprintf("concurrent-%d", id))
			if err != nil {
				errChan <- err
			}
		}(i)

		// Read operation
		go func() {
			defer wg.Done()
			<-startChan
			_, err := manager.Get(ctx, "ticket-0")
			if err != nil {
				errChan <- err
			}
		}()

		// Find operation
		go func() {
			defer wg.Done()
			<-startChan
			_, err := manager.FindTicket(ctx, "ticket-1")
			if err != nil {
				errChan <- err
			}
		}()

		// Current ticket operation
		go func() {
			defer wg.Done()
			<-startChan
			_, err := manager.GetCurrentTicket(ctx)
			if err != nil {
				errChan <- err
			}
		}()
	}

	// Give goroutines time to block
	time.Sleep(10 * time.Millisecond)

	// Cancel context first
	cancel()

	// Then start all operations
	close(startChan)

	// Wait for all operations
	wg.Wait()
	close(errChan)

	// Count cancelled operations
	errorCount := 0
	for err := range errChan {
		if err != nil {
			errorCount++
			assert.True(t, errors.Is(err, context.Canceled))
		}
	}

	// All operations should have been cancelled
	assert.Equal(t, 50, errorCount, "All operations should fail with cancelled context")
}

// TestContextTimeoutScenarios tests various timeout scenarios
func TestContextTimeoutScenarios(t *testing.T) {
	manager, _ := setupTestManager(t)

	tests := []struct {
		name      string
		timeout   time.Duration
		preDelay  time.Duration
		operation func(context.Context) error
		wantErr   bool
	}{
		{
			name:     "immediate timeout",
			timeout:  1 * time.Microsecond,
			preDelay: 5 * time.Millisecond,
			operation: func(ctx context.Context) error {
				_, err := manager.Create(ctx, "timeout-test")
				return err
			},
			wantErr: true,
		},
		{
			name:     "sufficient timeout",
			timeout:  1 * time.Second,
			preDelay: 0,
			operation: func(ctx context.Context) error {
				_, err := manager.Create(ctx, "no-timeout-test")
				return err
			},
			wantErr: false,
		},
		{
			name:     "timeout during file write",
			timeout:  10 * time.Millisecond,
			preDelay: 0,
			operation: func(ctx context.Context) error {
				// Create large content that takes time to write
				largeContent := make([]byte, 50*1024*1024) // 50MB
				return writeFileWithContext(ctx, "/tmp/large-test.txt", largeContent, 0644)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			if tt.preDelay > 0 {
				time.Sleep(tt.preDelay)
			}

			err := tt.operation(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestPartialOperationHandling tests handling of partial operations when cancelled
func TestPartialOperationHandling(t *testing.T) {
	_, tmpDir := setupTestManager(t)

	// Test partial file write
	t.Run("partial file write", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		testFile := filepath.Join(tmpDir, "partial-write.txt")

		// Start writing in a goroutine
		errChan := make(chan error, 1)
		go func() {
			largeData := make([]byte, 10*1024*1024) // 10MB
			errChan <- writeFileWithContext(ctx, testFile, largeData, 0644)
		}()

		// Cancel quickly
		time.Sleep(1 * time.Millisecond)
		cancel()

		err := <-errChan
		assert.Error(t, err)

		// If file was created, it should be cleaned up or incomplete
		if _, statErr := os.Stat(testFile); statErr == nil {
			// File exists, verify it's not the full size
			info, _ := os.Stat(testFile)
			assert.Less(t, info.Size(), int64(10*1024*1024), "File should not be fully written")
		}
	})
}

// TestContextWithValues tests that context values are preserved through operations
func TestContextWithValues(t *testing.T) {
	type contextKey string
	const userKey contextKey = "user"
	const requestIDKey contextKey = "request-id"

	manager, _ := setupTestManager(t)

	ctx := context.WithValue(context.Background(), userKey, "test-user")
	ctx = context.WithValue(ctx, requestIDKey, "12345")
	ctx, cancel := context.WithCancel(ctx)
	cancel()

	// Values should be preserved even after cancellation
	assert.Equal(t, "test-user", ctx.Value(userKey))
	assert.Equal(t, "12345", ctx.Value(requestIDKey))

	// But operations should fail
	_, err := manager.Create(ctx, "value-test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation cancelled")
}

// BenchmarkManagerContextOverhead benchmarks context checking overhead
func BenchmarkManagerContextOverhead(b *testing.B) {
	tmpDir := b.TempDir()
	cfg := config.Default()
	cfg.Tickets.Dir = "tickets"
	manager := NewManager(cfg, tmpDir)
	ctx := context.Background()

	// Create a ticket for benchmarking Get operation
	ticket, err := manager.Create(ctx, "bench-ticket")
	require.NoError(b, err)

	b.Run("Get", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = manager.Get(ctx, ticket.ID)
		}
	})

	b.Run("List", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = manager.List(ctx, StatusFilterAll)
		}
	})

	b.Run("FindTicket", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = manager.FindTicket(ctx, ticket.ID)
		}
	})
}

// BenchmarkCancelledContextOverhead benchmarks operations with cancelled contexts
func BenchmarkCancelledContextOverhead(b *testing.B) {
	tmpDir := b.TempDir()
	cfg := config.Default()
	cfg.Tickets.Dir = "tickets"
	manager := NewManager(cfg, tmpDir)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, _ = manager.List(ctx, StatusFilterAll)
	}
}

// BenchmarkFileOperationsWithContext benchmarks file operations with context
func BenchmarkFileOperationsWithContext(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "bench.txt")
	testData := []byte("benchmark test data")

	// Write initial file
	err := os.WriteFile(testFile, testData, 0644)
	require.NoError(b, err)

	b.Run("readFileWithContext", func(b *testing.B) {
		ctx := context.Background()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = readFileWithContext(ctx, testFile)
		}
	})

	b.Run("writeFileWithContext", func(b *testing.B) {
		ctx := context.Background()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = writeFileWithContext(ctx, testFile, testData, 0644)
		}
	})
}
