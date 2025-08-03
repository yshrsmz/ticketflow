---
priority: 2
description: Refactor tests to avoid using os.Chdir for better parallelization
created_at: "2025-08-03T11:30:12+09:00"
started_at: "2025-08-03T14:09:31+09:00"
closed_at: null
related:
    - parent:250801-003207-improve-test-coverage
---

# Refactor tests to remove os.Chdir usage

Refactor test files that use `os.Chdir` to improve test parallelization. Tests that change the working directory cannot run in parallel because they modify global state.

## Context

Several test files use `os.Chdir` to change the working directory during test execution:
- `cmd/ticketflow/handlers_test.go`
- `internal/cli/cleanup_test.go`
- Various integration tests in `test/integration/`

This prevents these tests from running in parallel with `t.Parallel()`, which could significantly speed up test execution.

## Tasks

- [ ] Identify all test files using `os.Chdir`
- [ ] Refactor tests to use absolute paths instead of changing directories
- [ ] Use `cmd.Dir` field when executing commands instead of changing global directory
- [ ] Add `t.Parallel()` to tests that can now run concurrently
- [ ] Ensure all tests still pass after refactoring
- [ ] Run `make test` to verify all tests pass
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update CLAUDE.md with best practices for avoiding os.Chdir in tests
- [ ] Get developer approval before closing

## Implementation Strategy

Instead of:
```go
os.Chdir(testDir)
cmd := exec.Command("git", "init")
cmd.Run()
```

Use:
```go
cmd := exec.Command("git", "init")
cmd.Dir = testDir
cmd.Run()
```

## Notes

This is a follow-up from PR #31 code review suggestions. The goal is to improve test performance by enabling parallel execution where possible.

## Update: Detailed Design for Proper Solution

After initial investigation, it was discovered that simply replacing `os.Chdir` with `cmd.Dir` is insufficient because the ticketflow application itself expects to run from the project root directory. A more comprehensive solution is needed.

### Problem Analysis

The ticketflow application relies on being executed from the project root directory (similar to git) and uses relative paths to:
1. Find the git repository root using `git rev-parse --show-toplevel`
2. Load the `.ticketflow.yaml` configuration file
3. Access ticket directories (todo/doing/done)
4. Manage git worktrees

This design prevents tests from running in parallel since `os.Chdir` modifies global state.

### Solution Approaches

#### Approach 1: Modify ticketflow to Accept Working Directory Parameter (Recommended)

Add a working directory parameter throughout the application, similar to git's `-C` flag.

**Implementation Details:**
1. Update Git package to use `cmd.Dir = repoPath` for all commands
2. Add `-C` flag to CLI commands
3. Create `WithWorkingDirectory` option for CLI app creation
4. Update all file operations to use absolute paths

**Example API Changes:**
```go
// CLI usage
ticketflow -C /path/to/repo new feature-x

// Test usage
app, err := cli.NewAppWithWorkingDir(ctx, repoPath)
```

**Benefits:**
- Enables full test parallelization
- Clean, explicit API design
- No process overhead
- Follows git's pattern

**Effort:** 2-3 days

#### Approach 2: Create Test Harness with Process Isolation

Run each test in a separate process with its own working directory.

**Implementation:**
- Build test binary for each test run
- Execute commands via subprocess
- Parse JSON output for assertions

**Benefits:**
- No production code changes
- Complete test isolation

**Drawbacks:**
- Significant performance overhead
- Complex debugging
- Cannot test internal APIs

**Effort:** 1-2 days

#### Approach 3: Accept Sequential Execution

Organize tests into parallel and sequential categories.

**Implementation:**
- Move tests that don't need `os.Chdir` to `test/integration/parallel/`
- Keep `os.Chdir` tests in `test/integration/sequential/`
- Update Makefile to run each category appropriately

**Benefits:**
- No code changes required
- Immediate implementation

**Drawbacks:**
- Limited parallelization
- Doesn't solve fundamental issue

**Effort:** 0.5-1 day

### Recommendation

**Approach 1** is recommended because it:
- Provides the cleanest architecture
- Enables full test parallelization
- Follows established patterns (git's `-C` flag)
- Improves overall codebase quality

### Implementation Plan for Approach 1 (Simplified)

**Key Discovery**: `FindProjectRoot` already accepts a `startPath` parameter. We only need to make the CLI layer configurable!

1. **Phase 1: Add Working Directory Option to CLI** (Day 1)
   - Add `workingDir` field to `App` struct (defaults to ".")
   - Create `WithWorkingDirectory` option function
   - Update `NewApp` to use `workingDir` instead of hardcoded "."
   - Update all calls from `FindProjectRoot(ctx, ".")` to `FindProjectRoot(ctx, app.workingDir)`

2. **Phase 2: Update Component Initialization** (Day 2)
   - Ensure `git.New()` uses the working directory path
   - Update `ticket.NewManager()` to use absolute paths
   - Update config loading to work relative to working directory
   - Verify all file operations use absolute paths

3. **Phase 3: Test Migration** (Day 3)
   - Create `NewAppWithWorkingDir` helper for tests
   - Remove all `os.Chdir` calls from tests
   - Add `t.Parallel()` to all tests
   - Document the new testing pattern

### Important Clarifications

- **No -C flag needed for daily use** - this is primarily for internal testing
- **No changes to user experience** - defaults to current directory
- **FindProjectRoot remains unchanged** - already has the needed functionality
- **Minimal code changes** - mostly threading the working directory through

### Expected Performance Impact

With Approach 1:
- Integration tests can run with full parallelization
- Expected 3-4x speedup on multi-core machines
- No runtime performance impact
- Cleaner test execution without global state changes