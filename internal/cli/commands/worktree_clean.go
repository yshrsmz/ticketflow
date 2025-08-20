package commands

import (
	"context"
	"flag"
	"fmt"

	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/command"
)

// WorktreeCleanCommand implements the worktree clean subcommand
type WorktreeCleanCommand struct{}

// NewWorktreeCleanCommand creates a new worktree clean command
func NewWorktreeCleanCommand() command.Command {
	return &WorktreeCleanCommand{}
}

// Name returns the command name
func (c *WorktreeCleanCommand) Name() string {
	return "clean"
}

// Aliases returns alternative names for this command
func (c *WorktreeCleanCommand) Aliases() []string {
	return nil
}

// Description returns a short description of the command
func (c *WorktreeCleanCommand) Description() string {
	return "Remove orphaned worktrees"
}

// Usage returns the usage string for the command
func (c *WorktreeCleanCommand) Usage() string {
	return "worktree clean [--format text|json]"
}

// worktreeCleanFlags holds the flags for the worktree clean command
type worktreeCleanFlags struct {
	format      string
	formatShort string
}

// normalize merges short and long form flags (short form takes precedence)
func (f *worktreeCleanFlags) normalize() {
	// Only use formatShort if it was explicitly set (not empty)
	if f.formatShort != "" {
		f.format = f.formatShort
	}
}

// SetupFlags configures the flag set for this command
func (c *WorktreeCleanCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	flags := &worktreeCleanFlags{}
	// Long forms
	fs.StringVar(&flags.format, "format", FormatText, "Output format (text|json)")
	// Short forms
	fs.StringVar(&flags.formatShort, "o", "", "Output format (short form)")
	return flags
}

// Validate checks if the provided flags and arguments are valid
func (c *WorktreeCleanCommand) Validate(flags interface{}, args []string) error {
	// No additional arguments expected
	if len(args) > 0 {
		return fmt.Errorf("worktree clean command takes no arguments")
	}

	// Safely assert flags type
	f, ok := flags.(*worktreeCleanFlags)
	if !ok {
		return fmt.Errorf("invalid flags type: expected *worktreeCleanFlags, got %T", flags)
	}

	// Normalize flags
	f.normalize()

	// Validate format flag
	if f.format != FormatText && f.format != FormatJSON {
		return fmt.Errorf("invalid format: %q (must be %q or %q)", f.format, FormatText, FormatJSON)
	}

	return nil
}

// Execute runs the command with the given context
func (c *WorktreeCleanCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Safely assert flags type
	f, ok := flags.(*worktreeCleanFlags)
	if !ok {
		return fmt.Errorf("invalid flags type: expected *worktreeCleanFlags, got %T", flags)
	}

	// Parse output format first
	outputFormat := cli.ParseOutputFormat(f.format)
	cli.SetGlobalOutputFormat(outputFormat) // Ensure errors are formatted correctly

	// Create App instance with format
	app, err := cli.NewAppWithFormat(ctx, outputFormat)
	if err != nil {
		return err
	}

	// Execute the clean operation
	result, err := app.CleanWorktrees(ctx)
	if err != nil {
		return err
	}

	// Handle JSON output if requested
	if outputFormat == cli.FormatJSON {
		output := map[string]interface{}{
			"success":           true,
			"cleaned_count":     result.CleanedCount,
			"cleaned_worktrees": result.CleanedWorktrees,
			"failed_worktrees":  result.FailedWorktrees,
			"total_worktrees":   result.TotalWorktrees,
			"active_tickets":    result.ActiveTickets,
		}
		return app.Output.PrintJSON(output)
	}

	// Text output is already handled by CleanWorktrees
	return nil
}
