---
priority: 2
description: "Complete migration of remaining result types to Printable interface"
created_at: "2025-08-19T01:30:44+09:00"
started_at: null
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

All implementations include comprehensive unit tests and maintain backward compatibility.

## Tasks

### High Priority - Complete Existing Result Types
- [ ] Create `WorktreeCleanResult` wrapper for `CleanWorktreesResult` and implement Printable
  - Wrap the existing CleanWorktreesResult struct
  - Implement TextRepresentation() and StructuredData()
  - Add comprehensive unit tests

### Medium Priority - Migrate Map-Based Results
- [ ] Analyze and migrate map-based results to typed structs:
  - [ ] `new` command - Create NewTicketResult struct
  - [ ] `close` command - Create CloseTicketResult struct  
  - [ ] `restore` command - Create RestoreResult struct
- [ ] Each should:
  - Define proper struct with typed fields
  - Implement Printable interface
  - Include unit tests
  - Maintain exact output format compatibility

### Final Steps
- [ ] Remove the map[string]interface{} fallback case once all commands are migrated
- [ ] Remove OutputWriter legacy wrapper once fully migrated
- [ ] Run `make test` to verify all tests pass
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update architecture documentation with final Printable pattern
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Implementation Guidelines

Follow the established patterns from Phase 1 and 2:

1. **Use wrapper pattern for existing types** - Don't modify domain models
2. **Preserve exact output format** - Backward compatibility is critical
3. **Add comprehensive unit tests** - Test both text and JSON output
4. **Use buffer pre-allocation** - Use the defined constants (smallBufferSize, mediumBufferSize, largeBufferSize)
5. **Document with comments** - Explain the purpose of each Printable implementation

## Success Criteria
- All commands return Printable types (no direct Printf/PrintJSON calls)
- Switch statement in output_writer.go is completely removed
- All tests pass including integration tests
- Output format remains unchanged from user perspective
- Code passes gofmt, go vet, and golangci-lint checks

## References
- Parent ticket: 250816-123703-improve-json-output-separation
- Previous work: 250816-203224-migrate-all-results-to-printable-interface
- PR #80: Initial migration implementation