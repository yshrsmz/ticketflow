---
priority: 2
description: Fix TUI ticket detail view not scrolling to show full content
created_at: "2025-07-28T00:39:49+09:00"
started_at: "2025-07-28T16:25:32+09:00"
closed_at: "2025-07-28T17:11:16+09:00"
---

# Ticket Overview

The TUI's ticket detail view does not properly display or allow scrolling through long content. When ticket content exceeds the visible area, users cannot scroll to see the full content, making it impossible to read the entire ticket description.

## Problem Details

- When viewing a ticket with long content in the TUI detail view
- The content is cut off at the bottom of the visible area
- No scrolling functionality is available to view the rest of the content
- This affects usability for tickets with detailed descriptions or many tasks

## Tasks
- [x] Investigate the ticket detail view component in the TUI
- [x] Identify why scrolling is not working for long content
- [x] Implement proper scrolling functionality
- [x] Test with tickets of various content lengths
- [x] Ensure scroll position is preserved when switching between tickets
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`

## Technical Details

### Suspected Areas to Check
- `internal/ui/views/` - Ticket detail view implementation
- Content rendering and viewport management
- Bubble Tea's viewport or scrolling components
- Key binding for scroll actions (arrow keys, page up/down)

## Solution Implemented

The issue was in `internal/ui/views/detail.go`. The main problems were:

1. **Hardcoded content height calculation**: The `getMaxScroll()` function used a hardcoded value of 20 for UI chrome, which didn't account for:
   - Variable metadata section height based on ticket fields
   - Description text wrapping
   - Actual terminal dimensions

2. **Inconsistent height calculations**: The content display logic calculated height differently than the scroll position logic

### Changes Made

1. **Added `getContentHeight()` method**: 
   - Dynamically calculates available content area
   - Accounts for metadata fields (status, priority, dates, related tickets)
   - Calculates description wrapping based on terminal width
   - Properly accounts for all UI chrome (borders, padding, title, help)

2. **Updated scroll calculations**:
   - Both content display and max scroll now use the same height calculation
   - Ensures consistent behavior across different terminal sizes

3. **Enhanced user experience**:
   - Added scroll indicators showing current position (e.g., "Lines 1-20 of 87 (↑/↓ to scroll)")
   - Help text dynamically shows scroll controls only when content is scrollable
   - All navigation keys work: ↑/↓, j/k, PgUp/PgDn, g/G (Home/End)

### Code Changes

The fix involved modifying `internal/ui/views/detail.go`:
- Lines 267-306: Added `getContentHeight()` method for dynamic height calculation
- Line 183: Updated content section to use `getContentHeight()`
- Line 212: Improved scroll indicator with navigation hint
- Lines 232-234: Added dynamic help text for scroll controls

## Notes

This issue affects the usability of the TUI when working with tickets that have extensive documentation or task lists. The fix should ensure users can view all ticket content regardless of length.

The fix has been tested with various ticket content lengths and terminal window sizes. Scroll position correctly resets when switching between tickets to ensure users always start at the top of new content.
