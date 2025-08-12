package commands

import (
	"context"
	"flag"
	"fmt"

	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/command"
)

// StatusCommand implements the status command using the new Command interface
type StatusCommand struct{}

// NewStatusCommand creates a new status command
func NewStatusCommand() command.Command {
	return &StatusCommand{}
}

// Name returns the command name
func (c *StatusCommand) Name() string {
	return "status"
}

// Aliases returns alternative names for this command
func (c *StatusCommand) Aliases() []string {
	return nil
}

// Description returns a short description of the command
func (c *StatusCommand) Description() string {
	return "Show the status of the current ticket"
}

// Usage returns the usage string for the command
func (c *StatusCommand) Usage() string {
	return "status [--format text|json]"
}

// statusFlags holds the flags for the status command
type statusFlags struct {
	format string
}

// SetupFlags configures flags for the command
func (c *StatusCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	flags := &statusFlags{}
	fs.StringVar(&flags.format, "format", "text", "Output format (text|json)")
	return flags
}

// Validate checks if the command arguments are valid
func (c *StatusCommand) Validate(flags interface{}, args []string) error {
	// Extract flags
	f := flags.(*statusFlags)

	// Validate format flag
	if f.format != "text" && f.format != "json" {
		return fmt.Errorf("invalid format: %q (must be 'text' or 'json')", f.format)
	}

	return nil
}

// Execute runs the status command
func (c *StatusCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	// Create App instance with dependencies
	app, err := cli.NewApp(ctx)
	if err != nil {
		return err
	}

	// Extract flags
	f := flags.(*statusFlags)
	outputFormat := cli.ParseOutputFormat(f.format)

	// Delegate to App's Status method
	return app.Status(ctx, outputFormat)
}
