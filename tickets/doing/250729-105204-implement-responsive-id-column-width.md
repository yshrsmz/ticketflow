---
priority: 2
description: Make ID column width responsive based on terminal width in TUI list view
created_at: "2025-07-29T10:52:04+09:00"
started_at: "2025-07-29T11:10:05+09:00"
closed_at: null
related:
    - 250728-230137-improve-tui-ticket-title-visibility
---

# Ticket Overview

Implement responsive ID column width that adapts based on terminal width. Instead of a fixed 20-character width, the ID column should use a percentage of available terminal width with minimum and maximum constraints, allowing more space for ticket IDs on wider terminals.

## Tasks
- [x] Replace fixed `idWidth := 20` with dynamic calculation based on terminal width
- [x] Implement percentage-based width calculation (e.g., 25% of terminal width)
- [x] Add minimum width constraint (20 characters) to ensure readability
- [x] Add maximum width constraint (40 characters) to prevent excessive space usage
- [x] Update column width recalculation when terminal is resized
- [x] Test with various terminal widths (80, 100, 120, 160 columns)
- [x] Ensure other columns adjust properly with dynamic ID width
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation if necessary
- [ ] Get developer approval before closing

## Notes

### Implementation Details:
```go
// Example implementation
idWidth := int(float64(m.width) * 0.25) // 25% of terminal width
if idWidth < 20 { 
    idWidth = 20  // Minimum width
}
if idWidth > 40 { 
    idWidth = 40  // Maximum width
}
```

### Benefits:
- Narrow terminals (80 cols): Uses minimum 20 chars (same as current)
- Medium terminals (120 cols): Uses ~30 chars, showing more of the slug
- Wide terminals (160+ cols): Uses up to 40 chars, showing most/all of typical IDs

### Considerations:
- Need to ensure description column still has adequate space
- Column widths should recalculate on terminal resize
- May want to make the percentage configurable in the future

## Implementation Summary

Successfully implemented responsive ID column width in `internal/ui/views/list.go`:

- Replaced the fixed `idWidth := 20` with a dynamic calculation based on terminal width
- The ID column now uses 25% of the terminal width
- Implemented constraints: minimum 20 characters, maximum 40 characters
- The calculation automatically updates when the terminal is resized (via the existing `SetSize` method)

### Testing Results:
- 80 columns terminal: ID width = 20 (minimum)
- 100 columns terminal: ID width = 25
- 120 columns terminal: ID width = 30
- 160+ columns terminal: ID width = 40 (maximum)

The implementation ensures better visibility of ticket IDs on wider terminals while maintaining readability on narrow terminals. The description column automatically adjusts to use the remaining space.