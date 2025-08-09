---
priority: 2
description: "Fix current-ticket.md removal without verification when closing tickets"
created_at: "2025-08-09T23:48:41+09:00"
started_at: null
closed_at: null
---

# Fix current-ticket.md removal without verification

## Problem
When closing a ticket provided as a CLI parameter, ticketflow currently removes current-ticket.md without checking if the parameter ticket matches what current-ticket.md is pointing to. This can lead to incorrect removal of the current ticket symlink.

Example scenario:
1. User has current-ticket.md pointing to ticket-A
2. User runs `ticketflow close ticket-B`
3. current-ticket.md gets removed even though it was pointing to ticket-A, not ticket-B

## Solution
Add verification before deleting current-ticket.md:
1. Read the symlink target of current-ticket.md if it exists
2. Compare the target with the ticket being closed
3. Only remove current-ticket.md if they match
4. Leave it untouched if it points to a different ticket

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [ ] Locate the close command implementation in `internal/cli/close.go`
- [ ] Add logic to read current-ticket.md symlink target
- [ ] Compare symlink target with the ticket being closed
- [ ] Only remove current-ticket.md if targets match
- [ ] Add unit tests for the new verification logic
- [ ] Test edge cases (no current-ticket.md, broken symlink, etc.)
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Implementation Notes
- The fix should be in the `Execute` method of the close command
- Use `os.Readlink()` to get the symlink target
- Handle cases where current-ticket.md doesn't exist or is not a symlink
- Consider adding a log message when preserving current-ticket.md because it points to a different ticket