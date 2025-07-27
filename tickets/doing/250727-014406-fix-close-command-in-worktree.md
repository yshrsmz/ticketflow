---
priority: 2
description: ""
created_at: 2025-07-27T01:44:06.585250076Z
started_at: 2025-07-27T02:56:46.835865674Z
closed_at: null
---

# 概要

The `ticketflow close` command and other commands fail when run from within a worktree because they cannot find `current-ticket.md`. This is a consequence of the issue described in ticket 250726-230008-current-ticket-not-exist.md.

## タスク
- [ ] Implement the fix from ticket 250726-230008 (copy ticket file and create symlink in worktree)
- [ ] Test `ticketflow close` command from within a worktree
- [ ] Test other commands that rely on `current-ticket.md` from worktree context
- [ ] Ensure commands work correctly in both worktree and non-worktree modes

## 技術仕様

Commands affected:
- `ticketflow close` - relies on `GetCurrentTicket()` which looks for `current-ticket.md`
- Any other commands that use `Manager.GetCurrentTicket()`

The fix involves modifying `StartTicket` in `internal/cli/commands.go` to:
1. Copy the ticket file from source branch to worktree
2. Create `current-ticket.md` symlink in the worktree

## メモ

This ticket depends on the analysis done in ticket 250726-230008-current-ticket-not-exist.md