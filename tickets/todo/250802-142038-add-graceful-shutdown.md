---
priority: 2
description: "Implement graceful shutdown with context cancellation"
created_at: "2025-08-02T14:20:38+09:00"
started_at: null
closed_at: null
related:
    - parent:250801-003206-add-context-support
---

# Add Graceful Shutdown Handling

Implement signal handling to gracefully cancel operations when the user interrupts the program (Ctrl+C).

## Context

Now that context support is implemented throughout the codebase, we need to add signal handling to actually trigger cancellation when users want to stop operations.

## Tasks

- [ ] Add signal handler for SIGINT and SIGTERM
- [ ] Create root context that gets cancelled on shutdown signal
- [ ] Update main.go to use cancellable context
- [ ] Pass cancellable context through CLI commands
- [ ] Add cleanup handling for interrupted operations
- [ ] Test graceful shutdown with long-running operations
- [ ] Add context to any identified long-running loops
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation with shutdown behavior
- [ ] Update the ticket with insights from resolving this ticket
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