---
priority: 2
description: Improve test coverage for command Execute methods
created_at: "2025-08-15T17:16:07+09:00"
started_at: "2025-08-15T17:44:10+09:00"
closed_at: null
related:
    - parent:250812-152927-migrate-remaining-commands
---

# Improve Command Test Coverage

Increase test coverage for command Execute methods. Current overall coverage is 42.8% with many Execute methods having 0% or low coverage.

## Current Coverage Status

### Commands with 0% Execute Coverage
- `new.go` Execute: 0.0%
- `restore.go` Execute: 0.0%
- `show.go` Execute: 0.0%
- `worktree_clean.go` Execute: 0.0%
- `worktree_list.go` Execute: 0.0%

### Commands with Low Execute Coverage
- `close.go` Execute: 29.2%
- `start.go` Execute: 43.8%
- `worktree.go` Execute: 53.3%
- `cleanup.go` Execute: 63.6%
- `status.go` Execute: 70.0%

### Commands with Good Coverage
- `list.go` Execute: 88.9%
- `version.go` Execute: 100.0%
- `help.go` Execute: 100.0%
- `init.go` Execute: 100.0%

## Tasks

- [ ] Add Execute method tests for `new` command
- [ ] Add Execute method tests for `restore` command
- [ ] Add Execute method tests for `show` command
- [ ] Add Execute method tests for `worktree_clean` command
- [ ] Add Execute method tests for `worktree_list` command
- [ ] Improve Execute method tests for `close` command
- [ ] Improve Execute method tests for `start` command
- [ ] Improve Execute method tests for `cleanup` command
- [ ] Run `make coverage` to verify improvement
- [ ] Aim for at least 70% coverage for all Execute methods
- [ ] Document any untestable code paths

## Testing Strategy

### For Each Command Test
1. Test successful execution with valid inputs
2. Test error cases (invalid flags, missing dependencies)
3. Test both text and JSON output formats
4. Test context cancellation handling (where applicable)
5. Use mocks for external dependencies (Git, file system)

### Example Test Structure
```go
func TestCommandExecute(t *testing.T) {
    t.Run("successful execution", func(t *testing.T) {
        // Setup mocks
        // Call Execute
        // Assert expected behavior
    })
    
    t.Run("handles invalid flags", func(t *testing.T) {
        // Test error handling
    })
    
    t.Run("context cancellation", func(t *testing.T) {
        // Test context handling
    })
}
```

## Success Criteria

- Overall test coverage > 60%
- All Execute methods have at least 70% coverage
- No regression in existing tests
- Clear documentation for any untestable paths

## Benefits

- Higher confidence in code reliability
- Better documentation through tests
- Easier refactoring with safety net
- Reduced bugs in production