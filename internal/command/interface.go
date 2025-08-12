package command

import (
	"context"
	"flag"
)

// Command represents a CLI command that can be executed
type Command interface {
	// Name returns the command name (e.g., "new", "start", "list")
	Name() string

	// Aliases returns alternative names for this command (e.g., ["-v", "--version"] for version)
	Aliases() []string

	// Description returns a short description of what the command does
	Description() string

	// Usage returns usage information for the command
	Usage() string

	// SetupFlags configures the flag set for this command
	// Returns a struct that will hold the parsed flag values
	SetupFlags(fs *flag.FlagSet) interface{}

	// Validate checks if the provided flags and arguments are valid
	// flags is the struct returned by SetupFlags after parsing
	// args are the remaining command-line arguments after flag parsing
	Validate(flags interface{}, args []string) error

	// Execute runs the command with the given context
	// flags is the struct returned by SetupFlags after parsing
	// args are the remaining command-line arguments after flag parsing
	Execute(ctx context.Context, flags interface{}, args []string) error
}

// Metadata holds basic information about a command
type Metadata struct {
	Name        string
	Description string
	Usage       string
	Examples    []Example
}

// Example represents a usage example for a command
type Example struct {
	Description string
	Command     string
}

// Result represents the outcome of a command execution
type Result struct {
	Success bool
	Message string
	Data    interface{} // Optional data for JSON output
}

// Registry manages the collection of available commands
type Registry interface {
	// Register adds a command to the registry
	Register(cmd Command) error

	// Get retrieves a command by name
	Get(name string) (Command, bool)

	// List returns all registered commands
	List() []Command
}
