---
priority: 3
description: Handle case when branch exists but points to different commit
created_at: "2025-08-03T12:14:50+09:00"
started_at: "2025-08-05T13:29:41+09:00"
closed_at: null
related:
    - parent:250726-183403-fix-branch-already-exist-on-start
---

# Handle Diverged Branch

## Overview
When starting a ticket, if a branch already exists but points to a different commit than expected (e.g., not at the default branch HEAD), we need to provide clear options to the user.

## Tasks
- [ ] Add method to check if branch diverged from expected base
- [ ] Implement interactive prompt for user choice
- [ ] Add option to use existing branch
- [ ] Add option to delete and recreate branch
- [ ] Add tests for diverged branch scenarios

## Technical Details
- Compare branch HEAD with default branch HEAD using `git rev-parse`
- Show clear information about the divergence (commits ahead/behind)
- Implement user choice handling with proper error recovery

## Acceptance Criteria
- User gets clear options when branch is diverged
- Each option works correctly (use existing, recreate, cancel)
- Tests cover all scenarios