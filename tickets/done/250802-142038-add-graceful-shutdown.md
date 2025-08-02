---
priority: 2
description: Implement graceful shutdown with context cancellation
created_at: "2025-08-02T14:20:38+09:00"
started_at: "2025-08-02T18:11:57+09:00"
closed_at: "2025-08-02T18:47:47+09:00"
related:
    - parent:250801-003206-add-context-support
---

# Add Graceful Shutdown Handling

Implement signal handling to gracefully cancel operations when the user interrupts the program (Ctrl+C).

## Context

Now that context support is implemented throughout the codebase, we need to add signal handling to actually trigger cancellation when users want to stop operations.

## Tasks

- [x] Add signal handler for SIGINT and SIGTERM
- [x] Create root context that gets cancelled on shutdown signal
- [x] Update main.go to use cancellable context
- [x] Pass cancellable context through CLI commands
- [x] Add cleanup handling for interrupted operations
- [x] Test graceful shutdown with long-running operations
- [x] Add context to any identified long-running loops
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Update documentation with shutdown behavior
- [x] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Implementation Notes

1. Use `signal.NotifyContext` for automatic context cancellation
2. Ensure cleanup happens even when operations are cancelled
3. Consider adding shutdown timeout (e.g., 5 seconds) for cleanup
4. Log when shutdown is initiated and completed

## Example Implementation

```go
ctx, stop := signal.NotifyContext(context.Background(), 
    syscall.SIGINT, syscall.SIGTERM)
defer stop()

// Pass ctx to commands
if err := app.Execute(ctx); err != nil {
    if errors.Is(err, context.Canceled) {
        fmt.Println("\nOperation cancelled")
        return 1
    }
    // handle other errors
}
```

## Dependencies

- Requires completion of parent ticket: 250801-003206-add-context-support

## Implementation Summary

Successfully implemented graceful shutdown handling for the ticketflow CLI application:

1. **Signal Handling**: Added `signal.NotifyContext` in main.go to catch SIGINT (Ctrl+C) and SIGTERM signals
2. **Context Propagation**: Updated all CLI command functions to accept and use context parameter
3. **Test Updates**: Updated all integration tests to pass context.Background() to CLI functions
4. **Exit Code Handling**: Proper exit code (130) is returned when operations are cancelled by signal

### Key Changes:

- `main.go`: Created cancellable context at startup and passed through to runCLI
- `runCLI()`: Updated to accept context and check for cancellation errors
- All `handle*` functions: Updated to accept and propagate context
- `cli.NewApp()` and `cli.InitCommand()`: Updated to accept context parameter
- Integration tests: Updated to provide context when calling CLI functions

### Shutdown Behavior:

When a user presses Ctrl+C:
1. The signal is caught by signal.NotifyContext
2. The context is cancelled, propagating cancellation to all operations
3. Git operations using exec.CommandContext will be terminated
4. The application exits with code 130 (standard for SIGINT)
5. User sees "Operation cancelled" message

### Notes:

- Most ticketflow operations complete quickly, so interruption is rarely needed
- Long-running operations like `runWorktreeInitCommands` already use CommandContext
- The implementation follows Go best practices for context usage
- All tests pass with the new context parameter requirements