package cli

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/testutil"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// BenchmarkNewTicketWithVariousSizes benchmarks ticket creation with different content sizes
func BenchmarkNewTicketWithVariousSizes(b *testing.B) {
	scenarios := []struct {
		name        string
		contentSize int
	}{
		{"small-100B", 100},
		{"medium-1KB", 1024},
		{"large-10KB", 10 * 1024},
		{"xlarge-100KB", 100 * 1024},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			env := testutil.SetupBenchmarkEnvironment(b)

			app := &App{
				Manager:     env.Manager,
				Git:         env.Git,
				Config:      env.Config,
				ProjectRoot: env.ProjectRoot,
				Output:      NewOutputWriter(io.Discard, io.Discard, FormatText),
			}

			ctx := context.Background()
			content := testutil.GenerateTicketContent(scenario.contentSize)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				slug := fmt.Sprintf("bench-ticket-%d", i)

				// Create ticket
				err := app.NewTicket(ctx, slug, "", FormatText)
				require.NoError(b, err)

				// Write content to simulate realistic usage
				ticketID := generateTicketID(slug)
				err = env.Manager.WriteContent(ctx, ticketID, content)
				require.NoError(b, err)
			}
		})
	}
}

// BenchmarkListTicketsLargeRepository benchmarks listing with realistic large repositories
func BenchmarkListTicketsLargeRepository(b *testing.B) {
	scenarios := []struct {
		name         string
		totalTickets int
		filter       string
		format       OutputFormat
	}{
		{"100-all-text", 100, "all", FormatText},
		{"100-all-json", 100, "all", FormatJSON},
		{"500-all-text", 500, "all", FormatText},
		{"500-all-json", 500, "all", FormatJSON},
		{"1000-all-text", 1000, "all", FormatText},
		{"1000-all-json", 1000, "all", FormatJSON},
		{"1000-todo-text", 1000, "todo", FormatText},
		{"1000-doing-text", 1000, "doing", FormatText},
		{"1000-done-text", 1000, "done", FormatText},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			env := testutil.SetupBenchmarkEnvironment(b)

			// Create large repository
			b.StopTimer()
			testutil.CreateLargeRepository(b, env, scenario.totalTickets)
			b.StartTimer()

			app := &App{
				Manager:     env.Manager,
				Git:         env.Git,
				Config:      env.Config,
				ProjectRoot: env.ProjectRoot,
				Output:      NewOutputWriter(io.Discard, io.Discard, scenario.format),
			}

			ctx := context.Background()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				err := app.ListTickets(ctx, ticket.Status(scenario.filter), 0, scenario.format)
				require.NoError(b, err)
			}

			// Report memory usage for large datasets
			if scenario.totalTickets >= 1000 {
				testutil.MeasureMemoryUsage(b, scenario.name)
			}
		})
	}
}

// BenchmarkStartTicketConcurrent benchmarks concurrent ticket start operations
func BenchmarkStartTicketConcurrent(b *testing.B) {
	concurrencyLevels := []int{1, 2, 4, 8}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("concurrency-%d", concurrency), func(b *testing.B) {
			env := testutil.SetupBenchmarkEnvironment(b)

			// Disable worktrees for concurrent testing
			env.Config.Worktree.Enabled = false

			app := &App{
				Manager:     env.Manager,
				Git:         env.Git,
				Config:      env.Config,
				ProjectRoot: env.ProjectRoot,
				Output:      NewOutputWriter(io.Discard, io.Discard, FormatText),
			}

			ctx := context.Background()

			// Pre-create tickets
			b.StopTimer()
			ticketIDs := testutil.CreateBenchmarkTickets(b, env, b.N, "todo")
			b.StartTimer()

			b.ResetTimer()
			b.ReportAllocs()

			// Use semaphore to limit concurrency
			sem := make(chan struct{}, concurrency)
			var wg sync.WaitGroup

			for i := 0; i < b.N; i++ {
				wg.Add(1)
				sem <- struct{}{} // Acquire semaphore

				go func(idx int) {
					defer wg.Done()
					defer func() { <-sem }() // Release semaphore

					err := app.StartTicket(ctx, ticketIDs[idx], false)
					if err != nil {
						b.Error(err)
					}

					// Clean up
					_, _ = app.Git.Exec(ctx, "add", ".")
					_, _ = app.Git.Exec(ctx, "commit", "-m", fmt.Sprintf("Start ticket %d", idx))
					_, _ = app.Git.Exec(ctx, "checkout", "main")
				}(i)
			}

			wg.Wait()
		})
	}
}

// BenchmarkCloseTicketWithReason benchmarks close operations with different reason sizes
func BenchmarkCloseTicketWithReason(b *testing.B) {
	reasonSizes := []struct {
		name string
		size int
	}{
		{"no-reason", 0},
		{"short-reason", 50},
		{"medium-reason", 200},
		{"long-reason", 1000},
	}

	for _, rs := range reasonSizes {
		b.Run(rs.name, func(b *testing.B) {
			env := testutil.SetupBenchmarkEnvironment(b)

			// Disable worktrees for close operations
			env.Config.Worktree.Enabled = false

			app := &App{
				Manager:     env.Manager,
				Git:         env.Git,
				Config:      env.Config,
				ProjectRoot: env.ProjectRoot,
				Output:      NewOutputWriter(io.Discard, io.Discard, FormatText),
			}

			ctx := context.Background()

			// Pre-create and start tickets
			b.StopTimer()
			ticketIDs := make([]string, b.N)
			for i := 0; i < b.N; i++ {
				slug := fmt.Sprintf("bench-ticket-%d", i)
				t, err := env.Manager.Create(ctx, slug)
				require.NoError(b, err)

				ticketIDs[i] = t.ID

				// Start the ticket
				err = app.StartTicket(ctx, t.ID, false)
				require.NoError(b, err)

				// Commit changes
				_, _ = app.Git.Exec(ctx, "add", ".")
				_, _ = app.Git.Exec(ctx, "commit", "-m", "Start ticket")
			}

			// Generate reason if needed
			var reason string
			if rs.size > 0 {
				reason = testutil.GenerateTicketContent(rs.size)
			}

			b.StartTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Switch to ticket branch
				_, err := app.Git.Exec(ctx, "checkout", ticketIDs[i])
				require.NoError(b, err)

				// Close with reason
				if reason != "" {
					err = app.CloseTicketWithReason(ctx, reason, false)
				} else {
					err = app.CloseTicket(ctx, false)
				}
				require.NoError(b, err)
			}
		})
	}
}

// BenchmarkWorktreeOperations benchmarks worktree-specific operations
func BenchmarkWorktreeOperations(b *testing.B) {
	operations := []struct {
		name      string
		operation func(*App, context.Context, string) error
	}{
		{
			"create-worktree",
			func(app *App, ctx context.Context, ticketID string) error {
				return app.StartTicket(ctx, ticketID, false)
			},
		},
		{
			"sync-worktree",
			func(app *App, ctx context.Context, ticketID string) error {
				// Simulate sync by listing worktrees
				_, err := app.Git.ListWorktrees(ctx)
				return err
			},
		},
	}

	for _, op := range operations {
		b.Run(op.name, func(b *testing.B) {
			env := testutil.SetupBenchmarkEnvironment(b)

			// Enable worktrees
			env.Config.Worktree.Enabled = true
			env.Config.Worktree.BaseDir = "../.worktrees"
			env.Config.Worktree.InitCommands = []string{} // No init commands for benchmarks

			app := &App{
				Manager:     env.Manager,
				Git:         env.Git,
				Config:      env.Config,
				ProjectRoot: env.ProjectRoot,
				Output:      NewOutputWriter(io.Discard, io.Discard, FormatText),
			}

			ctx := context.Background()

			// Pre-create tickets
			b.StopTimer()
			ticketIDs := testutil.CreateBenchmarkTickets(b, env, b.N, "todo")
			b.StartTimer()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				err := op.operation(app, ctx, ticketIDs[i])
				if err != nil {
					b.Error(err)
				}

				// Cleanup worktree if created
				if op.name == "create-worktree" {
					worktreePath := filepath.Join(env.ProjectRoot, env.Config.Worktree.BaseDir, ticketIDs[i])
					_ = app.Git.RemoveWorktree(ctx, worktreePath)
					_, _ = app.Git.Exec(ctx, "checkout", "main")
				}
			}
		})
	}
}

// BenchmarkSearchAndFilter benchmarks ticket search and filtering operations
func BenchmarkSearchAndFilter(b *testing.B) {
	scenarios := []struct {
		name         string
		ticketCount  int
		searchTerm   string
		expectedHits int
	}{
		{"100-tickets-10pct-match", 100, "important", 10},
		{"500-tickets-10pct-match", 500, "important", 50},
		{"1000-tickets-10pct-match", 1000, "important", 100},
		{"1000-tickets-1pct-match", 1000, "critical", 10},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			env := testutil.SetupBenchmarkEnvironment(b)

			// Create tickets with searchable content
			b.StopTimer()
			ctx := context.Background()
			for i := 0; i < scenario.ticketCount; i++ {
				slug := fmt.Sprintf("ticket-%d", i)
				t, err := env.Manager.Create(ctx, slug)
				require.NoError(b, err)

				// Add searchable content to some tickets
				content := fmt.Sprintf("Ticket content %d", i)
				if i < scenario.expectedHits {
					content += " " + scenario.searchTerm
				}

				err = env.Manager.WriteContent(ctx, t.ID, content)
				require.NoError(b, err)
			}
			b.StartTimer()

			app := &App{
				Manager:     env.Manager,
				Git:         env.Git,
				Config:      env.Config,
				ProjectRoot: env.ProjectRoot,
				Output:      NewOutputWriter(io.Discard, io.Discard, FormatText),
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Simulate search by listing and filtering
				tickets, err := app.Manager.List(ctx, ticket.StatusFilterAll)
				require.NoError(b, err)

				// Filter tickets containing search term
				var matched []ticket.Ticket
				for _, t := range tickets {
					content, err := app.Manager.ReadContent(ctx, t.ID)
					if err == nil && strings.Contains(content, scenario.searchTerm) {
						matched = append(matched, t)
					}
				}

				if len(matched) != scenario.expectedHits {
					b.Errorf("Expected %d matches, got %d", scenario.expectedHits, len(matched))
				}
			}
		})
	}
}

// BenchmarkMemoryPressure benchmarks operations under memory pressure
func BenchmarkMemoryPressure(b *testing.B) {
	scenarios := []struct {
		name            string
		ticketCount     int
		simultaneousOps int
		contentSize     int
	}{
		{"low-pressure", 100, 10, 1024},
		{"medium-pressure", 500, 50, 10 * 1024},
		{"high-pressure", 1000, 100, 100 * 1024},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			env := testutil.SetupBenchmarkEnvironment(b)

			app := &App{
				Manager:     env.Manager,
				Git:         env.Git,
				Config:      env.Config,
				ProjectRoot: env.ProjectRoot,
				Output:      NewOutputWriter(io.Discard, io.Discard, FormatText),
			}

			ctx := context.Background()

			// Create initial tickets with content
			b.StopTimer()
			ticketIDs := make([]string, scenario.ticketCount)
			content := testutil.GenerateTicketContent(scenario.contentSize)

			for i := 0; i < scenario.ticketCount; i++ {
				slug := fmt.Sprintf("ticket-%d", i)
				t, err := env.Manager.Create(ctx, slug)
				require.NoError(b, err)

				err = env.Manager.WriteContent(ctx, t.ID, content)
				require.NoError(b, err)

				ticketIDs[i] = t.ID
			}
			b.StartTimer()

			// Measure initial memory
			testutil.MeasureMemoryUsage(b, "before")

			b.ResetTimer()
			b.ReportAllocs()

			// Perform simultaneous operations
			var wg sync.WaitGroup
			sem := make(chan struct{}, scenario.simultaneousOps)

			for i := 0; i < b.N; i++ {
				wg.Add(1)
				sem <- struct{}{}

				go func(idx int) {
					defer wg.Done()
					defer func() { <-sem }()

					// Perform mixed operations
					ticketIdx := idx % len(ticketIDs)

					// Read content
					_, _ = env.Manager.ReadContent(ctx, ticketIDs[ticketIdx])

					// List tickets
					_, _ = app.Manager.List(ctx, ticket.StatusFilterAll)

					// Get specific ticket
					_, _ = app.Manager.Get(ctx, ticketIDs[ticketIdx])
				}(i)
			}

			wg.Wait()

			// Measure final memory
			testutil.MeasureMemoryUsage(b, "after")
		})
	}
}
