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
   - Tests on Ubuntu only (cost-effective)
   - Tests with Go 1.24
   - Runs unit and integration tests
   - Checks code formatting and runs go vet
   - Includes golangci-lint for comprehensive linting
   - All actions updated to latest versions (checkout@v4, setup-go@v5, cache@v4, golangci-lint-action@v8)

3. **Replaced Codecov with custom PR comment coverage**
   - No external dependencies or tokens required
   - Posts nicely formatted coverage report as PR comment
   - Shows total coverage with color-coded emoji (🟢 80%+, 🟡 60-79%, 🔴 <60%)
   - Collapsible detailed coverage by package
   - Auto-updates existing comment on new pushes

## Notes

This issue was discovered while working on ticket 250727-231907-fix-cleanup-force-flag. The test failure is not related to any recent changes but is a pre-existing issue with macOS-specific path handling.

## Insights and Lessons Learned

1. **macOS Symlink Behavior**: macOS uses symlinks for system directories (e.g., `/var` → `/private/var`, `/tmp` → `/private/tmp`). When writing cross-platform tests that compare file paths, always use `filepath.EvalSymlinks()` to resolve symlinks before comparison to ensure consistency across different operating systems.

2. **GitHub Actions Evolution**: The ecosystem has moved significantly in 2024:
   - All major actions have updated to v4+ to support Node.js 20
   - Codecov now requires tokens even for public repos (v4)
   - GitHub's cache action will undergo architecture changes in Feb 2025
   - Always check for the latest versions when setting up workflows

3. **Cost-Effective CI**: Running tests only on Ubuntu saves significant costs compared to macOS runners, which are 10x more expensive. Since the symlink fix ensures cross-platform compatibility, testing on one platform is sufficient for most cases.

4. **Simple Solutions Often Best**: Instead of using complex third-party coverage services, a simple 30-line GitHub script can provide elegant PR comment coverage reports. This approach:
   - Eliminates external dependencies
   - Keeps data within GitHub
   - Provides immediate feedback in PR context
   - Is easier to maintain and customize

5. **AWK in GitHub Actions**: When using AWK in GitHub Actions YAML, dollar signs must be escaped as `$$` to prevent YAML interpretation. This is a common gotcha when writing shell scripts in workflow files.

6. **Testing Coverage Tools**: Always test coverage generation commands locally before adding to CI. The `go tool cover` output format can vary, and parsing it requires careful attention to edge cases.
