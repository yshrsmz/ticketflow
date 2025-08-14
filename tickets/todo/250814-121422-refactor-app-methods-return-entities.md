---
priority: 2
description: "Refactor App methods to return primary entities"
created_at: "2025-08-14T12:14:22+09:00"
started_at: null
closed_at: null
related:
    - parent:250814-013846-migrate-close-command
---

# Refactor App methods to return primary entities

Refactor App methods to return the primary entity they operate on, eliminating the need for commands to re-fetch data for JSON output.

## Background

During the close command migration, we identified that App methods only return errors, forcing commands to re-fetch ticket data for JSON output. After consulting with golang-pro and golang-cli-architect agents, we determined that returning entities is appropriate and legitimate for a daily-use developer tool.

### Expert Consensus
- **golang-pro**: "This isn't about clean architecture for its own sake. It's about making your daily tool more reliable, testable, and pleasant to work with."
- **golang-cli-architect**: "Your middle-ground design is not over-engineering - it's appropriate engineering for a daily-use developer tool."

## Tasks

### 1. Update App Method Signatures
- [ ] Update CloseTicket to return `(*ticket.Ticket, error)`
- [ ] Update CloseTicketWithReason to return `(*ticket.Ticket, error)`
- [ ] Update CloseTicketByID to return `(*ticket.Ticket, error)`
- [ ] Update StartTicket to return `(*ticket.Ticket, error)`
- [ ] Update NewTicket to return `(*ticket.Ticket, error)`
- [ ] Update RestoreTicket to return `(*ticket.Ticket, error)` (when implemented)

### 2. Update Migrated Commands
- [ ] Update close command to use returned ticket
  - [ ] Remove re-fetching logic in outputCloseSuccessJSON
  - [ ] Use returned ticket for JSON output
- [ ] Update start command to use returned ticket
  - [ ] Remove internal JSON handling from App method
  - [ ] Move JSON formatting to command layer
- [ ] Update new command to use returned ticket
  - [ ] Simplify JSON output logic

### 3. Add Helper Methods for Derived Data
- [ ] Create internal/cli/helpers.go file
- [ ] Add CalculateDuration(ticket *ticket.Ticket) time.Duration
- [ ] Add ExtractParentID(ticket *ticket.Ticket) string
- [ ] Add GetWorktreePath(ticketID string) (string, error)

### 4. Update Tests
- [ ] Update App method tests to verify returned tickets
- [ ] Update command tests to mock returned tickets
- [ ] Ensure backward compatibility tests pass
- [ ] Add tests for new helper methods

### 5. Documentation
- [ ] Update COMMAND_MIGRATION_GUIDE.md with new pattern
- [ ] Add examples of using returned entities
- [ ] Document helper method usage

## Benefits

1. **Eliminates re-fetching** - No more duplicate ticket reads after operations
2. **Better testability** - Can assert on returned values directly
3. **Cleaner command code** - Commands focus on presentation, not data retrieval
4. **Performance** - Reduces file I/O operations
5. **Consistency** - All operations follow the same pattern
6. **Idiomatic Go** - Follows standard `(T, error)` return pattern

## Implementation Notes

### What We're Doing
- Return the primary entity from operations that modify it
- Keep backward compatibility (callers can ignore returned ticket)
- Add focused helper methods for derived data

### What We're NOT Doing
- NOT creating complex result structs
- NOT adding unnecessary abstractions
- NOT changing operations that don't naturally return entities (cleanup, etc.)

### Example Implementation
```go
// Before
func (app *App) CloseTicket(ctx context.Context, force bool) error {
    // ... close logic ...
    return nil
}

// After
func (app *App) CloseTicket(ctx context.Context, force bool) (*ticket.Ticket, error) {
    // ... close logic ...
    return closedTicket, nil
}

// Command usage
ticket, err := app.CloseTicket(ctx, force)
if err != nil {
    return err
}
// Use ticket directly for JSON output
```

## Success Criteria

- [ ] All App methods return appropriate entities
- [ ] No more re-fetching in commands
- [ ] All tests pass
- [ ] Commands are cleaner and more focused
- [ ] Performance improvement measurable (fewer file reads)

## References

- Close command implementation that identified this need
- Architectural discussion with golang-pro and golang-cli-architect agents
- Patterns from successful CLI tools (git, docker, kubectl)

## Estimated Time

- **App method updates**: 2 hours
- **Command updates**: 2 hours
- **Helper methods**: 1 hour
- **Testing**: 2 hours
- **Total**: ~1 day of focused work

## Priority

**HIGH** - Should be done BEFORE the restore command migration. This ensures:
1. Restore command uses the clean pattern from the start
2. No need to refactor restore later
3. All future commands see the improved pattern
4. Less total work (implement once correctly vs. implement then refactor)