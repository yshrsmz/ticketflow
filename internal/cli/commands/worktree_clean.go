package commands

import (
	"context"
	"flag"
	"fmt"

	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/command"
)

// WorktreeCleanCommand implements the worktree clean subcommand
type WorktreeCleanCommand struct{}

// NewWorktreeCleanCommand creates a new worktree clean command
func NewWorktreeCleanCommand() command.Command {
	return &WorktreeCleanCommand{}
}

// Name returns the command name
func (c *WorktreeCleanCommand) Name() string {
	return "clean"
}

// Aliases returns alternative names for this command
func (c *WorktreeCleanCommand) Aliases() []string {
	return nil
}

// Description returns a short description of the command
func (c *WorktreeCleanCommand) Description() string {
	return "Remove orphaned worktrees"
}

// Usage returns the usage string for the command
func (c *WorktreeCleanCommand) Usage() string {
	return "worktree clean"
}

// SetupFlags configures the flag set for this command
func (c *WorktreeCleanCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	// No flags for the clean command
	return nil
}

// Validate checks if the provided flags and arguments are valid
func (c *WorktreeCleanCommand) Validate(flags interface{}, args []string) error {
	// No additional arguments expected
	if len(args) > 0 {
		return fmt.Errorf("worktree clean command takes no arguments")
	}
	return nil
}

// Execute runs the command with the given context
func (c *WorktreeCleanCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	app, err := cli.NewApp(ctx)
	if err != nil {
		return err
	}

	return app.CleanWorktrees(ctx)
}
