// Package command provides a simple command interface for CLI applications.
//
// This package defines the Command interface and Registry for organizing
// CLI commands in a maintainable way. It's designed to help break up large
// command files into smaller, focused units.
//
// # Basic Usage
//
// Define a command by implementing the Command interface:
//
//	type NewTicketCommand struct {
//	    app *cli.App
//	}
//
//	func (c *NewTicketCommand) Name() string { return "new" }
//	func (c *NewTicketCommand) Description() string { return "Create a new ticket" }
//	func (c *NewTicketCommand) Usage() string { return "new [flags] <slug>" }
//
//	func (c *NewTicketCommand) SetupFlags(fs *flag.FlagSet) interface{} {
//	    flags := &newTicketFlags{}
//	    fs.IntVar(&flags.priority, "priority", 2, "ticket priority")
//	    return flags
//	}
//
//	func (c *NewTicketCommand) Validate(flags interface{}, args []string) error {
//	    if len(args) < 1 {
//	        return errors.New("ticket slug required")
//	    }
//	    return nil
//	}
//
//	func (c *NewTicketCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
//	    f := flags.(*newTicketFlags)
//	    slug := args[0]
//	    return c.app.NewTicket(ctx, slug, f.priority)
//	}
//
// # Registry Usage
//
// Use the Registry to manage multiple commands:
//
//	registry := command.NewRegistry()
//	registry.Register(&NewTicketCommand{app: app})
//	registry.Register(&ListTicketsCommand{app: app})
//	registry.Register(&StartTicketCommand{app: app})
//
//	// Get and execute a command
//	if cmd, ok := registry.Get("new"); ok {
//	    fs := flag.NewFlagSet(cmd.Name(), flag.ExitOnError)
//	    flags := cmd.SetupFlags(fs)
//	    fs.Parse(args)
//
//	    if err := cmd.Validate(flags, fs.Args()); err != nil {
//	        return err
//	    }
//
//	    return cmd.Execute(ctx, flags, fs.Args())
//	}
//
// # Design Philosophy
//
// This package prioritizes simplicity and practicality:
//   - No unnecessary abstractions or performance optimizations
//   - Clean separation of concerns (flags, validation, execution)
//   - Easy to test individual commands in isolation
//   - Supports context for cancellation of long-running operations
//   - Thread-safe registry for concurrent access
package command
