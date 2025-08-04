package ticket

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yshrsmz/ticketflow/internal/config"
)

// BenchmarkManagerCreate benchmarks ticket creation
func BenchmarkManagerCreate(b *testing.B) {
	tmpDir := b.TempDir()
	cfg := config.Default()
	cfg.Tickets.Dir = "tickets"
	manager := NewManager(cfg, tmpDir)
	ctx := context.Background()

	// Create ticket directory structure
	todoDir := filepath.Join(tmpDir, cfg.Tickets.Dir, "todo")
	if err := os.MkdirAll(todoDir, 0755); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		slug := fmt.Sprintf("benchmark-ticket-%d", i)
		_, err := manager.Create(ctx, slug)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkManagerGet benchmarks getting a single ticket
func BenchmarkManagerGet(b *testing.B) {
	tmpDir := b.TempDir()
	cfg := config.Default()
	cfg.Tickets.Dir = "tickets"
	manager := NewManager(cfg, tmpDir)
	ctx := context.Background()

	// Create ticket directory structure
	todoDir := filepath.Join(tmpDir, cfg.Tickets.Dir, "todo")
	if err := os.MkdirAll(todoDir, 0755); err != nil {
		b.Fatal(err)
	}

	// Create a ticket to benchmark getting
	ticket, err := manager.Create(ctx, "benchmark-ticket")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := manager.Get(ctx, ticket.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkManagerList benchmarks listing tickets with different filters
func BenchmarkManagerList(b *testing.B) {
	scenarios := []struct {
		name         string
		ticketCount  int
		statusFilter StatusFilter
	}{
		{"10-tickets-all", 10, StatusFilterAll},
		{"50-tickets-all", 50, StatusFilterAll},
		{"100-tickets-all", 100, StatusFilterAll},
		{"100-tickets-active", 100, StatusFilterActive},
		{"100-tickets-todo", 100, StatusFilterTodo},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			tmpDir := b.TempDir()
			cfg := config.Default()
			cfg.Tickets.Dir = "tickets"
			manager := NewManager(cfg, tmpDir)
			ctx := context.Background()

			// Create ticket directory structure
			todoDir := filepath.Join(tmpDir, cfg.Tickets.Dir, "todo")
			doingDir := filepath.Join(tmpDir, cfg.Tickets.Dir, "doing")
			doneDir := filepath.Join(tmpDir, cfg.Tickets.Dir, "done")
			for _, dir := range []string{todoDir, doingDir, doneDir} {
				if err := os.MkdirAll(dir, 0755); err != nil {
					b.Fatal(err)
				}
			}

			// Create tickets distributed across statuses
			for i := 0; i < scenario.ticketCount; i++ {
				slug := fmt.Sprintf("benchmark-ticket-%d", i)
				ticket, err := manager.Create(ctx, slug)
				if err != nil {
					b.Fatal(err)
				}

				// Move some tickets to different statuses
				switch i % 3 {
				case 1: // Move to doing
					oldPath := filepath.Join(todoDir, ticket.ID+".md")
					newPath := filepath.Join(doingDir, ticket.ID+".md")
					if err := os.Rename(oldPath, newPath); err != nil {
						b.Fatal(err)
					}
				case 2: // Move to done
					oldPath := filepath.Join(todoDir, ticket.ID+".md")
					newPath := filepath.Join(doneDir, ticket.ID+".md")
					if err := os.Rename(oldPath, newPath); err != nil {
						b.Fatal(err)
					}
				}
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := manager.List(ctx, scenario.statusFilter)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkManagerUpdate benchmarks updating tickets
func BenchmarkManagerUpdate(b *testing.B) {
	tmpDir := b.TempDir()
	cfg := config.Default()
	cfg.Tickets.Dir = "tickets"
	manager := NewManager(cfg, tmpDir)
	ctx := context.Background()

	// Create ticket directory structure
	todoDir := filepath.Join(tmpDir, cfg.Tickets.Dir, "todo")
	if err := os.MkdirAll(todoDir, 0755); err != nil {
		b.Fatal(err)
	}

	// Create a ticket to update
	ticket, err := manager.Create(ctx, "benchmark-ticket")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ticket.Content = fmt.Sprintf("Updated content %d", i)
		if err := manager.Update(ctx, ticket); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkManagerFindTicket benchmarks finding tickets
func BenchmarkManagerFindTicket(b *testing.B) {
	scenarios := []struct {
		name        string
		ticketCount int
		findIndex   int // Which ticket to find (0 = first, -1 = last)
	}{
		{"10-tickets-first", 10, 0},
		{"10-tickets-last", 10, -1},
		{"50-tickets-first", 50, 0},
		{"50-tickets-last", 50, -1},
		{"100-tickets-first", 100, 0},
		{"100-tickets-last", 100, -1},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			tmpDir := b.TempDir()
			cfg := config.Default()
			cfg.Tickets.Dir = "tickets"
			manager := NewManager(cfg, tmpDir)
			ctx := context.Background()

			// Create ticket directory structure
			todoDir := filepath.Join(tmpDir, cfg.Tickets.Dir, "todo")
			if err := os.MkdirAll(todoDir, 0755); err != nil {
				b.Fatal(err)
			}

			// Create tickets
			var ticketIDs []string
			for i := 0; i < scenario.ticketCount; i++ {
				slug := fmt.Sprintf("benchmark-ticket-%d", i)
				ticket, err := manager.Create(ctx, slug)
				if err != nil {
					b.Fatal(err)
				}
				ticketIDs = append(ticketIDs, ticket.ID)
			}

			// Determine which ticket ID to find
			findID := ticketIDs[0]
			if scenario.findIndex == -1 {
				findID = ticketIDs[len(ticketIDs)-1]
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := manager.FindTicket(ctx, findID)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkManagerReadWriteContent benchmarks reading and writing ticket content
func BenchmarkManagerReadWriteContent(b *testing.B) {
	tmpDir := b.TempDir()
	cfg := config.Default()
	cfg.Tickets.Dir = "tickets"
	manager := NewManager(cfg, tmpDir)
	ctx := context.Background()

	// Create ticket directory structure
	todoDir := filepath.Join(tmpDir, cfg.Tickets.Dir, "todo")
	if err := os.MkdirAll(todoDir, 0755); err != nil {
		b.Fatal(err)
	}

	// Create a ticket
	ticket, err := manager.Create(ctx, "benchmark-ticket")
	if err != nil {
		b.Fatal(err)
	}

	// Test content of various sizes
	contentSizes := []struct {
		name string
		size int
	}{
		{"small-100B", 100},
		{"medium-1KB", 1024},
		{"large-10KB", 10 * 1024},
	}

	for _, contentSize := range contentSizes {
		// Generate content of specified size
		content := strings.Repeat("x", contentSize.size)

		b.Run("WriteContent-"+contentSize.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if err := manager.WriteContent(ctx, ticket.ID, content); err != nil {
					b.Fatal(err)
				}
			}
		})

		// Write content once for read benchmark
		if err := manager.WriteContent(ctx, ticket.ID, content); err != nil {
			b.Fatal(err)
		}

		b.Run("ReadContent-"+contentSize.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, err := manager.ReadContent(ctx, ticket.ID)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkManagerCurrentTicket benchmarks getting and setting current ticket
func BenchmarkManagerCurrentTicket(b *testing.B) {
	tmpDir := b.TempDir()
	cfg := config.Default()
	cfg.Tickets.Dir = "tickets"
	manager := NewManager(cfg, tmpDir)
	ctx := context.Background()

	// Create ticket directory structure
	todoDir := filepath.Join(tmpDir, cfg.Tickets.Dir, "todo")
	if err := os.MkdirAll(todoDir, 0755); err != nil {
		b.Fatal(err)
	}

	// Create a ticket
	ticket, err := manager.Create(ctx, "benchmark-ticket")
	if err != nil {
		b.Fatal(err)
	}

	b.Run("SetCurrentTicket", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if err := manager.SetCurrentTicket(ctx, ticket); err != nil {
				b.Fatal(err)
			}
		}
	})

	// Set current ticket once for get benchmark
	if err := manager.SetCurrentTicket(ctx, ticket); err != nil {
		b.Fatal(err)
	}

	b.Run("GetCurrentTicket", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, err := manager.GetCurrentTicket(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

