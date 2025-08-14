---
priority: 2
description: "Migrate restore command to new Command interface"
created_at: "2025-08-14T11:15:07+09:00"
started_at: null
closed_at: null
related:
    - parent:250812-152927-migrate-remaining-commands
---

# Migrate restore command to new Command interface

Migrate the `restore` command to use the new Command interface, enabling recovery of accidentally closed tickets. This command is the inverse of `close` and completes the core ticket lifecycle management (new → start → close → restore).

## Why This Command Next?

1. **Simplest Remaining Command**: Only 7 lines of implementation, making it ideal for a quick win
2. **Completes Core Lifecycle**: Forms the essential workflow with new/start/close/restore
3. **High User Value**: Critical safety net for accidental closes
4. **Pattern Establishment**: Sets pattern for zero-argument, current-context commands
5. **Low Risk**: Minimal complexity reduces implementation risk

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

### 1. Create Command File and Structure
- [ ] Create `internal/cli/commands/restore.go` implementing the Command interface
- [ ] Define command struct (no flags needed for this simple command)
- [ ] Implement basic Command interface methods (Name, Aliases, Description, Usage)

### 2. Implement SetupFlags Method
- [ ] Add `--format` / `-o` string flag for output format (json/text, default: text)
- [ ] Normalize flag values (merge short and long forms)
- [ ] No other flags needed (restore is intentionally simple)

### 3. Implement Validate Method
- [ ] Validate that no arguments are provided (restore only works on current ticket)
- [ ] Validate format flag value (must be "json" or "text")
- [ ] Store validated format for Execute method

### 4. Implement Execute Method
- [ ] Get App instance using `cli.NewApp(ctx)` pattern
- [ ] Call `app.RestoreTicket(ctx)` to restore the current ticket
- [ ] After successful restore, retrieve ticket data for JSON formatting
- [ ] Handle JSON output formatting
- [ ] Format errors as JSON when format flag is set to json
- [ ] Return appropriate success message in text mode

### 5. Add JSON Output Support
- [ ] Define JSON output structure for restored ticket
- [ ] Include fields: ticket_id, status, restored_at, previous_status, duration_in_done
- [ ] Include worktree_path and parent_ticket if available
- [ ] Format and marshal JSON response based on format flag

### 6. Create Comprehensive Unit Tests
- [ ] Create `internal/cli/commands/restore_test.go`
- [ ] Test command metadata (name, aliases, description)
- [ ] Test flag setup and validation
- [ ] Test no-arguments validation
- [ ] Test JSON and text output formats
- [ ] Test error cases with mock App
- [ ] Achieve >80% test coverage

### 7. Integration Testing
- [ ] Test restore of recently closed ticket
- [ ] Test error when no current ticket exists
- [ ] Test error when ticket is not in done status
- [ ] Test JSON output format correctness
- [ ] Test that worktree is preserved after restore

### 8. Register Command and Clean Up
- [ ] Register restore command in `main.go` command registry
- [ ] Remove restore case from switch statement (around line 220)
- [ ] Verify command appears in help text
- [ ] Ensure handleRestore function is removed if no longer used

### 9. Code Quality and Documentation
- [ ] Run `make test` to ensure all tests pass
- [ ] Run `make vet` for static analysis
- [ ] Run `make fmt` for code formatting
- [ ] Run `make lint` for linting issues
- [ ] Update `docs/COMMAND_MIGRATION_GUIDE.md` with restore command completion status
- [ ] Document the zero-argument pattern for future commands

### 10. Final Verification
- [ ] Manual testing of restore functionality
- [ ] Verify ticket moves from done → doing correctly
- [ ] Ensure error messages are consistent with other commands
- [ ] Get developer approval before closing

## Implementation Notes

### Current Implementation Analysis
- Located in switch statement around line 220 in main.go
- Calls `handleRestore(ctx)` which is only 7 lines
- Simply delegates to `app.RestoreTicket(ctx)`
- No flags or arguments currently
- Error messages are already user-friendly

### Available App Method
The App struct provides one method for restoring tickets:
- `RestoreTicket(ctx context.Context) error` - Restores current ticket from done → doing

**Note**: This method only returns an error. For JSON output, the command must:
- Call `app.RestoreTicket(ctx)`
- If successful, retrieve ticket information using ticket manager
- Format and return the JSON response

### Migration Requirements
1. **App Dependency**: Use `cli.NewApp(ctx)` directly to leverage existing `App.RestoreTicket` method
2. **No Positional Arguments**: Restore only works on current ticket (enforce zero arguments)
3. **Simple Behavior**: Always restores current ticket from done → doing
4. **Validation**:
   - No arguments allowed
   - Current ticket must exist
   - Ticket must be in done status
5. **Output Formatting**:
   - Simple success message in text format
   - JSON output support for AI/tooling integration
   - Format flag (`--format`/`-o`) with json/text options
6. **Error Handling**: Clear, consistent error messages matching other commands

### Expected Behavior
- Validates current ticket exists and is in done status
- Moves ticket from done → doing status
- Updates timestamps appropriately
- Preserves worktree if it exists
- Returns structured JSON output when `--format json` is specified

### JSON Output Structure

```json
{
  "ticket_id": "250814-111507-migrate-restore-command",
  "status": "doing",
  "restored_at": "2025-08-14T12:00:00+09:00",
  "previous_status": "done",
  "duration_in_done": "2h30m",
  "worktree_path": "../ticketflow.worktrees/250814-111507-migrate-restore-command",
  "parent_ticket": "250812-152927-migrate-remaining-commands",
  "message": "Ticket restored successfully"
}
```

## Pattern Building on Previous Migrations

This migration builds on:
1. **State Modification** (from `new`, `start`, `close`): Changes ticket state
2. **Zero Arguments Pattern**: First command with strictly no arguments
3. **Current Context Only**: Works only with current ticket in worktree
4. **App Method Reuse**: Leverages existing `App.RestoreTicket`
5. **JSON Output** (from other commands): Consistent format flag pattern

## New Patterns to Establish

1. **Zero-Argument Command**: Strictly no positional arguments allowed
2. **Current-Only Operation**: No option to specify ticket ID
3. **Recovery/Undo Pattern**: Template for other undo operations
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

    // Perform the restore operation
    err = app.RestoreTicket(ctx)
    
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

    // For JSON output, gather ticket info after successful restore
    if r.format == FormatJSON {
        // NOTE: Currently requires re-fetching ticket data
        // This will be improved when App methods return entities (see refactoring ticket)
        ticketID := getCurrentTicketID()
        ticket, _ := app.TicketManager.GetTicket(ticketID)
        
        jsonData := map[string]interface{}{
            "ticket_id": ticket.ID,
            "status": "doing",
            "restored_at": time.Now().Format(time.RFC3339),
            "previous_status": "done",
            // Calculate other fields...
        }
        return outputJSON(jsonData)
    }

    fmt.Println("✅ Ticket restored successfully")
    return nil
}
```

### Note on App Method Refactoring
Based on architectural discussions during the close command implementation, we've decided to refactor App methods to return primary entities in a separate ticket. This will eliminate the need to re-fetch ticket data for JSON output. The restore command will be updated as part of that refactoring effort.

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

1. **Current Ticket Detection**: Must determine if in a worktree with current ticket
2. **State Validation**: Ensure ticket is actually in done status
3. **Worktree Preservation**: Verify worktree remains intact after restore
4. **Timestamp Updates**: Handle appropriate timestamp modifications
5. **Error Messages**: Clear feedback when no ticket or wrong status
6. **JSON Data Gathering**: Retrieve ticket info post-restore for JSON output

## Dependencies
- Builds on patterns from: `new`, `start`, `close` commands
- Pairs with: `close` command (inverse operation)
- Will inform patterns for: other recovery/undo operations
- Related to state management established in previous migrations

## Testing Strategy

### Unit Test Coverage Areas
1. **Command metadata**: Name, aliases, description validation
2. **Flag handling**: Format flag values
3. **Validation logic**: No arguments enforcement
4. **Mock App interactions**: Verify RestoreTicket called correctly
5. **Output formatting**: Both JSON and text modes
6. **Error scenarios**: No current ticket, wrong status

### Integration Test Scenarios
1. Restore recently closed ticket
2. Error when no current ticket
3. Error when ticket not in done status
4. JSON output format verification
5. Worktree preservation check

## Implementation Dependencies

### Constants and Utilities
- Format constants (`FormatText`, `FormatJSON`) from `internal/cli/commands/new.go`
- Consider extracting to shared location
- Use `cli.ParseOutputFormat()` for format validation
- Need `getCurrentTicketID()` helper for JSON output

### Data Gathering for JSON Output
Since App.RestoreTicket doesn't return ticket data, the command must:
1. Determine which ticket was restored
2. Use `app.TicketManager.GetTicket()` to retrieve the restored ticket
3. Calculate duration in done status from timestamps
4. Include worktree path if available

## References

- See `docs/COMMAND_MIGRATION_GUIDE.md` for migration patterns
- Review `internal/cli/commands/close.go` for inverse operation patterns
- Check current `handleRestore` implementation in main.go (around line 220)
- Command interface at `internal/command/interface.go`
- Format constants at `internal/cli/commands/new.go`
- Simplest command migration example for reference