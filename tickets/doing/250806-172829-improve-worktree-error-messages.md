---
priority: 2
description: Improve error messages for worktree-related failures to guide users to solutions
created_at: "2025-08-06T17:28:29+09:00"
started_at: "2025-08-09T10:08:03+09:00"
closed_at: null
related:
    - parent:250803-121506-worktree-recovery
---

# Improve Worktree Error Messages

## Overview
Enhance error messages when worktree operations fail to provide clear, actionable guidance to users. Instead of generic git errors, provide specific instructions on how to resolve common worktree issues.

## Tasks
- [x] Identify common worktree error patterns in git stderr output
- [x] Enhance worktree error detection in existing error converter
- [x] Add helper functions for worktree-specific error enhancement
- [x] Add unit tests for error message enhancement
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [ ] Get developer approval before closing

## Technical Details

### Common Worktree Errors to Handle

1. **Corrupted worktree reference**
   - Git error: `fatal: '<path>' is not a working tree`
   - Enhanced: `Error: Worktree appears to be corrupted\nFix: Run 'git worktree prune' to clean up corrupted references, then retry your command`

2. **Directory already exists**
   - Git error: `fatal: '<path>' already exists`
   - Enhanced: `Error: Worktree directory already exists\nFix: Remove the directory manually or use 'ticketflow cleanup' if it's an old ticket`

3. **Branch already checked out**
   - Git error: `fatal: '<branch>' is already checked out`
   - Enhanced: `Error: Branch is already checked out in another worktree\nFix: Use 'git worktree list' to find where it's checked out`

4. **Missing worktree directory**
   - Git error: `fatal: could not create work tree dir`
   - Enhanced: `Error: Cannot create worktree directory\nFix: Check permissions and disk space, or try 'git worktree prune' if references are corrupted`

### Implementation Approach

1. Enhance error conversion in `internal/cli/error_converter.go`:
   - Add worktree-specific error pattern detection in `ConvertError()` function
   - Check for `GitError` and `WorktreeError` types that contain worktree operations
   - Pattern match on git stderr output already captured in errors

2. Create helper functions for worktree error enhancement:
```go
func enhanceWorktreeError(gitErr *ticketerrors.GitError) *CLIError {
    errStr := gitErr.Err.Error()
    
    // Pattern matching for common worktree errors
    switch {
    case strings.Contains(errStr, "is not a working tree"):
        return NewError(ErrWorktreeRemoveFailed, "Worktree appears to be corrupted",
            gitErr.Error(),
            []string{
                "Run 'git worktree prune' to clean up corrupted references",
                "Then retry your command",
            })
    // ... other patterns
    }
    
    return nil // No enhancement needed
}
```

3. Leverage existing error infrastructure:
   - Use existing `CLIError` type with suggestions field
   - Integrate with current error flow (no changes needed in commands)
   - Errors automatically flow through `ConvertError()` for enhancement

### Files to Modify
- Update `internal/cli/error_converter.go` - Add worktree error pattern detection
- Update `internal/cli/error_converter_test.go` - Add tests for enhanced messages
- Verify error handling in `internal/cli/commands.go` (worktree operations)
- Verify error handling in `internal/cli/cleanup.go` (cleanup operations)

## Acceptance Criteria
- [x] Common worktree errors show helpful, actionable messages
- [x] Error messages include specific fix instructions
- [x] Original error is preserved for debugging
- [x] No regression in existing error handling
- [x] Test coverage for all enhanced error patterns

## Notes
This is a lightweight improvement that provides immediate value to users without adding complexity to the codebase. It follows the decision to keep ticketflow focused on ticket management while helping users resolve git issues themselves.

## Implementation Insights

### Key Discoveries
1. **Existing Infrastructure**: The codebase already had a robust error handling system with `ConvertError()` function that transforms internal errors to user-friendly CLI errors. This made integration straightforward.

2. **Error Type Fields**: Had to correct initial assumptions about error struct fields:
   - `GitError` uses `Branch` field, not `Ref`
   - `WorktreeError` uses `Path` field, not `TicketID`

3. **Pattern Detection**: Git error messages are consistent enough for reliable pattern matching using simple `strings.Contains()` checks, avoiding the need for complex regex.

### Implementation Approach
- Added two new helper functions: `enhanceWorktreeGitError()` and `enhanceWorktreeError()`
- Modified `ConvertError()` to check for worktree-related operations and apply enhancements
- Used the existing `CLIError` structure with its `Suggestions` field for actionable guidance
- Preserved original error details in the `Details` field for debugging

### Testing Considerations
- Used table-driven tests with comprehensive coverage of all error patterns
- Ensured both positive cases (enhanced errors) and negative cases (unrelated errors) are tested
- Fixed test assertions to use partial string matching for suggestions, as exact text may vary

### Files Actually Modified
- `internal/cli/error_converter.go` - Added enhancement logic and helper functions
- `internal/cli/error_converter_test.go` - Added comprehensive test coverage

The implementation successfully integrates with the existing error handling flow without requiring changes to command implementations, as errors automatically flow through the converter.

## Code Review Improvements

### Golang-Pro Review Results
The implementation was reviewed and enhanced with the following improvements:

1. **Code Refactoring**
   - Eliminated duplicate switch statements by extracting common logic into `enhanceWorktreeErrorString()` 
   - Reduced code duplication by ~40 lines
   - Improved maintainability and extensibility

2. **Additional Error Patterns**
   - Added **permission denied** error handling with guidance for permission issues
   - Added **locked worktree** error handling with instructions for dealing with lock files
   
3. **Final Error Patterns Covered** (7 total):
   - Corrupted worktree reference
   - Directory already exists
   - Branch already checked out  
   - Cannot create work tree directory
   - Invalid git reference
   - Permission denied (added)
   - Locked worktree (added)

All tests pass, code quality checks (vet, fmt, lint) pass. Implementation is ready for merge.