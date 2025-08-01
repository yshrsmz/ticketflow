---
priority: 2
description: Fix stale branch detection in cleanup command - branches for done tickets not being identified
created_at: "2025-07-29T22:14:02+09:00"
started_at: "2025-07-29T22:15:35+09:00"
closed_at: "2025-07-30T00:39:21+09:00"
---

# Ticket Overview

The `ticketflow cleanup` command is not detecting stale branches properly. After completing tickets and moving them to done/, their associated git branches remain but are not identified as stale by the cleanup command. The cleanup output shows "Cleaned 0 stale branch(es)" even when there are branches for done tickets that should be cleaned up.

## Problem Details

Observed behavior:
- Three tickets were completed and moved to done/:
  - 250729-105128-implement-two-column-id-display
  - 250729-105204-implement-responsive-id-column-width
  - 250729-105236-implement-tui-display-mode-toggle
- Their worktrees were successfully cleaned ("Cleaned 3 orphaned worktree(s)")
- However, their git branches still exist but show "Cleaned 0 stale branch(es)"

Expected behavior:
- The cleanup command should identify these branches as stale since their tickets are in done/
- It should offer to delete these local branches

## Investigation Points

1. The `cleanStaleBranches` function in `internal/cli/cleanup.go` checks if branch names match ticket IDs
2. It's possible the branch names don't exactly match the ticket IDs
3. Need to verify:
   - What branches actually exist (`git branch`)
   - How they compare to ticket IDs
   - Whether the ticket ID lookup is working correctly

## Tasks
- [x] List all local branches and compare with ticket IDs to identify the mismatch
- [x] Debug the cleanStaleBranches function to see why branches aren't matching
- [x] Fix the branch name matching logic if needed
- [x] Consider if branch naming convention has changed or is inconsistent
- [x] Add logging/debug output to help diagnose similar issues in the future
- [x] Test the fix with multiple done tickets
- [x] Run `make test` to ensure no regressions
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Update tests if the branch matching logic changes
- [x] Document any changes to branch naming conventions
- [x] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Notes

The issue might be related to:
- Branch naming format vs ticket ID format
- Case sensitivity in matching
- Special characters or formatting differences
- The ticket status lookup not working as expected

## Resolution

The issue was identified in the `cleanStaleBranches` function in `internal/cli/cleanup.go`. The function was calling `app.Manager.List("")` which, according to the `getDirectoriesForStatus` function logic, only returns tickets from todo and doing directories when given an empty string.

The fix was simple: change the parameter from `""` to `"all"` to include done tickets:
- `app.Manager.List("")` → `app.Manager.List("all")`

This change was made in two places:
1. In `cleanStaleBranches` function (line 108)
2. In `CleanupStats` function (line 187)

After the fix, the cleanup command correctly identifies and removes branches for done tickets. The issue was not related to branch naming or matching logic, but simply that done tickets weren't being loaded for comparison.

