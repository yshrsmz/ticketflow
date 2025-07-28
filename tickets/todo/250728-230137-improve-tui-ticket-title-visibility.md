---
priority: 2
description: "Improve ticket title visibility in TUI list view by handling long date-prefixed filenames"
created_at: "2025-07-28T23:01:37+09:00"
started_at: null
closed_at: null
---

# Ticket Overview

In the TUI list view, ticket titles are displayed using their full filenames which include a date prefix (e.g., "250728-230137-improve-tui-ticket-title-visibility"). Due to limited screen space in the title column, users can barely see the actual ticket content after the date prefix, making it difficult to identify tickets at a glance. We need to improve how ticket titles are displayed while maintaining the ability to identify tickets.

## Tasks
- [ ] Analyze current TUI list view implementation and column width handling
- [ ] Research different approaches for displaying long titles in constrained spaces
- [ ] Implement solution (options include: truncate date prefix, show only slug part, use description field, responsive column widths)
- [ ] Add proper ellipsis handling for overflow text
- [ ] Ensure ticket ID remains visible for identification
- [ ] Test with various terminal widths and ticket title lengths
- [ ] Consider adding tooltip or expanded view on hover/selection
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation if necessary
- [ ] Update README.md if new UI behavior is introduced
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Notes

### Current Issues:
- Ticket filenames follow pattern: `YYMMDD-HHMMSS-slug-description.md`
- The date prefix alone takes 13 characters (250728-230137)
- In narrow terminals or with multiple columns, the slug portion gets cut off
- Users need to see the meaningful part of the ticket (the slug) to identify it

### Code Analysis:
- The ID column width is fixed at 20 characters (see `list.go` line 236: `idWidth := 20`)
- Current display logic at line 305: `id := truncate(t.ID, idWidth)`
- The Ticket struct has separate fields available: `ID`, `Slug`, and `Description`
- Truncation happens in the `truncate()` function which adds "..." when text exceeds width

### Detailed Solution Proposals:

#### Solution 1: Show Only Slug in List View
- **Implementation**: Change line 305 to `id := truncate(t.Slug, idWidth)`
- **Pros**: Maximum visibility for meaningful content, simple one-line change
- **Cons**: No date/time info visible, potential confusion with duplicate slugs
- **Example**: `improve-tui-ticket-t...` instead of `250728-230137-improv...`

#### Solution 2: Smart Truncation (Date + Slug End)
- **Implementation**: Create new function `smartTruncateID()` that shows first 6 chars + "..." + last N chars
- **Code Example**:
  ```go
  func smartTruncateID(id string, maxWidth int) string {
      if len(id) <= maxWidth { return id }
      parts := strings.Split(id, "-")
      if len(parts) >= 3 {
          datePrefix := parts[0][:6] // YYMMDD
          slugPart := strings.Join(parts[2:], "-")
          remaining := maxWidth - 9 // 6 for date, 3 for "..."
          if remaining > 0 && len(slugPart) > remaining {
              return datePrefix + "..." + slugPart[len(slugPart)-remaining:]
          }
      }
      return truncate(id, maxWidth)
  }
  ```
- **Example**: `250728...visibility`

#### Solution 3: Two-Column ID Display
- **Implementation**: Split ID column into Date (13 chars) and Slug (variable width)
- **Changes Required**:
  - Adjust header format in line 242-247
  - Modify column width calculations
  - Update row formatting in line 311-315
- **Pros**: Full date visible, slug gets dedicated space
- **Cons**: More complex layout changes, less space for description

#### Solution 4: Use Description Field as Primary Display
- **Implementation**: 
  ```go
  displayText := t.Description
  if displayText == "" {
      displayText = t.Slug
  }
  id := truncate(displayText, idWidth)
  ```
- **Pros**: User-controlled display text, most flexible
- **Cons**: Requires users to maintain descriptions, inconsistent if some tickets lack descriptions

#### Solution 5: Responsive ID Column Width
- **Implementation**: Calculate width based on content or terminal width percentage
  ```go
  idWidth := int(float64(m.width) * 0.25) // 25% of terminal width
  if idWidth < 20 { idWidth = 20 }        // Minimum width
  if idWidth > 40 { idWidth = 40 }        // Maximum width
  ```
- **Pros**: Adapts to available space
- **Cons**: Layout shifts as window resizes

#### Solution 6: Selection-Based Expansion
- **Implementation**: Show full ID only for selected ticket
  ```go
  if i == m.cursor {
      id = t.ID // Show full ID when selected
  } else {
      id = truncate(t.Slug, idWidth)
  }
  ```
- **Pros**: Clean default view, full info on demand
- **Cons**: ID changes as cursor moves

#### Solution 7: Toggle Display Mode
- **Implementation**: Add display mode state and toggle key
  ```go
  type DisplayMode int
  const (
      DisplayID DisplayMode = iota
      DisplaySlug
      DisplayDescription
  )
  // In Update() add case "d": m.displayMode = (m.displayMode + 1) % 3
  ```
- **Pros**: User preference, no information loss
- **Cons**: Additional state complexity

### Implementation Strategy:
1. Create branch `experiment-tui-title-display`
2. Implement each solution as a separate commit for easy comparison
3. Test scenarios:
   - Narrow terminal (80 columns)
   - Wide terminal (120+ columns)
   - Long ticket slugs
   - Tickets with/without descriptions
4. Consider hybrid approach combining best features
5. Get user feedback on preferred approach

### Considerations:
- Must maintain ability to uniquely identify tickets
- Should work well with both wide and narrow terminals
- Consider accessibility and readability
- Maintain consistency with CLI output format
- Ensure ticket operations (start, close, etc.) still work with displayed text