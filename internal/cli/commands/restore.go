package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/command"
)

// RestoreCommand implements the restore command using the new Command interface
type RestoreCommand struct{}

// NewRestoreCommand creates a new restore command
func NewRestoreCommand() command.Command {
	return &RestoreCommand{}
}

// Name returns the command name
func (r *RestoreCommand) Name() string {
	return "restore"
}

// Aliases returns alternative names for this command
func (r *RestoreCommand) Aliases() []string {
	return nil
}

// Description returns a short description of the command
func (r *RestoreCommand) Description() string {
	return "Restore the current-ticket.md symlink in a worktree"
}

// Usage returns the usage string for the command
func (r *RestoreCommand) Usage() string {
	return "restore [--format text|json]"
}

// restoreFlags holds the flags for the restore command
type restoreFlags struct {
	format StringFlag
}

// SetupFlags configures the flag set for this command
func (r *RestoreCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	flags := &restoreFlags{}
	RegisterString(fs, &flags.format, "format", "o", FormatText, "Output format (text|json)")
	return flags
}

// Validate checks if the provided flags and arguments are valid
func (r *RestoreCommand) Validate(flags interface{}, args []string) error {
	// Defensive type assertion
	f, ok := flags.(*restoreFlags)
	if !ok {
		return fmt.Errorf("invalid flags type: expected *restoreFlags, got %T", flags)
	}

	// No arguments allowed for restore command
	if len(args) > 0 {
		return fmt.Errorf("restore command does not accept any arguments")
	}

	// Validate format value using resolved value
	format := f.format.Value()
	if format != FormatText && format != FormatJSON {
		return fmt.Errorf("invalid format: %q (must be %q or %q)", format, FormatText, FormatJSON)
	}

	return nil
}

// Execute runs the command with the given context
func (r *RestoreCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	f := flags.(*restoreFlags)

	// Get resolved format value
	format := f.format.Value()

	// Parse output format first
	outputFormat := cli.ParseOutputFormat(format)
	cli.SetGlobalOutputFormat(outputFormat) // Ensure errors are formatted correctly

	// Create App instance with format
	app, err := cli.NewAppWithFormat(ctx, outputFormat)
	if err != nil {
		return err
	}

	// Perform the restore operation - now returns the ticket directly
	ticket, err := app.RestoreCurrentTicket(ctx)

	if err != nil {
		if outputFormat == cli.FormatJSON {
			// Return error in JSON format
			result := map[string]interface{}{
				"error":   err.Error(),
				"success": false,
			}
			return app.Output.PrintResult(result)
		}
		return err
	}

	// Get worktree path
	var worktreePath string
	if cwd, err := os.Getwd(); err == nil {
		worktreePath = cwd
	}

	// Extract parent ticket
	var parentTicket string
	for _, rel := range ticket.Related {
		if strings.HasPrefix(rel, "parent:") {
			parentTicket = strings.TrimPrefix(rel, "parent:")
			break
		}
	}

	// Create result
	result := &cli.RestoreTicketResult{
		Ticket:       ticket,
		SymlinkPath:  "current-ticket.md",
		TargetPath:   fmt.Sprintf("tickets/doing/%s.md", ticket.ID),
		ParentTicket: parentTicket,
		WorktreePath: worktreePath,
	}

	return app.Output.PrintResult(result)
}
