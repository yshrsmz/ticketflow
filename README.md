# TicketFlow

A git worktree-based ticket management system optimized for AI collaboration.

## Phase 1, 2, 3 & 4 Implementation Status ✅

This implementation includes Phase 1 (MVP), Phase 2 (Worktree Integration), Phase 3 (TUI), and Phase 4 (Advanced Features):

### Core Features Implemented:

**Phase 1 - Core Functionality:**
- ✅ Configuration management with YAML files
- ✅ Ticket model with YAML frontmatter and Markdown content
- ✅ Basic CLI commands (init, new, list, show, start, close, restore, status)
- ✅ Git branch management for tickets
- ✅ Current ticket symlink tracking
- ✅ JSON output support for AI integration

**Phase 2 - Worktree Integration:**
- ✅ Git worktree creation on ticket start
- ✅ Automatic worktree removal on ticket close
- ✅ Worktree subcommands (list, clean)
- ✅ Init commands execution in new worktrees
- ✅ Orphaned worktree cleanup
- ✅ Parallel development support

**Phase 3 - Terminal UI (TUI):**
- ✅ Interactive TUI with Bubble Tea framework
- ✅ Ticket list view with filtering and navigation
- ✅ Ticket detail view with content display
- ✅ New ticket creation form
- ✅ Worktree management view
- ✅ Keyboard navigation and help overlay
- ✅ Comprehensive styling with lipgloss

**Phase 4 - Advanced Features:**
- ✅ Progress tracking with tasks and percentage
- ✅ Progress reporting for active tickets
- ✅ Auto-cleanup functionality for old tickets
- ✅ Automatic ticket archiving
- ✅ Orphaned worktree cleanup
- ✅ Stale branch removal
- ✅ Progress visualization in TUI

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

```bash
# Build from source
make build

# Install to GOPATH/bin
make install
```

## Quick Start

### Interactive TUI Mode:

Simply run `ticketflow` without any arguments to launch the interactive terminal UI:

```bash
ticketflow
```

TUI Features:
- Browse and filter tickets by status and priority
- View ticket details with scrolling
- Create new tickets with form input
- Start/close tickets with visual feedback
- Manage worktrees interactively
- Keyboard shortcuts with help overlay (press `?`)

### Basic Workflow (without worktrees):

1. Initialize in your git project:
```bash
ticketflow init
```

2. Create a new ticket:
```bash
ticketflow new implement-feature
```

3. Start working (creates branch):
```bash
ticketflow start 250124-150000-implement-feature
# Now on branch: 250124-150000-implement-feature
```

4. Close ticket when done:
```bash
ticketflow close
# Merges to main, ready to commit
```

### Worktree Workflow (recommended):

1. Enable worktrees in `.ticketflow.yaml`:
```yaml
worktree:
  enabled: true
  base_dir: "./.worktrees"
  init_commands:
    - "npm install"  # or your setup commands
```

2. Start work (creates separate worktree):
```bash
ticketflow start 250124-150000-implement-feature
# Created worktree: ./.worktrees/250124-150000-implement-feature
cd ./.worktrees/250124-150000-implement-feature
```

3. Work in isolated environment, then close:
```bash
ticketflow close
# Merges changes and removes worktree
```

## CLI Commands

- `ticketflow init` - Initialize ticket system
- `ticketflow new <slug>` - Create new ticket
- `ticketflow list [--status todo|doing|done] [--format json]` - List tickets
- `ticketflow show <ticket-id> [--format json]` - Show ticket details
- `ticketflow start <ticket-id> [--no-push]` - Start work on ticket
- `ticketflow close [--no-push] [--force]` - Complete current ticket
- `ticketflow restore` - Restore current-ticket symlink
- `ticketflow status [--format json]` - Show current status
- `ticketflow worktree list [--format json]` - List all worktrees
- `ticketflow worktree clean` - Remove orphaned worktrees
- `ticketflow progress update <ticket> <percentage>` - Update ticket progress
- `ticketflow progress show <ticket>` - Show ticket progress
- `ticketflow progress add-task <ticket> <description>` - Add task to ticket
- `ticketflow progress complete-task <ticket> <index>` - Complete a task
- `ticketflow progress report` - Generate progress report
- `ticketflow cleanup [--dry-run]` - Clean up old tickets and branches

## Implementation Complete

All four phases of TicketFlow have been successfully implemented. The system is fully functional with:
- Core ticket management functionality
- Git worktree integration for parallel development
- Interactive terminal UI
- Advanced features including progress tracking and auto-cleanup

The system is ready for production use and AI agent integration.

## Development

```bash
# Run tests
make test

# Run specific test suites
make test-unit
make test-integration

# Format code
make fmt

# Build for all platforms
make build-all
```

## License

MIT