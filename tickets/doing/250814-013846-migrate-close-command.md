---
priority: 2
description: Migrate close command to new Command interface
created_at: "2025-08-14T01:38:46+09:00"
started_at: "2025-08-14T09:59:44+09:00"
closed_at: null
related:
    - parent:250812-152927-migrate-remaining-commands
---

# Migrate close command to new Command interface

Migrate the `close` command to use the new Command interface, completing the core ticket lifecycle commands (new → start → close). This command closes tickets and handles the final state transition.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

### 1. Create Command File and Structure
- [x] Create `internal/cli/commands/close.go` implementing the Command interface
- [x] Define command struct with fields for force, reason, and format flags
- [x] Implement basic Command interface methods (Name, Aliases, Description, Usage)

### 2. Implement SetupFlags Method
- [x] Add `--force` / `-f` boolean flag to force close with uncommitted changes
- [x] Add `--reason` string flag for closing reason (added to commit message)
- [x] Add `--format` / `-o` string flag for output format (json/text, default: text)
- [x] Normalize flag values (merge short and long forms)

### 3. Implement Validate Method
- [x] Validate argument count (0 or 1 arguments allowed)
- [x] Validate format flag value (must be "json" or "text")
- [x] Store validated arguments for Execute method

### 4. Implement Execute Method with Dual-Mode Logic
- [x] Get App instance using `cli.NewApp(ctx)` pattern
- [x] Implement dual behavior:
  - [x] No args: Call `app.CloseTicket(ctx, force)` or `app.CloseTicketWithReason(ctx, reason, force)`
  - [x] With ID: Call `app.CloseTicketByID(ctx, ticketID, reason, force)`
- [x] After successful close, retrieve ticket data for JSON formatting
- [x] Handle gathering of duration, parent info, and other metadata for JSON output
- [x] Format errors as JSON when format flag is set to json
- [x] Return appropriate error messages

### 5. Add JSON Output Support
- [x] Define JSON output structures for both modes (current ticket vs by ID)
- [x] Include fields: ticket_id, status, closed_at, close_reason, commit_created, force_used
- [x] For current ticket mode: also include duration, parent_ticket, worktree_path
- [x] Format and marshal JSON response based on format flag

### 6. Create Comprehensive Unit Tests
- [ ] Create `internal/cli/commands/close_test.go`
- [ ] Test command metadata (name, aliases, description)
- [ ] Test flag setup and validation
- [ ] Test dual-mode behavior (no args vs with ID)
- [ ] Test JSON and text output formats
- [ ] Test error cases with mock App
- [ ] Achieve >80% test coverage

### 7. Integration Testing
- [ ] Test close current ticket in worktree
- [ ] Test close specific ticket by ID
- [ ] Test error when ticket not found
- [ ] Test error when ticket already closed
- [ ] Test force close with uncommitted changes
- [ ] Test close with reason provided
- [ ] Test JSON output format correctness
- [ ] Test no current ticket error (when no args in main repo)

### 8. Register Command and Clean Up
- [ ] Register close command in `main.go` command registry
- [ ] Remove close case from switch statement (lines 202-219)
- [ ] Remove `handleClose` function if no longer used
- [ ] Verify command appears in help text

### 9. Code Quality and Documentation
- [ ] Run `make test` to ensure all tests pass
- [ ] Run `make vet` for static analysis
- [ ] Run `make fmt` for code formatting
- [ ] Run `make lint` for linting issues
- [ ] Update `docs/COMMAND_MIGRATION_GUIDE.md` with close command completion status
- [ ] Add any new patterns discovered to the guide

### 10. Final Verification
- [ ] Manual testing of all command variations
- [ ] Verify backward compatibility
- [ ] Ensure error messages are consistent with other commands
- [ ] Get developer approval before closing

## Implementation Notes

### Current Implementation Analysis
- Located in switch statement around lines 202-219 in main.go
- Calls `handleClose(ctx, ticketID, force, reason)` which delegates to App methods
- Takes optional argument: ticket ID (if not provided, uses current ticket)
- Has flags:
  - `--force` / `-f` for forcing close with uncommitted changes
  - `--reason` for adding close reason to commit message
- Currently lacks JSON output support

### Available App Methods
The App struct provides three methods for closing tickets (all return only errors, not ticket data):
1. `CloseTicket(ctx context.Context, force bool) error` - Closes current ticket in worktree
2. `CloseTicketWithReason(ctx context.Context, reason string, force bool) error` - Closes current ticket with reason
3. `CloseTicketByID(ctx context.Context, ticketID, reason string, force bool) error` - Closes specific ticket by ID

**Note**: These methods only return errors. For JSON output, the command must:
- Call the appropriate close method
- If successful, retrieve ticket information separately using ticket manager
- Format and return the JSON response

### Migration Requirements
1. **App Dependency**: Use `cli.NewApp(ctx)` directly to leverage existing `App.CloseTicket` method
2. **Optional Positional Arguments**: Handle 0 or 1 arguments (MinArgs: 0, MaxArgs: 1)
3. **Dual Behavior**:
   - No args: Detect and close current ticket from worktree
   - With ID: Close specified ticket
4. **Validation**:
   - Ticket exists and is not already closed
   - Check for uncommitted changes (unless forced)
   - Current ticket detection when no ID provided
5. **Output Formatting**:
   - Consistent success/error messages in text format
   - JSON output support for AI/tooling integration
   - Format flag (`--format`/`-o`) with json/text options
6. **Commit Integration**: Handle reason flag for commit message enhancement
7. **Error Handling**: Clear, consistent error messages matching other commands

### Expected Behavior
- Validates ticket exists and is not already closed
- Moves ticket from doing → done status
- Sets closed_at timestamp
- Creates "Close ticket" commit (with optional reason)
- Handles uncommitted changes check (bypassable with --force)
- Returns structured JSON output when `--format json` is specified

### JSON Output Structures

For closing current ticket (no args):
```json
{
  "ticket_id": "250814-013846-migrate-close-command",
  "status": "done",
  "closed_at": "2025-08-14T02:00:00+09:00",
  "close_reason": "Migration completed successfully",
  "duration": "2h30m",
  "parent_ticket": "250812-152927-migrate-remaining-commands",
  "worktree_path": "../ticketflow.worktrees/250814-013846-migrate-close-command",
  "commit_created": true,
  "force_used": false
}
```

For closing by ID:
```json
{
  "ticket_id": "250813-123456-some-feature",
  "status": "done",
  "closed_at": "2025-08-14T02:00:00+09:00",
  "close_reason": "Feature completed",
  "branch": "250813-123456-some-feature",
  "commit_created": true,
  "force_used": false
}
```

## Pattern Building on Previous Migrations

This migration builds on:
1. **State Modification** (from `new`, `start`): Changes ticket state and filesystem
2. **Optional Positional Arguments**: New pattern - 0 or 1 arguments
3. **Boolean Flags** (from `start`): --force flag with both long and short forms
4. **String Flags**: --reason flag for additional context
5. **App Method Reuse**: Leverages existing `App.CloseTicket`
6. **JSON Output** (from other commands): Consistent format flag pattern
7. **Current Context Detection**: Determining current ticket from worktree

## New Patterns to Establish

1. **Optional Positional Arguments**: First command with 0-1 positional args
2. **Current Context Detection**: Inferring ticket from current directory
3. **Commit Message Enhancement**: Adding custom reason to commits
4. **Dual Mode Operation**: Different behavior based on argument presence
5. **Uncommitted Changes Validation**: Safety check with force override

## Implementation Code Examples

### Command Structure
```go
type CloseCommand struct {
    force  bool
    reason string
    format string
    args   []string
}
```

### Dual-Mode Execute Logic (Corrected)
```go
func (c *CloseCommand) Execute(ctx context.Context) error {
    app, err := cli.NewApp(ctx)
    if err != nil {
        return err
    }

    // Perform the close operation
    if len(c.args) == 0 {
        // Close current ticket
        if c.reason != "" {
            err = app.CloseTicketWithReason(ctx, c.reason, c.force)
        } else {
            err = app.CloseTicket(ctx, c.force)
        }
    } else {
        // Close specific ticket by ID
        err = app.CloseTicketByID(ctx, c.args[0], c.reason, c.force)
    }

    if err != nil {
        if c.format == FormatJSON {
            // Return error in JSON format
            return outputJSON(map[string]interface{}{
                "error": err.Error(),
                "success": false,
            })
        }
        return err
    }

    // For JSON output, gather ticket info after successful close
    if c.format == FormatJSON {
        // Retrieve closed ticket information
        ticketID := c.args[0]
        if len(c.args) == 0 {
            // Get current ticket ID for no-args mode
            ticketID = getCurrentTicketID()
        }
        
        // Read ticket data and build JSON response
        ticket, _ := app.TicketManager.GetTicket(ticketID)
        jsonData := buildJSONResponse(ticket, c.force, c.reason)
        return outputJSON(jsonData)
    }

    return nil
}
```

### Flag Normalization Pattern
```go
func (c *CloseCommand) SetupFlags(fs *flag.FlagSet) {
    fs.BoolVar(&c.force, "force", false, "Force close even with uncommitted changes")
    fs.BoolVar(&c.force, "f", false, "Force close even with uncommitted changes (shorthand)")
    fs.StringVar(&c.reason, "reason", "", "Reason for closing the ticket")
    fs.StringVar(&c.format, "format", "text", "Output format (json|text)")
    fs.StringVar(&c.format, "o", "text", "Output format (shorthand)")
}
```

## Estimated Time
**4-6 hours** based on:
- Similar complexity to `start` command (2-3 hours base)
- Additional complexity from dual behavior modes (+1 hour)
- Comprehensive testing requirements (+1-2 hours)
- Current ticket detection logic and validation

## Why This Command Next?

1. **Completes Core Lifecycle**: new → start → close forms the essential workflow
2. **Natural Progression**: Logical next step after `start` command
3. **Moderate Complexity**: Right level of challenge after `start`
4. **High User Value**: Frequently used command in daily workflow
5. **Unblocks Related Commands**: `restore` naturally follows `close`
6. **Pattern Completion**: Establishes remaining state-change patterns

## Technical Considerations

1. **Current Ticket Detection**: Determine ticket from worktree directory
2. **Git Status Check**: Validate clean working directory
3. **Force Flag Handling**: Override safety checks when forced
4. **Reason Integration**: Append reason to commit message properly
5. **State Transitions**: Atomic updates to ticket status
6. **Dual Mode Logic**: Handle both current and specific ticket paths
7. **Error Recovery**: Handle partial state changes gracefully
8. **Testing Complexity**: Mock both modes of operation
9. **JSON Output Structure**: Consistent schema for both modes
10. **Error Message Consistency**: Match format from other migrated commands

## Dependencies
- Builds on patterns from: `new`, `start`, `show` commands
- Will inform patterns for: `restore`, `cleanup` commands
- Related to worktree management established in `start`

## Testing Strategy

### Unit Test Coverage Areas
1. **Command metadata**: Name, aliases, description validation
2. **Flag handling**: Force, reason, format flag combinations
3. **Validation logic**: Argument count, format values
4. **Mock App interactions**: Verify correct methods called with right parameters
5. **Output formatting**: Both JSON and text modes
6. **Error scenarios**: Missing ticket, already closed, uncommitted changes

### Integration Test Scenarios
1. Close current ticket from within worktree
2. Close specific ticket by ID from main repo
3. Force close with uncommitted changes
4. Close with custom reason
5. Error handling for edge cases

## Implementation Dependencies

### Constants and Utilities
- Format constants (`FormatText`, `FormatJSON`) are currently defined in `internal/cli/commands/new.go`
  - Consider importing from there or extracting to a shared location like `internal/cli/output.go`
- Use `cli.ParseOutputFormat()` to parse and validate the format flag value
- The `outputJSON()` helper function will need to be implemented or imported

### Data Gathering for JSON Output
Since App methods don't return ticket data, the command must:
1. Track which ticket was closed (especially for no-args mode)
2. Use `app.TicketManager.GetTicket()` to retrieve the closed ticket
3. Calculate derived fields like duration from timestamps
4. Determine worktree path for current ticket mode

## References

- See `docs/COMMAND_MIGRATION_GUIDE.md` for migration patterns
- Review `internal/cli/commands/start.go` for similar state-changing patterns
- Check current `handleClose` implementation in main.go (lines 202-219, 299-306)
- Review `TestApp_CloseTicket_WithMocks` for test patterns
- Start command PR #62 for reference on state-changing commands with force flag
- Command interface at `internal/command/interface.go`
- Format constants currently at `internal/cli/commands/new.go`

## Progress and Insights

### Current Status (2025-08-14)
- **Ticket Refinement**: ✅ Completed comprehensive refinement based on codebase analysis
- **Technical Review**: ✅ Addressed issues identified by golang-pro review
- **PR Created**: ✅ PR #63 created with refined ticket ready for implementation
- **Next Ticket**: ✅ Created restore command migration ticket (250814-111507-migrate-restore-command)

### Key Insights from Analysis

1. **App Methods Don't Return Data**: 
   - Discovered that App methods only return errors, not structured data
   - Commands must retrieve ticket data separately for JSON output
   - This pattern affects all state-changing commands

2. **JSON Output Complexity**:
   - JSON formatting requires post-operation data gathering
   - Need helper functions like `getCurrentTicketID()` and `buildJSONResponse()`
   - Consider creating shared utilities for common JSON operations

3. **Dual-Mode Pattern**:
   - First command with optional positional arguments (0 or 1)
   - Establishes pattern for context-aware operations
   - Complexity comes from handling both current ticket and by-ID modes

4. **Constants Organization**:
   - Format constants scattered across commands (new.go)
   - Should be consolidated in a shared location
   - Pattern affects all commands needing output formatting

5. **Testing Considerations**:
   - Dual-mode requires comprehensive test coverage
   - Mock complexity increases with optional arguments
   - Integration tests crucial for verifying both modes

### Implementation Recommendations

1. **Start with shared utilities**: Extract format constants and JSON helpers first
2. **Implement simple path first**: Get no-args mode working before by-ID mode  
3. **Test incrementally**: Build tests alongside implementation
4. **Document patterns**: Update migration guide with dual-mode pattern

### Time Estimate Adjustment
- Original: 3-4 hours
- Revised: 4-6 hours (due to JSON data gathering complexity)