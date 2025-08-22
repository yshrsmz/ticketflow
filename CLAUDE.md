# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
TicketFlow is a git worktree-based ticket management system written in Go. It provides both a TUI (Terminal User Interface) and CLI for managing development tickets using a directory-based status tracking system (todo/doing/done).

## Build and Development Commands

### Essential Commands
```bash
# Initialize development environment (first time setup)
make init               # Sets up dependencies, git hooks, and worktree

# Build the application (binary will be placed at repository root as ./ticketflow)
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

### Git Hooks (Lefthook)
The project uses Lefthook for git hooks management. Hooks are automatically installed when you run `make init` or `make setup-hooks`.

**Pre-commit hooks** (fast checks):
- `gofmt` - Formats Go code automatically
- `go vet` - Static analysis
- `golangci-lint --fast` - Quick linting

**Pre-push hooks** (comprehensive checks):
- `make test` - Runs all tests
- `make build` - Verifies build succeeds
- `make lint` - Full linting

To skip hooks temporarily: `git commit --no-verify` or `git push --no-verify`

### Cross-platform builds
```bash
make build-all          # Build for all platforms
make release-archives   # Create release archives with checksums
```

## Ticket Format

Tickets are stored as Markdown files with YAML frontmatter:

```yaml
---
priority: 2
description: "Brief description of the ticket"
created_at: "2025-08-09T12:00:00+09:00"
started_at: "2025-08-09T13:00:00+09:00"
closed_at: null
related:
    - "parent:250809-121723-parent-ticket"
    - "blocks:250810-090000-dependent-feature"
    - "related:250811-100000-similar-work"
---

# Ticket Title

Detailed description and tasks...
```

### Related Field Format
The `related` field uses quoted strings to prevent YAML parsing issues:
- **Parent relationships**: `"parent:ticket-id"`
- **Blocking relationships**: `"blocks:ticket-id"` or `"blocked-by:ticket-id"`
- **General relationships**: `"related:ticket-id"`

Note: Quotes are required to prevent automated tools from incorrectly flagging the colon as a YAML syntax error.

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
7. **Context-Aware Operations**: All long-running operations support context.Context for cancellation and timeouts (see docs/context-usage.md)

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
- **Context Usage**: All long-running operations should accept `context.Context` as first parameter and check for cancellation

### Version Management
Version information is embedded at build time using ldflags:
- Version number from git tags
- Build timestamp
- Git commit hash
These are displayed in `ticketflow version` command.

## Development Workflow for New Features

1. Create a feature ticket: `ticketflow new my-feature`
   - Or create a sub-ticket with explicit parent: `ticketflow new --parent parent-ticket-id my-sub-feature`
   - Note: Flags must come before the ticket slug
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

## GitHub PR Management

### IMPORTANT: Correct way to check PR status with gh command
When checking PR status, use `statusCheckRollup` instead of `checks` (which doesn't exist):

```bash
# CORRECT - Use statusCheckRollup for CI status
gh pr view 63 --json statusCheckRollup,reviews,comments,title,body

# WRONG - checks field doesn't exist
gh pr view 63 --json checks  # This will error!
```

### Common PR fields for --json flag:
- `statusCheckRollup` - CI/checks status
- `reviews` - PR reviews
- `comments` - PR comments  
- `commits` - List of commits
- `title`, `body` - PR title and description
- `files` - Changed files
- `state` - PR state (open/closed/merged)
- `latestReviews` - Latest review from each reviewer

### Example: Full PR status check
```bash
# Get comprehensive PR information
gh pr view 63 --json statusCheckRollup,latestReviews,comments,state,mergeable

# Check CI status specifically
gh pr view 63 --json statusCheckRollup --jq '.statusCheckRollup'

# List all available fields (if unsure)
gh pr view 63 --json 2>&1 | grep "Available fields"
```

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

### Testing Strategy for CLI Commands

#### Key Principles
After architectural analysis, we've adopted an integration-first testing strategy for CLI commands:

1. **Execute methods are orchestrators** - They coordinate multiple components (git, file system, config) and should be tested as integration tests, not unit tests with mocks
2. **Wrong abstraction level** - Mocking at Manager/Git level creates brittle tests that don't verify real behavior  
3. **Industry patterns** - Tools like git, docker, and kubectl use integration tests for commands, not unit tests with mocks
4. **Test harness approach** - Create real test environments with:
   - Temporary git repositories
   - Actual file system operations
   - Real ticket structures
   - Proper configuration files
5. **Focus on behavior, not implementation** - Test what users experience, not internal method calls

#### Test Organization
- **Unit tests**: For pure business logic (validators, parsers, formatters)
- **Integration tests**: For command Execute methods (use `_integration_test.go` suffix)
- **Test harness**: Shared test infrastructure in `internal/cli/commands/testharness/`

#### Writing Integration Tests
```go
func TestCommand_Execute_Integration(t *testing.T) {
    // Create test environment
    env := testharness.NewTestEnvironment(t)
    
    // Setup test data
    env.CreateTicket("test-ticket", ticket.StatusTodo)
    env.RunGit("add", ".")
    env.RunGit("commit", "-m", "Setup")
    
    // Execute command
    cmd := commands.NewStartCommand()
    err := cmd.Execute(ctx, flags, []string{"test-ticket"})
    
    // Verify behavior
    assert.True(t, env.FileExists("tickets/doing/test-ticket.md"))
    assert.Contains(t, env.LastCommitMessage(), "Start ticket")
}
```

#### Coverage Expectations
- Integration tests provide meaningful coverage of Execute methods
- Don't chase coverage percentages with mocks - prefer real behavior verification
- Accept that some error paths may only be testable through integration tests

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

### Current-ticket.md Handling
**IMPORTANT**: `current-ticket.md` is just a symlink that points to a ticket in `tickets/doing/`. When updating or committing ticket changes, always update the actual file in `tickets/doing/`, NOT the symlink. The symlink should remain gitignored.

## Repository Structure Guidelines

### File Placement Rules
**NEVER create documentation files in the repository root directory**. The root should only contain essential files like:
- README.md (main project documentation)
- Makefile, go.mod, go.sum (build configuration)
- .ticketflow.yaml (project configuration)
- LICENSE, CONTRIBUTING.md (if applicable)

### Where to Put Documentation
- **Permanent documentation**: Place in `docs/` directory
- **Temporary notes/proposals**: Add to the ticket description or PR description
- **Implementation details**: Use code comments inline with the code
- **Architecture decisions**: Create ADR (Architecture Decision Records) in `docs/adr/`
- **Test documentation**: Place in test directories or `test/README.md`

Never create files like REFACTORING_PROPOSAL.md, IMPROVEMENTS_SUMMARY.md, or similar in the repository root.
