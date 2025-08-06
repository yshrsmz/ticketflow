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

## Tasks
- [ ] Add automatic recovery with `git worktree prune` on worktree errors
- [ ] Implement retry mechanism after pruning
- [ ] Add `ticketflow doctor --fix-worktrees` command
- [ ] Add detection for orphaned worktree directories
- [ ] Add tests for recovery scenarios

## Technical Details
- Detect worktree-related errors in git command output
- Run `git worktree prune` automatically when appropriate
- Implement `doctor` subcommand with various fix options
- Check for worktree directories that exist without git tracking

## Acceptance Criteria
- Worktree errors are automatically recovered when possible
- Manual recovery command works for complex cases
- Clear messages guide users through recovery
- No data loss during recovery operations