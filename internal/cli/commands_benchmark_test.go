package cli

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/testutil"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// BenchmarkCreateTicket benchmarks the full create ticket command
func BenchmarkCreateTicket(b *testing.B) {
	env := testutil.SetupBenchmarkEnvironment(b)

	app := &App{
		Manager:     env.Manager,
		Git:         env.Git,
		Config:      env.Config,
		ProjectRoot: env.ProjectRoot,
		Output:      NewOutputWriter(io.Discard, io.Discard, FormatText),
	}

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		slug := fmt.Sprintf("benchmark-ticket-%d", i)
		err := app.NewTicket(ctx, slug, "", FormatText)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStartTicket benchmarks the start ticket operation
func BenchmarkStartTicket(b *testing.B) {
	scenarios := []struct {
		name            string
		worktreeEnabled bool
	}{
		{"with-worktree", true},
		{"without-worktree", false},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			env := testutil.SetupBenchmarkEnvironment(b)

			env.Config.Worktree.Enabled = scenario.worktreeEnabled
			if scenario.worktreeEnabled {
				env.Config.Worktree.BaseDir = "../.worktrees"
				// Disable init commands for benchmark
				env.Config.Worktree.InitCommands = []string{}
			}

			app := &App{
				Manager:     env.Manager,
				Git:         env.Git,
				Config:      env.Config,
				ProjectRoot: env.ProjectRoot,
				Output:      NewOutputWriter(io.Discard, io.Discard, FormatText),
			}

			ctx := context.Background()

			// Pre-create tickets for benchmarking
			b.StopTimer()
			ticketIDs := testutil.CreateBenchmarkTickets(b, env, b.N, "todo")

			// Commit the created tickets to avoid uncommitted changes
			if _, err := app.Git.Exec(ctx, "add", "."); err != nil {
				b.Fatalf("Failed to add tickets: %v", err)
			}
			if _, err := app.Git.Exec(ctx, "commit", "-m", "Add benchmark tickets"); err != nil {
				b.Fatalf("Failed to commit tickets: %v", err)
			}
			b.StartTimer()

			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				err := app.StartTicket(ctx, ticketIDs[i], false)
				if err != nil {
					b.Fatal(err)
				}

				// For non-worktree mode, commit the changes to avoid uncommitted changes error
				if !scenario.worktreeEnabled {
					if _, err := app.Git.Exec(ctx, "add", "."); err != nil {
						b.Logf("Warning: Failed to add changes: %v", err)
					}
					if _, err := app.Git.Exec(ctx, "commit", "-m", "Benchmark commit"); err != nil {
						b.Logf("Warning: Failed to commit changes: %v", err)
					}
				}

				// Clean up by switching back to main branch
				if _, err := app.Git.Exec(ctx, "checkout", "main"); err != nil {
					b.Logf("Warning: Failed to checkout main: %v", err)
				}

				// Remove worktree if created
				if scenario.worktreeEnabled {
					worktreePath := fmt.Sprintf("%s/../.worktrees/%s", env.ProjectRoot, ticketIDs[i])
					_ = app.Git.RemoveWorktree(ctx, worktreePath)
				}
			}
		})
	}
}

// BenchmarkCloseTicket benchmarks the close ticket operation
func BenchmarkCloseTicket(b *testing.B) {
	// Setup once for all iterations
	env := testutil.SetupBenchmarkEnvironment(b)

	// Disable worktrees for close benchmark to avoid conflicts
	env.Config.Worktree.Enabled = false

	app := &App{
		Manager:     env.Manager,
		Git:         env.Git,
		Config:      env.Config,
		ProjectRoot: env.ProjectRoot,
		Output:      NewOutputWriter(io.Discard, io.Discard, FormatText),
	}

	ctx := context.Background()

	// Pre-create enough tickets for all benchmark iterations
	// Use a reasonable number since b.N might be large
	numTickets := b.N
	if numTickets > 100 {
		numTickets = 100
	}
	
	b.StopTimer()
	
	// Create and prepare tickets only once
	ticketIDs := testutil.CreateBenchmarkTicketsWithPrefix(b, env, numTickets, "todo", fmt.Sprintf("close-bench-%d", time.Now().Unix()))
	
	// Commit the created tickets - check for errors
	if _, err := app.Git.Exec(ctx, "add", "."); err != nil {
		b.Fatalf("Failed to add tickets: %v", err)
	}
	if _, err := app.Git.Exec(ctx, "commit", "-m", "Add benchmark tickets"); err != nil {
		b.Fatalf("Failed to commit tickets: %v", err)
	}

	// Start all tickets
	for i := 0; i < numTickets; i++ {
		err := app.StartTicket(ctx, ticketIDs[i], false)
		if err != nil {
			b.Fatalf("Failed to start ticket %s: %v", ticketIDs[i], err)
		}
		
		// Commit the changes on the feature branch (ticket moved to doing)
		if _, err := app.Git.Exec(ctx, "add", "."); err != nil {
			b.Fatalf("Failed to add changes for ticket %s: %v", ticketIDs[i], err)
		}
		if _, err := app.Git.Exec(ctx, "commit", "-m", fmt.Sprintf("Start ticket %s", ticketIDs[i])); err != nil {
			b.Fatalf("Failed to commit changes for ticket %s: %v", ticketIDs[i], err)
		}
		
		// Switch back to main for next ticket
		if _, err := app.Git.Exec(ctx, "checkout", "main"); err != nil {
			b.Fatalf("Failed to checkout main: %v", err)
		}
	}
	
	b.StartTimer()
	b.ResetTimer()
	b.ReportAllocs()

	// Benchmark only the close operation
	for i := 0; i < b.N; i++ {
		ticketIdx := i % numTickets
		
		// Switch to ticket branch
		_, err := app.Git.Exec(ctx, "checkout", ticketIDs[ticketIdx])
		if err != nil {
			b.Fatal(err)
		}

		// Use force=true to skip uncommitted changes check (current-ticket.md symlink)
		err = app.CloseTicket(ctx, true)
		if err != nil {
			// Skip if already closed (when b.N > numTickets)
			continue
		}
	}
}

// BenchmarkListTickets benchmarks listing tickets with different filters
func BenchmarkListTickets(b *testing.B) {
	scenarios := []struct {
		name        string
		filter      ticket.Status
		ticketCount int
		format      OutputFormat
	}{
		{"10-tickets-all-text", ticket.Status("all"), 10, FormatText},
		{"10-tickets-all-json", ticket.Status("all"), 10, FormatJSON},
		{"100-tickets-all-text", ticket.Status("all"), 100, FormatText},
		{"100-tickets-all-json", ticket.Status("all"), 100, FormatJSON},
		{"100-tickets-todo", ticket.StatusTodo, 100, FormatText},
		{"100-tickets-doing", ticket.StatusDoing, 100, FormatText},
		{"100-tickets-done", ticket.StatusDone, 100, FormatText},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			env := testutil.SetupBenchmarkEnvironment(b)

			app := &App{
				Manager:     env.Manager,
				Git:         env.Git,
				Config:      env.Config,
				ProjectRoot: env.ProjectRoot,
				Output:      NewOutputWriter(io.Discard, io.Discard, scenario.format),
			}

			ctx := context.Background()

			// Pre-create tickets with different statuses
			b.StopTimer()
			todoCount := scenario.ticketCount * 40 / 100
			doingCount := scenario.ticketCount * 30 / 100
			doneCount := scenario.ticketCount - todoCount - doingCount

			// Create todo tickets with unique prefix per scenario
			testutil.CreateBenchmarkTicketsWithPrefix(b, env, todoCount, "todo", fmt.Sprintf("%s-todo", scenario.name))

			// Create doing tickets with unique prefix
			testutil.CreateBenchmarkTicketsWithPrefix(b, env, doingCount, "doing", fmt.Sprintf("%s-doing", scenario.name))

			// Create done tickets with unique prefix
			testutil.CreateBenchmarkTicketsWithPrefix(b, env, doneCount, "done", fmt.Sprintf("%s-done", scenario.name))

			b.StartTimer()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				err := app.ListTickets(ctx, scenario.filter, 0, scenario.format)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkSearchTickets benchmarks searching tickets by content
func BenchmarkSearchTickets(b *testing.B) {
	scenarios := []struct {
		name        string
		ticketCount int
		searchTerm  string
	}{
		{"10-tickets", 10, "benchmark"},
		{"50-tickets", 50, "benchmark"},
		{"100-tickets", 100, "benchmark"},
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

			// Pre-create tickets with searchable content
			b.StopTimer()
			ticketIDs := testutil.CreateBenchmarkTickets(b, env, scenario.ticketCount, "todo")

			// Add content to tickets
			for _, id := range ticketIDs {
				content := fmt.Sprintf("This is a benchmark ticket with search term: %s", scenario.searchTerm)
				err := env.Manager.WriteContent(ctx, id, content)
				require.NoError(b, err)
			}
			b.StartTimer()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				err := app.ListTickets(ctx, "all", 0, FormatText)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
