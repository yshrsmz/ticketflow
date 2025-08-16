---
priority: 3
description: Centralize format constants to avoid duplication across commands
created_at: "2025-08-15T17:14:51+09:00"
started_at: "2025-08-16T23:39:32+09:00"
closed_at: null
related:
    - parent:250812-152927-migrate-remaining-commands
---

# Centralize Format Constants

Move format constants (FormatText, FormatJSON) to a central location to avoid duplication across command files.

## Current State

Format constants are currently duplicated or inconsistently used across command files:

**Files with duplicate constant definitions:**
- `internal/cli/commands/new.go` - defines FormatText and FormatJSON (exported)
- `internal/cli/commands/worktree_list.go` - defines formatText and formatJSON (unexported)

**Files using constants from new.go:**
- `internal/cli/commands/cleanup.go`
- `internal/cli/commands/close.go`
- `internal/cli/commands/start.go`
- `internal/cli/commands/restore.go`
- `internal/cli/commands/worktree_clean.go`

**Files using hardcoded strings instead of constants:**
- `internal/cli/commands/list.go` - uses hardcoded "text" and "json"
- `internal/cli/commands/show.go` - uses hardcoded "text" and "json"

## Tasks

- [x] Create `internal/cli/commands/constants.go` with FormatText and FormatJSON string constants
- [x] Remove duplicate FormatText/FormatJSON definitions from `new.go`
- [x] Remove duplicate formatText/formatJSON definitions from `worktree_list.go`
- [x] Update files currently using constants from new.go:
  - [x] `cleanup.go`
  - [x] `close.go`
  - [x] `start.go`
  - [x] `restore.go`
  - [x] `worktree_clean.go`
- [x] Update files using hardcoded strings to use constants:
  - [x] `list.go` (replace "text" and "json" strings)
  - [x] `show.go` (replace "text" and "json" strings)
- [x] Update `new.go` and `worktree_list.go` to use centralized constants
- [x] Update test files to use constants instead of hardcoded strings where applicable
- [x] Ensure no duplicate definitions remain
- [x] Run `make test` to verify no breakage
- [x] Run `make fmt`, `make vet` and `make lint`

## Implementation Approach

1. Create `internal/cli/commands/constants.go` with:
```go
package commands

// Format string constants for command flags
// These match the values expected by cli.ParseOutputFormat()
const (
    FormatText = "text"  // Maps to cli.FormatText
    FormatJSON = "json"  // Maps to cli.FormatJSON
)
```

2. Remove duplicate definitions from:
   - `new.go` (lines with FormatText and FormatJSON constants)
   - `worktree_list.go` (lines with formatText and formatJSON constants)

3. Update all command files to use the centralized constants instead of:
   - Their own constant definitions
   - Hardcoded "text" and "json" strings
   - References to constants from other command files

## Benefits

- Single source of truth for format constants
- Easier to maintain and modify
- Reduces code duplication
- Prevents inconsistencies
- Consistent usage pattern across all commands

## Notes

- The `cli` package already has `cli.FormatText` and `cli.FormatJSON` as `OutputFormat` type constants
- Command-level string constants are needed for flag parsing before conversion via `cli.ParseOutputFormat()`
- This refactoring ensures consistent string usage at the command level