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
- [ ] Modify column width calculations in `list.go` to accommodate separate date and slug columns
- [ ] Update header format (lines 242-247) to show "Date" and "Slug" instead of "ID"
- [ ] Parse ticket ID to extract date and slug parts in the display loop
- [ ] Update row formatting (lines 311-315) to display date and slug in separate columns
- [ ] Adjust total width calculations to ensure proper layout
- [ ] Test with various terminal widths to ensure responsive behavior
- [ ] Ensure truncation works properly for long slugs in the slug column
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
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