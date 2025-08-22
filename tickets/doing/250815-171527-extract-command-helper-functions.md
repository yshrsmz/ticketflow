---
priority: 3
description: Extract common helper functions used across commands
created_at: "2025-08-15T17:15:27+09:00"
started_at: "2025-08-21T17:46:24+09:00"
closed_at: null
related:
    - parent:250812-152927-migrate-remaining-commands
---

# Extract Command Helper Functions

Extract common patterns and helper functions that are duplicated across command implementations.

## Common Patterns to Extract

### 1. Parent Ticket Extraction
Currently duplicated logic for extracting parent from ticket relationships.

### 2. Format Validation
Common pattern for validating format flags (text|json).

### 3. Flag Normalization
Pattern for merging short and long form flags.

### 4. Safe Type Assertion
Common pattern for safe type assertion with error handling.

## Implementation Plan

After analyzing the codebase:
- **9 commands** use type assertions with `ok := flags.(*commandFlags)`
- **6+ commands** duplicate format validation logic
- **3 commands** duplicate parent ticket extraction logic
- Existing `flag_utils.go` already handles flag normalization well

### Refactoring Approach
1. **Rename**: `flag_utils.go` → `flag_types.go` (better describes content)
2. **Create**: `validation.go` for validation and extraction helpers

## Tasks

- [x] Rename `internal/cli/commands/flag_utils.go` to `flag_types.go`
- [x] Create `internal/cli/commands/validation.go` file
- [x] Extract `ExtractParentFromTicket` helper function
- [x] Extract `ValidateFormat` helper function  
- [x] Create generic `AssertFlags[T]` helper for safe type assertions
- [x] Update close.go to use validation helpers
- [x] Update new.go to use validation helpers
- [x] Update restore.go to use validation helpers
- [x] Update list.go to use validation helpers
- [x] Update show.go to use validation helpers
- [x] Update status.go to use validation helpers
- [x] Update start.go to use validation helpers
- [x] Update cleanup.go to use validation helpers
- [x] Update worktree_clean.go to use validation helpers
- [x] Update worktree_list.go to use validation helpers
- [x] Update version.go to use validation helpers
- [x] Create validation_test.go with unit tests
- [x] Update existing tests to match new imports (flag_utils_test.go)
- [x] Run `make test` to verify no breakage
- [x] Run `make vet`, `make fmt` and `make lint`

## Example Implementations

### Safe Type Assertion Helper
```go
func AssertFlags[T any](flags interface{}) (*T, error) {
    f, ok := flags.(*T)
    if !ok {
        return nil, fmt.Errorf("invalid flags type: expected *%T, got %T", *new(T), flags)
    }
    return f, nil
}
```

### Format Validation Helper
```go
func ValidateFormat(format string) error {
    if format != FormatText && format != FormatJSON {
        return fmt.Errorf("invalid format: %q (must be 'text' or 'json')", format)
    }
    return nil
}
```

### Parent Extraction Helper
```go
func ExtractParentFromTicket(ticket *ticket.Ticket) string {
    if ticket == nil || len(ticket.Related) == 0 {
        return ""
    }
    for _, rel := range ticket.Related {
        if strings.HasPrefix(rel, "parent:") {
            return strings.TrimPrefix(rel, "parent:")
        }
    }
    return ""
}
```

## Benefits

- Reduces code duplication
- Single source of truth for common patterns
- Easier to maintain and test
- Improves code consistency
- Makes commands more focused on their specific logic
