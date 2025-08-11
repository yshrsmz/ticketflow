# TicketFlow

A git worktree-based ticket management system optimized for AI-assisted development.

## Overview

TicketFlow is a modern ticket management system that integrates seamlessly with Git workflows. It uses git worktrees to enable parallel work on multiple tickets while maintaining a clean project structure. Designed with both human developers and AI assistants in mind, it provides both an intuitive TUI (Terminal User Interface) and a comprehensive CLI.

### Core Features:

**Ticket Management:**
- ✅ Directory-based status tracking (todo/doing/done)
- ✅ Markdown files with YAML frontmatter
- ✅ Sub-ticket support with parent relationships
- ✅ Current ticket symlink tracking
- ✅ JSON output support for AI integration

**Git Integration:**
- ✅ Git worktree creation for isolated development
- ✅ Branch management per ticket
- ✅ Manual control over all Git operations (no auto-push/merge)
- ✅ Worktree persistence after ticket closure
- ✅ Post-PR merge cleanup commands

**Terminal UI (TUI):**
- ✅ Interactive interface with Bubble Tea framework
- ✅ Tab-based navigation (ALL/TODO/DOING/DONE)
- ✅ Real-time search functionality
- ✅ Ticket creation, editing, and management
- ✅ Worktree management view
- ✅ Keyboard shortcuts with help overlay
- ✅ Start/close tickets directly from TUI

### Project Structure:
```
ticketflow/
├── cmd/ticketflow/          # CLI entry point
├── internal/
│   ├── config/             # Configuration management
│   ├── ticket/             # Ticket model and manager
│   ├── git/                # Git operations
│   ├── cli/                # CLI commands and output
│   └── ui/                 # TUI implementation
│       ├── components/     # Reusable UI components
│       ├── styles/         # Styling and themes
│       └── views/          # View implementations
└── test/                   # Integration tests
```

## Installation

### Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/yshrsmz/ticketflow/releases).

**Linux (AMD64)**
```bash
curl -L https://github.com/yshrsmz/ticketflow/releases/latest/download/ticketflow-linux-amd64.tar.gz | tar xz
sudo mv ticketflow /usr/local/bin/
```

**Linux (ARM64)**
```bash
curl -L https://github.com/yshrsmz/ticketflow/releases/latest/download/ticketflow-linux-arm64.tar.gz | tar xz
sudo mv ticketflow /usr/local/bin/
```

**macOS (Intel)**
```bash
curl -L https://github.com/yshrsmz/ticketflow/releases/latest/download/ticketflow-darwin-amd64.tar.gz | tar xz
sudo mv ticketflow /usr/local/bin/
```

**macOS (Apple Silicon)**
```bash
curl -L https://github.com/yshrsmz/ticketflow/releases/latest/download/ticketflow-darwin-arm64.tar.gz | tar xz
sudo mv ticketflow /usr/local/bin/
```

### From Source

```bash
# Clone the repository
git clone https://github.com/yshrsmz/ticketflow
cd ticketflow

# Build the binary
make build

# Install to GOPATH/bin
make install

# Or build for specific platforms
make build-linux   # Build for Linux (amd64, arm64)
make build-mac     # Build for macOS (amd64, arm64)
make build-all     # Build for all platforms
```

### Using Go

```bash
go install github.com/yshrsmz/ticketflow/cmd/ticketflow@latest
```

## Quick Start

### Interactive TUI Mode

Launch the interactive TUI by running `ticketflow` without arguments:

```bash
ticketflow
```

**TUI Features:**
- Tab navigation between TODO, DOING, DONE, and ALL tickets
- Search tickets with `/` (real-time filtering)
- Create new tickets with `n`
- Start work on tickets with `s`
- View ticket details with `Enter`
- Edit tickets in external editor with `e`
- Close tickets with `c` (in detail view)
- View worktrees with `w`
- Help overlay with `?`

### Basic Workflow

1. **Initialize TicketFlow in your project**:
```bash
cd your-project
ticketflow init
```

2. **Create a new ticket**:
```bash
ticketflow new implement-feature
# Creates: tickets/todo/250124-150000-implement-feature.md

# Or create a sub-ticket with explicit parent
ticketflow new --parent 250124-150000-implement-feature add-tests
```

3. **Edit ticket details**:
```bash
$EDITOR tickets/todo/250124-150000-implement-feature.md
```

4. **Start working on the ticket**:
```bash
ticketflow start 250124-150000-implement-feature
# Creates branch and moves ticket to doing/
```

5. **Close ticket when done**:
```bash
ticketflow close
# Moves ticket to done/ and commits the change
```

6. **Push and create PR**:
```bash
git push origin 250124-150000-implement-feature
# Create PR on GitHub/GitLab/etc
```

### Worktree Workflow (Recommended)

1. **Enable worktrees in `.ticketflow.yaml`**:
```yaml
worktree:
  enabled: true
  base_dir: "../ticketflow.worktrees"  # Relative to project root
  init_commands:
    - "git fetch origin"
    # - "npm install"
    # - "make deps"
```

2. **Start work (creates worktree)**:
```bash
ticketflow start 250124-150000-implement-feature
# Creates worktree at ../ticketflow.worktrees/250124-150000-implement-feature
```

3. **Navigate to worktree and develop**:
```bash
cd ../ticketflow.worktrees/250124-150000-implement-feature
# Make changes, commit as usual
```

4. **Close ticket when done**:
```bash
ticketflow close
# Ticket marked as done, worktree remains for PR
```

5. **After PR is merged, clean up**:
```bash
cd ../../your-project
ticketflow cleanup 250124-150000-implement-feature
# Removes worktree and deletes local branch
```

### Closing Abandoned or Invalid Tickets

Sometimes tickets need to be closed without merging (e.g., requirements changed, duplicate work, invalid approach). TicketFlow supports closing tickets with a reason:

```bash
# Close current ticket with reason
ticketflow close --reason "Requirements changed"

# Close specific ticket from main repository
ticketflow close 250124-150000-abandoned-feature --reason "Duplicate of #123"

# Close ticket from anywhere with reason
ticketflow close 250124-150000-invalid-approach --reason "Better solution found in ticket #456"
```

When closing with a reason:
- The reason is stored in the ticket's frontmatter as `closure_reason`
- A closure note is added to the ticket content
- The git commit message includes the reason

### Auto-cleanup

Remove orphaned worktrees and stale branches for done tickets:

```bash
# Preview what would be cleaned
ticketflow cleanup --dry-run

# Perform cleanup
ticketflow cleanup
```

The auto-cleanup command will:
- Remove worktrees for tickets that no longer exist or are in done status
- Delete local git branches for tickets that are marked as done
- Show statistics of what was cleaned

## CLI Commands

### Core Commands

| Command | Description |
|---------|-------------|
| `ticketflow` | Launch interactive TUI |
| `ticketflow init` | Initialize ticket system in current repository |
| `ticketflow new [options] <slug>` | Create a new ticket |
| `ticketflow list [options]` | List tickets |
| `ticketflow show <id> [options]` | Show ticket details |
| `ticketflow start <id>` | Start working on a ticket |
| `ticketflow close [ticket] [options]` | Close current or specific ticket |
| `ticketflow restore` | Restore current-ticket symlink |
| `ticketflow status [options]` | Show current status |
| `ticketflow cleanup <id> [options]` | Clean up specific ticket after PR merge |
| `ticketflow cleanup [options]` | Auto-cleanup orphaned worktrees and stale branches |

### Worktree Commands

| Command | Description |
|---------|-------------|
| `ticketflow worktree list [options]` | List all worktrees |
| `ticketflow worktree clean` | Remove orphaned worktrees |

### Common Options

- `--status STATUS` - Filter by status (todo/doing/done)
- `--format FORMAT` - Output format (text/json)
- `--force, -f` - Force operation without confirmation
- `--count N` - Limit number of results

### Command-Specific Options

**new command:**
- `--parent TICKET_ID` - Specify parent ticket ID explicitly
- `-p TICKET_ID` - Short form of --parent

**Note:** Flags must come before the ticket slug (e.g., `ticketflow new --parent parent-id my-ticket`)

**close command:**
- `--force, -f` - Force close with uncommitted changes
- `--reason` - Reason for closing the ticket (required when closing abandoned/invalid tickets)

**cleanup command:**
- `--force` - Skip confirmation prompts (for specific ticket cleanup)
- `--dry-run` - Show what would be cleaned without making changes (for auto-cleanup)

- `--help, -h` - Show command help for any command

## Configuration

TicketFlow uses `.ticketflow.yaml` for configuration:

```yaml
# Git settings
git:
  default_branch: "main"

# Worktree settings  
worktree:
  enabled: true
  base_dir: "../.worktrees"
  init_commands:
    - "git fetch origin"
    # Add your project-specific setup commands

# Ticket settings
tickets:
  dir: "tickets"
  todo_dir: "todo"
  doing_dir: "doing" 
  done_dir: "done"
  
  # Template for new tickets
  template: |
    # Summary
    
    ## Tasks
    - [ ] 
    
    ## Notes

# Output settings
output:
  default_format: "text"
  json_pretty: true

# Timeout settings
timeouts:
  git: 30          # Timeout for git operations in seconds (max: 3600)
  init_commands: 60 # Timeout for worktree init commands in seconds (max: 3600)
```

## Sub-ticket Workflow

Create sub-tickets while working on a parent ticket:

### Method 1: Implicit Parent (from worktree)
```bash
# In parent worktree
cd ../ticketflow.worktrees/250124-150000-user-system

# Create sub-tickets (automatically linked to current ticket)
ticketflow new user-model
ticketflow new user-auth
```

### Method 2: Explicit Parent (from anywhere)
```bash
# From main repository or any location
ticketflow new --parent 250124-150000-user-system user-model
ticketflow new -p 250124-150000-user-system user-auth

# Start sub-ticket (branches from parent)
ticketflow start 250124-151000-user-model

# Work in sub-ticket worktree
cd ../ticketflow.worktrees/250124-151000-user-model
# ... implement ...

# Create PR targeting parent branch
git push origin 250124-151000-user-model
```

## AI Integration

TicketFlow is designed for seamless AI integration:

```bash
# Get structured data
ticketflow list --format json
ticketflow show 250124-150000 --format json
ticketflow status --format json

# AI-friendly error messages
export TICKETFLOW_OUTPUT_FORMAT=json
```

## Troubleshooting

### Restore Current Ticket

If the current ticket link is broken:
```bash
ticketflow restore
```

### Clean Orphaned Worktrees

Remove worktrees without active tickets:
```bash
ticketflow worktree clean
```

### Version Information

Check version and build info:
```bash
ticketflow version
```

## Performance

TicketFlow is optimized for handling large numbers of tickets efficiently:

- **Concurrent file operations**: List operations use parallel loading for 10+ tickets
- **Smart resource management**: Automatic worker pooling based on CPU cores
- **Optimized for scale**: 50%+ faster listing for 100+ tickets compared to sequential loading
- **Memory efficient**: Pre-allocated buffers and minimal allocations in hot paths
- **Context-aware**: All operations support cancellation for responsive UI

See [benchmark results](docs/benchmark-results.md) for detailed performance metrics.

## Development

```bash
# Run all tests
make test

# Run specific test suites
make test-unit        # Unit tests only
make test-integration # Integration tests only

# Format code
make fmt

# Check code quality
make vet
make lint            # Requires golangci-lint

# Build for all platforms
make build-all       # Creates binaries in dist/

# Show version info
make version

# Create a release
make release TAG=v1.0.0
make release-build   # Build release binaries
```

## Key Design Principles

1. **No Automatic Git Operations**: You control when to push, merge, or clean up
2. **Flat Worktree Structure**: All worktrees at the same level for simplicity  
3. **PR-based Workflow**: Designed for code review processes
4. **Local-first**: Everything works offline, no external services required
5. **AI-friendly**: Structured data formats and clear command outputs

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

MIT License - see LICENSE file for details