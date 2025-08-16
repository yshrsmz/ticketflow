package commands

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/yshrsmz/ticketflow/internal/command"
)

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

// workflowContent contains the comprehensive development workflow guide
const workflowContent = `# TicketFlow Development Workflow

## Overview
TicketFlow is a git worktree-based ticket management system that helps manage development tasks using a directory-based status tracking system (todo/doing/done).

## Development Workflow for New Features

### 1. Create a Feature Ticket
` + "```bash" + `
# Create a new ticket
ticketflow new my-feature

# Or create a sub-ticket with explicit parent
ticketflow new --parent parent-ticket-id my-sub-feature
` + "```" + `

**Note**: Flags must come before the ticket slug.

### 2. Start Work on the Ticket
` + "```bash" + `
# Start work (creates worktree automatically if enabled)
ticketflow start <ticket-id>
` + "```" + `

This command:
- Moves the ticket from todo/ to doing/
- Creates a git worktree for parallel development (if enabled)
- Checks out a new branch for your work

### 3. Navigate to the Worktree
` + "```bash" + `
# Navigate to the created worktree
cd ../ticketflow.worktrees/<ticket-id>
` + "```" + `

### 4. Make Your Changes
- Implement the feature/fix in the worktree
- All changes are isolated from the main repository
- You can switch between multiple tickets without stashing

### 5. Test Your Changes
` + "```bash" + `
# Run tests
make test

# Run specific test suites if available
make test-unit
make test-integration
` + "```" + `

### 6. Run Code Quality Checks
` + "```bash" + `
# Format code
make fmt

# Run static analysis
make vet

# Run linter
make lint
` + "```" + `

### 7. Commit and Push Changes
` + "```bash" + `
# Add and commit your changes
git add .
git commit -m "Implement feature X"

# Push to remote
git push -u origin <branch-name>
` + "```" + `

### 8. Create a Pull Request
` + "```bash" + `
# Create PR using GitHub CLI
gh pr create

# Or push and create via web interface
git push -u origin <branch>
` + "```" + `

### 9. Wait for Review and Approval
**IMPORTANT**: 
- Check if the ticket contains approval requirements
- If approval is required, DO NOT close the ticket until explicitly approved
- Wait for developer review and feedback

### 10. Close the Ticket (After Approval)
**CRITICAL**: Always close tickets from within the worktree directory!

` + "```bash" + `
# Make sure you're in the worktree directory
pwd  # Should show ../ticketflow.worktrees/<ticket-id>

# Close the ticket (creates close commit on feature branch)
ticketflow close

# Push the close commit
git push
` + "```" + `

**Common mistake to avoid**:
- Don't go back to the main repository to close the ticket
- This would create the close commit on the wrong branch

### 11. Clean Up After PR Merge
` + "```bash" + `
# After the PR is merged, clean up the worktree
ticketflow cleanup <ticket-id>
` + "```" + `

This removes the worktree and cleans up the branch.

## Important Commands

### Ticket Management
` + "```bash" + `
# List all tickets
ticketflow list

# Show ticket details
ticketflow show <ticket-id>

# Check current status
ticketflow status

# Restore a closed ticket
ticketflow restore <ticket-id>
` + "```" + `

### Worktree Management
` + "```bash" + `
# List all worktrees
ticketflow worktree list

# Clean orphaned worktrees
ticketflow worktree clean
` + "```" + `

## Configuration

TicketFlow is configured via .ticketflow.yaml in your repository root:

` + "```yaml" + `
# Example configuration
worktrees:
  enabled: true
  base_dir: ../ticketflow.worktrees
  init_commands:
    - npm install
    - make build

git:
  default_branch: main

tickets:
  directory: tickets
` + "```" + `

## Tips and Best Practices

1. **Always work in worktrees** - This keeps your main repository clean and allows parallel work
2. **Close tickets from the worktree** - Ensures the close commit is on the feature branch
3. **Use sub-tickets** - Break large features into smaller, manageable tickets with parent relationships
4. **Follow the workflow** - The structured approach helps maintain a clean git history
5. **Regular cleanup** - Use ` + "`ticketflow cleanup`" + ` after PRs are merged to keep worktrees organized

## Integration with AI Tools

This workflow output can be integrated with various AI coding assistants:

` + "```bash" + `
# Save to Claude configuration
ticketflow workflow > CLAUDE.md

# Append to Cursor rules
ticketflow workflow >> .cursorrules

# Save to any AI tool configuration
ticketflow workflow > .ai/instructions.md
` + "```" + `

## Getting Help

` + "```bash" + `
# Show help for any command
ticketflow help <command>

# Show general help
ticketflow help

# Show version information
ticketflow version
` + "```" + `
`
