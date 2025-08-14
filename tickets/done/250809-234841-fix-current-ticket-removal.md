---
priority: 2
description: Fix current-ticket.md removal without verification when closing tickets
created_at: "2025-08-09T23:48:41+09:00"
started_at: "2025-08-14T16:55:38+09:00"
closed_at: "2025-08-14T17:51:58+09:00"
---

# Fix current-ticket.md removal without verification

## Problem
When closing a ticket provided as a CLI parameter, ticketflow currently removes current-ticket.md without checking if the parameter ticket matches what current-ticket.md is pointing to. This can lead to incorrect removal of the current ticket symlink.

**Bug Location**: `internal/cli/commands.go:1319` in `moveTicketToDoneWithReason` function

Example scenario:
1. User has current-ticket.md pointing to ticket-A
2. User runs `ticketflow close ticket-B`
3. current-ticket.md gets removed even though it was pointing to ticket-A, not ticket-B

## Root Cause
The `moveTicketToDoneWithReason` function unconditionally calls `app.Manager.SetCurrentTicket(ctx, nil)` which removes the current-ticket.md symlink regardless of which ticket is being closed.

## Solution
**Simplified Approach**: Add an `isCurrentTicket` boolean parameter to `moveTicketToDoneWithReason` function.

The calling code already knows whether it's closing the current ticket:
- `closeCurrentTicketInternal` → pass `true`
- `closeAndCommitTicket` (via `CloseTicketByID`) → pass `false`

Only remove current-ticket.md when `isCurrentTicket == true`.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [x] Modify `moveTicketToDoneWithReason` in `internal/cli/commands.go` to accept `isCurrentTicket bool` parameter
- [x] Update the symlink removal logic to only execute when `isCurrentTicket == true`
- [x] Update `closeCurrentTicketInternal` to pass `true` for isCurrentTicket
- [x] Update `closeAndCommitTicket` to pass `false` for isCurrentTicket
- [x] Add unit test for `CloseTicketByID` when closing non-current ticket
- [x] Add integration test verifying current-ticket.md preservation when closing other tickets
- [x] Test edge cases (no current-ticket.md, broken symlink, etc.)
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Update the ticket with insights from resolving this ticket
- [x] Code review by golang-pro agent - PASSED with no issues
- [x] Create pull request (#66)
- [x] Address CI failures (linting issues)
- [x] Address Copilot review comments
- [ ] Get developer approval before closing

## Implementation Notes
- The simpler parameter-based approach is preferred over symlink checking
- This bug affects both worktree and non-worktree modes
- Consider adding debug logging when preserving current-ticket.md
- Currently no test coverage for this scenario - that's why the bug wasn't caught

## Testing Gap Identified
There are no existing tests for closing a ticket by ID when it's not the current ticket. This gap in test coverage allowed this bug to slip through.

## Resolution Insights

### Bug Analysis
The root cause was a simple logic error where `moveTicketToDoneWithReason` unconditionally removed the current-ticket.md symlink whenever ANY ticket was closed. This happened because the function didn't know whether it was closing the current ticket or a different one.

### Solution Approach
Instead of the initially proposed symlink checking approach, we implemented a simpler solution:
- Added an `isCurrentTicket` boolean parameter to `moveTicketToDoneWithReason`
- The calling functions already know whether they're closing the current ticket, so they pass the appropriate value
- This avoids the complexity of symlink reading and comparison

### Testing Improvements
Added comprehensive test coverage:
1. **Unit Test**: Verifies that `SetCurrentTicket(nil)` is NOT called when closing a non-current ticket
2. **Integration Tests**: End-to-end verification of three scenarios:
   - Preserving current-ticket.md when closing a different ticket
   - Removing current-ticket.md when closing the current ticket
   - Graceful handling when current-ticket.md doesn't exist

### Lessons Learned
1. **Simplicity wins**: The parameter-based approach is cleaner than symlink checking
2. **Test coverage gaps**: Missing tests for common scenarios can hide bugs
3. **Clear separation of concerns**: Functions should know their context (current vs non-current)
4. **Integration tests are valuable**: They catch real-world usage patterns that unit tests might miss

## Code Review Results

### Golang-Pro Agent Review - PASSED ✅
The code changes were reviewed by the golang-pro agent with focus on:
- Code quality and Go idioms
- Logic correctness
- Test coverage
- Performance considerations
- Edge case handling
- Best practices

**Review Summary:**
- **No issues found** - Implementation is correct and production-ready
- Code is clean, idiomatic Go following best practices
- Fix properly addresses the bug with minimal changes
- Tests are comprehensive covering all scenarios
- All edge cases handled (missing symlinks, worktree modes)
- No performance concerns (simple boolean check)

**Optional architectural note:** The reviewer noted that when `CloseTicketByID` is called with the current ticket's ID, it delegates to the normal close flow. This is functionally correct and maintains consistency with existing patterns.

## Implementation Summary

### Changes Made
1. **Core Fix**: Added `isCurrentTicket` boolean parameter to `moveTicketToDoneWithReason`
2. **Conditional Logic**: Only remove current-ticket.md when `isCurrentTicket == true`
3. **Caller Updates**: 
   - `closeCurrentTicketInternal` passes `true`
   - `closeAndCommitTicket` passes `false`
4. **Test Coverage**: Added unit and integration tests for all scenarios
5. **Documentation**: Updated ticket with insights and lessons learned

### Commits
- `refine: Update ticket with simplified implementation approach`
- `fix: Add isCurrentTicket parameter to prevent incorrect symlink removal`
- `test: Add unit test for preserving current-ticket.md`
- `test: Add integration tests for current-ticket.md preservation`
- `test: Fix unit and integration tests`
- `style: Apply go fmt formatting`
- `docs: Update ticket with implementation insights`

### Status
**✅ Implementation complete and verified**
- All tasks completed
- Tests passing
- Code reviewed and approved by automated review
- Pull Request #66 created and passing all CI checks
- Ready for developer approval and merge

## Pull Request Review Process

### Copilot Review Feedback
The PR received automated review from GitHub Copilot with the following suggestions:

1. **Linting Issues** (Fixed):
   - QF1003: Use switch statement instead of if-else chain
   - errcheck: Handle os.Remove error return value

2. **Code Comments** (Improved):
   - Clarified why `closeCurrentTicketInternal` passes true
   - Clarified why `CloseTicketByID` passes false

3. **Architectural Suggestions** (Considered but not changed):
   - **os.Chdir usage**: Kept existing pattern as it's used throughout integration tests and documented as a known limitation
   - **NewAppWithWorkingDir signature**: Kept as-is since it's an existing test helper used across the codebase

### CI/CD Process
- Initial CI run failed due to linting issues
- Fixed linting issues with proper switch statement and error handling
- All subsequent CI runs passed successfully
- Final status: ✅ Lint PASS, ✅ Test PASS

### Additional Insights
1. **Importance of CI linting**: Caught style issues that weren't detected locally
2. **Value of automated reviews**: Copilot provided useful suggestions for code clarity
3. **Pattern consistency**: Important to follow existing codebase patterns even if not ideal
4. **Comment clarity**: Explicit comments about boolean parameters improve maintainability