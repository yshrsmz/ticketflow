package commands

import (
	"context"
	flag "github.com/spf13/pflag"

	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/command"
)

// InitCommand implements the init command using the new Command interface
type InitCommand struct{}

// NewInitCommand creates a new init command
func NewInitCommand() command.Command {
	return &InitCommand{}
}

// Name returns the command name
func (c *InitCommand) Name() string {
	return "init"
}

// Aliases returns alternative names for this command
func (c *InitCommand) Aliases() []string {
	return nil
}

// Description returns a short description of the command
func (c *InitCommand) Description() string {
	return "Initialize a new ticketflow project"
}

// Usage returns the usage string for the command
func (c *InitCommand) Usage() string {
	return "init"
}

// SetupFlags configures flags for the command
func (c *InitCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	// Init command has no flags
	return nil
}

// Validate checks if the command arguments are valid
func (c *InitCommand) Validate(flags interface{}, args []string) error {
	// No validation needed for init command
	// It doesn't require existing config and creates its own structure
	return nil
}

// Execute runs the init command
func (c *InitCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Delegate to the existing init command implementation
	// This handles all the initialization logic including:
	// - Finding the git project root
	// - Creating default config
	// - Creating ticket directories
	// - Updating .gitignore
	return cli.InitCommand(ctx)
}
