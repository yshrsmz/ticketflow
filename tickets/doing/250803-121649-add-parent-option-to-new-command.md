---
priority: 3
description: Add --parent flag to ticketflow new command for explicit parent relationships
created_at: "2025-08-03T12:16:49+09:00"
started_at: "2025-08-03T22:50:03+09:00"
closed_at: null
---

# Add Parent Option to New Command

## Overview
Currently, parent-child relationships are only created when running `ticketflow new` from within a parent ticket's worktree. Add a `--parent` flag to explicitly specify a parent ticket from any location.

This is a standalone feature enhancement, not a sub-ticket of the branch fix.

## Tasks
- [ ] Add --parent flag to new command
- [ ] Validate parent ticket exists and is valid
- [ ] Add parent relationship to ticket metadata
- [ ] Update help text and documentation
- [ ] Add tests for parent flag functionality

## Technical Details
```bash
# Current behavior (implicit parent from worktree)
cd ../ticketflow.worktrees/parent-ticket
ticketflow new sub-ticket

# Proposed behavior (explicit parent)
ticketflow new --parent parent-ticket-id sub-ticket
ticketflow new -p parent-ticket-id sub-ticket  # short form
```

Implementation:
- Add flag to CLI parser: `flag.String("parent", "", "Parent ticket ID")`
- Validate parent ticket exists in any status (todo/doing/done)
- Add to ticket's Related field: `parent:<ticket-id>`
- Should work from any directory, not just worktrees

## Acceptance Criteria
- Can create sub-tickets with explicit parent from main repo
- Can create sub-tickets with explicit parent from any worktree
- Parent validation provides clear error if ticket doesn't exist
- Help text clearly explains the option
- Tests cover both valid and invalid parent scenarios