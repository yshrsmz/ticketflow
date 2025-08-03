---
priority: 4
description: "Add permission checking and concurrent operation handling"
created_at: "2025-08-03T12:15:21+09:00"
started_at: null
closed_at: null
related:
  - "parent:250726-183403-fix-branch-already-exist-on-start"
---

# Robustness Improvements

## Overview
Improve robustness by adding permission checking before operations and handling concurrent operations on the same ticket.

## Tasks
- [ ] Add permission checking for worktree directories
- [ ] Implement file-based locking for ticket operations
- [ ] Add timeout mechanism for lock acquisition
- [ ] Add --force flag to override stale locks
- [ ] Add graceful fallback when permissions fail
- [ ] Add tests for permission and concurrency scenarios

## Technical Details
- Check write permissions before creating directories
- Implement lock files in tickets directory (`.ticketID.lock`)
- Use file locks with PID and timestamp for stale detection
- Provide clear error messages with fix suggestions
- Fall back to non-worktree mode if permissions insufficient

## Acceptance Criteria
- Permission errors are caught early with helpful messages
- Concurrent operations are properly serialized
- Stale locks can be detected and cleaned
- System remains usable even with permission restrictions