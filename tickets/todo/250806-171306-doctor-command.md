---
priority: 2
description: "Implement ticketflow doctor command for manual worktree recovery and diagnostics"
created_at: "2025-08-06T17:13:06+09:00"
started_at: null
closed_at: null
related:
    - parent:250803-121506-worktree-recovery
    - depends_on:250806-171131-worktree-error-detection
---

# Doctor Command Implementation

## Overview
Implement a `ticketflow doctor` command that provides manual recovery and diagnostic capabilities for worktree issues. This command will help users diagnose and fix complex worktree problems that automatic recovery cannot handle.

## Tasks
- [ ] Create doctor command infrastructure in `cmd/ticketflow/`
- [ ] Implement `internal/cli/doctor.go` with doctor operations
- [ ] Add `--fix-worktrees` flag for worktree recovery
- [ ] Add `--check-only` flag for diagnostic mode
- [ ] Implement orphaned worktree directory detection
- [ ] Add `--verbose` flag for detailed output
- [ ] Create comprehensive test suite
- [ ] Add documentation and help text
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Get developer approval before closing

## Technical Details
### Command Structure
```
ticketflow doctor [flags]

Flags:
  --fix-worktrees   Fix corrupted worktree references
  --check-only      Check for issues without fixing
  --verbose         Show detailed diagnostic information
  --json           Output results in JSON format
```

### Doctor Operations
1. **Check Mode** (`--check-only`):
   - List all worktrees and their status
   - Detect orphaned worktree directories
   - Identify corrupted .git/worktrees entries
   - Check for branch/worktree mismatches
   - Validate ticket file locations

2. **Fix Mode** (`--fix-worktrees`):
   - Run `git worktree prune` for corrupted entries
   - Clean orphaned worktree directories (with confirmation)
   - Fix branch references
   - Recover lost ticket files

### Implementation Files
- `cmd/ticketflow/handlers_doctor.go` - Command handler
- `internal/cli/doctor.go` - Core doctor logic
- `internal/cli/doctor_test.go` - Tests

### Diagnostic Output Format
```
Checking worktree health...

✓ 3 healthy worktrees found
⚠ 1 orphaned directory found:
  - ../ticketflow.worktrees/old-ticket
✗ 1 corrupted worktree reference:
  - feature-branch (directory missing)

Run with --fix-worktrees to repair issues.
```

### JSON Output Structure
```json
{
  "healthy_worktrees": 3,
  "issues": [
    {
      "type": "orphaned_directory",
      "path": "../ticketflow.worktrees/old-ticket",
      "fixable": true
    },
    {
      "type": "corrupted_reference",
      "branch": "feature-branch",
      "fixable": true
    }
  ],
  "summary": "2 issues found, all fixable"
}
```

## Acceptance Criteria
- [ ] Doctor command correctly identifies all worktree issues
- [ ] Fix mode safely repairs corrupted worktrees
- [ ] Check-only mode doesn't modify anything
- [ ] Clear user feedback for all operations
- [ ] JSON output works for automation
- [ ] Verbose mode provides debugging information
- [ ] No data loss during fix operations
- [ ] Comprehensive test coverage (>80%)

## Notes
This is phase 3 of the worktree recovery implementation. The doctor command provides manual control over recovery operations for cases where automatic recovery is not appropriate or has failed. It should be safe by default, requiring explicit flags for any destructive operations.