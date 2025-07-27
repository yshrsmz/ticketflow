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
- [x] CleanupTicket 関数で force フラグの処理を確認
- [x] 確認プロンプトのロジックを修正
- [x] テストケースを追加
- [x] 動作確認

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

## 調査結果と修正内容

### 根本原因
Go の flag パーサーの仕様により、フラグは位置引数（ticket ID）の前に指定する必要があった。
`ticketflow cleanup <ticket-id> --force` の順序では、`--force` がフラグとして認識されない。

### 実装した修正
1. **確認プロンプトのロジック修正** (`internal/cli/commands.go`):
   - クリーンアップ情報の表示を force フラグのチェック外に移動
   - force が false の場合のみ確認プロンプトを表示するように修正

2. **ヘルプテキストの更新** (`cmd/ticketflow/main.go`):
   - 使用方法を `ticketflow cleanup [options] <ticket>` に更新
   - 正しい使用例を追加: `ticketflow cleanup --force 250124-150000-implement-auth`

3. **統合テストの追加** (`test/integration/cleanup_test.go`):
   - force フラグが正しく動作することを確認するテストケースを追加
   - テストは成功し、修正が正しく機能することを確認

### 正しい使用方法
```bash
# 正しい - フラグを ticket ID の前に指定
ticketflow cleanup --force <ticket-id>

# 間違い - これは動作しない
ticketflow cleanup <ticket-id> --force
```

### その他の発見
- 統合テストで `TestWorktreeWorkflow` が macOS のパス解決（`/var` vs `/private/var`）の違いにより失敗することを発見
- この問題は本修正とは無関係のため、別チケット（250728-001024-fix-worktree-test-macos）として記録