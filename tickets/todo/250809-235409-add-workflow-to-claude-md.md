---
priority: 2
description: "Append recommended development workflow to CLAUDE.md during ticketflow init"
created_at: "2025-08-09T23:54:09+09:00"
started_at: null
closed_at: null
---

# Append recommended development workflow to CLAUDE.md during init

## Overview
When users run `ticketflow init`, the command should append the recommended development workflow using ticketflow to the CLAUDE.md file. This will help AI assistants (like Claude) understand how to properly use ticketflow for development tasks.

## Requirements
1. The workflow section should be appended to CLAUDE.md (not replace existing content)
2. Should check if CLAUDE.md already contains the workflow section to avoid duplicates
3. The workflow should include:
   - How to create tickets
   - How to start work with worktrees
   - How to navigate to worktrees
   - How to close tickets properly (from within worktree)
   - How to handle PR creation and approval
   - How to cleanup after merge

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [ ] Locate the init command implementation in `internal/cli/init.go`
- [ ] Create a template string with the recommended workflow content
- [ ] Add logic to check if CLAUDE.md exists
- [ ] Check if workflow section already exists in CLAUDE.md (avoid duplicates)
- [ ] Append workflow section to CLAUDE.md if not present
- [ ] Handle file creation if CLAUDE.md doesn't exist
- [ ] Add unit tests for the CLAUDE.md workflow appending logic
- [ ] Test edge cases (no CLAUDE.md, existing workflow section, file permissions)
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Implementation Notes
- The workflow content should be similar to what's currently in the existing CLAUDE.md "Development Workflow for New Features" section
- Use a marker comment or heading to identify the workflow section (e.g., "## Development Workflow with TicketFlow")
- Consider making the workflow content a constant or embedded resource for maintainability
- The appending should be idempotent - running init multiple times shouldn't duplicate the content