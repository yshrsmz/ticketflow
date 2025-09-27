# TicketFlow Project Overview

## Purpose
TicketFlow is a git worktree-based ticket management system written in Go that provides both TUI and CLI interfaces for managing development tickets using a directory-based status tracking system (todo/doing/done).

## Tech Stack
- **Language**: Go 1.23+ 
- **TUI Framework**: Bubble Tea (charmbracelet/bubbletea)
- **CLI Flags**: spf13/pflag (recently migrated from standard flag package)
- **Testing**: testify framework for assertions
- **Version Control**: Git with worktree support

## Project Structure
```
ticketflow/
├── cmd/ticketflow/          # CLI entry point
├── internal/
│   ├── config/             # Configuration management
│   ├── ticket/             # Ticket model and manager
│   ├── git/                # Git operations
│   ├── cli/                # CLI commands and output
│   │   └── commands/       # Individual command implementations
│   └── ui/                 # TUI implementation
│       ├── components/     # Reusable UI components
│       ├── styles/         # Styling and themes
│       └── views/          # View implementations
├── test/integration/       # Integration tests
├── scripts/githooks/       # Git hooks for quality checks
├── docs/                   # Documentation
└── tickets/               # Ticket files (todo/doing/done)
```

## Key Features
- Directory-based status tracking (tickets move between todo/, doing/, done/)
- Git worktree integration for parallel development
- Markdown files with YAML frontmatter for tickets
- Support for sub-tickets with parent relationships
- JSON output mode for AI tool integration
- Comprehensive TUI with keyboard navigation