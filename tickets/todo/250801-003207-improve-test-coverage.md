---
priority: 2
description: "Improve test coverage with comprehensive unit tests, table-driven tests, and benchmarks"
created_at: "2025-08-01T00:32:07+09:00"
started_at: null
closed_at: null
parent: 250801-002917-extract-interfaces-for-testability
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
- [ ] Add tests for error scenarios in `internal/ticket/manager.go`
- [ ] Add tests for edge cases in `internal/git/git.go`
- [ ] Add tests for validation logic
- [ ] Add tests for configuration parsing
- [ ] Achieve >80% test coverage

### Table-Driven Tests
- [ ] Convert complex function tests to table-driven format
- [ ] Add test cases for boundary conditions
- [ ] Include both positive and negative test cases
- [ ] Document test case purposes

### Benchmark Tests
- [ ] Add benchmarks for ticket listing operations
- [ ] Add benchmarks for file I/O operations
- [ ] Add benchmarks for string building operations
- [ ] Add benchmarks for search/filter operations

### Integration Tests
- [ ] Enhance integration tests for full workflows
- [ ] Add tests for error recovery scenarios
- [ ] Test concurrent operations
- [ ] Test with various configurations

### Quality Assurance
- [ ] Run `make coverage` and review report
- [ ] Ensure all new code has tests
- [ ] Document any untestable code
- [ ] Run `make test` with race detector
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update testing documentation
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

## Notes

This ticket depends on the interface extraction ticket being completed first, as mocks will make testing much easier.

Focus on testing behavior, not implementation details. Tests should be resilient to refactoring as long as the behavior remains the same.

Use test coverage as a guide, not a goal. 100% coverage doesn't mean bug-free code.