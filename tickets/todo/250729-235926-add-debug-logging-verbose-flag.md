---
priority: 2
description: "Add --verbose or --debug flag to enable debug logging for troubleshooting"
created_at: "2025-07-29T23:59:26+09:00"
started_at: null
closed_at: null
---

# Ticket Overview

Add a `--verbose` or `--debug` flag to ticketflow commands to enable debug logging output. This will help users and developers troubleshoot issues by providing detailed information about what the application is doing internally.

## Background

During the investigation of the stale branch detection issue (PR #18), it was identified that having debug logging would have made the troubleshooting process easier. Currently, there's no way to see detailed information about internal operations like:
- Which branches are being evaluated during cleanup
- Why certain branches are or aren't considered stale
- Git command executions and their results
- File operations and ticket loading details

## Requirements

1. **Global Flag**: Add a `--verbose` or `--debug` flag that can be used with any ticketflow command
2. **Log Levels**: Implement at least two log levels:
   - Normal: Default, minimal output
   - Debug/Verbose: Detailed operational information
3. **Consistent Format**: Debug logs should have a consistent format with timestamps and context
4. **Performance**: Debug logging should have minimal performance impact when disabled

## Implementation Approach

1. **Logging Library**: Evaluate whether to use Go's standard `log` package or a more feature-rich library like `logrus` or `zap`
2. **Global Configuration**: Make the debug flag available globally across all commands
3. **Strategic Placement**: Add debug logs at key decision points, especially:
   - Branch cleanup logic (showing which branches are evaluated and why)
   - Git operations (commands being executed)
   - Ticket operations (loading, filtering, status changes)
   - Worktree operations
   - Error conditions and edge cases

## Tasks
- [ ] Research and choose appropriate logging approach (stdlib vs third-party)
- [ ] Add `--verbose` or `--debug` flag to CLI argument parsing
- [ ] Create logging utility/wrapper with appropriate log levels
- [ ] Add debug logs to cleanup operations (priority for stale branch detection)
- [ ] Add debug logs to git operations
- [ ] Add debug logs to ticket manager operations
- [ ] Add debug logs to worktree operations
- [ ] Test debug output with various commands
- [ ] Ensure debug output goes to stderr (not stdout) to not interfere with JSON output
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation with examples of using debug flag
- [ ] Update README.md with troubleshooting section mentioning debug flag
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Example Usage

```bash
# Debug cleanup operations
ticketflow cleanup --verbose

# Debug ticket creation
ticketflow new "my-feature" --debug

# Debug with JSON output (debug to stderr, JSON to stdout)
ticketflow list --format json --verbose
```

## Notes

- Consider whether both `--verbose` and `--debug` should be supported with different verbosity levels, or just one flag
- Debug output should go to stderr to not interfere with structured output (JSON)
- Consider adding log categories/tags to filter specific types of debug output
- Performance impact should be minimal when debug logging is disabled
- Consider adding a config file option to enable debug logging by default for development