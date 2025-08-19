---
priority: 2
description: Complete migration of remaining result types to Printable interface
created_at: "2025-08-19T01:30:44+09:00"
started_at: "2025-08-19T14:22:12+09:00"
closed_at: null
related:
    - parent:250816-123703-improve-json-output-separation
---

# Complete Printable Migration

## Overview
Continue the migration to Printable interface pattern for the remaining result types. This is Phase 3 of the migration started in ticket 250816-203224-migrate-all-results-to-printable-interface.

## Background
Phase 1 and Phase 2 successfully migrated:
- TicketResult (wrapper for *ticket.Ticket)
- TicketListResult (enhanced with dynamic column width)
- WorktreeListResult (new result type)
- StatusResult (migrated from helper functions)
- StartResult (wrapper for StartTicketResult)
- CleanupResult (already implements Printable interface)

All implementations include comprehensive unit tests and maintain backward compatibility.

**Note**: CleanupResult already exists and implements the Printable interface (internal/cli/printable.go:59-94). The CleanWorktreesResult mentioned in the original plan appears to be an unused structure that doesn't need migration.

## Tasks

### Primary Tasks - Migrate Remaining Commands to Printable Pattern
- [ ] `new` command (internal/cli/commands/new.go:136-149)
  - [ ] Create NewTicketResult struct with typed fields
  - [ ] Implement TextRepresentation() and StructuredData()
  - [ ] Replace map[string]interface{} usage
  - [ ] Add comprehensive unit tests
- [ ] `close` command (internal/cli/commands/close.go:178-248)
  - [ ] Create CloseTicketResult struct with typed fields
  - [ ] Implement TextRepresentation() and StructuredData()
  - [ ] Replace map[string]interface{} usage and outputCloseErrorJSON helper
  - [ ] Add comprehensive unit tests
- [ ] `restore` command (internal/cli/commands/restore.go:123-157)
  - [ ] Create RestoreTicketResult struct with typed fields
  - [ ] Implement TextRepresentation() and StructuredData()
  - [ ] Replace map[string]interface{} usage
  - [ ] Add comprehensive unit tests

### Cleanup & Verification
- [ ] Remove the map[string]interface{} fallback case from output_writer.go (lines 102-109)
- [ ] Remove OutputWriter legacy wrapper from output_writer.go (lines 136-186)
- [ ] Update any remaining direct calls to Printf/PrintJSON to use PrintResult
- [ ] Run `make test` to verify all tests pass
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update the ticket with implementation insights
- [ ] Get developer approval before closing

## Implementation Guidelines

Follow the established patterns from Phase 1 and 2:

1. **Use wrapper pattern for existing types** - Don't modify domain models
2. **Preserve exact output format** - Backward compatibility is critical
3. **Add comprehensive unit tests** - Test both text and JSON output
4. **Use buffer pre-allocation** - Use the defined constants (smallBufferSize, mediumBufferSize, largeBufferSize)
5. **Document with comments** - Explain the purpose of each Printable implementation

## Success Criteria
- The 3 remaining commands (new, close, restore) use typed Printable structs instead of map[string]interface{}
- Map fallback case removed from textOutputFormatter.PrintResult() (output_writer.go:102-109)
- OutputWriter legacy wrapper completely removed (output_writer.go:136-186)
- All unit tests pass with comprehensive coverage of new result types
- Integration tests confirm backward compatibility
- Output format remains unchanged from user perspective
- Code passes gofmt, go vet, and golangci-lint checks

## References
- Parent ticket: 250816-123703-improve-json-output-separation
- Previous work: 250816-203224-migrate-all-results-to-printable-interface
- PR #80: Initial migration implementation