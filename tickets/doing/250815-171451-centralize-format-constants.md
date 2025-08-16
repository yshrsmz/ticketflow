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

Format constants are currently duplicated in multiple command files:
- `internal/cli/commands/cleanup.go`
- `internal/cli/commands/close.go`
- `internal/cli/commands/list.go`
- `internal/cli/commands/new.go`
- `internal/cli/commands/restore.go`
- `internal/cli/commands/show.go`
- `internal/cli/commands/start.go`
- `internal/cli/commands/worktree_list.go`

## Tasks

- [ ] Create a central constants file (e.g., `internal/cli/commands/constants.go`)
- [ ] Move FormatText and FormatJSON constants to the central file
- [ ] Update all command files to use the centralized constants
- [ ] Ensure no duplicate definitions remain
- [ ] Run `make test` to verify no breakage
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update any documentation that references these constants

## Implementation Approach

1. Create `internal/cli/commands/constants.go` with:
```go
package commands

const (
    FormatText = "text"
    FormatJSON = "json"
)
```

2. Remove local definitions from all command files
3. Update imports if necessary

## Benefits

- Single source of truth for format constants
- Easier to maintain and modify
- Reduces code duplication
- Prevents inconsistencies