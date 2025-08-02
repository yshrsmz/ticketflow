package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/ticket"
	"github.com/yshrsmz/ticketflow/internal/ui"
)

// Build variables set by ldflags
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Set up graceful shutdown handling:
	// - Catches SIGINT (Ctrl+C) and SIGTERM signals
	// - Context cancellation propagates through all operations
	// - Git operations using exec.CommandContext will be terminated
	// - Returns exit code 130 (standard for SIGINT) on interruption
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// No arguments = TUI mode
	if len(os.Args) == 1 {
		runTUI()
		return
	}

	// CLI mode
	if err := runCLI(ctx); err != nil {
		// Check if the error is due to context cancellation
		if ctx.Err() != nil {
			fmt.Fprintf(os.Stderr, "\nOperation cancelled\n")
			os.Exit(130) // Standard exit code for SIGINT
		}
		cli.HandleError(err)
		os.Exit(1)
	}
}

func runTUI() {
	// Find git root using default timeout
	g := git.New(".")
	root, err := g.RootPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: not in a git repository\n")
		os.Exit(1)
	}

	// Load config
	cfg, err := config.Load(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load config: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run 'ticketflow init' to initialize the ticket system\n")
		os.Exit(1)
	}

	// Recreate git client with configured timeout
	g = git.NewWithTimeout(".", cfg.GetGitTimeout())

	// Create ticket manager
	manager := ticket.NewManager(cfg, root)

	// Create and run TUI
	model := ui.New(cfg, manager, g, root)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to run TUI: %v\n", err)
		os.Exit(1)
	}
}

func runCLI(ctx context.Context) error {
	// Define subcommands
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)

	newCmd := flag.NewFlagSet("new", flag.ExitOnError)

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listStatus := listCmd.String("status", "", "Filter by status (todo|doing|done)")
	listCount := listCmd.Int("count", 20, "Maximum number of tickets to show")
	listFormat := listCmd.String("format", "text", "Output format (text|json)")

	showCmd := flag.NewFlagSet("show", flag.ExitOnError)
	showFormat := showCmd.String("format", "text", "Output format (text|json)")

	startCmd := flag.NewFlagSet("start", flag.ExitOnError)

	closeCmd := flag.NewFlagSet("close", flag.ExitOnError)
	closeForce := closeCmd.Bool("force", false, "Force close with uncommitted changes")
	closeForceShort := closeCmd.Bool("f", false, "Force close (short form)")

	restoreCmd := flag.NewFlagSet("restore", flag.ExitOnError)

	statusCmd := flag.NewFlagSet("status", flag.ExitOnError)
	statusFormat := statusCmd.String("format", "text", "Output format (text|json)")

	worktreeCmd := flag.NewFlagSet("worktree", flag.ExitOnError)

	cleanupCmd := flag.NewFlagSet("cleanup", flag.ExitOnError)
	cleanupDryRun := cleanupCmd.Bool("dry-run", false, "Show what would be cleaned without making changes")
	cleanupForce := cleanupCmd.Bool("force", false, "Skip confirmation prompts")

	migrateCmd := flag.NewFlagSet("migrate", flag.ExitOnError)
	migrateDryRun := migrateCmd.Bool("dry-run", false, "Show what would be updated without making changes")

	// Parse command
	if len(os.Args) < 2 {
		printUsage()
		return nil
	}

	switch os.Args[1] {
	case "init":
		if err := initCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
		return handleInit(ctx)

	case "new":
		if err := newCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
		if newCmd.NArg() < 1 {
			return fmt.Errorf("missing slug argument")
		}
		return handleNew(ctx, newCmd.Arg(0), *listFormat)

	case "list":
		if err := listCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
		return handleList(ctx, *listStatus, *listCount, *listFormat)

	case "show":
		if err := showCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
		if showCmd.NArg() < 1 {
			return fmt.Errorf("missing ticket argument")
		}
		return handleShow(ctx, showCmd.Arg(0), *showFormat)

	case "start":
		if err := startCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
		if startCmd.NArg() < 1 {
			return fmt.Errorf("missing ticket argument")
		}
		return handleStart(ctx, startCmd.Arg(0), false)

	case "close":
		if err := closeCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
		force := *closeForce || *closeForceShort
		return handleClose(ctx, false, force)

	case "restore":
		if err := restoreCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
		return handleRestore(ctx)

	case "status":
		if err := statusCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
		return handleStatus(ctx, *statusFormat)

	case "worktree":
		if len(os.Args) < 3 {
			printWorktreeUsage()
			return nil
		}
		if err := worktreeCmd.Parse(os.Args[3:]); err != nil {
			return err
		}
		return handleWorktree(ctx, os.Args[2], worktreeCmd.Args())

	case "cleanup":
		if err := cleanupCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
		if cleanupCmd.NArg() > 0 {
			// New cleanup command with ticket ID
			return handleCleanupTicket(ctx, cleanupCmd.Arg(0), *cleanupForce)
		}
		// Old auto-cleanup command
		return handleCleanup(ctx, *cleanupDryRun)

	case "migrate":
		if err := migrateCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
		return handleMigrateDates(ctx, *migrateDryRun)

	case "help", "-h", "--help":
		printUsage()
		return nil

	case "version", "-v", "--version":
		fmt.Printf("ticketflow version %s\n", Version)
		if Version != "dev" || GitCommit != "unknown" {
			fmt.Printf("  Git commit: %s\n", GitCommit)
			fmt.Printf("  Built at:   %s\n", BuildTime)
		}
		return nil

	default:
		return fmt.Errorf("unknown command: %s", os.Args[1])
	}
}

func handleInit(ctx context.Context) error {
	// Special case: init doesn't require existing config
	return cli.InitCommand(ctx)
}

func handleNew(ctx context.Context, slug, format string) error {
	app, err := cli.NewApp(ctx)
	if err != nil {
		return err
	}

	outputFormat := cli.ParseOutputFormat(format)
	cli.GlobalOutputFormat = outputFormat
	return app.NewTicket(ctx, slug, outputFormat)
}

func handleList(ctx context.Context, status string, count int, format string) error {
	app, err := cli.NewApp(ctx)
	if err != nil {
		return err
	}

	var ticketStatus ticket.Status
	if status != "" {
		ticketStatus = ticket.Status(status)
		if !isValidStatus(ticketStatus) {
			return fmt.Errorf("invalid status: %s", status)
		}
	}

	outputFormat := cli.ParseOutputFormat(format)
	cli.GlobalOutputFormat = outputFormat
	return app.ListTickets(ctx, ticketStatus, count, outputFormat)
}

func handleShow(ctx context.Context, ticketID, format string) error {
	app, err := cli.NewApp(ctx)
	if err != nil {
		return err
	}

	// Get the ticket
	t, err := app.Manager.Get(ctx, ticketID)
	if err != nil {
		return err
	}

	outputFormat := cli.ParseOutputFormat(format)
	cli.GlobalOutputFormat = outputFormat
	if outputFormat == cli.FormatJSON {
		// For JSON, just output the ticket data
		return outputJSON(map[string]interface{}{
			"ticket": map[string]interface{}{
				"id":          t.ID,
				"path":        t.Path,
				"status":      string(t.Status()),
				"priority":    t.Priority,
				"description": t.Description,
				"created_at":  t.CreatedAt.Time,
				"started_at":  t.StartedAt.Time,
				"closed_at":   t.ClosedAt.Time,
				"related":     t.Related,
				"content":     t.Content,
			},
		})
	}

	// Text format
	fmt.Printf("ID: %s\n", t.ID)
	fmt.Printf("Status: %s\n", t.Status())
	fmt.Printf("Priority: %d\n", t.Priority)
	fmt.Printf("Description: %s\n", t.Description)
	fmt.Printf("Created: %s\n", t.CreatedAt.Format(time.RFC3339))

	if t.StartedAt.Time != nil {
		fmt.Printf("Started: %s\n", t.StartedAt.Time.Format(time.RFC3339))
	}

	if t.ClosedAt.Time != nil {
		fmt.Printf("Closed: %s\n", t.ClosedAt.Time.Format(time.RFC3339))
	}

	if len(t.Related) > 0 {
		fmt.Printf("Related: %s\n", strings.Join(t.Related, ", "))
	}

	fmt.Printf("\n%s\n", t.Content)

	return nil
}

func handleStart(ctx context.Context, ticketID string, noPush bool) error {
	app, err := cli.NewApp(ctx)
	if err != nil {
		return err
	}

	return app.StartTicket(ctx, ticketID)
}

func handleClose(ctx context.Context, noPush, force bool) error {
	app, err := cli.NewApp(ctx)
	if err != nil {
		return err
	}

	return app.CloseTicket(ctx, force)
}

func handleRestore(ctx context.Context) error {
	app, err := cli.NewApp(ctx)
	if err != nil {
		return err
	}

	return app.RestoreCurrentTicket(ctx)
}

func handleStatus(ctx context.Context, format string) error {
	app, err := cli.NewApp(ctx)
	if err != nil {
		return err
	}

	outputFormat := cli.ParseOutputFormat(format)
	cli.GlobalOutputFormat = outputFormat
	return app.Status(ctx, outputFormat)
}

func handleWorktree(ctx context.Context, subcommand string, args []string) error {
	app, err := cli.NewApp(ctx)
	if err != nil {
		return err
	}

	switch subcommand {
	case "list":
		format := "text"
		if len(args) > 0 && args[0] == "--format" && len(args) > 1 {
			format = args[1]
		}
		outputFormat := cli.ParseOutputFormat(format)
		cli.GlobalOutputFormat = outputFormat
		return app.ListWorktrees(ctx, outputFormat)

	case "clean":
		return app.CleanWorktrees(ctx)

	default:
		printWorktreeUsage()
		return fmt.Errorf("unknown worktree subcommand: %s", subcommand)
	}
}

func isValidStatus(status ticket.Status) bool {
	switch status {
	case ticket.StatusTodo, ticket.StatusDoing, ticket.StatusDone:
		return true
	default:
		return false
	}
}

func printUsage() {
	fmt.Printf(`TicketFlow - Git worktree-based ticket management system (v%s)

USAGE:
  ticketflow                          Start TUI (interactive mode)
  ticketflow init                     Initialize ticket system
  ticketflow new <slug>               Create new ticket
  ticketflow list [options]           List tickets
  ticketflow show <ticket> [options]  Show ticket details
  ticketflow start <ticket> [options] Start working on ticket
  ticketflow close [options]          Complete current ticket
  ticketflow restore                  Restore current-ticket link
  ticketflow status [options]         Show current status
  ticketflow worktree <command>       Manage worktrees
  ticketflow cleanup [options] <ticket> Clean up after PR merge
  ticketflow cleanup [options]        Auto-cleanup orphaned worktrees and stale branches
  ticketflow migrate [options]        Migrate ticket dates to new format
  ticketflow help                     Show this help
  ticketflow version                  Show version

OPTIONS:
  list:
    --status STATUS    Filter by status (todo|doing|done)
    --count N          Maximum number of tickets to show (default: 20)
    --format FORMAT    Output format: text|json (default: text)

  show:
    --format FORMAT    Output format: text|json (default: text)

  start:
    (no options)

  close:
    --force, -f        Force close with uncommitted changes

  status:
    --format FORMAT    Output format: text|json (default: text)

  worktree:
    list [--format FORMAT]  List all worktrees
    clean                   Remove orphaned worktrees

  cleanup <ticket>:
    --force                Skip confirmation prompts

  cleanup (auto):
    --dry-run              Show what would be cleaned without making changes

  migrate:
    --dry-run              Show what would be updated without making changes

EXAMPLES:
  # Initialize in current git repository
  ticketflow init

  # Create a new ticket
  ticketflow new implement-auth

  # Create a sub-ticket (from within a worktree)
  cd ../.worktrees/250124-150000-implement-auth
  ticketflow new auth-database

  # List all todo tickets
  ticketflow list --status todo

  # Start working on a ticket
  ticketflow start 250124-150000-implement-auth

  # Close the current ticket
  ticketflow close

  # Get current status as JSON
  ticketflow status --format json

  # Clean up after PR merge (with force flag)
  ticketflow cleanup --force 250124-150000-implement-auth

  # Auto-cleanup orphaned worktrees and stale branches for done tickets
  ticketflow cleanup --dry-run  # Preview what would be cleaned
  ticketflow cleanup            # Perform cleanup

Use 'ticketflow <command> -h' for command-specific help.
`, Version)
}

func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func printWorktreeUsage() {
	fmt.Println(`TicketFlow Worktree Management

USAGE:
  ticketflow worktree list [--format json]   List all worktrees
  ticketflow worktree clean                   Remove orphaned worktrees

DESCRIPTION:
  The worktree command manages git worktrees associated with tickets.

  list    Shows all worktrees with their paths, branches, and HEAD commits
  clean   Removes worktrees that don't have corresponding active tickets

EXAMPLES:
  # List all worktrees
  ticketflow worktree list

  # List worktrees in JSON format
  ticketflow worktree list --format json

  # Clean up orphaned worktrees
  ticketflow worktree clean`)
}

func handleCleanup(ctx context.Context, dryRun bool) error {
	app, err := cli.NewApp(ctx)
	if err != nil {
		return err
	}

	if dryRun {
		// Show cleanup statistics first
		if err := app.CleanupStats(ctx); err != nil {
			return err
		}
		fmt.Println("\nDry-run mode: No changes will be made.")
	}

	_, err = app.AutoCleanup(ctx, dryRun)
	return err
}

func handleCleanupTicket(ctx context.Context, ticketID string, force bool) error {
	app, err := cli.NewApp(ctx)
	if err != nil {
		return err
	}

	return app.CleanupTicket(ctx, ticketID, force)
}

func handleMigrateDates(ctx context.Context, dryRun bool) error {
	app, err := cli.NewApp(ctx)
	if err != nil {
		return err
	}

	return app.MigrateDates(ctx, dryRun)
}
