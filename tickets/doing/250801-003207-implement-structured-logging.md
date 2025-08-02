---
priority: 2
description: Replace fmt.Printf with structured logging using log/slog for better observability
created_at: "2025-08-01T00:32:07+09:00"
started_at: "2025-08-02T19:30:13+09:00"
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
- [x] Create logging package in `internal/log/`
- [x] Configure slog with appropriate handlers
- [x] Add log level configuration
- [x] Create logger factory functions

### Replace Printf Statements
- [ ] Replace all `fmt.Printf` in `internal/ticket/manager.go` (no printf found)
- [ ] Replace all `fmt.Printf` in `internal/git/git.go` (no printf found)
- [x] Replace all `fmt.Printf` in `internal/cli/commands.go`
- [ ] Replace all `fmt.Printf` in `internal/ui/` (where appropriate)

### Add Contextual Logging
- [x] Add ticket ID to relevant log entries
- [x] Add operation names to log entries
- [x] ~~Add timing information for operations~~ (not needed for now)
- [x] Add error details to error logs

### Configuration
- [x] Add log level configuration to .ticketflow.yaml
- [x] Add log format configuration (text/json)
- [x] Add log output configuration (file/stderr)
- [x] Document logging configuration options

### Quality Assurance
- [x] Ensure user-facing output is not affected
- [x] Verify log levels work correctly
- [x] Test JSON output format
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Update documentation
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

## UPDATE: Simplified Approach Based on Feedback

After initial implementation, the approach has been simplified based on user feedback:
- **No logging by default** - Silent operation unless explicitly enabled
- **No YAML configuration** - Remove logging from config file
- **CLI flags only** - Add command-line options for logging control

### Revised Tasks

- [x] Remove logging configuration from YAML/config system
- [x] Implement no-op logger as default (silent operation)
- [x] Add CLI flags: --log-level, --log-format, --log-output
- [x] Update all logger initialization to use CLI flags
- [x] Remove logging documentation from README
- [x] Clean up example files

## Implementation Summary (Initial Approach)

### What Was Done

1. **Created logging infrastructure** in `internal/log/`:
   - `logger.go`: Main logger implementation with configuration support
   - `global.go`: Global logger instance and convenience functions
   - `logger_test.go`: Comprehensive test coverage
   - Supports JSON/text formats, configurable levels, and output destinations

2. **Integrated with configuration system**:
   - Added `logging` field to config struct with level, format, and output options
   - Logger is reconfigured when config is loaded in both TUI and CLI modes
   - Created example configuration file showing logging options

3. **Added structured logging to key operations**:
   - `main.go`: Application startup and mode selection
   - `commands.go`: Ticket operations (new, start, close)
   - `cleanup.go`: Worktree and branch cleanup operations
   - `migrate_dates.go`: Date migration process

4. **Maintained user experience**:
   - All user-facing output remains unchanged (still using fmt.Printf)
   - Logging is added alongside user output for observability
   - Debug logs are hidden by default (info level)

### Key Insights

1. **Separation of concerns**: It's critical to distinguish between user-facing output and diagnostic logging. User output should remain clean and focused, while logging provides detailed operational information.

2. **Context propagation**: Using methods like `WithTicket()`, `WithOperation()`, and `WithError()` makes it easy to add consistent contextual information to log entries.

3. **Minimal disruption**: The implementation was careful not to change any user-visible behavior while adding comprehensive logging throughout the application.

4. **Configuration flexibility**: Supporting multiple output formats (text/JSON) and destinations (stderr/stdout/file) allows users to integrate with their preferred log aggregation tools.

### Areas Not Completed

1. **UI logging**: The TUI components in `internal/ui/` were not modified as they use the Bubble Tea framework which has its own output handling. No fmt.Printf statements were found that needed replacement.

2. **Timing information**: Not implemented as it's not needed for now.

### Next Steps

None - implementation is complete with the simplified CLI-flag based approach.

### Additional Work Completed

1. **Documentation Updates**:
   - Added logging configuration to the main README.md configuration section
   - Created a dedicated "Logging" section with detailed examples
   - Documented integration with observability tools

2. **Examples and Demos**:
   - Created examples to demonstrate logging (later removed as part of simplification)

3. **Code Quality**:
   - All fmt.Printf statements in CLI packages have been augmented with structured logging
   - User-facing output remains unchanged
   - No logging needed in UI package (uses Bubble Tea framework)
   - All tests pass, code is formatted and linted

## Simplified Implementation Summary

Based on user feedback, the implementation has been revised to be much simpler:

1. **Removed YAML Configuration**:
   - Deleted logging configuration from Config struct
   - Removed all logging-related YAML parsing
   - No logging configuration in .ticketflow.yaml files

2. **Implemented Silent Default**:
   - Created NewNoOp() function that returns a logger with io.Discard output
   - Global logger defaults to no-op logger (silent operation)
   - No logs are generated unless explicitly enabled via CLI flags

3. **Added CLI Flag Support**:
   - Created internal/cli/logging.go with LoggingOptions struct
   - Added --log-level, --log-format, --log-output flags to all commands
   - ConfigureLogging() function sets up logging only when flags are provided

4. **Updated Documentation**:
   - Removed logging section from README
   - Updated help text to show logging flags
   - Cleaned up example files

### Key Benefits of Simplified Approach:
- Zero noise by default - perfect for human and AI usage
- Simple opt-in via CLI flags when debugging is needed
- No configuration file clutter
- Easier to use and understand

## Final Status

All implementation work is complete. The structured logging feature has been successfully implemented with a simplified CLI-flag based approach that:

1. Maintains silent operation by default (no logs unless explicitly enabled)
2. Provides full structured logging capabilities via CLI flags
3. Keeps configuration files clean and focused
4. Preserves all user-facing output exactly as before

Ready for developer review and approval.