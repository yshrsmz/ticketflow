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
- [ ] Identify common worktree error patterns in git stderr output
- [ ] Create helper function to detect and enhance worktree errors
- [ ] Update error handling in worktree operations to use enhanced messages
- [ ] Add unit tests for error message enhancement
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
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

1. Create `internal/git/error_enhancement.go`:
```go
func EnhanceWorktreeError(err error) error {
    if err == nil {
        return nil
    }
    
    stderr := err.Error()
    
    // Map common patterns to helpful messages
    for pattern, message := range worktreeErrorMap {
        if matched := pattern.MatchString(stderr); matched {
            return fmt.Errorf("%s\n\nOriginal error: %v", message, err)
        }
    }
    
    return err // Return original if no enhancement needed
}
```

2. Update worktree operations to use enhancement:
- `internal/cli/start.go` - When creating worktree
- `internal/cli/cleanup.go` - When removing worktree
- Any other places where worktree operations occur

### Files to Modify
- Create `internal/git/error_enhancement.go`
- Create `internal/git/error_enhancement_test.go`
- Update `internal/cli/start.go`
- Update `internal/cli/cleanup.go`

## Acceptance Criteria
- [ ] Common worktree errors show helpful, actionable messages
- [ ] Error messages include specific fix instructions
- [ ] Original error is preserved for debugging
- [ ] No regression in existing error handling
- [ ] Test coverage for all enhanced error patterns

## Notes
This is a lightweight improvement that provides immediate value to users without adding complexity to the codebase. It follows the decision to keep ticketflow focused on ticket management while helping users resolve git issues themselves.