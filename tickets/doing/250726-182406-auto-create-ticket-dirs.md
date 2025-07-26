---
priority: 2
description: auto create ticket directories if it does not exist
created_at: 2025-07-26T18:24:06.535876+09:00
started_at: 2025-07-26T20:29:50.30535+09:00
closed_at: null
---

# 概要

Currently when we execute `ticketflow start` and `doing` directory does not exist, it crashes with error

```
dist/ticketflow-linux-arm64 start 250726-181410-fix-empty-status-tab
  ⎿  Running initialization commands...
       $ git status

  ⎿  Error: failed to move ticket to doing: rename /workspaces/ticketflow/tickets/todo/250726-181410-fix-empty-status-tab.md
     /workspaces/ticketflow/tickets/doing/250726-181410-fix-empty-status-tab.md: no such file or directory
```

I think this should be because we don't create todo/doing/done directory if it does not exist.

## タスク
- [x] analyze the issue
- [x] update this doc with analysis result and solution
- [x] update tasks
- [x] implement directory auto-creation
- [x] add tests
- [x] verify fix works

## 技術仕様

### Root Cause Analysis
The issue occurs in `internal/cli/commands.go` in the `StartTicket` and `CloseTicket` functions. When moving tickets between status directories (todo → doing → done), the code attempts to rename the file without first ensuring the target directory exists.

### Solution
Added `os.MkdirAll` calls before attempting to move ticket files:

1. In `StartTicket` (line 334): Create doing directory before moving from todo
2. In `CloseTicket` (line 515): Create done directory before moving from doing

### Implementation Details
```go
// Ensure doing directory exists
if err := os.MkdirAll(doingPath, 0755); err != nil {
    // Rollback worktree/branch changes
    return fmt.Errorf("failed to create doing directory: %w", err)
}
```

### Tests Added
Created comprehensive integration tests in `test/integration/directory_creation_test.go`:
- `TestDirectoryAutoCreation`: Tests directory creation without worktrees
- `TestDirectoryCreationWithWorktrees`: Tests directory creation with worktrees enabled

## メモ

- The fix ensures that missing directories are created automatically during ticket transitions
- No changes needed for todo directory creation as it's already handled in `Manager.Create()`
- All existing tests pass with the new changes

### test feedback 1

I got following error

```sh
% ./dist/ticketflow-darwin-arm64 start 250726-182406-auto-create-ticket-dirs
Running initialization commands...
  $ git status
Error: failed to stage ticket move: git add /Users/a12897/repos/github.com/yshrsmz/ticketflow/tickets/todo/250726-182406-auto-create-ticket-dirs.md /Users/a12897/repos/github.com/yshrsmz/ticketflow/tickets/doing/250726-182406-auto-create-ticket-dirs.md failed: exit status 128
fatal: pathspec '/Users/a12897/repos/github.com/yshrsmz/ticketflow/tickets/todo/250726-182406-auto-create-ticket-dirs.md' did not match any files
```

#### Solution
The issue was that after moving the file with `os.Rename`, we were trying to `git add` both the old and new paths. However, the old path no longer exists after the rename, causing git to fail.

Fixed by changing the git add command to use `-A` flag with the directories instead of individual files:
```go
// Old code:
if err := app.Git.Add(oldPath, newPath); err != nil {

// New code:
if err := app.Git.Add("-A", filepath.Dir(oldPath), filepath.Dir(newPath)); err != nil {
```

This tells git to add all changes (including deletions and additions) in both directories, properly tracking the file move operation.
