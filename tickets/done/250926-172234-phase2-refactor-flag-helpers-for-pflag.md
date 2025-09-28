---
priority: 2
description: 'Phase 2: Refactor custom flag helpers to use pflag''s native shorthand support'
created_at: "2025-09-26T17:22:34+09:00"
started_at: "2025-09-28T00:05:58+09:00"
closed_at: "2025-09-28T20:07:01+09:00"
related:
    - parent:250924-143504-migrate-to-pflag-for-flexible-cli-args
    - blocked-by:250926-165945-phase1-pflag-basic-import-migration
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

### Issues Discovered in Phase 1
During Phase 1 migration, we discovered that pflag handles single-character flags differently:
- pflag requires `StringVarP`/`BoolVarP` for combined long/short registration
- Using `StringVar` with a single character doesn't create a shorthand flag
- The reflection-based compatibility layer in Phase 1 is a temporary workaround
- Tests had to be updated from `-s` to `--s` for single-char flags registered with StringVar

## Implementation Plan

### 1. Refactor flag_types.go

**CRITICAL: Remove Phase 1 Reflection Workaround**
The Phase 1 implementation added reflection to call pflag's StringVarP/BoolVarP methods.
This MUST be removed and replaced with proper pflag usage.

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

## Implementation Status (Completed)

### ✅ Completed Tasks
1. **Implemented Option A** - Removed all helper functions entirely
2. **Deleted flag_types.go** (154 lines) and **flag_types_test.go** (230 lines)
3. **Updated all 6 commands** to use pflag directly:
   - close.go ✅
   - start.go ✅
   - cleanup.go ✅
   - new.go ✅
   - restore.go ✅
   - worktree_clean.go ✅
4. **Simplified flag structs** from StringFlag/BoolFlag to string/bool
5. **Removed all .Value() calls** - now using direct field access
6. **Added comprehensive flag parsing tests** in TestNewCommand_FlagParsing
7. **Updated all test files** to work with simple types
8. **Fixed all issues from code review**:
   - Added proper CLI flag parsing tests
   - Removed outdated comments about flag utilities
   - Updated references to Value() methods

### 📊 Final Results
- **332 lines removed** (212 insertions, 544 deletions)
- **All tests passing** ✅
- **Code formatted and vetted** ✅
- **PR created**: https://github.com/yshrsmz/ticketflow/pull/101

### 🔍 Key Insights Gained
1. **pflag's native behavior is "last flag wins"** - When both long and short forms are provided, the last one takes precedence. This differs from our original custom logic but is actually more standard.

2. **Reflection workaround was more complex than expected** - The Phase 1 reflection hack added significant complexity (using reflect.ValueOf to call StringVarP/BoolVarP dynamically). Direct usage is much cleaner.

3. **Test coverage improved** - By adding TestNewCommand_FlagParsing, we now properly test actual CLI flag parsing behavior rather than just setting struct fields directly.

4. **Code review via Codex was valuable** - The Codex review caught important issues:
   - Test at line 90 wasn't actually testing flag parsing
   - Several outdated comments remained
   - Need for proper flag parsing tests

5. **Simple types are better** - Using plain string/bool instead of custom StringFlag/BoolFlag types makes the code more approachable and reduces cognitive load.

### 🎯 Next Steps
- Await PR review and approval
- Once merged, Phase 2 will be complete
- Parent ticket (pflag migration) can proceed to any remaining phases or be closed if complete