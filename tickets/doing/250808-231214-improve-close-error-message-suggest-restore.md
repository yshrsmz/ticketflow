---
priority: 2
description: Improve error message when ticketflow close fails to suggest using ticketflow restore
created_at: "2025-08-08T23:12:14+09:00"
started_at: "2025-08-08T23:35:16+09:00"
closed_at: null
---

# Ticket Overview

When `ticketflow close` command fails (e.g., due to missing symlink or corrupted state), the error message should be more helpful and suggest using `ticketflow restore` to recover the symlink and restore the ticket to a working state.

## Context

Currently, when `ticketflow close` encounters an error, users may not know that they can use `ticketflow restore` to recover from a broken state. This improvement will make the error message more actionable and help users resolve issues more quickly.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [ ] Find where `ticketflow close` command handles errors
- [ ] Identify the specific error conditions where restore would help (e.g., missing symlink, broken worktree state)
- [ ] Update error messages to include suggestion about using `ticketflow restore`
- [ ] Ensure the suggestion is clear and includes the correct command syntax
- [ ] Test the updated error messages in various failure scenarios
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation if necessary
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Implementation Notes

The error message should be helpful and actionable, something like:
```
Error: Failed to close ticket: [original error]

Suggestion: If the ticket's symlink or worktree state is corrupted, you can try:
  ticketflow restore <ticket-id>

This will attempt to restore the ticket to a working state.
```

## Notes

This improvement will enhance user experience by providing clear recovery steps when operations fail.