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
	fs.StringVar(&flags.format, "format", FormatText, "Output format (text|json)")
	return flags
}

// Validate checks if the command arguments are valid
func (c *StatusCommand) Validate(flags interface{}, args []string) error {
	// Safely assert flags type
	f, ok := flags.(*statusFlags)
	if !ok {
		return fmt.Errorf("invalid flags type: expected *statusFlags, got %T", flags)
	}

	// Validate format flag
	if f.format != FormatText && f.format != FormatJSON {
		return fmt.Errorf("invalid format: %q (must be 'text' or 'json')", f.format)
	}

	return nil
}

// Execute runs the status command
func (c *StatusCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Safely extract flags
	f, ok := flags.(*statusFlags)
	if !ok {
		return fmt.Errorf("invalid flags type: expected *statusFlags, got %T", flags)
	}
	outputFormat := cli.ParseOutputFormat(f.format)
	cli.SetGlobalOutputFormat(outputFormat) // Ensure errors are formatted correctly

	// Create App instance with the correct format from the start
	app, err := cli.NewAppWithFormat(ctx, outputFormat)
	if err != nil {
		return err
	}

	// Delegate to App's Status method
	return app.Status(ctx, outputFormat)
}
