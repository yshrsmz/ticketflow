package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/yshrsmz/ticketflow/internal/cli"
)

// Command represents a CLI command with its configuration
type Command struct {
	Name       string
	MinArgs    int
	SetupFlags func(*flag.FlagSet) interface{}
	Validate   func(*flag.FlagSet, interface{}) error
	Execute    func(context.Context, *flag.FlagSet, interface{}) error
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
