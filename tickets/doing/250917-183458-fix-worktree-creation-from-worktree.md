---
priority: 2
description: Fix worktree base directory calculation when running from within a worktree
created_at: "2025-09-17T18:34:58+09:00"
started_at: "2025-09-17T18:38:24+09:00"
closed_at: null
---

# Fix Worktree Creation Path When Running From Worktree

## Problem

When running `ticketflow start <ticket-id>` from within an existing worktree, the new worktree base directory is calculated relative to the current worktree instead of the main repository. This creates nested worktree structures like:

```
../ticketflow.worktrees/existing-ticket/../ticketflow.worktrees/new-ticket
```

Instead of the expected:
```
../ticketflow.worktrees/new-ticket
```

## Root Cause Analysis

The issue occurs because:
1. `git rev-parse --show-toplevel` returns the worktree's path when run from within a worktree
2. `FindProjectRoot()` in `internal/git/git.go` uses this command to determine the project root
3. `GetWorktreePath()` then calculates the worktree base directory relative to this incorrect root
4. This causes new worktrees to be created relative to the current worktree, not the main repository

## Solution

We need to detect if we're in a worktree and find the main repository path. Git provides commands to identify:
- If we're in a worktree: `git rev-parse --git-common-dir` returns different path than `--git-dir` in worktrees
- Main repository path: `git worktree list` shows the main worktree (first entry) or use `--git-common-dir` to find .git directory

## Implementation (Completed)

The fix has been implemented with a different approach than originally planned:

1. **Added `FindMainRepositoryRoot()` function**:
   - New function that uses `--git-common-dir` to always get the main repository path
   - Returns the parent directory of the common .git directory
   - Handles both regular repos and worktrees correctly

2. **Kept `FindProjectRoot()` unchanged**:
   - Still uses `--show-toplevel` for backward compatibility
   - Returns the current project root (may be a worktree)

3. **Updated integration points**:
   - `cli/commands.go`: Uses `FindMainRepositoryRoot()` for worktree base path calculation
   - `ui/app.go`: Uses `FindMainRepositoryRoot()` for UI worktree operations
   - Both have fallback logic if the new function fails

## Tasks

- [x] Research git commands for worktree detection and main repo path discovery
- [x] ~~Implement `IsWorktree()` helper function~~ (Not needed with chosen approach)
- [x] ~~Implement `GetMainRepoPath()` helper function~~ (Implemented as `FindMainRepositoryRoot()`)
- [x] ~~Update `FindProjectRoot()` to handle worktree scenarios~~ (Kept unchanged for compatibility)
- [x] Add new `FindMainRepositoryRoot()` function using `--git-common-dir`
- [x] Update `cli/commands.go` to use `FindMainRepositoryRoot()` for worktree paths
- [x] Update `ui/app.go` to use `FindMainRepositoryRoot()` for worktree paths
- [x] Add unit test `TestFindMainRepositoryRoot_FromWorktree`
- [x] Add integration test `TestStartTicketFromWithinWorktree` for nested worktree scenario
- [x] Test manual workflow: create ticket A, start it, then from A's worktree create and start ticket B
- [x] Run `make test` to verify all tests pass
- [x] Run `make fmt` to fix code formatting
- [ ] Run `make vet` and `make lint`
- [x] Update the ticket with implementation insights and any edge cases discovered
- [ ] Get developer approval before closing

## Testing Scenarios

1. **From main repository**: `ticketflow start ticket-1` should create worktree at `../ticketflow.worktrees/ticket-1`
2. **From worktree**: Navigate to `../ticketflow.worktrees/ticket-1`, then `ticketflow start ticket-2` should create at `../ticketflow.worktrees/ticket-2` (not nested)
3. **Nested worktrees**: Ensure the fix handles multiple levels correctly

## Implementation Notes

- This is a critical bug that affects the usability when working with multiple tickets
- The fix is backwards compatible - `FindProjectRoot()` remains unchanged
- The new `FindMainRepositoryRoot()` function is only used for worktree base path calculation
- No migration needed as the fix prevents future nested worktrees without affecting existing ones

## Key Implementation Details

1. **Two separate functions approach**:
   - `FindProjectRoot()`: Returns current project root (may be worktree) - unchanged for compatibility
   - `FindMainRepositoryRoot()`: Always returns main repository root - new function for worktree paths

2. **Why this approach is better**:
   - Maintains 100% backward compatibility
   - Clear separation of concerns
   - Explicit function names indicate intent
   - Graceful fallback if new function fails

3. **Testing**:
   - Unit test verifies `FindMainRepositoryRoot()` works from within a worktree
   - Integration test `TestStartTicketFromWithinWorktree` verifies end-to-end behavior
   - Test passes: worktrees created from within worktrees are siblings, not nested