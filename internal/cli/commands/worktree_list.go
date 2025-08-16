package commands

import (
	"context"
	"flag"
	"fmt"

	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/command"
)

// WorktreeListCommand implements the worktree list subcommand
type WorktreeListCommand struct{}

// NewWorktreeListCommand creates a new worktree list command
func NewWorktreeListCommand() command.Command {
	return &WorktreeListCommand{}
}

// Name returns the command name
func (c *WorktreeListCommand) Name() string {
	return "list"
}

// Aliases returns alternative names for this command
func (c *WorktreeListCommand) Aliases() []string {
	return nil
}

// Description returns a short description of the command
func (c *WorktreeListCommand) Description() string {
	return "List all worktrees"
}

// Usage returns the usage string for the command
func (c *WorktreeListCommand) Usage() string {
	return "worktree list [--format json]"
}

// worktreeListFlags holds the flags for the worktree list command
type worktreeListFlags struct {
	format      string
	formatShort string
}

// SetupFlags configures the flag set for this command
func (c *WorktreeListCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	flags := &worktreeListFlags{}
	fs.StringVar(&flags.format, "format", FormatText, "Output format (text, json)")
	fs.StringVar(&flags.formatShort, "o", FormatText, "Output format (text, json)")
	return flags
}

// Validate checks if the provided flags and arguments are valid
func (c *WorktreeListCommand) Validate(flags interface{}, args []string) error {
	// No additional arguments expected
	if len(args) > 0 {
		return fmt.Errorf("worktree list command takes no arguments")
	}

	// If flags is nil, there's nothing to validate
	if flags == nil {
		return nil
	}

	f := flags.(*worktreeListFlags)

	// Handle short form
	if f.formatShort != "" && f.formatShort != FormatText {
		f.format = f.formatShort
	}

	// Validate format (empty string defaults to text which is valid)
	if f.format != "" && f.format != FormatText && f.format != FormatJSON {
		return fmt.Errorf("invalid format: %s (must be '%s' or '%s')", f.format, FormatText, FormatJSON)
	}

	return nil
}

// Execute runs the command with the given context
func (c *WorktreeListCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Default format
	format := FormatText

	if flags != nil {
		f := flags.(*worktreeListFlags)
		if f.format != "" {
			format = f.format
		}
	}

	outputFormat := cli.ParseOutputFormat(format)

	// Create App instance with the correct format from the start
	app, err := cli.NewAppWithFormat(ctx, outputFormat)
	if err != nil {
		return err
	}

	return app.ListWorktrees(ctx, outputFormat)
}
