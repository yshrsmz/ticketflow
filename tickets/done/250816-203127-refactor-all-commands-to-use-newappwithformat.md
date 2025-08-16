---
priority: 2
description: Refactor all CLI commands to use NewAppWithFormat for cleaner initialization
created_at: "2025-08-16T20:31:27+09:00"
started_at: "2025-08-16T22:44:15+09:00"
closed_at: "2025-08-16T23:27:18+09:00"
related:
    - parent:250816-123703-improve-json-output-separation
---

# Refactor All Commands to Use NewAppWithFormat

## Problem
Currently, many commands still use the old pattern of creating an App then mutating it after flag parsing. This creates awkward state mutations and makes the code harder to understand.

## Current State
All commands have been successfully refactored! ✅

## Tasks
- [x] Update `list.go` to use NewAppWithFormat
- [x] Update `show.go` to use NewAppWithFormat
- [x] Update `new.go` to use NewAppWithFormat
- [x] Update `restore.go` to use NewAppWithFormat
- [x] Update `worktree_clean.go` to use NewAppWithFormat (added JSON format support)
- [x] Update `start.go` and `close.go` to use NewAppWithFormat via new helper
- [x] Create `getAppWithFormat` helper for start/close commands
- [x] Remove deprecated `getApp` function
- [x] Run `make test` to verify all commands work
- [x] Run `make vet`, `make fmt` and `make lint`

## Implementation Pattern
Replace:
```go
app, err := cli.NewApp(ctx)
if err != nil {
    return err
}
// ... parse flags ...
app.Output = cli.NewOutputWriter(os.Stdout, os.Stderr, outputFormat)
app.StatusWriter = cli.NewStatusWriter(os.Stdout, outputFormat)
```

With:
```go
// ... parse flags first ...
outputFormat := cli.ParseOutputFormat(f.format)
app, err := cli.NewAppWithFormat(ctx, outputFormat)
if err != nil {
    return err
}
```

## Benefits
- Cleaner initialization flow
- App is immutable after creation
- No awkward state mutations
- Follows industry best practices (Kong, Docker, kubectl)

## Notes
This is a follow-up to the JSON/text output separation work. The pattern has been proven to work in three commands already.

## Key Insights & Implementation Notes

### Commands Refactored
1. **Simple commands** (`list`, `show`, `new`, `restore`): Straightforward replacement of the pattern
2. **Commands with test helpers** (`start`, `close`): Required creating `getAppWithFormat` helper to maintain test compatibility
3. **Commands needing enhancement** (`worktree_clean`): Added JSON format support as part of the refactoring

### Technical Improvements
- **CleanWorktreesResult struct**: Added structured return type for `CleanWorktrees` method to support JSON output
- **Removed technical debt**: Deleted deprecated `getApp()` function that was no longer needed
- **Test updates**: Fixed integration test to handle new return type from `CleanWorktrees`

### Pattern Benefits Realized
- All commands now follow consistent initialization pattern
- Flag parsing happens before App creation (no more mutations)
- Output format is determined at App creation time
- JSON/text output handling is cleaner and more predictable

### Completion Status
✅ All 7 identified commands refactored
✅ Additional commands (`start`, `close`) discovered and refactored
✅ JSON support added to `worktree_clean`
✅ All tests passing
✅ Linting clean

This completes the refactoring work and establishes a clean, consistent pattern for all CLI commands.