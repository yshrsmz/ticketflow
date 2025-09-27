---
priority: 1
description: "Phase 1: Basic pflag import migration - mechanical replacement of flag imports"
created_at: "2025-09-26T16:59:45+09:00"
started_at: null
closed_at: null
related:
  - "parent:250924-143504-migrate-to-pflag-for-flexible-cli-args"
---

# Phase 1: Basic pflag Import Migration

## Objective
Replace Go's standard `flag` package with `spf13/pflag` across the codebase. This is a mechanical change that immediately enables interspersed flag support.

## Scope
- Add pflag dependency
- Update 20 Go files that import the flag package
- Fix one instance of ExitOnError behavior
- No functional changes to flag registration logic (that's Phase 2)

## Implementation Steps

### 1. Add Dependency
```bash
go get github.com/spf13/pflag
```

### 2. Update Imports
Replace in all 20 files:
```go
// Before
import "flag"

// After
import flag "github.com/spf13/pflag"
```

### Files to Update
Production code (non-test files):
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
1. Run `make test` to ensure all tests pass
2. Manually verify interspersed flags work:
   ```bash
   ./ticketflow show ticket-123 --format json  # Should now work!
   ./ticketflow show --format json ticket-123  # Should still work
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