---
priority: 2
description: "Migrate status command to new Command interface"
created_at: "2025-08-12T18:55:38+09:00"
started_at: null
closed_at: null
related:
    - parent:250810-003001-refactor-command-interface
---

# Migrate status command to new Command interface

Migrate the `status` command to use the new Command interface. This is the first read-only command that requires App dependency, establishing the pattern for future migrations.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [ ] Create `internal/cli/commands/status.go` implementing the Command interface
- [ ] Implement App dependency injection pattern (following migration guide)
- [ ] Handle the -o/--output flag for JSON output format
- [ ] Add comprehensive unit tests with mock App
- [ ] Update main.go to register status command
- [ ] Remove status case from switch statement
- [ ] Test status command functionality (with and without current ticket)
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update migration guide with completion status
- [ ] Get developer approval before closing

## Implementation Notes

### Current Implementation
- Located in switch statement around line 193 in main.go
- Calls `handleStatus(ctx, format)` 
- Has one flag: `-o` for output format (json/text)
- Requires App instance to get current ticket status

### Migration Requirements
1. **App Dependency**: First command needing `cli.App` - establish factory pattern
2. **Flag Handling**: Parse `-o` flag for output format
3. **Error Handling**: Properly handle "no current ticket" scenario
4. **Testing**: Mock App for unit tests

### Expected Behavior
- Shows current ticket information if one exists
- Returns appropriate error if no current ticket
- Supports JSON output format with `-o json` flag
- Default text output shows ticket ID, status, description, and duration

## References

- See `docs/COMMAND_MIGRATION_GUIDE.md` section on "Commands with App Dependencies"
- Review `internal/cli/commands/version.go` for basic command structure
- Check current `handleStatus` implementation in main.go
- App factory pattern example in migration guide

## Why This Command Next?

1. **Simple Read-Only**: No state modifications, lower risk
2. **Establishes Pattern**: First command with App dependency
3. **Single Flag**: Simple flag handling to implement
4. **Quick Win**: Estimated 2-3 hours to complete
5. **Foundation**: Pattern will be reused for list, show, and other commands