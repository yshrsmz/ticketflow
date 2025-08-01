---
priority: 1
description: "Implement consistent error handling patterns with custom error types and proper error wrapping"
created_at: "2025-08-01T00:30:48+09:00"
started_at: null
closed_at: null
---

# Standardize Error Handling

Implement consistent error handling patterns throughout the codebase with custom error types, proper error wrapping, and comprehensive input validation.

## Context

Current error handling in the codebase is inconsistent:
- Some functions return generic errors without context
- Error messages are not standardized
- Missing input validation in many public APIs
- No custom error types for domain-specific errors
- Inconsistent use of error wrapping

Proper error handling is crucial for:
- Debugging and troubleshooting
- Providing meaningful feedback to users
- Maintaining system reliability
- Following Go best practices

## Tasks

### Custom Error Types
- [ ] Create `internal/errors/errors.go` with domain-specific error types
  - [ ] `TicketNotFoundError`
  - [ ] `WorktreeExistsError`
  - [ ] `InvalidTicketStateError`
  - [ ] `GitOperationError`
  - [ ] `ConfigurationError`
- [ ] Implement error type checking helpers (e.g., `IsTicketNotFound()`)

### Error Wrapping Implementation
- [ ] Update all error returns to use `fmt.Errorf` with `%w` verb
- [ ] Add contextual information to errors (e.g., ticket ID, operation name)
- [ ] Ensure error chains maintain proper context throughout call stack

### Input Validation
- [ ] Add validation for all public API methods in `internal/ticket/manager.go`
- [ ] Add validation for CLI command inputs in `internal/cli/`
- [ ] Add validation for configuration values in `internal/config/`
- [ ] Create reusable validation functions for common patterns

### Error Message Standardization
- [ ] Define error message templates for common scenarios
- [ ] Update all error messages to follow consistent format
- [ ] Include actionable information in user-facing errors
- [ ] Separate internal errors from user-facing messages

### Quality Assurance
- [ ] Add tests for all custom error types
- [ ] Add tests for error wrapping and unwrapping
- [ ] Ensure error messages are helpful and consistent
- [ ] Run `make test` to ensure no regressions
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation if necessary
- [ ] Get developer approval before closing

## Implementation Guidelines

### Custom Error Type Pattern
```go
// internal/errors/errors.go
type TicketNotFoundError struct {
    ID string
}

func (e TicketNotFoundError) Error() string {
    return fmt.Sprintf("ticket not found: %s", e.ID)
}

func IsTicketNotFound(err error) bool {
    var e TicketNotFoundError
    return errors.As(err, &e)
}
```

### Error Wrapping Pattern
```go
// Before
if err != nil {
    return err
}

// After
if err != nil {
    return fmt.Errorf("failed to load ticket %s: %w", id, err)
}
```

### Input Validation Pattern
```go
func (tm *Manager) StartTicket(id string) error {
    if id == "" {
        return fmt.Errorf("ticket ID cannot be empty")
    }
    
    if !isValidTicketID(id) {
        return fmt.Errorf("invalid ticket ID format: %s", id)
    }
    
    // ... rest of implementation
}
```

### Error Message Standards
- Start with lowercase (Go convention)
- Be specific about what failed
- Include relevant context (IDs, filenames, etc.)
- Suggest fixes when possible
- Keep messages concise

## Notes

Good error handling is fundamental to a maintainable codebase. This ticket will make debugging easier and improve the user experience by providing clear, actionable error messages.

Consider using the `errors` package for wrapping and `errors.Is/As` for type checking. This approach is more flexible than string comparison and maintains the error chain.