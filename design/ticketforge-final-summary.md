# TicketForge Implementation Summary

## 🎯 実装の要点

Claude Codeで実装する際の重要なポイントをまとめました。

## 📁 最終的なディレクトリ構造

```
ticketforge/                       # プロジェクトルート
├── .ticketforge.yaml             # 設定ファイル
├── tickets/                      # チケット管理ディレクトリ
│   ├── todo/                     # 未開始 (started_at: null)
│   │   └── 250125-093000-add-tests.md
│   ├── doing/                    # 作業中 (started_at: set, closed_at: null)
│   │   └── 250124-150000-implement-auth.md
│   └── done/                     # 完了 (closed_at: set)
│       └── 250123-110000-setup-ci.md
├── current-ticket.md             # -> tickets/doing/250124-150000-implement-auth.md
└── src/                          # プロジェクトのソースコード

../.worktrees/                    # Worktreeディレクトリ（設定可能）
└── 250124-150000-implement-auth/ # チケットIDと同名
    ├── .git                      # メインリポジトリへの参照
    ├── src/                      # メインと同じ構造
    └── current-ticket.md         # シンボリックリンク
```

## 🔄 チケットのライフサイクル

```mermaid
graph LR
    A[new] -->|ファイル作成| B[tickets/todo/]
    B -->|start| C[tickets/doing/]
    C -->|+ worktree作成| D[作業中]
    D -->|close| E[tickets/done/]
    E -->|+ worktree削除| F[完了]
```

## 📝 実装の優先順位（更新版）

### Phase 1: コア機能
1. **設定管理** (`internal/config/`)
   - `.ticketforge.yaml` の読み込み/作成
   - デフォルト値の設定
   - ディレクトリ構造（todo/doing/done）のサポート

2. **チケットモデル** (`internal/ticket/`)
   - Markdown + YAML frontmatterのパース
   - 状態管理（ディレクトリベース）
   - ファイル操作（作成、移動、読み込み）

3. **基本CLIコマンド**
   - `init`: 初期化（ディレクトリ作成含む）
   - `new`: tickets/todo/ に作成
   - `list`: 各ディレクトリから読み込み
   - `start`: todo → doing へ移動（worktreeなし）
   - `close`: doing → done へ移動

### Phase 2: Git Worktree統合
1. **Git操作** (`internal/git/`)
   - 基本的なGitコマンドのラッパー
   - エラーハンドリング

2. **Worktree管理** (`internal/git/worktree.go`)
   - `git worktree add -b` での作成
   - 初期化コマンドの実行
   - worktreeの削除

3. **コマンドの拡張**
   - `start`: worktree作成を追加
   - `close`: worktree削除を追加
   - `worktree list/clean`: 管理コマンド

### Phase 3: TUI実装
- Bubble Teaによるインタラクティブな操作
- ディレクトリ別のタブ表示（TODO/DOING/DONE）

### Phase 4: 高度な機能
- JSON出力（`--format json`）
- エラーメッセージのJSON対応
- 移行ツール（既存プロジェクト対応）

## 🔑 重要な実装詳細

### 1. チケットID = ブランチ名 = Worktreeディレクトリ名
```go
ticketID := "250124-150000-implement-auth"
branchName := ticketID                    // 同じ
worktreePath := "../.worktrees/" + ticketID  // 同じ
```

### 2. ディレクトリ移動とGit
```bash
# startコマンドでの移動
git add tickets/todo/250124-150000-implement-auth.md
git add tickets/doing/250124-150000-implement-auth.md
git commit -m "Start ticket: 250124-150000-implement-auth"

# closeコマンドでの移動
git add tickets/doing/250124-150000-implement-auth.md
git add tickets/done/250124-150000-implement-auth.md
git commit -m "Archive ticket: 250124-150000-implement-auth"
```

### 3. Worktree作成フロー
```go
// 新しいブランチとworktreeを同時に作成
git worktree add -b 250124-150000-implement-auth ../worktrees/250124-150000-implement-auth

// 初期化コマンドを実行（エラーは警告のみ）
cd ../worktrees/250124-150000-implement-auth
git pull origin main --rebase
npm install  // etc.
```

### 4. チケット検索の優先順位
```go
// FindTicket()の検索順序
1. tickets/todo/
2. tickets/doing/
3. tickets/done/
```

## ⚠️ 実装時の注意点

1. **ディレクトリの自動作成**
   - 各コマンドで必要なディレクトリを自動作成
   - `os.MkdirAll()` でエラーを防ぐ

2. **後方互換性**
   - 既存のフラット構造からの移行を考慮
   - `init` コマンドで移行オプションを提供

3. **エラーハンドリング**
   - ファイル移動失敗時のロールバック
   - Worktree作成失敗時のクリーンアップ

4. **相対パスの使用**
   - シンボリックリンクは相対パスで作成
   - worktreeパスも設定に応じて相対/絶対を選択

## 🚀 実装開始

```bash
# プロジェクトセットアップ
mkdir ticketforge
cd ticketforge
go mod init github.com/yourusername/ticketforge

# 依存関係の追加
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/stretchr/testify
go get gopkg.in/yaml.v3

# ディレクトリ構造の作成
mkdir -p cmd/ticketforge
mkdir -p internal/{config,ticket,git,ui/views,ui/styles,cli}
mkdir -p test/{integration,e2e,testutil}

# Phase 1から実装開始！
```

## 📚 参照ドキュメント

1. **要件定義と実装ガイド** - 全体の仕様
2. **実装例** - 具体的なコード例
3. **Git Workflow仕様** - Git/Worktreeの詳細
4. **ディレクトリ構造更新** - todo/doing/doneの仕様

すべての準備が整いました。Phase 1から順番に実装を進めてください！

Good luck with the implementation! 🎉