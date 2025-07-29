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
- [x] Add DisplayMode type and constants to `list.go`
- [x] Add displayMode field to TicketListModel struct
- [x] Implement 'd' key handler in Update() to cycle display modes
- [x] Modify display logic to show different content based on current mode
- [x] Add visual indicator showing current display mode (e.g., in header or status bar)
- [x] Update help text to include 'd' key functionality
- [ ] Persist display mode preference (optional, for future enhancement)
- [x] Test all three modes with various ticket types
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
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

## Implementation Results and Decision

### Implementation Outcome:
The display mode toggle feature was successfully implemented and tested. The implementation:
- Added a DisplayMode type with three modes (ID, Slug, Description)
- Implemented the 'd' key handler to cycle between modes
- Added a visual indicator in the header showing the current mode
- Updated help text to document the new functionality
- All tests passed and code quality checks were successful

### Feature Analysis:
**Pros:**
- Feature worked exactly as designed
- Users could press 'd' to easily cycle between display modes
- Provided flexibility for different user preferences
- Clean implementation that integrated well with existing code

**Cons:**
- Added complexity to the UI state management
- Potential for user confusion with mode switching
- Requires user interaction to see different information
- Additional cognitive load for users to remember current mode

### Decision: Not Proceeding
After implementing and testing this solution, we've decided not to proceed with the display mode toggle in favor of the simpler responsive width approach (ticket 250729-105204). The reasons for this decision:

1. **Simplicity**: The responsive width solution provides most of the benefits without requiring any user interaction
2. **User Experience**: No need for users to learn a new keybinding or remember their current display mode
3. **Automatic Optimization**: The responsive approach automatically shows as much information as possible based on terminal width
4. **Reduced Complexity**: Less state to manage in the UI, making the codebase simpler to maintain

While the toggle feature worked well and was properly implemented, the responsive width solution provides a better balance between showing more ticket ID information and maintaining simplicity. The automatic adjustment based on terminal width is more intuitive and requires zero user interaction, making it the preferred approach.