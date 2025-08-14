---
priority: 2
description: Migrate restore command to new Command interface
created_at: "2025-08-14T11:15:07+09:00"
started_at: "2025-08-14T15:15:27+09:00"
closed_at: null
related:
    - parent:250812-152927-migrate-remaining-commands
---

# Migrate restore command to new Command interface

Migrate the `restore` command to use the new Command interface, enabling restoration of the `current-ticket.md` symlink when working in a worktree. This command fixes broken or missing symlinks by detecting the current git branch and re-establishing the link to the corresponding ticket file.

## Why This Command Next?

1. **Simplest Remaining Command**: Only 7 lines of implementation, making it ideal for a quick win
2. **Essential for Worktree Workflow**: Fixes broken symlinks that can occur during development
3. **High User Value**: Critical for maintaining worktree context and navigation
4. **Pattern Establishment**: Sets pattern for zero-argument, current-context commands
5. **Low Risk**: Minimal complexity reduces implementation risk

## Important Clarification

**This command does NOT restore closed tickets back to "doing" status.** It only restores the `current-ticket.md` symlink that points to the active ticket in a worktree. The symlink can become broken or missing due to various git operations or file system issues.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

### 1. Create Command File and Structure
- [x] Create `internal/cli/commands/restore.go` implementing the Command interface
- [x] Define command struct (no flags needed for this simple command)
- [x] Implement basic Command interface methods (Name, Aliases, Description, Usage)

### 2. Implement SetupFlags Method
- [x] Add `--format` / `-o` string flag for output format (json/text, default: text)
- [x] Normalize flag values (merge short and long forms)
- [x] No other flags needed (restore is intentionally simple)

### 3. Implement Validate Method
- [x] Validate that no arguments are provided (restore only works on current ticket)
- [x] Validate format flag value (must be "json" or "text")
- [x] Store validated format for Execute method

### 4. Implement Execute Method
- [x] Get App instance using `cli.NewApp(ctx)` pattern
- [x] Call `app.RestoreCurrentTicket(ctx)` to restore the symlink
- [x] The method now returns `(*ticket.Ticket, error)` so use the returned ticket directly
- [x] Handle JSON output formatting
- [x] Format errors as JSON when format flag is set to json
- [x] Return appropriate success message in text mode

### 5. Add JSON Output Support
- [x] Define JSON output structure for symlink restoration
- [x] Include fields: ticket_id, status, symlink_restored, worktree_path
- [x] Include parent_ticket if available from ticket metadata
- [x] Format and marshal JSON response based on format flag

### 6. Create Comprehensive Unit Tests
- [x] Create `internal/cli/commands/restore_test.go`
- [x] Test command metadata (name, aliases, description)
- [x] Test flag setup and validation
- [x] Test no-arguments validation
- [x] Test JSON and text output formats
- [x] Test error cases with mock App
- [x] Achieve >80% test coverage

### 7. Integration Testing
- [ ] Test symlink restoration in a worktree
- [ ] Test error when not in a worktree
- [ ] Test error when branch doesn't correspond to a ticket
- [ ] Test JSON output format correctness
- [ ] Test when symlink already exists and is correct

### 8. Register Command and Clean Up
- [x] Register restore command in `main.go` command registry
- [x] Remove restore case from switch statement (around line 189)
- [x] Verify command appears in help text
- [x] Ensure handleRestore function is removed if no longer used

### 9. Code Quality and Documentation
- [x] Run `make test` to ensure all tests pass
- [x] Run `make vet` for static analysis
- [x] Run `make fmt` for code formatting
- [x] Run `make lint` for linting issues
- [x] Update `docs/COMMAND_MIGRATION_GUIDE.md` with restore command completion status
- [x] Document the zero-argument pattern for future commands

### 10. Final Verification
- [ ] Manual testing of symlink restoration functionality
- [ ] Verify symlink is created correctly pointing to the right ticket
- [ ] Ensure error messages are consistent with other commands
- [ ] Get developer approval before closing

## Implementation Notes

### Current Implementation Analysis
- Located in switch statement around line 189 in main.go
- Calls `handleRestore(ctx)` which is only 7 lines
- Simply delegates to `app.RestoreCurrentTicket(ctx)`
- No flags or arguments currently
- Error messages are already user-friendly

### Available App Method
The App struct provides one method for restoring the symlink:
- `RestoreCurrentTicket(ctx context.Context) (*ticket.Ticket, error)` - Restores the current-ticket.md symlink

**Note**: After the App refactoring (completed in ticket 250814-121422), this method now returns the ticket entity directly, so there's no need to re-fetch the ticket data for JSON output.

### Migration Requirements
1. **App Dependency**: Use `cli.NewApp(ctx)` directly to leverage existing `App.RestoreCurrentTicket` method
2. **No Positional Arguments**: Restore only works in current worktree context (enforce zero arguments)
3. **Simple Behavior**: Always restores the current-ticket.md symlink based on git branch
4. **Validation**:
   - No arguments allowed
   - Must be in a worktree
   - Branch must correspond to a valid ticket
5. **Output Formatting**:
   - Simple success message in text format
   - JSON output support for AI/tooling integration
   - Format flag (`--format`/`-o`) with json/text options
6. **Error Handling**: Clear, consistent error messages matching other commands

### Expected Behavior
- Validates that we're in a worktree context
- Determines ticket ID from current git branch
- Creates or fixes the current-ticket.md symlink
- Points symlink to the correct ticket file in tickets/doing/
- Returns structured JSON output when `--format json` is specified

### JSON Output Structure

```json
{
  "ticket_id": "250814-111507-migrate-restore-command",
  "status": "doing",
  "symlink_restored": true,
  "symlink_path": "current-ticket.md",
  "target_path": "tickets/doing/250814-111507-migrate-restore-command.md",
  "worktree_path": "../ticketflow.worktrees/250814-111507-migrate-restore-command",
  "parent_ticket": "250812-152927-migrate-remaining-commands",
  "message": "Current ticket symlink restored"
}
```

## Pattern Building on Previous Migrations

This migration builds on:
1. **Worktree Operations** (from `start`): Works within worktree context
2. **Zero Arguments Pattern**: First command with strictly no arguments
3. **Current Context Only**: Works only with current worktree context
4. **App Method Reuse**: Leverages existing `App.RestoreCurrentTicket`
5. **JSON Output** (from other commands): Consistent format flag pattern

## New Patterns to Establish

1. **Zero-Argument Command**: Strictly no positional arguments allowed
2. **Current-Only Operation**: No option to specify ticket ID
3. **Symlink Management Pattern**: Template for other symlink-related operations
4. **Simplified Command**: Intentionally minimal flags for clarity

## Implementation Code Examples

### Command Structure
```go
type RestoreCommand struct {
    format string
}
```

### Simple Execute Logic
```go
func (r *RestoreCommand) Execute(ctx context.Context) error {
    app, err := cli.NewApp(ctx)
    if err != nil {
        return err
    }

    // Perform the restore operation - now returns the ticket directly
    ticket, err := app.RestoreCurrentTicket(ctx)
    
    if err != nil {
        if r.format == FormatJSON {
            // Return error in JSON format
            return outputJSON(map[string]interface{}{
                "error": err.Error(),
                "success": false,
            })
        }
        return err
    }

    // For JSON output, use the returned ticket directly
    if r.format == FormatJSON {
        jsonData := map[string]interface{}{
            "ticket_id": ticket.ID,
            "status": string(ticket.Status),
            "symlink_restored": true,
            "symlink_path": "current-ticket.md",
            "target_path": fmt.Sprintf("tickets/doing/%s.md", ticket.ID),
            "message": "Current ticket symlink restored",
        }
        
        // Add optional fields if available
        if ticket.ParentID != "" {
            jsonData["parent_ticket"] = ticket.ParentID
        }
        
        return outputJSON(jsonData)
    }

    fmt.Println("✅ Current ticket symlink restored")
    return nil
}
```

### Note on App Method Refactoring
The App method refactoring (ticket 250814-121422) has been completed, so `RestoreCurrentTicket` now returns `(*ticket.Ticket, error)`. This means:
- The restore command can use the returned ticket directly for JSON output
- No need for re-fetching workarounds in the implementation
- Clean implementation pattern from the start

### Minimal Flag Setup
```go
func (r *RestoreCommand) SetupFlags(fs *flag.FlagSet) {
    fs.StringVar(&r.format, "format", "text", "Output format (json|text)")
    fs.StringVar(&r.format, "o", "text", "Output format (shorthand)")
}
```

### Strict Validation
```go
func (r *RestoreCommand) Validate(args []string) error {
    if len(args) > 0 {
        return fmt.Errorf("restore command does not accept any arguments")
    }
    
    if r.format != "json" && r.format != "text" {
        return fmt.Errorf("invalid format: %s (must be 'json' or 'text')", r.format)
    }
    
    return nil
}
```

## Estimated Time
**2-3 hours** based on:
- Simplest command implementation (30 minutes)
- Minimal flag handling (15 minutes)
- Basic validation logic (15 minutes)
- JSON output support (30 minutes)
- Comprehensive testing (1 hour)
- Integration and cleanup (30 minutes)

## Technical Considerations

1. **Worktree Context**: Must be executed from within a worktree
2. **Branch Validation**: Current git branch must correspond to a valid ticket
3. **Symlink Creation**: Handle cases where symlink already exists or is broken
4. **Path Resolution**: Correctly resolve relative paths for symlink target
5. **Error Messages**: Clear feedback when not in worktree or no corresponding ticket
6. **JSON Data**: Use the ticket returned by RestoreCurrentTicket for output

## Dependencies
- Builds on patterns from: `new`, `start`, `close` commands
- Related to: `start` command (both work with worktrees and symlinks)
- Will inform patterns for: other symlink management operations
- Leverages App refactoring completed in ticket 250814-121422

## Testing Strategy

### Unit Test Coverage Areas
1. **Command metadata**: Name, aliases, description validation
2. **Flag handling**: Format flag values
3. **Validation logic**: No arguments enforcement
4. **Mock App interactions**: Verify RestoreCurrentTicket called correctly
5. **Output formatting**: Both JSON and text modes
6. **Error scenarios**: Not in worktree, no corresponding ticket

### Integration Test Scenarios
1. Successful symlink restoration in worktree
2. Error when not in a worktree
3. Error when branch doesn't match any ticket
4. JSON output format verification
5. Behavior when symlink already exists and is correct
6. Behavior when symlink is broken or points to wrong file

## Implementation Dependencies

### Constants and Utilities
- Format constants (`FormatText`, `FormatJSON`) from `internal/cli/commands/new.go`
- Consider extracting to shared location
- Use `cli.ParseOutputFormat()` for format validation
- Need `getCurrentTicketID()` helper for JSON output

### Data Gathering for JSON Output
With the completed App refactoring, `RestoreCurrentTicket` returns `(*ticket.Ticket, error)`, so:
1. Use the returned ticket directly for all JSON fields
2. No need to re-fetch or determine ticket ID separately
3. Include symlink-specific information (paths, restoration status)
4. Add parent ticket ID if present in the ticket metadata

## References

- See `docs/COMMAND_MIGRATION_GUIDE.md` for migration patterns
- Review `internal/cli/commands/close.go` for inverse operation patterns
- Check current `handleRestore` implementation in main.go (around line 220)
- Command interface at `internal/command/interface.go`
- Format constants at `internal/cli/commands/new.go`
- Simplest command migration example for reference