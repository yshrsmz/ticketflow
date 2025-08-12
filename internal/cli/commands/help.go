package commands

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/yshrsmz/ticketflow/internal/command"
)

// HelpCommand implements the help command using the new Command interface
type HelpCommand struct {
	registry command.Registry
	version  string
}

// NewHelpCommand creates a new help command with the registry
func NewHelpCommand(registry command.Registry, version string) command.Command {
	return &HelpCommand{
		registry: registry,
		version:  version,
	}
}

// Name returns the command name
func (c *HelpCommand) Name() string {
	return "help"
}

// Aliases returns alternative names for this command
func (c *HelpCommand) Aliases() []string {
	return []string{"-h", "--help"}
}

// Description returns a short description of the command
func (c *HelpCommand) Description() string {
	return "Show help information"
}

// Usage returns the usage string for the command
func (c *HelpCommand) Usage() string {
	return "help [command]"
}

// SetupFlags configures flags for the command
func (c *HelpCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	// Help command has no flags
	return nil
}

// Validate checks if the command arguments are valid
func (c *HelpCommand) Validate(flags interface{}, args []string) error {
	// No validation needed for help command
	return nil
}

// Execute runs the help command
func (c *HelpCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	// If a specific command is requested, show its help
	if len(args) > 0 {
		return c.showCommandHelp(args[0])
	}

	// Show general help
	return c.showGeneralHelp()
}

// showGeneralHelp displays the general help message
func (c *HelpCommand) showGeneralHelp() error {
	// Handle version string that might already have 'v' prefix
	versionStr := c.version
	if !strings.HasPrefix(versionStr, "v") {
		versionStr = "v" + versionStr
	}
	fmt.Printf("TicketFlow - Git worktree-based ticket management system (%s)\n", versionStr)
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  ticketflow                          Start TUI (interactive mode)")
	fmt.Println("  ticketflow <command> [options]      Run a command")
	fmt.Println()

	// Get migrated commands from registry
	migratedCommands := c.getMigratedCommands()

	// Define unmigrated commands in order
	// TODO: Remove this hardcoded list once all commands are migrated
	unmigratedCommands := []struct {
		usage       string
		description string
	}{
		{"new <slug> [options]", "Create new ticket"},
		{"list [options]", "List tickets"},
		{"show <ticket> [options]", "Show ticket details"},
		{"start <ticket> [options]", "Start working on ticket"},
		{"close [ticket] [options]", "Close current or specific ticket"},
		{"restore", "Restore current-ticket link"},
		{"status [options]", "Show current status"},
		{"worktree <command>", "Manage worktrees"},
		{"cleanup [options] <ticket>", "Clean up after PR merge"},
		{"cleanup [options]", "Auto-cleanup orphaned worktrees and stale branches"},
		{"migrate [options]", "Migrate ticket dates to new format"},
	}

	fmt.Println("COMMANDS:")

	// Show unmigrated commands
	for _, cmd := range unmigratedCommands {
		fmt.Printf("  %-30s %s\n", cmd.usage, cmd.description)
	}

	// Show migrated commands
	for _, cmd := range migratedCommands {
		usage := cmd.Name()
		if cmd.Usage() != "" && cmd.Usage() != cmd.Name() {
			usage = cmd.Usage()
		}
		fmt.Printf("  %-30s %s\n", usage, cmd.Description())
	}

	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  All commands support logging options:")
	fmt.Println("    --log-level LEVEL   Log level (debug, info, warn, error)")
	fmt.Println("    --log-format FORMAT Log format (text, json)")
	fmt.Println("    --log-output OUTPUT Log output (stderr, stdout, or file path)")
	fmt.Println()
	fmt.Println("  new:")
	fmt.Println("    --parent TICKET    Specify parent ticket ID")
	fmt.Println("    -p TICKET          Short form of --parent")
	fmt.Println("    --format FORMAT    Output format: text|json (default: text)")
	fmt.Println()
	fmt.Println("  list:")
	fmt.Println("    --status STATUS    Filter by status (todo|doing|done)")
	fmt.Println("    --count N          Maximum number of tickets to show (default: 20)")
	fmt.Println("    --format FORMAT    Output format: text|json (default: text)")
	fmt.Println()
	fmt.Println("  show:")
	fmt.Println("    --format FORMAT    Output format: text|json (default: text)")
	fmt.Println()
	fmt.Println("  start:")
	fmt.Println("    --force            Force recreate worktree if it already exists")
	fmt.Println("    --format FORMAT    Output format: text|json (default: text)")
	fmt.Println()
	fmt.Println("  close:")
	fmt.Println("    --format FORMAT    Output format: text|json (default: text)")
	fmt.Println()
	fmt.Println("  status:")
	fmt.Println("    --format FORMAT    Output format: text|json (default: text)")
	fmt.Println()
	fmt.Println("  worktree:")
	fmt.Println("    remove <ticket>    Remove worktree for a specific ticket")
	fmt.Println("    list               List all worktrees")
	fmt.Println()
	fmt.Println("  cleanup:")
	fmt.Println("    --dry-run          Preview cleanup without making changes")
	fmt.Println("    --format FORMAT    Output format: text|json (default: text)")
	fmt.Println()
	fmt.Println("  migrate:")
	fmt.Println("    --force            Force migration even if already migrated")
	fmt.Println("    --format FORMAT    Output format: text|json (default: text)")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  ticketflow new feature-xyz --parent TASK-123")
	fmt.Println("  ticketflow list --status doing")
	fmt.Println("  ticketflow start feature-xyz")
	fmt.Println("  ticketflow close")
	fmt.Println()
	fmt.Println("For more information about a command, use:")
	fmt.Println("  ticketflow help <command>")

	return nil
}

// getMigratedCommands returns a sorted list of migrated commands from the registry
func (c *HelpCommand) getMigratedCommands() []command.Command {
	if c.registry == nil {
		return nil
	}

	commands := c.registry.List()

	// Sort commands by name for consistent output
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name() < commands[j].Name()
	})

	return commands
}

// showCommandHelp displays help for a specific command
func (c *HelpCommand) showCommandHelp(cmdName string) error {
	// Try to get the command from the registry
	if c.registry != nil {
		if cmd, ok := c.registry.Get(cmdName); ok {
			fmt.Printf("Command: %s\n", cmd.Name())
			fmt.Printf("Description: %s\n", cmd.Description())
			fmt.Printf("Usage: ticketflow %s\n", cmd.Usage())

			if aliases := cmd.Aliases(); len(aliases) > 0 {
				fmt.Printf("Aliases: %s\n", strings.Join(aliases, ", "))
			}

			return nil
		}
	}

	// For unmigrated commands, show a simple message
	switch cmdName {
	case "init", "new", "list", "show", "start", "close", "restore",
		"status", "worktree", "cleanup", "migrate":
		fmt.Printf("Command: %s\n", cmdName)
		fmt.Println("Use 'ticketflow help' to see available options for this command.")
		return nil
	default:
		return fmt.Errorf("unknown command: %s", cmdName)
	}
}
