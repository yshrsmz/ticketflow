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
- [ ] Create `internal/ui/components/close_dialog.go` with text input for closure reason
- [ ] Implement dialog state management (show/hide, input focus, validation)
- [ ] Add confirmation and cancel button handling with proper key bindings (Enter/ESC)

### Phase 2: Close Flow Integration  
- [ ] Extend `closeTicket()` in `internal/ui/app.go:419` to show dialog when 'c' pressed
- [ ] Create `closeTicketWithReason()` method that calls CLI's `CloseTicketWithReason`
- [ ] Modify `moveTicketToDoneAndCommit()` at line 702 to accept optional reason parameter
- [ ] Add branch merge detection using `git.IsBranchMerged()` to determine if reason required

### Phase 3: UI Display Updates
- [ ] Update `internal/ui/views/detail.go` to display `ClosureReason` field when present
- [ ] Add closure reason indicator (e.g., "⚠" icon) to `internal/ui/views/list.go` for abandoned tickets
- [ ] Update help overlay in `internal/ui/components/help.go` with new close shortcuts

### Phase 4: State Management & Shortcuts
- [ ] Implement 'c' for normal close (show dialog only if reason required)
- [ ] Implement 'C' or 'shift+c' for force close with reason
- [ ] Handle ESC key to cancel dialog without closing ticket
- [ ] Ensure dialog state properly resets between uses

### Phase 5: Testing & Validation
- [ ] Test normal close flow remains unchanged for completed tickets
- [ ] Test close with reason for abandoned tickets
- [ ] Test branch merge detection and automatic reason requirement
- [ ] Test dialog cancellation and state cleanup
- [ ] Verify error handling for empty/whitespace-only reasons

## Acceptance Criteria (Updated)

### Core Functionality
- [ ] Dialog appears when closing ticket with 'c' key in detail view
- [ ] Can enter closure reason in text input field
- [ ] Can confirm with Enter or cancel with ESC
- [ ] Empty/whitespace-only reasons are rejected with error message
- [ ] Dialog automatically appears when branch is not merged (reason required)

### Display & Indicators
- [ ] Closure reason shown in ticket detail view under metadata
- [ ] Abandoned tickets (with closure reason) show "⚠" icon in list view
- [ ] Help overlay documents new close shortcuts ('c' and 'C')

### Integration & Consistency
- [ ] TUI uses same `CloseTicketWithReason` backend as CLI
- [ ] Branch merge detection works same as CLI implementation
- [ ] Error messages match CLI format and content
- [ ] Context cancellation handled properly throughout

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