package main

import (
	"context"
	"flag"

	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/command"
)

// executeNewCommand executes commands that implement the new Command interface
func executeNewCommand(ctx context.Context, cmd command.Command, args []string) error {
	// Create flag set for this command
	// Use ContinueOnError to handle errors explicitly rather than exiting
	fs := flag.NewFlagSet(cmd.Name(), flag.ContinueOnError)

	// Setup command-specific flags
	cmdFlags := cmd.SetupFlags(fs)

	// Add logging flags (same as existing system)
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
	if err := cmd.Validate(cmdFlags, fs.Args()); err != nil {
		return err
	}

	// Execute command
	return cmd.Execute(ctx, cmdFlags, fs.Args())
}
