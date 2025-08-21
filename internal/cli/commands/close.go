package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/command"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// CloseTicketInternalError represents an internal error during ticket closing
type CloseTicketInternalError struct {
	IsCurrentTicket bool
	Message         string
}

// Error implements the error interface
func (e *CloseTicketInternalError) Error() string {
	return fmt.Sprintf("internal error: %s (isCurrentTicket=%v)", e.Message, e.IsCurrentTicket)
}

// CloseCommand implements the close command using the new Command interface
type CloseCommand struct{}

// NewCloseCommand creates a new close command
func NewCloseCommand() command.Command {
	return &CloseCommand{}
}

// Name returns the command name
func (c *CloseCommand) Name() string {
	return "close"
}

// Aliases returns alternative names for this command
func (c *CloseCommand) Aliases() []string {
	return nil
}

// Description returns a short description of the command
func (c *CloseCommand) Description() string {
	return "Close a ticket"
}

// Usage returns the usage string for the command
func (c *CloseCommand) Usage() string {
	return "close [--force] [--reason <message>] [--format text|json] [<ticket-id>]"
}

// closeFlags holds the flags for the close command
type closeFlags struct {
	force  BoolFlag
	reason string
	format StringFlag
	args   []string // Store validated arguments
}

// SetupFlags configures flags for the command
func (c *CloseCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	flags := &closeFlags{}
	RegisterBool(fs, &flags.force, "force", "f", "Force close even with uncommitted changes")
	fs.StringVar(&flags.reason, "reason", "", "Reason for closing the ticket")
	RegisterString(fs, &flags.format, "format", "o", FormatText, "Output format (text|json)")
	return flags
}

// Validate checks if the command arguments are valid
func (c *CloseCommand) Validate(flags interface{}, args []string) error {
	// Check for unexpected extra arguments
	if len(args) > 1 {
		return fmt.Errorf("unexpected arguments after ticket ID: %v", args[1:])
	}

	// Safely assert flags type
	f, ok := flags.(*closeFlags)
	if !ok {
		return fmt.Errorf("invalid flags type: expected *closeFlags, got %T", flags)
	}

	// Store arguments for Execute method
	f.args = args

	// Validate format flag using resolved value
	format := f.format.Value()
	if format != FormatText && format != FormatJSON {
		return fmt.Errorf("invalid format: %q (must be %q or %q)", format, FormatText, FormatJSON)
	}

	return nil
}

// Execute runs the close command
func (c *CloseCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	// Check for context cancellation at the start
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Safely assert flags type
	f, ok := flags.(*closeFlags)
	if !ok {
		return fmt.Errorf("invalid flags type: expected *closeFlags, got %T", flags)
	}

	// Get resolved values
	format := f.format.Value()
	force := f.force.Value()

	// Parse output format first
	outputFormat := cli.ParseOutputFormat(format)
	cli.SetGlobalOutputFormat(outputFormat) // Ensure errors are formatted correctly

	// Create App instance with format
	app, err := getAppWithFormat(ctx, outputFormat)
	if err != nil {
		return err
	}

	// Perform the close operation based on whether ticket ID was provided
	var closeErr error
	var closedTicket *ticket.Ticket
	var ticketID string
	var mode string
	var branch string

	if len(f.args) == 0 {
		// No ticket ID provided - close current ticket
		mode = "current"
		if f.reason != "" {
			closedTicket, closeErr = app.CloseTicketWithReason(ctx, f.reason, force)
		} else {
			closedTicket, closeErr = app.CloseTicket(ctx, force)
		}
	} else {
		// Ticket ID provided - close specific ticket
		mode = "by_id"
		ticketID = f.args[0]
		branch = ticketID // Branch name is usually the ticket ID
		closedTicket, closeErr = app.CloseTicketByID(ctx, ticketID, f.reason, force)
	}

	// Handle errors
	if closeErr != nil {
		if outputFormat == cli.FormatJSON {
			// For JSON, return structured error
			result := map[string]interface{}{
				"success": false,
				"error":   closeErr.Error(),
			}
			return app.Output.PrintResult(result)
		}
		return closeErr
	}

	// Calculate duration if current ticket mode
	var duration time.Duration
	if mode == "current" && closedTicket.StartedAt.Time != nil && closedTicket.ClosedAt.Time != nil {
		duration = closedTicket.ClosedAt.Time.Sub(*closedTicket.StartedAt.Time)
	}

	// Extract parent ticket
	var parentTicket string
	for _, rel := range closedTicket.Related {
		if strings.HasPrefix(rel, "parent:") {
			parentTicket = strings.TrimPrefix(rel, "parent:")
			break
		}
	}

	// Get worktree path for current mode
	var worktreePath string
	if mode == "current" {
		if cwd, err := os.Getwd(); err == nil {
			worktreePath = cwd
		}
	}

	// Create result
	result := &cli.CloseTicketResult{
		Ticket:        closedTicket,
		Mode:          mode,
		ForceUsed:     force,
		CommitCreated: true, // App methods always create commits on success
		CloseReason:   f.reason,
		Duration:      duration,
		ParentTicket:  parentTicket,
		WorktreePath:  worktreePath,
		Branch:        branch,
	}

	return app.Output.PrintResult(result)
}
