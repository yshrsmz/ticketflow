---
priority: 2
description: Add toggle key to cycle between ID, slug, and description display modes in TUI list
created_at: "2025-07-29T10:52:36+09:00"
started_at: "2025-07-29T11:10:09+09:00"
closed_at: null
related:
    - 250728-230137-improve-tui-ticket-title-visibility
---

# Ticket Overview

Implement a display mode toggle in the TUI list view that allows users to cycle between showing full ID, slug only, or description in the first column. Users can press 'd' to switch between display modes based on their preference and current needs.

## Tasks
- [ ] Add DisplayMode type and constants to `list.go`
- [ ] Add displayMode field to TicketListModel struct
- [ ] Implement 'd' key handler in Update() to cycle display modes
- [ ] Modify display logic to show different content based on current mode
- [ ] Add visual indicator showing current display mode (e.g., in header or status bar)
- [ ] Update help text to include 'd' key functionality
- [ ] Persist display mode preference (optional, for future enhancement)
- [ ] Test all three modes with various ticket types
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation with new keyboard shortcut
- [ ] Update README.md to mention display mode toggle feature
- [ ] Get developer approval before closing

## Notes

### Implementation Details:
```go
type DisplayMode int

const (
    DisplayID DisplayMode = iota          // Show full ticket ID (default)
    DisplaySlug                           // Show slug only
    DisplayDescription                    // Show description (fallback to slug if empty)
)

// In TicketListModel
displayMode DisplayMode

// In Update() method
case "d":
    m.displayMode = (m.displayMode + 1) % 3
    // Could also show a temporary message indicating the new mode

// In View() method, around line 305
var displayText string
switch m.displayMode {
case DisplaySlug:
    displayText = t.Slug
case DisplayDescription:
    if t.Description != "" {
        displayText = t.Description
    } else {
        displayText = t.Slug
    }
default:
    displayText = t.ID
}
id := truncate(displayText, idWidth)
```

### User Experience:
- Default mode shows full ID (current behavior)
- Press 'd' to cycle: ID → Slug → Description → ID
- Mode indicator could be shown in header: "Display: ID | Slug | Description"
- Mode persists during current session
- Future: Could save preference in config file

### Benefits:
- Users can choose display based on their workflow
- No information is permanently lost
- Flexible for different use cases (detailed work vs. quick overview)