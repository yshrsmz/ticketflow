---
priority: 2
description: Fix TestWorktreeWorkflow test failure on macOS due to symlink path resolution
created_at: "2025-07-28T00:16:06+09:00"
started_at: "2025-07-28T13:21:29+09:00"
closed_at: null
related:
    - parent:250727-231907-fix-cleanup-force-flag
---

# Ticket Overview

The `TestWorktreeWorkflow` integration test is failing on macOS because of path resolution differences. macOS has `/var` as a symlink to `/private/var`, which causes the test to fail when comparing expected vs actual paths.

## Error Details

```
Error Trace: test/integration/worktree_test.go:82
Error:      Not equal: 
            expected: "/var/folders/..."
            actual  : "/private/var/folders/..."
```

## Tasks
- [x] Investigate the path comparison issue in TestWorktreeWorkflow
- [x] Fix the test to handle macOS symlink resolution properly
- [x] Ensure the fix works on both macOS and Linux
- [x] Run `make test-integration` to verify all tests pass
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Set up GitHub Actions to run tests on PRs

## Solution

Fixed the test by using `filepath.EvalSymlinks()` to resolve symlinks before comparing paths. This handles the macOS `/var` → `/private/var` symlink issue while maintaining compatibility with Linux.

## Changes Made

1. **Fixed TestWorktreeWorkflow** (test/integration/worktree_test.go:82-87)
   - Added symlink resolution using `filepath.EvalSymlinks()` for both expected and actual paths
   - This ensures consistent path comparison across macOS and Linux

2. **Added GitHub Actions workflow** (.github/workflows/test.yml)
   - Tests on both Ubuntu and macOS
   - Tests with Go 1.22 and 1.23
   - Runs unit and integration tests
   - Checks code formatting and runs go vet
   - Includes golangci-lint for comprehensive linting

## Notes

This issue was discovered while working on ticket 250727-231907-fix-cleanup-force-flag. The test failure is not related to any recent changes but is a pre-existing issue with macOS-specific path handling.
