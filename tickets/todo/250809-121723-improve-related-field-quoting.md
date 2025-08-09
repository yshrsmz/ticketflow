---
priority: 2
description: "Standardize quoting of related field values to prevent false-positive code reviews"
created_at: "2025-08-09T12:17:23+09:00"
started_at: null
closed_at: null
related:
    - "parent:250806-172829-improve-worktree-error-messages"
---

# Improve Related Field Quoting

## Overview
Standardize the use of quotes around `related` field values in ticket YAML frontmatter to prevent false-positive suggestions from automated code review tools (like GitHub Copilot) that incorrectly identify patterns like `parent:ticket-id` as YAML syntax errors.

## Problem
Currently, the `related` field values are written without quotes:
```yaml
related:
    - parent:250803-121506-worktree-recovery
```

Automated review tools often flag this as a YAML formatting issue, suggesting to add a space after the colon (e.g., `parent: 250803-121506-worktree-recovery`), which would actually break the ticketflow parsing logic.

## Solution
Wrap all `related` field values in quotes to make it explicit that these are single string values, not YAML key-value pairs:
```yaml
related:
    - "parent:250803-121506-worktree-recovery"
    - "blocks:250804-090000-dependent-feature"
    - "related:250805-100000-similar-work"
```

## Tasks
- [ ] Update ticket creation logic to use quoted format for new tickets
- [ ] Update ticket update operations (e.g., when setting parent) to use quoted format
- [ ] Add migration script or command to update existing tickets (optional, for consistency)
- [ ] Update tests to use and expect quoted format
- [ ] Update documentation and examples to show quoted format
- [ ] Verify backward compatibility (unquoted values should still work)
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Get developer approval before closing

## Implementation Details

### Files to Update
1. **Ticket Creation/Update**:
   - `internal/cli/commands.go` - Update parent setting logic to use quoted format
   - Any other places that programmatically add related entries

2. **Tests**:
   - Update test fixtures to use quoted format
   - Ensure tests verify both quoted and unquoted formats work

3. **Documentation**:
   - `CLAUDE.md` - Update examples
   - `design/*.md` - Update any design docs with examples
   - README if it contains examples

### Backward Compatibility
The change must be backward compatible:
- Existing tickets without quotes must continue to work
- The parsing logic (`strings.HasPrefix`, `strings.TrimPrefix`) doesn't need changes
- Both formats should be accepted, but new entries should use quotes

### Example Code Changes
```go
// Before
t.Related = append(t.Related, fmt.Sprintf("parent:%s", parentTicketID))

// After  
t.Related = append(t.Related, fmt.Sprintf(`"parent:%s"`, parentTicketID))
```

Note: The quotes should be part of the string value itself, not just YAML formatting.

## Acceptance Criteria
- [ ] New tickets created with parent relationship use quoted format
- [ ] Existing tickets with unquoted format still work correctly
- [ ] Automated review tools no longer flag related field values as formatting issues
- [ ] All tests pass with the new format
- [ ] Documentation is updated with the new format

## Benefits
1. Eliminates false-positive code review suggestions
2. Makes YAML structure more explicit and less ambiguous
3. Improves developer experience by reducing noise in PR reviews
4. Maintains full backward compatibility

## Notes
This is a quality-of-life improvement that will prevent confusion and unnecessary review comments. The quoted format is more explicit about the intent that these are single string values containing colons, not YAML nested structures.