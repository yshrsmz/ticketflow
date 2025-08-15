package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/cli/commands"
	"github.com/yshrsmz/ticketflow/internal/command"
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

// commandRegistry holds all migrated commands
var commandRegistry = command.NewRegistry()

func init() {
	// Register migrated commands
	if err := commandRegistry.Register(commands.NewVersionCommand(Version, GitCommit, BuildTime)); err != nil {
		// Log error but continue - allow program to run with degraded functionality
		// This should never happen in practice but we handle it gracefully
		fmt.Fprintf(os.Stderr, "Warning: failed to register version command: %v\n", err)
	}

	// Register help command with registry access
	if err := commandRegistry.Register(commands.NewHelpCommand(commandRegistry, Version)); err != nil {
		// Log error but continue - allow program to run with degraded functionality
		// This should never happen in practice but we handle it gracefully
		fmt.Fprintf(os.Stderr, "Warning: failed to register help command: %v\n", err)
	}

	// Register init command
	if err := commandRegistry.Register(commands.NewInitCommand()); err != nil {
		// Log error but continue - allow program to run with degraded functionality
		// This should never happen in practice but we handle it gracefully
		fmt.Fprintf(os.Stderr, "Warning: failed to register init command: %v\n", err)
	}

	// Register status command
	if err := commandRegistry.Register(commands.NewStatusCommand()); err != nil {
		// Log error but continue - allow program to run with degraded functionality
		// This should never happen in practice but we handle it gracefully
		fmt.Fprintf(os.Stderr, "Warning: failed to register status command: %v\n", err)
	}
	// Register list command
	if err := commandRegistry.Register(commands.NewListCommand()); err != nil {
		// Log error but continue - allow program to run with degraded functionality
		// This should never happen in practice but we handle it gracefully
		fmt.Fprintf(os.Stderr, "Warning: failed to register list command: %v\n", err)
	}
	// Register show command
	if err := commandRegistry.Register(commands.NewShowCommand()); err != nil {
		// Log error but continue - allow program to run with degraded functionality
		// This should never happen in practice but we handle it gracefully
		fmt.Fprintf(os.Stderr, "Warning: failed to register show command: %v\n", err)
	}
	// Register new command
	if err := commandRegistry.Register(commands.NewNewCommand()); err != nil {
		// Log error but continue - allow program to run with degraded functionality
		// This should never happen in practice but we handle it gracefully
		fmt.Fprintf(os.Stderr, "Warning: failed to register new command: %v\n", err)
	}
	// Register start command
	if err := commandRegistry.Register(commands.NewStartCommand()); err != nil {
		// Log error but continue - allow program to run with degraded functionality
		// This should never happen in practice but we handle it gracefully
		fmt.Fprintf(os.Stderr, "Warning: failed to register start command: %v\n", err)
	}
	// Register close command
	if err := commandRegistry.Register(commands.NewCloseCommand()); err != nil {
		// Log error but continue - allow program to run with degraded functionality
		// This should never happen in practice but we handle it gracefully
		fmt.Fprintf(os.Stderr, "Warning: failed to register close command: %v\n", err)
	}
	// Register cleanup command
	if err := commandRegistry.Register(commands.NewCleanupCommand()); err != nil {
		// Log error but continue - allow program to run with degraded functionality
		// This should never happen in practice but we handle it gracefully
		fmt.Fprintf(os.Stderr, "Warning: failed to register cleanup command: %v\n", err)
	}

	if err := commandRegistry.Register(commands.NewRestoreCommand()); err != nil {
		// Log error but continue - allow program to run with degraded functionality
		// This should never happen in practice but we handle it gracefully
		fmt.Fprintf(os.Stderr, "Warning: failed to register restore command: %v\n", err)
	}

	// Register worktree command
	if err := commandRegistry.Register(commands.NewWorktreeCommand()); err != nil {
		// Log error but continue - allow program to run with degraded functionality
		// This should never happen in practice but we handle it gracefully
		fmt.Fprintf(os.Stderr, "Warning: failed to register worktree command: %v\n", err)
	}
}

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

	// Load config with a short timeout context for TUI mode
	loadCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cfg, err := config.LoadWithContext(loadCtx, root)
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

// startFlags holds command-line flags for the 'start' command
func runCLI(ctx context.Context) error {
	// Parse command
	if len(os.Args) < 2 {
		// Show help when no command is provided
		if cmd, ok := commandRegistry.Get("help"); ok {
			return executeNewCommand(ctx, cmd, []string{})
		}
		// Fallback if help command is not registered (should never happen)
		fmt.Fprintln(os.Stderr, "Error: help command not available")
		return fmt.Errorf("help command not available")
	}

	// Check if command is in the new registry (for migrated commands)
	// This now handles both direct commands and aliases
	if cmd, ok := commandRegistry.Get(os.Args[1]); ok {
		return executeNewCommand(ctx, cmd, os.Args[2:])
	}

	// All commands have been migrated to the registry
	return fmt.Errorf("unknown command: %s", os.Args[1])
}

func isValidStatus(status ticket.Status) bool {
	switch status {
	case ticket.StatusTodo, ticket.StatusDoing, ticket.StatusDone:
		return true
	default:
		return false
	}
}
