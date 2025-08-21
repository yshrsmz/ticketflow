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

## Tasks

- [ ] Create `internal/cli/commands/helpers.go` file
- [ ] Extract `extractParentFromTicket` helper function
- [ ] Extract `validateFormat` helper function  
- [ ] Extract `normalizeFlags` helper for flag merging
- [ ] Create generic `assertFlags[T]` helper for safe type assertions
- [ ] Update commands to use shared helpers
- [ ] Remove duplicate implementations
- [ ] Run `make test` to verify no breakage
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Add tests for new helper functions

## Example Implementations

### Safe Type Assertion Helper
```go
func assertFlags[T any](flags interface{}) (*T, error) {
    f, ok := flags.(*T)
    if !ok {
        return nil, fmt.Errorf("invalid flags type: expected *%T, got %T", *new(T), flags)
    }
    return f, nil
}
```

### Format Validation Helper
```go
func validateFormat(format string) error {
    if format != FormatText && format != FormatJSON {
        return fmt.Errorf("invalid format: %q (must be 'text' or 'json')", format)
    }
    return nil
}
```

### Parent Extraction Helper
```go
func extractParentFromTicket(ticket *ticket.Ticket) string {
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