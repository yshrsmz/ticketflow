---
priority: 2
description: "Fix integration tests failing on PR branches"
created_at: "2025-07-28T14:25:09+09:00"
started_at: null
closed_at: null
related:
    - parent:250728-001606-fix-worktree-test-macos
---

# Ticket Overview

Integration tests are failing in CI when run on PR branches with the error "Invalid branch for starting ticket". The tests expect to run on the main branch but GitHub Actions runs them on the PR branch.

## Problem

Multiple integration tests fail with:
```
Error: Invalid branch for starting ticket
```

This affects:
- TestCleanupTicketWithForceFlag
- TestDirectoryAutoCreation  
- TestDirectoryCreationWithWorktrees
- TestCompleteWorkflow
- TestRestoreWorkflow
- TestStartTicket_WorktreeCreatedAfterCommit

## Tasks
- [ ] Investigate why tests require main branch
- [ ] Consider options:
  - [ ] Make tests branch-agnostic
  - [ ] Skip certain tests when not on main branch
  - [ ] Use a different approach for PR testing
- [ ] Implement chosen solution
- [ ] Run `make test-integration` to verify tests pass
- [ ] Test on both main branch and PR branches
- [ ] Update CI workflow if needed

## Notes

This issue blocks all PRs from passing CI. The tests work fine locally on the main branch but fail in GitHub Actions PR context.