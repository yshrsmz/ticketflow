---
priority: 1
description: "Fix code quality issues including naming conventions, magic strings, and Go idioms"
created_at: "2025-08-01T00:31:26+09:00"
started_at: null
closed_at: null
---

# Fix Code Quality Issues

Address various code quality issues throughout the codebase including inconsistent naming conventions, magic strings/numbers, and non-idiomatic Go patterns.

## Context

The codebase has several code quality issues that impact maintainability:
- Inconsistent receiver names (sometimes `m`, sometimes `tm` for Manager)
- Magic strings and numbers scattered throughout
- Incorrect capitalization (e.g., `Id` instead of `ID`)
- Unnecessary else blocks after returns
- Inconsistent variable naming
- Missing constants for repeated values

These issues make the code:
- Harder to maintain and modify
- More prone to bugs from typos
- Less idiomatic and harder for Go developers to work with
- More difficult to refactor safely

## Tasks

### Naming Convention Fixes
- [ ] Standardize receiver names across all types (use consistent short names)
- [ ] Fix capitalization issues (`ID` not `Id`, `URL` not `Url`)
- [ ] Ensure variable names follow Go conventions
- [ ] Fix method names to be idiomatic Go

### Magic String/Number Elimination
- [ ] Create constants for ticket statuses ("todo", "doing", "done")
- [ ] Create constants for file extensions and patterns
- [ ] Create constants for default values and limits
- [ ] Replace all magic strings with named constants

### Code Structure Improvements
- [ ] Remove unnecessary else blocks after returns
- [ ] Simplify nested if statements where possible
- [ ] Remove redundant type declarations
- [ ] Fix shadowed variable declarations

### Specific Files to Review
- [ ] `internal/ticket/manager.go` - Receiver names, magic strings
- [ ] `internal/git/git.go` - Magic strings, naming conventions
- [ ] `internal/cli/commands.go` - Else blocks, variable names
- [ ] `internal/ui/app.go` - Magic strings, naming
- [ ] `internal/config/config.go` - Default values as constants

### Quality Assurance
- [ ] Run `make fmt` to ensure consistent formatting
- [ ] Run `make vet` to catch common issues
- [ ] Run `make lint` to check for style violations
- [ ] Run `make test` to ensure no regressions
- [ ] Update documentation if necessary
- [ ] Get developer approval before closing

## Implementation Guidelines

### Receiver Naming
```go
// Bad - inconsistent receiver names
func (m *Manager) StartTicket() {}
func (tm *Manager) CloseTicket() {}

// Good - consistent short name
func (m *Manager) StartTicket() {}
func (m *Manager) CloseTicket() {}
```

### Constants for Magic Values
```go
// Bad - magic strings
if status == "todo" {
    // ...
}

// Good - named constants
const (
    StatusTodo  = "todo"
    StatusDoing = "doing"
    StatusDone  = "done"
)

if status == StatusTodo {
    // ...
}
```

### Eliminating Else After Return
```go
// Bad - unnecessary else
if err != nil {
    return err
} else {
    return nil
}

// Good - no else needed
if err != nil {
    return err
}
return nil
```

### Go Naming Conventions
- Acronyms should be all caps: `ID`, `URL`, `API`, `HTTP`
- Interface names should be descriptive: `TicketManager` not `Manager`
- Package names should be lowercase: `ticketflow` not `TicketFlow`
- Exported functions start with capital: `NewManager` not `newManager`

## Notes

These changes are mostly mechanical but will significantly improve code quality and maintainability. Focus on consistency - once a pattern is chosen, apply it everywhere.

Use `golangci-lint` to catch many of these issues automatically. Consider adding it to the CI pipeline to prevent regression.