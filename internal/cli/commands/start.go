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
	force       bool
	forceShort  bool
	format      string
	formatShort string
}

// normalize merges short and long form flags (short form takes precedence)
func (f *startFlags) normalize() {
	if f.forceShort {
		f.force = f.forceShort
	}
	if f.formatShort != "" {
		f.format = f.formatShort
	}
}

// SetupFlags configures flags for the command
func (c *StartCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	flags := &startFlags{}
	// Long forms
	fs.BoolVar(&flags.force, "force", false, "Force recreate worktree if it already exists")
	fs.StringVar(&flags.format, "format", FormatText, "Output format (text|json)")
	// Short forms
	fs.BoolVar(&flags.forceShort, "f", false, "Force recreate worktree (short form)")
	fs.StringVar(&flags.formatShort, "o", FormatText, "Output format (short form)")
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

	// Merge short and long forms (short form takes precedence if both provided)
	f.normalize()

	// Validate format flag
	if f.format != FormatText && f.format != FormatJSON {
		return fmt.Errorf("invalid format: %q (must be %q or %q)", f.format, FormatText, FormatJSON)
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

	// Create App instance with dependencies
	app, err := cli.NewApp(ctx)
	if err != nil {
		return err
	}

	// Extract flags - Validate already checked this, but we still need to handle the type assertion
	// for defensive programming (e.g., if Execute is called directly without Validate)
	f, ok := flags.(*startFlags)
	if !ok {
		return fmt.Errorf("invalid flags type: expected *startFlags, got %T", flags)
	}

	// Get the ticket ID from the first positional argument
	ticketID := args[0]

	// Parse output format
	outputFormat := cli.ParseOutputFormat(f.format)

	// Use the existing StartTicket method from App which handles all the business logic
	return app.StartTicket(ctx, ticketID, f.force, outputFormat)
}
