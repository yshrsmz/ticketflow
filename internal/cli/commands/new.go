package commands

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/command"
)

// Format constants for output formats
const (
	FormatText = "text"
	FormatJSON = "json"
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
	parent      string
	parentShort string
	format      string
	formatShort string
}

// normalize merges short and long form flags (short form takes precedence)
func (f *newFlags) normalize() {
	if f.parentShort != "" {
		f.parent = f.parentShort
	}
	if f.formatShort != "" {
		f.format = f.formatShort
	}
}

// SetupFlags configures flags for the command
func (c *NewCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	flags := &newFlags{}
	// Long forms
	fs.StringVar(&flags.parent, "parent", "", "Parent ticket ID")
	fs.StringVar(&flags.format, "format", FormatText, "Output format (text|json)")
	// Short forms
	fs.StringVar(&flags.parentShort, "p", "", "Parent ticket ID (short form)")
	fs.StringVar(&flags.formatShort, "o", FormatText, "Output format (short form)")
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
	f.normalize()

	// Validate format flag
	if f.format != FormatText && f.format != FormatJSON {
		return fmt.Errorf("invalid format: %q (must be %q or %q)", f.format, FormatText, FormatJSON)
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

	// Use the existing NewTicket method from App which handles all the business logic
	ticket, err := app.NewTicket(ctx, slug, f.parent)
	if err != nil {
		return err
	}

	// Handle JSON output if requested
	if f.format == FormatJSON {
		output := map[string]interface{}{
			"ticket": map[string]interface{}{
				"id":   ticket.ID,
				"path": ticket.Path,
			},
		}
		// Extract parent ticket ID from Related field if available
		for _, rel := range ticket.Related {
			if strings.HasPrefix(rel, "parent:") {
				output["parent_ticket"] = strings.TrimPrefix(rel, "parent:")
				break
			}
		}
		return app.Output.PrintJSON(output)
	}

	// Text format output is already handled by NewTicket
	return nil
}
