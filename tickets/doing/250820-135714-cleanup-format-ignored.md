---
priority: 2
description: Fix cleanup command ignoring --format parameter
created_at: "2025-08-20T13:57:14+09:00"
started_at: "2025-08-20T14:00:33+09:00"
closed_at: null
---

# Fix cleanup command ignoring --format parameter

## Problem
The `ticketflow cleanup` command doesn't respect the `--format` parameter. When running `ticketflow cleanup --format json`, it still outputs human-readable format instead of JSON.

## Root Cause
The bug was in the `normalize()` function that merges long and short form flags. The function incorrectly overwrote the `--format` value with the default value from the `-o` short form flag:

```go
// BUGGY CODE:
if f.formatShort != "" {  // Always true since default is "text"
    f.format = f.formatShort  // Overwrites user's --format with default
}
```

Since `formatShort` always has the default value "text" (even when not explicitly set), it would always overwrite the user's `--format json` setting.

## Solution Implemented
Fixed the normalize logic to only use the short form if it was explicitly set to a non-default value:

```go
// FIXED CODE:
if f.formatShort != "" && f.formatShort != FormatText {
    f.format = f.formatShort
}
```

## Commands Affected and Fixed
- `cleanup` - Fixed ✅
- `close` - Fixed ✅
- `new` - Fixed ✅
- `restore` - Already had correct implementation
- `start` - Fixed ✅
- `worktree_clean` - Fixed ✅
- `worktree_list` - Already had correct implementation

Commands without the issue (no `-o` short form):
- `list`, `show`, `status` - Not affected

## Tasks
- [x] Investigate cleanup command implementation in `internal/cli/commands/cleanup.go`
- [x] Identify the normalize function bug causing --format to be ignored
- [x] Fix normalize logic in cleanup and all affected commands
- [x] Verify JSON output works correctly with --format json
- [x] Test dry-run mode with JSON format
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Update the ticket with root cause and solution
- [ ] Get developer approval before closing

## Testing
Verified that all the following now work correctly:
- `ticketflow cleanup --format json` - outputs JSON ✅
- `ticketflow cleanup --format json --dry-run` - outputs JSON ✅
- `ticketflow cleanup -o json` - outputs JSON (unchanged) ✅
- Other affected commands with `--format json` - all fixed ✅

## Notes
This was a systemic issue affecting 6 commands that had both `--format` and `-o` flags with the buggy normalize function. The fix ensures that the short form flag only takes precedence when explicitly set by the user, not when it just has the default value.