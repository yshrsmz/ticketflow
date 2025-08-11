---
priority: 2
description: "Migrate all remaining commands to new architecture"
created_at: "2025-08-10T00:30:33+09:00"
started_at: null
closed_at: null
---

# Task 5.1: Complete Command Migration

**Duration**: 3 days  
**Complexity**: Medium  
**Phase**: 5 - Migration and Cleanup  
**Dependencies**: All Phase 2 tasks

Migrate all remaining commands to the new command registry architecture.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [ ] List all remaining commands to migrate
- [ ] Migrate each command to implement Command interface
- [ ] Register all commands in the registry
- [ ] Update main.go to use registry exclusively
- [ ] Ensure consistent error handling
- [ ] Update help text for all commands
- [ ] Add integration tests for all commands
- [ ] Verify backward compatibility
- [ ] Test command aliases
- [ ] Document any behavior changes
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation if necessary
- [ ] Update README.md
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Implementation Notes

- Key files: All files in internal/cli/
- Test each command thoroughly after migration
- Maintain consistent UX across commands
- Use lessons learned from first command migrations

## Expected Outcomes

- All commands using new architecture
- Consistent command behavior
- Ready for legacy code removal