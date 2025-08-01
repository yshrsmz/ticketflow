---
priority: 1
description: Break down large functions into smaller, focused functions following Single Responsibility Principle
created_at: "2025-08-01T00:30:10+09:00"
started_at: "2025-08-01T11:10:01+09:00"
closed_at: null
---

# Decompose Large Functions

Break down large functions into smaller, focused functions that follow the Single Responsibility Principle. This will improve code readability, maintainability, and testability.

## Context

The codebase contains several large functions that are doing too many things:
- `StartTicket` in `internal/ticket/manager.go` is 448 lines long
- `CloseTicket` in `internal/ticket/manager.go` is complex with multiple responsibilities
- Several functions in `internal/cli/commands.go` and `internal/ui/app.go` exceed 100 lines

Large functions are harder to:
- Understand and reason about
- Test effectively
- Modify without introducing bugs
- Reuse in different contexts

## Tasks

### Primary Decomposition Targets
- [ ] Decompose `StartTicket` (448 lines) in `internal/ticket/manager.go`
  - [ ] Extract ticket validation logic
  - [ ] Extract worktree creation logic
  - [ ] Extract branch setup logic
  - [ ] Extract status update logic
  - [ ] Extract file movement operations
- [ ] Decompose `CloseTicket` in `internal/ticket/manager.go`
  - [ ] Extract worktree cleanup logic
  - [ ] Extract branch operations
  - [ ] Extract ticket archival logic
- [ ] Extract common ticket movement operations into utilities

### Secondary Targets
- [ ] Review and decompose large functions in `internal/cli/commands.go`
- [ ] Review and decompose large functions in `internal/ui/app.go`
- [ ] Create helper functions for repeated patterns

### Quality Assurance
- [ ] Ensure each function has a single, clear responsibility
- [ ] Add unit tests for newly extracted functions
- [ ] Run `make test` to ensure no regressions
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation if necessary
- [ ] Get developer approval before closing

## Implementation Guidelines

### Function Decomposition Strategy
1. Identify logical sections within large functions
2. Extract each section into a well-named helper function
3. Pass dependencies explicitly (avoid hidden state)
4. Return errors consistently
5. Keep functions under 50 lines when possible

### Naming Conventions
- Use descriptive verb-noun combinations (e.g., `validateTicket`, `createWorktree`)
- Keep names concise but clear
- Use consistent naming patterns across similar operations

### Example Decomposition for StartTicket
```go
// Before: One large function
func (tm *Manager) StartTicket(id string) error {
    // 448 lines of mixed logic
}

// After: Multiple focused functions
func (tm *Manager) StartTicket(id string) error {
    ticket, err := tm.loadAndValidateTicket(id)
    if err != nil {
        return err
    }
    
    if err := tm.createTicketWorktree(ticket); err != nil {
        return err
    }
    
    if err := tm.setupTicketBranch(ticket); err != nil {
        return err
    }
    
    if err := tm.updateTicketStatus(ticket, StatusDoing); err != nil {
        return err
    }
    
    return nil
}
```

## Notes

This ticket depends on the interface extraction ticket (250801-002917) being completed first, as the decomposed functions will be easier to test with proper interfaces in place.

Focus on making each function do one thing well. If a function name contains "and" or requires multiple sentences to describe, it probably needs further decomposition.