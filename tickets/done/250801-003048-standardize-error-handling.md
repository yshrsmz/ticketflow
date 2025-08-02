---
priority: 1
description: Implement consistent error handling patterns with custom error types and proper error wrapping
created_at: "2025-08-01T00:30:48+09:00"
started_at: "2025-08-01T18:32:27+09:00"
closed_at: "2025-08-02T10:46:37+09:00"
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
- [x] Create `internal/errors/errors.go` with domain-specific error types
  - [x] `TicketError` with operation context
  - [x] `WorktreeError` with path information
  - [x] `GitError` with branch details
  - [x] `ConfigError` with field validation info
  - [x] Sentinel errors for common conditions
- [x] Implement error type checking helpers (`IsNotFound()`, `IsAlreadyExists()`)

### Error Wrapping Implementation
- [x] Update all error returns to use `fmt.Errorf` with `%w` verb
- [x] Add contextual information to errors (e.g., ticket ID, operation name)
- [x] Ensure error chains maintain proper context throughout call stack

### Input Validation
- [x] Configuration validation in `internal/config/`
- [x] Ticket validation remains as-is (already comprehensive)
- [x] CLI error conversion with user-friendly messages

### Error Message Standardization
- [x] Replace string-based error checking with `errors.Is()`
- [x] Update all error messages to follow consistent format
- [x] Include actionable information in CLI errors via `error_converter.go`
- [x] Separate internal errors from user-facing messages

### Quality Assurance
- [x] Updated tests to use error type assertions
- [x] Fixed all failing tests
- [x] Run `make test` to ensure no regressions - ALL PASS
- [x] Run `make vet`, `make fmt` and `make lint` - ALL PASS
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

## Implementation Summary

Successfully implemented standardized error handling across the entire codebase:

1. **Created `internal/errors` package** with:
   - Sentinel errors for common conditions (ErrTicketNotFound, ErrNotGitRepo, etc.)
   - Structured error types: TicketError, GitError, WorktreeError, ConfigError
   - Helper functions: IsNotFound(), IsAlreadyExists()

2. **Fixed string-based error checking**:
   - Replaced `strings.Contains(err.Error(), "not found")` with `errors.Is(err, ErrTicketNotFound)`
   - Updated all packages to use proper error types

3. **Added CLI error converter** (`internal/cli/error_converter.go`):
   - Converts internal errors to user-friendly CLI errors
   - Provides helpful suggestions for common error scenarios
   - Maintains existing CLI error structure for JSON/text output

4. **Updated all tests** to use proper error assertions instead of string matching

All tests pass and the codebase now has consistent, maintainable error handling that follows Go best practices.

## Key Insights from Implementation

### 1. **Error Wrapping vs Custom Types Trade-off**
Initially considered using only error wrapping with `fmt.Errorf`, but custom error types proved more valuable for:
- Type-safe error checking with `errors.As()`
- Structured data (e.g., TicketID, Branch, Path) accessible to callers
- Clear domain boundaries between different error categories

### 2. **Sentinel Errors are Still Valuable**
Despite having custom types, sentinel errors (like `ErrTicketNotFound`) remain the best choice for:
- Well-known conditions that multiple packages need to check
- Simple boolean checks with `errors.Is()`
- Maintaining backward compatibility with existing error handling patterns

### 3. **Error Context Chains**
Added context chain support to TicketError for complex operations:
```go
NewTicketErrorWithContext("create", ticketID, err, "worktree", "init")
// Output: "worktree > init > create ticket 123: underlying error"
```
This helps trace errors through multiple layers without losing information.

### 4. **Validation in Constructors**
Added validation to error constructors to catch programming errors early:
- Empty operation names return an error immediately
- Nil underlying errors are rejected
- This prevents malformed errors from propagating through the system

### 5. **CLI Error Conversion Pattern**
The error converter bridges internal and external error representations:
- Internal errors focus on technical accuracy
- CLI errors focus on user experience with actionable suggestions
- This separation allows changing user messages without touching business logic

### 6. **Test Update Strategy**
Updating tests revealed how tightly coupled they were to error strings:
- Changed from exact string matching to semantic error checking
- Tests now verify error behavior, not implementation details
- This makes tests more resilient to error message changes

### 7. **Error Formatting Consistency**
Established clear patterns for error messages:
- Operations: lowercase, verb form ("create", "remove", "push")
- Context included when available (ticket ID, branch name, file path)
- Underlying errors preserved with `%w` for full stack traces

### 8. **Future Considerations**
- Consider adding error codes for programmatic handling
- Structured logging integration would benefit from error types
- Error metrics/monitoring could use error type information
- Consider i18n support for user-facing error messages