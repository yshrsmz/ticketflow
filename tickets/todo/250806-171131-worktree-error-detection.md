---
priority: 1
description: "Enhance git error detection to identify worktree-specific corruption patterns"
created_at: "2025-08-06T17:11:31+09:00"
started_at: null
closed_at: null
related:
    - parent:250803-121506-worktree-recovery
---

# Worktree Error Detection Infrastructure

## Overview
Enhance the error detection capabilities in the git package to specifically identify worktree corruption and failure patterns. This will enable automatic recovery mechanisms to trigger appropriately.

## Tasks
- [ ] Enhance `Git.Exec()` to detect worktree error patterns in stderr
- [ ] Create `WorktreeCorruptionError` type in `internal/errors/`
- [ ] Implement error pattern matching for worktree-specific issues
- [ ] Add comprehensive unit tests for error detection
- [ ] Document error patterns and recovery triggers
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Get developer approval before closing

## Technical Details
### Error Patterns to Detect
- "fatal: '<path>' is not a working tree"
- "fatal: '<path>' is already a worktree"
- "error: Worktree '<path>' is corrupt"
- "fatal: could not create work tree dir"
- "fatal: target '<branch>' already exists"

### Implementation Approach
1. Create `internal/git/error_patterns.go` with pattern matching logic
2. Enhance `Git.Exec()` method to parse stderr systematically
3. Return structured `WorktreeCorruptionError` with error type and context
4. Maintain backward compatibility with existing error handling

### Files to Modify
- `/internal/git/git.go` - Enhance `Exec()` method
- `/internal/errors/errors.go` - Add `WorktreeCorruptionError`
- `/internal/git/worktree.go` - Add error detection helpers

### New Files
- `/internal/git/error_patterns.go` - Error pattern matching
- `/internal/git/error_patterns_test.go` - Pattern matching tests

## Acceptance Criteria
- [ ] All specified worktree error patterns are correctly detected
- [ ] Error detection doesn't break existing error handling
- [ ] Structured errors provide context for recovery decisions
- [ ] 100% test coverage for error pattern matching
- [ ] Performance impact is negligible (< 1ms per git command)

## Notes
This is phase 1 of the worktree recovery implementation. It provides the foundation for automatic recovery mechanisms without actually performing any recovery actions.