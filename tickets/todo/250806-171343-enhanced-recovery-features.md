---
priority: 3
description: "Add advanced recovery features for complex worktree corruption scenarios"
created_at: "2025-08-06T17:13:43+09:00"
started_at: null
closed_at: null
related:
    - parent:250803-121506-worktree-recovery
    - depends_on:250806-171131-worktree-error-detection
    - depends_on:250806-171235-automatic-worktree-recovery
    - depends_on:250806-171306-doctor-command
---

# Enhanced Recovery Features

## Overview
Implement advanced recovery capabilities for complex worktree corruption scenarios that go beyond basic `git worktree prune`. This includes partial corruption repair, metadata backup/restore, and detailed recovery reporting.

## Tasks
- [ ] Implement worktree metadata backup system
- [ ] Add partial corruption repair logic
- [ ] Create recovery journal for tracking operations
- [ ] Add recovery statistics and reporting
- [ ] Implement dry-run mode for all recovery operations
- [ ] Add worktree state validation
- [ ] Create recovery configuration options
- [ ] Add comprehensive test coverage
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Get developer approval before closing

## Technical Details
### Metadata Backup System
- Backup `.git/worktrees/` metadata before operations
- Store backups in `.ticketflow/recovery/backups/`
- Implement restore capability with rollback
- Auto-cleanup old backups (keep last 10)

### Partial Corruption Repair
- Reconstruct missing HEAD files
- Repair commondir references
- Fix gitdir pointers
- Recover from incomplete worktree additions

### Recovery Journal
Create `.ticketflow/recovery/journal.json`:
```json
{
  "entries": [
    {
      "timestamp": "2025-08-06T17:00:00Z",
      "operation": "prune_worktree",
      "target": "feature-branch",
      "result": "success",
      "details": {
        "error_detected": "directory_missing",
        "action_taken": "pruned",
        "backup_location": "backups/2025-08-06-170000"
      }
    }
  ]
}
```

### Recovery Configuration
Add to `.ticketflow.yaml`:
```yaml
recovery:
  enabled: true
  auto_recovery: true
  max_retries: 3
  backup_before_recovery: true
  keep_backups: 10
  dry_run_default: false
  verbose_logging: false
```

### Advanced Recovery Operations
1. **State Validation**:
   - Verify worktree consistency
   - Check branch tracking
   - Validate ticket file locations
   - Ensure git index integrity

2. **Smart Recovery**:
   - Analyze corruption type
   - Choose appropriate recovery strategy
   - Minimize data loss
   - Preserve user changes when possible

3. **Recovery Strategies**:
   - `PRUNE_ONLY`: Simple git worktree prune
   - `REPAIR_METADATA`: Fix .git/worktrees entries
   - `RECREATE_WORKTREE`: Remove and recreate
   - `MANUAL_INTERVENTION`: Requires user action

## Acceptance Criteria
- [ ] Metadata backup/restore works reliably
- [ ] Partial corruption can be repaired without full prune
- [ ] Recovery journal tracks all operations
- [ ] Statistics provide useful debugging information
- [ ] Dry-run mode shows what would be done
- [ ] Configuration options control recovery behavior
- [ ] No data loss for recoverable scenarios
- [ ] Clear documentation for advanced features
- [ ] Test coverage > 80%

## Notes
This is phase 4 of the worktree recovery implementation. These are advanced features that handle edge cases and provide enterprise-level recovery capabilities. They build on the foundation of the previous phases but are optional enhancements for users who need more sophisticated recovery options.