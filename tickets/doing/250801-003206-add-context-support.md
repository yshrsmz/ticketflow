---
priority: 2
description: Add context.Context support for cancellation and timeouts in long-running operations
created_at: "2025-08-01T00:32:06+09:00"
started_at: "2025-08-02T11:49:52+09:00"
closed_at: null
---

# Add Context Support

Implement proper context.Context support throughout the codebase to enable cancellation and timeouts for long-running operations, especially git commands and file I/O operations.

## Context

Currently, the application doesn't support cancellation of long-running operations. This can lead to:
- Users unable to cancel stuck operations
- Resource waste from abandoned operations
- Poor user experience with unresponsive commands
- No timeout control for external commands

Adding context support will:
- Enable graceful cancellation of operations
- Allow timeout configuration
- Improve resource management
- Follow Go best practices for concurrent operations

## Tasks

### API Updates
- [x] Add context parameter to all public methods in `internal/ticket/manager.go`
- [x] Add context parameter to all methods in `internal/git/git.go`
- [x] Update CLI command handlers to accept context
- [x] Update TUI operations to use context

### Implementation
- [x] Implement context cancellation for git operations
- [ ] Add context support to file I/O operations
- [ ] Implement timeout handling for external commands
- [ ] Add graceful shutdown handling

### Specific Updates
- [x] Update `exec.Command` calls to use `CommandContext` (mostly done, see review)
- [ ] Add context to long-running loops
- [x] Implement proper context propagation
- [ ] Add timeout configuration options

### Quality Assurance
- [ ] Add tests for cancellation behavior
- [ ] Test timeout functionality
- [ ] Verify graceful shutdown works correctly
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation if necessary
- [ ] Get developer approval before closing

### Code Review Findings (golang-pro)
- [x] Comprehensive interface updates with context as first parameter
- [x] Proper use of exec.CommandContext for main git operations
- [x] Good context error checking at operation boundaries
- [x] Consistent patterns across the codebase
- [ ] Fix utility functions still using exec.Command (IsGitRepo, FindProjectRoot)
- [ ] Update editor and init commands to use exec.CommandContext
- [ ] Add timeout configuration support
- [ ] Add context support to file I/O operations

## Implementation Guidelines

### Context Parameter Pattern
```go
// Before
func (m *Manager) StartTicket(id string) error {
    // ...
}

// After
func (m *Manager) StartTicket(ctx context.Context, id string) error {
    // Check context at operation boundaries
    if err := ctx.Err(); err != nil {
        return fmt.Errorf("operation cancelled: %w", err)
    }
    
    // Use context for external commands
    cmd := exec.CommandContext(ctx, "git", "worktree", "add")
    // ...
}
```

### Timeout Configuration
```go
// Add to config
type Config struct {
    // ...
    CommandTimeout time.Duration `yaml:"commandTimeout"`
}

// Use in commands
ctx, cancel := context.WithTimeout(ctx, cfg.CommandTimeout)
defer cancel()
```

## Progress Summary

### Completed (2025-08-02)
- Added context.Context as first parameter to all TicketManager and GitClient interface methods
- Updated all implementations to use exec.CommandContext for git operations
- Added context error checking at operation boundaries
- Updated all CLI and TUI operations to pass context.Background()
- All tests updated and passing
- Code formatted and linted

### golang-pro Review Grade: B+
The implementation is well-executed with proper patterns and comprehensive coverage. Minor gaps identified:
1. Utility functions (IsGitRepo, FindProjectRoot) still use exec.Command
2. Editor and init commands not using context-aware execution
3. No timeout configuration implemented yet
4. File I/O operations pending context support

## Notes

This is a breaking change that will require updating all method signatures. Consider implementing in phases:
1. Add context to internal implementations
2. Update public APIs with backward compatibility
3. Deprecate old methods
4. Remove deprecated methods in next major version

Ensure all context cancellations are properly handled to avoid resource leaks.