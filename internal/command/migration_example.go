// Package command migration example - showing how to migrate from the current system
// This file demonstrates the migration path and can be deleted after migration is complete

package command

import (
	"context"
	"flag"
	"fmt"
	// Note: This import would be needed for real implementation
	// "github.com/yshrsmz/ticketflow/internal/cli"
)

// Example 1: Migrate the "version" command (simplest case)
type VersionCommand struct {
	Version   string
	GitCommit string
	BuildTime string
}

func (c *VersionCommand) Name() string        { return "version" }
func (c *VersionCommand) Aliases() []string   { return []string{"-v", "--version"} }
func (c *VersionCommand) Description() string { return "Show version information" }
func (c *VersionCommand) Usage() string       { return "version" }

func (c *VersionCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	// No flags for version command
	return nil
}

func (c *VersionCommand) Validate(flags interface{}, args []string) error {
	// No validation needed
	return nil
}

func (c *VersionCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	fmt.Printf("ticketflow version %s\n", c.Version)
	if c.Version != "dev" || c.GitCommit != "unknown" {
		fmt.Printf("  Git commit: %s\n", c.GitCommit)
		fmt.Printf("  Built at:   %s\n", c.BuildTime)
	}
	return nil
}

// Example 2: Migrate the "list" command (with flags)
type ListCommand struct {
	// app would be used in real implementation
	// app *cli.App
}

type listFlags struct {
	status string
	count  int
	format string
}

func (c *ListCommand) Name() string        { return "list" }
func (c *ListCommand) Aliases() []string   { return nil }
func (c *ListCommand) Description() string { return "List tickets" }
func (c *ListCommand) Usage() string       { return "list [flags]" }

func (c *ListCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	flags := &listFlags{}
	fs.StringVar(&flags.status, "status", "", "Filter by status (todo|doing|done)")
	fs.IntVar(&flags.count, "count", 20, "Maximum number of tickets to show")
	fs.StringVar(&flags.format, "format", "text", "Output format (text|json)")
	return flags
}

func (c *ListCommand) Validate(flags interface{}, args []string) error {
	f := flags.(*listFlags)
	if f.status != "" && f.status != "todo" && f.status != "doing" && f.status != "done" {
		return fmt.Errorf("invalid status: %s", f.status)
	}
	if f.format != "text" && f.format != "json" {
		return fmt.Errorf("invalid format: %s", f.format)
	}
	return nil
}

func (c *ListCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	_ = flags.(*listFlags) // Would be used in real implementation
	// This would call the existing handleList function or inline its logic
	// NOTE: In real implementation, you'd need to convert string to proper types
	// This is just an example showing the pattern
	// return c.app.ListTickets(ctx, ticket.Status(f.status), f.count, cli.OutputFormat(f.format))
	return nil // Example implementation
}

// Example 3: Migrate the "new" command (with required arguments)
type NewCommand struct {
	// app would be used in real implementation
	// app *cli.App
}

type newFlags struct {
	parent      string
	parentShort string
	format      string
}

func (c *NewCommand) Name() string        { return "new" }
func (c *NewCommand) Aliases() []string   { return nil }
func (c *NewCommand) Description() string { return "Create a new ticket" }
func (c *NewCommand) Usage() string       { return "new [flags] <slug>" }

func (c *NewCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	flags := &newFlags{}
	fs.StringVar(&flags.parent, "parent", "", "Parent ticket ID")
	fs.StringVar(&flags.parentShort, "p", "", "Parent ticket ID (short form)")
	fs.StringVar(&flags.format, "format", "text", "Output format (text|json)")
	return flags
}

func (c *NewCommand) Validate(flags interface{}, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing slug argument")
	}
	// Could add slug validation here
	return nil
}

func (c *NewCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	_ = flags.(*newFlags) // Would be used in real implementation
	_ = args[0]           // slug - Would be used in real implementation
	// This would call the existing handleNew function or inline its logic
	// NOTE: In real implementation, you'd need to convert string to proper types
	// return c.app.NewTicket(ctx, slug, parent, cli.OutputFormat(f.format))
	return nil // Example implementation
}

// Migration Bridge - adapts new Command interface to old Command struct
type CommandAdapter struct {
	cmd Command
}

// Adapt converts a new Command to work with the existing parseAndExecute function
func (a *CommandAdapter) ToLegacyCommand() interface{} {
	// This creates a struct compatible with the existing Command type in command.go
	type legacyCommand struct {
		Name         string
		MinArgs      int
		MinArgsError string
		SetupFlags   func(*flag.FlagSet) interface{}
		Validate     func(*flag.FlagSet, interface{}) error
		Execute      func(context.Context, *flag.FlagSet, interface{}) error
	}

	legacy := legacyCommand{
		Name: a.cmd.Name(),
	}

	// Adapt SetupFlags
	legacy.SetupFlags = a.cmd.SetupFlags

	// Adapt validation - merge Validate method with existing patterns
	legacy.Validate = func(fs *flag.FlagSet, flags interface{}) error {
		return a.cmd.Validate(flags, fs.Args())
	}

	// Adapt Execute
	legacy.Execute = func(ctx context.Context, fs *flag.FlagSet, flags interface{}) error {
		return a.cmd.Execute(ctx, flags, fs.Args())
	}

	return legacy
}

// MigrationStep shows how to incrementally migrate commands
func MigrationStep1_UseRegistryForDispatch() {
	// Step 1: Create registry and register commands
	registry := NewRegistry()

	// Register migrated commands
	if err := registry.Register(&VersionCommand{
		Version:   "1.0.0",
		GitCommit: "abc123",
		BuildTime: "2024-01-01",
	}); err != nil {
		// Handle error in real implementation
		_ = err
	}

	// In main.go, replace the switch statement with:
	/*
		if cmd, ok := registry.Get(os.Args[1]); ok {
			// Use the new command
			return executeCommand(ctx, cmd, os.Args[2:])
		} else {
			// Fall back to old switch for unmigrated commands
			switch os.Args[1] {
			case "list":
				// old implementation
			// ... other unmigrated commands
			}
		}
	*/
}

func MigrationStep2_MoveCommandsToSeparateFiles() {
	// Step 2: Move each command to its own file
	// internal/cli/commands/version.go
	// internal/cli/commands/list.go
	// internal/cli/commands/new.go
	// etc.

	// Each file would have an init() function that registers itself:
	/*
		func init() {
			DefaultRegistry.Register(&VersionCommand{...})
		}
	*/
}

func MigrationStep3_RemoveOldCommandStruct() {
	// Step 3: Once all commands are migrated:
	// - Remove the old Command struct from command.go
	// - Remove parseAndExecute function
	// - Replace with a simple executeCommand function that works with the new interface
	// - Remove the large switch statement from main.go
}

// The executeCommand function would be the new unified command executor
// Example implementation (after migration is complete):
//
// func executeCommand(ctx context.Context, cmd Command, args []string) error {
//     fs := flag.NewFlagSet(cmd.Name(), flag.ExitOnError)
//     flags := cmd.SetupFlags(fs)
//     loggingOpts := cli.AddLoggingFlags(fs)
//
//     if err := fs.Parse(args); err != nil {
//         return err
//     }
//
//     if err := cli.ConfigureLogging(loggingOpts); err != nil {
//         return err
//     }
//
//     if err := cmd.Validate(flags, fs.Args()); err != nil {
//         return err
//     }
//
//     return cmd.Execute(ctx, flags, fs.Args())
// }
