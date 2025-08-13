---
priority: 2
description: "Migrate new command to new Command interface"
created_at: "2025-08-13T17:50:42+09:00"
started_at: null
closed_at: null
related:
    - parent:250812-152927-migrate-remaining-commands
---

# Migrate new command to new Command interface

Migrate the `new` command to use the new Command interface, continuing the pattern established by previous migrations. This command creates new tickets and is the first state-changing command to be migrated.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [ ] Create `internal/cli/commands/new.go` implementing the Command interface
- [ ] Implement App dependency using `cli.NewApp(ctx)` pattern
- [ ] Handle positional argument for ticket slug (MinArgs: 1)
- [ ] Implement flags:
  - [ ] `--parent` / `-p` for parent ticket ID
  - [ ] `--format` / `-o` for output format (text/json)
- [ ] Add slug validation (alphanumeric and hyphens only)
- [ ] Handle parent ticket validation and relationship
- [ ] Add comprehensive unit tests with mock App
- [ ] Update main.go to register new command
- [ ] Remove new case from switch statement
- [ ] Test new command functionality with various scenarios:
  - [ ] Valid slug creation
  - [ ] Invalid slug validation
  - [ ] Parent ticket relationship
  - [ ] JSON output format
  - [ ] Empty/missing slug
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update migration guide with completion status
- [ ] Get developer approval before closing

## Implementation Notes

### Current Implementation
- Located in switch statement around line 191-213 in main.go
- Calls `handleNew(ctx, slug, parent, format)`
- Takes one required argument: ticket slug
- Has two flags:
  - `--parent` / `-p` for parent ticket ID
  - `--format` / `-o` for output format

### Migration Requirements
1. **App Dependency**: Use `cli.NewApp(ctx)` directly
2. **Positional Arguments**: Required slug argument with validation
3. **Parent Ticket**: Handle parent ticket resolution and validation
4. **Slug Validation**: Ensure alphanumeric and hyphens only
5. **Output Formatting**: Support both text and JSON output formats
6. **Error Handling**: Clear messages for invalid slugs or missing parents

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

## Why This Command Next?

1. **Core Functionality**: Essential for ticket creation workflow
2. **Moderate Complexity**: Good stepping stone to more complex commands
3. **Foundation Building**: Establishes patterns for state-changing commands
4. **No Dependencies**: Can be implemented immediately
5. **High Impact**: Frequently used command in daily workflow

## Technical Considerations

1. **Slug Validation**: Must preserve existing validation rules
2. **Parent Resolution**: Use app.Manager.Get() for parent validation
3. **File Creation**: Ensure atomic file operations
4. **Flag Merging**: Handle both long and short flag forms
5. **Testing**: Mock file system operations for unit tests
6. **Backward Compatibility**: Preserve exact output format

## References

- See `docs/COMMAND_MIGRATION_GUIDE.md` for migration patterns
- Review `internal/cli/commands/show.go` for positional argument pattern
- Check current `handleNew` implementation in main.go (line ~314)
- Show command PR: #60 for reference implementation