---
priority: 2
description: "Batch git operations for parallel execution and better performance"
created_at: "2025-08-10T00:30:17+09:00"
started_at: null
closed_at: null
---

# Task 3.2: Git Operations Batching

**Duration**: 1 day  
**Complexity**: Medium  
**Phase**: 3 - Performance Optimizations  
**Dependencies**: Task 2.3 (Worker Pool)

Implement concurrent git operations batching for improved performance on parallel queries.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [ ] Analyze current git operations in internal/git/
- [ ] Identify operations that can be batched
- [ ] Use git plumbing commands for performance
- [ ] Implement git for-each-ref instead of multiple branch calls
- [ ] Add concurrent execution for independent queries
- [ ] Create GitBatch struct for operation grouping
- [ ] Implement command coalescing
- [ ] Add unit tests for batched operations
- [ ] Benchmark individual vs batched operations
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation if necessary
- [ ] Update README.md
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Implementation Notes

- Example: git for-each-ref instead of multiple git branch calls
- Use plumbing commands over porcelain for speed
- Batch size limits to avoid command line length issues
- Key files: internal/git/*.go

## Expected Outcomes

- 50% reduction in git command executions
- Faster worktree operations
- Improved performance for status checks