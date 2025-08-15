---
priority: 3
description: "Improve test coverage for maintenance commands with partial coverage"
created_at: "2025-08-15T17:56:24+09:00"
started_at: null
closed_at: null
related:
    - "parent:250815-171607-improve-command-test-coverage"
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

## Tasks

### Setup
- [ ] Analyze existing tests to identify coverage gaps
- [ ] Reuse test utilities from other test improvement tickets

### Cleanup Command Tests
- [ ] Review existing 63.6% coverage to identify gaps
- [ ] Test successful cleanup of merged tickets
- [ ] Test cleanup with force flag
- [ ] Test error handling for active worktrees
- [ ] Test error handling for unmerged branches
- [ ] Test both text and JSON output formats
- [ ] Add tests for edge cases not currently covered

### Worktree Command Tests
- [ ] Review existing 53.3% coverage to identify gaps
- [ ] Test worktree subcommand routing
- [ ] Test error handling for invalid subcommands
- [ ] Test help display for worktree command
- [ ] Test both text and JSON output formats
- [ ] Add tests for uncovered command paths

### Status Command Tests (Minor improvements)
- [ ] Review existing 70.0% coverage (already meets target)
- [ ] Add any missing edge case tests
- [ ] Test error scenarios not currently covered
- [ ] Ensure JSON output format is fully tested

### Verification
- [ ] Run `make test` to ensure all tests pass
- [ ] Run `make coverage` to verify coverage improvements
- [ ] Run `make vet`, `make fmt` and `make lint`

### Documentation
- [ ] Document any legitimately untestable code paths
- [ ] Note which code paths were already well-tested
- [ ] Update the ticket with insights from implementation
- [ ] Get developer approval before closing

## Acceptance Criteria

- [ ] `cleanup.go` Execute method has ≥70% test coverage
- [ ] `worktree.go` Execute method has ≥70% test coverage
- [ ] `status.go` Execute method maintains ≥70% test coverage
- [ ] All tests pass with `make test`
- [ ] No regression in existing tests
- [ ] Test code follows project conventions
- [ ] Focus on meaningful coverage, not just line count

## Notes

Priority 3 (Medium) - These commands already have partial coverage, so they're lower risk.
Estimated effort: 1-2 days

The goal is to fill coverage gaps and ensure comprehensive testing, not just hit coverage numbers.