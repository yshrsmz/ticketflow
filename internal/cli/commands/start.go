package commands

import (
	"context"
	"flag"
	"fmt"

	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/command"
)

// StartCommand implements the start command using the new Command interface
type StartCommand struct{}

// NewStartCommand creates a new start command
func NewStartCommand() command.Command {
	return &StartCommand{}
}

// Name returns the command name
func (c *StartCommand) Name() string {
	return "start"
}

// Aliases returns alternative names for this command
func (c *StartCommand) Aliases() []string {
	return nil
}

// Description returns a short description of the command
func (c *StartCommand) Description() string {
	return "Start work on a ticket"
}

// Usage returns the usage string for the command
func (c *StartCommand) Usage() string {
	return "start [--force] [--format text|json] <ticket-id>"
}

// startFlags holds the flags for the start command
type startFlags struct {
	force  BoolFlag
	format StringFlag
}

// SetupFlags configures flags for the command
func (c *StartCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	flags := &startFlags{}
	RegisterBool(fs, &flags.force, "force", "f", "Force recreate worktree if it already exists")
	RegisterString(fs, &flags.format, "format", "o", FormatText, "Output format (text|json)")
	return flags
}

// Validate checks if the command arguments are valid
func (c *StartCommand) Validate(flags interface{}, args []string) error {
	// Check for required ticket ID argument
	if len(args) < 1 {
		return fmt.Errorf("missing ticket argument")
	}

	// Check for unexpected extra arguments
	if len(args) > 1 {
		return fmt.Errorf("unexpected arguments after ticket ID: %v", args[1:])
	}

	// Safely assert flags type
	f, ok := flags.(*startFlags)
	if !ok {
		return fmt.Errorf("invalid flags type: expected *startFlags, got %T", flags)
	}

	// Validate format flag using resolved value
	format := f.format.Value()
	if format != FormatText && format != FormatJSON {
		return fmt.Errorf("invalid format: %q (must be %q or %q)", format, FormatText, FormatJSON)
	}

	return nil
}

// Execute runs the start command
func (c *StartCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	// Check for context cancellation early
	// This is a defensive programming practice to fail fast if context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Extract flags - Validate already checked this, but we still need to handle the type assertion
	// for defensive programming (e.g., if Execute is called directly without Validate)
	f, ok := flags.(*startFlags)
	if !ok {
		return fmt.Errorf("invalid flags type: expected *startFlags, got %T", flags)
	}

	// Check args length for defensive programming (in case Execute is called without Validate)
	if len(args) < 1 {
		return fmt.Errorf("missing ticket argument")
	}

	// Get resolved values
	format := f.format.Value()
	force := f.force.Value()

	// Parse output format first
	outputFormat := cli.ParseOutputFormat(format)
	cli.SetGlobalOutputFormat(outputFormat) // Ensure errors are formatted correctly

	// Create App instance with format
	app, err := getAppWithFormat(ctx, outputFormat)
	if err != nil {
		return err
	}

	// Get the ticket ID from the first positional argument
	ticketID := args[0]

	// Use the existing StartTicket method from App which handles all the business logic
	result, err := app.StartTicket(ctx, ticketID, force)
	if err != nil {
		return err
	}

	// Create StartResult wrapper for Printable interface
	startResult := &cli.StartResult{
		StartTicketResult: result,
		WorktreeEnabled:   app.Config.Worktree.Enabled,
	}

	return app.Output.PrintResult(startResult)
}
