---
priority: 1
description: Extract interfaces from concrete implementations to improve testability and enable dependency injection
created_at: "2025-08-01T00:29:17+09:00"
started_at: "2025-08-01T00:42:50+09:00"
closed_at: null
---

# Extract Interfaces for Testability

Extract interfaces from concrete implementations to improve testability and enable dependency injection throughout the codebase. This is a foundational change that will enable better unit testing and make the codebase more maintainable.

## Context

Currently, the codebase has tightly coupled concrete implementations which make unit testing difficult. By extracting interfaces and implementing dependency injection, we can:
- Create mock implementations for testing
- Improve code flexibility and maintainability
- Follow Go best practices for interface design
- Enable better separation of concerns

## Tasks

### Interface Extraction
- [ ] Extract `TicketManager` interface from the concrete implementation in `internal/ticket/manager.go`
- [ ] Extract `GitClient` interface from the concrete implementation in `internal/git/git.go`
- [ ] Create `internal/ticket/interfaces.go` for ticket-related interfaces
- [ ] Create `internal/git/interfaces.go` for git-related interfaces

### Dependency Injection
- [ ] Implement dependency injection in `cmd/ticketflow/main.go`
- [ ] Update CLI commands in `internal/cli/` to use injected dependencies
- [ ] Update TUI application in `internal/ui/app.go` to use injected dependencies

### Testing Infrastructure
- [ ] Create mock implementations in `internal/mocks/` directory
- [ ] Add unit tests demonstrating the use of mocks
- [ ] Ensure all existing tests pass

### Quality Assurance
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation if necessary
- [ ] Update README.md
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Implementation Details

### Files to Modify
- `/internal/ticket/manager.go` - Extract TicketManager interface
- `/internal/git/git.go` - Extract GitClient interface
- `/internal/ticket/interfaces.go` - New file for interfaces
- `/internal/git/interfaces.go` - New file for interfaces
- `/internal/mocks/` - New directory for mock implementations
- `/cmd/ticketflow/main.go` - Implement dependency injection
- `/internal/cli/commands.go` - Update to use injected dependencies
- `/internal/ui/app.go` - Update to use injected dependencies

### Interface Design Guidelines
- Keep interfaces small and focused (Interface Segregation Principle)
- Define interfaces where they are used, not where they are implemented
- Use descriptive method names that express intent
- Return errors as the last return value
- Avoid unnecessary abstraction - only extract what's needed for testing

## Notes

This is a high-priority foundational change that will unblock many other improvements. It should be completed before decomposing large functions or other refactoring work.

Key interfaces to extract:
- `TicketManager`: Core ticket operations (Create, Start, Close, List, etc.)
- `GitClient`: Git operations (CreateWorktree, RemoveWorktree, CreateBranch, etc.)