package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/ticket"
	"time"
)

// BenchmarkCreateTicket benchmarks the full create ticket command
func BenchmarkCreateTicket(b *testing.B) {
	tmpDir := b.TempDir()
	setupBenchmarkGitRepo(b, tmpDir)

	app := &App{
		Manager:     ticket.NewManager(config.Default(), tmpDir),
		Git:         git.New(tmpDir),
		Config:      config.Default(),
		ProjectRoot: tmpDir,
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
			tmpDir := b.TempDir()
			setupBenchmarkGitRepo(b, tmpDir)

			cfg := config.Default()
			cfg.Worktree.Enabled = scenario.worktreeEnabled
			if scenario.worktreeEnabled {
				cfg.Worktree.BaseDir = "../.worktrees"
			}

			app := &App{
				Manager:     ticket.NewManager(cfg, tmpDir),
				Git:         git.New(tmpDir),
				Config:      cfg,
				ProjectRoot: tmpDir,
			}

			ctx := context.Background()

			// Pre-create tickets for benchmarking
			b.StopTimer()
			ticketIDs := make([]string, b.N)
			for i := 0; i < b.N; i++ {
				slug := fmt.Sprintf("benchmark-ticket-%d", i)
				err := app.NewTicket(ctx, slug, "", FormatText)
				if err != nil {
					b.Fatal(err)
				}
				ticketIDs[i] = generateTicketID(slug)
			}
			b.StartTimer()

			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				err := app.StartTicket(ctx, ticketIDs[i], false)
				if err != nil {
					b.Fatal(err)
				}

				// Clean up by switching back to main branch
				_, _ = app.Git.Exec(ctx, "checkout", "main")

				// Remove worktree if created
				if scenario.worktreeEnabled {
					worktreePath := filepath.Join(tmpDir, cfg.Worktree.BaseDir, ticketIDs[i])
					_ = app.Git.RemoveWorktree(ctx, worktreePath)
				}
			}
		})
	}
}

// BenchmarkCloseTicket benchmarks the close ticket operation
func BenchmarkCloseTicket(b *testing.B) {
	tmpDir := b.TempDir()
	setupBenchmarkGitRepo(b, tmpDir)

	cfg := config.Default()
	app := &App{
		Manager:     ticket.NewManager(cfg, tmpDir),
		Git:         git.New(tmpDir),
		Config:      cfg,
		ProjectRoot: tmpDir,
	}

	ctx := context.Background()

	// Pre-create and start tickets
	b.StopTimer()
	ticketIDs := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		slug := fmt.Sprintf("benchmark-ticket-%d", i)
		err := app.NewTicket(ctx, slug, "", FormatText)
		if err != nil {
			b.Fatal(err)
		}

		// Get the actual ticket ID
		tickets, err := app.Manager.List(ctx, ticket.StatusFilterTodo)
		if err != nil {
			b.Fatal(err)
		}

		// Find the ticket we just created
		var ticketID string
		for _, t := range tickets {
			if t.Slug == slug {
				ticketID = t.ID
				break
			}
		}
		if ticketID == "" {
			b.Fatalf("Could not find ticket with slug %s", slug)
		}
		ticketIDs[i] = ticketID

		// Start the ticket to move it to doing status
		err = app.StartTicket(ctx, ticketID, false)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StartTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Switch to the ticket branch
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

// BenchmarkListTickets benchmarks listing tickets with different counts
func BenchmarkListTickets(b *testing.B) {
	scenarios := []struct {
		name        string
		ticketCount int
		format      OutputFormat
	}{
		{"10-tickets-text", 10, FormatText},
		{"10-tickets-json", 10, FormatJSON},
		{"50-tickets-text", 50, FormatText},
		{"50-tickets-json", 50, FormatJSON},
		{"100-tickets-text", 100, FormatText},
		{"100-tickets-json", 100, FormatJSON},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			tmpDir := b.TempDir()
			setupBenchmarkGitRepo(b, tmpDir)

			app := &App{
				Manager:     ticket.NewManager(config.Default(), tmpDir),
				Git:         git.New(tmpDir),
				Config:      config.Default(),
				ProjectRoot: tmpDir,
			}

			ctx := context.Background()

			// Create tickets
			for i := 0; i < scenario.ticketCount; i++ {
				slug := fmt.Sprintf("benchmark-ticket-%d", i)
				err := app.NewTicket(ctx, slug, "", FormatText)
				if err != nil {
					b.Fatal(err)
				}
			}

			// Redirect output to discard it during benchmark
			oldStdout := os.Stdout
			devNull, err := os.Open(os.DevNull)
			if err != nil {
				b.Fatal(err)
			}
			defer devNull.Close()
			os.Stdout = devNull
			defer func() { os.Stdout = oldStdout }()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				err := app.ListTickets(ctx, "all", 0, scenario.format)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// generateTicketID generates a ticket ID from a slug
func generateTicketID(slug string) string {
	return generateTimestamp() + "-" + slug
}

// generateTimestamp generates a timestamp for ticket IDs
func generateTimestamp() string {
	return time.Now().Format("060102-150405")
}

// setupBenchmarkGitRepo creates a minimal git repository for benchmarking
func setupBenchmarkGitRepo(b *testing.B, tmpDir string) {
	b.Helper()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	require.NoError(b, cmd.Run())

	// Configure git locally (not globally)
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
