package commands

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

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
	fs.StringVar(&flags.format, "format", "text", "Output format (text|json)")
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
	f, ok := flags.(*showFlags)
	if !ok {
		return fmt.Errorf("invalid flags type: expected *showFlags, got %T", flags)
	}

	// Validate format flag (empty string defaults to "text" for backward compatibility)
	if f.format == "" {
		f.format = "text"
	}
	if f.format != "text" && f.format != "json" {
		return fmt.Errorf("invalid format: %q (must be 'text' or 'json')", f.format)
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

	// Create App instance with dependencies
	app, err := cli.NewApp(ctx)
	if err != nil {
		return err
	}

	// Safely extract flags
	f, ok := flags.(*showFlags)
	if !ok {
		return fmt.Errorf("invalid flags type: expected *showFlags, got %T", flags)
	}

	// Get the ticket using the first positional argument
	ticketID := args[0]
	t, err := app.Manager.Get(ctx, ticketID)
	if err != nil {
		return err
	}

	// Output based on format
	outputFormat := cli.ParseOutputFormat(f.format)
	if outputFormat == cli.FormatJSON {
		// For JSON, output the ticket data with exact structure from original
		// Note: We preserve the original behavior where nil times are output as null
		return app.Output.PrintJSON(map[string]interface{}{
			"ticket": map[string]interface{}{
				"id":          t.ID,
				"path":        t.Path,
				"status":      string(t.Status()),
				"priority":    t.Priority,
				"description": t.Description,
				"created_at":  t.CreatedAt.Time, // Always non-nil
				"started_at":  t.StartedAt.Time, // Can be nil, outputs as null in JSON
				"closed_at":   t.ClosedAt.Time,  // Can be nil, outputs as null in JSON
				"related":     t.Related,
				"content":     t.Content,
			},
		})
	}

	// Text format output
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
