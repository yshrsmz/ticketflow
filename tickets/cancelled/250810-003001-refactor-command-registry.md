---
priority: 2
description: "Build command registry with O(1) lookup and self-registration"
created_at: "2025-08-10T00:30:01+09:00"
started_at: null
closed_at: null
---

# Task 2.2: Command Registry Implementation

**Duration**: 2 days  
**Complexity**: Medium  
**Phase**: 2 - Command Architecture  
**Dependencies**: Task 2.1 (Command Interface)

Build command registry alongside existing switch statement. Implement self-registering commands using init() functions.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [ ] Create internal/command/registry.go
- [ ] Implement Registry struct with map[string]Command
- [ ] Add Register() and Get() methods
- [ ] Support command aliases
- [ ] Implement help text generation
- [ ] Add init() functions for command registration
- [ ] Create factory pattern for command instantiation
- [ ] Add unit tests for registry operations
- [ ] Update main.go to use registry alongside switch
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation if necessary
- [ ] Update README.md
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Implementation Notes

- Use singleton pattern for global registry
- O(1) lookup performance with map
- Reference git's command registration pattern
- Keep switch statement during migration

## Expected Outcomes

- Eliminate 300+ line switch statement
- Self-documenting command system
- Easy to add new commands