---
priority: 2
description: Fix ticketflow close command when run from worktree - current-ticket.md not found
created_at: 2025-07-27T01:44:06.585250076Z
started_at: 2025-07-27T11:07:51.537269441Z
closed_at: null
---

# 概要

The `ticketflow close` command and other commands fail when run from within a worktree because they cannot find `current-ticket.md`. This issue has been analyzed and a fix has already been implemented.

## 現状分析

調査の結果、`StartTicket` 関数（`internal/cli/commands.go:384-407`）に既に修正が実装されていることを確認しました：

1. **ワークツリー作成時の処理**：
   - チケットファイルをワークツリーの `tickets/doing/` ディレクトリにコピー
   - ワークツリー内に `current-ticket.md` シンボリックリンクを作成

2. **実装済みのコード**：
   ```go
   // If using worktree, copy ticket file and create symlink
   if app.Config.Worktree.Enabled && worktreePath != "" {
       // Copy ticket file to worktree
       // Create current-ticket.md symlink in worktree
   }
   ```

## 問題の原因

`GetCurrentTicket()` は `projectRoot` から `current-ticket.md` を探すため、ワークツリー内から実行すると見つけられない問題がありましたが、既に対処済みです。

## タスク
- [x] Implement the fix from ticket 250726-230008 (copy ticket file and create symlink in worktree) - 実装済み
- [ ] Test `ticketflow close` command from within a worktree
- [ ] Test other commands that rely on `current-ticket.md` from worktree context
- [ ] Ensure commands work correctly in both worktree and non-worktree modes
- [ ] 既存の実装が正しく動作することを確認

## 技術仕様

Commands affected:
- `ticketflow close` - relies on `GetCurrentTicket()` which looks for `current-ticket.md`
- Any other commands that use `Manager.GetCurrentTicket()`

既に実装されている修正内容（`internal/cli/commands.go`）：
1. ワークツリー作成時にチケットファイルをコピー（line 393-400）
2. ワークツリー内に `current-ticket.md` シンボリックリンクを作成（line 403-406）

## 次のステップ

1. 実装済みの修正が正しく動作するかテストを実施
2. 問題が解決されていない場合は、追加の修正を検討
3. テストケースの追加を検討

## メモ

This ticket is related to ticket 250726-230008-current-ticket-not-exist.md. The implementation has been completed but needs testing to ensure it works correctly.
