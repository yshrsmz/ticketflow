---
priority: 2
description: "Migrate from standard flag to spf13/pflag to support flags after positional arguments"
created_at: "2025-09-24T14:35:04+09:00"
started_at: null
closed_at: null
related: []
---

# Migrate to pflag for Flexible CLI Arguments

## Problem Statement

Currently, ticketflow uses Go's standard `flag` package which stops parsing at the first non-flag argument. This causes errors when AI coding agents (and users) naturally place flags after positional arguments:

```bash
# Current behavior - ERROR
ticketflow show ticket-123 --format json
# Error: unexpected arguments after ticket ID: [--format json]

# Required workaround - WORKS but unintuitive
ticketflow show --format json ticket-123
```

This affects commands: `show`, `start`, `close`, `cleanup`, and others that take a ticket ID.

## Solution

Migrate from standard `flag` package to `spf13/pflag` which supports interspersed flags. While pflag is mostly a drop-in replacement, we need to explicitly enable interspersed mode since it defaults to false for compatibility.

**CRITICAL**: pflag requires `SetInterspersed(true)` to allow flags after positional arguments!

## Why pflag Over Other Options

After analyzing multiple CLI packages:
- **pflag**: Drop-in replacement, 2-3 hours migration, solves problem immediately
- **cobra**: Would require 2-3 days of refactoring entire command structure
- **urfave/cli v2**: Doesn't support flags after args (same problem we have)
- **kong**: Requires moderate refactoring (1-2 days), struct-based approach

## Implementation Plan

### Phase 1: Add Dependency
```bash
go get github.com/spf13/pflag
```

### Phase 2: Update Imports
Replace across all files:
```go
// Before
import "flag"

// After
import flag "github.com/spf13/pflag"
```

### Phase 3: Enable Interspersed Mode
After creating each FlagSet, enable interspersed parsing:
```go
fs := pflag.NewFlagSet(cmd.Name(), pflag.ContinueOnError)
fs.SetInterspersed(true) // CRITICAL: Allow flags after positional args
```

Locations that need this change:
- `cmd/ticketflow/executor.go:15` - Main executor
- `internal/cli/commands/worktree.go:86` - Worktree subcommands
- Any other `NewFlagSet` calls found during migration

### Phase 4: Refactor Flag Helpers
The custom flag helpers in `internal/cli/commands/flag_types.go` need refactoring:
- Current: Registers short/long flags separately with precedence logic
- pflag: Use `*VarP` methods for built-in shorthand support
- Example: `fs.StringVarP(&flags.format, "format", "f", "text", "Output format")`

### Phase 5: Test All Commands
Verify that flags work in any position:
```bash
# All of these should work after migration
ticketflow show ticket-123 --format json
ticketflow show --format json ticket-123
ticketflow start ticket-123 --format json
ticketflow close ticket-123 --format json
```

### Phase 6: Update Tests
- Update test cases to verify flexible flag positioning
- Add new test cases for interspersed flags
- Ensure backward compatibility (old style still works)

## Files to Modify

### Core Files
- `cmd/ticketflow/executor.go` - Main flag parsing logic
- `internal/cli/logging.go` - Logging flags setup

### Command Files (30+ files)
- `internal/cli/commands/*.go` - All command implementations
- `internal/cli/commands/*_test.go` - All command tests

### Documentation
- `CLAUDE.md` - Update AI integration guidelines
- `README.md` - Update usage examples if needed

## Testing Checklist

- [ ] All existing tests pass
- [ ] Flags work before positional args (backward compatibility)
- [ ] Flags work after positional args (new capability)
- [ ] Flags work interspersed with multiple args
- [ ] Help text still displays correctly
- [ ] Error messages remain clear
- [ ] JSON output format works in all positions
- [ ] **NEW: Integration tests for real CLI parsing path** (not just unit tests)
- [ ] **NEW: Test nested commands (e.g., `worktree clean`) with interspersed flags**
- [ ] **NEW: Verify short/long flag precedence still works correctly**

## Migration Commands

```bash
# Add dependency
go get github.com/spf13/pflag

# Run tests before changes
make test

# SAFER: Use goimports or manual edits per package instead of global sed
# Avoid mangling multi-line imports and doc strings
go get -u golang.org/x/tools/cmd/goimports

# For each package, update imports manually or with tooling
# Then add SetInterspersed(true) after each NewFlagSet call

# Run tests after changes
make test

# Test the fix works with interspersed flags
./ticketflow show $(ls tickets/todo/*.md | head -1 | xargs basename .md) --format json
```

## Success Criteria

1. AI agents can use natural flag ordering without errors
2. All existing functionality preserved
3. No breaking changes for current users
4. Tests provide coverage for new capability
5. ~~Migration completed in under 3 hours~~ **Revised: 4-6 hours** (due to flag helper refactoring and comprehensive testing)

## Benefits

- **Immediate Fix**: Solves the AI agent integration issue
- **Drop-in Replacement**: Minimal code changes required
- **Battle-tested**: Used by Kubernetes, Docker, and other major projects
- **Future-proof**: If we later want Cobra (which uses pflag internally), this is a stepping stone
- **Better UX**: More intuitive command-line usage for all users

## Rollback Plan

If issues arise, reverting is straightforward:
1. Revert the import changes
2. Remove pflag dependency from go.mod
3. Run `go mod tidy`

Since pflag is API-compatible with standard flag, the risk is minimal.

## Known Complexities (from Codex review)

1. **SetInterspersed Required**: pflag does NOT enable interspersed by default - must explicitly call `SetInterspersed(true)`
2. **Flag Helper Refactoring**: Custom short/long flag logic in `flag_types.go` needs complete rewrite for pflag's `*VarP` API
3. **Test Coverage Gap**: No existing integration tests for real CLI parsing path - need to add these
4. **Estimate Adjustment**: 30+ files to modify, plus testing = likely 4-6 hours, not 2-3

## References

- pflag documentation: https://github.com/spf13/pflag
- Issue discussion: Current conversation about CLI flexibility
- Standard flag limitations: https://github.com/golang/go/issues/36744
- Codex feasibility review: Identified critical SetInterspersed requirement