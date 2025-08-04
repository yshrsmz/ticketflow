---
priority: 2
description: ""
created_at: "2025-08-03T13:08:21+09:00"
started_at: "2025-08-04T17:27:48+09:00"
closed_at: null
related:
    - parent:250726-183403-fix-branch-already-exist-on-start
---

# Ticket Overview

Add a `--force` flag to the `ticketflow start` command to allow recreating worktrees when they already exist. This feature will improve developer experience when they need to recreate a worktree without manual cleanup.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [ ] Add `--force` flag to the start command CLI
- [ ] Implement logic to remove existing worktree when --force is used
- [ ] Add confirmation prompt when using --force (unless --yes is also provided)
- [ ] Update command help text
- [ ] Add unit tests for the --force flag behavior
- [ ] Add integration tests
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation if necessary
- [ ] Update README.md
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Technical Specification

### Current Behavior
- When a worktree already exists, the command fails with an error
- Users must manually run `ticketflow cleanup` or remove the worktree

### Proposed Behavior
- Add `--force` flag to `ticketflow start`
- When used with existing worktree:
  1. Show warning message
  2. Remove existing worktree
  3. Create new worktree
- Consider adding `--yes` flag to skip confirmation

### Implementation Details
1. Add flag to start command in CLI
2. Check flag in `StartTicket` method
3. If force and worktree exists:
   - Call `RemoveWorktree`
   - Proceed with normal creation
4. Add appropriate logging and user feedback

## Notes

This feature was suggested during code review as a quality-of-life improvement for developers who need to recreate worktrees. Related to the parent ticket that fixed the "branch already exists" error.