---
priority: 2
description: "Migrate list, new, and start commands to new architecture"
created_at: "2025-08-10T00:30:01+09:00"
started_at: null
closed_at: null
---

# Task 2.4: Migrate First Commands

**Duration**: 2 days  
**Complexity**: Medium  
**Phase**: 2 - Command Architecture  
**Dependencies**: Task 2.2 (Command Registry), Task 2.3 (Worker Pool)

Migrate list, new, and start commands to demonstrate the new architecture. Maintain backward compatibility.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [ ] Refactor internal/cli/list.go to implement Command interface
- [ ] Refactor internal/cli/new.go to implement Command interface  
- [ ] Refactor internal/cli/start.go to implement Command interface
- [ ] Register commands in command registry
- [ ] Add performance metadata to each command
- [ ] Implement validation rules
- [ ] Update main.go to use registry for these commands
- [ ] Ensure backward compatibility
- [ ] Add integration tests for migrated commands
- [ ] Benchmark before/after performance
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation if necessary
- [ ] Update README.md
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Implementation Notes

- Start with most frequently used commands
- Keep existing handler functions during migration
- Verify no behavior changes
- Key files: internal/cli/{list,new,start}.go

## Expected Outcomes

- Proof of concept for new architecture
- Performance baseline for comparison
- Template for remaining command migrations