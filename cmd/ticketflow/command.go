package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/yshrsmz/ticketflow/internal/cli"
)

// Command represents a CLI command with its configuration
type Command struct {
	// Name is the command name used for the flag set
	Name string

	// MinArgs is the minimum number of positional arguments required
	MinArgs int

	// MinArgsError is an optional custom error message for missing arguments.
	// If not provided, a generic message will be used.
	MinArgsError string

	// SetupFlags is an optional function to configure command-specific flags.
	// It should return a pointer to a struct containing the flag values.
	SetupFlags func(*flag.FlagSet) interface{}

	// Validate is an optional function for additional validation beyond MinArgs.
	// It receives the parsed flag set and the result from SetupFlags.
	Validate func(*flag.FlagSet, interface{}) error

	// Execute is the required function that implements the command logic.
	// It receives the context, parsed flag set, and the result from SetupFlags.
	Execute func(context.Context, *flag.FlagSet, interface{}) error
}

// parseAndExecute handles the common pattern of parsing flags, configuring logging, and executing a command
func parseAndExecute(ctx context.Context, cmd Command, args []string) error {
	// Create flag set
	fs := flag.NewFlagSet(cmd.Name, flag.ExitOnError)

	// Setup command-specific flags
	var cmdFlags interface{}
	if cmd.SetupFlags != nil {
		cmdFlags = cmd.SetupFlags(fs)
	}

	// Add logging flags
	loggingOpts := cli.AddLoggingFlags(fs)

	// Parse flags
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Configure logging
	if err := cli.ConfigureLogging(loggingOpts); err != nil {
		return err
	}

	// Validate arguments
	if cmd.MinArgs > 0 && fs.NArg() < cmd.MinArgs {
		if cmd.MinArgsError != "" {
			return fmt.Errorf("%s", cmd.MinArgsError)
		}
		return fmt.Errorf("missing required arguments for %s command", cmd.Name)
	}

	// Additional validation if provided
	if cmd.Validate != nil {
		if err := cmd.Validate(fs, cmdFlags); err != nil {
			return err
		}
	}

	// Execute command
	return cmd.Execute(ctx, fs, cmdFlags)
}
