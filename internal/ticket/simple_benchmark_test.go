package ticket

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/yshrsmz/ticketflow/internal/config"
)

// BenchmarkManagerList benchmarks the most critical operation - listing tickets
// Run with: go test -bench=BenchmarkManagerList ./internal/ticket
func BenchmarkManagerList(b *testing.B) {
	scenarios := []struct {
		name         string
		ticketCount  int
		statusFilter StatusFilter
	}{
		{"10-tickets", 10, StatusFilterAll},
		{"50-tickets", 50, StatusFilterAll},
		{"100-tickets", 100, StatusFilterAll},
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
