package commands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/command"
)

// versionFlags holds the flags for the version command
type versionFlags struct {
	format string
}

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
	return "version [--format text|json]"
}

// SetupFlags configures flags for the command
func (c *VersionCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	flags := &versionFlags{}
	fs.StringVar(&flags.format, "format", FormatText, "Output format (text|json)")
	return flags
}

// Validate checks if the command arguments are valid
func (c *VersionCommand) Validate(flags interface{}, args []string) error {
	if flags != nil {
		f := flags.(*versionFlags)
		if f.format != FormatText && f.format != FormatJSON {
			return fmt.Errorf("invalid format: %s (must be '%s' or '%s')", f.format, FormatText, FormatJSON)
		}
	}
	return nil
}

// Execute runs the version command
func (c *VersionCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Determine output format
	format := FormatText
	if flags != nil {
		f := flags.(*versionFlags)
		format = f.format
	}

	// Set global format for error handling
	outputFormat := cli.ParseOutputFormat(format)
	cli.SetGlobalOutputFormat(outputFormat)

	// Output version information based on format
	if format == FormatJSON {
		versionInfo := map[string]interface{}{
			"version":    c.Version,
			"git_commit": c.GitCommit,
			"build_time": c.BuildTime,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(versionInfo)
	}

	// Text format output
	fmt.Printf("ticketflow version %s\n", c.Version)
	if c.Version != "dev" || c.GitCommit != "unknown" {
		fmt.Printf("  Git commit: %s\n", c.GitCommit)
		fmt.Printf("  Built at:   %s\n", c.BuildTime)
	}
	return nil
}
