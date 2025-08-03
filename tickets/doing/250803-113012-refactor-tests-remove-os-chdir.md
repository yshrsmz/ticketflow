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