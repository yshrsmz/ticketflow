# TicketForge Final Specification

## 基本方針

- **マージはGitHub PR経由**: `ticketforge`はマージ操作を行わない
- **明示的なGit操作**: 自動pushは行わない。ユーザーが明示的にpushする
- **チケット管理に集中**: ブランチ管理とチケット状態管理を分離

## 設定ファイル (.ticketforge.yaml)

```yaml
# Git設定
git:
  default_branch: "main"
  
# Worktree設定
worktree:
  enabled: true
  base_dir: "../.worktrees"  # 相対パスまたは絶対パス
  
  # worktree作成後の初期化コマンド
  init_commands:
    - "git fetch origin"
    # - "npm install"
    # - "make setup"
    
  # ネストしたworktreeの設定
  allow_nested: true
  nested_structure: "parent.sub"  # サブタスク用のディレクトリ構造
  max_nest_depth: 2
    
# チケット設定  
tickets:
  dir: "tickets"
  
  # 状態別のサブディレクトリ
  todo_dir: "todo"
  doing_dir: "doing"
  done_dir: "done"
  
  # チケットテンプレート
  template: |
    # 概要
    
    ## タスク
    - [ ] 
    
    ## メモ

# 出力設定
output:
  default_format: "text"  # text|json
  json_pretty: true
```

## コマンド一覧

### 基本コマンド

```bash
# TUI起動（引数なし）
ticketforge

# システム初期化
ticketforge init

# チケット操作
ticketforge new <slug>                    # チケット作成
ticketforge list [--status STATUS]        # 一覧表示
ticketforge show <ticket-id>              # 詳細表示
ticketforge start <ticket-id>             # 作業開始（worktree作成）
ticketforge close                         # 作業完了（マージなし）
ticketforge restore                       # current-ticketリンク復元

# Worktree管理
ticketforge worktree list                 # worktree一覧
ticketforge cleanup <ticket-id>           # PRマージ後のクリーンアップ

# その他
ticketforge status                        # 現在の状態表示
ticketforge help                          # ヘルプ表示
```

### オプション

```bash
# 出力形式
--format json                             # JSON出力（読み取り操作のみ）

# チケット作成時
ticketforge new <slug> --here             # 現在のブランチに作成（worktree内）

# Cleanup時
ticketforge cleanup <ticket-id> --force   # 確認をスキップ
```

## ワークフロー

### 1. 基本的なワークフロー

```bash
# 1. チケット作成
[main]$ ticketforge new implement-auth
Created ticket: tickets/todo/250124-150000-implement-auth.md

# 2. 作業開始
[main]$ ticketforge start 250124-150000-implement-auth
Creating worktree: ../.worktrees/250124-150000-implement-auth
Moving ticket: todo → doing
Committed: "Start ticket: 250124-150000-implement-auth"

Worktree created. Next steps:
1. cd ../.worktrees/250124-150000-implement-auth
2. Start your work
3. git push -u origin 250124-150000-implement-auth

# 3. Worktreeに移動して作業
[main]$ cd ../.worktrees/250124-150000-implement-auth
[implement-auth]$ # 実装作業...

# 4. 変更をコミット
[implement-auth]$ git add .
[implement-auth]$ git commit -m "Implement authentication feature"

# 5. チケット完了
[implement-auth]$ ticketforge close
Moving ticket: doing → done
Committed: "Close ticket: 250124-150000-implement-auth"

✅ Ticket closed: 250124-150000-implement-auth
   Branch: 250124-150000-implement-auth
   Status: doing → done

📋 Next steps:
1. Push your changes:
   git push origin 250124-150000-implement-auth
   
2. Create Pull Request on your Git service
   
3. After PR is merged:
   ticketforge cleanup 250124-150000-implement-auth

# 6. プッシュとPR作成
[implement-auth]$ git push origin 250124-150000-implement-auth
[implement-auth]$ gh pr create  # またはGitHub UIから

# 7. PRマージ後のクリーンアップ
[implement-auth]$ cd ../../project
[main]$ git pull
[main]$ ticketforge cleanup 250124-150000-implement-auth
🌳 Removing worktree: ../.worktrees/250124-150000-implement-auth
🌿 Deleting local branch: 250124-150000-implement-auth
✅ Cleanup completed
```

### 2. サブチケットのワークフロー

```bash
# 1. 親タスクで作業中
[user-system]$ pwd
/path/to/.worktrees/user-system

# 2. サブチケットを作成
[user-system]$ ticketforge new user-model
📍 Creating ticket in worktree branch: user-system
✅ Created ticket: tickets/todo/250124-153000-user-model.md

# 3. サブチケットを開始
[user-system]$ ticketforge start user-model
🌳 Creating nested worktree:
   Parent: ../.worktrees/user-system
   Child:  ../.worktrees/user-system.sub/user-model

# 4. サブタスクで作業
[user-system]$ cd ../user-system.sub/user-model
[user-model]$ # 実装...
[user-model]$ git add .
[user-model]$ git commit -m "Implement user model"

# 5. サブチケット完了
[user-model]$ ticketforge close
✅ Ticket closed: user-model

📋 Next steps:
1. Push your changes:
   git push origin user-model
   
2. Create Pull Request:
   Base: user-system  ← 親ブランチに向けてPR
   Compare: user-model

# 6. 親ブランチに向けてPR作成
[user-model]$ git push origin user-model
[user-model]$ gh pr create --base user-system
```

## Git操作の詳細

### startコマンドの動作

```go
func (m *Manager) StartTicket(ticketID string) error {
    // 1. Worktree作成（新しいブランチも同時に作成）
    git worktree add -b 250124-150000-implement-auth ../worktrees/250124-150000-implement-auth
    
    // 2. チケット移動（worktree内で実行）
    cd ../worktrees/250124-150000-implement-auth
    git mv tickets/todo/250124-150000-implement-auth.md tickets/doing/
    # started_atを更新
    git add tickets/
    git commit -m "Start ticket: 250124-150000-implement-auth"
    
    // 3. 初期化コマンド実行
    git fetch origin
    # その他の設定されたコマンド
    
    // 4. プッシュはユーザーに委ねる
    // 自動pushは行わない
}
```

### closeコマンドの動作

```go
func (m *Manager) CloseTicket() error {
    // 1. チケット移動（現在のworktreeで実行）
    git mv tickets/doing/250124-150000-implement-auth.md tickets/done/
    # closed_atを更新
    git add tickets/
    git commit -m "Close ticket: 250124-150000-implement-auth"
    
    // 2. マージは行わない
    // 3. プッシュは行わない
    // 4. worktreeも削除しない（PR作成のため）
    
    // 5. 次のステップを案内
    // - git push
    // - PR作成
    // - cleanup
}
```

## ディレクトリ構造

```
project/                              # メインリポジトリ
├── .ticketforge.yaml
├── tickets/
│   ├── todo/                        # 未開始
│   │   └── 250125-093000-add-tests.md
│   ├── doing/                       # 作業中
│   │   └── 250124-150000-implement-auth.md
│   └── done/                        # 完了
│       └── 250123-110000-setup-ci.md
├── current-ticket.md -> tickets/doing/250124-150000-implement-auth.md
└── src/

../.worktrees/                       # Worktreeディレクトリ
├── 250124-150000-implement-auth/    # アクティブなworktree
└── 250124-150000-implement-auth.sub/  # サブタスク用
    └── 250124-153000-user-model/
```

## エラーハンドリング

### よくあるエラーと対処

```bash
# mainブランチ以外でstart
[feature/other]$ ticketforge start some-ticket
Error: Must be on 'main' branch to start new ticket
Suggestions:
1. Switch to main: git checkout main
2. Or use --from-current-branch flag (not recommended)

# 未コミットの変更がある状態でclose
[feature]$ ticketforge close
Error: Uncommitted changes detected
Suggestions:
1. Commit your changes: git add . && git commit -m "message"
2. Or use --force flag to ignore (not recommended)

# Worktree内で新しいチケットをstart
[feature]$ ticketforge start another-ticket
Error: Cannot start new ticket from within a worktree
Suggestions:
1. Go back to main repository: cd ../../project
2. Or complete current work first: ticketforge close
```

## JSON出力形式

```bash
$ ticketforge list --format json
{
  "tickets": [
    {
      "id": "250124-150000-implement-auth",
      "path": "tickets/doing/250124-150000-implement-auth.md",
      "status": "doing",
      "priority": 1,
      "description": "User authentication implementation",
      "created_at": "2025-01-24T15:00:00Z",
      "started_at": "2025-01-24T15:30:00Z",
      "closed_at": null,
      "has_worktree": true,
      "worktree_path": "../.worktrees/250124-150000-implement-auth"
    }
  ],
  "summary": {
    "total": 5,
    "todo": 2,
    "doing": 1,
    "done": 2
  }
}
```

## 実装の優先順位

### Phase 1: コア機能
- 設定ファイル管理
- チケットモデル（todo/doing/done）
- 基本CLIコマンド（init, new, list, start, close）
- Worktreeなしでの動作

### Phase 2: Worktree統合
- Git操作の抽象化
- Worktree作成・削除
- cleanupコマンド
- restoreコマンド

### Phase 3: 高度な機能
- サブチケット（ネストしたworktree）
- JSON出力
- エラーハンドリングの洗練

### Phase 4: TUI
- Bubble Teaによる実装
- インタラクティブな操作

## まとめ

この仕様により：
1. **明示的な操作**: 自動push/mergeなし、ユーザーが完全にコントロール
2. **GitHub統合**: PR中心のワークフロー
3. **柔軟性**: サブチケットによる階層的なタスク管理
4. **シンプル**: チケット管理とGit操作の責務を分離