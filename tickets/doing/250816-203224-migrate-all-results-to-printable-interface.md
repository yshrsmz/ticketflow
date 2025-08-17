---
priority: 2
description: Migrate all result types to implement Printable interface to eliminate switch statement
created_at: "2025-08-16T20:32:24+09:00"
started_at: "2025-08-17T23:35:16+09:00"
closed_at: null
related:
    - parent:250816-123703-improve-json-output-separation
---

# Migrate All Results to Printable Interface

## Problem
The current OutputWriter uses a large switch statement to handle different result types. This is tightly coupled and will grow huge over time. We need to migrate to the Printable interface pattern (inspired by kubectl's ResourcePrinter).

## Current State
- `Printable` interface has been created in `internal/cli/printable.go`
- `CleanupResult` already implements Printable ✅
- `TicketListResult` example implementation exists ✅
- OutputWriter checks for Printable first, falls back to switch ✅

## Tasks
- [ ] Make ticket.Ticket implement Printable
- [ ] Create WorktreeListResult wrapper and implement Printable
- [ ] Create StatusResult wrapper and implement Printable
- [ ] Migrate any map[string]interface{} results to typed structs with Printable
- [ ] Update all commands to return Printable types
- [ ] Remove the switch statement from textResultWriter.PrintResult
- [ ] Remove deprecated print methods from textResultWriter
- [ ] Run `make test` to verify all output still works
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation about the Printable pattern

## Implementation Pattern

### For existing types:
```go
// Add to existing type (e.g., ticket.Ticket)
func (t *Ticket) TextRepresentation() string {
    // Format for human reading
}

func (t *Ticket) StructuredData() interface{} {
    // Return data for JSON serialization
}
```

### For slice results, create wrapper:
```go
type WorktreeListResult struct {
    Worktrees []git.WorktreeInfo
    Count     int
}

func (r *WorktreeListResult) TextRepresentation() string {
    // Format as table or list
}

func (r *WorktreeListResult) StructuredData() interface{} {
    // Return structured data
}
```

## Benefits
- Each result type owns its formatting logic (Single Responsibility)
- No central switch statement to maintain
- Easy to add new result types
- Follows kubectl's proven ResourcePrinter pattern
- Clean separation between business logic and presentation

## Architecture Reference
This follows the pattern used by:
- **kubectl**: ResourcePrinter interface with PrintObj method
- **docker**: Structured result types with formatters
- **git**: Separation between plumbing (data) and porcelain (formatting)

## Notes
Once all types implement Printable, we can completely remove the switch statement and have a much cleaner, more maintainable architecture.