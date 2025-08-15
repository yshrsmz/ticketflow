---
priority: 2
description: Improve test coverage for command Execute methods
created_at: "2025-08-15T17:16:07+09:00"
started_at: "2025-08-15T17:44:10+09:00"
closed_at: null
related:
    - parent:250812-152927-migrate-remaining-commands
    - "blocks:250815-175448-test-coverage-core-workflow-commands"
    - "blocks:250815-175548-test-coverage-zero-coverage-commands"
    - "blocks:250815-175624-test-coverage-maintenance-commands"
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

## Status Update

After analysis, this ticket was determined to be too large for a single implementation (estimated 5-7 days of work). It has been split into three manageable sub-tickets:

### Sub-tickets Created

1. **250815-175448-test-coverage-core-workflow-commands** (Priority 1)
   - Focus: `start` (43.8%) and `close` (29.2%) commands
   - Critical workflow commands that need immediate attention
   - Estimated: 2 days

2. **250815-175548-test-coverage-zero-coverage-commands** (Priority 2)
   - Focus: `new`, `restore`, `show`, `worktree_clean`, `worktree_list` (all at 0%)
   - Commands with zero coverage need comprehensive testing
   - Estimated: 2-3 days

3. **250815-175624-test-coverage-maintenance-commands** (Priority 3)
   - Focus: `cleanup` (63.6%), `worktree` (53.3%), `status` (70.0%)
   - Commands with partial coverage need improvement
   - Estimated: 1-2 days

## Tasks

- [x] Analyze scope and determine if ticket should be split
- [x] Create sub-ticket for core workflow commands (start, close)
- [x] Create sub-ticket for zero coverage commands
- [x] Create sub-ticket for maintenance commands
- [x] Complete sub-ticket: 250815-175448-test-coverage-core-workflow-commands ✅
- [ ] Complete sub-ticket: 250815-175548-test-coverage-zero-coverage-commands
- [ ] Complete sub-ticket: 250815-175624-test-coverage-maintenance-commands
- [ ] Run `make coverage` to verify overall improvement
- [ ] Verify all Execute methods have at least 70% coverage
- [ ] Close parent ticket after all sub-tickets complete

## Testing Strategy (UPDATED)

### Important Discovery
After implementing tests for the first sub-ticket (250815-175448), we discovered that mock-heavy unit testing for Execute methods is fundamentally flawed. Following consultation with architectural patterns from tools like git, docker, and kubectl, we've shifted to an integration-first testing approach.

### New Strategy: Integration Testing for Execute Methods
1. **Execute methods are orchestrators** - They coordinate multiple components and should be tested as integrated units
2. **Use test harness with real environments** - Create actual git repos, files, and configurations in temp directories
3. **Test user-visible behavior** - Focus on what users experience, not internal method calls
4. **Mock only at system boundaries** - Only mock things like network calls or system resources when absolutely necessary

### Test Harness Approach
```go
// Create real test environment
env := testharness.NewTestEnvironment(t)
env.CreateTicket("test-ticket", ticket.StatusTodo)
env.RunGit("add", ".")
env.RunGit("commit", "-m", "Setup")

// Execute actual command
cmd := commands.NewStartCommand()
err := cmd.Execute(ctx, flags, []string{"test-ticket"})

// Verify real outcomes
assert.True(t, env.FileExists("tickets/doing/test-ticket.md"))
assert.Contains(t, env.LastCommitMessage(), "Start ticket")
```

### Coverage Achieved (First Sub-ticket)
- `close.go` Execute: **91.7%** (up from 29.2%) ✅
- `start.go` Execute: **94.4%** (up from 43.8%) ✅

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