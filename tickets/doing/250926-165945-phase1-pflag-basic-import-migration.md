---
priority: 1
description: 'Phase 1: Basic pflag import migration - mechanical replacement of flag imports'
created_at: "2025-09-26T16:59:45+09:00"
started_at: "2025-09-27T21:07:45+09:00"
closed_at: null
related:
    - parent:250924-143504-migrate-to-pflag-for-flexible-cli-args
---

# Phase 1: Basic pflag Import Migration

## Objective
Replace Go's standard `flag` package with `spf13/pflag` across the codebase. This is a mechanical change that immediately enables interspersed flag support.

## Scope
- Add pflag dependency (verify it's not already present first)
- Update all 39 Go files that import the flag package (20 production + 19 test files)
- Fix one instance of ExitOnError behavior
- No functional changes to flag registration logic (that's Phase 2)
- Test files are updated for consistency even though they don't directly benefit from interspersed flags

## Implementation Steps

### 1. Add Dependency
First verify pflag is not already in go.mod, then:
```bash
go get github.com/spf13/pflag
```

### 2. Update Imports
Replace in all 39 files (both production and test files):
```go
// Before
import "flag"

// After
import flag "github.com/spf13/pflag"
```

### Files to Update

#### Production code (20 files):
- `cmd/ticketflow/executor.go`
- `internal/cli/logging.go`
- `internal/cli/commands/version.go`
- `internal/cli/commands/list.go`
- `internal/cli/commands/close.go`
- `internal/cli/commands/show.go`
- `internal/cli/commands/start.go`
- `internal/cli/commands/cleanup.go`
- `internal/cli/commands/new.go`
- `internal/cli/commands/worktree.go`
- `internal/cli/commands/worktree_list.go`
- `internal/cli/commands/status.go`
- `internal/cli/commands/restore.go`
- `internal/cli/commands/help.go`
- `internal/cli/commands/worktree_clean.go`
- `internal/cli/commands/init.go`
- `internal/cli/commands/flag_types.go`
- `internal/cli/commands/workflow.go`
- `internal/command/interface.go`
- `internal/command/migration_example.go`

#### Test files (19 files):
- `internal/cli/commands/start_test.go`
- `internal/cli/commands/new_test.go`
- `internal/cli/commands/worktree_list_test.go`
- `internal/cli/commands/init_test.go`
- `internal/cli/commands/status_test.go`
- `internal/cli/commands/flag_types_test.go`
- `internal/cli/commands/cleanup_integration_test.go`
- `internal/cli/commands/restore_test.go`
- `internal/cli/commands/close_test.go`
- `internal/cli/commands/version_test.go`
- `internal/cli/commands/worktree_clean_test.go`
- `internal/cli/commands/show_test.go`
- `internal/cli/commands/help_test.go`
- `internal/cli/commands/worktree_test.go`
- `internal/cli/commands/cleanup_test.go`
- `internal/cli/commands/list_test.go`
- `internal/command/interface_test.go`
- `internal/command/registry_test.go`

Note: Test files are updated for consistency and to ensure the test suite continues to work correctly with pflag.

### 3. Fix ExitOnError Instance
In `internal/cli/commands/worktree.go:86`, change:
```go
// Before
fs := flag.NewFlagSet(fmt.Sprintf("worktree %s", subcmdName), flag.ExitOnError)

// After
fs := flag.NewFlagSet(fmt.Sprintf("worktree %s", subcmdName), flag.ContinueOnError)
```

## Testing
After making these changes:
1. Run `make test` to ensure all tests pass (no test changes should be needed)
2. Build the binary: `make build`
3. Manually verify interspersed flags work:
   ```bash
   ./ticketflow show ticket-123 --format json  # Should now work!
   ./ticketflow show --format json ticket-123  # Should still work
   ./ticketflow start ticket-123 -f            # Short flags after args
   ./ticketflow cleanup ticket-123 --force     # Long flags after args
   ```

## Success Criteria
- All existing tests pass without modification
- Flags can be placed after positional arguments
- No breaking changes to existing command usage
- No SetInterspersed calls needed (pflag defaults to true)

## Notes
- This is a pure import change - no logic modifications
- The RegisterString/RegisterBool helpers will still work but will be optimized in Phase 2
- Test files are also updated to maintain consistency
- pflag defaults to interspersed mode (SetInterspersed(true)), so no explicit configuration needed
- The only behavioral change is fixing ExitOnError to ContinueOnError for proper error handling

## Risks and Mitigation
- **Risk**: Minimal - pflag is API-compatible with standard flag package
- **Rollback**: Simple - revert imports and remove dependency
- **Testing**: Comprehensive test suite ensures no regressions