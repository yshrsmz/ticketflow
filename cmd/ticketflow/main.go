package main

import (
	"context"
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

	// TUI mode does not use the cancellable context created above,
	// because the Bubble Tea framework manages its own signal handling
	// and shutdown logic internally.
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

// Define flag structures for commands that need them
type listFlags struct {
	status string
	count  int
	format string
}

type showFlags struct {
	format string
}

type closeFlags struct {
	force      bool
	forceShort bool
}

type statusFlags struct {
	format string
}

type cleanupFlags struct {
	dryRun bool
	force  bool
}

type migrateFlags struct {
	dryRun bool
}

// newFlags holds command-line flags for the 'new' command
type newFlags struct {
	parent      string // Parent ticket ID (long form)
	parentShort string // Parent ticket ID (short form)
	format      string // Output format (text or json)
}

func runCLI(ctx context.Context) error {
	// Parse command
	if len(os.Args) < 2 {
		printUsage()
		return nil
	}

	// Parse command
	switch os.Args[1] {
	case "init":
		return parseAndExecute(ctx, Command{
			Name: "init",
			Execute: func(ctx context.Context, fs *flag.FlagSet, flags interface{}) error {
				return handleInit(ctx)
			},
		}, os.Args[2:])

	case "new":
		return parseAndExecute(ctx, Command{
			Name:         "new",
			MinArgs:      1,
			MinArgsError: "missing slug argument",
			SetupFlags: func(fs *flag.FlagSet) interface{} {
				flags := &newFlags{}
				fs.StringVar(&flags.parent, "parent", "", "Parent ticket ID")
				fs.StringVar(&flags.parentShort, "p", "", "Parent ticket ID (short form)")
				fs.StringVar(&flags.format, "format", "text", "Output format (text|json)")
				return flags
			},
			Execute: func(ctx context.Context, fs *flag.FlagSet, cmdFlags interface{}) error {
				flags := cmdFlags.(*newFlags)
				parent := flags.parent
				if parent == "" {
					parent = flags.parentShort
				}
				return handleNew(ctx, fs.Arg(0), parent, flags.format)
			},
		}, os.Args[2:])

	case "list":
		return parseAndExecute(ctx, Command{
			Name: "list",
			SetupFlags: func(fs *flag.FlagSet) interface{} {
				flags := &listFlags{}
				fs.StringVar(&flags.status, "status", "", "Filter by status (todo|doing|done)")
				fs.IntVar(&flags.count, "count", 20, "Maximum number of tickets to show")
				fs.StringVar(&flags.format, "format", "text", "Output format (text|json)")
				return flags
			},
			Execute: func(ctx context.Context, fs *flag.FlagSet, cmdFlags interface{}) error {
				flags := cmdFlags.(*listFlags)
				return handleList(ctx, flags.status, flags.count, flags.format)
			},
		}, os.Args[2:])

	case "show":
		return parseAndExecute(ctx, Command{
			Name:         "show",
			MinArgs:      1,
			MinArgsError: "missing ticket argument",
			SetupFlags: func(fs *flag.FlagSet) interface{} {
				flags := &showFlags{}
				fs.StringVar(&flags.format, "format", "text", "Output format (text|json)")
				return flags
			},
			Execute: func(ctx context.Context, fs *flag.FlagSet, cmdFlags interface{}) error {
				flags := cmdFlags.(*showFlags)
				return handleShow(ctx, fs.Arg(0), flags.format)
			},
		}, os.Args[2:])

	case "start":
		return parseAndExecute(ctx, Command{
			Name:         "start",
			MinArgs:      1,
			MinArgsError: "missing ticket argument",
			Execute: func(ctx context.Context, fs *flag.FlagSet, flags interface{}) error {
				return handleStart(ctx, fs.Arg(0), false)
			},
		}, os.Args[2:])

	case "close":
		return parseAndExecute(ctx, Command{
			Name: "close",
			SetupFlags: func(fs *flag.FlagSet) interface{} {
				flags := &closeFlags{}
				fs.BoolVar(&flags.force, "force", false, "Force close with uncommitted changes")
				fs.BoolVar(&flags.forceShort, "f", false, "Force close (short form)")
				return flags
			},
			Execute: func(ctx context.Context, fs *flag.FlagSet, cmdFlags interface{}) error {
				flags := cmdFlags.(*closeFlags)
				force := flags.force || flags.forceShort
				return handleClose(ctx, false, force)
			},
		}, os.Args[2:])

	case "restore":
		return parseAndExecute(ctx, Command{
			Name: "restore",
			Execute: func(ctx context.Context, fs *flag.FlagSet, flags interface{}) error {
				return handleRestore(ctx)
			},
		}, os.Args[2:])

	case "status":
		return parseAndExecute(ctx, Command{
			Name: "status",
			SetupFlags: func(fs *flag.FlagSet) interface{} {
				flags := &statusFlags{}
				fs.StringVar(&flags.format, "format", "text", "Output format (text|json)")
				return flags
			},
			Execute: func(ctx context.Context, fs *flag.FlagSet, cmdFlags interface{}) error {
				flags := cmdFlags.(*statusFlags)
				return handleStatus(ctx, flags.format)
			},
		}, os.Args[2:])

	case "worktree":
		if len(os.Args) < 3 {
			printWorktreeUsage()
			return nil
		}
		return parseAndExecute(ctx, Command{
			Name: "worktree",
			Execute: func(ctx context.Context, fs *flag.FlagSet, flags interface{}) error {
				return handleWorktree(ctx, os.Args[2], fs.Args())
			},
		}, os.Args[3:])

	case "cleanup":
		return parseAndExecute(ctx, Command{
			Name: "cleanup",
			SetupFlags: func(fs *flag.FlagSet) interface{} {
				flags := &cleanupFlags{}
				fs.BoolVar(&flags.dryRun, "dry-run", false, "Show what would be cleaned without making changes")
				fs.BoolVar(&flags.force, "force", false, "Skip confirmation prompts")
				return flags
			},
			Execute: func(ctx context.Context, fs *flag.FlagSet, cmdFlags interface{}) error {
				flags := cmdFlags.(*cleanupFlags)
				if fs.NArg() > 0 {
					// New cleanup command with ticket ID
					return handleCleanupTicket(ctx, fs.Arg(0), flags.force)
				}
				// Old auto-cleanup command
				return handleCleanup(ctx, flags.dryRun)
			},
		}, os.Args[2:])

	case "migrate":
		return parseAndExecute(ctx, Command{
			Name: "migrate",
			SetupFlags: func(fs *flag.FlagSet) interface{} {
				flags := &migrateFlags{}
				fs.BoolVar(&flags.dryRun, "dry-run", false, "Show what would be updated without making changes")
				return flags
			},
			Execute: func(ctx context.Context, fs *flag.FlagSet, cmdFlags interface{}) error {
				flags := cmdFlags.(*migrateFlags)
				return handleMigrateDates(ctx, flags.dryRun)
			},
		}, os.Args[2:])

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

func handleNew(ctx context.Context, slug, parent, format string) error {
	app, err := cli.NewApp(ctx)
	if err != nil {
		return err
	}

	outputFormat := cli.ParseOutputFormat(format)
	return app.NewTicket(ctx, slug, parent, outputFormat)
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
	if outputFormat == cli.FormatJSON {
		// For JSON, just output the ticket data
		return app.Output.PrintJSON(map[string]interface{}{
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
	app.Output.Printf("ID: %s\n", t.ID)
	app.Output.Printf("Status: %s\n", t.Status())
	app.Output.Printf("Priority: %d\n", t.Priority)
	app.Output.Printf("Description: %s\n", t.Description)
	app.Output.Printf("Created: %s\n", t.CreatedAt.Format(time.RFC3339))

	if t.StartedAt.Time != nil {
		app.Output.Printf("Started: %s\n", t.StartedAt.Time.Format(time.RFC3339))
	}

	if t.ClosedAt.Time != nil {
		app.Output.Printf("Closed: %s\n", t.ClosedAt.Time.Format(time.RFC3339))
	}

	if len(t.Related) > 0 {
		app.Output.Printf("Related: %s\n", strings.Join(t.Related, ", "))
	}

	app.Output.Printf("\n%s\n", t.Content)

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
  ticketflow new <slug> [options]     Create new ticket
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
  All commands support logging options:
    --log-level LEVEL   Log level (debug, info, warn, error)
    --log-format FORMAT Log format (text, json)
    --log-output OUTPUT Log output (stderr, stdout, or file path)

  new:
    --parent TICKET    Specify parent ticket ID
    -p TICKET          Short form of --parent
    --format FORMAT    Output format: text|json (default: text)

  list:
    --status STATUS    Filter by status (todo|doing|done)
    --count N          Maximum number of tickets to show (default: 20)
    --format FORMAT    Output format: text|json (default: text)

  show:
    --format FORMAT    Output format: text|json (default: text)

  start:
    (no specific options)

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
