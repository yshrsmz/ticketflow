package commands

import (
	"context"
	"flag"
	"fmt"

	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/command"
)

// ShowCommand implements the show command using the new Command interface
type ShowCommand struct{}

// NewShowCommand creates a new show command
func NewShowCommand() command.Command {
	return &ShowCommand{}
}

// Name returns the command name
func (c *ShowCommand) Name() string {
	return "show"
}

// Aliases returns alternative names for this command
func (c *ShowCommand) Aliases() []string {
	return nil
}

// Description returns a short description of the command
func (c *ShowCommand) Description() string {
	return "Show ticket details"
}

// Usage returns the usage string for the command
func (c *ShowCommand) Usage() string {
	return "show <ticket-id> [--format text|json]"
}

// showFlags holds the flags for the show command
type showFlags struct {
	format string
}

// SetupFlags configures flags for the command
func (c *ShowCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	flags := &showFlags{}
	fs.StringVar(&flags.format, "format", FormatText, "Output format (text|json)")
	return flags
}

// Validate checks if the command arguments are valid
func (c *ShowCommand) Validate(flags interface{}, args []string) error {
	// Check for required ticket ID argument
	if len(args) < 1 {
		return fmt.Errorf("missing ticket ID argument")
	}

	// Check for unexpected extra arguments
	if len(args) > 1 {
		return fmt.Errorf("unexpected arguments after ticket ID: %v", args[1:])
	}

	// Safely assert flags type
	f, err := AssertFlags[showFlags](flags)
	if err != nil {
		return err
	}

	// Validate format flag (empty string defaults to text for backward compatibility)
	if f.format == "" {
		f.format = FormatText
	}
	if err := ValidateFormat(f.format); err != nil {
		return err
	}

	return nil
}

// Execute runs the show command
func (c *ShowCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Safely extract flags
	f, err := AssertFlags[showFlags](flags)
	if err != nil {
		return err
	}

	// Parse output format first
	outputFormat := cli.ParseOutputFormat(f.format)
	cli.SetGlobalOutputFormat(outputFormat) // Ensure errors are formatted correctly

	// Create App instance with format
	app, err := cli.NewAppWithFormat(ctx, outputFormat)
	if err != nil {
		return err
	}

	// Get the ticket using the first positional argument
	ticketID := args[0]
	t, err := app.Manager.Get(ctx, ticketID)
	if err != nil {
		return err
	}

	// Use TicketResult for Printable interface
	result := &cli.TicketResult{
		Ticket: t,
	}

	return app.Output.PrintResult(result)
}
