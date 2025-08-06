---
priority: 1
description: "Implement automatic recovery with git worktree prune and retry mechanisms"
created_at: "2025-08-06T17:12:35+09:00"
started_at: null
closed_at: null
related:
    - parent:250803-121506-worktree-recovery
    - depends_on:250806-171131-worktree-error-detection
---

# Automatic Worktree Recovery Mechanism

## Overview
Implement automatic recovery mechanisms that detect worktree errors and attempt to fix them transparently using `git worktree prune` and retry logic. This builds on the error detection infrastructure to provide seamless recovery from common worktree corruption issues.

## Tasks
- [ ] Create `internal/recovery` package for recovery mechanisms
- [ ] Implement `WorktreeRecovery` struct with recovery methods
- [ ] Add automatic `git worktree prune` on specific errors
- [ ] Implement retry logic with exponential backoff
- [ ] Integrate recovery into existing worktree operations
- [ ] Add comprehensive integration tests
- [ ] Add logging for recovery operations
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Get developer approval before closing

## Technical Details
### Recovery Infrastructure
Create `internal/recovery/worktree.go` with:
```go
type WorktreeRecovery struct {
    git       GitClient
    logger    *log.Logger
    maxRetries int
}

// Core methods:
- DetectCorruption(ctx, error) bool
- AttemptRecovery(ctx) error
- RecoverWithRetry(ctx, operation) error
```

### Integration Points
- `internal/cli/start.go` - Wrap worktree creation
- `internal/cli/cleanup.go` - Wrap worktree removal
- `internal/git/worktree.go` - Add recovery wrapper methods

### Retry Strategy
- Maximum 3 retries with exponential backoff
- Initial delay: 100ms
- Max delay: 2 seconds
- Log each recovery attempt

### Recovery Flow
1. Detect recoverable worktree error
2. Log recovery attempt
3. Run `git worktree prune`
4. Retry original operation
5. If still failing, return original error with recovery context

## Acceptance Criteria
- [ ] Automatic recovery triggers on known worktree errors
- [ ] Recovery attempts are logged for debugging
- [ ] Retry mechanism uses proper backoff strategy
- [ ] Failed recovery returns meaningful error messages
- [ ] No data loss during recovery operations
- [ ] Integration tests cover common corruption scenarios
- [ ] Recovery doesn't mask non-worktree errors

## Notes
This is phase 2 of the worktree recovery implementation. It depends on the error detection infrastructure from phase 1. The recovery should be transparent to users in most cases, with clear logging when recovery is attempted.