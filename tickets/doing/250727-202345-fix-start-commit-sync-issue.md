---
priority: 1
description: start時の親ブランチコミットとworktreeの同期問題を解決する
created_at: 2025-07-27T20:23:45.005609+09:00
started_at: 2025-07-27T20:26:07.59654+09:00
closed_at: null
---

# 概要

`ticketflow start`コマンド実行時に、親ブランチ側でチケットファイルの移動（todo→doing）をコミットしているが、このコミットがworktree側に含まれないため、以下の問題が発生している：

1. **チケットの重複**: worktree側ではtodoディレクトリに、親ブランチ側ではdoingディレクトリに同じチケットが存在する状態になる
2. **履歴の齟齬**: 親ブランチの`started_at`更新コミットがworktreeに反映されない
3. **マージ時の競合リスク**: 後でworktreeをマージする際に、チケットファイルの移動に関する競合が発生する可能性

## 現在の動作

1. `ticketflow start <ticket-id>`実行
2. 親ブランチで:
   - チケットをtodo→doingに移動
   - `started_at`を更新
   - コミット作成
3. worktree作成（親ブランチのコミット前の状態から分岐）
4. 結果: worktreeにはtodoにチケットが残る

## 解決案

### 案1: worktree作成後にチケット移動をコミット
- worktreeを先に作成
- その後で親ブランチでチケット移動をコミット
- worktree側でも同じ変更を反映

### 案2: チケット移動のコミットを廃止
- start時はファイル移動のコミットを作らない
- worktree内でのみチケットステータスを管理
- cleanup時にマージと同時にステータス更新

### 案3: rebaseによる同期
- 現状のまま親ブランチでコミット
- worktree作成後、親ブランチの最新コミットをrebase

## タスク
- [ ] 現在のstart処理の実装を詳細に調査
- [ ] 各解決案のメリット・デメリットを整理
- [ ] 最適な解決方法を決定
- [ ] 実装方針を策定
- [ ] テストケースを設計
- [ ] 実装
- [ ] 既存のworktreeへの影響を考慮

## 技術仕様

### 影響範囲
- `internal/cli/start.go`: startコマンドの実装
- `internal/ticket/manager.go`: チケット移動処理
- `internal/git/worktree.go`: worktree作成処理

### 考慮事項
- 既存のワークフローとの互換性
- ユーザーの期待する動作
- Gitの履歴の整合性

## メモ

- この問題は優先度高（priority: 1）として設定
- worktreeベースの開発フローの根幹に関わる問題
- 解決方法によってはワークフロー全体の見直しが必要になる可能性
