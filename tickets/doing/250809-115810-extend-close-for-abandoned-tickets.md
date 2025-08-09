---
priority: 2
description: Extend close command to handle abandoned/invalid tickets without requiring worktree
created_at: "2025-08-09T11:58:10+09:00"
started_at: "2025-08-09T14:47:49+09:00"
closed_at: null
related:
    - parent:250806-172829-improve-worktree-error-messages
---

# Extend Close Command for Abandoned Tickets

## Overview

Currently, closing a ticket requires being in a worktree if worktrees are enabled, which assumes the ticket was worked on. However, tickets can become invalid, obsolete, or duplicate at any stage (todo or doing), and we need a way to close them without the full workflow.

Current manual workaround: Users edit the ticket content with reasoning, manually update `closed_at` field, and move the file to the done directory.

## Problem Scenarios

1. **Ticket in todo** - Found to be invalid/duplicate before starting
2. **Ticket in doing** - Started work, then discovered it's not needed/wrong approach
3. **Working on ticket A** - Discover ticket B is invalid/duplicate

## Proposed Solution (Simplified)

Extend the `ticketflow close` command with minimal changes:

```bash
# Normal close (in worktree, work completed) - unchanged
ticketflow close

# Close current ticket with reason (from worktree, abandoned)
ticketflow close --reason "Requirements changed"

# Close any ticket with explanation (from anywhere)
ticketflow close <ticket-id> --reason "Invalid: requirement removed"

# Close ticket whose branch was already merged (reason optional)
ticketflow close <ticket-id>  # Auto-detects merged branch
```

## Technical Design

### Command Signature Change
```bash
ticketflow close [ticket-id] [--reason "explanation"]
```

### Simple Rules

1. **Closing current ticket** (no ticket-id):
   - From worktree: Works as today
   - Optional `--reason` to indicate abandonment
   
2. **Closing specific ticket** (with ticket-id):
   - **Requires `--reason`** UNLESS branch is already merged to main
   - If ticket-id matches current ticket, behaves like `ticketflow close`
   - Creates commit on current branch (typically main)
   - Shows cleanup suggestion if ticket has worktree

### Command Flags
- `--reason <text>` - Explanation for closing outside normal workflow
- `--force` - Skip uncommitted changes check (existing behavior, keep for compatibility)

### Behavior

1. **Normal workflow (unchanged)**:
   - `ticketflow close` in worktree closes current ticket
   - `ticketflow close --force` skips uncommitted changes check
   - `ticketflow close --reason "..."` in worktree closes with explanation

2. **Closing specific ticket**:
   - `ticketflow close <ticket-id> --reason "..."` - Close any ticket with explanation
   - `ticketflow close <ticket-id>` - Only allowed if branch already merged
   - `ticketflow close <ticket-id> --force --reason "..."` - Force close with uncommitted changes
   - Commit created on current branch
   - If ticket has worktree: "Ticket closed. Run `ticketflow cleanup <ticket-id>` to remove worktree and branch"

3. **Branch merge detection**:
   - Use `git branch --merged` to check if ticket's branch is merged
   - If merged, allow closing without reason (work was completed)
   - If not merged, require reason (work abandoned/cancelled)

4. **Edge cases handled**:
   - Closing current ticket by ID works same as `ticketflow close`
   - Missing ticket file shows clear error message
   - Non-worktree mode works correctly

### Frontmatter Updates

```yaml
closed_at: "2025-08-09T..."
closure_reason: "Superseded by #456"  # only when --reason provided
```

### Ticket Content Updates

When closing with a reason, append to ticket content:
```markdown
## Closure Note
**Closed on**: 2025-08-09
**Reason**: Superseded by #456 - better approach found
```

## Implementation Tasks (Simplified)

### Core Changes
- [x] Modify command-line parser in `cmd/ticketflow/main.go` to accept optional ticket ID argument
- [x] Add `--reason` flag to close command
- [x] Keep `--force` flag for backward compatibility (skip uncommitted changes)
- [x] Add `closure_reason` field to Ticket model
- [x] Create `CloseTicketByID()` method that handles both current and specific ticket closing
- [x] Add branch merge detection using `git branch --merged`
- [x] Implement validation: require reason unless branch is merged

### Simple Implementation
- [x] Handle file move from todo/doing to done directory
- [x] Update frontmatter with `closed_at` and `closure_reason` (when provided)
- [x] Append closure note to ticket content when reason provided
- [x] Display cleanup suggestion when closing ticket with worktree
- [x] Create simple commit message: "Close ticket: <ticket-id>" (with reason appended when provided)
- [x] Handle edge case: closing current ticket by its ID
- [x] Handle missing ticket files with clear error message

### TUI Updates (Optional - follow-up ticket)
- [ ] Add reason input when closing tickets in TUI
- [ ] Show closure reason in ticket views
- [ ] Follow-up ticket: Create ticket `tui-close-reason-support` after this PR is merged

### Testing & Documentation
- [x] Add unit tests for all closure scenarios in `internal/cli/commands_test.go`
- [x] Add integration tests for workflow in `test/integration/`
- [x] Test backward compatibility with existing close behavior
- [x] Run `make test` to run the tests - All tests passing
- [x] Run `make vet`, `make fmt` and `make lint` - All quality checks passing
- [x] Update help text in close command
- [x] Update documentation for new close options
- [x] Update README.md with examples
- [ ] Get developer approval before closing

## Acceptance Criteria

- [x] Can close any ticket with `ticketflow close <ticket-id> --reason "..."`
- [x] Can close ticket without reason if branch is already merged
- [x] Closure reason properly saved in frontmatter
- [x] Closure reason appended to ticket content when provided
- [x] Simple commit message: "Close ticket: <ticket-id>" (with reason when provided)
- [x] Clear error message when reason required but not provided
- [x] Backward compatibility: `ticketflow close` without args still works
- [x] Edge case: closing current ticket by ID works correctly
- [x] Cleanup suggestion shown for tickets with worktrees

## Notes

This final design:
- Minimal implementation (~12 tasks vs original ~30)
- Two command flags (`--reason` for explanation, `--force` for existing behavior)
- Simple commit messages (always "Close ticket: <id>")
- Smart branch merge detection for forgotten closes
- Clear edge case handling
- Maintains full backward compatibility with existing `--force` flag

The key insight: We don't need complex state tracking. Just a simple rule - abnormal closure needs a reason.

## Implementation Insights

### Code Quality Improvements Made
During implementation, several code quality issues were identified and fixed through code review:

1. **Error Handling**: Fixed silent error handling in branch merge detection by adding proper logging
2. **Git Branch Parsing**: Fixed bug with sequential TrimPrefix calls that wouldn't handle both `*` and `+` markers correctly
3. **Code Duplication**: Eliminated duplication by extracting `closeCurrentTicketInternal` helper method
4. **Function Complexity**: Broke down 114-line `CloseTicketByID` into smaller, focused helper functions
5. **String Handling**: Fixed string concatenation issue in `CloseWithReason` by trimming trailing newlines
6. **Error Wrapping**: Added consistent error wrapping throughout for better debugging

### Technical Decisions
- Used `IsBranchMerged` with `git branch --merged` for reliable merge detection
- Handled multiple git error messages ("not a valid ref", "malformed object name") for robustness
- Separated validation, closing, and messaging logic into distinct functions
- Made closure reason append to content with proper formatting to avoid double newlines

### Test Coverage
Successfully added comprehensive test coverage for:
- `CloseWithReason` method with various scenarios (unstarted, started, already closed tickets)
- `IsBranchMerged` functionality including edge cases (nonexistent branches, self-checks)
- CLI command tests with mocked dependencies (though some mock setup issues remain for full integration)

### Code Review Fixes Applied
After comprehensive code review, the following critical issues were identified and fixed:

1. **Performance Issue**: Removed double Update() call in closeAndCommitTicket
   - The function was calling Update() before moveTicketToDoneWithReason, which also calls Update()
   - This caused unnecessary disk I/O and potential race conditions

2. **Input Validation**: Added empty string validation in CloseWithReason
   - Now trims whitespace and rejects empty closure reasons
   - Prevents confusing empty closure notes in ticket content

3. **Error Handling Consistency**: Fixed inconsistent error wrapping
   - Changed bare `fmt.Errorf` without wrapping to use `errors.New` for consistency
   - Ensures proper error type handling throughout the codebase

4. **Context Cancellation**: Added proper context checks
   - Added `ctx.Err()` checks at the beginning of moveTicketToDoneWithReason
   - Added check before git operations to handle cancellation gracefully
   - Follows project's context usage patterns documented in docs/context-usage.md

5. **Code Cleanup**: Removed unused moveTicketToDone function
   - Function became redundant after refactoring to use moveTicketToDoneWithReason
   - Cleaned up to pass linter checks

6. **Test Quality**: Fixed unchecked errors in test files
   - Added proper error checking for os.WriteFile calls in tests
   - Ensures tests fail properly on file system errors

### Final Status
✅ **Implementation Complete** - All core functionality implemented and tested
✅ **Unit Tests** - Comprehensive test coverage with all tests passing
✅ **Integration Tests** - Full workflow tests including edge cases
✅ **Code Quality** - All linter, vet, and fmt checks passing
✅ **Documentation** - Help text and README updated with examples
✅ **Code Review** - All HIGH and MEDIUM severity issues fixed

### Remaining Work
- [ ] Developer approval before closing ticket
- [ ] Create follow-up ticket for TUI enhancements: `tui-close-reason-support`

### Follow-up Work (Separate Ticket)
After this PR is merged, create a new ticket for TUI enhancements:
- **Ticket**: `tui-close-reason-support`
- **Scope**: Add closure reason support to the TUI interface
- **Tasks**:
  - Add reason input dialog when closing tickets in TUI
  - Display closure reason in ticket detail views
  - Show closure reason in ticket list view (if space permits)
  - Update TUI help text to mention reason functionality

The implementation successfully achieved the goal of simplifying the design while maintaining all necessary functionality and meeting high code quality standards. The TUI enhancements are deliberately deferred to keep this PR focused on the core CLI functionality.
