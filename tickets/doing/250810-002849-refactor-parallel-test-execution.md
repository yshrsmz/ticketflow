---
priority: 2
description: Complete parallel test execution for remaining unit tests
created_at: "2025-08-10T00:28:49+09:00"
started_at: "2025-08-11T23:58:03+09:00"
closed_at: null
---

# Task 1.4: Complete Parallel Test Execution for Unit Tests

**Duration**: 2-3 hours  
**Complexity**: Trivial  
**Phase**: 1 - Foundation  
**Dependencies**: None

Add `t.Parallel()` to remaining unit tests that don't have it yet. Integration tests were already parallelized in ticket 250803-113012.

## Current State
- **Integration tests**: ✅ Already parallelized (refactored in ticket 250803-113012)  
- **Unit tests**: ~34% already use `t.Parallel()`, ~66% still sequential
- **Current performance**: Tests run in 5.7s (down from 11.4s sequential)
- **Documentation**: Already exists in `docs/testing-patterns.md`

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [ ] Identify unit tests without `t.Parallel()` (skip benchmark tests)
- [ ] Add `t.Parallel()` to remaining unit test functions in `internal/` packages
- [ ] Verify all tests pass with race detection (`make test`)
- [ ] Measure final performance improvement
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Get developer approval before closing

## Implementation Notes

- **Important**: Integration tests are already parallelized (they no longer use `os.Chdir`)
- Focus only on unit tests in `internal/` packages that don't have `t.Parallel()`
- Skip benchmark tests (they shouldn't be parallel for accurate measurements)
- All tests already use `t.TempDir()` for isolation, so they should be safe to parallelize
- The Makefile already includes `-race` flag for test execution
- Key files to update:
  - `internal/ticket/manager_test.go` (most test functions)
  - `internal/cli/commands_helpers_test.go`
  - `internal/cli/errors_test.go`
  - `internal/cli/output_test.go`
  - `internal/git/git_test.go`
  - `internal/git/worktree_test.go`

## Expected Outcomes

- Additional 20-30% test suite speedup (from current 5.7s to ~4s)
- Consistent parallelization across all non-benchmark tests
- Better CPU utilization during test runs

## Related Documentation:
- Full refactoring discussion: docs/20250810-refactor-discussion.md
- Executive summary: docs/20250810-refactor-summary.md
- Ticket overview: docs/20250810-refactor-tickets.md
