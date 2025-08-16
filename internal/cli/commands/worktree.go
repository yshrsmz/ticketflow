package commands

import (
	"context"
	"flag"
	"fmt"

	"github.com/yshrsmz/ticketflow/internal/command"
)

const (
	// errUnknownSubcommand is the error message for unknown worktree subcommands
	errUnknownSubcommand = "unknown worktree subcommand: %s"
)

// WorktreeCommand implements the worktree parent command using the new Command interface
type WorktreeCommand struct {
	subcommands map[string]command.Command
}

// NewWorktreeCommand creates a new worktree command with its subcommands
func NewWorktreeCommand() command.Command {
	return &WorktreeCommand{
		subcommands: map[string]command.Command{
			"list":  NewWorktreeListCommand(),
			"clean": NewWorktreeCleanCommand(),
		},
	}
}

// Name returns the command name
func (c *WorktreeCommand) Name() string {
	return "worktree"
}

// Aliases returns alternative names for this command
func (c *WorktreeCommand) Aliases() []string {
	return nil
}

// Description returns a short description of the command
func (c *WorktreeCommand) Description() string {
	return "Manage git worktrees associated with tickets"
}

// Usage returns the usage string for the command
func (c *WorktreeCommand) Usage() string {
	return "worktree <subcommand> [options]"
}

// SetupFlags configures the flag set for this command
func (c *WorktreeCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	// No flags for the parent command
	return nil
}

// Validate checks if the provided flags and arguments are valid
func (c *WorktreeCommand) Validate(flags interface{}, args []string) error {
	// We don't validate here - let subcommands handle their own validation
	// If no subcommand is provided, we'll show usage in Execute
	return nil
}

// Execute runs the command with the given context
func (c *WorktreeCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if len(args) == 0 {
		c.printUsage()
		return nil
	}

	subcmdName := args[0]
	subcmd, ok := c.subcommands[subcmdName]
	if !ok {
		c.printUsage()
		return fmt.Errorf(errUnknownSubcommand, subcmdName)
	}

	// Parse flags for the subcommand
	fs := flag.NewFlagSet(fmt.Sprintf("worktree %s", subcmdName), flag.ExitOnError)
	subcmdFlags := subcmd.SetupFlags(fs)

	// Parse remaining arguments
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	// Validate the subcommand
	if err := subcmd.Validate(subcmdFlags, fs.Args()); err != nil {
		return err
	}

	// Execute the subcommand
	return subcmd.Execute(ctx, subcmdFlags, fs.Args())
}

// printUsage prints the usage information for the worktree command
func (c *WorktreeCommand) printUsage() {
	fmt.Println(`TicketFlow Worktree Management

USAGE:
  ticketflow worktree list [--format json]   List all worktrees
  ticketflow worktree clean                   Remove orphaned worktrees

DESCRIPTION:
  The worktree command manages git worktrees associated with tickets.

  list    Shows all worktrees with their paths, branches, and HEAD commits
  clean   Removes worktrees that don't have corresponding active tickets

EXAMPLES:
  # List all worktrees
  ticketflow worktree list

  # List worktrees in JSON format
  ticketflow worktree list --format json

  # Clean up orphaned worktrees
  ticketflow worktree clean`)
}
