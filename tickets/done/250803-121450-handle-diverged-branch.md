---
priority: 3
description: Handle case when branch exists but points to different commit
created_at: "2025-08-03T12:14:50+09:00"
started_at: "2025-08-05T13:29:41+09:00"
closed_at: "2025-08-06T14:30:32+09:00"
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
- [x] Address all PR review comments
- [x] Implement non-interactive mode for CI/CD
- [x] Fix all lint and test failures
- [x] Add comprehensive error handling and recovery
- [x] Document non-interactive mode behavior

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

## Code Review and Fixes Applied

After review by golang-pro agent, the following improvements were made:

1. **Added input validation** to `BranchDivergedFrom` method to validate branch names
2. **Improved error messages** in prompt utility to show valid options when invalid input is provided
3. **Added error recovery** when branch recreation fails - attempts to recreate the deleted branch
4. **Removed unused code** - cleaned up unused `simulateUserInput` test helper
5. **Added placeholder for debug logging** when branch divergence is detected

## Insights and Learnings

### 1. **Branch State Management Complexity**
Working on this feature revealed the complexity of managing branch states in a worktree-based workflow. When a branch exists but has diverged, there are multiple valid user intentions:
- They may want to continue with their existing work (use existing)
- They may want to start fresh (recreate)
- They may have created the branch accidentally (cancel)

### 2. **Error Recovery Importance**
The review highlighted a critical error path: if we delete a branch but fail to create the worktree, we've left the repository in a worse state. Adding recovery logic to recreate the branch in this case prevents data loss.

### 3. **User Experience Considerations**
The interactive prompt needs to be both informative and safe:
- Clear information about the divergence (X commits ahead, Y behind)
- Safe defaults (recreate is default to avoid accidentally using wrong branch)
- Clear error messages with valid options listed

### 4. **Testing Challenges**
Integration testing interactive prompts is challenging because:
- Standard input cannot be easily mocked in integration tests
- The test verifies the error message rather than the full flow
- Consider future refactoring to make prompts more testable (e.g., injectable reader)

### 5. **Git Command Security**
Branch name validation is critical to prevent command injection. The existing `isValidBranchName` function provides good protection, but it must be used consistently across all methods that accept branch names as parameters.

### 6. **Performance Considerations**
The implementation is efficient:
- Uses `git rev-list --count` for fast commit counting
- Only checks divergence when branch already exists
- Avoids unnecessary git operations

### 7. **CI/CD Integration Challenges**
The integration with CI systems revealed important considerations:
- Interactive prompts fail in non-interactive environments (no stdin)
- Environment variable detection is more reliable than terminal detection
- Default behavior must be safe and predictable in automated contexts
- Tests need to handle both interactive and non-interactive modes

### 8. **Code Review Value**
The PR review process uncovered several improvements:
- Edge cases in git operations (repos without origin, single commit repos)
- Consistency in using defined constants vs hardcoded strings
- Security considerations even with validated inputs
- The importance of explaining design decisions in comments

## Future Improvements

1. **Configuration Option**: Add `divergenceStrategy` config to allow automatic handling without prompts
2. **Timeout Handling**: Add timeout for interactive prompts to prevent hanging
3. **Better Testability**: Refactor prompt utility to accept a reader interface for easier testing
4. **Logging**: Add structured logging for debugging branch divergence detection
5. **Branch Cleanup**: Consider adding a command to clean up diverged branches in bulk

## PR Review Comments and Fixes

### Initial Review Comments (All Addressed ✅)
1. **Parse errors for commit counts** - Changed to return errors instead of silently ignoring
2. **Rollback logic safety** - Added check for HEAD^ existence before attempting reset  
3. **Testability improvement** - Refactored prompt functions to accept io.Reader parameter
4. **Documentation** - Added comment about git version requirements for --initial-branch flag

### Additional Review Comments (All Addressed ✅)
1. **GetDefaultBranch error handling** - Added check for origin remote existence and fallback to git config init.defaultBranch
2. **Git constants usage** - Updated rollback to use proper git constants (SubcmdReset, FlagHard)
3. **Security clarification** - Added comments explaining branch name validation prevents injection
4. **Test design rationale** - Explained why integration test uses raw git commands

### CI/CD Fixes Applied
1. **Lint errors fixed**:
   - Fixed ineffectual assignment to `err` in worktree.go
   - Fixed unchecked os.Setenv/Unsetenv returns in tests
   - Applied go fmt formatting

2. **Test failures fixed**:
   - Implemented non-interactive mode for CI environments
   - Added automatic detection of CI (via environment variables)
   - Tests now use default options in non-interactive mode
   - Created comprehensive documentation for non-interactive mode

### Key Implementation Highlights

#### Non-Interactive Mode
Added intelligent detection and handling for non-interactive environments:
- Detects CI environments (CI, GITHUB_ACTIONS, GITLAB_CI, etc.)
- Supports TICKETFLOW_NON_INTERACTIVE environment variable
- Uses terminal detection as fallback
- Automatically selects default options in prompts

#### Error Handling Improvements
- GetDefaultBranch now handles multiple edge cases gracefully
- Proper error propagation for parse failures
- Safe rollback with HEAD^ existence check

## Status

The feature is complete with all PR review comments addressed and CI fully passing. The implementation includes:
- ✅ Branch divergence detection and user prompting
- ✅ Non-interactive mode for CI/CD environments  
- ✅ Comprehensive error handling and recovery
- ✅ Full test coverage with passing CI
- ✅ All code review feedback incorporated

PR #40 is ready for final review and merge.