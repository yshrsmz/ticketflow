---
priority: 2
description: "Fix list command JSON summary to include all ticket counts"
created_at: 2025-07-26T21:12:29.237378+09:00
started_at: 2025-07-26T21:30:18+09:00
closed_at: 2025-07-26T21:35:00+09:00
---

# 概要

tree command says there's some tickets in done directory:

```sh
% tree tickets     
tickets
├── done
│   ├── 250726-181410-fix-empty-status-tab.md
│   └── 250726-182406-auto-create-ticket-dirs.md
└── todo
    ├── 250726-181525-global-shortcut-misbehavior.md
    ├── 250726-183403-fix-branch-already-exist-on-start.md
    └── 250726-211229-list-command-empty-doing-done.md

3 directories, 5 files
```

but `ticketflow list` returns only tickets in todo directory

```sh
% ./dist/ticketflow-darwin-arm64 list --format json               
{
  "summary": {
    "doing": 0,
    "done": 0,
    "todo": 3,
    "total": 3
  },
  "tickets": [
    {
      "closed_at": null,
      "created_at": "2025-07-26T21:12:29.237378+09:00",
      "description": "",
      "has_worktree": false,
      "id": "250726-211229-list-command-empty-doing-done",
      "path": "/Users/a12897/repos/github.com/yshrsmz/ticketflow/tickets/todo/250726-211229-list-command-empty-doing-done.md",
      "priority": 2,
      "related": null,
      "started_at": null,
      "status": "todo"
    },
    {
      "closed_at": null,
      "created_at": "2025-07-26T18:34:03.720259+09:00",
      "description": "",
      "has_worktree": false,
      "id": "250726-183403-fix-branch-already-exist-on-start",
      "path": "/Users/a12897/repos/github.com/yshrsmz/ticketflow/tickets/todo/250726-183403-fix-branch-already-exist-on-start.md",
      "priority": 2,
      "related": null,
      "started_at": null,
      "status": "todo"
    },
    {
      "closed_at": null,
      "created_at": "2025-07-26T18:15:25.212321+09:00",
      "description": "",
      "has_worktree": false,
      "id": "250726-181525-global-shortcut-misbehavior",
      "path": "/Users/a12897/repos/github.com/yshrsmz/ticketflow/tickets/todo/250726-181525-global-shortcut-misbehavior.md",
      "priority": 2,
      "related": null,
      "started_at": null,
      "status": "todo"
    }
  ]
}
```


## タスク
- [x] Investigate the list command implementation
- [x] Find why done tickets are not being listed
- [x] Fix the getDirectoriesForStatus function to include done directory when no filter is specified
- [x] Fix config file mismatch (archive_dir -> done_dir)
- [x] Test with both JSON and table formats

## 技術仕様

The issue was revised based on user feedback:
- The `list` command without status filter should only show active tickets (todo and doing) - this is the intended behavior
- However, the JSON format summary should show counts for ALL tickets including done

### Actual Fix Applied:

1. **Kept** `getDirectoriesForStatus` to return only todo/doing for empty filter (this is correct behavior)

2. **Fixed config issue**: Updated `.ticketflow.yaml`:
```yaml
tickets:
    done_dir: done  # Changed from archive_dir
```

3. **Updated JSON output** in `outputTicketListJSON` to always fetch all tickets for summary:
```go
// Always calculate full summary from all tickets
allTickets, err := app.Manager.List("all")
// ... calculate counts from allTickets
```

4. **Updated status command** to also use `List("all")` for correct summary counts

## メモ

- The list command correctly shows only active tickets (todo, doing) when no status filter is specified
- The JSON summary now includes counts for all tickets including done
- The status command also shows correct total counts
- Both JSON and table output formats work correctly

## Implementation Insights

During the fix, I discovered:

1. **User Intent**: The original behavior of showing only active tickets (todo/doing) in the list command was actually intentional. The issue was only with the JSON summary counts.

2. **Manager Design**: The `getDirectoriesForStatus` function in manager.go was correctly designed with different behaviors:
   - Empty string ("") = active tickets only (todo + doing)
   - "all" or any other string = all tickets (todo + doing + done)
   - Specific status strings work as filters

3. **Config Field Naming**: The YAML config was using `archive_dir` but the Go struct expected `done_dir`. This mismatch would have caused issues if not caught.

4. **Separation of Concerns**: The fix showed good separation between:
   - What tickets to display (filtered list)
   - What counts to show in summary (always all tickets)
   
This ensures users see relevant active work by default while still getting a complete overview in the summary.
