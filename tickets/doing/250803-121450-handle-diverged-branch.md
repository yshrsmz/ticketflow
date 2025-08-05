---
priority: 3
description: Handle case when branch exists but points to different commit
created_at: "2025-08-03T12:14:50+09:00"
started_at: "2025-08-05T13:29:41+09:00"
closed_at: null
related:
    - parent:250726-183403-fix-branch-already-exist-on-start
---

# Handle Diverged Branch

## Overview
When starting a ticket, if a branch already exists but points to a different commit than expected (e.g., not at the default branch HEAD), we need to provide clear options to the user.

## Tasks
- [x] Add method to check if branch diverged from expected base
- [x] Implement interactive prompt for user choice
- [x] Add option to use existing branch
- [x] Add option to delete and recreate branch
- [x] Add tests for diverged branch scenarios

## Technical Details
- Compare branch HEAD with default branch HEAD using `git rev-parse`
- Show clear information about the divergence (commits ahead/behind)
- Implement user choice handling with proper error recovery

## Implementation Design

### Problem Analysis
Currently, when `ticketflow start` is called and a branch already exists, the `AddWorktree` method in `internal/git/worktree.go` simply uses the existing branch without checking if it points to the expected commit. This can lead to confusion if:
1. The branch was created manually or by another process
2. The branch was left from a previous attempt and has diverged
3. The branch points to an older commit

### Solution Design (Updated with Review Feedback)

#### 1. Add Branch Divergence Detection (in `internal/git/git.go`)
```go
// GetDefaultBranch returns the configured default branch (main/master)
func (g *Git) GetDefaultBranch(ctx context.Context) (string, error)

// BranchDivergedFrom checks if a branch has diverged from a base branch
func (g *Git) BranchDivergedFrom(ctx context.Context, branch, baseBranch string) (bool, error)

// GetBranchCommit gets the commit hash a branch points to
func (g *Git) GetBranchCommit(ctx context.Context, branch string) (string, error)

// GetBranchDivergenceInfo returns commits ahead/behind between branches
func (g *Git) GetBranchDivergenceInfo(ctx context.Context, branch, baseBranch string) (ahead, behind int, error)
```

Implementation notes:
- Use `git rev-parse origin/HEAD` to detect default branch properly
- Validate branch names using existing `isValidBranchName` function
- Use `git rev-list --count` for accurate ahead/behind counts
- Add constants to `constants.go`: `SubcmdRevList = "rev-list"`, `FlagCount = "--count"`

#### 2. Enhanced Error Types (in `internal/errors/errors.go`)
```go
// Sentinel error
var ErrBranchDiverged = errors.New("branch has diverged from expected base")

// BranchDivergenceError provides detailed divergence information
type BranchDivergenceError struct {
    Branch     string
    BaseBranch string
    Ahead      int
    Behind     int
}

func (e *BranchDivergenceError) Error() string
func (e *BranchDivergenceError) Is(target error) bool
func NewBranchDivergenceError(branch, baseBranch string, ahead, behind int) error
```

#### 3. Create Reusable Prompt Utility (new file: `internal/cli/prompt.go`)
```go
// PromptOption represents a choice in a prompt
type PromptOption struct {
    Key         string
    Description string
    IsDefault   bool
}

// Prompt displays options and returns selected key
func Prompt(message string, options []PromptOption) (string, error)

// ConfirmPrompt displays a yes/no prompt
func ConfirmPrompt(message string, defaultYes bool) bool
```

#### 4. Update AddWorktree Logic (in `internal/git/worktree.go`)
- Get default branch using `GetDefaultBranch()` instead of current branch
- Check divergence when branch exists
- Return `BranchDivergenceError` with details when diverged
- Continue with existing behavior if not diverged

#### 5. Handle Divergence in CLI (in `internal/cli/commands.go`)
Add new method:
```go
func (app *App) handleBranchDivergence(ctx context.Context, t *ticket.Ticket, 
    worktreePath string, divergenceErr *ticketerrors.BranchDivergenceError) (string, error)
```

Update `createAndSetupWorktree`:
- Use `errors.As` to check for `BranchDivergenceError`
- Call `handleBranchDivergence` when detected
- Show clear divergence info (commits ahead/behind)
- Present three options: use existing, recreate, cancel
- Handle each option appropriately

#### 6. Configuration Enhancement (optional future improvement)
Consider adding to `.ticketflow.yaml`:
```yaml
worktree:
  divergenceStrategy: "prompt" # or "use-existing", "recreate"
```

#### 7. Update Interfaces (in `internal/git/interfaces.go`)
Add new methods to `GitClient` interface:
```go
type GitClient interface {
    WorktreeClient
    GetDefaultBranch(ctx context.Context) (string, error)
    BranchDivergedFrom(ctx context.Context, branch, baseBranch string) (bool, error)
    GetBranchCommit(ctx context.Context, branch string) (string, error)
    GetBranchDivergenceInfo(ctx context.Context, branch, baseBranch string) (ahead, behind int, err error)
}

```

### Testing Strategy

#### 1. Unit Tests (in `internal/git/git_test.go`)
```go
func TestGetDefaultBranch(t *testing.T)
func TestBranchDivergedFrom(t *testing.T)
func TestGetBranchCommit(t *testing.T)
func TestGetBranchDivergenceInfo(t *testing.T)
```

Use table-driven tests:
```go
tests := []struct {
    name       string
    setupFunc  func(t *testing.T, g *Git) // Setup git state
    branch     string
    baseBranch string
    wantAhead  int
    wantBehind int
    wantErr    bool
}
```

#### 2. Integration Tests (in `test/integration/branch_divergence_test.go`)
Test scenarios:
- Branch exists at expected commit (should proceed without prompt)
- Branch exists but ahead of base (should prompt with divergence info)
- Branch exists but behind base (should prompt with divergence info)
- Branch exists with different history (should prompt)
- User selects "use existing" option
- User selects "recreate" option
- User selects "cancel" option
- Invalid branch names are rejected

#### 3. Mock Tests (in `internal/cli/commands_test.go`)
- Mock GitClient to simulate divergence scenarios
- Test prompt handling without actual git operations
- Verify error handling and rollback behavior

#### 4. Logging and Debugging
Add structured logging:
```go
logger.Debug("checking branch divergence",
    slog.String("branch", branch),
    slog.String("baseBranch", baseBranch),
    slog.Int("ahead", ahead),
    slog.Int("behind", behind))
```

## Acceptance Criteria
- User gets clear, informative prompt when branch has diverged
- Divergence information shows exact commits ahead/behind
- All three options work correctly:
  - Use existing: preserves branch history
  - Recreate: deletes and creates fresh branch
  - Cancel: rolls back the operation cleanly
- Proper error handling with descriptive messages
- No git operations leave repository in inconsistent state
- Tests cover all divergence scenarios
- Branch name validation prevents command injection

## Implementation Summary

Successfully implemented branch divergence detection and handling with the following changes:

1. **Git Methods Added** (in `internal/git/git.go`):
   - `GetDefaultBranch()` - Detects main/master branch with fallback
   - `GetBranchCommit()` - Gets commit hash for a branch
   - `BranchDivergedFrom()` - Checks if branches have diverged
   - `GetBranchDivergenceInfo()` - Returns commits ahead/behind

2. **Error Handling** (in `internal/errors/errors.go`):
   - Added `ErrBranchDiverged` sentinel error
   - Created `BranchDivergenceError` type with detailed info
   - Implements `Is()` method for proper error matching

3. **Interactive Prompt** (new file `internal/cli/prompt.go`):
   - `Prompt()` - Displays options and handles user choice
   - `ConfirmPrompt()` - Yes/no confirmation helper
   - Supports default options

4. **Worktree Updates** (in `internal/git/worktree.go`):
   - `AddWorktree()` now checks for branch divergence
   - Returns `BranchDivergenceError` when divergence detected

5. **CLI Integration** (in `internal/cli/commands.go`):
   - `handleBranchDivergence()` - Shows divergence info and handles user choice
   - Three options: use existing, recreate, or cancel
   - Proper rollback on cancellation

6. **Interface Updates** (in `internal/git/interfaces.go`):
   - Added new methods to `GitClient` interface
   - Added `BranchExists` to `BasicGitClient`

7. **Tests Added**:
   - Unit tests for all new git methods
   - Integration tests for divergence scenarios
   - Mock implementation updated for testing

The implementation ensures users are informed when a branch has diverged and can make an informed choice about how to proceed.