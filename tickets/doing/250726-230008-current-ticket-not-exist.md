---
priority: 1
description: current-ticket.md does not exist in worktree
created_at: 2025-07-26T23:00:08.843045+09:00
started_at: 2025-07-27T02:36:27.257840763Z
closed_at: null
---

# 概要

When worktrees are enabled, the `current-ticket.md` symlink is created in the main repository but not in the worktree itself. This causes issues when code tries to access `current-ticket.md` from within a worktree context.

## Root Cause

1. `StartTicket` creates a worktree for the ticket branch
2. `SetCurrentTicket` creates the `current-ticket.md` symlink in the main repository (via `m.projectRoot`)
3. When working in the worktree, any code that tries to access `current-ticket.md` fails because the symlink only exists in the main repo

## タスク
- [x] Determine if `current-ticket.md` should exist in worktrees - YES
- [ ] Copy the ticket file from source branch to worktree (to preserve uncommitted changes)
- [ ] Create `current-ticket.md` symlink in worktree after creating the worktree
- [ ] Update `StartTicket` in `internal/cli/commands.go` to implement both changes

## 技術仕様

The issue occurs in the interaction between:
- `internal/cli/commands.go`: `StartTicket()` - creates worktree
- `internal/ticket/manager.go`: `SetCurrentTicket()` - creates symlink in main repo only
- Any code that tries to access `current-ticket.md` from worktree context

### Solution Implementation
After creating the worktree in `StartTicket`, we need to:

1. **Copy the ticket file from source to worktree**
   - After moving ticket to `doing/` in source branch
   - Before committing in source branch
   - Copy the file to the same path in the worktree to preserve any uncommitted edits
   - This ensures user's work on the ticket description is not lost

2. **Create `current-ticket.md` symlink in the worktree**
   - The symlink should point to the ticket file within that worktree's filesystem
   - Example: `./tickets/doing/250726-123456-my-feature.md`

Example flow:
- Source branch moves ticket: `tickets/todo/X.md` → `tickets/doing/X.md`
- Copy to worktree: `cp tickets/doing/X.md ../250726-123456-my-feature/tickets/doing/X.md`
- Create symlink in worktree: `ln -s tickets/doing/X.md ../250726-123456-my-feature/current-ticket.md`

Code location to modify: `internal/cli/commands.go` after:
1. Ticket move to doing/ in source branch
2. Before git commit in source branch
3. After worktree creation

## メモ

This is likely affecting the `ticketflow close` command and other commands when run from within a worktree.
