---
priority: 2
description: Migrate new command to new Command interface
created_at: "2025-08-13T17:50:42+09:00"
started_at: "2025-08-13T18:01:59+09:00"
closed_at: null
related:
    - parent:250812-152927-migrate-remaining-commands
---

# Migrate new command to new Command interface

Migrate the `new` command to use the new Command interface, continuing the pattern established by previous migrations. This command creates new tickets and is the first state-changing command to be migrated.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [x] Create `internal/cli/commands/new.go` implementing the Command interface
- [x] Implement App dependency using `cli.NewApp(ctx)` pattern
- [x] Handle positional argument for ticket slug (MinArgs: 1)
- [x] Implement flags:
  - [x] `--parent` / `-p` for parent ticket ID
  - [x] `--format` / `-o` for output format (text/json)
  - [x] Ensure short flag forms (-p, -o) work correctly alongside long forms
- [x] Add slug validation (alphanumeric and hyphens only)
- [x] Handle parent ticket validation and relationship
- [x] Preserve backward compatibility of output formats (text and JSON must match exactly)
- [x] Add comprehensive unit tests with mock App
- [x] Update main.go to register new command
- [x] Remove new case from switch statement
- [x] Test new command functionality with various scenarios:
  - [x] Valid slug creation
  - [x] Invalid slug validation
  - [x] Parent ticket relationship (exists, not done, no circular references)
  - [x] Short vs long flag forms
  - [x] JSON output format (exact format matching)
  - [x] Text output format (exact format matching)
  - [x] Empty/missing slug
  - [x] Error message consistency
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Update migration guide with completion status
- [ ] Get developer approval before closing

## Implementation Notes

### Implementation Strategy
Based on analysis, the recommended approach is:
1. Create command structure with all interface methods in `internal/cli/commands/new.go`
2. Leverage existing `App.NewTicket` method for all business logic (no reimplementation needed)
3. Focus on argument parsing, validation, and flag handling in the command layer
4. Ensure helper methods remain accessible through the App struct
5. Follow the pattern established by `show.go` for positional arguments

### Current Implementation
- Located in switch statement around line 191-213 in main.go
- Calls `handleNew(ctx, slug, parent, format)`
- Takes one required argument: ticket slug
- Has two flags:
  - `--parent` / `-p` for parent ticket ID
  - `--format` / `-o` for output format

### Migration Requirements
1. **App Dependency**: Use `cli.NewApp(ctx)` directly to leverage existing `App.NewTicket` method
2. **Positional Arguments**: Required slug argument with validation
3. **Parent Ticket**: Handle parent ticket resolution and validation (exists, not done, no circular references)
4. **Slug Validation**: Ensure alphanumeric and hyphens only
5. **Output Formatting**: Support both text and JSON output formats with exact backward compatibility
6. **Error Handling**: Preserve exact error messages for consistency with current implementation
7. **Flag Handling**: Both long (--parent, --format) and short (-p, -o) forms must work correctly

### Expected Behavior
- Creates new ticket with provided slug
- Validates slug format (alphanumeric and hyphens)
- Optionally sets parent ticket relationship
- Outputs created ticket info in requested format
- Creates ticket file in todo directory
- Supports both long and short flag forms

## Pattern Differences from Previous Migrations

This is the first migrated command that:
1. **Modifies state** - Creates new ticket files
2. **Has complex validation** - Slug format and parent ticket
3. **Manages relationships** - Parent-child ticket linking

## Estimated Time
**2-3 hours** based on:
- Show command took ~50 minutes (read-only)
- New is more complex (state-changing)
- Requires parent ticket validation
- More complex flag handling (short and long forms)

**Actual Time**: ~2 hours (completed efficiently within estimate)

## Why This Command Next?

1. **Core Functionality**: Essential for ticket creation workflow
2. **Moderate Complexity**: Good stepping stone to more complex commands
3. **Foundation Building**: Establishes patterns for state-changing commands
4. **No Dependencies**: Can be implemented immediately
5. **High Impact**: Frequently used command in daily workflow

## Technical Considerations

1. **Slug Validation**: Must preserve existing validation rules and error messages
2. **Parent Resolution**: Use app.Manager.Get() for parent validation with all edge cases
3. **File Creation**: Ensure atomic file operations
4. **Flag Merging**: Handle both long and short flag forms correctly (may need separate StringVar calls)
5. **Testing**: Mock file system operations for unit tests, include table-driven tests
6. **Backward Compatibility**: Preserve exact output format for both text and JSON
7. **Business Logic Reuse**: Leverage existing `App.NewTicket` method rather than reimplementing
8. **Error Consistency**: Maintain exact error messages including helpful suggestions

## Key Insights from Implementation

### 1. **Flag Handling Pattern**
- Dual flag support (long and short forms) requires separate StringVar calls
- Short forms take precedence when both are provided
- Created `normalize()` helper method to cleanly merge flag values

### 2. **Code Review Improvements**
After golang-pro review (Grade: B+), implemented:
- Constants for format values to avoid magic strings
- Context cancellation check at Execute start
- Improved flag position documentation in Usage string
- Cleaner code organization with helper methods

### 3. **Testing Considerations**
- MockApp already defined in status_test.go (avoid redeclaration)
- Table-driven tests provide excellent coverage for validation scenarios
- Test exact error message format for backward compatibility

### 4. **Pattern Establishment**
This migration establishes patterns for:
- State-changing commands (vs read-only)
- Complex validation logic integration
- Parent-child relationship handling
- Reusing existing business logic from App methods

### 5. **Backward Compatibility Success**
- All existing functionality preserved exactly
- Output formats (text/JSON) match byte-for-byte
- Error messages maintain exact format with helpful suggestions
- Flag behavior identical to original implementation

## References

- See `docs/COMMAND_MIGRATION_GUIDE.md` for migration patterns
- Review `internal/cli/commands/show.go` for positional argument pattern
- Check current `handleNew` implementation in main.go (line ~314)
- Show command PR: #60 for reference implementation

## Completion Status
✅ **All technical tasks completed successfully**
- Implementation follows established patterns
- All tests passing (unit and integration)
- Code review improvements implemented
- Migration guide updated
- Ready for developer approval