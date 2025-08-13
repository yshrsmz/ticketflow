package commands

import (
	"context"
	"flag"
	"fmt"

	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/command"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// Default value for count flag
const defaultCount = 20

// ListCommand implements the list command using the new Command interface
type ListCommand struct{}

// NewListCommand creates a new list command
func NewListCommand() command.Command {
	return &ListCommand{}
}

// Name returns the command name
func (c *ListCommand) Name() string {
	return "list"
}

// Aliases returns alternative names for this command
func (c *ListCommand) Aliases() []string {
	return []string{"ls"}
}

// Description returns a short description of the command
func (c *ListCommand) Description() string {
	return "List tickets"
}

// Usage returns the usage string for the command
func (c *ListCommand) Usage() string {
	return "list [--status todo|doing|done|all] [--count N] [--format text|json]"
}

// listFlags holds the flags for the list command
type listFlags struct {
	status      string
	statusShort string
	count       int
	countShort  int
	format      string
}

// SetupFlags configures flags for the command
func (c *ListCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	flags := &listFlags{}
	fs.StringVar(&flags.status, "status", "", "Filter by status (todo|doing|done|all)")
	fs.StringVar(&flags.statusShort, "s", "", "Filter by status (todo|doing|done|all)")
	fs.IntVar(&flags.count, "count", defaultCount, "Number of tickets to show")
	fs.IntVar(&flags.countShort, "c", defaultCount, "Number of tickets to show")
	fs.StringVar(&flags.format, "format", "text", "Output format (text|json)")
	return flags
}

// Validate checks if the command arguments are valid
func (c *ListCommand) Validate(flags interface{}, args []string) error {
	// Extract flags
	f := flags.(*listFlags)

	// Merge short and long form flags (short takes precedence if both provided)
	if f.statusShort != "" {
		f.status = f.statusShort
	}
	if f.countShort != defaultCount {
		f.count = f.countShort
	}

	// Validate format flag
	if f.format != "text" && f.format != "json" {
		return fmt.Errorf("invalid format: %q (must be 'text' or 'json')", f.format)
	}

	// Validate count flag
	if f.count < 0 {
		return fmt.Errorf("count must be non-negative, got %d", f.count)
	}

	// Validate status flag if provided
	if f.status != "" && !isValidListStatus(f.status) {
		return fmt.Errorf("invalid status: %q (must be 'todo', 'doing', 'done', or 'all')", f.status)
	}

	return nil
}

// Execute runs the list command
func (c *ListCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	// Create App instance with dependencies
	app, err := cli.NewApp(ctx)
	if err != nil {
		return err
	}

	// Extract flags
	f := flags.(*listFlags)

	// Convert status string to ticket.Status
	var ticketStatus ticket.Status
	if f.status != "" {
		ticketStatus = ticket.Status(f.status)
	}

	outputFormat := cli.ParseOutputFormat(f.format)

	// Delegate to App's ListTickets method
	return app.ListTickets(ctx, ticketStatus, f.count, outputFormat)
}

// isValidListStatus checks if the status is valid for list command
func isValidListStatus(status string) bool {
	switch status {
	case "", string(ticket.StatusTodo), string(ticket.StatusDoing), string(ticket.StatusDone), "all":
		return true
	default:
		return false
	}
}
