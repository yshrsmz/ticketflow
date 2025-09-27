---
priority: 2
description: "Phase 2: Refactor custom flag helpers to use pflag's native shorthand support"
created_at: "2025-09-26T17:22:34+09:00"
started_at: null
closed_at: null
related:
  - "parent:250924-143504-migrate-to-pflag-for-flexible-cli-args"
  - "blocked-by:250926-165945-phase1-pflag-basic-import-migration"
---

# Phase 2: Refactor Flag Helpers for pflag

## Objective
Refactor the custom flag helper functions to leverage pflag's built-in support for shorthand flags, simplifying our codebase and removing unnecessary complexity.

## Prerequisites
- Phase 1 must be completed (basic pflag import migration)

## Scope
- Refactor `RegisterString` and `RegisterBool` functions in `flag_types.go`
- Update 6 commands that use these helpers
- Simplify or eliminate the StringFlag and BoolFlag types

## Current Implementation Issues
The current implementation has custom logic for handling short/long flag precedence:
- `StringFlag` tracks which form was set with a `shortSet` field
- `RegisterString` uses `fs.Func` for custom short flag handling
- This complexity is unnecessary with pflag

## Implementation Plan

### 1. Refactor flag_types.go

**Option A: Direct Migration (Recommended)**
Remove the helper functions entirely and use pflag directly in commands:
```go
// Before (with helper)
RegisterString(fs, &flags.format, "format", "o", FormatText, "Output format")

// After (direct pflag)
fs.StringVarP(&flags.format, "format", "o", FormatText, "Output format")
```

**Option B: Simplified Helpers**
Keep helpers but simplify to thin wrappers:
```go
func RegisterString(fs *flag.FlagSet, p *string, longName, shortName, defaultValue, usage string) {
    if longName != "" && shortName != "" {
        fs.StringVarP(p, longName, shortName, defaultValue, usage)
    } else if longName != "" {
        fs.StringVar(p, longName, defaultValue, usage)
    } else if shortName != "" {
        // pflag requires a long name, use short as long if only short provided
        fs.StringVarP(p, shortName, "", defaultValue, usage)
    }
}
```

### 2. Update Command Structs

Simplify flag structs in commands:
```go
// Before
type startFlags struct {
    force  BoolFlag
    format StringFlag
}

// After
type startFlags struct {
    force  bool
    format string
}
```

### 3. Update Commands

Commands using RegisterString/RegisterBool (6 files):
- `internal/cli/commands/close.go`
- `internal/cli/commands/start.go`
- `internal/cli/commands/cleanup.go`
- `internal/cli/commands/new.go`
- `internal/cli/commands/restore.go`
- `internal/cli/commands/worktree_clean.go`

Example changes:
```go
// Before
RegisterBool(fs, &flags.force, "force", "f", "Force operation")
RegisterString(fs, &flags.format, "format", "o", FormatText, "Output format")
// ...
if f.format.Value() == "json" { ... }

// After
fs.BoolVarP(&flags.force, "force", "f", false, "Force operation")
fs.StringVarP(&flags.format, "format", "o", FormatText, "Output format")
// ...
if f.format == "json" { ... }
```

### 4. Update Tests

Update test files that use the flag helpers:
- `internal/cli/commands/flag_types_test.go` - Remove or simplify tests
- Various `*_test.go` files that test flag behavior

## Benefits
- Removes ~50 lines of custom code
- Eliminates thread-safety concerns with `shortSet` field
- Leverages battle-tested pflag functionality
- Cleaner, more idiomatic Go code

## Testing
1. Run all existing tests: `make test`
2. Verify short/long flag behavior still works:
   ```bash
   ./ticketflow start -f ticket-123        # Short form
   ./ticketflow start --force ticket-123   # Long form
   ./ticketflow start ticket-123 -f        # Interspersed
   ```
3. Ensure precedence still works if both forms provided

## Success Criteria
- All flag helper complexity removed
- Commands use pflag's native *VarP methods
- Tests pass without modification
- Short and long forms work correctly

## Decision Point
Before implementation, decide between:
- **Option A**: Remove helpers entirely (cleaner, more explicit)
- **Option B**: Keep simplified helpers (maintains abstraction)

Recommendation: Option A for transparency and simplicity.