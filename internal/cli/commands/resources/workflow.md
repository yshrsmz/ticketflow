# TicketFlow Workflow Guide

## Overview
TicketFlow is a git worktree-based ticket management system that helps manage development tasks using a directory-based status tracking system (todo/doing/done).

This guide shows how to use ticketflow to manage tickets and development workflow in your project.

## Workflow for Managing Tasks

### 1. Create a Feature Ticket
```bash
# Create a new ticket
ticketflow new my-feature

# Or create a sub-ticket with explicit parent
ticketflow new --parent parent-ticket-id my-sub-feature
```

**Note**: Flags must come before the ticket slug.

### 2. Start Work on the Ticket
```bash
# Start work (creates worktree automatically if enabled)
ticketflow start <ticket-id>
```

This command:
- Moves the ticket from todo/ to doing/
- Creates a git worktree for parallel development (if enabled)
- Checks out a new branch for your work

### 3. Navigate to the Worktree
```bash
# Navigate to the created worktree
cd ../ticketflow.worktrees/<ticket-id>
```

### 4. Make Your Changes
- Implement the feature/fix in the worktree
- All changes are isolated from the main repository
- You can switch between multiple tickets without stashing

### 5. Test Your Changes
```bash
# Run your project's tests
# The exact commands depend on your project type:
#   Node.js: npm test, yarn test
#   Go: go test ./..., make test
#   Rust: cargo test
#   Python: pytest, python -m unittest
#   Java: mvn test, gradle test

<your-test-command>
```

### 6. Run Code Quality Checks
```bash
# Run your project's code quality tools
# The exact commands depend on your project:
#   Node.js: npm run lint, prettier --write
#   Go: go fmt ./..., go vet ./..., golangci-lint run
#   Rust: cargo fmt, cargo clippy
#   Python: black ., flake8, mypy
#   Java: checkstyle, spotbugs

<your-lint-command>
<your-format-command>
```

### 7. Commit and Push Changes
```bash
# Add and commit your changes
git add .
git commit -m "Implement feature X"

# Push to remote
git push -u origin <branch-name>
```

### 8. Create a Pull Request
```bash
# Create PR using GitHub CLI
gh pr create

# Or push and create via web interface
git push -u origin <branch>
```

### 9. Wait for Review and Approval
**IMPORTANT**: 
- Check if the ticket contains approval requirements
- If approval is required, DO NOT close the ticket until explicitly approved
- Wait for developer review and feedback

### 10. Close the Ticket (After Approval)
**CRITICAL**: Always close tickets from within the worktree directory!

```bash
# Make sure you're in the worktree directory
pwd  # Should show ../ticketflow.worktrees/<ticket-id>

# Close the ticket (creates close commit on feature branch)
ticketflow close

# Push the close commit
git push
```

**Common mistake to avoid**:
- Don't go back to the main repository to close the ticket
- This would create the close commit on the wrong branch

### 11. Clean Up After PR Merge
```bash
# After the PR is merged, clean up the worktree
ticketflow cleanup <ticket-id>
```

This removes the worktree and cleans up the branch.

## Important Commands

### Ticket Management
```bash
# List all tickets
ticketflow list

# Show ticket details
ticketflow show <ticket-id>

# Check current status
ticketflow status

# Restore a closed ticket
ticketflow restore <ticket-id>
```

### Worktree Management
```bash
# List all worktrees
ticketflow worktree list

# Clean orphaned worktrees
ticketflow worktree clean
```

## Configuration

TicketFlow is configured via .ticketflow.yaml in your repository root:

```yaml
# Example configuration
worktrees:
  enabled: true
  base_dir: ../<project>.worktrees
  # Configure project-specific setup commands
  init_commands:
    # Node.js projects:
    - npm install
    # Go projects:
    - go mod download
    # Rust projects:
    - cargo build
    # Python projects:
    - pip install -r requirements.txt
    # Or any custom setup:
    - ./scripts/setup.sh

git:
  default_branch: main

tickets:
  directory: tickets
```

### Project-Specific Commands

The `init_commands` are executed automatically when you start work on a ticket. Configure these based on your project's needs:

- **Node.js**: `npm install`, `npm run build`
- **Go**: `go mod download`, `go build ./...`
- **Rust**: `cargo fetch`, `cargo build`
- **Python**: `pip install -r requirements.txt`, `python setup.py develop`
- **Ruby**: `bundle install`
- **Java**: `mvn install`, `gradle build`

## Tips and Best Practices

1. **Always work in worktrees** - This keeps your main repository clean and allows parallel work
2. **Close tickets from the worktree** - Ensures the close commit is on the feature branch
3. **Use sub-tickets** - Break large features into smaller, manageable tickets with parent relationships
4. **Follow the workflow** - The structured approach helps maintain a clean git history
5. **Regular cleanup** - Use `ticketflow cleanup` after PRs are merged to keep worktrees organized

## Integration with Development Tools

### AI Coding Assistants

Export this workflow guide to help AI assistants understand your project workflow:

```bash
# Save to Claude configuration
ticketflow workflow > CLAUDE.md

# Append to Cursor rules
ticketflow workflow >> .cursorrules

# Save to any AI tool configuration
ticketflow workflow > .ai/instructions.md
```

### Custom Scripts

You can integrate ticketflow with your project's build system:

```bash
# In package.json scripts:
"scripts": {
  "ticket:new": "ticketflow new",
  "ticket:start": "ticketflow start",
  "ticket:status": "ticketflow status"
}

# In Makefile:
ticket-new:
	ticketflow new $(ARGS)

ticket-start:
	ticketflow start $(ID)
```

## Getting Help

```bash
# Show help for any command
ticketflow help <command>

# Show general help
ticketflow help

# Show version information
ticketflow version
```