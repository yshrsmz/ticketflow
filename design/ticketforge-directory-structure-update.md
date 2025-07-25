# TicketForge Directory Structure Update

## 更新されたディレクトリ構造

### 基本構造

```
project/
├── .ticketforge.yaml
├── tickets/
│   ├── todo/      # 未開始のチケット (started_at: null)
│   │   └── 250124-150000-implement-auth.md
│   ├── doing/     # 作業中のチケット (started_at: set, closed_at: null)
│   │   └── 250124-140000-fix-bug.md
│   └── done/      # 完了したチケット (closed_at: set)
│       └── 250124-130000-setup-ci.md
├── current-ticket.md -> tickets/doing/250124-140000-fix-bug.md
└── src/
```

## 設定ファイルの更新

```yaml
# .ticketforge.yaml
tickets:
  # ベースディレクトリ
  dir: "tickets"
  
  # 状態別のサブディレクトリ（tickets/配下）
  todo_dir: "todo"
  doing_dir: "doing" 
  done_dir: "done"
  
  # チケットテンプレート
  template: |
    # 概要
    ...
```

## 各コマンドの動作変更

### 1. `ticketforge new` - チケット作成

```go
func (m *Manager) Create(slug string) (*Ticket, error) {
    // チケットIDを生成
    ticketID := GenerateID(slug)
    
    // todoディレクトリにファイルを作成
    todoDir := filepath.Join(m.config.Tickets.Dir, m.config.Tickets.TodoDir)
    if err := os.MkdirAll(todoDir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create todo directory: %w", err)
    }
    
    ticketPath := filepath.Join(todoDir, ticketID + ".md")
    
    // チケットファイルを作成
    ticket := &Ticket{
        ID:          ticketID,
        Slug:        slug,
        Path:        ticketPath,
        Priority:    2,
        Description: "",
        CreatedAt:   time.Now(),
        Content:     m.config.Tickets.Template,
    }
    
    // ファイルに書き込み
    if err := m.SaveTicket(ticket); err != nil {
        return nil, err
    }
    
    return ticket, nil
}
```

**実行例**:
```bash
$ ticketforge new implement-auth
Created ticket file: tickets/todo/250124-150000-implement-auth.md
Please edit the file to add title, description and details.
```

### 2. `ticketforge start` - 作業開始

```go
func (m *Manager) StartTicket(ticketID string) error {
    // 1. todoディレクトリからチケットを探す
    todoPath := m.findTicketInDir(ticketID, m.config.Tickets.TodoDir)
    if todoPath == "" {
        // doingディレクトリも確認（既に開始済みの可能性）
        doingPath := m.findTicketInDir(ticketID, m.config.Tickets.DoingDir)
        if doingPath != "" {
            return ErrTicketAlreadyStarted
        }
        return ErrTicketNotFound
    }
    
    // 2. チケットを読み込み
    ticket, err := m.LoadTicketFromPath(todoPath)
    if err != nil {
        return err
    }
    
    // 3. Worktree作成（既存のフローと同じ）
    // ...
    
    // 4. チケットのstarted_atを更新
    ticket.StartedAt = timePtr(time.Now())
    
    // 5. todoからdoingへファイルを移動
    doingDir := filepath.Join(m.config.Tickets.Dir, m.config.Tickets.DoingDir)
    if err := os.MkdirAll(doingDir, 0755); err != nil {
        return err
    }
    
    newPath := filepath.Join(doingDir, filepath.Base(todoPath))
    
    // ファイル内容を更新してから移動
    if err := m.SaveTicket(ticket); err != nil {
        return err
    }
    
    if err := os.Rename(todoPath, newPath); err != nil {
        return fmt.Errorf("failed to move ticket to doing: %w", err)
    }
    
    // 6. Git操作
    if err := m.git.Add(todoPath, newPath); err != nil {
        return err
    }
    
    if err := m.git.Commit(fmt.Sprintf("Start ticket: %s", ticketID)); err != nil {
        return err
    }
    
    // 7. 以降は既存のフローと同じ（worktree初期化など）
    // ...
    
    return nil
}
```

**Git操作の流れ**:
```bash
# ファイル移動をGitに認識させる
git add tickets/todo/250124-150000-implement-auth.md
git add tickets/doing/250124-150000-implement-auth.md
git commit -m "Start ticket: 250124-150000-implement-auth"
```

### 3. `ticketforge close` - 作業完了

```go
func (m *Manager) CloseTicket(force bool) error {
    // 1. 現在のチケットを特定（doingディレクトリから）
    currentBranch, _ := m.git.CurrentBranch()
    ticketID := currentBranch
    
    doingPath := m.findTicketInDir(ticketID, m.config.Tickets.DoingDir)
    if doingPath == "" {
        return ErrTicketNotFound
    }
    
    // 2-7. 既存のフロー（チケット更新、マージなど）
    // ...
    
    // 8. doingからdoneへファイルを移動
    doneDir := filepath.Join(m.config.Tickets.Dir, m.config.Tickets.DoneDir)
    if err := os.MkdirAll(doneDir, 0755); err != nil {
        return err
    }
    
    donePath := filepath.Join(doneDir, filepath.Base(doingPath))
    
    if err := os.Rename(doingPath, donePath); err != nil {
        return fmt.Errorf("failed to archive ticket: %w", err)
    }
    
    // 9. アーカイブをコミット
    if err := m.git.Add(doingPath, donePath); err != nil {
        return err
    }
    
    if err := m.git.Commit(fmt.Sprintf("Archive ticket: %s", ticketID)); err != nil {
        return err
    }
    
    // 10. 以降は既存のフロー（worktree削除など）
    // ...
    
    return nil
}
```

### 4. `ticketforge list` - 一覧表示

```go
func (m *Manager) List(status Status) ([]*Ticket, error) {
    var tickets []*Ticket
    
    // 状態に応じてディレクトリを選択
    dirs := m.getDirectoriesForStatus(status)
    
    for _, dir := range dirs {
        dirPath := filepath.Join(m.config.Tickets.Dir, dir)
        
        files, err := os.ReadDir(dirPath)
        if err != nil {
            if os.IsNotExist(err) {
                continue // ディレクトリがなければスキップ
            }
            return nil, err
        }
        
        for _, file := range files {
            if !strings.HasSuffix(file.Name(), ".md") {
                continue
            }
            
            ticketPath := filepath.Join(dirPath, file.Name())
            ticket, err := m.LoadTicketFromPath(ticketPath)
            if err != nil {
                log.Printf("Warning: failed to load ticket %s: %v", ticketPath, err)
                continue
            }
            
            tickets = append(tickets, ticket)
        }
    }
    
    // ソート（優先度順、作成日時順）
    sort.Slice(tickets, func(i, j int) bool {
        if tickets[i].Priority != tickets[j].Priority {
            return tickets[i].Priority < tickets[j].Priority
        }
        return tickets[i].CreatedAt.Before(tickets[j].CreatedAt)
    })
    
    return tickets, nil
}

func (m *Manager) getDirectoriesForStatus(status Status) []string {
    switch status {
    case StatusTodo:
        return []string{m.config.Tickets.TodoDir}
    case StatusDoing:
        return []string{m.config.Tickets.DoingDir}
    case StatusDone:
        return []string{m.config.Tickets.DoneDir}
    case "": // 全て（デフォルト）
        return []string{m.config.Tickets.TodoDir, m.config.Tickets.DoingDir}
    default:
        return []string{m.config.Tickets.TodoDir, m.config.Tickets.DoingDir, m.config.Tickets.DoneDir}
    }
}
```

## ヘルパー関数

```go
// チケットIDから特定のディレクトリ内でファイルを探す
func (m *Manager) findTicketInDir(ticketID, subDir string) string {
    dirPath := filepath.Join(m.config.Tickets.Dir, subDir)
    pattern := filepath.Join(dirPath, ticketID + ".md")
    
    matches, _ := filepath.Glob(pattern)
    if len(matches) > 0 {
        return matches[0]
    }
    
    // 部分一致も試す（ユーザーが時刻部分を省略した場合）
    pattern = filepath.Join(dirPath, "*-" + ticketID + ".md")
    matches, _ = filepath.Glob(pattern)
    if len(matches) == 1 {
        return matches[0]
    }
    
    return ""
}

// 任意のディレクトリからチケットを検索
func (m *Manager) FindTicket(ticketID string) (string, error) {
    // todo → doing → done の順で検索
    for _, dir := range []string{
        m.config.Tickets.TodoDir,
        m.config.Tickets.DoingDir,
        m.config.Tickets.DoneDir,
    } {
        if path := m.findTicketInDir(ticketID, dir); path != "" {
            return path, nil
        }
    }
    
    return "", ErrTicketNotFound
}
```

## 移行時の考慮事項

既存のプロジェクトからの移行を考慮：

```go
// InitializeDirectories creates the status directories if they don't exist
func (m *Manager) InitializeDirectories() error {
    dirs := []string{
        filepath.Join(m.config.Tickets.Dir, m.config.Tickets.TodoDir),
        filepath.Join(m.config.Tickets.Dir, m.config.Tickets.DoingDir),
        filepath.Join(m.config.Tickets.Dir, m.config.Tickets.DoneDir),
    }
    
    for _, dir := range dirs {
        if err := os.MkdirAll(dir, 0755); err != nil {
            return fmt.Errorf("failed to create directory %s: %w", dir, err)
        }
    }
    
    return nil
}

// MigrateExistingTickets moves tickets from flat structure to status directories
func (m *Manager) MigrateExistingTickets() error {
    // tickets/ 直下の .md ファイルを探す
    files, err := os.ReadDir(m.config.Tickets.Dir)
    if err != nil {
        return err
    }
    
    for _, file := range files {
        if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
            continue
        }
        
        oldPath := filepath.Join(m.config.Tickets.Dir, file.Name())
        ticket, err := m.LoadTicketFromPath(oldPath)
        if err != nil {
            log.Printf("Warning: failed to load ticket for migration: %s", oldPath)
            continue
        }
        
        // 状態に応じて移動先を決定
        var targetDir string
        switch ticket.Status() {
        case StatusTodo:
            targetDir = m.config.Tickets.TodoDir
        case StatusDoing:
            targetDir = m.config.Tickets.DoingDir
        case StatusDone:
            targetDir = m.config.Tickets.DoneDir
        }
        
        newPath := filepath.Join(m.config.Tickets.Dir, targetDir, file.Name())
        
        fmt.Printf("Migrating %s to %s\n", file.Name(), targetDir)
        if err := os.Rename(oldPath, newPath); err != nil {
            return fmt.Errorf("failed to migrate %s: %w", file.Name(), err)
        }
    }
    
    return nil
}
```

## 利点

1. **視覚的な進捗管理**: `ls tickets/*` で各状態のチケット数が一目でわかる
2. **シンプルな操作**: ファイルマネージャーでドラッグ&ドロップも可能
3. **Git履歴の明確化**: ファイル移動がチケットの状態遷移を表す
4. **検索の効率化**: 状態別にディレクトリが分かれているため検索が速い

## 表示例

```bash
$ tree tickets/
tickets/
├── todo/
│   ├── 250125-093000-add-tests.md
│   └── 250125-094500-refactor-api.md
├── doing/
│   └── 250124-150000-implement-auth.md
└── done/
    ├── 250123-110000-setup-ci.md
    └── 250124-130000-fix-critical-bug.md

$ ticketforge list
📋 Ticket List
---------------------------
[DOING]
- 250124-150000-implement-auth    Priority: 1
  User authentication implementation

[TODO]
- 250125-093000-add-tests         Priority: 2
  Add unit tests for auth module
- 250125-094500-refactor-api      Priority: 3
  Refactor API endpoints
```

この変更により、ticket.shと同様の直感的なディレクトリ構造でチケットを管理できるようになります。