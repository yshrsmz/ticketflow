---
priority: 2
description: "Refactor cli package to respect JSON format setting in AutoCleanup and related functions"
created_at: "2025-08-16T12:37:03+09:00"
started_at: null
closed_at: null
related:
    - parent:250815-175624-test-coverage-maintenance-commands
---

# Improve JSON Output Separation in CLI Package

## Problem
The AutoCleanup function and its helper methods in `internal/cli/cleanup.go` output status messages directly to stdout using `fmt.Printf` and `fmt.Println`, regardless of the output format setting. This causes mixed text/JSON output when JSON format is requested, making it difficult for tools to parse the JSON response.

## Current Behavior
When running `ticketflow cleanup --format json`, the output contains both text status messages and JSON:
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
When JSON format is specified, only valid JSON should be output to stdout. Status messages should either be:
1. Suppressed entirely in JSON mode
2. Output to stderr instead of stdout
3. Included within the JSON structure

## Affected Functions
- `AutoCleanup()` - Uses fmt.Println for status messages
- `cleanOrphanedWorktrees()` - Uses fmt.Println for progress
- `cleanStaleBranches()` - Uses fmt.Printf for progress
- `CleanupStats()` - Uses fmt.Println for statistics
- `CleanupTicket()` - Uses fmt.Printf for confirmation prompts

## Proposed Solution
Pass the output format setting through to the cli package methods, either:
1. Add format parameter to AutoCleanup and related methods
2. Add an OutputWriter interface that can handle both text and JSON modes
3. Use the existing app.Output for all output operations

## Tasks
- [ ] Analyze current output patterns in cli package
- [ ] Design solution for format-aware output
- [ ] Refactor AutoCleanup to respect format setting
- [ ] Refactor cleanOrphanedWorktrees to respect format setting
- [ ] Refactor cleanStaleBranches to respect format setting
- [ ] Refactor CleanupStats to respect format setting
- [ ] Refactor CleanupTicket confirmation prompts
- [ ] Update integration tests to verify clean JSON output
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update CLAUDE.md if API changes
- [ ] Get developer approval before closing

## Notes
This issue was discovered while implementing integration tests for the cleanup command. The tests had to work around the mixed output by extracting JSON from the combined text+JSON output.