---
priority: 2
description: Add 'workflow' command to print development workflow guide
created_at: "2025-08-09T23:54:09+09:00"
started_at: "2025-08-17T00:39:09+09:00"
closed_at: null
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
- [ ] Add integration test to verify command output
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Implementation Notes
- The workflow content should be based on the existing CLAUDE.md "Development Workflow for New Features" section
- Keep it simple - just a command that prints text to stdout
- No flags or complex options needed initially
- Users can redirect output as needed: `ticketflow workflow > CLAUDE.md` or `ticketflow workflow >> .cursorrules`