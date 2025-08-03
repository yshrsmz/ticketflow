# Fix for "branch already exists" error when starting work on a ticket

## Summary of Changes

This fix addresses the issue where `ticketflow start` would fail with "fatal: a branch named '<branch>' already exists" when attempting to create a worktree for a branch that already exists in the repository.

## Changes Made

### 1. Added BranchExists method to git package (`internal/git/git.go`)
- New method `BranchExists(ctx context.Context, branch string) (bool, error)` 
- Uses `git show-ref --verify --quiet refs/heads/<branch>` to check if a branch exists
- Returns false when branch doesn't exist (expected behavior)
- Returns error only for unexpected failures

### 2. Updated AddWorktree to handle existing branches (`internal/git/worktree.go`)
- Modified `AddWorktree` to check if branch exists before attempting to create worktree
- If branch exists: uses `git worktree add <path> <branch>` (without -b flag)
- If branch doesn't exist: uses `git worktree add <path> -b <branch>` (original behavior)

### 3. Added git command constants (`internal/git/constants.go`)
- Added `SubcmdShowRef = "show-ref"` 
- Added `FlagVerify = "--verify"`
- Added `FlagQuiet = "--quiet"`

### 4. Improved error messages for existing worktrees
- Enhanced error messages in `checkExistingWorktree` (CLI) to show worktree path
- Enhanced error messages in `setupTicketBranchOrWorktree` (TUI) to show worktree path
- Now displays: "Worktree for ticket <id> already exists at: <path>"

### 5. Added comprehensive tests
- Unit tests for `BranchExists` method in `internal/git/git_test.go`
- Tests for `AddWorktree` with existing/non-existing branches in `internal/git/worktree_test.go`
- Updated mock expectations in `internal/cli/commands_helpers_test.go`

## Testing

All tests pass:
- Unit tests: `go test ./internal/git`
- CLI tests: `go test ./internal/cli`
- Integration tests: `go test ./test/integration`
- Full test suite: `make test`

## How it works

When `ticketflow start` is called:
1. It checks if a worktree already exists for the ticket
2. If not, it attempts to create a new worktree
3. The `AddWorktree` method now:
   - Checks if the branch already exists using `BranchExists`
   - If yes: creates worktree pointing to existing branch
   - If no: creates worktree with a new branch

This allows users to restart work on a ticket even if the branch already exists from a previous incomplete cleanup or manual branch creation.