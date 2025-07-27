---
priority: 2
description: "Fix TestWorktreeWorkflow test failure on macOS due to symlink path resolution"
created_at: "2025-07-28T00:10:24+09:00"
started_at: null
closed_at: null
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
- [ ] Investigate the path comparison issue in TestWorktreeWorkflow
- [ ] Fix the test to handle macOS symlink resolution properly
- [ ] Ensure the fix works on both macOS and Linux
- [ ] Run `make test-integration` to verify all tests pass
- [ ] Run `make vet`, `make fmt` and `make lint`

## Notes

This issue was discovered while working on ticket 250727-231907-fix-cleanup-force-flag. The test failure is not related to any recent changes but is a pre-existing issue with macOS-specific path handling.