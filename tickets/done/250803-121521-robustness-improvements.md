---
priority: 4
description: Add permission checking and concurrent operation handling
created_at: "2025-08-03T12:15:21+09:00"
started_at: "2025-08-22T17:46:50+09:00"
closed_at: "2025-08-22T18:05:27+09:00"
related:
    - parent:250726-183403-fix-branch-already-exist-on-start
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

## Analysis and Decision (2025-08-22)

After thorough analysis of the codebase and ticketflow's actual usage patterns, we've decided to **NOT IMPLEMENT** these features based on the YAGNI (You Aren't Gonna Need It) principle.

### Why These Features Are Unnecessary

#### 1. **Permission Checking** ❌
- Ticketflow is a local, single-user tool where users have full control over their directories
- Current behavior already fails with clear errors if directories can't be created
- Pre-emptive permission checks add complexity without solving real problems
- In practice, permission errors are extremely rare (only occurs with disk full or wrong directory)

#### 2. **File-based Locking** ❌
- Single user, single machine usage pattern makes concurrent access a non-issue
- Git itself doesn't use file locks for most operations and works perfectly
- Would require multiple deliberate simultaneous `ticketflow start` commands to cause issues
- Adds significant complexity: lock files, timeouts, cleanup, stale detection = potential bugs

#### 3. **Timeout Mechanism** ❌
- Only needed if we implement locking (which we shouldn't)
- Adds complexity for a problem that doesn't exist

#### 4. **Force Flag for Locks** ❌
- The `--force` flag already exists for worktree recreation (the actual useful case)
- Lock override is only needed if we add unnecessary locking

#### 5. **Fallback Mechanisms** ❌
- Silently degrading to non-worktree mode when permissions fail is confusing
- Better approach: Fail fast with clear error messages
- Users should fix root causes, not have the tool hide problems

### Evidence from Real Usage

Looking at 50+ completed tickets, we've addressed:
- Branch divergence handling ✅ (real git issue)
- Worktree sync problems ✅ (real git issue)
- UI improvements ✅ (real usability issue)
- Performance optimization ✅ (real issue with 100+ tickets)
- Command architecture ✅ (real maintainability issue)

**Zero** permission issues. **Zero** concurrency problems. **Zero** lock corruption incidents.

### The Cost of Over-Engineering

Adding these features would:
- Increase codebase complexity permanently
- Create new failure modes (what if lock mechanism breaks?)
- Make the tool harder to understand for contributors
- Require ongoing maintenance for features that solve theoretical problems
- Violate ticketflow's design principle of simplicity

### Better Alternatives

If robustness concerns arise in actual usage:
1. **Simple safety check** (if really needed):
   ```go
   // One-liner check before critical operations
   if _, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600); err != nil {
       return fmt.Errorf("cannot create worktree: %w", err)
   }
   ```

2. **Documentation**: Add troubleshooting guide for edge cases

3. **Focus on real features** that improve daily usage:
   - Plugin system for custom workflows
   - Better TUI features
   - GitHub/GitLab API integration
   - Ticket templates
   - Time tracking

### Conclusion

This ticket represents defensive programming against problems that don't exist in practice. Ticketflow's current "fail fast with clear errors" approach is the correct design for a simple, local development tool. Adding these features would violate YAGNI and make the codebase unnecessarily complex without providing real value to users.

**Decision**: Won't implement - unnecessary complexity for a local-only tool.