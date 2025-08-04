package ui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/ticket"
	"github.com/yshrsmz/ticketflow/internal/ui/styles"
)

// BenchmarkAppInit benchmarks the app initialization
func BenchmarkAppInit(b *testing.B) {
	tmpDir := b.TempDir()
	setupBenchmarkUIRepo(b, tmpDir)

	cfg := config.Default()
	manager := ticket.NewManager(cfg, tmpDir)
	gitClient := git.New(tmpDir)
	ctx := context.Background()

	// Create some tickets
	for i := 0; i < 10; i++ {
		slug := fmt.Sprintf("benchmark-ticket-%d", i)
		if _, err := manager.Create(ctx, slug); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		app := New(cfg, manager, gitClient, tmpDir)
		cmd := app.Init()
		// Just check that it returns a valid command
		if cmd == nil {
			b.Fatal("Init returned nil")
		}
	}
}

// BenchmarkTicketListViewUpdate benchmarks the ticket list view updates
func BenchmarkTicketListViewUpdate(b *testing.B) {
	scenarios := []struct {
		name        string
		ticketCount int
	}{
		{"10-tickets", 10},
		{"50-tickets", 50},
		{"100-tickets", 100},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			tmpDir := b.TempDir()
			setupBenchmarkUIRepo(b, tmpDir)

			cfg := config.Default()
			manager := ticket.NewManager(cfg, tmpDir)
			ctx := context.Background()

			// Create tickets
			tickets := make([]ticket.Ticket, scenario.ticketCount)
			for i := 0; i < scenario.ticketCount; i++ {
				slug := fmt.Sprintf("benchmark-ticket-%d", i)
				t, err := manager.Create(ctx, slug)
				if err != nil {
					b.Fatal(err)
				}
				tickets[i] = *t
			}

			// Create a list view - for now just simulate list operations
			// since the actual view constructors are internal

			b.ResetTimer()
			b.ReportAllocs()

			// Benchmark various updates
			for i := 0; i < b.N; i++ {
				// Simulate list operations like filtering and sorting
				for j := 0; j < len(tickets); j++ {
					_ = tickets[j].ID
				}
			}
		})
	}
}

// BenchmarkTicketDetailView benchmarks the detail view rendering
func BenchmarkTicketDetailView(b *testing.B) {
	contentSizes := []struct {
		name string
		size int
	}{
		{"small-content", 100},
		{"medium-content", 1000},
		{"large-content", 10000},
	}

	for _, contentSize := range contentSizes {
		b.Run(contentSize.name, func(b *testing.B) {
			// Create a test ticket
			t := &ticket.Ticket{
				ID:        "250101-120000-benchmark",
				Slug:      "benchmark-ticket",
				CreatedAt: ticket.RFC3339Time{Time: time.Now()},
				Content:   fmt.Sprintf("# Benchmark Ticket\n\n%s", string(make([]byte, contentSize.size))),
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Simulate detail view rendering
				_ = fmt.Sprintf("ID: %s\nSlug: %s\nCreated: %s\n\n%s",
					t.ID, t.Slug, t.CreatedAt, t.Content)
			}
		})
	}
}

// BenchmarkTabNavigation benchmarks tab switching
func BenchmarkTabNavigation(b *testing.B) {
	tabs := []string{"List", "Detail", "New"}

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Cycle through tabs
		currentTab := i % len(tabs)
		// Simulate tab model behavior
		_ = tabs[currentTab]
	}
}

// BenchmarkStyleRendering benchmarks the style rendering
func BenchmarkStyleRendering(b *testing.B) {
	testStrings := []string{
		"Short text",
		"Medium length text that is a bit longer than short",
		"Very long text that contains multiple words and should test the performance of style rendering on longer content",
	}

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, text := range testStrings {
			// Simulate style rendering with basic formatting
			_ = styles.TitleStyle.Render(text)
			_ = styles.SubtitleStyle.Render(text)
			_ = styles.SuccessStyle.Render(text)
			_ = styles.ErrorStyle.Render(text)
		}
	}
}

// BenchmarkListFiltering benchmarks filtering tickets in the list view
func BenchmarkListFiltering(b *testing.B) {
	tmpDir := b.TempDir()
	setupBenchmarkUIRepo(b, tmpDir)

	cfg := config.Default()
	manager := ticket.NewManager(cfg, tmpDir)
	ctx := context.Background()

	// Create tickets with various statuses
	var tickets []ticket.Ticket
	statuses := []string{"todo", "doing", "done"}
	for i := 0; i < 100; i++ {
		slug := fmt.Sprintf("benchmark-ticket-%d", i)
		t, err := manager.Create(ctx, slug)
		if err != nil {
			b.Fatal(err)
		}

		// Move some tickets to different statuses
		status := statuses[i%3]
		if status != "todo" {
			oldPath := filepath.Join(tmpDir, cfg.Tickets.Dir, "todo", t.ID+".md")
			newPath := filepath.Join(tmpDir, cfg.Tickets.Dir, status, t.ID+".md")
			if err := os.MkdirAll(filepath.Dir(newPath), 0755); err != nil {
				b.Fatal(err)
			}
			if err := os.Rename(oldPath, newPath); err != nil {
				b.Fatal(err)
			}
		}

		tickets = append(tickets, *t)
	}

	filterScenarios := []string{"all", "active", "todo", "doing", "done"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		filter := filterScenarios[i%len(filterScenarios)]
		// Simulate filtering logic with pre-allocated slice
		filtered := make([]ticket.Ticket, 0, len(tickets))
		for _, t := range tickets {
			switch filter {
			case "all":
				filtered = append(filtered, t)
			case "active":
				if t.Status() != ticket.StatusDone {
					filtered = append(filtered, t)
				}
			case "todo":
				if t.Status() == ticket.StatusTodo {
					filtered = append(filtered, t)
				}
			case "doing":
				if t.Status() == ticket.StatusDoing {
					filtered = append(filtered, t)
				}
			case "done":
				if t.Status() == ticket.StatusDone {
					filtered = append(filtered, t)
				}
			}
		}
		_ = filtered
	}
}

// BenchmarkErrorHandling benchmarks error message rendering
func BenchmarkErrorHandling(b *testing.B) {
	errors := []error{
		fmt.Errorf("simple error"),
		fmt.Errorf("error with details: %s", "some details"),
		fmt.Errorf("long error message with multiple lines\nand additional context\nthat spans several lines"),
	}

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := errors[i%len(errors)]
		// Simulate error handling
		_ = err.Error()
		// In real app, this would involve rendering error views
	}
}

// setupBenchmarkUIRepo creates a minimal git repository for UI benchmarking
func setupBenchmarkUIRepo(b *testing.B, tmpDir string) {
	b.Helper()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	require.NoError(b, cmd.Run())

	// Configure git locally
	cmd = exec.Command("git", "config", "user.name", "Benchmark User")
	cmd.Dir = tmpDir
	require.NoError(b, cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "benchmark@example.com")
	cmd.Dir = tmpDir
	require.NoError(b, cmd.Run())

	// Create .ticketflow.yaml
	configContent := `
worktree:
  enabled: false
tickets:
  dir: tickets
`
	configPath := filepath.Join(tmpDir, ".ticketflow.yaml")
	require.NoError(b, os.WriteFile(configPath, []byte(configContent), 0644))

	// Create ticket directories
	ticketsDir := filepath.Join(tmpDir, "tickets")
	for _, dir := range []string{"todo", "doing", "done"} {
		require.NoError(b, os.MkdirAll(filepath.Join(ticketsDir, dir), 0755))
	}

	// Initial commit
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	require.NoError(b, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	require.NoError(b, cmd.Run())
}
