---
priority: 2
description: Migrate cleanup command to new Command interface
created_at: "2025-08-14T18:11:07+09:00"
started_at: "2025-08-14T21:54:18+09:00"
closed_at: "2025-08-15T10:29:23+09:00"
related:
    - parent:250812-152927-migrate-remaining-commands
---

# Migrate Cleanup Command to New Interface

Migrate the cleanup command from the old switch-based system to the new Command interface pattern, following the established migration guide.

## Context

The cleanup command is a dual-mode command that handles both automatic cleanup of orphaned resources and specific ticket cleanup. It's one of the last remaining commands that needs to be migrated to the new Command interface as part of the architectural refactoring effort.

## Implementation Details

The cleanup command operates in two distinct modes:

### Mode 1: Auto-Cleanup (no arguments)
- Triggered by: `ticketflow cleanup` or `ticketflow cleanup --dry-run`
- Automatically detects and removes orphaned worktrees (worktrees without active tickets)
- Cleans up stale branches (branches for done tickets without worktrees)
- Supports `--dry-run` flag to preview what would be cleaned without making changes
- Returns `CleanupResult` with statistics about cleaned resources

### Mode 2: Ticket-Specific Cleanup (with ticket ID)
- Triggered by: `ticketflow cleanup <ticket-id>` or `ticketflow cleanup --force <ticket-id>`
- Validates the specified ticket exists and is closed (in done/ directory)
- Removes the associated git worktree if it exists
- Removes the associated git branch if it exists
- Supports `--force` flag to skip confirmation prompts
- Currently doesn't return the cleaned ticket (should be refactored to follow new pattern)

## Tasks

- [x] Create internal/cli/cleanup_command.go implementing the Command interface
- [x] Define cleanupFlags struct with: dryRun, force, format, formatShort fields
- [x] Implement dual-mode Execute method (route based on presence of ticket ID in args)
- [x] Move auto-cleanup logic (AutoCleanup, CleanupStats) to work with new command
- [x] Move ticket-specific cleanup logic (CleanupTicket) to work with new command
- [x] Add JSON output support for both modes (CleanupResult and ticket cleanup)
- [x] Refactor CleanupTicket to return (*ticket.Ticket, error) for consistency with new pattern
- [x] Store args in flags struct during Validate (follow close command pattern)
- [x] Update main.go to register the new cleanup command
- [x] Remove old cleanup case from switch statement in main.go
- [x] Add/update tests for both auto-cleanup and ticket-specific modes
- [x] Test JSON output formatting for both modes
- [x] Run `make test` to ensure all tests pass
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Update COMMAND_MIGRATION_GUIDE.md to mark cleanup as migrated
- [x] Update parent ticket (250812-152927-migrate-remaining-commands) to mark this task as complete
- [x] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Implementation Notes

### Command Structure Pattern
Follow the close command's dual-mode pattern since it has nearly identical behavior:
1. Check for ticket ID in args during Execute
2. Route to appropriate App method based on mode
3. Handle different return types for each mode

### Key Implementation Points
- Use the normalize() helper for flag parsing (handles short/long format flags)
- The Execute method should check context cancellation at the start
- Validate format flag values in Validate() method
- Handle JSON errors appropriately for both CleanupResult and ticket objects
- Preserve existing confirmation prompt logic (handled within App methods)
- Auto-cleanup's dry-run mode shows statistics via CleanupStats() then returns

### Files to Reference
- `internal/cli/cleanup.go` - Current implementation with both AutoCleanup and CleanupTicket
- `internal/cli/helpers.go` - Contains normalize() and other helper functions
- `cmd/ticketflow/main.go` - Shows current dual-mode routing in Execute function
- COMMAND_MIGRATION_GUIDE.md - Standard migration patterns and requirements

### Testing Considerations
- Integration tests exist in `test/integration/cleanup_extended_test.go`
- Unit tests exist in `internal/cli/cleanup_test.go`
- Ensure both modes are tested with new command structure
- Test JSON output for both CleanupResult and ticket responses
- Verify dry-run mode doesn't make actual changes

## Notes

The cleanup command is more complex than initially described due to its dual-mode nature, but the migration is straightforward because similar patterns are already established from previous command migrations. The key is properly handling the two distinct operational modes and their different return types.

## Migration Insights

### Key Learnings

1. **Dual-Mode Pattern**: The cleanup command successfully follows the dual-mode pattern established by the close command, confirming this is a robust pattern for commands that can operate with or without arguments.

2. **Return Value Refactoring**: Successfully refactored `CleanupTicket` to return `(*ticket.Ticket, error)` for consistency with the new pattern. This required updating both the implementation and all test files.

3. **JSON Output Complexity**: Different return types (CleanupResult vs ticket.Ticket) required separate JSON output functions, highlighting the importance of type-specific formatting.

4. **Test Updates**: The return value change required updating multiple test files, demonstrating the importance of comprehensive test coverage when refactoring.

### Implementation Details

- Created `internal/cli/commands/cleanup.go` with full Command interface implementation
- Implemented separate execution paths for auto-cleanup and ticket-specific cleanup
- Added comprehensive JSON output support for both modes
- Successfully removed old cleanup implementation from main.go
- All existing tests pass with minimal modifications

### Challenges Resolved

1. **Type Errors**: Initial implementation had incorrect field references (e.g., `t.Title` instead of `t.Description`), resolved by checking actual struct definitions.

2. **Status String Conversion**: The Status type needed direct string conversion rather than a `.String()` method.

3. **Test Compilation**: Multiple test files needed updates to handle the new return signature of `CleanupTicket`.

### Migration Complete

The cleanup command has been successfully migrated to the new Command interface with all functionality preserved and enhanced with JSON output support.

## Code Review Results (golang-pro)

The implementation has been thoroughly reviewed by the golang-pro agent and received **full approval** with no issues found:

### ✅ Review Passed All Criteria:
- **Code Quality**: Follows Go idioms perfectly with clean separation of concerns
- **Error Handling**: Comprehensive and robust with proper error wrapping
- **Test Coverage**: Well-tested with both unit and integration tests
- **Consistency**: Perfectly aligned with other migrated commands
- **Performance**: Efficient implementation with no unnecessary allocations

### Key Strengths Confirmed:
- Proper context handling for cancellation support
- Safe type assertions with appropriate error returns
- Dual-mode operation correctly supports both auto-cleanup and ticket-specific cleanup
- Flag normalization follows established patterns (short/long forms)
- JSON and text output properly formatted for both modes
- Return value refactoring handled correctly across all files

### Final Status:
✅ **Production Ready** - The cleanup command migration is complete, tested, reviewed, and ready for deployment.