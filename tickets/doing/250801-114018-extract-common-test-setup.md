---
priority: 4
description: Extract common test setup code to reduce duplication in tests
created_at: "2025-08-01T11:40:18+09:00"
started_at: "2025-08-05T00:46:13+09:00"
closed_at: null
related:
    - parent:250801-003010-decompose-large-functions
---

# Extract Common Test Setup

Extract common test setup code into helper functions to reduce duplication and improve test maintainability.

## Context

Following the function decomposition work and addition of new tests, the golang-pro agent suggested extracting common test setup patterns. This will reduce code duplication across test files and make tests easier to maintain.

## Tasks

- [ ] Create test helper package or file for common utilities
- [ ] Extract common App setup for CLI tests
  ```go
  func setupTestApp(t *testing.T) *App {
      mockGit := new(mocks.MockGitClient)
      mockManager := new(mocks.MockTicketManager)
      return &App{
          Config:  testConfig(),
          Git:     mockGit,
          Manager: mockManager,
      }
  }
  ```
- [ ] Extract common Model setup for UI tests
- [ ] Create test fixture builders for common data structures
  - Ticket fixtures with various states
  - Config fixtures for different scenarios
  - Git worktree fixtures
- [ ] Extract common assertion helpers
- [ ] Create test data generators for table-driven tests
- [ ] Update existing tests to use the new helpers
- [ ] Document test helper usage in test README
- [ ] Run `make test` to ensure all tests still pass
- [ ] Get developer approval before closing

## Example Test Helpers

```go
// testhelpers/fixtures.go
package testhelpers

// TicketFixture creates a test ticket with sensible defaults
func TicketFixture(opts ...TicketOption) *ticket.Ticket {
    t := &ticket.Ticket{
        ID:          "test-ticket-123",
        Path:        "/test/tickets/todo/test-ticket-123.md",
        Description: "Test ticket",
        Priority:    1,
        CreatedAt:   ticket.RFC3339Time{Time: time.Now()},
    }
    for _, opt := range opts {
        opt(t)
    }
    return t
}

// TicketOption modifies a ticket fixture
type TicketOption func(*ticket.Ticket)

// WithStatus sets the ticket status
func WithStatus(status ticket.Status) TicketOption {
    return func(t *ticket.Ticket) {
        switch status {
        case ticket.StatusDoing:
            now := time.Now()
            t.StartedAt = ticket.RFC3339TimePtr{Time: &now}
        case ticket.StatusDone:
            now := time.Now()
            t.StartedAt = ticket.RFC3339TimePtr{Time: &now}
            t.ClosedAt = ticket.RFC3339TimePtr{Time: &now}
        }
    }
}
```

## Benefits

- Reduced code duplication in tests
- Consistent test setup across the codebase
- Easier to add new tests
- Single place to update when data structures change
- Better test readability

## Acceptance Criteria

- Common test patterns are extracted into reusable helpers
- No test logic is duplicated across files
- Test helpers are well-documented
- All tests continue to pass
- Test coverage is maintained or improved

## Notes

Suggested by golang-pro agent during code review. This is a lower priority improvement but will significantly improve test maintainability as the codebase grows.