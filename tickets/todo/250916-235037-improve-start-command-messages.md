---
priority: 2
description: "Improve start command messages and error hints for doing tickets"
created_at: "2025-09-16T23:50:37+09:00"
started_at: null
closed_at: null
---

# Improve Start Command Messages for Doing Tickets

## Overview

The `ticketflow start` command already supports creating/recreating worktrees for tickets in "doing" status when using the `--force` flag. However, the user experience can be improved:

1. The error message when attempting to start a doing ticket without `--force` doesn't mention that `--force` can be used to create a worktree
2. The success message incorrectly shows "Status: todo → doing" even when the ticket was already in doing status

## Tasks

- [ ] Update error message in `validateTicketForStart` to suggest using `--force` for worktree creation
- [ ] Add `OriginalStatus` field to `StartTicketResult` struct to track the ticket's status before the operation
- [ ] Add `IsRecreatingWorktree` field to `StartTicketResult` to distinguish between creating vs recreating
- [ ] Update `StartTicket` method to capture and pass the original status
- [ ] Fix status display in `printable.go` to show correct status transition (e.g., "doing → doing (worktree recreated)")
- [ ] Update output messages to distinguish between "Worktree created" vs "Worktree recreated"
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Add/update tests for the new functionality
- [ ] Update documentation if necessary
- [ ] Get developer approval before closing

## Technical Details

### Files to modify:
1. `internal/cli/commands.go`:
   - Line 38: Add new fields to `StartTicketResult`
   - Line 452: Capture original status
   - Line 500-505: Set new fields in result
   - Line 993-994: Update error message with --force suggestion

2. `internal/cli/printable.go`:
   - Line 357: Conditionally show "created" vs "recreated"
   - Line 362: Show dynamic status transition
   - Line 378: Similar changes for branch mode

## Notes

This improvement was identified when a user wanted to create a worktree for a ticket that was already in "doing" status (e.g., after the worktree was accidentally deleted). The functionality already exists with `--force`, but it's not discoverable without reading the code.