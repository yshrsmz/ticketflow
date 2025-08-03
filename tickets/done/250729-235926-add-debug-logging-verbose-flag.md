---
priority: 2
description: "Add --verbose or --debug flag to enable debug logging for troubleshooting"
created_at: "2025-07-29T23:59:26+09:00"
started_at: "2025-08-03T14:00:00+09:00"
closed_at: "2025-08-03T14:00:00+09:00"
---

# IMPLEMENTED: Feature Already Available

This feature has been **FULLY IMPLEMENTED** as part of ticket `250801-003207-implement-structured-logging`. The debug logging functionality is now available via the `--log-level debug` flag.

## Available Functionality

The implemented solution provides all requested capabilities:

### ✅ Global Flag Support
- **`--log-level debug`**: Enables debug logging for any ticketflow command
- **`--log-format json|text`**: Choose between structured JSON or human-readable text logs
- **`--log-output stderr|stdout|file`**: Control where logs are written

### ✅ Log Levels Implemented
- **Silent (default)**: No logs unless explicitly enabled
- **Debug**: Detailed operational information for troubleshooting
- **Info**: General informational messages
- **Warn**: Warning conditions
- **Error**: Error conditions

### ✅ Structured Logging with Context
- Consistent format with timestamps and contextual information
- Machine-readable JSON format for log aggregation tools
- Contextual fields like ticket_id, operation, error details

### ✅ Strategic Placement
Debug logs are implemented in all critical areas:
- ✅ Ticket operations (loading, filtering, status changes)
- ✅ Git operations and command executions
- ✅ Worktree operations
- ✅ Cleanup operations (including branch evaluation logic)
- ✅ Error conditions and edge cases

## Usage Examples

All the originally requested functionality is now available:

```bash
# Debug cleanup operations (equivalent to requested --verbose)
ticketflow cleanup --log-level debug

# Debug ticket creation (equivalent to requested --debug)
ticketflow new "my-feature" --log-level debug

# Debug with JSON output (logs to stderr, JSON to stdout)
ticketflow list --format json --log-level debug

# Additional capabilities beyond original requirements:
ticketflow start my-ticket --log-level debug --log-format json
ticketflow close --log-level info --log-output /tmp/ticketflow.log
```

## Implementation Details

- **Logging Library**: Uses Go's standard `log/slog` package for structured logging
- **Performance**: Zero overhead when logging is disabled (default silent operation)
- **Output Separation**: Logs go to stderr by default, preserving JSON output on stdout
- **No Configuration Required**: Pure CLI-flag based approach, no config file clutter

## Related Implementation

This feature was implemented as part of the comprehensive structured logging work in ticket `250801-003207-implement-structured-logging`, which provided:

1. Complete logging infrastructure in `internal/log/`
2. CLI flag integration for all commands
3. Structured logging throughout the codebase
4. Silent-by-default operation
5. Full observability capabilities

The implementation exceeded the original requirements by providing structured logging capabilities and multiple output formats while maintaining the simplicity of CLI-only configuration.

## Status: COMPLETED ✅

This ticket is now **COMPLETE** as all requested functionality has been implemented and is available in the current codebase. Users can enable debug logging using `--log-level debug` flag with any ticketflow command.