# TicketFlow 包括的設計書

## 1. システム概要

### 1.1 概要

TicketFlowは、生成AIとの協働に最適化されたgit worktreeベースのチケット管理システムです。ticket.shの設計思想を継承しつつ、worktree統合による並行作業のサポートとモダンなUIを提供します。

### 1.2 主な特徴

- **Markdown + YAML frontmatter**によるチケット管理
- **Git worktree**による並行作業のサポート
- **ディレクトリベースの状態管理**（todo/doing/done）
- **人間向けTUI**とAI向けCLIの両立
- **PRベースのワークフロー**（自動merge/pushなし）
- **サブチケット機能**による階層的タスク管理

### 1.3 設計原則

1. **シンプル**: 外部サービスへの依存なし、ローカルで完結
2. **明示的**: 自動的なGit操作は最小限、ユーザーが制御
3. **透明性**: チケットの状態がディレクトリ構造で可視化
4. **柔軟性**: AI/人間どちらも使いやすいインターフェース

## 2. システム構成

### 2.1 技術スタック

- **言語**: Go
- **TUIフレームワーク**: Bubble Tea (github.com/charmbracelet/bubbletea)
- **設定ファイル**: YAML (gopkg.in/yaml.v3)
- **テスト**: testify (github.com/stretchr/testify)

### 2.2 ディレクトリ構造

#### プロジェクト構造
```
ticketflow/                          # ツールのソースコード
├── cmd/
│   └── ticketflow/
│       └── main.go                  # エントリーポイント
├── internal/
│   ├── config/                      # 設定管理
│   │   ├── config.go
│   │   └── config_test.go
│   ├── ticket/                      # チケット管理
│   │   ├── ticket.go               # チケットモデル
│   │   ├── manager.go              # チケット操作
│   │   └── status.go               # 状態管理
│   ├── git/                        # Git操作
│   │   ├── git.go                  # Git基本操作
│   │   └── worktree.go             # Worktree管理
│   ├── cli/                        # CLIインターフェース
│   │   ├── app.go                  # アプリケーション
│   │   ├── commands.go             # コマンド実装
│   │   ├── output.go               # 出力フォーマット
│   │   └── errors.go               # エラーハンドリング
│   └── ui/                         # TUIインターフェース
│       ├── app.go                  # TUIアプリケーション
│       ├── views/                  # 各画面
│       │   ├── list.go
│       │   ├── detail.go
│       │   ├── new.go
│       │   └── worktree.go
│       └── styles/                 # スタイル定義
│           └── theme.go
├── test/                           # テスト
│   ├── integration/
│   ├── e2e/
│   └── testutil/
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

#### ユーザープロジェクト構造
```
project/                             # ユーザーのプロジェクト
├── .ticketflow.yaml                # 設定ファイル
├── tickets/                        # チケット管理ディレクトリ
│   ├── todo/                       # 未開始チケット
│   │   └── 250125-093000-add-tests.md
│   ├── doing/                      # 作業中チケット
│   │   └── 250124-150000-implement-auth.md
│   └── done/                       # 完了チケット
│       └── 250123-110000-setup-ci.md
├── current-ticket.md               # 現在のチケットへのシンボリックリンク
└── src/                           # プロジェクトのソースコード

../.worktrees/                      # Worktreeディレクトリ
├── 250124-150000-implement-auth/   # 通常のworktree
└── 250124-150000-implement-auth.sub/  # サブタスク用
    └── 250124-153000-user-model/
```

## 3. データ構造

### 3.1 チケットファイル形式

```yaml
---
# 基本メタデータ（ticket.sh互換）
priority: 2                          # 優先度 (1-3)
description: "ユーザー認証APIの実装"    # 簡潔な説明
created_at: "2025-01-24T10:00:00Z"   # 作成日時（ISO 8601）
started_at: null                     # 開始日時（作業開始時に設定）
closed_at: null                      # 完了日時（クローズ時に設定）

# 拡張フィールド（オプション）
related: ["250123-140000-api-design"] # 関連チケットID
---

# チケットタイトル

## 概要

チケットの詳細な説明をここに記述します。

## タスク
- [ ] JWT生成ロジックの実装
- [ ] 認証ミドルウェアの作成
- [ ] リフレッシュトークン対応

## メモ

実装に関する追加情報や注意事項。
```

### 3.2 チケットの状態

チケットの状態はディレクトリとメタデータから決定されます：

| 状態 | ディレクトリ | started_at | closed_at |
|------|------------|------------|-----------|
| TODO | `todo/` | null | null |
| DOING | `doing/` | 設定済み | null |
| DONE | `done/` | 設定済み | 設定済み |

### 3.3 設定ファイル (.ticketflow.yaml)

```yaml
# Git設定
git:
  default_branch: "main"             # デフォルトブランチ
  
# Worktree設定
worktree:
  enabled: true                      # worktree機能の有効/無効
  base_dir: "../.worktrees"          # worktreeのベースディレクトリ
  
  # worktree作成後の初期化コマンド
  init_commands:
    - "git fetch origin"
    # - "npm install"
    # - "make setup"
    
  # ネストしたworktree（サブタスク）の設定
  allow_nested: true                 # サブタスクのworktreeを許可
  nested_structure: "parent.sub"     # サブタスクディレクトリの命名規則
  max_nest_depth: 2                  # 最大ネスト深度
    
# チケット設定  
tickets:
  dir: "tickets"                     # チケットディレクトリ
  
  # 状態別のサブディレクトリ
  todo_dir: "todo"
  doing_dir: "doing"
  done_dir: "done"
  
  # 新規チケットのテンプレート
  template: |
    # 概要
    
    ## タスク
    - [ ] 
    
    ## メモ

# 出力設定
output:
  default_format: "text"             # デフォルト出力形式 (text|json)
  json_pretty: true                  # JSON整形出力
```

## 4. コマンド仕様

### 4.1 コマンド一覧

| コマンド | 説明 | 実行可能な場所 |
|---------|------|---------------|
| `ticketflow` | TUI起動 | どこでも |
| `ticketflow init` | システム初期化 | Gitリポジトリ |
| `ticketflow new <slug>` | チケット作成 | どこでも（警告付き） |
| `ticketflow list` | チケット一覧 | メインリポジトリ |
| `ticketflow show <id>` | チケット詳細 | メインリポジトリ |
| `ticketflow start <id>` | 作業開始 | メインリポジトリ推奨 |
| `ticketflow close` | 作業完了 | worktree |
| `ticketflow restore` | リンク復元 | worktree |
| `ticketflow cleanup <id>` | 後片付け | メインリポジトリ |
| `ticketflow worktree list` | worktree一覧 | メインリポジトリ |

### 4.2 各コマンドの詳細

#### init - システム初期化
```bash
ticketflow init
```
- `.ticketflow.yaml`を作成
- `tickets/todo`、`tickets/doing`、`tickets/done`ディレクトリを作成
- `.gitignore`に`current-ticket.md`を追加

#### new - チケット作成
```bash
ticketflow new implement-auth
```
- slugは英小文字、数字、ハイフンのみ
- `YYMMDD-HHMMSS-<slug>`形式のIDを生成
- `tickets/todo/`ディレクトリに作成
- どのブランチでも実行可能（mainブランチ以外では警告）

#### start - 作業開始
```bash
ticketflow start 250124-150000-implement-auth
```
1. 新しいworktreeとブランチを作成
2. チケットを`todo/` → `doing/`に移動
3. `started_at`を設定
4. 初期化コマンドを実行
5. `current-ticket.md`シンボリックリンクを作成

**サブチケットの場合**（worktree内から実行）:
- 親ブランチから分岐
- `../.worktrees/parent.sub/child/`に作成

#### close - 作業完了
```bash
ticketflow close
```
1. チケットを`doing/` → `done/`に移動
2. `closed_at`を設定
3. 変更をコミット
4. **マージやプッシュは行わない**
5. PR作成を案内

#### cleanup - 後片付け
```bash
ticketflow cleanup 250124-150000-implement-auth
```
PRマージ後の後片付け：
- worktreeを削除
- ローカルブランチを削除

### 4.3 JSON出力

読み取り系コマンドは`--format json`オプションでJSON出力可能：

```json
{
  "tickets": [
    {
      "id": "250124-150000-implement-auth",
      "path": "tickets/doing/250124-150000-implement-auth.md",
      "status": "doing",
      "priority": 1,
      "description": "ユーザー認証APIの実装",
      "created_at": "2025-01-24T15:00:00Z",
      "started_at": "2025-01-24T15:30:00Z",
      "closed_at": null,
      "related": ["250123-140000-api-design"],
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

## 5. ワークフロー

### 5.1 基本的なワークフロー

```bash
# 1. システム初期化（初回のみ）
[main]$ ticketflow init

# 2. チケット作成
[main]$ ticketflow new implement-auth
Created ticket: tickets/todo/250124-150000-implement-auth.md

# 3. チケット編集（説明やタスクを記入）
[main]$ $EDITOR tickets/todo/250124-150000-implement-auth.md

# 4. 作業開始
[main]$ ticketflow start 250124-150000-implement-auth
Creating worktree: ../.worktrees/250124-150000-implement-auth
Moving ticket: todo → doing
Next: cd ../.worktrees/250124-150000-implement-auth

# 5. worktreeで開発
[main]$ cd ../.worktrees/250124-150000-implement-auth
[implement-auth]$ # 実装作業...
[implement-auth]$ git add .
[implement-auth]$ git commit -m "Implement authentication"

# 6. 作業完了
[implement-auth]$ ticketflow close
Moving ticket: doing → done
Committed: "Close ticket: 250124-150000-implement-auth"

Next steps:
1. Push your changes: git push origin 250124-150000-implement-auth
2. Create Pull Request on your Git service
3. After PR is merged: ticketflow cleanup 250124-150000-implement-auth

# 7. プッシュとPR作成
[implement-auth]$ git push origin 250124-150000-implement-auth
# GitHub/GitLab等でPRを作成

# 8. PRマージ後の後片付け
[implement-auth]$ cd ../../project
[main]$ git pull
[main]$ ticketflow cleanup 250124-150000-implement-auth
Removing worktree: ../.worktrees/250124-150000-implement-auth
Deleting branch: 250124-150000-implement-auth
✅ Cleanup completed
```

### 5.2 サブチケットのワークフロー

```bash
# 1. 親タスクで作業中
[user-system]$ pwd
/path/to/.worktrees/user-system

# 2. タスクが大きすぎることに気づき、サブチケットを作成
[user-system]$ ticketflow new user-model
Creating ticket in worktree branch: user-system
Created: tickets/todo/250124-153000-user-model.md

[user-system]$ ticketflow new user-auth
Created: tickets/todo/250124-153100-user-auth.md

# 3. 現在の作業を一旦コミット
[user-system]$ git add .
[user-system]$ git commit -m "WIP: Creating sub-tasks"

# 4. サブチケットを開始（worktreeから直接）
[user-system]$ ticketflow start user-model
Creating nested worktree:
  Parent: ../.worktrees/user-system
  Child:  ../.worktrees/user-system.sub/user-model

# 5. サブタスクで作業
[user-system]$ cd ../user-system.sub/user-model
[user-model]$ # モデルの実装...
[user-model]$ git commit -m "Implement user model"

# 6. サブタスク完了
[user-model]$ ticketflow close
Next: Create PR with base branch 'user-system'

# 7. 親ブランチに向けてPR作成
[user-model]$ git push origin user-model
# PR: user-system ← user-model

# 8. 親タスクで続きの作業
[user-model]$ cd ../../user-system
[user-system]$ git pull  # マージされたサブタスクを取り込む
```

## 6. エラーハンドリング

### 6.1 エラーコード

```go
const (
    // システムエラー
    ErrNotGitRepo           = "NOT_GIT_REPO"
    ErrConfigNotFound       = "CONFIG_NOT_FOUND"
    ErrPermissionDenied     = "PERMISSION_DENIED"
    
    // チケットエラー
    ErrTicketNotFound       = "TICKET_NOT_FOUND"
    ErrTicketExists         = "TICKET_EXISTS"
    ErrTicketAlreadyStarted = "TICKET_ALREADY_STARTED"
    
    // Git/Worktreeエラー
    ErrGitDirtyWorkspace    = "GIT_DIRTY_WORKSPACE"
    ErrWorktreeExists       = "WORKTREE_EXISTS"
    ErrMaxNestDepth         = "MAX_NEST_DEPTH"
)
```

### 6.2 エラーメッセージ形式

**人間向け（text）**:
```
Error: Ticket not found
File 'tickets/todo/250124-150000-auth.md' does not exist.

Suggestions:
1. Check ticket ID: ticketflow list
2. Create new ticket: ticketflow new auth
```

**AI向け（JSON）**:
```json
{
  "error": {
    "code": "TICKET_NOT_FOUND",
    "message": "Ticket not found",
    "details": "File 'tickets/todo/250124-150000-auth.md' does not exist",
    "suggestions": [
      "Check ticket ID with 'ticketflow list'",
      "Create new ticket with 'ticketflow new auth'"
    ]
  }
}
```

## 7. TUI仕様

### 7.1 画面構成

1. **チケット一覧画面**（起動時デフォルト）
   - TODO/DOING/DONEタブ
   - チケットの検索
   - ステータス別の表示

2. **チケット詳細画面**
   - Markdownプレビュー
   - メタデータ表示
   - 操作ボタン

3. **新規作成画面**
   - slug入力
   - 優先度選択
   - 説明入力

4. **Worktree管理画面**
   - アクティブなworktree一覧
   - ステータス表示（clean/modified）

### 7.2 キーバインド

| キー | アクション |
|-----|-----------|
| `Tab` | タブ切り替え |
| `j`/`k` | 上下移動 |
| `Enter` | 詳細表示/選択 |
| `n` | 新規作成 |
| `s` | 作業開始 |
| `c` | クローズ |
| `w` | Worktree一覧 |
| `/` | 検索 |
| `?` | ヘルプ |
| `q` | 終了 |

## 8. 実装計画

### Phase 1: コア機能（MVP）
- [ ] 設定ファイル管理
- [ ] チケットモデル（todo/doing/done）
- [ ] 基本CLIコマンド（init, new, list）
- [ ] 簡易的なstart/close（worktreeなし）

### Phase 2: Worktree統合
- [ ] Git操作の抽象化
- [ ] Worktree管理機能
- [ ] start/closeのworktree対応
- [ ] cleanupコマンド

### Phase 3: 高度な機能
- [ ] サブチケット（ネストしたworktree）
- [ ] JSON出力対応
- [ ] エラーハンドリングの洗練
- [ ] restoreコマンド

### Phase 4: TUI実装
- [ ] Bubble Tea基本構造
- [ ] 各画面の実装
- [ ] キーバインド
- [ ] スタイリング

## 9. テスト戦略

### 9.1 テストカバレッジ目標
- Unit Tests: 80%以上
- Integration Tests: 主要ワークフロー
- E2E Tests: クリティカルパス

### 9.2 テストケース
- チケットのライフサイクル
- Worktree操作
- エラーケース
- サブチケットワークフロー

## 10. 将来の拡張可能性

- プラグインシステム
- カスタムワークフロー
- チーム機能（共有設定）
- 統計・レポート機能

---

この設計書は、TicketFlowの完全な仕様を定義しています。実装はPhase 1から順次進めることで、段階的に機能を追加していく計画です。