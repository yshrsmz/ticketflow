---
priority: 2
description: Add test coverage for commands with 0% Execute method coverage
created_at: "2025-08-15T17:55:48+09:00"
started_at: "2025-08-16T00:24:30+09:00"
closed_at: null
related:
    - parent:250815-171607-improve-command-test-coverage
---

# Test Coverage: Zero Coverage Commands

## Overview

Add comprehensive test coverage for commands that currently have 0% Execute method coverage. These commands are simpler than the core workflow commands but still need proper testing to ensure reliability.

## Current Coverage
- `new.go` Execute: 0.0%
- `restore.go` Execute: 0.0%
- `show.go` Execute: 0.0%
- `worktree_clean.go` Execute: 0.0%
- `worktree_list.go` Execute: 0.0%

## Target Coverage
- Achieve at least 70% coverage for all Execute methods
- Cover all major code paths and error scenarios

## Important Note: Testing Strategy Update
**Based on learnings from 250815-175448-test-coverage-core-workflow-commands:**
- Use integration tests with test harness instead of mock-heavy unit tests
- Leverage the `testharness` package created in the first sub-ticket
- Test real behavior with actual git repos and file operations
- See CLAUDE.md for updated testing guidelines

## Tasks

### Setup
- [x] Review the new test harness in `internal/cli/commands/testharness/`
- [x] Follow integration testing patterns from close_integration_test.go and start_integration_test.go

### New Command Tests
- [x] Test successful ticket creation with valid slug
- [x] Test creation with parent ticket flag
- [x] Test error handling for invalid slug format
- [x] Test error handling for duplicate tickets
- [x] Test both text and JSON output formats
- [x] Test template application

### Restore Command Tests
- [x] Test successful ticket restoration from done to todo
- [x] Test error handling for non-existent tickets
- [x] Test error handling for tickets not in done status
- [x] Test both text and JSON output formats
- [x] Test file operations and state transitions

### Show Command Tests
- [x] Test successful display of existing ticket
- [x] Test error handling for non-existent tickets
- [x] Test both text and JSON output formats
- [x] Test markdown rendering
- [x] Test metadata display

### Worktree Clean Command Tests
- [x] Test successful cleanup of merged worktrees
- [x] Test handling of unmerged branches
- [x] Test error handling for missing worktrees
- [x] Test both text and JSON output formats
- [x] Test dry-run mode if available

### Worktree List Command Tests
- [x] Test listing all worktrees
- [x] Test filtering by status
- [x] Test error handling when no worktrees exist
- [x] Test both text and JSON output formats
- [x] Test sorting and formatting

### Verification
- [x] Run `make test` to ensure all tests pass
- [x] Run `make coverage` to verify coverage improvements
- [x] Run `make vet`, `make fmt` and `make lint`

### Documentation
- [ ] Document any legitimately untestable code paths
- [ ] Add comments for complex test setups
- [ ] Update the ticket with insights from implementation
- [ ] Get developer approval before closing

## Implementation Summary

Successfully added comprehensive integration tests for all 5 commands with 0% Execute method coverage:

### Coverage Achieved:
- `new.go` Execute: **85.7%** ✅
- `restore.go` Execute: **95.0%** ✅
- `show.go` Execute: **92.3%** ✅
- `worktree_clean.go` Execute: **75.0%** ✅
- `worktree_list.go` Execute: **90.0%** ✅

### Key Implementation Details:
1. Used integration testing approach with real git repos and file operations
2. Leveraged existing testharness package from parent ticket
3. Followed table-driven test patterns for comprehensive scenario coverage
4. Fixed several test expectations to match actual command behavior:
   - Restore command requires being on a ticket branch
   - Worktree clean only keeps worktrees for "doing" status tickets
   - Error messages are case-sensitive
5. All tests pass with `make test`, `make vet`, `make fmt`, and `make lint`

### Files Created/Modified:
- `worktree_clean_integration_test.go` (198 lines)
- `worktree_list_integration_test.go` (294 lines)
- `show_integration_test.go` (278 lines)
- `restore_integration_test.go` (347 lines)
- `new_integration_test.go` (397 lines)

Total: ~1,514 lines of comprehensive integration tests

## Acceptance Criteria

- [x] All 5 commands have ≥70% Execute method test coverage
- [x] All tests pass with `make test`
- [x] No regression in existing tests
- [x] Test code follows project conventions
- [x] Uses table-driven tests for multiple scenarios
- [x] Mock dependencies are properly isolated (used integration tests instead)

## Notes

Priority 2 (High) - These commands have zero coverage and need immediate attention.
Estimated effort: 2-3 working days (0.5 days per command average)

## Dependencies
- Requires shared test utilities from ticket 250815-175448-test-coverage-core-workflow-commands