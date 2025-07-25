# TicketFlow - Requirements and Implementation Guide

## 1. プロジェクト概要

**TicketFlow**は、生成AIとの協働に最適化されたgit worktreeベースのチケット管理システムです。

### 主な特徴
- ticket.shの設計思想を継承（Markdown + YAML frontmatter）
- Git worktreeによる並行作業のサポート
- 人間向けのTUIとAI向けのCLIの両立
- 個人開発での利用を想定（AI 80% / 人間 20%の協働）

### 技術スタック
- 言語: Go
- TUIフレームワーク: Bubble Tea (github.com/charmbracelet/bubbletea)
- テストフレームワーク: testify (github.com/stretchr/testify)

## 2. ディレクトリ構造

```
ticketflow/
├── cmd/
│   └── ticketflow/
│       └── main.go              # エントリーポイント
├── internal/
│   ├── config/
│   │   ├── config.go            # 設定ファイル管理
│   │   └── config_test.go
│   ├── ticket/
│   │   ├── ticket.go            # チケットモデル
│   │   ├── ticket_test.go
│   │   ├── manager.go           # チケット操作ロジック
│   │   └── manager_test.go
│   ├── git/
│   │   ├── git.go               # Git操作の抽象化
│   │   ├── git_test.go
│   │   ├── worktree.go          # Worktree管理
│   │   └── worktree_test.go
│   ├── ui/
│   │   ├── app.go               # TUIアプリケーション
│   │   ├── views/
│   │   │   ├── list.go          # チケット一覧画面
│   │   │   ├── detail.go        # チケット詳細画面
│   │   │   ├── new.go           # 新規作成画面
│   │   │   └── worktree.go      # Worktree管理画面
│   │   ├── styles/
│   │   │   └── theme.go         # UIテーマ定義
│   │   └── components/
│   │       └── help.go          # ヘルプオーバーレイ
│   └── cli/
│       ├── commands.go          # CLIコマンド実装
│       ├── output.go            # 出力フォーマット管理
│       └── errors.go            # エラーハンドリング
├── test/
│   ├── integration/
│   │   └── workflow_test.go     # 統合テスト
│   ├── e2e/
│   │   └── cli_test.go          # E2Eテスト
│   └── testutil/
│       ├── git.go               # Gitテストユーティリティ
│       └── filesystem.go        # ファイルシステムヘルパー
├── .gitignore
├── .ticketflow.yaml.example    # 設定ファイルサンプル
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 3. データ構造

### 3.1 チケットファイル形式

```yaml
---
# ticket.sh互換のフィールド
priority: 2
description: "ユーザー認証APIの実装"
created_at: "2025-01-24T10:00:00Z"
started_at: null
closed_at: null

# 拡張フィールド（オプション）
related: ["240623-database-schema"]
---

# 概要

チケットの詳細内容をMarkdownで記述

## タスク
- [ ] JWT生成ロジックの実装
- [ ] 認証ミドルウェアの作成
- [ ] リフレッシュトークン対応

## メモ
追加の注意事項など
```

### 3.2 設定ファイル形式 (.ticketflow.yaml)

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
    - "git pull origin main --rebase"
    - "npm install"  # プロジェクトに応じて変更
    
  auto_operations:
    create_on_start: true
    remove_on_close: true
    cleanup_orphaned: true
    
# チケット設定  
tickets:
  dir: "tickets"
  archive_dir: "tickets/done"
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

## 4. コマンド仕様

### 4.1 基本コマンド

```bash
# TUI起動（引数なし）
ticketflow

# 初期化
ticketflow init

# チケット操作
ticketflow new <slug>                    # 新規作成
ticketflow list [--status STATUS] [--format json]  # 一覧表示
ticketflow show <ticket-id> [--format json]        # 詳細表示
ticketflow start <ticket-id> [--no-push]           # 作業開始
ticketflow close [--no-push] [--force]             # 作業完了
ticketflow restore                                  # リンク復元

# Worktree管理
ticketflow worktree list [--format json]
ticketflow worktree clean

# その他
ticketflow status [--format json]        # 現在の状態
ticketflow help                          # ヘルプ表示
```

### 4.2 JSON出力形式

#### list --format json
```json
{
  "tickets": [
    {
      "id": "250124-150000-implement-auth",
      "path": "tickets/250124-150000-implement-auth.md",
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
    "total": 15,
    "todo": 8,
    "doing": 3,
    "done": 4
  }
}
```

## 5. エラーコード

```go
const (
    // System errors
    ErrNotGitRepo      = "NOT_GIT_REPO"
    ErrConfigNotFound  = "CONFIG_NOT_FOUND"
    ErrConfigInvalid   = "CONFIG_INVALID"
    ErrPermissionDenied = "PERMISSION_DENIED"
    
    // Ticket errors
    ErrTicketNotFound  = "TICKET_NOT_FOUND"
    ErrTicketExists    = "TICKET_EXISTS"
    ErrTicketInvalid   = "TICKET_INVALID"
    ErrTicketNotStarted = "TICKET_NOT_STARTED"
    ErrTicketAlreadyStarted = "TICKET_ALREADY_STARTED"
    ErrTicketAlreadyClosed = "TICKET_ALREADY_CLOSED"
    
    // Git errors
    ErrGitDirtyWorkspace = "GIT_DIRTY_WORKSPACE"
    ErrGitBranchExists   = "GIT_BRANCH_EXISTS"
    ErrGitMergeFailed    = "GIT_MERGE_FAILED"
    ErrGitPushFailed     = "GIT_PUSH_FAILED"
    
    // Worktree errors
    ErrWorktreeExists    = "WORKTREE_EXISTS"
    ErrWorktreeNotFound  = "WORKTREE_NOT_FOUND"
    ErrWorktreeCreateFailed = "WORKTREE_CREATE_FAILED"
    ErrWorktreeRemoveFailed = "WORKTREE_REMOVE_FAILED"
)
```

## 6. 実装の優先順位

### Phase 1: コア機能（MVP）
1. 設定ファイル管理 (`internal/config/`)
2. チケットモデルとパーサー (`internal/ticket/ticket.go`)
3. 基本的なCLIコマンド (`internal/cli/`)
   - init
   - new
   - list
   - start (worktreeなし)
   - close

### Phase 2: Worktree統合
1. Git操作の抽象化 (`internal/git/`)
2. Worktree管理 (`internal/git/worktree.go`)
3. startコマンドのworktree対応
4. worktreeコマンド群

### Phase 3: TUI実装
1. Bubble Teaアプリケーション基盤 (`internal/ui/app.go`)
2. 各画面の実装
   - チケット一覧
   - チケット詳細
   - 新規作成
   - Worktree管理

### Phase 4: 高度な機能
1. JSON出力対応
2. エラーハンドリングの洗練
3. 進捗レポート機能
4. 自動クリーンアップ

## 7. テスト戦略

### 7.1 テストの種類と目標
- Unit Tests: 80%以上のカバレッジ
- Integration Tests: 主要ワークフローを網羅
- E2E Tests: クリティカルパスのみ

### 7.2 テストファイルの配置
- 各実装ファイルと同じディレクトリに `*_test.go` を配置
- 統合テストは `test/integration/` に配置
- E2Eテストは `test/e2e/` に配置

## 8. Makefile

```makefile
.PHONY: build test clean install

# デフォルトターゲット
all: test build

# ビルド
build:
	go build -o ticketflow ./cmd/ticketflow

# インストール
install: build
	cp ticketflow $(GOPATH)/bin/

# テスト
test:
	go test -v ./...

test-unit:
	go test -v -race ./internal/...

test-integration:
	go test -v -race ./test/integration/...

test-e2e: build
	go test -v ./test/e2e/...

# カバレッジ
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# リント
lint:
	golangci-lint run

# クリーン
clean:
	rm -f ticketflow
	rm -f coverage.out coverage.html

# 依存関係の更新
deps:
	go mod download
	go mod tidy

# 開発用の実行
run: build
	./ticketflow

# TUIの実行
run-tui: build
	./ticketflow
```

## 9. 開発開始の手順

1. プロジェクトディレクトリの作成
   ```bash
   mkdir ticketflow
   cd ticketflow
   ```

2. Go モジュールの初期化
   ```bash
   go mod init github.com/yshrsmz/ticketflow
   ```

3. 依存関係の追加
   ```bash
   go get github.com/charmbracelet/bubbletea
   go get github.com/charmbracelet/lipgloss
   go get github.com/stretchr/testify
   go get gopkg.in/yaml.v3
   ```

4. Phase 1から順に実装を進める

## 10. Git worktreeの考慮事項

- チケットIDとworktreeディレクトリ名を同一にする
- ブランチ名もチケットIDと同じ（feature/プレフィックスなし）
- worktree作成時に初期化コマンドを実行
- 孤立したworktreeの自動検出と削除

## 11. AIとの協働のベストプラクティス

- チケット作成時は構造化されたMarkdownを使用
- タスクリストは明確で実行可能な単位に分割
- 関連ファイルのパスを明記
- 技術要件を具体的に記載

---

このドキュメントを参考に、Phase 1から順次実装を進めてください。質問があれば適宜確認してください。
