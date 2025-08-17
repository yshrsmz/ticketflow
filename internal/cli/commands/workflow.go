package commands

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/yshrsmz/ticketflow/internal/command"
)

//go:embed resources/workflow.md
var workflowContent string

// WorkflowCommand implements the workflow command using the Command interface
type WorkflowCommand struct {
	output io.Writer
}

// NewWorkflowCommand creates a new workflow command
func NewWorkflowCommand() command.Command {
	return &WorkflowCommand{
		output: os.Stdout,
	}
}

// Name returns the command name
func (c *WorkflowCommand) Name() string {
	return "workflow"
}

// Aliases returns alternative names for this command
func (c *WorkflowCommand) Aliases() []string {
	return nil
}

// Description returns a short description of the command
func (c *WorkflowCommand) Description() string {
	return "Print development workflow guide"
}

// Usage returns the usage string for the command
func (c *WorkflowCommand) Usage() string {
	return "workflow"
}

// SetupFlags configures flags for the command
func (c *WorkflowCommand) SetupFlags(fs *flag.FlagSet) interface{} {
	// Workflow command has no flags
	return nil
}

// Validate checks if the command arguments are valid
func (c *WorkflowCommand) Validate(flags interface{}, args []string) error {
	// No validation needed for workflow command
	return nil
}

// Execute runs the workflow command
func (c *WorkflowCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Print the workflow content to the output writer
	if _, err := fmt.Fprint(c.output, workflowContent); err != nil {
		return fmt.Errorf("failed to write workflow content: %w", err)
	}
	return nil
}
