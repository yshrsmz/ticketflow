---
priority: 2
description: Migrate list command to new Command interface
created_at: "2025-08-12T21:36:13+09:00"
started_at: "2025-08-13T11:54:16+09:00"
closed_at: null
related:
    - parent:250810-003001-refactor-command-interface
---

# Migrate list command to new Command interface

Migrate the `list` command to use the new Command interface, building on the pattern established by the status command migration. This command lists tickets with various filtering options.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [ ] Create `internal/cli/commands/list.go` implementing the Command interface
- [ ] Implement App dependency using `cli.NewApp(ctx)` pattern
- [ ] Handle multiple flags:
  - [ ] `--status` flag for filtering (todo/doing/done/all)
  - [ ] `--count` flag for limiting results (default: 20)
  - [ ] `--format` flag for output format (text/json)
- [ ] Add status value validation (todo/doing/done/all/"")
- [ ] Add comprehensive unit tests with mock App
- [ ] Update main.go to register list command
- [ ] Remove list case from switch statement
- [ ] Test list command functionality with various flag combinations
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update migration guide with completion status
- [ ] Get developer approval before closing

## Implementation Notes

### Current Implementation
- Located in switch statement around line 200-230 in main.go
- Calls `handleList(ctx, status, count, format)`
- Has three flags:
  - `-s/--status`: Filter by status (default: "active" which shows todo+doing)
  - `-c/--count`: Number of tickets to show (default: 20)
  - `--format`: Output format (text/json)

### Migration Requirements
1. **App Dependency**: Use `cli.NewApp(ctx)` directly (following status pattern)
2. **Multiple Flags**: Handle 3 flags vs status command's single flag
3. **Status Validation**: Validate status values (todo/doing/done/all/"")
4. **Count Validation**: Ensure count is positive integer
5. **Default Handling**: Properly handle default values for each flag

### Expected Behavior
- Lists tickets based on status filter
- Empty status or "active" shows todo + doing tickets
- "all" shows all tickets regardless of status
- Specific status shows only tickets with that status
- Respects count limit for number of results
- Supports both text and JSON output formats

## Pattern Reuse from Status Command

1. **Direct App Creation**: Use `cli.NewApp(ctx)` in Execute method
2. **Flag Structure**: Define unexported `listFlags` struct
3. **Test Pattern**: Follow status_test.go structure
4. **Error Handling**: Let App.ListTickets handle business logic errors
5. **Output Format**: Use `cli.ParseOutputFormat()` for consistency

## Estimated Time
**25-30 minutes** based on:
- Status command took 20 minutes
- Additional complexity for 2 more flags
- Status validation logic
- Slightly more complex testing

## Why This Command Next?

1. **Natural Progression**: More complex than status but still read-only
2. **High User Impact**: One of the most frequently used commands
3. **Pattern Building**: Extends the App dependency pattern with more flags
4. **Low Risk**: Read-only operation, no data modifications
5. **Foundation**: Sets up patterns for show command and other read operations

## References

- See `docs/COMMAND_MIGRATION_GUIDE.md` for migration patterns
- Review `internal/cli/commands/status.go` for established pattern
- Check current `handleList` implementation in main.go
- Status command PR: #57 for reference implementation