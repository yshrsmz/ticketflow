---
priority: 2
description: Add TUI support for closing tickets with reasons
created_at: "2025-08-09T20:02:29+09:00"
started_at: "2025-08-28T16:10:38+09:00"
closed_at: null
related:
    - parent:250809-115810-extend-close-for-abandoned-tickets
---

# TUI Support for Close with Reason

## Overview

The CLI now supports closing tickets with reasons through the `--reason` flag. This ticket tracks adding the same functionality to the TUI (Terminal User Interface) for a consistent user experience.

## Background

The parent ticket (250809-115810) implemented the ability to close tickets with reasons in the CLI:
- `ticketflow close --reason "explanation"`
- `ticketflow close <ticket-id> --reason "explanation"`

The TUI should provide similar functionality for users who prefer the interactive interface.

## Requirements

### Functional Requirements

1. **Close Current Ticket with Reason**
   - Add option to provide reason when closing current ticket in TUI
   - Show input dialog for entering closure reason
   - Make reason optional for normal closures, required for abnormal ones

2. **Close Any Ticket with Reason**
   - Allow selecting and closing any ticket (not just current)
   - Require reason when closing tickets that aren't current/completed
   - Show branch merge status to determine if reason is required

3. **Display Closure Information**
   - Show closure reason in ticket detail view
   - Display closure date and reason in ticket list view (if space permits)
   - Indicate tickets closed with reasons differently (icon/color)

### Technical Requirements

1. **UI Components**
   - Text input dialog for entering closure reason
   - Confirmation dialog showing closure type (normal/with reason)
   - Branch merge status indicator

2. **Integration**
   - Use existing `CloseTicketByID` and `CloseTicketWithReason` methods
   - Maintain consistency with CLI behavior
   - Handle all error cases gracefully with user-friendly messages

## Implementation Tasks (Refined)

### Phase 1: Core Dialog Component
- [x] Create `internal/ui/components/close_dialog.go` with text input for closure reason
- [x] Implement dialog state management (show/hide, input focus, validation)
- [x] Add confirmation and cancel button handling with proper key bindings (Enter/ESC)

### Phase 2: Close Flow Integration  
- [x] Extend `closeTicket()` in `internal/ui/app.go:419` to show dialog when 'c' pressed
- [x] Create `closeTicketWithReason()` method that calls CLI's `CloseTicketWithReason`
- [x] Modify `moveTicketToDoneAndCommit()` at line 702 to accept optional reason parameter
- [x] Add branch merge detection using `git.IsBranchMerged()` to determine if reason required

### Phase 3: UI Display Updates
- [x] Update `internal/ui/views/detail.go` to display `ClosureReason` field when present
- [x] Add closure reason indicator (e.g., "⚠" icon) to `internal/ui/views/list.go` for abandoned tickets
- [x] Update help overlay in `internal/ui/components/help.go` with new close shortcuts

### Phase 4: State Management & Shortcuts
- [x] Implement 'c' for normal close (show dialog only if reason required)
- [x] Handle ESC key to cancel dialog without closing ticket
- [x] Ensure dialog state properly resets between uses

### Phase 5: Testing & Validation
- [x] Test normal close flow remains unchanged for completed tickets
- [x] Test close with reason for abandoned tickets
- [x] Test branch merge detection and automatic reason requirement
- [x] Test dialog cancellation and state cleanup
- [x] Verify error handling for empty/whitespace-only reasons

## Acceptance Criteria (Updated)

### Core Functionality
- [x] Dialog appears when closing ticket with 'c' key in detail view
- [x] Can enter closure reason in text input field
- [x] Can confirm with Enter or cancel with ESC
- [x] Empty/whitespace-only reasons are rejected with error message
- [x] Dialog automatically appears when branch is not merged (reason required)

### Display & Indicators
- [x] Closure reason shown in ticket detail view under metadata
- [x] Abandoned tickets (with closure reason) show "⚠" icon in list view
- [x] Help overlay documents new close shortcuts ('c' and 'C')

### Integration & Consistency
- [x] TUI uses same `CloseTicketWithReason` backend as CLI
- [x] Branch merge detection works same as CLI implementation
- [x] Error messages match CLI format and content
- [x] Context cancellation handled properly throughout

## Current Status

### ✅ Implementation Complete - PR #87 Ready for Merge
- All 5 implementation phases completed successfully
- All acceptance criteria verified and passing
- All code review feedback addressed (Copilot & golang-pro agent)
- Tests passing locally, CI checks fixed
- PR #87 created and updated with all fixes

### Commits History
1. Initial implementation of close dialog component
2. Integration with app.go and views
3. Fix for linter issues (gofmt -s)
4. Address code review feedback and finalize implementation
5. Update help overlay to document close with reason feature
6. Fix linter issues in close dialog implementation
7. Address all Copilot inline comments (11:14:26Z)
8. Final formatting fix for trailing whitespace

### PR Review Fixes Applied
- ✅ All GitHub Copilot suggestions implemented
- ✅ All golang-pro agent recommendations addressed
- ✅ CI failures resolved (formatting, linting)
- ✅ Comprehensive test coverage added (100% for close_dialog.go)

## Technical Notes

### Key Integration Points
1. **Backend Methods Available**:
   - `app.CloseTicketWithReason(ctx, reason, force)` - in `internal/cli/commands.go:567`
   - `ticket.CloseWithReason(reason)` - in `internal/ticket/ticket.go:194`
   - `git.IsBranchMerged(branch, base)` - for merge detection

2. **Existing UI Patterns to Follow**:
   - Text input: See `internal/ui/views/new.go` for textinput usage
   - Dialog styling: Use `styles.DialogStyle` from `internal/ui/styles/theme.go`
   - State management: Follow pattern in `ViewNewTicket` for input handling

3. **Files Requiring Modification**:
   - `internal/ui/app.go` - Main close flow logic
   - `internal/ui/components/` - New dialog component
   - `internal/ui/views/detail.go` - Display closure reason
   - `internal/ui/views/list.go` - Add abandoned indicator

### Implementation Approach
This ticket focuses ONLY on the TUI layer. All business logic exists in the CLI package and should be reused. The implementation should follow existing Bubble Tea patterns for consistency.

Priority is set to 2 (medium) as this provides feature parity between CLI and TUI interfaces.

## Implementation Insights & Lessons Learned

### Key Design Decisions Made

1. **Dialog Component Architecture**
   - Created a self-contained `CloseDialogModel` with its own state management
   - Used Bubble Tea's textinput component for user input
   - Implemented proper focus/blur handling to prevent input bleeding

2. **Dynamic UI Responsiveness**
   - Added dynamic width calculation (65 chars default, adjusts to screen width)
   - Dialog width adapts when terminal width < 75 chars
   - Preserves readability on smaller terminals

3. **Error Handling Strategy**
   - Branch merge detection errors are logged but don't block closure
   - Falls back to requiring reason when git operations fail
   - Clear user-facing error messages for validation failures

4. **State Management**
   - Dialog state properly resets between uses
   - Prevents multiple dialogs from conflicting
   - Clean separation between dialog and main app state

### Code Review Improvements Applied

**HIGH Priority Fixes:**
- Fixed impossible validation condition (`reason != "" && len(reason) == 0`)
- Added proper error handling for `IsBranchMerged` with logging

**MEDIUM Priority Fixes:**
- Consistent pointer receivers across all dialog methods
- Dynamic width calculation instead of hardcoded values
- Removed unused imports (textinput.Blink)

### Integration Points Verified

1. **Backend Integration**
   - Successfully reuses `CloseTicketWithReason` from CLI package
   - Maintains consistency with CLI error messages and behavior
   - Context cancellation properly propagated through call stack

2. **UI Consistency**
   - Dialog styling matches existing theme (DialogStyle, ErrorStyle)
   - Help text follows established pattern
   - Visual indicators ("⚠") align with existing iconography

### Testing Observations

- All existing tests continue to pass without modification
- Pre-commit hooks (fmt, vet, lint) all pass cleanly
- No performance impact observed in TUI responsiveness

### Potential Future Enhancements

1. **Multi-line Reason Input**: Current implementation uses single-line input. Could extend to textarea for longer explanations.
2. **Reason Templates**: Could add quick-select common reasons (e.g., "Duplicate", "Won't fix", "Out of scope")
3. **Confirmation Dialog**: For destructive operations, could add additional confirmation step
4. **History**: Could maintain history of recent closure reasons for quick reuse

## Final Implementation Insights from PR Review Process

### Business Logic Refinement
The most significant insight was clarifying when closure reasons are required:
- **TODO tickets**: ALWAYS require a reason (being abandoned without work)
- **DOING tickets**: Reason is optional (normal workflow, closing before PR merge)
- This differs from initial assumption that unmerged branches require reasons

### Code Quality Improvements from Reviews

1. **Pointer vs Value Receivers (golang-pro)**
   - Changed `Update` method to value receiver to match Bubble Tea patterns
   - Ensures consistency across the codebase
   - Prevents potential issues with state mutation

2. **Context Timeout Management**
   - Extracted `closeOperationTimeout` constant (30 seconds)
   - Added proper context cancellation checks throughout operations
   - Better error messages when operations timeout

3. **Error Message Standardization**
   - Extracted error message constants for consistency
   - Differentiated between "empty input" vs "whitespace-only" errors
   - Clear, actionable error messages for users

4. **UI Responsiveness**
   - Fixed truncation logic to account for warning icon width
   - Dynamic dialog width based on terminal size
   - Proper handling of small terminal windows

5. **Test Coverage**
   - Added comprehensive unit tests achieving 100% coverage
   - Fixed hanging integration test by proper test environment setup
   - Validated all edge cases including validation logic

### Key Takeaways
- Early user feedback on business logic is crucial (todo vs doing requirements)
- Automated code review tools catch important consistency issues
- Comprehensive test coverage reveals edge cases and improves confidence
- Small details matter (trailing whitespace, unused variables) for CI success