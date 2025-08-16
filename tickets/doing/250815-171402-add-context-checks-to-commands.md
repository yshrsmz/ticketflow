---
priority: 2
description: Add early context cancellation checks to commands for consistency
created_at: "2025-08-15T17:14:02+09:00"
started_at: "2025-08-16T13:59:55+09:00"
closed_at: null
related:
    - parent:250812-152927-migrate-remaining-commands
---

# Add Context Checks to Commands

Add early context cancellation checks to all commands for consistency. Currently only some commands have this check.

## Background

During the command migration, it was discovered that some commands have early context cancellation checks while others don't. The project's own documentation (`docs/context-usage.md`) recommends checking context early as a best practice.

## Commands Missing Context Checks

- `list.go`
- `show.go`
- `restore.go`
- `init.go`
- `version.go`
- `worktree_list.go`
- `worktree_clean.go`

## Tasks

- [ ] Add context check to list command Execute method
- [ ] Add context check to show command Execute method
- [ ] Add context check to restore command Execute method
- [ ] Add context check to init command Execute method
- [ ] Add context check to version command Execute method
- [ ] Add context check to worktree_list command Execute method
- [ ] Add context check to worktree_clean command Execute method
- [ ] Run `make test` to verify no breakage
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation if necessary

## Implementation Pattern

Add this check at the beginning of each Execute method:

```go
// Check if context is already cancelled
select {
case <-ctx.Done():
    return ctx.Err()
default:
}
```

## Reference

- Project guideline: `docs/context-usage.md`
- Example implementation: `internal/cli/commands/status.go`