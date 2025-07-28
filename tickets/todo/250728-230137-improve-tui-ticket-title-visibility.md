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

### Potential Solutions:
1. **Show only slug in list view**: Remove date prefix in display, keep full ID for operations
2. **Use ticket description field**: Display description from YAML frontmatter instead of filename
3. **Truncate intelligently**: Show first few chars of date + "..." + last part of slug
4. **Responsive columns**: Adjust column widths based on terminal size
5. **Two-line display**: Show date on first line, slug on second line (if space permits)

### Considerations:
- Must maintain ability to uniquely identify tickets
- Should work well with both wide and narrow terminals
- Consider accessibility and readability
- Maintain consistency with CLI output format