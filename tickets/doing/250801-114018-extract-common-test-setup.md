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

- [x] Create test helper package or file for common utilities
  - Created `internal/testutil` package with multiple specialized files
- [x] Extract common App setup for CLI tests
  - Created `MockSetup` struct in `mocks.go` with comprehensive mock management
- [x] Extract common Model setup for UI tests
  - Included in `MockSetup` with output capture utilities
- [x] Create test fixture builders for common data structures
  - Ticket fixtures with various states (using functional options pattern)
  - Config fixtures for different scenarios
  - Git repository fixtures with helper methods
- [x] Extract common assertion helpers
  - JSON output assertions
  - Error message assertions
  - File/directory existence assertions
  - Output content assertions
- [x] Create test data generators for table-driven tests
  - `GenerateTicketID` for unique IDs
  - `TicketContent` for generating frontmatter
  - Various fixture functions with options
- [x] Update existing tests to use the new helpers
  - Refactored `cmd/ticketflow/test_helpers.go`
  - Updated `internal/cli/test_helpers.go` with deprecation notices
- [x] Document test helper usage in test README
  - Created comprehensive README.md with examples and migration guide
- [x] Run `make test` to ensure all tests still pass
  - All tests passing after refactoring
- [x] Address golang-pro review feedback
  - Fixed critical bugs (priority conversion, type assertions)
  - Enhanced error handling with stderr capture
  - Added thread safety to OutputCapture
  - Improved code organization
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

## Implementation Details

### Package Structure
Created `internal/testutil` package with the following files:
- **fixtures.go**: Factory functions for tickets and configs using functional options pattern
- **git.go**: Git repository setup and operations with safety measures
- **mocks.go**: Centralized mock management with expectation helpers
- **filesystem.go**: File/directory operations and project setup utilities
- **context.go**: Context testing utilities for cancellation scenarios
- **assertions.go**: Common assertion helpers for output and errors
- **output.go**: Thread-safe output capture (enhanced existing file)

### Key Improvements from Review
1. **Fixed Critical Bugs**:
   - Priority string conversion using `strconv.Itoa()` instead of unsafe character arithmetic
   - Safe type handling in `TicketContent` with nil checks and proper assertions
   - Fixed string trimming using `strings.TrimSpace()`

2. **Enhanced Error Handling**:
   - Added stderr capture for git commands for better debugging
   - Created `execCommand` helper that captures both stdout and stderr
   - Improved error messages with detailed git failure information

3. **Thread Safety**:
   - Added mutex protection to `OutputCapture` for concurrent test execution
   - Prevents race conditions in parallel test runs

### Safety Measures
- **Git Configuration**: Always configures git locally in test directories, never globally
- **Path Validation**: Test utilities work within designated test directories
- **Resource Cleanup**: Automatic cleanup using `t.Cleanup()` for temporary resources

### Migration Impact
- Minimal changes required to existing tests
- Deprecated old helper functions with clear migration path
- Backward compatibility maintained where possible

## Notes

Suggested by golang-pro agent during code review. This is a lower priority improvement but will significantly improve test maintainability as the codebase grows.

### Future Enhancements
- Consider adding TestContext struct to bundle common test resources
- Add batch git operations for performance
- Implement config parameter usage in CreateConfigFile
- Add more specialized builders for complex test scenarios