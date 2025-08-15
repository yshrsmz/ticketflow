---
priority: 2
description: "Add test coverage for commands with 0% Execute method coverage"
created_at: "2025-08-15T17:55:48+09:00"
started_at: null
closed_at: null
related:
    - "parent:250815-171607-improve-command-test-coverage"
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

## Tasks

### Setup
- [ ] Review existing test patterns from well-tested commands (list, version, help)
- [ ] Reuse test utilities created in the core workflow ticket if available

### New Command Tests
- [ ] Test successful ticket creation with valid slug
- [ ] Test creation with parent ticket flag
- [ ] Test error handling for invalid slug format
- [ ] Test error handling for duplicate tickets
- [ ] Test both text and JSON output formats
- [ ] Test template application

### Restore Command Tests
- [ ] Test successful ticket restoration from done to todo
- [ ] Test error handling for non-existent tickets
- [ ] Test error handling for tickets not in done status
- [ ] Test both text and JSON output formats
- [ ] Test file operations and state transitions

### Show Command Tests
- [ ] Test successful display of existing ticket
- [ ] Test error handling for non-existent tickets
- [ ] Test both text and JSON output formats
- [ ] Test markdown rendering
- [ ] Test metadata display

### Worktree Clean Command Tests
- [ ] Test successful cleanup of merged worktrees
- [ ] Test handling of unmerged branches
- [ ] Test error handling for missing worktrees
- [ ] Test both text and JSON output formats
- [ ] Test dry-run mode if available

### Worktree List Command Tests
- [ ] Test listing all worktrees
- [ ] Test filtering by status
- [ ] Test error handling when no worktrees exist
- [ ] Test both text and JSON output formats
- [ ] Test sorting and formatting

### Verification
- [ ] Run `make test` to ensure all tests pass
- [ ] Run `make coverage` to verify coverage improvements
- [ ] Run `make vet`, `make fmt` and `make lint`

### Documentation
- [ ] Document any legitimately untestable code paths
- [ ] Add comments for complex test setups
- [ ] Update the ticket with insights from implementation
- [ ] Get developer approval before closing

## Acceptance Criteria

- [ ] All 5 commands have ≥70% Execute method test coverage
- [ ] All tests pass with `make test`
- [ ] No regression in existing tests
- [ ] Test code follows project conventions
- [ ] Uses table-driven tests for multiple scenarios
- [ ] Mock dependencies are properly isolated

## Notes

Priority 2 (High) - These commands have zero coverage and need immediate attention.
Estimated effort: 2-3 days