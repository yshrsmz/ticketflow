---
priority: 2
description: Implement configurable timeouts for operations
created_at: "2025-08-02T14:12:19+09:00"
started_at: "2025-08-02T17:20:34+09:00"
closed_at: null
related:
    - parent:250801-003206-add-context-support
---

# Add Timeout Configuration Support

Implement configurable timeouts for operations to prevent commands from running indefinitely.

## Context

Currently all operations use context.Background() without any timeout. This ticket adds configuration options to set default timeouts for different types of operations, improving reliability and user experience.

## Tasks

- [x] Add timeout configuration to config.yaml structure
- [x] Define timeout fields for different operation types (git, file I/O, etc.)
- [x] Update config package with timeout parsing and validation
- [x] Modify CLI commands to use timeout from config
- [ ] Add command-line flags to override timeout values (deferred - not needed for initial implementation)
- [x] Implement graceful timeout handling with proper error messages
- [x] Add default timeout values (e.g., 30s for git operations)
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Update documentation with timeout configuration examples
- [x] Update README.md with timeout configuration
- [x] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Configuration Design

```yaml
# .ticketflow.yaml
timeouts:
  git: 30           # Timeout for git operations in seconds
  init_commands: 60 # Timeout for worktree init commands in seconds
```

Note: Simplified the design to focus on the two most critical timeout scenarios. File I/O operations remain using context cancellation without explicit timeouts.

## Implementation Notes

1. Use `context.WithTimeout` instead of `context.Background()`
2. Allow per-command timeout overrides via CLI flags
3. Ensure timeout errors are clearly reported to users
4. Consider different timeouts for different git operations (clone vs status)

## Dependencies

- Requires completion of parent ticket: 250801-003206-add-context-support

## Implementation Insights

### Key Decisions Made

1. **Simplified Configuration**: Instead of implementing timeouts for all operation types, focused on the two most critical areas:
   - Git operations (30s default)
   - Worktree init commands (60s default)

2. **Backward Compatibility**: The implementation maintains full backward compatibility:
   - If timeouts are not configured, sensible defaults are used
   - Existing configurations continue to work without modification
   - Zero/negative values fall back to defaults

3. **Context-Aware Design**: The timeout implementation respects existing context deadlines:
   - If a context already has a deadline, it's not overridden
   - This allows for proper timeout composition in nested operations

4. **Error Messaging**: Implemented specific error messages for timeout scenarios:
   - Git operations report "operation timed out after Xs"
   - Init commands show which specific command timed out

### Technical Implementation Details

1. **Constants for Defaults**: Defined constants in `config/constants.go` to avoid magic numbers
2. **Factory Pattern**: Added `NewWithTimeout()` alongside existing `New()` for Git struct
3. **Timeout Preservation**: Worktree operations preserve the parent Git instance's timeout
4. **Comprehensive Testing**: Added tests for timeout functionality including edge cases

### Code Quality Improvements from Review

1. Replaced hardcoded timeout values with named constants
2. Enhanced error messages to clearly indicate timeout vs other failures
3. Updated project configuration to demonstrate the feature

### Future Enhancements (Not Implemented)

1. **Per-Operation Timeouts**: Could allow different timeouts for fetch, clone, etc.
2. **Command-Line Overrides**: Could add flags like `--timeout-git=45`
3. **Retry Logic**: Could implement exponential backoff for timeout failures
4. **Timeout Metrics**: Could log warnings when operations approach timeout

### Lessons Learned

1. **Start Simple**: Beginning with just two timeout types was the right approach
2. **Test Edge Cases**: Important to test both timeout and successful scenarios
3. **Clear Error Messages**: Users need to know when timeouts occur vs other failures
4. **Review Feedback**: The golang-pro review provided valuable improvements