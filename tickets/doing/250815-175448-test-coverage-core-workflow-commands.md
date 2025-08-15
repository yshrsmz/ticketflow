---
priority: 1
description: Improve test coverage for core workflow commands (start and close)
created_at: "2025-08-15T17:54:48+09:00"
started_at: "2025-08-15T18:13:29+09:00"
closed_at: null
related:
    - parent:250815-171607-improve-command-test-coverage
---

# Test Coverage: Core Workflow Commands

## Overview

Improve test coverage for the essential workflow commands that are critical to the ticket management flow. These commands have low coverage and are frequently used by developers.

## Status Summary

✅ **COMPLETE** - Ready for developer review and approval

- **Coverage Achieved**: 
  - `close.go` Execute: **91.7%** (target was 70%, exceeded by 21.7%)
  - `start.go` Execute: **94.4%** (target was 70%, exceeded by 24.4%)
- **Approach Changed**: Shifted from mock-heavy unit testing to integration testing with real git operations
- **Code Quality**: Two golang-pro reviews completed, all issues resolved
- **Security**: Fixed directory traversal vulnerability and race conditions
- **Deliverables**: Created reusable test harness, comprehensive integration tests, updated documentation

## Current Coverage
- `close.go` Execute: 29.2%
- `start.go` Execute: 43.8%

## Target Coverage
- Achieve at least 70% coverage for both Execute methods
- Focus on critical user paths and error handling

## Tasks

### Testing Strategy Revision (Based on CLI Architecture Analysis)
- [x] Consulted golang-cli-architect for better testing patterns
- [x] Recognized that Execute methods are orchestrators, not unit-testable
- [x] Adopted integration testing approach following git/docker/kubectl patterns

### Test Harness Implementation
- [x] Created testharness package with complete test environment
- [x] Implemented git repo initialization and ticket management
- [x] Added helpers for worktree, commit, and file operations

### Integration Tests Implementation
- [x] Wrote comprehensive integration tests for close command
  - Test closing current ticket
  - Test closing by ID
  - Test with reason and force flags
  - Test error scenarios
  - Test JSON output format
  - Test context cancellation
- [x] Wrote comprehensive integration tests for start command
  - Test starting with worktree creation
  - Test force flag behavior
  - Test without worktree mode
  - Test parent relationships
  - Test error scenarios
  - Test init commands
  - Test JSON output format

### Verification
- [x] Run `make test` to ensure all tests pass
- [x] Run `make coverage` to verify coverage improvements
- [x] Run `make vet`, `make fmt` and `make lint`

### Documentation
- [x] Documented new testing philosophy in ticket
- [x] Update CLAUDE.md with testing strategy guidance
- [x] Updated parent and sibling tickets with new testing approach

### Code Review & Quality
- [x] Initial golang-pro code review completed
- [x] Fixed race condition in symlink creation
- [x] Added 30-second context timeouts to integration tests
- [x] Fixed directory traversal vulnerability in WriteFile
- [x] Improved string building efficiency
- [x] Enhanced error messages with more context
- [x] Final golang-pro review confirms production-ready code
- [ ] Get developer approval before closing

## Acceptance Criteria

- [x] `close.go` Execute method improves from 29.2% to ≥70% test coverage (achieved 91.7%)
- [x] `start.go` Execute method improves from 43.8% to ≥70% test coverage (achieved 94.4%)
- [x] All tests pass with `make test`
- [x] No regression in existing tests
- [x] Test code follows project conventions and uses table-driven tests where appropriate
- [x] Mock dependencies are properly isolated (replaced with integration tests)

## Testing Strategy Insights

After consulting with the golang-cli-architect, we discovered that our initial approach of heavy mocking was fundamentally flawed. Key insights:

1. **Execute methods are orchestrators** - They coordinate multiple components and are inherently integration-focused
2. **Wrong abstraction level** - Mocking at Manager/Git level creates brittle tests that don't verify real behavior
3. **Industry patterns** - Tools like git, docker, and kubectl use integration tests for commands, not unit tests with mocks
4. **Test harness approach** - Creating a real test environment with temp directories and actual git repos provides genuine confidence

### Implementation Changes

Instead of mock-heavy unit tests, we implemented:
- **testharness package**: Complete test environment with real git repo, file system, and config
- **Integration tests**: Test actual command execution with real dependencies
- **Focused scope**: Test user-visible behavior, not implementation details
- **Better coverage**: Integration tests provide meaningful coverage that verifies the tool actually works

This approach aligns with the "Don't fight the framework" principle - CLI commands are about orchestrating side effects, so integration testing is the natural fit.

### Additional Insights from Implementation

1. **Test Harness Reusability**: The `testharness` package we created is highly reusable and will significantly speed up testing for the remaining commands in sibling tickets.

2. **Security Considerations in Tests**: Even test code needs security considerations - we discovered and fixed a potential directory traversal vulnerability in the test harness.

3. **Context Timeouts Are Critical**: Integration tests that interact with external systems (git) must have timeouts to prevent hanging CI/CD pipelines.

4. **Coverage vs Quality**: We exceeded the 70% target (achieving 91.7% and 94.4%) but more importantly, the tests actually verify real behavior rather than just hitting code paths.

5. **Factory Pattern Benefits**: The `app_factory.go` pattern allows clean dependency injection without modifying production code structure, maintaining separation of concerns.

6. **Working Directory Management**: Integration tests that use `os.Chdir()` cannot run in parallel - this is an acceptable tradeoff for testing CLI tools that expect to run from project root.

7. **Error Message Quality**: Good error messages with context (like we added to close.go) dramatically improve debugging when tests fail.

## Notes

Priority 1 (Critical) - These are the most important commands in the workflow and need robust testing.
Estimated effort: 2 working days (aggressive target - may need adjustment)

### Approval Process
"Get developer approval before closing" means:
1. Complete all implementation and testing
2. Create PR with all changes
3. Get PR reviewed and approved by developer
4. Only close ticket after explicit approval is given