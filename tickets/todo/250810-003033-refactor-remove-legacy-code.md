---
priority: 2
description: "Remove old switch statement and handler functions"
created_at: "2025-08-10T00:30:33+09:00"
started_at: null
closed_at: null
---

# Task 5.2: Remove Legacy Code

**Duration**: 1 day  
**Complexity**: Low  
**Phase**: 5 - Migration and Cleanup  
**Dependencies**: Task 5.1 (Complete Command Migration)

Remove the old 300+ line switch statement and all unused handler functions.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [ ] Remove switch statement from main.go
- [ ] Delete unused handler functions
- [ ] Clean up deprecated interfaces
- [ ] Remove old command execution code
- [ ] Update imports and dependencies
- [ ] Run dead code detection tools
- [ ] Ensure no orphaned code remains
- [ ] Verify all tests still pass
- [ ] Check code coverage hasn't dropped
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation if necessary
- [ ] Update README.md
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Implementation Notes

- Key files: cmd/ticketflow/main.go, internal/cli/
- Use tools to detect dead code
- Be careful not to remove still-used utilities
- Keep git history clean with meaningful commits

## Expected Outcomes

- Cleaner, more maintainable codebase
- Reduced binary size
- Easier to understand main.go