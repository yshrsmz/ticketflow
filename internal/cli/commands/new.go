package commands

import (
	"context"
	"flag"
	"fmt"

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
	return "new <slug> [--parent <ticket-id>] [--format text|json]"
}

// newFlags holds the flags for the new command
type newFlags struct {
	parent      string
	parentShort string
	format      string
	formatShort string
}

// SetupFlags configures flags for the command
func (c *NewCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	flags := &newFlags{}
	// Long forms
	fs.StringVar(&flags.parent, "parent", "", "Parent ticket ID")
	fs.StringVar(&flags.format, "format", "text", "Output format (text|json)")
	// Short forms
	fs.StringVar(&flags.parentShort, "p", "", "Parent ticket ID (short form)")
	fs.StringVar(&flags.formatShort, "o", "text", "Output format (short form)")
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
	f, ok := flags.(*newFlags)
	if !ok {
		return fmt.Errorf("invalid flags type: expected *newFlags, got %T", flags)
	}

	// Merge short and long forms (short form takes precedence if both provided)
	if f.parentShort != "" {
		f.parent = f.parentShort
	}
	if f.formatShort != "" && f.formatShort != "text" {
		f.format = f.formatShort
	}

	// Validate format flag
	if f.format != "text" && f.format != "json" {
		return fmt.Errorf("invalid format: %q (must be 'text' or 'json')", f.format)
	}

	return nil
}

// Execute runs the new command
func (c *NewCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	// Create App instance with dependencies
	app, err := cli.NewApp(ctx)
	if err != nil {
		return err
	}

	// Safely extract flags
	f, ok := flags.(*newFlags)
	if !ok {
		return fmt.Errorf("invalid flags type: expected *newFlags, got %T", flags)
	}

	// Get the slug from the first positional argument
	slug := args[0]

	// Parse output format
	outputFormat := cli.ParseOutputFormat(f.format)

	// Use the existing NewTicket method from App which handles all the business logic
	return app.NewTicket(ctx, slug, f.parent, outputFormat)
}
