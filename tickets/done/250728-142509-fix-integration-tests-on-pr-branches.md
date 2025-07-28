---
priority: 2
description: Fix integration tests failing on PR branches
created_at: "2025-07-28T14:25:09+09:00"
started_at: "2025-07-28T14:59:41+09:00"
closed_at: "2025-07-28T15:09:24+09:00"
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
- [x] Investigate why tests require main branch
- [x] Consider options:
  - [x] Make tests branch-agnostic
  - [ ] Skip certain tests when not on main branch
  - [ ] Use a different approach for PR testing
- [x] Implement chosen solution
- [x] Run `make test-integration` to verify tests pass
- [x] Test on both main branch and PR branches
- [ ] Update CI workflow if needed

## Solution

The issue was that `ticketflow start` validates that you're on either the default branch (main) or a valid ticket branch. Since PR branches in GitHub Actions (e.g., `pull/123/merge`) are neither, the tests failed.

Fixed by updating the `setupTestRepo` helper function in `test/integration/workflow_test.go` to ensure tests always start from the main branch, regardless of which branch they're executed from.

## Key Insights

1. **Test Isolation**: Integration tests create temporary git repositories for each test case using `setupTestRepo()`. These are completely separate from the actual codebase repository.

2. **Branch Context Mismatch**: When GitHub Actions runs tests on a PR, it checks out a merge commit (e.g., `pull/123/merge`). However, the test repositories inherit some git context, causing them to fail ticketflow's branch validation.

3. **Minimal Fix**: By ensuring test repositories explicitly checkout `main` branch during setup, we maintain test isolation while satisfying ticketflow's branch requirements. No changes to production code were needed.

4. **CI Compatibility**: This approach works regardless of which branch the CI runner is on, making tests truly branch-agnostic without compromising the application's branch validation logic.

## Notes

This issue blocks all PRs from passing CI. The tests work fine locally on the main branch but fail in GitHub Actions PR context.