package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/command"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

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
	force       bool
	forceShort  bool
	reason      string
	format      string
	formatShort string
	args        []string // Store validated arguments
}

// normalize merges short and long form flags using logical OR for booleans
// and preferring non-empty strings for string flags
func (f *closeFlags) normalize() {
	// Use logical OR for boolean flags - true if either is set
	f.force = f.force || f.forceShort

	// For string flags, prefer non-empty value (short form if both set)
	if f.formatShort != "" {
		f.format = f.formatShort
	}
}

// SetupFlags configures flags for the command
func (c *CloseCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	flags := &closeFlags{}
	// Long forms
	fs.BoolVar(&flags.force, "force", false, "Force close even with uncommitted changes")
	fs.StringVar(&flags.reason, "reason", "", "Reason for closing the ticket")
	fs.StringVar(&flags.format, "format", FormatText, "Output format (text|json)")
	// Short forms
	fs.BoolVar(&flags.forceShort, "f", false, "Force close (short form)")
	fs.StringVar(&flags.formatShort, "o", FormatText, "Output format (short form)")
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

	// Merge short and long forms (short form takes precedence if both provided)
	f.normalize()

	// Validate format flag
	if f.format != FormatText && f.format != FormatJSON {
		return fmt.Errorf("invalid format: %q (must be %q or %q)", f.format, FormatText, FormatJSON)
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

	// Get app instance
	app, err := cli.NewApp(ctx)
	if err != nil {
		return err
	}

	// Safely assert flags type
	f, ok := flags.(*closeFlags)
	if !ok {
		return fmt.Errorf("invalid flags type: expected *closeFlags, got %T", flags)
	}

	// Perform the close operation based on whether ticket ID was provided
	var closeErr error
	var closedTicket *ticket.Ticket
	var ticketID string

	if len(f.args) == 0 {
		// No ticket ID provided - close current ticket
		if f.reason != "" {
			closedTicket, closeErr = app.CloseTicketWithReason(ctx, f.reason, f.force)
		} else {
			closedTicket, closeErr = app.CloseTicket(ctx, f.force)
		}
	} else {
		// Ticket ID provided - close specific ticket
		ticketID = f.args[0]
		closedTicket, closeErr = app.CloseTicketByID(ctx, ticketID, f.reason, f.force)
	}

	// Handle errors based on format
	if closeErr != nil {
		if f.format == FormatJSON {
			return outputCloseErrorJSON(app, closeErr)
		}
		return closeErr
	}

	// Handle success output based on format
	if f.format == FormatJSON {
		return outputCloseSuccessJSON(ctx, app, closedTicket, f.reason, f.force, len(f.args) == 0)
	}

	// For text format, the App methods already print success messages
	return nil
}

// outputCloseErrorJSON outputs an error in JSON format
func outputCloseErrorJSON(app *cli.App, err error) error {
	output := map[string]interface{}{
		"success": false,
		"error":   err.Error(),
	}
	return app.Output.PrintJSON(output)
}

// outputCloseSuccessJSON outputs success information in JSON format
func outputCloseSuccessJSON(ctx context.Context, app *cli.App, closedTicket *ticket.Ticket, reason string, force bool, isCurrentTicket bool) error {
	// Build the output structure
	output := map[string]interface{}{
		"success":        true,
		"force_used":     force,
		"commit_created": true, // App methods always create commits on success
	}

	// Set mode based on whether current ticket or by ID
	if isCurrentTicket {
		output["mode"] = "current"
	} else {
		output["mode"] = "by_id"
	}

	// Use the returned ticket information directly
	if closedTicket != nil {
		output["ticket_id"] = closedTicket.ID
		output["status"] = string(closedTicket.Status())

		if closedTicket.ClosedAt.Time != nil {
			output["closed_at"] = closedTicket.ClosedAt.Time.Format(time.RFC3339)
		}

		// Calculate duration for current ticket mode
		if isCurrentTicket && closedTicket.StartedAt.Time != nil && closedTicket.ClosedAt.Time != nil {
			duration := closedTicket.ClosedAt.Time.Sub(*closedTicket.StartedAt.Time)
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60
			output["duration"] = fmt.Sprintf("%dh%dm", hours, minutes)
		}

		// Add parent ticket if available - extract from Related field
		for _, rel := range closedTicket.Related {
			if strings.HasPrefix(rel, "parent:") {
				output["parent_ticket"] = strings.TrimPrefix(rel, "parent:")
				break
			}
		}

		// Add worktree path for current ticket
		if isCurrentTicket {
			cwd, err := os.Getwd()
			if err == nil {
				output["worktree_path"] = cwd
			}
		} else {
			// For by-ID mode, include the branch name
			output["branch"] = closedTicket.ID
		}
	}

	if reason != "" {
		output["close_reason"] = reason
	}

	return app.Output.PrintJSON(output)
}
