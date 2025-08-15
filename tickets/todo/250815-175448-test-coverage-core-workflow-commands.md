---
priority: 1
description: "Improve test coverage for core workflow commands (start and close)"
created_at: "2025-08-15T17:54:48+09:00"
started_at: null
closed_at: null
related:
    - "parent:250815-171607-improve-command-test-coverage"
---

# Test Coverage: Core Workflow Commands

## Overview

Improve test coverage for the essential workflow commands that are critical to the ticket management flow. These commands have low coverage and are frequently used by developers.

## Current Coverage
- `close.go` Execute: 29.2%
- `start.go` Execute: 43.8%

## Target Coverage
- Achieve at least 70% coverage for both Execute methods
- Focus on critical user paths and error handling

## Tasks

### Setup
- [ ] Create shared test utilities for mocking common dependencies
- [ ] Extract MockApp structure to a reusable test helper

### Close Command Tests
- [ ] Test successful ticket closure with valid ticket ID
- [ ] Test closure with commit message generation
- [ ] Test error handling for non-existent tickets
- [ ] Test error handling when ticket not in "doing" status
- [ ] Test both text and JSON output formats
- [ ] Test context cancellation handling
- [ ] Test git operations failure scenarios

### Start Command Tests  
- [ ] Test successful ticket start with valid ticket ID
- [ ] Test worktree creation and branch setup
- [ ] Test error handling for non-existent tickets
- [ ] Test error handling when ticket already started
- [ ] Test both text and JSON output formats
- [ ] Test context cancellation handling
- [ ] Test git operations failure scenarios
- [ ] Test init command execution

### Verification
- [ ] Run `make test` to ensure all tests pass
- [ ] Run `make coverage` to verify coverage improvements
- [ ] Run `make vet`, `make fmt` and `make lint`

### Documentation
- [ ] Document any legitimately untestable code paths
- [ ] Add comments explaining complex test scenarios
- [ ] Update the ticket with insights from implementation
- [ ] Get developer approval before closing

## Acceptance Criteria

- [ ] `close.go` Execute method has ≥70% test coverage
- [ ] `start.go` Execute method has ≥70% test coverage
- [ ] All tests pass with `make test`
- [ ] No regression in existing tests
- [ ] Test code follows project conventions and uses table-driven tests where appropriate
- [ ] Mock dependencies are properly isolated

## Notes

Priority 1 (Critical) - These are the most important commands in the workflow and need robust testing.
Estimated effort: 2 days