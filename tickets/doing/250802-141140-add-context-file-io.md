---
priority: 2
description: ""
created_at: "2025-08-02T14:11:40+09:00"
started_at: "2025-08-02T14:56:01+09:00"
closed_at: null
related:
    - parent:250801-003206-add-context-support
---

# Add Context Support to File I/O Operations

Implement context-aware file I/O operations to enable cancellation and timeouts for file reading and writing operations.

## Context

The current implementation has context support for git operations and other external commands, but file I/O operations still use standard Go functions without context support. This ticket adds context-aware wrappers for file operations.

## Tasks

- [x] Create context-aware file I/O helpers in internal/ticket/manager.go
- [x] Add context checks before expensive file operations
- [x] Update ReadContent method to support context cancellation
- [x] Update WriteContent method to support context cancellation
- [x] Update file operations in ticket creation/update
- [x] Add proper error handling for context cancellation during I/O
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation if necessary

## Implementation Notes

Since Go's standard library doesn't have built-in context support for file I/O, we'll need to:
1. Check context before starting operations
2. For large files, read/write in chunks with context checks between chunks
3. Consider using goroutines with select for true cancellation support

## Dependencies

- Requires completion of parent ticket: 250801-003206-add-context-support
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Notes

### Implementation Summary

1. **Created context-aware helper functions**:
   - `readFileWithContext()`: Reads files with context support, using chunked reading for large files (>1MB)
   - `writeFileWithContext()`: Writes files with context support, using chunked writing for large files (>1MB)

2. **Updated file operations**:
   - Modified `Create()` to use `writeFileWithContext()` when creating new ticket files
   - Modified `Update()` to use `writeFileWithContext()` when updating ticket files
   - Modified `loadTicket()` to use `readFileWithContext()` when reading ticket files

3. **Context checks added**:
   - All file I/O operations now check context before starting
   - Large file operations check context between chunks (64KB chunks)
   - Proper error messages returned for cancelled operations

4. **Comprehensive tests added**:
   - Tests for successful read/write operations
   - Tests for context cancellation before operations
   - Tests for context cancellation during large file operations
   - Tests for existing Manager methods with cancelled contexts
   - Test for creating tickets with large templates (>1MB)

All tests pass and linters are clean.