---
priority: 2
description: Migrate from standard flag to spf13/pflag to support flags after positional arguments
created_at: "2025-09-24T14:35:04+09:00"
started_at: "2025-09-26T16:41:16+09:00"
closed_at: null
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

Migrate from standard `flag` package to `spf13/pflag` which supports interspersed flags. pflag is mostly a drop-in replacement and **interspersed mode is enabled by default** (contrary to earlier assumptions).

**UPDATE**: pflag defaults to `interspersed = true` when creating a new FlagSet, so no explicit `SetInterspersed(true)` calls are needed!

## Why pflag Over Other Options

After analyzing multiple CLI packages:
- **pflag**: Drop-in replacement, 2-3 hours migration, solves problem immediately
- **cobra**: Would require 2-3 days of refactoring entire command structure
- **urfave/cli v2**: Doesn't support flags after args (same problem we have)
- **kong**: Requires moderate refactoring (1-2 days), struct-based approach

## Implementation Plan

### Phase 1: Basic Migration (Sub-ticket 1)
```bash
go get github.com/spf13/pflag
```

Update all imports from `"flag"` to `flag "github.com/spf13/pflag"` across 20 files.

### Phase 2: Refactor Flag Helpers (Sub-ticket 2)
Update the custom flag helpers in `internal/cli/commands/flag_types.go` and the 6 commands that use them:
- Convert `RegisterString` calls to use pflag's `StringVarP` method
- Convert `RegisterBool` calls to use pflag's `BoolVarP` method
- Simplify the StringFlag/BoolFlag types since pflag handles short/long forms natively

### Phase 3: Comprehensive Testing (Sub-ticket 3)
Add integration tests to verify interspersed flag support:
- Test all commands with flags after positional arguments
- Ensure backward compatibility (flags before args still work)
- Add test cases for edge cases (multiple args, mixed flags)
- Update existing tests that validate "unexpected arguments" errors

### Key Changes Required

1. **Import Updates** (20 files):
   - Simple alias import change
   - No SetInterspersed needed (defaults to true)
   - Fix one instance of ExitOnError in worktree.go:86

2. **Flag Helper Refactoring** (7 files):
   ```go
   // Before: Custom RegisterString/RegisterBool
   RegisterString(fs, &flags.format, "format", "o", FormatText, "...")

   // After: Use pflag's native *VarP methods
   fs.StringVarP(&flags.format, "format", "o", FormatText, "...")
   ```

3. **Testing**:
   - Verify all commands work with flags in any position
   - Update tests that expect "unexpected arguments" errors
   - Add comprehensive integration tests

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
5. Migration completed in 2-3 hours (simpler than originally thought since SetInterspersed not needed)

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

## Implementation Strategy

**Breaking this into 3 sub-tickets for staged delivery:**

1. **Sub-ticket 1**: Basic pflag migration (imports only)
   - Low risk, mechanical change
   - Enables interspersed flags immediately
   - ~1 hour effort

2. **Sub-ticket 2**: Refactor flag helpers
   - Update RegisterString/RegisterBool functions
   - Simplify flag precedence logic
   - ~1 hour effort

3. **Sub-ticket 3**: Comprehensive testing
   - Add integration tests for interspersed flags
   - Update existing test expectations
   - ~1 hour effort

This staged approach reduces risk and makes review easier.

## References

- pflag documentation: https://github.com/spf13/pflag
- Issue discussion: Current conversation about CLI flexibility
- Standard flag limitations: https://github.com/golang/go/issues/36744
- **Corrected**: pflag defaults to interspersed=true, no explicit SetInterspersed needed