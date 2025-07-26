---
priority: 2
description: "Fix global shortcuts interfering with text input in TUI"
created_at: 2025-07-26T18:15:25.212321+09:00
started_at: 2025-07-26T22:41:43+09:00
closed_at: 2025-07-26T22:52:00+09:00
---

# 概要

When creating a ticket from TUI, global shortcut is still active and for example pressing 'w' while writing ticket description causes the app to display worktree list.

## expected behavior

User should be able to edit text without being interrupted by global shortcuts.

## タスク
- [x] Analyze global shortcut handling in app.go
- [x] Identify text input implementation in new.go
- [x] Find root cause of shortcut interference
- [x] Implement fix to disable shortcuts during text input

## 技術仕様

### Root Cause Analysis

The issue occurs because global shortcuts in `app.go` are processed before view-specific handling. In the `Update` function (lines 119-141), shortcuts like 'w' for worktree view are checked regardless of the current view state.

When a user is typing in text fields in the new ticket view, pressing 'w' triggers the global shortcut handler before the text input component can process it.

### Code Analysis

1. **Global shortcut handler** (`internal/ui/app.go:119-141`):
   - Processes keys like 'w', 'q', '?' globally
   - Does not check if a text input is currently focused

2. **New ticket view** (`internal/ui/views/new.go`):
   - Has text inputs (textinput, textarea) that need all keyboard input
   - Currently cannot prevent global shortcuts from firing

### Proposed Fix

The best approach is to skip ALL global shortcuts (including '?') when in text input views, except for ctrl+c to quit:

```go
// In app.go Update function, modify the global shortcut handling:
case tea.KeyMsg:
    // Help overlay takes precedence
    if m.help.IsVisible() {
        switch msg.String() {
        case "?", "esc", "q":
            m.help.Hide()
            return m, nil
        }
        return m, nil
    }

    // Skip most global shortcuts when in text input views
    if m.view == ViewNewTicket {
        // Only handle ctrl+c for emergency exit
        if msg.String() == "ctrl+c" {
            return m, tea.Quit
        }
        // Let all other keys (including '?', 'w', 'q') pass to the view
    } else {
        // Normal global shortcuts for non-text-input views
        switch msg.String() {
        case "?":
            m.help.Toggle()
            return m, nil
        case "ctrl+c":
            return m, tea.Quit
        case "q":
            if m.view == ViewTicketList {
                return m, tea.Quit
            }
            m.view = m.previousView
            return m, nil
        case "w":
            if m.view != ViewWorktreeList {
                m.previousView = m.view
                m.view = ViewWorktreeList
                return m, m.worktreeList.Init()
            }
        }
    }
```

This ensures that:
1. Users can type any character (including '?', 'w', 'q') in text fields
2. The help overlay can still be dismissed if accidentally opened
3. Ctrl+C always works as an emergency exit
4. For help in the new ticket view, we could add a hint like "Press ESC and then ? for help"

## メモ

- This is a common issue in TUI applications where global shortcuts conflict with text input
- The fix should be minimal and only affect behavior when text input is active
- Consider also checking for other views that might have text input in the future
- Alternative approaches considered:
  - Using a modifier key (like Alt+w) for global shortcuts - but this is less discoverable
  - Having a "command mode" vs "insert mode" like vim - but this adds complexity
  - The chosen approach (disabling shortcuts in text views) is the simplest and most intuitive

## Implementation Details

### Changes Made:

1. **app.go (lines 118-127)**: Added conditional logic to skip global shortcuts when in text input modes:
   - Checks for `ViewNewTicket` OR when list view is in search mode
   - Only `ctrl+c` is processed for emergency exit
   - All other keys pass through to the text input components
   - Global shortcuts remain active in all other views/modes

2. **list.go (lines 364-367)**: Added `IsSearchMode()` method to expose search mode state

3. **new.go (line 241)**: Added help text "esc then ? for help" to guide users on accessing help

### Testing Notes:
- Build successful with `make build`
- Code formatting and vetting passed
- Manual testing required to verify:
  - In new ticket view: Characters like 'w', 'q', '?' can be typed in text fields
  - In list view search mode: Same characters can be typed in search
  - ESC exits search mode or cancels forms
  - After ESC, global shortcuts work again
  - ctrl+c still exits the application from any mode

### Additional Discovery:
Found that the list view also has a search mode (activated with `/`) that needs the same fix. This has been included in the implementation.

## Insights from Implementation

1. **Multiple Text Input Contexts**: Initially focused only on the new ticket creation view, but discovered the list view also has text input (search mode). This highlights the importance of checking all views for similar functionality.

2. **Clean Abstraction**: The solution uses a boolean flag `isInTextInputMode` to clearly express the condition, making it easy to add more views in the future if needed.

3. **Minimal Impact**: The fix is surgical - it only affects key handling when actually in text input mode, preserving all existing behavior in other contexts.

4. **User Experience Consideration**: Added help text guidance ("esc then ? for help") to ensure users aren't confused about how to access help when global shortcuts are disabled.

5. **Future-Proofing**: The pattern established here (checking for text input mode before processing global shortcuts) can be easily extended when new views with text input are added.

### Lessons Learned:
- Always audit the entire codebase for similar patterns when fixing an issue
- TUI applications need careful consideration of modal vs non-modal input handling
- Clear communication to users (via help text) is crucial when behavior changes contextually
