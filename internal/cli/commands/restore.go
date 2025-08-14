package commands

import (
	"context"
	"encoding/json"
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
	format      string
	formatShort string
}

// SetupFlags configures the flag set for this command
func (r *RestoreCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	flags := &restoreFlags{}

	// Output format flags
	fs.StringVar(&flags.format, "format", FormatText, "Output format (text|json)")
	fs.StringVar(&flags.formatShort, "o", FormatText, "Output format (short form)")

	return flags
}

// Validate checks if the provided flags and arguments are valid
func (r *RestoreCommand) Validate(flags interface{}, args []string) error {
	f := flags.(*restoreFlags)

	// No arguments allowed for restore command
	if len(args) > 0 {
		return fmt.Errorf("restore command does not accept any arguments")
	}

	// Normalize format flag (prefer short form if both are set)
	if f.formatShort != FormatText {
		f.format = f.formatShort
	}

	// Validate format value
	if f.format != FormatText && f.format != FormatJSON {
		return fmt.Errorf("invalid format: %q (must be %q or %q)", f.format, FormatText, FormatJSON)
	}

	return nil
}

// Execute runs the command with the given context
func (r *RestoreCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	f := flags.(*restoreFlags)

	// Get App instance
	app, err := cli.NewApp(ctx)
	if err != nil {
		return err
	}

	// Perform the restore operation - now returns the ticket directly
	ticket, err := app.RestoreCurrentTicket(ctx)

	if err != nil {
		if f.format == FormatJSON {
			// Return error in JSON format
			return outputJSON(map[string]interface{}{
				"error":   err.Error(),
				"success": false,
			})
		}
		return err
	}

	// For JSON output, use the returned ticket directly
	if f.format == FormatJSON {
		// Get the current working directory for worktree path
		cwd, _ := os.Getwd()

		jsonData := map[string]interface{}{
			"ticket_id":        ticket.ID,
			"status":           string(ticket.Status()),
			"symlink_restored": true,
			"symlink_path":     "current-ticket.md",
			"target_path":      fmt.Sprintf("tickets/doing/%s.md", ticket.ID),
			"worktree_path":    cwd,
			"message":          "Current ticket symlink restored",
			"success":          true,
		}

		// Extract parent ticket ID from Related field if available
		for _, rel := range ticket.Related {
			if strings.HasPrefix(rel, "parent:") {
				jsonData["parent_ticket"] = strings.TrimPrefix(rel, "parent:")
				break
			}
		}

		return outputJSON(jsonData)
	}

	// Text output
	fmt.Println("✅ Current ticket symlink restored")
	return nil
}

// outputJSON formats and outputs data as JSON
func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

