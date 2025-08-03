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
- [ ] Update AddWorktree to handle existing branches (use branch without -b if exists)
- [ ] Add unit tests for BranchExists method
- [ ] Add tests for AddWorktree with existing/non-existing branches
- [ ] Improve error message when worktree already exists (show path)

## Sub-tickets Created
- [ ] 250803-121450-handle-diverged-branch - Handle case when branch points to different commit
- [ ] 250803-121506-worktree-recovery - Add worktree recovery mechanisms
- [ ] 250803-121521-robustness-improvements - Add permission and concurrency handling

## 技術仕様 (Core Fix Only)

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

### Core Fix Focus
This ticket focuses only on fixing the immediate "branch already exists" error. Edge cases are handled in sub-tickets:
- **250803-121450-handle-diverged-branch**: Branch pointing to different commit
- **250803-121506-worktree-recovery**: Corrupted worktree references  
- **250803-121521-robustness-improvements**: Permission issues and concurrent operations

## メモ

[追加の注意事項やメモ]
