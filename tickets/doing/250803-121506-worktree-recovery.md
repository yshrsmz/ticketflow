---
priority: 3
description: Add worktree recovery mechanisms for corrupted references
created_at: "2025-08-03T12:15:06+09:00"
started_at: "2025-08-06T14:35:55+09:00"
closed_at: null
related:
    - parent:250726-183403-fix-branch-already-exist-on-start
---

# Worktree Recovery

## Overview
Handle corrupted worktree references where the worktree directory is deleted but git still tracks it, or .git/worktrees entries are corrupted.

## Analysis Complete - Ticket Split Required

After thorough analysis of the codebase, this ticket has been determined to be too large for a single implementation. It has been split into 4 focused sub-tickets that should be implemented in sequence:

### Sub-tickets Created:

1. **250806-171131-worktree-error-detection** (Priority: 1)
   - Enhance git error detection to identify worktree-specific corruption patterns
   - Create foundation for recovery mechanisms
   - **Must be completed first**

2. **250806-171235-automatic-worktree-recovery** (Priority: 1)
   - Implement automatic recovery with `git worktree prune`
   - Add retry logic with exponential backoff
   - Depends on: worktree-error-detection

3. **250806-171306-doctor-command** (Priority: 2)
   - Implement `ticketflow doctor` command
   - Add `--fix-worktrees` flag for manual recovery
   - Provide diagnostic capabilities
   - Depends on: worktree-error-detection

4. **250806-171343-enhanced-recovery-features** (Priority: 3)
   - Advanced recovery features for complex scenarios
   - Metadata backup/restore system
   - Recovery journal and statistics
   - Depends on: all previous tickets

## Original Tasks (Now Distributed)
- [→ Ticket 1] Add automatic recovery with `git worktree prune` on worktree errors
- [→ Ticket 2] Implement retry mechanism after pruning
- [→ Ticket 3] Add `ticketflow doctor --fix-worktrees` command
- [→ Ticket 3] Add detection for orphaned worktree directories
- [→ All] Add tests for recovery scenarios

## Technical Analysis Summary
The analysis revealed:
- ✅ `PruneWorktrees()` already exists in `internal/git/worktree.go`
- ✅ Error infrastructure exists but needs enhancement for worktree-specific errors
- ❌ No `doctor` command infrastructure currently exists
- ❌ No retry mechanism infrastructure
- ❌ Current `Git.Exec()` doesn't identify worktree corruption specifically

## Recommendation
**This ticket should be closed** after creating the sub-tickets. The implementation should proceed with the sub-tickets in order, starting with the error detection infrastructure as it provides the foundation for all recovery mechanisms.

## Acceptance Criteria (Met by Sub-tickets)
- ✅ Worktree errors are automatically recovered when possible (Ticket 2)
- ✅ Manual recovery command works for complex cases (Ticket 3)
- ✅ Clear messages guide users through recovery (All tickets)
- ✅ No data loss during recovery operations (All tickets)