package commands

import (
	"context"
	"flag"
	"fmt"

	"github.com/yshrsmz/ticketflow/internal/command"
)

// VersionCommand implements the version command using the new Command interface
type VersionCommand struct {
	Version   string
	GitCommit string
	BuildTime string
}

// NewVersionCommand creates a new version command with the provided build information
func NewVersionCommand(version, gitCommit, buildTime string) command.Command {
	return &VersionCommand{
		Version:   version,
		GitCommit: gitCommit,
		BuildTime: buildTime,
	}
}

// Name returns the command name
func (c *VersionCommand) Name() string {
	return "version"
}

// Aliases returns alternative names for this command
func (c *VersionCommand) Aliases() []string {
	return []string{"-v", "--version"}
}

// Description returns a short description of the command
func (c *VersionCommand) Description() string {
	return "Show version information"
}

// Usage returns the usage string for the command
func (c *VersionCommand) Usage() string {
	return "version"
}

// SetupFlags configures flags for the command
func (c *VersionCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	// Version command has no flags
	return nil
}

// Validate checks if the command arguments are valid
func (c *VersionCommand) Validate(flags interface{}, args []string) error {
	// No validation needed for version command
	return nil
}

// Execute runs the version command
func (c *VersionCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	fmt.Printf("ticketflow version %s\n", c.Version)
	if c.Version != "dev" || c.GitCommit != "unknown" {
		fmt.Printf("  Git commit: %s\n", c.GitCommit)
		fmt.Printf("  Built at:   %s\n", c.BuildTime)
	}
	return nil
}