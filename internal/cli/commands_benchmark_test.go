package cli

import (
	"context"
	"fmt"
	"io"
	"testing"

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
			_, _ = app.Git.Exec(ctx, "add", ".")
			_, _ = app.Git.Exec(ctx, "commit", "-m", "Add benchmark tickets")
			b.StartTimer()

			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				err := app.StartTicket(ctx, ticketIDs[i], false)
				if err != nil {
					b.Fatal(err)
				}

				// For non-worktree mode, commit the changes to avoid uncommitted changes error
				if !scenario.worktreeEnabled {
					_, _ = app.Git.Exec(ctx, "add", ".")
					_, _ = app.Git.Exec(ctx, "commit", "-m", "Benchmark commit")
				}

				// Clean up by switching back to main branch
				_, _ = app.Git.Exec(ctx, "checkout", "main")

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

	// Pre-create and start tickets
	b.StopTimer()
	ticketIDs := testutil.CreateBenchmarkTickets(b, env, b.N, "todo")
	
	for i := 0; i < b.N; i++ {
		err := app.StartTicket(ctx, ticketIDs[i], false)
		if err != nil {
			b.Fatal(err)
		}
		// Commit changes
		_, _ = app.Git.Exec(ctx, "add", ".")
		_, _ = app.Git.Exec(ctx, "commit", "-m", "Start ticket")
	}
	b.StartTimer()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Switch to ticket branch
		_, err := app.Git.Exec(ctx, "checkout", ticketIDs[i])
		if err != nil {
			b.Fatal(err)
		}

		err = app.CloseTicket(ctx, false)
		if err != nil {
			b.Fatal(err)
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

			// Create todo tickets
			testutil.CreateBenchmarkTickets(b, env, todoCount, "todo")

			// Create doing tickets
			testutil.CreateBenchmarkTickets(b, env, doingCount, "doing")

			// Create done tickets
			testutil.CreateBenchmarkTickets(b, env, doneCount, "done")

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