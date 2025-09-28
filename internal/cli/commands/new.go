package commands

import (
	"context"
	"fmt"
	flag "github.com/spf13/pflag"

	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/command"
)

// NewCommand implements the new command using the new Command interface
type NewCommand struct{}

// NewNewCommand creates a new 'new' command
func NewNewCommand() command.Command {
	return &NewCommand{}
}

// Name returns the command name
func (c *NewCommand) Name() string {
	return "new"
}

// Aliases returns alternative names for this command
func (c *NewCommand) Aliases() []string {
	return nil
}

// Description returns a short description of the command
func (c *NewCommand) Description() string {
	return "Create a new ticket"
}

// Usage returns the usage string for the command
func (c *NewCommand) Usage() string {
	return "new [--parent <ticket-id>] [--format text|json] <slug>"
}

// newFlags holds the flags for the new command
type newFlags struct {
	parent string
	format string
}

// normalize is no longer needed - pflag handles this automatically

// SetupFlags configures flags for the command
func (c *NewCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	flags := &newFlags{}
	// Use pflag's StringVarP to register both long and short forms
	fs.StringVarP(&flags.parent, "parent", "p", "", "Parent ticket ID")
	fs.StringVarP(&flags.format, "format", "o", FormatText, "Output format (text|json)")
	return flags
}

// Validate checks if the command arguments are valid
func (c *NewCommand) Validate(flags interface{}, args []string) error {
	// Check for required slug argument
	if len(args) < 1 {
		return fmt.Errorf("missing slug argument")
	}

	// Check for unexpected extra arguments
	if len(args) > 1 {
		return fmt.Errorf("unexpected arguments after slug: %v", args[1:])
	}

	// Safely assert flags type
	f, err := AssertFlags[newFlags](flags)
	if err != nil {
		return err
	}

	// Validate format flag
	if err := ValidateFormat(f.format); err != nil {
		return err
	}

	return nil
}

// Execute runs the new command
func (c *NewCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	// Check for context cancellation early
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Safely extract flags
	f, err := AssertFlags[newFlags](flags)
	if err != nil {
		return err
	}

	// Get flag values
	format := f.format
	parent := f.parent

	// Parse output format first
	outputFormat := cli.ParseOutputFormat(format)
	cli.SetGlobalOutputFormat(outputFormat) // Ensure errors are formatted correctly

	// Create App instance with format
	app, err := cli.NewAppWithFormat(ctx, outputFormat)
	if err != nil {
		return err
	}

	// Get the slug from the first positional argument
	slug := args[0]

	// Use the existing NewTicket method from App which handles all the business logic
	ticket, err := app.NewTicket(ctx, slug, parent)
	if err != nil {
		return err
	}

	// Extract parent ticket ID from Related field
	parentTicketID := ExtractParentFromTicket(ticket)
	// Fall back to explicit parent if not in Related field
	if parentTicketID == "" {
		parentTicketID = parent
	}

	// Create the result and use PrintResult
	result := &cli.NewTicketResult{
		Ticket:       ticket,
		ParentTicket: parentTicketID,
	}

	return app.Output.PrintResult(result)
}
