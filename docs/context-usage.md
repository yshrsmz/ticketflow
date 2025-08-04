# Context Usage in TicketFlow

This document describes how `context.Context` is used throughout the TicketFlow codebase for cancellation and timeout support.

## Overview

TicketFlow uses Go's `context.Context` to provide cancellation support for long-running operations. This ensures the application remains responsive and can gracefully handle interruptions (e.g., Ctrl+C).

## Architecture

### Signal Handling in main()

The application sets up signal handling at startup to catch interrupts:

```go
// main.go
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()
```

**Note**: TUI mode uses Bubble Tea's built-in signal handling instead of the cancellable context.

## Context Usage Patterns

### 1. Git Operations

All git operations use `exec.CommandContext` for automatic process termination on context cancellation:

```go
// internal/git/git.go
func (g *Git) Exec(ctx context.Context, args ...string) (string, error) {
    cmd := exec.CommandContext(ctx, GitCmd, args...)
    // ... configure timeout
    output, err := cmd.Output()
    // ... handle errors
}
```

### 2. Configuration Loading/Saving

Config operations support context cancellation with atomic writes:

```go
// internal/config/config.go

// Load with context support
cfg, err := config.LoadWithContext(ctx, projectRoot)

// Save with context support  
err := cfg.SaveWithContext(ctx, configPath)
```

Features:
- File size validation (1MB limit for config files)
- Atomic writes using temporary files
- Context checks before and during I/O operations

### 3. Ticket Manager File Operations

The ticket manager uses sophisticated context-aware file I/O:

```go
// internal/ticket/manager.go

// Chunked reading for large files
func readFileWithContext(ctx context.Context, path string) ([]byte, error) {
    // Check context before starting
    if err := ctx.Err(); err != nil {
        return nil, fmt.Errorf("operation cancelled: %w", err)
    }
    
    // For large files, read in chunks with context checks
    for {
        if err := ctx.Err(); err != nil {
            return nil, fmt.Errorf("operation cancelled during read: %w", err)
        }
        // Read chunk...
    }
}
```

### 4. CLI Commands

All CLI commands accept context and propagate it down:

```go
// internal/cli/commands.go
func (app *App) StartTicket(ctx context.Context, ticketID string) error {
    // Context is passed to all sub-operations
    ticket, err := app.Manager.Get(ctx, ticketID)
    // ...
    err = app.Git.CheckoutBranch(ctx, branchName)
    // ...
}
```

## Timeout Configuration

Timeouts are configurable via `.ticketflow.yaml`:

```yaml
timeouts:
  git: 30          # Git operations timeout (seconds)
  init_commands: 60 # Worktree init commands timeout (seconds)
```

Usage in code:
```go
// Get configured timeout
timeout := cfg.GetGitTimeout() // Returns time.Duration

// Create timeout context
ctx, cancel := context.WithTimeout(ctx, timeout)
defer cancel()
```

## Best Practices

### 1. Always Check Context Early

Check context at the beginning of operations:

```go
func SomeOperation(ctx context.Context) error {
    if err := ctx.Err(); err != nil {
        return fmt.Errorf("operation cancelled: %w", err)
    }
    // ... proceed with operation
}
```

### 2. Check Context in Loops

For operations that process data in loops:

```go
for _, item := range items {
    // Check context in each iteration
    if err := ctx.Err(); err != nil {
        return fmt.Errorf("operation cancelled: %w", err)
    }
    // Process item...
}
```

### 3. Use Context-Aware APIs

Always prefer context-aware versions:

```go
// Good
cmd := exec.CommandContext(ctx, "git", "status")

// Avoid
cmd := exec.Command("git", "status")
```

### 4. Propagate Context

Always pass context down the call chain:

```go
func HighLevel(ctx context.Context) error {
    // Pass context to lower-level functions
    return LowLevel(ctx, data)
}
```

## Testing Context Cancellation

Example test for context cancellation:

```go
func TestOperationWithCancelledContext(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    cancel() // Cancel immediately
    
    err := SomeOperation(ctx)
    require.Error(t, err)
    assert.Contains(t, err.Error(), "operation cancelled")
}
```

## Error Handling

Context cancellation errors should be handled gracefully:

```go
if err := operation(ctx); err != nil {
    // Check if error is due to context cancellation
    if ctx.Err() != nil {
        // Handle cancellation (e.g., cleanup, user message)
        return fmt.Errorf("operation cancelled by user")
    }
    // Handle other errors
    return fmt.Errorf("operation failed: %w", err)
}
```

## Migration Guide

When adding context support to existing functions:

1. Add `ctx context.Context` as the first parameter
2. Create a wrapper for backward compatibility:
   ```go
   func OldFunction(param string) error {
       return NewFunction(context.Background(), param)
   }
   ```
3. Update callers gradually to use context-aware version
4. Add context checks at appropriate points
5. Add tests for context cancellation

## Summary

Context support in TicketFlow ensures:
- Responsive cancellation of long-running operations
- Proper cleanup on interruption
- Configurable timeouts for different operation types
- Graceful shutdown handling

All new code should follow these patterns to maintain consistent context usage throughout the application.