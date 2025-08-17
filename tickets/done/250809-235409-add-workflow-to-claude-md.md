---
priority: 2
description: Add 'workflow' command to print development workflow guide
created_at: "2025-08-09T23:54:09+09:00"
started_at: "2025-08-17T00:39:09+09:00"
closed_at: "2025-08-17T12:20:43+09:00"
---

# Add 'workflow' command to print development workflow guide

## Overview
Create a simple command `ticketflow workflow` that prints the ticketflow development workflow guide to stdout. This allows users to integrate the workflow guide with any AI tool (Claude, Cursor, Copilot), documentation system, or their preferred workflow setup.

## Requirements
1. Print comprehensive workflow guide to stdout in markdown format
2. Include all essential workflows:
   - How to create tickets
   - How to start work with worktrees
   - How to navigate to worktrees
   - How to close tickets properly (from within worktree)
   - How to handle PR creation and approval
   - How to cleanup after merge
3. Be tool-agnostic - users decide where to pipe the output
4. Simple implementation - just print and exit, no file manipulation

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [x] Create `workflow.go` command file in `internal/cli/commands/`
- [x] Register the workflow command in the CLI router
- [x] Embed the workflow content as a string constant
- [x] Implement Execute method that prints to stdout
- [x] Add integration test to verify command output
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Update the ticket with insights from resolving this ticket
- [x] Implement code review improvements (io.Writer abstraction, error handling)
- [x] Add comprehensive unit tests with 100% coverage
- [x] Address PR review comments (embed.FS, test refactoring)
- [x] Create PR and pass CI checks
- [ ] Get developer approval before closing

## Implementation Notes
- The workflow content should be based on the existing CLAUDE.md "Development Workflow for New Features" section
- Keep it simple - just a command that prints text to stdout
- No flags or complex options needed initially
- Users can redirect output as needed: `ticketflow workflow > CLAUDE.md` or `ticketflow workflow >> .cursorrules`

## Implementation Insights

### Initial Implementation
- Implemented as a simple command following the existing Command interface pattern (similar to version.go)
- Started with workflow content as a large const string in the Go file
- Integration tests verify both direct output and shell redirection scenarios
- Tests use dynamic project root detection to work from any directory
- Command is AI-agnostic - users decide where to pipe the output
- No file manipulation needed, follows Unix philosophy of outputting to stdout

### Improvements from Code Review
- **Testability**: Added io.Writer abstraction to enable unit testing without capturing stdout
- **Error Handling**: Properly handle and wrap write errors with context
- **Test Coverage**: Added comprehensive unit tests achieving 100% coverage, including:
  - Context cancellation scenarios
  - Write error handling
  - Document structure validation
  - Markdown format verification

### Refactoring from PR Review
- **Maintainability**: Moved from large string constant to `//go:embed` directive
  - Created `resources/workflow.md` for better content management
  - Used Go's embed package for compile-time embedding
  - Makes content editing easier with proper markdown syntax highlighting
- **DRY Principle**: Extracted duplicated test code into `findProjectRoot()` helper
- **No Runtime Impact**: Content is embedded at compile time, no file I/O overhead

### Key Learnings
1. **Go Embed Package**: The `//go:embed` directive is excellent for embedding static content while keeping code clean
2. **Test Helpers**: Extracting common test logic improves maintainability and reduces duplication
3. **Interface Design**: Using io.Writer instead of direct stdout makes code more testable
4. **PR Workflow**: Iterative improvements through code review lead to better quality
5. **Unix Philosophy**: Simple tools that do one thing well and output to stdout are highly composable

### Future Enhancements
- Could add format flags (--format=plain, --format=markdown, --format=html)
- Could support multiple AI tool formats (Cursor, Copilot specific formats)
- Could add interactive mode to select specific sections to output