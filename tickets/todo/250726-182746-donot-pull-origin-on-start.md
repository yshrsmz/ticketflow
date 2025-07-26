---
priority: 2
description: ""
created_at: 2025-07-26T18:27:46.051888+09:00
started_at: null
closed_at: null
---

# 概要

Fix the following error

```sh
! dist/ticketflow-linux-arm64 start 250726-181410-fix-empty-status-tab
  ⎿  Running initialization commands...
       $ git pull origin main --rebase
     Warning: Command failed: exit status 1
     … +5 lines (ctrl+r to expand)
  ⎿ Error: failed to move ticket to doing: rename /workspaces/ticketflow/tickets/todo/250726-181410-fix-empty-status-tab.md /workspaces/ticketfl
    ow/tickets/doing/250726-181410-fix-empty-status-tab.md: no such file or directory
```

We should not pull, just stick to local repository

## タスク
- [ ] タスク1
- [ ] タスク2
- [ ] タスク3

## 技術仕様

[必要に応じて技術的な詳細を記述]

## メモ

[追加の注意事項やメモ]
