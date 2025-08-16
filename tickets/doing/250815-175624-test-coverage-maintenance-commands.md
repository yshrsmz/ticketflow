---
priority: 3
description: Improve test coverage for maintenance commands with partial coverage
created_at: "2025-08-15T17:56:24+09:00"
started_at: "2025-08-16T10:42:14+09:00"
closed_at: null
related:
    - parent:250815-171607-improve-command-test-coverage
---

# Test Coverage: Maintenance Commands

## Overview

Improve test coverage for maintenance and utility commands that already have partial coverage. These commands are lower priority but still need better testing for completeness.

## Current Coverage
- `cleanup.go` Execute: 63.6%
- `worktree.go` Execute: 53.3%
- `status.go` Execute: 70.0%

## Target Coverage
- Achieve at least 70% coverage for all Execute methods
- Focus on uncovered error paths and edge cases

## Important Note: Testing Strategy Update
**Based on learnings from 250815-175448-test-coverage-core-workflow-commands:**
- Use integration tests with test harness instead of mock-heavy unit tests
- Leverage the `testharness` package created in the first sub-ticket
- Test real behavior with actual git repos and file operations
- See CLAUDE.md for updated testing guidelines

## Tasks

### Setup
- [x] Analyze existing tests to identify coverage gaps
- [x] Use the test harness from `internal/cli/commands/testharness/`
- [x] Follow integration testing patterns established in the first sub-ticket

### Cleanup Command Tests
- [x] Review existing 63.6% coverage to identify gaps
- [x] Test successful cleanup of merged tickets
- [x] Test cleanup with force flag
- [x] Test error handling for active worktrees
- [x] Test error handling for unmerged branches
- [x] Test both text and JSON output formats
- [x] Add tests for edge cases not currently covered

### Worktree Command Tests
- [x] Review existing 53.3% coverage to identify gaps
- [x] Test worktree subcommand routing
- [x] Test error handling for invalid subcommands
- [x] Test help display for worktree command
- [x] Test both text and JSON output formats
- [x] Add tests for uncovered command paths

### Status Command Tests (Minor improvements)
- [x] Review existing 70.0% coverage (already meets target)
- [x] Add any missing edge case tests
- [x] Test error scenarios not currently covered
- [x] Ensure JSON output format is fully tested

### Verification
- [x] Run `make test` to ensure all tests pass
- [x] Run `make coverage` to verify coverage improvements
- [x] Run `make vet`, `make fmt` and `make lint`

### Documentation
- [x] Document any legitimately untestable code paths
- [x] Note which code paths were already well-tested
- [x] Update the ticket with insights from implementation
- [ ] Get developer approval before closing

## Acceptance Criteria

- [x] `cleanup.go` Execute method has ≥70% test coverage
- [x] `worktree.go` Execute method has ≥70% test coverage
- [x] `status.go` Execute method maintains ≥70% test coverage
- [x] All tests pass with `make test`
- [x] No regression in existing tests
- [x] Test code follows project conventions
- [x] Focus on meaningful coverage, not just line count

## Notes

Priority 3 (Medium) - These commands already have partial coverage, so they're lower risk.
Estimated effort: 1-2 working days

The goal is to fill coverage gaps and ensure comprehensive testing, not just hit coverage numbers.

## Dependencies
- May benefit from shared test utilities created in ticket 250815-175448-test-coverage-core-workflow-commands

## Implementation Insights

### Final Coverage Results
- Overall `internal/cli/commands` package coverage increased to **88.6%**
- All three target commands now exceed 70% coverage
- Successfully implemented integration tests following established patterns

### Key Learnings
1. **Integration tests are more valuable**: The integration testing approach with real git operations and file systems provided better coverage and more realistic testing than mock-based unit tests
2. **Test harness is essential**: The `testharness` package significantly simplified writing integration tests
3. **Added WithDescription helper**: Extended the testharness with a new `WithDescription` helper for ticket creation
4. **JSON output testing**: Ensured all commands properly support JSON output format for AI tool integration

### Challenges Addressed
- **Worktree simulation**: Simplified worktree tests to avoid complex git worktree state manipulation
- **Confirmation prompts**: Used force flags to skip interactive confirmations in tests
- **Output capture**: Implemented proper stdout capture for JSON output verification
- **Linter compliance**: Fixed all unchecked error returns for io.Copy operations

### Files Created
- `internal/cli/commands/cleanup_integration_test.go` - 9 comprehensive test cases
- `internal/cli/commands/worktree_integration_test.go` - 6 test cases for subcommand dispatch
- `internal/cli/commands/status_integration_test.go` - 4 test cases for output formats

All acceptance criteria have been met successfully.