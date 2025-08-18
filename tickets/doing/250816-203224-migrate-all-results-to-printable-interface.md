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
- `TicketListResult` already implements Printable ✅
- OutputWriter checks for Printable first, falls back to switch ✅

## Actual Result Types in Use
Based on codebase analysis, these are the result types currently being handled:

### In Switch Statement:
1. `*ticket.Ticket` - Used by show command
2. `[]*ticket.Ticket` - Used by list command (should use TicketListResult)
3. `map[string]interface{}` - Generic fallback for various commands

### Direct Output (No Result Types):
- `worktree list` - Uses Printf/PrintJSON directly
- `status` - Uses helper functions directly
- `start`, `close`, `new`, `restore` - Return maps directly for JSON

### Existing but Not in Switch:
- `StartTicketResult` - Handled inline in start command
- `CleanWorktreesResult` - Handled inline in worktree clean command

## Tasks Completed (Phase 1)
- [x] Create `TicketResult` wrapper for single `*ticket.Ticket` and implement Printable
  - **Important**: Don't modify ticket.Ticket directly - it's a domain model
- [x] Update list command to use existing `TicketListResult` instead of `[]*ticket.Ticket`
- [x] Create `WorktreeListResult` struct and implement Printable
  - Migrate worktree list command from direct Printf to result type
- [x] Update commands to return Printable types:
  - [x] `show` command - return TicketResult
  - [x] `list` command - return TicketListResult (enhanced with dynamic column width)
  - [x] `worktree list` - return WorktreeListResult
- [x] Remove migrated cases from the switch statement in textResultWriter.PrintResult
- [x] Remove unused `outputTicketListText` and `outputTicketListJSON` functions
- [x] Keep `map[string]interface{}` case as final fallback
- [x] Run `make test` to verify all output still works
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Address PR review feedback from Copilot
- [x] Performance optimizations (use fmt.Fprintf directly to strings.Builder)
- [x] Remove legacy `[]*ticket.Ticket` case and `printTicketList` method completely
  - No production code uses this legacy path
  - Removed test cases for the legacy path
  - All tests continue to pass

## Tasks Remaining (Future Work)
- [ ] Create `StatusResult` struct and implement Printable
  - Migrate status command from helper functions to result type
- [ ] Create `StartResult` wrapper for `StartTicketResult` and implement Printable
- [ ] Create `WorktreeCleanResult` wrapper for `CleanWorktreesResult` and implement Printable
- [ ] Update remaining commands to return Printable types:
  - `status` - return StatusResult
  - `start` - return StartResult
  - `worktree clean` - return WorktreeCleanResult
- [ ] Migrate simple map results to typed structs where appropriate:
  - `new`, `close`, `restore` commands
- [ ] Update documentation about the Printable pattern

## Implementation Pattern

### For domain models, create wrapper:
```go
// DON'T modify ticket.Ticket directly
// Instead, create a wrapper:
type TicketResult struct {
    Ticket *ticket.Ticket
}

func (r *TicketResult) TextRepresentation() string {
    // Format for human reading (preserve existing format)
}

func (r *TicketResult) StructuredData() interface{} {
    return r.Ticket // Return the ticket for JSON serialization
}
```

### For commands with direct output, create result type:
```go
type WorktreeListResult struct {
    Worktrees []git.WorktreeInfo
    Count     int
}

func (r *WorktreeListResult) TextRepresentation() string {
    // Move Printf logic here, preserve exact format
}

func (r *WorktreeListResult) StructuredData() interface{} {
    return struct {
        Worktrees []git.WorktreeInfo `json:"worktrees"`
        Count     int                `json:"count"`
    }{
        Worktrees: r.Worktrees,
        Count:     r.Count,
    }
}
```

### For existing result types, create Printable wrapper:
```go
type StartResult struct {
    *StartTicketResult
}

func (r *StartResult) TextRepresentation() string {
    // Format the start result for text output
}

func (r *StartResult) StructuredData() interface{} {
    return r.StartTicketResult
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

## Implementation Insights

### Key Learnings
1. **Wrapper Pattern Success**: Creating wrappers (like TicketResult) for domain models keeps the domain layer pure while adding presentation concerns separately.

2. **Performance Considerations**:
   - Using `fmt.Fprintf` directly to `strings.Builder` avoids intermediate string allocations
   - Two-pass algorithms (find max, then format) can be cleaner than single-pass with intermediate storage
   - For typical CLI use cases (10-100 items), code clarity > micro-optimizations

3. **Testing Challenges**:
   - When migrating to new patterns, test mocks need updates (e.g., ListTickets now calls Manager.List twice for summary)
   - Test expectations must be updated to match new output formats

4. **Code Review Insights**:
   - Named constants (GitSHAFullLength, GitSHAShortLength) improve maintainability over magic numbers
   - Dead code removal is important - unreachable switch cases should be removed
   - Not all performance suggestions are improvements - clarity often wins

5. **Migration Strategy**:
   - Incremental migration with fallback mechanism allows gradual adoption
   - Can migrate high-value/frequently-used commands first
   - Backward compatibility maintained throughout migration

### Design Decisions
- **Dynamic vs Fixed Column Width**: Chose dynamic width for better UX with varying ID lengths
- **Summary in Text Output**: Removed from TicketListResult text output for cleaner separation (summary is metadata, not list content)
- **Direct Output vs Buffering**: Chose direct formatting over buffering all lines for better memory efficiency

### Future Considerations
- Consider creating a `ResultBuilder` helper for common formatting patterns
- May want to standardize error representation in Printable implementations
- Could add formatting options (verbose, compact) to Printable interface

## Notes
Once all types implement Printable, we can completely remove the switch statement and have a much cleaner, more maintainable architecture. The PR #80 demonstrates a solid foundation for continuing this pattern.