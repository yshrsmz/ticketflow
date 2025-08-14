---
priority: 2
description: Fix current-ticket.md removal without verification when closing tickets
created_at: "2025-08-09T23:48:41+09:00"
started_at: "2025-08-14T16:55:38+09:00"
closed_at: null
---

# Fix current-ticket.md removal without verification

## Problem
When closing a ticket provided as a CLI parameter, ticketflow currently removes current-ticket.md without checking if the parameter ticket matches what current-ticket.md is pointing to. This can lead to incorrect removal of the current ticket symlink.

**Bug Location**: `internal/cli/commands.go:1319` in `moveTicketToDoneWithReason` function

Example scenario:
1. User has current-ticket.md pointing to ticket-A
2. User runs `ticketflow close ticket-B`
3. current-ticket.md gets removed even though it was pointing to ticket-A, not ticket-B

## Root Cause
The `moveTicketToDoneWithReason` function unconditionally calls `app.Manager.SetCurrentTicket(ctx, nil)` which removes the current-ticket.md symlink regardless of which ticket is being closed.

## Solution
**Simplified Approach**: Add an `isCurrentTicket` boolean parameter to `moveTicketToDoneWithReason` function.

The calling code already knows whether it's closing the current ticket:
- `closeCurrentTicketInternal` → pass `true`
- `closeAndCommitTicket` (via `CloseTicketByID`) → pass `false`

Only remove current-ticket.md when `isCurrentTicket == true`.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [ ] Modify `moveTicketToDoneWithReason` in `internal/cli/commands.go` to accept `isCurrentTicket bool` parameter
- [ ] Update the symlink removal logic to only execute when `isCurrentTicket == true`
- [ ] Update `closeCurrentTicketInternal` to pass `true` for isCurrentTicket
- [ ] Update `closeAndCommitTicket` to pass `false` for isCurrentTicket
- [ ] Add unit test for `CloseTicketByID` when closing non-current ticket
- [ ] Add integration test verifying current-ticket.md preservation when closing other tickets
- [ ] Test edge cases (no current-ticket.md, broken symlink, etc.)
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Implementation Notes
- The simpler parameter-based approach is preferred over symlink checking
- This bug affects both worktree and non-worktree modes
- Consider adding debug logging when preserving current-ticket.md
- Currently no test coverage for this scenario - that's why the bug wasn't caught

## Testing Gap Identified
There are no existing tests for closing a ticket by ID when it's not the current ticket. This gap in test coverage allowed this bug to slip through.