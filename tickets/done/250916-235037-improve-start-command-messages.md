---
priority: 2
description: Improve start command messages and error hints for doing tickets
created_at: "2025-09-16T23:50:37+09:00"
started_at: "2025-09-17T00:01:13+09:00"
closed_at: "2025-09-17T13:53:08+09:00"
---

# Improve Start Command Messages for Doing Tickets

## Overview

The `ticketflow start` command already supports creating/recreating worktrees for tickets in "doing" status when using the `--force` flag. However, the user experience can be improved:

1. The error message when attempting to start a doing ticket without `--force` doesn't mention that `--force` can be used to create a worktree
2. The success message incorrectly shows "Status: todo → doing" even when the ticket was already in doing status

## Tasks

### Initial Implementation (Completed)
- [x] Update error message in `validateTicketForStart` to suggest using `--force` for worktree creation
- [x] Add `OriginalStatus` field to `StartTicketResult` struct to track the ticket's status before the operation
- [x] Add `IsRecreatingWorktree` field to `StartTicketResult` to distinguish between creating vs recreating
- [x] Update `StartTicket` method to capture and pass the original status
- [x] Fix status display in `printable.go` to show correct status transition (e.g., "doing → doing (worktree recreated)")
- [x] Update output messages to distinguish between "Worktree created" vs "Worktree recreated"
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Add/update tests for the new functionality

### Follow-up Refinements (Completed)
- [x] Gate the `--force` suggestion behind worktree-enabled check
- [x] Add branch mode-specific suggestions when worktrees are disabled
- [x] Remove "branch recreated" language in branch mode (simplified to just show status transition)
- [x] Add defensive fallback for empty OriginalStatus (defaults to "todo")
- [x] Add test coverage for empty OriginalStatus fallback
- [x] Add test for worktree recreation text output
- [x] Extend StructuredData test to verify new fields
- [x] Add integration test for branch mode suggestions

### Test Isolation Fix (Completed)
- [x] Fixed critical test isolation issue where tests modified the main repository when run from git hooks in worktrees
- [x] Implemented `env -u` solution in pre-push and pre-commit hooks to clean git environment variables
- [x] Reverted unsuccessful GIT_CEILING_DIRECTORIES approach that didn't work with worktrees
- [x] Restored test parallelization by removing unnecessary t.Setenv() calls

### Pending
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

## Implementation Summary

### Files Modified:
1. **`internal/cli/commands.go`**:
   - Added `OriginalStatus` and `IsRecreatingWorktree` fields to `StartTicketResult` struct
   - Updated `StartTicket` to capture original status and determine if recreating worktree
   - Modified `validateTicketForStart` to provide context-aware suggestions based on worktree mode

2. **`internal/cli/printable.go`**:
   - Added defensive fallback for empty `OriginalStatus` (defaults to `todo`)
   - Shows "Worktree recreated" vs "Worktree created" based on `IsRecreatingWorktree`
   - Displays accurate status transitions (e.g., "doing → doing (worktree recreated)")
   - Removed confusing "branch recreated" language in branch mode
   - Added new fields to `StructuredData()` for JSON output

3. **`internal/cli/printable_test.go`**:
   - Added test for empty OriginalStatus fallback behavior
   - Added test for worktree recreation text output
   - Extended StructuredData test to verify new fields are included

4. **`test/integration/worktree_force_test.go`**:
   - Enhanced tests to verify error messages include appropriate suggestions
   - Added assertions for branch mode suggestions when worktrees disabled

## Insights and Learnings

### 1. Context-Aware Error Messages
The implementation revealed the importance of providing different suggestions based on the user's configuration. When worktrees are disabled, suggesting `--force` would be misleading since it doesn't work in branch mode. Instead, we provide actionable alternatives like `git checkout` and `ticketflow status`.

### 2. Defensive Programming
Adding a fallback for empty `OriginalStatus` prevents potential display issues. While this shouldn't happen in normal operation, having the fallback ensures the UI remains consistent even if there are unexpected edge cases.

### 3. User Experience Considerations
- The "branch recreated" language was removed because it was confusing - in branch mode, we're just switching branches, not recreating anything
- Distinguishing between "created" and "recreated" for worktrees helps users understand what actually happened
- Omitting the commit message for recreations avoids confusion (no commit is made when recreating)

### 4. Testing Strategy
The integration tests proved valuable in catching edge cases, particularly around the different behavior between worktree and branch modes. The unit tests ensure the display logic works correctly for all scenarios.

### 5. Code Review Feedback Integration
The golang-pro review identified the missing JSON fields in `StructuredData()`, which was important for maintaining consistency in the API output. This highlights the value of thorough code review in catching completeness issues.

### 6. Critical Test Isolation Issue in Worktrees
Discovered a critical issue where tests run from git hooks within a worktree would inherit git environment variables (GIT_DIR, GIT_WORK_TREE, GIT_COMMON_DIR) pointing to the parent repository. This caused tests to modify the actual repository instead of creating isolated test environments.

**Why GIT_CEILING_DIRECTORIES didn't work**: The `.git` file in worktrees (which contains `gitdir: /path/to/parent/.git/worktrees/name`) allows git to discover the parent repository even with GIT_CEILING_DIRECTORIES set. This is because git reads the `.git` file directly before checking ceiling directories.

**The Solution**: Using `env -u` to unset git environment variables before running tests ensures they start with a clean environment. This is a minimal, surgical fix that:
- Prevents tests from accessing the parent repository
- Allows tests to create proper isolated test repositories
- Has no side effects on the hook's environment after tests complete
- Works consistently across different platforms

### 7. Git Hook Script Improvements
Updated both pre-commit and pre-push hooks to use `env -u GIT_DIR -u GIT_WORK_TREE -u GIT_COMMON_DIR` before running make commands. This ensures any git operations within the make targets (like tests) don't inherit the worktree context.

## Notes

This improvement was identified when a user wanted to create a worktree for a ticket that was already in "doing" status (e.g., after the worktree was accidentally deleted). The functionality already exists with `--force`, but it's not discoverable without reading the code.

The implementation successfully makes this hidden feature discoverable while also improving the accuracy of status messages throughout the application.
