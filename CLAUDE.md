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
6. **Small, Focused Functions**: Following Single Responsibility Principle, complex operations are decomposed into smaller helper functions for better maintainability and testability

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
- Helper functions are thoroughly tested with table-driven tests for better coverage

### Code Organization Best Practices
- **Function Size**: Keep functions under 50 lines when possible
- **Single Responsibility**: Each function should do one thing well
- **Clear Naming**: Use descriptive verb-noun combinations (e.g., `validateTicket`, `createWorktree`)
- **Error Handling**: Wrap errors with context using `fmt.Errorf` for better debugging
- **Helper Functions**: Extract common validation, setup, and cleanup logic into reusable helpers
- **Early Returns**: Use guard clauses to reduce nesting and improve readability

### Version Management
Version information is embedded at build time using ldflags:
- Version number from git tags
- Build timestamp
- Git commit hash
These are displayed in `ticketflow version` command.

## Development Workflow for New Features

1. Create a feature ticket: `ticketflow new my-feature`
2. Start work (creates worktree): `ticketflow start <ticket-id>`
   - You don't need to actually start work. just execute ticketflow command.
   - Actual work is done by another Coding Agent by launching new editor
3. Navigate to the worktree: `cd ../ticketflow.worktrees/<ticket-id>`
4. Make changes in the worktree
5. Run tests: `make test`
6. Run linters: `make fmt vet lint`
7. Commit and push changes
8. Create PR: `git push -u origin <branch>` and `gh pr create`
9. **IMPORTANT: Wait for developer approval before closing the ticket**
   - Check if the ticket contains approval requirements (e.g., "Get developer approval before closing")
   - If approval is required, DO NOT close the ticket until explicitly approved
   - Developer will review the PR and provide feedback or approval
10. **Only after approval, close the ticket FROM THE WORKTREE**: `ticketflow close`
11. Push the branch with the close commit: `git push`
12. After PR merge: `ticketflow cleanup <ticket-id>`

## Ticket Lifecycle Management

### CRITICAL: Always Read Ticket Requirements
Before closing any ticket, ALWAYS:
1. Read the ticket content for any specific requirements or approval steps
2. Look for phrases like "Get developer approval before closing" or "Requires review"
3. If approval is required, complete the implementation and create PR, but DO NOT close until approved
4. The developer will indicate approval through PR comments or direct communication

### IMPORTANT: Closing Tickets with Worktrees
When using worktrees (the default mode), you MUST close tickets from within the worktree directory, not from the main repository. This ensures the "Close ticket" commit is created on the feature branch.

**Correct workflow:**
```bash
cd ../ticketflow.worktrees/<ticket-id>  # Stay in the worktree
ticketflow close                        # Creates close commit on feature branch
git push                                # Push the complete branch with close commit
```

**Common mistake to avoid:**
```bash
# WRONG: Don't do this!
cd /path/to/main/repo      # Going back to main repo
ticketflow close           # This creates commit on wrong branch!
```

### Why this matters
- The ticket file exists in the feature branch, not in main
- Closing from the main repo may create the commit on the main branch
- This can cause confusing rebase issues where the ticket file disappears
- The close commit should be part of the PR, not a separate commit on main

## Important Notes
- The TUI is launched when running `ticketflow` without arguments
- JSON output mode (`-o json`) is designed for AI tool integration
- All git operations are explicit - no automatic pushing or merging
- Worktrees are created under the configured worktreeBaseDir (this project uses: ../ticketflow.worktrees)
- The application supports sub-tickets with parent relationships via YAML frontmatter
- **Always work within the worktree directory when making ticket-related changes**

## AI Integration Guidelines
- **ALWAYS use `--format json` when running ticketflow commands** to get structured output for better parsing and analysis
- JSON output provides comprehensive ticket information including metadata, relationships, and timestamps

## Testing Guidelines

### Git Configuration in Tests
- **NEVER use `git config --global` in tests** - This modifies the user's git configuration and can cause commits with wrong authors
- Always configure git locally within test directories: `cmd.Dir = tmpDir` before running git commands
- When writing test helpers that use git, add warning comments about not using `--global`
- Example of correct test git configuration:
  ```go
  // Configure git locally in the test repo (not globally)
  cmd := exec.Command("git", "config", "user.name", "Test User")
  cmd.Dir = tmpDir  // Critical: sets the working directory
  cmd.Run()
  ```
- This issue was discovered when test code using `--global` caused subsequent commits to have wrong authors ("Test User" instead of the actual developer)

### Test Parallelization and os.Chdir
- **Unit tests**: Should use `t.Parallel()` for better performance when they don't modify global state
- **Integration tests**: Cannot use `t.Parallel()` because they use `os.Chdir` to change working directory
- **Why os.Chdir is required**: The ticketflow application is designed to work from the project root directory (similar to git), expecting:
  - `.ticketflow.yaml` configuration file in the current directory
  - `tickets/` directory structure relative to the current directory
  - Git repository in the current directory
- **Best practices for tests using os.Chdir**:
  ```go
  originalWd, err := os.Getwd()
  require.NoError(t, err)
  defer func() {
      err := os.Chdir(originalWd)
      require.NoError(t, err)
  }()
  require.NoError(t, os.Chdir(testDir))
  ```
- See `test/integration/README.md` for detailed explanation of integration test requirements
