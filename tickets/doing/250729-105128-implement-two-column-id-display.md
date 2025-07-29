---
priority: 2
description: Split ticket ID display into separate date and slug columns in TUI list view
created_at: "2025-07-29T10:51:28+09:00"
started_at: "2025-07-29T11:10:01+09:00"
closed_at: null
related:
    - 250728-230137-improve-tui-ticket-title-visibility
---

# Ticket Overview

Implement a two-column display for ticket IDs in the TUI list view, splitting the current ID column into a Date column (13 chars) and a Slug column (variable width). This will ensure the full date is always visible while giving more space to display the meaningful slug portion of ticket IDs.

## Tasks
- [x] Modify column width calculations in `list.go` to accommodate separate date and slug columns
- [x] Update header format (lines 242-247) to show "Date" and "Slug" instead of "ID"
- [x] Parse ticket ID to extract date and slug parts in the display loop
- [x] Update row formatting (lines 311-315) to display date and slug in separate columns
- [x] Adjust total width calculations to ensure proper layout
- [x] Test with various terminal widths to ensure responsive behavior
- [x] Ensure truncation works properly for long slugs in the slug column
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation if necessary
- [ ] Get developer approval before closing

## Notes

### Implementation Details:
- Date column should be fixed at 13 characters (YYMMDD-HHMMSS format)
- Slug column should use remaining available space after other columns
- Need to handle tickets that might not follow the expected ID format gracefully
- Consider alignment and padding between columns for readability

### Example Layout:
```
Date          Slug                  Status  Pri  Description
────────────────────────────────────────────────────────────
250728-230137 improve-tui-ticket... todo    2    Improve ticket title visibility...
250728-171759 fix-tui-breakage-a... done    2    TUI breaks when returning from...
```

### Implementation Results (2025-07-29):

The two-column ID display was successfully implemented and tested. All functionality works as expected:
- Date and slug are correctly split into separate columns
- Date column maintains fixed 13-character width
- Slug column uses available space with proper truncation
- Layout responds correctly to terminal width changes
- Fixed an existing description truncation bug during implementation

However, after completing the implementation, several insights emerged:

1. **Increased Complexity**: The solution required significant changes to the column width calculation logic, header formatting, and row rendering. This added complexity to the codebase that needs to be maintained.

2. **Layout Calculations**: Managing two columns instead of one required additional calculations for spacing, padding, and truncation. The code became more intricate with more edge cases to handle.

3. **Limited Benefit**: While the solution works, the benefit over a simpler responsive single-column approach is minimal. Users can still see the full date and a meaningful portion of the slug.

4. **Alternative Solution**: The responsive width solution (implemented in ticket 250729-105204) achieves similar goals with much less complexity. It allows the ID column to expand dynamically based on content while maintaining readability.

### Decision:

After careful consideration, we've decided not to proceed with this two-column solution. While it was successfully implemented and tested, the added complexity doesn't justify the marginal improvement over the simpler responsive width approach. The responsive width solution provides:
- Similar visibility for both date and slug parts
- Much simpler implementation
- Easier maintenance
- Better flexibility for future enhancements

This ticket is being closed as "explored but not adopted" - the implementation proved the concept works but highlighted that a simpler solution better serves the project's needs.