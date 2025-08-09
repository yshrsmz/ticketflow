---
priority: 2
description: Improve error message when ticketflow close fails to suggest using ticketflow restore
created_at: "2025-08-08T23:12:14+09:00"
started_at: "2025-08-08T23:35:16+09:00"
closed_at: "2025-08-09T10:05:50+09:00"
---

# Ticket Overview

When `ticketflow close` command fails (e.g., due to missing symlink or corrupted state), the error message should be more helpful and suggest using `ticketflow restore` to recover the symlink and restore the ticket to a working state.

## Context

Currently, when `ticketflow close` encounters an error, users may not know that they can use `ticketflow restore` to recover from a broken state. This improvement will make the error message more actionable and help users resolve issues more quickly.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [x] Find where `ticketflow close` command handles errors
- [x] Identify the specific error conditions where restore would help (e.g., missing symlink, broken worktree state)
- [x] Update error messages to include suggestion about using `ticketflow restore`
- [x] Ensure the suggestion is clear and includes the correct command syntax
- [x] Test the updated error messages in various failure scenarios
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation if necessary
- [x] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Implementation Notes

The error message should be helpful and actionable, something like:
```
Error: Failed to close ticket: [original error]

Suggestion: If the ticket's symlink or worktree state is corrupted, you can try:
  ticketflow restore <ticket-id>

This will attempt to restore the ticket to a working state.
```

## Implementation Details

The improvements were made to three key areas:

1. **validateTicketForClose function** in `internal/cli/commands.go`:
   - Added detection for symlink-related errors when GetCurrentTicket fails
   - Added specific error message suggesting `ticketflow restore` when symlink issues are detected
   - Improved error messages for "No active ticket" scenario to include restore suggestion

2. **ConvertError function** in `internal/cli/error_converter.go`:
   - Updated the generic ErrTicketNotStarted conversion to include restore suggestion
   - This ensures consistent messaging across all ticket-not-started errors

3. **Error message improvements**:
   - "No active ticket" now suggests: `ticketflow restore` as an option
   - Symlink read failures now explicitly suggest restore as the recovery method
   - All suggestions are clear with exact command syntax

The actual error messages now provide:
- Clear indication of the problem
- Multiple recovery options in order of likelihood
- Exact command syntax for each suggestion

Example error output:
```
Error Code: TICKET_NOT_STARTED
Message: No active ticket
Details: There is no ticket currently being worked on

Suggestions:
  - Start a ticket first: ticketflow start <ticket-id>
  - Restore current ticket link if in a worktree: ticketflow restore
  - List available tickets: ticketflow list
```

## Notes

This improvement will enhance user experience by providing clear recovery steps when operations fail. The implementation follows the existing error handling patterns in the codebase and maintains consistency with other error messages.