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

## Decision: Won't Implement

After thorough analysis and consideration of ticketflow's core purpose as a **ticket management tool**, we've decided NOT to implement complex worktree recovery mechanisms.

### Rationale

1. **Scope creep**: ticketflow is a ticket management tool, not a git repair utility. Implementing recovery mechanisms would blur this focus.

2. **Problem rarity**: Worktree corruption is extremely rare in practice. Most developers never encounter it.

3. **Existing solution**: Git already provides `git worktree prune` which solves 99% of worktree corruption issues. Users can run this command directly.

4. **User capability**: Developers using ticketflow are already comfortable with git and can handle basic troubleshooting.

5. **Cost vs benefit**: The implementation would require 4 complex sub-tickets worth of development for a problem that rarely occurs. The maintenance burden would be ongoing.

6. **Over-engineering**: This would be a classic case of building a complex solution for a non-problem.

### What We'll Do Instead

1. **Existing functionality**: Keep the current `PruneWorktrees()` call in cleanup operations (already implemented). ✅

2. **Better error messages**: When worktree errors occur, provide clear, actionable messages:
   ```
   Error: Worktree appears to be corrupted
   Fix: Run 'git worktree prune' to clean up corrupted references
   Then retry your ticketflow command
   ```
   → **Ticket created:** 250806-172829-improve-worktree-error-messages

3. **Documentation**: Add a troubleshooting section to the docs covering common worktree issues and their solutions.
   → **Ticket created:** 250806-172904-add-troubleshooting-docs

### Original Tasks (Not Implementing)
- ~~Add automatic recovery with `git worktree prune` on worktree errors~~
- ~~Implement retry mechanism after pruning~~
- ~~Add `ticketflow doctor --fix-worktrees` command~~
- ~~Add detection for orphaned worktree directories~~
- ~~Add tests for recovery scenarios~~

## Resolution

This ticket is being closed as "Won't Implement" based on the principle that ticketflow should focus on its core purpose: ticket management. Git worktree issues should be handled by git itself, with ticketflow providing helpful error messages to guide users to the solution.

The 4 sub-tickets that were initially created have been deleted as they are no longer needed.