---
priority: 1
description: "Fix current-ticket.md being created in both main repo and worktree"
created_at: "2025-08-08T17:42:50+09:00"
started_at: null
closed_at: null
---

# Fix current-ticket.md duplication in worktree mode

## Problem
When starting a ticket from the default branch with worktrees enabled, `current-ticket.md` symlink is created in both the main repository and the worktree. This causes confusion as the symlink should only exist in the worktree when worktree mode is enabled.

### Root Cause
In `internal/cli/commands.go`:
1. Line 986: `moveTicketToDoing()` calls `SetCurrentTicket()` which creates symlink in main repo
2. Line 475: Worktree is created afterward (doesn't inherit the symlink since it's gitignored)
3. Line 1192: `createWorktreeTicketSymlink()` creates another symlink in the worktree

This results in two symlinks existing simultaneously.

## Solution
In worktree mode, skip creating the symlink in the main repository during `moveTicketToDoing()`. The symlink should only be created in the worktree via `createWorktreeTicketSymlink()`.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [ ] Modify `moveTicketToDoing()` to conditionally call `SetCurrentTicket()` based on worktree mode
- [ ] Ensure symlink is only created in appropriate location for each mode
- [ ] Add test to verify symlink location in worktree mode
- [ ] Add test to verify symlink location in non-worktree mode
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Test manually with both worktree enabled and disabled
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Acceptance Criteria
- When starting a ticket with worktrees enabled, `current-ticket.md` only exists in the worktree
- When starting a ticket with worktrees disabled, `current-ticket.md` exists in the main repo
- No regression in non-worktree mode
- All existing tests pass

## Notes
This issue was discovered when running `ticketflow start` from the main branch with worktrees enabled.