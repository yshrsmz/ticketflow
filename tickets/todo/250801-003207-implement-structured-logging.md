---
priority: 2
description: "Replace fmt.Printf with structured logging using log/slog for better observability"
created_at: "2025-08-01T00:32:07+09:00"
started_at: null
closed_at: null
---

# Implement Structured Logging

Replace all `fmt.Printf` statements with structured logging using Go's `log/slog` package to improve observability, debugging, and log analysis.

## Context

Current logging approach has several limitations:
- Using `fmt.Printf` for all output mixing logs with user output
- No log levels (debug, info, warn, error)
- No structured data (everything is plain text)
- Difficult to filter or parse logs
- No consistent format across the application

Structured logging provides:
- Machine-readable log formats (JSON)
- Consistent log structure
- Easy filtering by level or attributes
- Better integration with log aggregation tools
- Contextual information with each log entry

## Tasks

### Setup Logging Infrastructure
- [ ] Create logging package in `internal/log/`
- [ ] Configure slog with appropriate handlers
- [ ] Add log level configuration
- [ ] Create logger factory functions

### Replace Printf Statements
- [ ] Replace all `fmt.Printf` in `internal/ticket/manager.go`
- [ ] Replace all `fmt.Printf` in `internal/git/git.go`
- [ ] Replace all `fmt.Printf` in `internal/cli/commands.go`
- [ ] Replace all `fmt.Printf` in `internal/ui/` (where appropriate)

### Add Contextual Logging
- [ ] Add ticket ID to relevant log entries
- [ ] Add operation names to log entries
- [ ] Add timing information for operations
- [ ] Add error details to error logs

### Configuration
- [ ] Add log level configuration to .ticketflow.yaml
- [ ] Add log format configuration (text/json)
- [ ] Add log output configuration (file/stderr)
- [ ] Document logging configuration options

### Quality Assurance
- [ ] Ensure user-facing output is not affected
- [ ] Verify log levels work correctly
- [ ] Test JSON output format
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation
- [ ] Get developer approval before closing

## Implementation Guidelines

### Logger Setup
```go
// internal/log/logger.go
func New(level slog.Level) *slog.Logger {
    opts := &slog.HandlerOptions{
        Level: level,
    }
    
    handler := slog.NewJSONHandler(os.Stderr, opts)
    return slog.New(handler)
}
```

### Logging Pattern
```go
// Before
fmt.Printf("Starting ticket: %s\n", ticketID)

// After
logger.Info("starting ticket",
    slog.String("ticket_id", ticketID),
    slog.String("operation", "start"),
)
```

### Contextual Logging
```go
// Add context to logger
logger := baseLogger.With(
    slog.String("ticket_id", ticket.ID),
    slog.String("status", ticket.Status),
)

// Use throughout operation
logger.Debug("creating worktree")
logger.Info("worktree created", slog.String("path", path))
```

### Log Levels
- **Debug**: Detailed information for debugging
- **Info**: General informational messages
- **Warn**: Warning conditions that might need attention
- **Error**: Error conditions that need immediate attention

## Notes

Be careful to distinguish between:
- User output (should remain as-is, not logged)
- Debug/diagnostic output (should use structured logging)
- Error messages (should be both user-facing and logged)

Consider using logger as a dependency injection to make testing easier.