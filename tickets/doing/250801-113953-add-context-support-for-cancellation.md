---
priority: 3
description: Add context.Context support for cancellation of long-running operations
created_at: "2025-08-01T11:39:53+09:00"
started_at: "2025-08-04T22:33:53+09:00"
closed_at: null
related:
    - parent:250801-003010-decompose-large-functions
---

# Add Context Support for Cancellation

Add `context.Context` support to allow cancellation of long-running operations, improving responsiveness and resource management.

## Context

Following the function decomposition work, the golang-pro agent suggested adding context.Context support for better control over long-running operations. This is particularly important for operations that involve git commands or file I/O which could potentially hang or take longer than expected.

## Tasks

- [ ] Update CLI command signatures to accept context.Context
  ```go
  func (app *App) StartTicket(ctx context.Context, ticketID string) error {
      // Allow cancellation of long-running operations
  }
  ```
- [ ] Add context support to Git interface methods
- [ ] Implement context cancellation in long-running git operations
- [ ] Add timeout contexts for operations with expected durations
- [ ] Update UI commands to propagate context from Bubble Tea
- [ ] Add context cancellation tests
- [ ] Update documentation with context usage examples
- [ ] Run `make test` to ensure all tests pass
- [ ] Get developer approval before closing

## Implementation Guidelines

1. **CLI Commands**: Pass context from main() down through command execution
2. **Git Operations**: Use exec.CommandContext for git commands
3. **File Operations**: Check context.Done() in loops
4. **UI Operations**: Use Bubble Tea's built-in context support
5. **Timeouts**: Set reasonable timeouts for different operation types

## Example Implementation

```go
// Git interface update
type GitClient interface {
    AddWorktree(ctx context.Context, path, branch string) error
    // ... other methods
}

// Implementation
func (g *Git) AddWorktree(ctx context.Context, path, branch string) error {
    cmd := exec.CommandContext(ctx, "git", "worktree", "add", path, branch)
    return cmd.Run()
}
```

## Acceptance Criteria

- All long-running operations support context cancellation
- Operations gracefully handle context cancellation
- No goroutine leaks when operations are cancelled
- Tests verify cancellation behavior
- Performance is not negatively impacted

## Notes

Suggested by golang-pro agent during code review. This improvement will make the application more robust and responsive, especially in CI/CD environments where timeouts are important.