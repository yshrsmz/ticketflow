---
priority: 2
description: Refactor cli package to respect JSON format setting in AutoCleanup and related functions
created_at: "2025-08-16T12:37:03+09:00"
started_at: "2025-08-16T18:15:21+09:00"
closed_at: null
related:
    - parent:250815-175624-test-coverage-maintenance-commands
---

# Improve JSON Output Separation in CLI Package

## Problem
Multiple functions in the `internal/cli` package output status messages directly to stdout using `fmt.Printf` and `fmt.Println`, regardless of the output format setting. This causes mixed text/JSON output when JSON format is requested, breaking JSON parsing for AI tools and automation.

## Current Behavior
When running commands with `--format json`, the output contains both text status messages and JSON:
```
Starting auto-cleanup...

Cleaning orphaned worktrees...
  Cleaned 0 orphaned worktree(s)

Cleaning stale branches...
  Cleaned 0 stale branch(es)
Auto-cleanup completed.
{"success": true, "result": {...}}
```

## Expected Behavior
When JSON format is specified, only valid JSON should be output to stdout. Status messages should be suppressed in JSON mode (output only in text mode).

## Affected Functions

### In `internal/cli/cleanup.go`:
- `AutoCleanup()` - Line 30, 41, 56, 64
- `cleanOrphanedWorktrees()` - Lines 77, 116, 121, 132
- `cleanStaleBranches()` - Lines 140, 175, 181, 193
- `CleanupStats()` - Lines 199-253

### In `internal/cli/commands.go`:
- `InitTicketSystem()` - Lines 159, 194-196
- `createWorktree()` - Lines 1347, 1391
- `runInitCommands()` - Lines 1479, 1491, 1514

### In `internal/cli/prompt.go`:
- `SelectOption()` - Lines 45, 58-59, 66, 68, 72
- `GetConfirmation()` - Lines 112, 126

### In `internal/cli/commands/` (for reference, not changing):
- `help.go` - Help text is always text mode
- `version.go` - Version info is always text mode
- `worktree.go` - Help text
- `restore.go` - Success message
- `cleanup.go` - Summary messages (lines 149, 197-204)

## Implementation Approach
Use Strategy pattern with two separate interfaces to cleanly separate concerns:

### StatusWriter Interface
- Handles progress/status messages during execution
- Has two implementations: `textStatusWriter` (prints) and `nullStatusWriter` (no-op for JSON)
- Replaces all `fmt.Printf/Println` calls with `app.Status.Printf/Println`

### OutputWriter Interface  
- Handles final structured data output
- Has two implementations: `jsonOutputWriter` and `textOutputWriter`
- Each implementation knows how to format data appropriately
- No format checking needed in business logic

```go
// StatusWriter for progress messages
type StatusWriter interface {
    Printf(format string, args ...interface{})
    Println(args ...interface{})
}

// OutputWriter for structured results
type OutputWriter interface {
    PrintResult(data interface{}) error
}
```

This approach:
- Clean separation of concerns (status vs data output)
- No repeated format checks throughout code
- Strategy pattern handles format internally
- Easy to test with mock implementations
- Consistent architecture across the codebase

## Tasks
- [x] Analyze current output patterns and verify all affected functions
- [x] Create StatusWriter interface and implementations (text and null)
- [x] Create new OutputWriter interface and implementations (json and text)
- [x] Update App struct to use new interfaces
- [x] Refactor cleanup.go to use app.Status instead of fmt
- [x] Refactor commands.go (InitTicketSystem, createWorktree, runInitCommands)
- [x] Refactor prompt.go functions to use app.Status
- [x] Update command-level output in cleanup.go command
- [ ] Remove old OutputWriter Printf/Println methods (deprecated, kept for compatibility)
- [x] Update integration tests to verify clean JSON output
- [x] Run `make test` to verify all tests pass
- [x] Run `make vet`, `make fmt` and `make lint`
- [ ] Update CLAUDE.md to document new architecture
- [ ] Get developer approval before closing

## Notes
- This issue affects AI tool integration, which requires clean JSON parsing
- The infrastructure (OutputWriter with GetFormat()) already exists
- Command-level functions in `internal/cli/commands/` that always output help text can remain as-is
- Focus on functions that mix status messages with structured data output

## Implementation Details

### Solution Implemented
Successfully refactored the CLI package using the Strategy pattern with two separate interfaces:

1. **StatusWriter Interface**: Manages progress/status messages during execution
   - `textStatusWriter`: Outputs messages to stdout (text mode)
   - `nullStatusWriter`: No-op implementation (JSON mode)
   
2. **ResultWriter Interface**: Handles final structured data output  
   - `jsonResultWriter`: Outputs JSON to stdout
   - `textResultWriter`: Formats data as human-readable text

### Key Changes Made

1. **New Files Created**:
   - `internal/cli/status_writer.go`: StatusWriter interface and implementations
   - `internal/cli/output_writer.go`: Renamed from output.go, now contains ResultWriter

2. **Refactored Files**:
   - `internal/cli/cleanup.go`: All fmt.Printf/Println replaced with app.StatusWriter
   - `internal/cli/commands.go`: Updated to use StatusWriter for all status messages
   - `internal/cli/prompt.go`: Refactored to use StatusWriter for prompts
   - `internal/cli/commands/cleanup.go`: Sets StatusWriter based on format
   - `internal/cli/commands/status.go`: Sets StatusWriter based on format
   - `internal/cli/commands/worktree_list.go`: Sets StatusWriter based on format

3. **Test Fixes**:
   - Added StatusWriter initialization to test fixtures
   - Added nil checks in cleanup functions for test compatibility
   - All tests now passing with clean separation of output

### Benefits Achieved
- Clean JSON output when `--format json` is specified
- No mixed text/JSON output breaking AI tool parsing
- Clear separation of concerns between status messages and data output
- Consistent architecture using Strategy pattern
- Easy to test with mock implementations
- Backward compatible (old Printf/Println methods deprecated but retained)

### Testing Verification
- All unit tests passing
- Integration tests updated and passing
- `make fmt`, `make vet`, and `make lint` all pass
- JSON output now properly formatted without status message contamination