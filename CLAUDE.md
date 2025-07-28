# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
TicketFlow is a git worktree-based ticket management system written in Go. It provides both a TUI (Terminal User Interface) and CLI for managing development tickets using a directory-based status tracking system (todo/doing/done).

## Build and Development Commands

### Essential Commands
```bash
# Build the application
make build

# Run all tests
make test

# Run specific test suites
make test-unit          # Unit tests only
make test-integration   # Integration tests only

# Run a single test
go test -v -run TestFunctionName ./path/to/package

# Code quality checks (run before committing)
make fmt    # Format code with go fmt
make vet    # Run go vet
make lint   # Run golangci-lint

# Generate test coverage
make coverage
```

### Cross-platform builds
```bash
make build-all          # Build for all platforms
make release-archives   # Create release archives with checksums
```

## Architecture Overview

### Core Package Structure
- **cmd/ticketflow/**: Main entry point, initializes either TUI or CLI mode based on arguments
- **internal/config/**: Configuration management, loads and validates .ticketflow.yaml
- **internal/ticket/**: Core ticket model and manager, handles YAML frontmatter and ticket lifecycle
- **internal/git/**: Git operations wrapper, manages worktrees and branches
- **internal/cli/**: CLI command implementations and output formatting (supports JSON for AI integration)
- **internal/ui/**: TUI implementation using Bubble Tea framework
  - **components/**: Reusable UI components (list, tabs, dialogs)
  - **views/**: Main application views (list, detail, new ticket, worktree management)
  - **styles/**: Theming and visual styling

### Key Design Patterns
1. **Directory-based Status**: Tickets move between todo/, doing/, and done/ directories to track status
2. **Worktree Integration**: Each ticket gets its own git worktree for parallel development
3. **Model-View Architecture**: Clear separation between business logic (internal packages) and UI layers
4. **Command Pattern**: CLI commands are self-contained with their own Execute methods
5. **Interface-based Git Operations**: Git package provides an abstraction over git commands

### Configuration (.ticketflow.yaml)
The application is configured via .ticketflow.yaml which controls:
- Git settings (default branch for worktrees)
- Worktree behavior (enabled/disabled, base directory, init commands)
- Ticket directories and templates
- Output format preferences

### Testing Strategy
- Unit tests use testify framework with mocks for external dependencies
- Integration tests in test/integration/ verify full workflows
- Test files follow Go convention (*_test.go) alongside implementation
- Mock implementations provided for Git operations to enable testing without real git repos

### Version Management
Version information is embedded at build time using ldflags:
- Version number from git tags
- Build timestamp
- Git commit hash
These are displayed in `ticketflow version` command.

## Development Workflow for New Features

1. Create a feature ticket: `ticketflow new my-feature`
2. Start work (creates worktree): `ticketflow start <ticket-id>`
3. Make changes in the worktree
4. Run tests: `make test`
5. Run linters: `make fmt vet lint`
6. Commit and push changes
7. After PR merge: `ticketflow cleanup <ticket-id>`

## Important Notes
- The TUI is launched when running `ticketflow` without arguments
- JSON output mode (`-o json`) is designed for AI tool integration
- All git operations are explicit - no automatic pushing or merging
- Worktrees are created under the configured worktreeBaseDir (default: ../)
- The application supports sub-tickets with parent relationships via YAML frontmatter

## AI Integration Guidelines
- **ALWAYS use `--format json` when running ticketflow commands** to get structured output for better parsing and analysis
- JSON output provides comprehensive ticket information including metadata, relationships, and timestamps
