---
priority: 2
description: "Extend close command to handle abandoned/invalid tickets without requiring worktree"
created_at: "2025-08-09T11:58:10+09:00"
started_at: null
closed_at: null
related:
    - parent:250806-172829-improve-worktree-error-messages
---

# Extend Close Command for Abandoned Tickets

## Overview

Currently, closing a ticket requires being in a worktree if worktrees are enabled, which assumes the ticket was worked on. However, tickets can become invalid, obsolete, or duplicate at any stage (todo or doing), and we need a way to close them without the full workflow.

Current manual workaround: Users edit the ticket content with reasoning, manually update `closed_at` field, and move the file to the done directory.

## Problem Scenarios

1. **Ticket in todo** - Found to be invalid/duplicate before starting
2. **Ticket in doing** - Started work, then discovered it's not needed/wrong approach
3. **Working on ticket A** - Discover ticket B is invalid/duplicate

## Proposed Solution

Extend the `ticketflow close` command to handle different closure scenarios:

```bash
# Normal close (in worktree, work completed)
ticketflow close

# Close without implementation (from anywhere)
ticketflow close <ticket-id> --reason "Invalid: requirement removed"

# Close with abandoned flag (clearer intent)
ticketflow close <ticket-id> --abandoned "Superseded by #456"

# Close from main repo when ticket has worktree
ticketflow close <ticket-id> --force --reason "Obsolete"
```

## Technical Design

### Command Flags
- `--reason <text>` - Explanation for closing without completion
- `--abandoned` - Explicitly mark as abandoned (optional, for clarity)
- `--force` - Allow closing from main repo even if worktree exists

### Behavior by Context

1. **Ticket in todo (never started)**:
   - Can be closed from main repo
   - Add closure note to ticket content
   - Update frontmatter with `closed_at` and optionally `closure_type`
   - Move to `done/` directory
   - Commit: "Close ticket (not implemented): <ticket-id>"

2. **Ticket in doing with worktree**:
   - If in worktree: Close normally with reason
   - If in main repo with `--force`: Close and suggest cleanup
   - Add closure note to ticket content
   - Update frontmatter
   - Move to `done/` directory
   - Commit: "Close ticket (abandoned): <ticket-id>"

3. **Ticket in doing without worktree** (non-worktree mode):
   - Similar to todo handling
   - Can close from current branch

### Frontmatter Updates

```yaml
closed_at: "2025-08-09T..."
closure_type: "abandoned"  # or "completed", "cancelled"
closure_reason: "Superseded by #456"
```

### Ticket Content Updates

When closing with a reason, append to ticket content:
```markdown
## Closure Note
**Closed on**: 2025-08-09
**Type**: Abandoned
**Reason**: Superseded by #456 - better approach found
```

## Tasks

- [ ] Add command flags (--reason, --abandoned, --force) to close command
- [ ] Implement logic to detect ticket status and location
- [ ] Add validation for closing from main repo vs worktree
- [ ] Implement frontmatter updates with closure metadata
- [ ] Add closure note to ticket content when reason provided
- [ ] Handle file move from todo/doing to done directory
- [ ] Create appropriate commit messages based on closure type
- [ ] Add confirmation prompt for destructive operations
- [ ] Update error messages for invalid operations
- [ ] Add unit tests for all closure scenarios
- [ ] Add integration tests for workflow
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation for new close options
- [ ] Update README.md with examples
- [ ] Get developer approval before closing

## Acceptance Criteria

- [ ] Can close todo tickets from main repo with reason
- [ ] Can close doing tickets from worktree with reason
- [ ] Can force-close doing tickets from main repo with --force
- [ ] Closure metadata is properly saved in frontmatter
- [ ] Closure reason is appended to ticket content
- [ ] Appropriate commit messages for different closure types
- [ ] Clear error messages when operation not allowed
- [ ] Backward compatibility with existing close behavior

## Notes

This change maintains backward compatibility - `ticketflow close` without flags continues to work as before for normal ticket completion. The new flags only add capabilities for handling abandoned/invalid tickets.

Consider future enhancement: `ticketflow cleanup --abandoned` to remove worktrees and branches for abandoned tickets in bulk.