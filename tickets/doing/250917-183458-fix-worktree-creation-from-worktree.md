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

## Implementation Plan

1. **Modify `FindProjectRoot()` function**:
   - Check if current directory is a worktree
   - If in worktree, find and return the main repository path
   - If not in worktree, use current behavior

2. **Add helper functions**:
   - `IsWorktree()` - Detect if current path is a worktree
   - `GetMainRepoPath()` - Get the main repository path from a worktree

3. **Update worktree path calculation**:
   - Ensure `GetWorktreePath()` always uses the main repository as the base

## Tasks

- [ ] Research git commands for worktree detection and main repo path discovery
- [ ] Implement `IsWorktree()` helper function in `internal/git/git.go`
- [ ] Implement `GetMainRepoPath()` helper function
- [ ] Update `FindProjectRoot()` to handle worktree scenarios
- [ ] Add unit tests for new helper functions
- [ ] Add integration test for creating worktree from within a worktree
- [ ] Test manual workflow: create ticket A, start it, then from A's worktree create and start ticket B
- [ ] Run `make test` to verify all tests pass
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update the ticket with implementation insights and any edge cases discovered
- [ ] Get developer approval before closing

## Testing Scenarios

1. **From main repository**: `ticketflow start ticket-1` should create worktree at `../ticketflow.worktrees/ticket-1`
2. **From worktree**: Navigate to `../ticketflow.worktrees/ticket-1`, then `ticketflow start ticket-2` should create at `../ticketflow.worktrees/ticket-2` (not nested)
3. **Nested worktrees**: Ensure the fix handles multiple levels correctly

## Notes

- This is a critical bug that affects the usability when working with multiple tickets
- The fix should be backwards compatible with existing worktree structures
- Consider adding a migration path if users have accidentally created nested worktrees