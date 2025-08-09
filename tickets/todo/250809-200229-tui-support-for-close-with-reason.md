---
priority: 2
description: Add TUI support for closing tickets with reasons
created_at: "2025-08-09T20:02:29+09:00"
started_at: null
closed_at: null
related:
    - "parent:250809-115810-extend-close-for-abandoned-tickets"
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

## Implementation Tasks

### UI Components
- [ ] Create reason input dialog component
- [ ] Add close confirmation dialog with reason preview
- [ ] Update ticket detail view to show closure reason
- [ ] Add closure reason indicator to list view

### Integration
- [ ] Add "Close with Reason" option to ticket actions menu
- [ ] Implement branch merge detection in TUI
- [ ] Add validation for required reasons
- [ ] Handle error messages and user feedback

### Testing
- [ ] Test closing current ticket with/without reason
- [ ] Test closing other tickets with required reason
- [ ] Test branch merge detection display
- [ ] Test error handling for invalid inputs

## Acceptance Criteria

- [ ] Can close current ticket with optional reason in TUI
- [ ] Can close any ticket with reason from TUI
- [ ] Closure reason displayed in ticket detail view
- [ ] Branch merge status shown when relevant
- [ ] Error messages match CLI behavior
- [ ] User experience is intuitive and consistent

## Notes

This is a follow-up enhancement to the CLI close with reason feature. The core logic already exists in the `internal/cli` package and should be reused where possible.

Priority is set to 2 (medium) as this is a nice-to-have enhancement that improves UX consistency but isn't blocking core functionality.