package testutil

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// BenchmarkEnvironment represents a complete benchmark test environment
type BenchmarkEnvironment struct {
	TmpDir      string
	Config      *config.Config
	Manager     *ticket.Manager
	Git         *git.Git
	ProjectRoot string
	TicketsDir  string
}

// SetupBenchmarkEnvironment creates a complete benchmark environment with git repo and ticket structure
func SetupBenchmarkEnvironment(b *testing.B) *BenchmarkEnvironment {
	b.Helper()

	tmpDir := b.TempDir()

	// Initialize git repo
	SetupBenchmarkGitRepo(b, tmpDir)

	// Create config
	cfg := config.Default()
	cfg.Tickets.Dir = "tickets"
	cfg.Worktree.Enabled = false // Default to disabled for benchmarks

	// Create ticket directories
	ticketsDir := filepath.Join(tmpDir, cfg.Tickets.Dir)
	for _, dir := range []string{"todo", "doing", "done"} {
		require.NoError(b, os.MkdirAll(filepath.Join(ticketsDir, dir), 0755))
	}

	// Create .ticketflow.yaml
	configContent := `
worktree:
  enabled: false
tickets:
  dir: tickets
`
	configPath := filepath.Join(tmpDir, ".ticketflow.yaml")
	require.NoError(b, os.WriteFile(configPath, []byte(configContent), 0644))

	// Commit the structure
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	require.NoError(b, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Setup benchmark environment")
	cmd.Dir = tmpDir
	require.NoError(b, cmd.Run())

	return &BenchmarkEnvironment{
		TmpDir:      tmpDir,
		Config:      cfg,
		Manager:     ticket.NewManager(cfg, tmpDir),
		Git:         git.New(tmpDir),
		ProjectRoot: tmpDir,
		TicketsDir:  ticketsDir,
	}
}

// SetupBenchmarkGitRepo creates a minimal git repository for benchmarking
func SetupBenchmarkGitRepo(b *testing.B, tmpDir string) {
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

	// Create initial file
	readmePath := filepath.Join(tmpDir, "README.md")
	require.NoError(b, os.WriteFile(readmePath, []byte("# Benchmark Repo"), 0644))

	// Initial commit
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	require.NoError(b, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	require.NoError(b, cmd.Run())
}

// CreateBenchmarkTickets creates a specified number of tickets for benchmarking
func CreateBenchmarkTickets(b *testing.B, env *BenchmarkEnvironment, count int, status string) []string {
	b.Helper()
	return CreateBenchmarkTicketsWithPrefix(b, env, count, status, "bench-ticket")
}

// CreateBenchmarkTicketsWithPrefix creates tickets with a custom prefix to avoid ID collisions
func CreateBenchmarkTicketsWithPrefix(b *testing.B, env *BenchmarkEnvironment, count int, status string, prefix string) []string {
	b.Helper()

	// Use context with timeout for long-running operations
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ticketIDs := make([]string, count)

	for i := 0; i < count; i++ {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			b.Fatalf("Ticket creation phase timed out after creating %d of %d tickets with prefix '%s' in status '%s': %v", i, count, prefix, status, ctx.Err())
		default:
		}

		// Use prefix and add nanosecond precision to ensure uniqueness
		slug := fmt.Sprintf("%s-%d-%d", prefix, i, time.Now().UnixNano())
		t, err := env.Manager.Create(ctx, slug)
		require.NoError(b, err, "failed to create ticket with slug: %s", slug)
		ticketIDs[i] = t.ID

		// Move to specified status if not "todo"
		if status != "todo" {
			oldPath := filepath.Join(env.TicketsDir, "todo", t.ID+".md")
			newPath := filepath.Join(env.TicketsDir, status, t.ID+".md")
			require.NoError(b, os.Rename(oldPath, newPath))
		}

		// Small sleep to ensure different timestamps for ticket IDs
		if i < count-1 {
			time.Sleep(time.Millisecond)
		}
	}

	return ticketIDs
}

// GenerateTicketContent generates ticket content of specified size
func GenerateTicketContent(size int) string {
	const chunk = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. "
	chunkLen := len(chunk)

	if size <= chunkLen {
		return chunk[:size]
	}

	// Use strings.Builder for efficient string concatenation
	var builder strings.Builder
	builder.Grow(size) // Pre-allocate capacity

	for builder.Len() < size {
		remaining := size - builder.Len()
		if remaining >= chunkLen {
			builder.WriteString(chunk)
		} else {
			builder.WriteString(chunk[:remaining])
		}
	}

	return builder.String()
}

// BenchmarkTimer provides utilities for controlling benchmark timing
type BenchmarkTimer struct {
	b           *testing.B
	startTime   time.Time
	stopped     bool
	elapsedTime time.Duration // Store elapsed time when stopped
}

// NewBenchmarkTimer creates a new benchmark timer
func NewBenchmarkTimer(b *testing.B) *BenchmarkTimer {
	return &BenchmarkTimer{
		b:           b,
		startTime:   time.Now(),
		stopped:     false,
		elapsedTime: 0,
	}
}

// Stop stops the timer for setup operations
func (bt *BenchmarkTimer) Stop() {
	if !bt.stopped {
		bt.b.StopTimer()
		bt.elapsedTime = time.Since(bt.startTime) // Store elapsed time when stopping
		bt.stopped = true
	}
}

// Start restarts the timer for measured operations
func (bt *BenchmarkTimer) Start() {
	if bt.stopped {
		bt.b.StartTimer()
		bt.stopped = false
		bt.startTime = time.Now()
		// Don't reset elapsedTime here - keep accumulated time for tracking total elapsed
	}
}

// Elapsed returns the elapsed time since the timer was started.
// When the timer is stopped, it returns the elapsed time that was captured at the moment Stop() was called.
// When the timer is running, it returns the current elapsed time since Start() was last called.
// This allows accurate measurement of specific code sections without including setup/cleanup time.
func (bt *BenchmarkTimer) Elapsed() time.Duration {
	if bt.stopped {
		return bt.elapsedTime // Return stored elapsed time when stopped
	}
	return time.Since(bt.startTime)
}

// TimeOp times a single operation within a benchmark
func (bt *BenchmarkTimer) TimeOp(name string, op func() error) error {
	start := time.Now()
	err := op()
	elapsed := time.Since(start)

	// Log operation time for analysis (will be captured in benchmark output)
	if elapsed > 100*time.Millisecond {
		bt.b.Logf("%s took %v", name, elapsed)
	}

	return err
}

// RunBenchmarkScenario runs a benchmark with multiple scenarios
func RunBenchmarkScenario(b *testing.B, scenarios []BenchmarkScenario) {
	for _, scenario := range scenarios {
		b.Run(scenario.Name, func(b *testing.B) {
			env := SetupBenchmarkEnvironment(b)
			scenario.Setup(b, env)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				if err := scenario.Run(env, i); err != nil {
					b.Fatal(err)
				}
			}

			if scenario.Cleanup != nil {
				scenario.Cleanup(b, env)
			}
		})
	}
}

// BenchmarkScenario represents a benchmark scenario configuration
type BenchmarkScenario struct {
	Name    string
	Setup   func(*testing.B, *BenchmarkEnvironment)
	Run     func(*BenchmarkEnvironment, int) error
	Cleanup func(*testing.B, *BenchmarkEnvironment)
}

// MeasureMemoryUsage captures memory statistics during benchmark
func MeasureMemoryUsage(b *testing.B, name string) {
	b.Helper()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	b.Logf("%s - Memory: Alloc=%v MB, TotalAlloc=%v MB, Sys=%v MB, NumGC=%v",
		name,
		bToMB(m.Alloc),
		bToMB(m.TotalAlloc),
		bToMB(m.Sys),
		m.NumGC,
	)
}

// bToMB converts bytes to megabytes
func bToMB(b uint64) float64 {
	return float64(b) / 1024 / 1024
}

// CreateLargeRepository creates a repository with many tickets for stress testing
func CreateLargeRepository(b *testing.B, env *BenchmarkEnvironment, totalTickets int) {
	b.Helper()

	timer := NewBenchmarkTimer(b)
	timer.Stop()

	// Calculate dynamic timeout based on ticket count
	// 1 second per ticket, minimum 2 minutes, maximum 30 minutes
	timeoutSeconds := totalTickets
	if timeoutSeconds < 120 {
		timeoutSeconds = 120 // Minimum 2 minutes
	} else if timeoutSeconds > 1800 {
		timeoutSeconds = 1800 // Maximum 30 minutes
	}

	// Use context with timeout for long-running operations
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	b.Logf("Using timeout of %v for %d tickets", time.Duration(timeoutSeconds)*time.Second, totalTickets)

	// Distribute tickets across statuses
	todoCount := totalTickets * 40 / 100
	doingCount := totalTickets * 20 / 100
	doneCount := totalTickets - todoCount - doingCount

	b.Logf("Creating large repository: %d todo, %d doing, %d done", todoCount, doingCount, doneCount)

	// Create todo tickets
	for i := 0; i < todoCount; i++ {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			b.Fatalf("Repository creation timed out while creating TODO tickets (created %d of %d): %v", i, todoCount, ctx.Err())
		default:
		}

		slug := fmt.Sprintf("todo-ticket-%d", i)
		_, err := env.Manager.Create(ctx, slug)
		require.NoError(b, err, "failed to create ticket with slug: %s", slug)
	}

	// Create doing tickets
	for i := 0; i < doingCount; i++ {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			b.Fatalf("Repository creation timed out while creating DOING tickets (created %d of %d): %v", i, doingCount, ctx.Err())
		default:
		}

		slug := fmt.Sprintf("doing-ticket-%d", i)
		t, err := env.Manager.Create(ctx, slug)
		require.NoError(b, err, "failed to create ticket with slug: %s", slug)

		// Move to doing
		oldPath := filepath.Join(env.TicketsDir, "todo", t.ID+".md")
		newPath := filepath.Join(env.TicketsDir, "doing", t.ID+".md")
		require.NoError(b, os.Rename(oldPath, newPath))
	}

	// Create done tickets
	for i := 0; i < doneCount; i++ {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			b.Fatalf("Repository creation timed out while creating DONE tickets (created %d of %d): %v", i, doneCount, ctx.Err())
		default:
		}

		slug := fmt.Sprintf("done-ticket-%d", i)
		t, err := env.Manager.Create(ctx, slug)
		require.NoError(b, err, "failed to create ticket with slug: %s", slug)

		// Move to done
		oldPath := filepath.Join(env.TicketsDir, "todo", t.ID+".md")
		newPath := filepath.Join(env.TicketsDir, "done", t.ID+".md")
		require.NoError(b, os.Rename(oldPath, newPath))
	}

	timer.Start()
}
