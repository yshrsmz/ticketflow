---
priority: 2
description: Improve test coverage with comprehensive unit tests, table-driven tests, and benchmarks
created_at: "2025-08-01T00:32:07+09:00"
started_at: "2025-08-02T23:45:43+09:00"
closed_at: null
related:
    - parent:250801-002917-extract-interfaces-for-testability
---

# Improve Test Coverage

Enhance test coverage throughout the codebase with comprehensive unit tests, table-driven tests for complex logic, and benchmarks for performance-critical operations.

## Context

Current test coverage has gaps:
- Missing error scenario tests
- Limited edge case coverage
- No benchmarks for performance-critical code
- Some packages have minimal or no tests
- Tests don't cover all code paths

Comprehensive testing will:
- Catch bugs before they reach production
- Enable confident refactoring
- Document expected behavior
- Identify performance regressions
- Improve code quality

## Tasks

### Unit Test Coverage
- [x] Add comprehensive tests for CLI commands in `cmd/ticketflow`
- [x] Add tests for error scenarios in `internal/cli`
- [x] Add tests for output formatting functions
- [x] Add tests for cleanup functionality
- [x] Achieve significant coverage improvements:
  - cmd/ticketflow: 7.1% → 51.3%
  - internal/cli: 29.3% → 43.4%

### Table-Driven Tests
- [x] Convert complex function tests to table-driven format
- [x] Add test cases for boundary conditions
- [x] Include both positive and negative test cases
- [x] Document test case purposes

### Benchmark Tests
- [x] Add benchmarks for formatDuration function
- [x] Add benchmarks for ticketToJSON function
- [x] Add benchmarks for parseOutputFormat function
- [x] Add allocation reporting to all benchmarks

### Code Quality Improvements
- [x] Implement golang-pro code review suggestions
- [x] Fix resource management (file descriptor leaks)
- [x] Improve test organization with test_helpers.go
- [x] Add test fixtures and constants
- [x] Fix environment variable handling with t.Setenv
- [x] Add parallel test execution where safe
- [x] Create package-level test documentation

### Quality Assurance
- [x] Run all tests and ensure they pass
- [x] Ensure all new code has tests
- [x] Fix test failures by removing parallel execution where needed
- [x] Run `make test` successfully
- [ ] Get developer approval before closing

## Implementation Guidelines

### Table-Driven Test Pattern
```go
func TestTicketManager_StartTicket(t *testing.T) {
    tests := []struct {
        name    string
        ticketID string
        setup   func(*testing.T, *Manager)
        wantErr bool
        errMsg  string
    }{
        {
            name:     "valid ticket",
            ticketID: "123-valid",
            setup:    func(t *testing.T, m *Manager) { /* create ticket */ },
            wantErr:  false,
        },
        {
            name:     "empty ticket ID",
            ticketID: "",
            wantErr:  true,
            errMsg:   "ticket ID cannot be empty",
        },
        {
            name:     "ticket not found",
            ticketID: "nonexistent",
            wantErr:  true,
            errMsg:   "ticket not found",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            m := NewManager()
            if tt.setup != nil {
                tt.setup(t, m)
            }
            
            err := m.StartTicket(tt.ticketID)
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Benchmark Pattern
```go
func BenchmarkListTickets(b *testing.B) {
    m := setupManagerWithTickets(b, 1000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := m.ListTickets(StatusTodo)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Mock Usage Pattern
```go
func TestManagerWithMockGit(t *testing.T) {
    mockGit := mocks.NewMockGitClient(t)
    mockGit.On("CreateWorktree", mock.Anything).Return(nil)
    
    m := NewManager(WithGitClient(mockGit))
    // Test manager behavior
    
    mockGit.AssertExpectations(t)
}
```

## Insights from Implementation

### Key Achievements
1. **Significant Coverage Improvements**: Increased test coverage dramatically for critical packages:
   - cmd/ticketflow: 7.1% → 51.3% (7x improvement)
   - internal/cli: 29.3% → 43.4% (1.5x improvement)

2. **Test Organization**: Created dedicated test helper files to reduce duplication and improve maintainability

3. **Resource Management**: Fixed file descriptor leaks in stdout/stderr capture by implementing proper defer cleanup

4. **Parallel Test Limitations**: Discovered that tests using `os.Chdir` cannot run in parallel as they modify global state

### Challenges Resolved
1. **Git Configuration Issues**: Tests were failing due to missing git user configuration. Resolved by setting global git config in test setup
2. **Mock Type Mismatches**: Fixed issues where mocks were returning `[]*ticket.Ticket` instead of `[]ticket.Ticket`
3. **Missing Mock Expectations**: Added comprehensive mock expectations for all test scenarios

### Best Practices Applied
- Used table-driven tests for comprehensive coverage
- Added benchmarks with allocation reporting
- Used `t.Setenv` for automatic environment variable cleanup
- Created test fixtures to standardize test data
- Extracted magic values into named constants
- Added package-level documentation for test organization

### Code Review Integration
Successfully integrated golang-pro agent code review suggestions, which significantly improved test quality and maintainability.

## Notes

This ticket depends on the interface extraction ticket being completed first, as mocks will make testing much easier.

Focus on testing behavior, not implementation details. Tests should be resilient to refactoring as long as the behavior remains the same.

Use test coverage as a guide, not a goal. 100% coverage doesn't mean bug-free code.