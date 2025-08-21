package commands

import (
	"context"
	"flag"
	"fmt"

	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/command"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// CleanupCommand implements the cleanup command using the new Command interface
type CleanupCommand struct{}

// NewCleanupCommand creates a new cleanup command
func NewCleanupCommand() command.Command {
	return &CleanupCommand{}
}

// Name returns the command name
func (c *CleanupCommand) Name() string {
	return "cleanup"
}

// Aliases returns alternative names for this command
func (c *CleanupCommand) Aliases() []string {
	return nil
}

// Description returns a short description of the command
func (c *CleanupCommand) Description() string {
	return "Clean up worktrees and branches"
}

// Usage returns the usage string for the command
func (c *CleanupCommand) Usage() string {
	return "cleanup [--dry-run] [--force] [--format text|json] [<ticket-id>]"
}

// cleanupFlags holds the flags for the cleanup command
type cleanupFlags struct {
	dryRun bool
	force  BoolFlag
	format StringFlag
	args   []string // Store validated arguments
}

// SetupFlags configures flags for the command
func (c *CleanupCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	flags := &cleanupFlags{}
	fs.BoolVar(&flags.dryRun, "dry-run", false, "Show what would be cleaned without making changes")
	RegisterBool(fs, &flags.force, "force", "f", "Skip confirmation prompts")
	RegisterString(fs, &flags.format, "format", "o", FormatText, "Output format (text|json)")
	return flags
}

// Validate checks if the command arguments are valid
func (c *CleanupCommand) Validate(flags interface{}, args []string) error {
	// Check for unexpected extra arguments
	if len(args) > 1 {
		return fmt.Errorf("unexpected arguments after ticket ID: %v", args[1:])
	}

	// Safely assert flags type
	f, ok := flags.(*cleanupFlags)
	if !ok {
		return fmt.Errorf("invalid flags type: expected *cleanupFlags, got %T", flags)
	}

	// Store arguments for Execute method
	f.args = args

	// Validate format flag using resolved value
	format := f.format.Value()
	if format != FormatText && format != FormatJSON {
		return fmt.Errorf("invalid format: %q (must be %q or %q)", format, FormatText, FormatJSON)
	}

	// dry-run flag only makes sense for auto-cleanup mode (no ticket ID)
	if f.dryRun && len(args) > 0 {
		return fmt.Errorf("--dry-run cannot be used when cleaning up a specific ticket")
	}

	return nil
}

// Execute runs the cleanup command
func (c *CleanupCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	// Check for context cancellation at the start
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Safely assert flags type
	f, ok := flags.(*cleanupFlags)
	if !ok {
		return fmt.Errorf("invalid flags type: expected *cleanupFlags, got %T", flags)
	}

	// Get resolved values
	format := f.format.Value()

	// Get app instance with the correct output format from the start
	outputFormat := cli.ParseOutputFormat(format)
	cli.SetGlobalOutputFormat(outputFormat) // Ensure errors are formatted correctly
	app, err := cli.NewAppWithFormat(ctx, outputFormat)
	if err != nil {
		return err
	}

	// Perform the cleanup operation based on whether ticket ID was provided
	if len(f.args) == 0 {
		// Auto-cleanup mode
		return c.executeAutoCleanup(ctx, app, f)
	}

	// Ticket-specific cleanup mode
	return c.executeTicketCleanup(ctx, app, f)
}

// executeAutoCleanup handles the auto-cleanup mode (no ticket ID provided)
func (c *CleanupCommand) executeAutoCleanup(ctx context.Context, app *cli.App, flags *cleanupFlags) error {
	// Get resolved format value
	format := flags.format.Value()

	// Perform auto-cleanup (or dry-run)
	result, err := app.AutoCleanup(ctx, flags.dryRun)
	if err != nil {
		if format == FormatJSON {
			return outputAutoCleanupErrorJSON(app, err)
		}
		return err
	}

	// Output results
	if format == FormatJSON {
		return outputAutoCleanupJSON(app, result)
	}

	// Text output
	return app.Output.PrintResult(result)
}

// executeTicketCleanup handles the ticket-specific cleanup mode
func (c *CleanupCommand) executeTicketCleanup(ctx context.Context, app *cli.App, flags *cleanupFlags) error {
	ticketID := flags.args[0]

	// Get resolved values
	format := flags.format.Value()
	force := flags.force.Value()

	// Perform ticket cleanup
	cleanedTicket, err := app.CleanupTicket(ctx, ticketID, force)
	if err != nil {
		if format == FormatJSON {
			return outputTicketCleanupErrorJSON(app, err)
		}
		return err
	}

	// Output results
	if format == FormatJSON {
		return outputTicketCleanupJSON(app, cleanedTicket)
	}

	// Text output
	outputTicketCleanupText(cleanedTicket)
	return nil
}

// outputTicketCleanupText outputs ticket cleanup results in text format
func outputTicketCleanupText(t *ticket.Ticket) {
	fmt.Printf("Successfully cleaned up ticket: %s\n", t.ID)
	fmt.Printf("Description: %s\n", t.Description)
}

// outputAutoCleanupJSON outputs auto-cleanup results in JSON format
func outputAutoCleanupJSON(app *cli.App, result *cli.CleanupResult) error {
	output := map[string]interface{}{
		"success": true,
		"result": map[string]interface{}{
			"orphaned_worktrees": result.OrphanedWorktrees,
			"stale_branches":     result.StaleBranches,
			"errors":             result.Errors,
		},
	}

	return app.Output.PrintJSON(output)
}

// outputTicketCleanupJSON outputs ticket cleanup results in JSON format
func outputTicketCleanupJSON(app *cli.App, t *ticket.Ticket) error {
	output := map[string]interface{}{
		"success": true,
		"ticket": map[string]interface{}{
			"id":          t.ID,
			"description": t.Description,
			"status":      string(t.Status()),
			"priority":    t.Priority,
			"created_at":  t.CreatedAt,
			"closed_at":   t.ClosedAt,
		},
	}

	return app.Output.PrintJSON(output)
}

// outputAutoCleanupErrorJSON outputs auto-cleanup error in JSON format
func outputAutoCleanupErrorJSON(app *cli.App, err error) error {
	output := map[string]interface{}{
		"success": false,
		"error":   err.Error(),
	}

	return app.Output.PrintJSON(output)
}

// outputTicketCleanupErrorJSON outputs ticket cleanup error in JSON format
func outputTicketCleanupErrorJSON(app *cli.App, err error) error {
	output := map[string]interface{}{
		"success": false,
		"error":   err.Error(),
	}

	return app.Output.PrintJSON(output)
}
