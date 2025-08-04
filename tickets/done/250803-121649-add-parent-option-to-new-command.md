---
priority: 3
description: Add --parent flag to ticketflow new command for explicit parent relationships
created_at: "2025-08-03T12:16:49+09:00"
started_at: "2025-08-03T22:50:03+09:00"
closed_at: "2025-08-04T12:22:05+09:00"
---

# Add Parent Option to New Command

## Overview
Currently, parent-child relationships are only created when running `ticketflow new` from within a parent ticket's worktree. Add a `--parent` flag to explicitly specify a parent ticket from any location.

This is a standalone feature enhancement, not a sub-ticket of the branch fix.

## Tasks
- [x] Add --parent flag to new command
- [x] Validate parent ticket exists and is valid
- [x] Add parent relationship to ticket metadata
- [x] Update help text and documentation
- [x] Add tests for parent flag functionality

## Technical Details
```bash
# Current behavior (implicit parent from worktree)
cd ../ticketflow.worktrees/parent-ticket
ticketflow new sub-ticket

# Proposed behavior (explicit parent)
ticketflow new --parent parent-ticket-id sub-ticket
ticketflow new -p parent-ticket-id sub-ticket  # short form
```

Implementation:
- Add flag to CLI parser: `flag.String("parent", "", "Parent ticket ID")`
- Validate parent ticket exists in any status (todo/doing/done)
- Add to ticket's Related field: `parent:<ticket-id>`
- Should work from any directory, not just worktrees

### Parent Option Handling
1. **Multiple Parents**: Not allowed. A ticket can have at most one parent. If multiple `--parent` flags are specified, validate and return an error.
2. **Explicit vs Implicit Parent Priority**:
   - If `--parent` flag is provided: Use explicit parent exclusively, ignore current worktree
   - If no `--parent` flag: Check if current branch is a ticket (implicit parent)
   - This gives users predictable behavior and full control

3. **Validation Requirements**:
   - Parent ticket must exist in any status (todo/doing/done)
   - Prevent circular dependencies (ticket cannot be its own parent)
   - Clear error messages for invalid parent IDs

4. **User Feedback**:
   - "Creating sub-ticket with parent: <explicit-parent-id>" (when using --parent)
   - "Creating ticket in branch: <current-branch>" (when using implicit parent)
   - Show warning if in ticket worktree but using different parent: "Using explicit parent '<parent-id>' instead of current worktree '<current-ticket>'"

## Acceptance Criteria
- Can create sub-tickets with explicit parent from main repo
- Can create sub-tickets with explicit parent from any worktree
- Parent validation provides clear error if ticket doesn't exist
- Help text clearly explains the option
- Tests cover both valid and invalid parent scenarios
- Explicit parent flag overrides implicit worktree parent
- Clear user feedback about which parent is being used

## Implementation Notes

### Key Insights from Implementation

1. **Flag Parsing Order**: Go's flag parsing requires flags to come BEFORE positional arguments. This is critical for the feature to work:
   ```bash
   ticketflow new --parent parent-id ticket-slug   # ✅ Correct
   ticketflow new ticket-slug --parent parent-id   # ❌ Won't work
   ```

2. **Self-Parent Prevention**: Added validation to prevent both `--parent slug` and `--parent generated-id` to avoid circular dependencies. The check happens before validating parent existence to avoid unnecessary database queries.

3. **Dual Flag Support**: Implemented both `--parent` and `-p` (short form) with validation to ensure they're not used together with different values.

4. **Clear User Feedback**: The implementation provides different messages for different scenarios:
   - "Creating sub-ticket with parent: X" (explicit parent)
   - "Creating ticket in branch: Y" (implicit parent from worktree)
   - "Using explicit parent 'X' instead of current worktree 'Y'" (override warning)

5. **Test Coverage**: Added comprehensive unit tests and integration tests covering all scenarios including edge cases like self-parenting and non-existent parents.

### Files Modified
- `cmd/ticketflow/main.go`: Added newFlags struct and flag parsing
- `internal/cli/commands.go`: Updated NewTicket to handle explicit parent parameter
- `internal/cli/commands_parent_test.go`: New comprehensive unit tests
- `test/integration/parent_flag_test.go`: New integration tests

The feature is fully functional and tested, ready for use.
