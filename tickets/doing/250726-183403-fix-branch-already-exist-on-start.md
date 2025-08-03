---
priority: 2
description: ""
created_at: "2025-07-26T18:34:03+09:00"
started_at: "2025-08-03T12:14:44+09:00"
closed_at: null
---

# 概要

fix the following error.

```sh
! dist/ticketflow-linux-arm64 start 250726-181410-fix-empty-status-tab
  ⎿  Error: failed to create worktree: git worktree add /workspaces/ticketflow/.worktrees/250726-181410-fix-empty-status-tab -b
     250726-181410-fix-empty-status-tab failed: exit status 255
     Preparing worktree (new branch '250726-181410-fix-empty-status-tab')
     fatal: a branch named '250726-181410-fix-empty-status-tab' already exists
```

## タスク
- [ ] Add BranchExists method to check if a branch exists
- [ ] Update AddWorktree to handle existing branches
- [ ] Improve error messages for better user guidance
- [ ] Add comprehensive tests for edge cases
- [ ] Update documentation with troubleshooting guide

## 技術仕様

### Problems
1. **Branch exists without worktree**: When `ticketflow start` is called and the branch exists but worktree doesn't, it fails with "fatal: a branch named '<branch>' already exists"
2. **Worktree already exists**: Currently returns an error, but could be handled more gracefully

### Root Causes
1. Branch can exist from previous incomplete cleanup or manual branch creation
2. The `git worktree add -b` command always tries to create a new branch
3. Worktree existence check prevents restarting work on a ticket

### Solution Design

#### 1. Add Branch Existence Check
Create new method in `internal/git/git.go`:
```go
func (g *Git) BranchExists(ctx context.Context, branch string) (bool, error)
```
Use `git show-ref --verify --quiet refs/heads/<branch>` to check branch existence

#### 2. Update AddWorktree to Handle Existing Branches
Modify `internal/git/worktree.go`:
- Check if branch exists before creating worktree
- If branch exists: `git worktree add <path> <branch>` (without -b)
- If branch doesn't exist: `git worktree add <path> -b <branch>` (current behavior)

#### 3. Improve Error Handling
- Enhance error messages in `checkExistingWorktree` to show worktree path
- Provide clear suggestions for users when errors occur
- Add option to force recreate if needed

### Implementation Steps
1. Implement `BranchExists` method with proper error handling
2. Update `AddWorktree` to conditionally use `-b` flag
3. Write unit tests for `BranchExists`
4. Write tests for `AddWorktree` with existing/non-existing branches
5. Add integration tests for the complete workflow
6. Update error messages to be more helpful

### Edge Cases to Handle

#### 1. Branch exists pointing to different commit than expected
**Scenario**: User manually created a branch with same name or branch exists from previous work
**Solution**:
- Check if branch's HEAD matches the commit where we expect to start (usually main/master HEAD)
- If different, provide options:
  ```
  Branch '250726-181410-fix-empty-status-tab' already exists but points to a different commit.
  Current branch HEAD: abc123 (2 commits ahead of main)
  Expected base: def456 (main)
  
  Options:
  1. Use existing branch (continue previous work)
  2. Delete and recreate branch from main
  3. Cancel operation
  ```
- Implementation: Compare `git rev-parse <branch>` with `git rev-parse <default-branch>`

#### 2. Corrupted worktree references
**Scenario**: Worktree directory deleted but git still tracks it, or .git/worktrees entry corrupted
**Solution**:
- When `git worktree add` fails with worktree errors, try:
  1. Run `git worktree prune` to clean stale entries
  2. Retry the operation
  3. If still fails, check if worktree path exists without git tracking
- Add recovery command: `ticketflow doctor --fix-worktrees`
- Implementation:
  ```go
  if err != nil && strings.Contains(err.Error(), "worktree") {
      // Try to prune and retry
      g.PruneWorktrees(ctx)
      // Retry add operation
  }
  ```

#### 3. Permission issues
**Scenario**: No write permissions to worktree directory or git directories
**Solution**:
- Check permissions before operations:
  ```go
  if err := checkWritePermission(worktreeBaseDir); err != nil {
      return fmt.Errorf("no write permission to %s: %w", worktreeBaseDir, err)
  }
  ```
- Provide clear error messages with fix suggestions:
  ```
  Error: Permission denied creating worktree at /path/to/worktrees
  Try: sudo chown -R $(whoami) /path/to/worktrees
  ```
- Fall back to non-worktree mode if configured

#### 4. Concurrent operations on same ticket
**Scenario**: Multiple users or processes trying to start the same ticket simultaneously
**Solution**:
- Implement file-based locking in tickets directory:
  ```go
  lockFile := filepath.Join(ticketDir, fmt.Sprintf(".%s.lock", ticketID))
  if err := acquireLock(lockFile, 30*time.Second); err != nil {
      return fmt.Errorf("ticket is being modified by another process")
  }
  defer releaseLock(lockFile)
  ```
- Add timeout for lock acquisition (30 seconds)
- Show which process holds the lock if possible
- Add `--force` flag to override stale locks

## メモ

[追加の注意事項やメモ]
