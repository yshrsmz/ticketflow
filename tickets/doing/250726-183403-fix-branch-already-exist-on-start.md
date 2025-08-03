---
priority: 2
description: ""
created_at: "2025-07-26T18:34:03+09:00"
started_at: "2025-08-03T12:14:44+09:00"
closed_at: null
---

# 概要

fix the following error.

```sh
! dist/ticketflow-linux-arm64 start 250726-181410-fix-empty-status-tab
  ⎿  Error: failed to create worktree: git worktree add /workspaces/ticketflow/.worktrees/250726-181410-fix-empty-status-tab -b
     250726-181410-fix-empty-status-tab failed: exit status 255
     Preparing worktree (new branch '250726-181410-fix-empty-status-tab')
     fatal: a branch named '250726-181410-fix-empty-status-tab' already exists
```

## タスク
- [x] Add BranchExists method to check if a branch exists
- [x] Update AddWorktree to handle existing branches (use branch without -b if exists)
- [x] Add unit tests for BranchExists method
- [x] Add tests for AddWorktree with existing/non-existing branches
- [x] Improve error message when worktree already exists (show path)
- [x] Create integration tests for branch exists scenario
- [x] Create shared worktree utility package
- [x] Add comprehensive logging for debugging
- [x] Address PR review comments (lint, validation, logging)
- [x] Refactor branch validation for better readability
- [x] Create PR and ensure CI passes

## Sub-tickets Created
- [ ] 250803-121450-handle-diverged-branch - Handle case when branch points to different commit
- [ ] 250803-121506-worktree-recovery - Add worktree recovery mechanisms
- [ ] 250803-121521-robustness-improvements - Add permission and concurrency handling
- [ ] 250803-130821-force-flag-for-worktree-recreation - Add --force flag to recreate worktrees
- [ ] 250803-130903-metrics-telemetry-for-worktree-operations - Add metrics/telemetry tracking

## 技術仕様 (Core Fix Only)

### Problems
1. **Branch exists without worktree**: When `ticketflow start` is called and the branch exists but worktree doesn't, it fails with "fatal: a branch named '<branch>' already exists"
2. **Worktree already exists**: Currently returns an error, but could be handled more gracefully

### Root Causes
1. Branch can exist from previous incomplete cleanup or manual branch creation
2. The `git worktree add -b` command always tries to create a new branch
3. Worktree existence check prevents restarting work on a ticket

### Solution Design

#### 1. Add Branch Existence Check
Create new method in `internal/git/git.go`:
```go
func (g *Git) BranchExists(ctx context.Context, branch string) (bool, error)
```
Use `git show-ref --verify --quiet refs/heads/<branch>` to check branch existence

#### 2. Update AddWorktree to Handle Existing Branches
Modify `internal/git/worktree.go`:
- Check if branch exists before creating worktree
- If branch exists: `git worktree add <path> <branch>` (without -b)
- If branch doesn't exist: `git worktree add <path> -b <branch>` (current behavior)

#### 3. Improve Error Handling
- Enhance error messages in `checkExistingWorktree` to show worktree path
- Provide clear suggestions for users when errors occur
- Add option to force recreate if needed

### Implementation Steps
1. Implement `BranchExists` method with proper error handling
2. Update `AddWorktree` to conditionally use `-b` flag
3. Write unit tests for `BranchExists`
4. Write tests for `AddWorktree` with existing/non-existing branches
5. Add integration tests for the complete workflow
6. Update error messages to be more helpful

### Core Fix Focus
This ticket focuses only on fixing the immediate "branch already exists" error. Edge cases are handled in sub-tickets:
- **250803-121450-handle-diverged-branch**: Branch pointing to different commit
- **250803-121506-worktree-recovery**: Corrupted worktree references  
- **250803-121521-robustness-improvements**: Permission issues and concurrent operations

## Implementation Summary

### Changes Made

#### 1. Added BranchExists method to git package (`internal/git/git.go`)
- New method `BranchExists(ctx context.Context, branch string) (bool, error)` 
- Uses `git show-ref --verify --quiet refs/heads/<branch>` to check if a branch exists
- Returns false when branch doesn't exist (expected behavior)
- Returns error only for unexpected failures

#### 2. Updated AddWorktree to handle existing branches (`internal/git/worktree.go`)
- Modified `AddWorktree` to check if branch exists before attempting to create worktree
- If branch exists: uses `git worktree add <path> <branch>` (without -b flag)
- If branch doesn't exist: uses `git worktree add <path> -b <branch>` (original behavior)

#### 3. Added git command constants (`internal/git/constants.go`)
- Added `SubcmdShowRef = "show-ref"` 
- Added `FlagVerify = "--verify"`
- Added `FlagQuiet = "--quiet"`

#### 4. Improved error messages for existing worktrees
- Enhanced error messages in `checkExistingWorktree` (CLI) to show worktree path
- Enhanced error messages in `setupTicketBranchOrWorktree` (TUI) to show worktree path
- Now displays: "Worktree for ticket <id> already exists at: <path>"

#### 5. Added comprehensive tests
- Unit tests for `BranchExists` method in `internal/git/git_test.go`
- Tests for `AddWorktree` with existing/non-existing branches in `internal/git/worktree_test.go`
- Updated mock expectations in `internal/cli/commands_helpers_test.go`

### Testing
All tests pass:
- Unit tests: `go test ./internal/git`
- CLI tests: `go test ./internal/cli`
- Integration tests: `go test ./test/integration`
- Full test suite: `make test`

### How it works
When `ticketflow start` is called:
1. It checks if a worktree already exists for the ticket
2. If not, it attempts to create a new worktree
3. The `AddWorktree` method now:
   - Checks if the branch already exists using `BranchExists`
   - If yes: creates worktree pointing to existing branch
   - If no: creates worktree with a new branch

This allows users to restart work on a ticket even if the branch already exists from a previous incomplete cleanup or manual branch creation.

## Additional Changes from Code Reviews

### Code Review Iterations
The PR went through multiple code review iterations with the golang-pro agent, resulting in significant improvements:

#### 1. Code Duplication Elimination
- Created shared `internal/worktree/utils.go` package
- Extracted common worktree path resolution logic
- Added comprehensive logging with fallback scenarios

#### 2. Branch Name Validation
- Initially used complex regex for validation
- Added security measures to prevent command injection
- Refactored to character-by-character validation for readability
- Added Unicode support for international branch names

#### 3. Error Handling Improvements
- Added error logging when worktree lookup fails
- Enhanced error messages to show actual worktree paths
- Improved debugging capabilities with structured logging

### PR Review Response
Addressed all PR review comments:
- Fixed lint errors (unchecked error returns)
- Added branch name validation for security
- Added error logging for debugging
- All CI checks passing (lint and tests)

## Insights and Learnings

### 1. Git Worktree Edge Cases
- Branches can exist without worktrees due to incomplete cleanup
- The `-b` flag fails when branch already exists
- Worktree state can become inconsistent with git state

### 2. Code Review Value
- Multiple review iterations led to significantly better code
- Extracted shared utilities improved maintainability
- Security considerations (command injection) were identified

### 3. Readability vs Complexity Trade-offs
- Complex regex was secure but hard to understand
- Character-by-character validation is more readable
- Helper functions improve code organization
- Clear comments are essential for validation logic

### 4. Test Coverage Importance
- Integration tests caught real-world scenarios
- Unit tests ensure edge cases are handled
- Mock updates required careful consideration

### 5. PR Management
- Created sub-tickets for edge cases to keep PR focused
- Deferred nice-to-have improvements for future work
- Balance between perfection and shipping working code

## メモ

The fix has been implemented, tested, and refined through multiple code review iterations. PR #32 is ready for merge with all CI checks passing. The implementation is production-ready with:
- Comprehensive test coverage
- Security considerations addressed
- Improved code readability
- Better error messages and logging
- Shared utilities for code reuse

Future enhancements tracked in sub-tickets for --force flag and metrics/telemetry.
