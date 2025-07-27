---
priority: 2
description: ticketflow cleanup コマンドの --force フラグが機能しない問題を修正
created_at: "2025-07-27T23:19:07+09:00"
started_at: "2025-07-27T23:45:51+09:00"
closed_at: null
---

# 概要

`ticketflow cleanup <ticket-id> --force` コマンドを実行しても、確認プロンプトがスキップされずに表示される問題を修正する。

## 問題の詳細

現在の動作:
```bash
$ ticketflow cleanup 250727-193854-standardize-date-format --force
🗑️  Cleanup for ticket: 250727-193854-standardize-date-format
   Description: チケット内のcreated_at, started_at, closed_atの日付フォーマットを統一する

This will:
  • Remove worktree: /path/to/worktree
  • Delete local branch: 250727-193854-standardize-date-format

Are you sure? (y/N): 
❌ Cleanup cancelled
```

期待される動作:
- `--force` フラグが指定された場合、確認プロンプトをスキップして直接クリーンアップを実行する

## タスク
- [ ] CleanupTicket 関数で force フラグの処理を確認
- [ ] 確認プロンプトのロジックを修正
- [ ] テストケースを追加
- [ ] 動作確認

## 技術仕様

### 調査箇所
- `internal/cli/commands.go` の `CleanupTicket` メソッド
- `cmd/ticketflow/main.go` の `handleCleanupTicket` 関数
- フラグの受け渡しが正しく行われているか確認

### 修正方針
1. force フラグが true の場合、確認プロンプトをスキップする条件分岐を追加
2. 既存の確認ロジックを force フラグでラップする

## メモ

- この問題は日付フォーマット統一作業の後処理で発見された
- 手動で `git worktree remove` と `git branch -d` を実行することで回避可能