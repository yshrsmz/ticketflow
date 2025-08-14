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

- [x] Modify `moveTicketToDoneWithReason` in `internal/cli/commands.go` to accept `isCurrentTicket bool` parameter
- [x] Update the symlink removal logic to only execute when `isCurrentTicket == true`
- [x] Update `closeCurrentTicketInternal` to pass `true` for isCurrentTicket
- [x] Update `closeAndCommitTicket` to pass `false` for isCurrentTicket
- [x] Add unit test for `CloseTicketByID` when closing non-current ticket
- [x] Add integration test verifying current-ticket.md preservation when closing other tickets
- [x] Test edge cases (no current-ticket.md, broken symlink, etc.)
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Implementation Notes
- The simpler parameter-based approach is preferred over symlink checking
- This bug affects both worktree and non-worktree modes
- Consider adding debug logging when preserving current-ticket.md
- Currently no test coverage for this scenario - that's why the bug wasn't caught

## Testing Gap Identified
There are no existing tests for closing a ticket by ID when it's not the current ticket. This gap in test coverage allowed this bug to slip through.

## Resolution Insights

### Bug Analysis
The root cause was a simple logic error where `moveTicketToDoneWithReason` unconditionally removed the current-ticket.md symlink whenever ANY ticket was closed. This happened because the function didn't know whether it was closing the current ticket or a different one.

### Solution Approach
Instead of the initially proposed symlink checking approach, we implemented a simpler solution:
- Added an `isCurrentTicket` boolean parameter to `moveTicketToDoneWithReason`
- The calling functions already know whether they're closing the current ticket, so they pass the appropriate value
- This avoids the complexity of symlink reading and comparison

### Testing Improvements
Added comprehensive test coverage:
1. **Unit Test**: Verifies that `SetCurrentTicket(nil)` is NOT called when closing a non-current ticket
2. **Integration Tests**: End-to-end verification of three scenarios:
   - Preserving current-ticket.md when closing a different ticket
   - Removing current-ticket.md when closing the current ticket
   - Graceful handling when current-ticket.md doesn't exist

### Lessons Learned
1. **Simplicity wins**: The parameter-based approach is cleaner than symlink checking
2. **Test coverage gaps**: Missing tests for common scenarios can hide bugs
3. **Clear separation of concerns**: Functions should know their context (current vs non-current)
4. **Integration tests are valuable**: They catch real-world usage patterns that unit tests might miss