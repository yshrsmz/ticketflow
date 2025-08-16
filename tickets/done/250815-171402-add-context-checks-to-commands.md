---
priority: 2
description: Add early context cancellation checks to commands for consistency
created_at: "2025-08-15T17:14:02+09:00"
started_at: "2025-08-16T13:59:55+09:00"
closed_at: "2025-08-16T16:16:56+09:00"
related:
    - parent:250812-152927-migrate-remaining-commands
---

# Add Context Checks to Commands

Add early context cancellation checks to all commands for consistency. Currently only some commands have this check.

## Background

During the command migration, it was discovered that some commands have early context cancellation checks while others don't. The project's own documentation (`docs/context-usage.md`) recommends checking context early as a best practice.

## Commands Missing Context Checks

After comprehensive analysis, the following commands lack early context cancellation checks:

- `list.go` - Starts directly with `cli.NewApp(ctx)`
- `show.go` - Starts directly with `cli.NewApp(ctx)`
- `restore.go` - Starts directly with `cli.NewApp(ctx)`
- `init.go` - Directly delegates to `cli.InitCommand(ctx)`
- `version.go` - Prints version info without context check
- `worktree_list.go` - Starts with format handling
- `worktree_clean.go` - Starts directly with `cli.NewApp(ctx)`
- `help.go` - Prints help without context check (discovered during analysis)
- `worktree.go` - Parent command lacking context check (discovered during analysis)

## Commands Already Having Context Checks

For reference, these commands already implement the pattern correctly:
- `cleanup.go` ✓
- `close.go` ✓
- `new.go` ✓
- `start.go` ✓
- `status.go` ✓

## Tasks

- [x] Add context check to list command Execute method
- [x] Add context check to show command Execute method
- [x] Add context check to restore command Execute method
- [x] Add context check to init command Execute method
- [x] Add context check to version command Execute method
- [x] Add context check to worktree_list command Execute method
- [x] Add context check to worktree_clean command Execute method
- [x] Add context check to help command Execute method
- [x] Add context check to worktree command Execute method
- [x] Run `make test` to verify no breakage
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Consider updating docs/context-usage.md to mention both patterns (ctx.Err() vs select)

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

## Implementation Insights

### Discoveries
1. **Additional Commands Found**: During analysis, discovered 2 more commands (`help.go` and `worktree.go`) that lacked context checks beyond the initially identified 7 commands.

2. **Test Impact**: Integration tests needed updates because the behavior changed from returning git-specific errors to returning immediate context cancellation errors. This is the expected and correct behavior.

3. **Pattern Consistency**: The codebase uses the select statement pattern rather than direct `ctx.Err()` checks. Both are functionally equivalent, but consistency is important for maintainability.

### Technical Details
- **Performance Impact**: Minimal - context checks are lightweight operations that prevent unnecessary resource allocation
- **Error Behavior Change**: Commands now fail fast with "context canceled" instead of proceeding to git operations that would fail with "Not in a git repository"
- **Test Coverage**: All 4 affected integration tests were updated to match the new behavior

### Documentation Enhancement
Updated `docs/context-usage.md` to document both context checking patterns:
- **Pattern A**: Direct `ctx.Err()` check
- **Pattern B**: Select statement (used consistently in CLI commands)

### Code Review Results
- Golang-pro agent review found the implementation to be excellent with no issues
- All changes follow Go best practices and idioms
- Consistent pattern applied across all 9 commands
- Tests properly updated to reflect behavioral changes

## Completion Status
✅ **All tasks completed successfully**
- 9 commands updated with context checks
- All tests passing
- Code quality checks passed (vet, fmt, lint)
- Documentation updated
- Code reviewed and approved by golang-pro agent