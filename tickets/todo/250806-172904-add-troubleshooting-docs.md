---
priority: 3
description: "Add troubleshooting documentation for common worktree issues"
created_at: "2025-08-06T17:29:04+09:00"
started_at: null
closed_at: null
related:
    - parent:250803-121506-worktree-recovery
---

# Add Troubleshooting Documentation

## Overview
Create comprehensive troubleshooting documentation that covers common worktree issues and their solutions. This will help users self-diagnose and fix problems without needing complex recovery mechanisms in ticketflow itself.

## Tasks
- [ ] Create `docs/troubleshooting.md` with worktree troubleshooting section
- [ ] Document common worktree issues and solutions
- [ ] Add examples of error messages and their fixes
- [ ] Include preventive measures
- [ ] Link from main README.md
- [ ] Get developer approval before closing

## Documentation Content

### Structure for `docs/troubleshooting.md`

```markdown
# Troubleshooting Guide

## Worktree Issues

### Common Problems and Solutions

#### 1. Corrupted Worktree References
**Symptoms:**
- Error: `fatal: '<path>' is not a working tree`
- Worktree directory was deleted but git still tracks it

**Solution:**
```bash
# Clean up corrupted worktree references
git worktree prune

# Retry your ticketflow command
ticketflow start <ticket-id>
```

#### 2. Branch Already Checked Out
**Symptoms:**
- Error: `fatal: '<branch>' is already checked out at '<path>'`

**Solution:**
```bash
# Find where the branch is checked out
git worktree list

# Either switch to that worktree or remove it
cd <worktree-path>  # Option 1: Use existing worktree
# OR
git worktree remove <worktree-path>  # Option 2: Remove old worktree
```

#### 3. Orphaned Worktree Directories
**Symptoms:**
- Worktree directories exist but aren't tracked by git
- Cannot create new worktree due to existing directory

**Solution:**
```bash
# List all worktrees
git worktree list

# Remove untracked directories manually
rm -rf ../ticketflow.worktrees/<orphaned-directory>

# Clean up git's worktree references
git worktree prune
```

#### 4. Cannot Create Worktree Directory
**Symptoms:**
- Error: `fatal: could not create work tree dir`

**Solution:**
```bash
# Check permissions
ls -la ../ticketflow.worktrees/

# Check disk space
df -h

# Try pruning first in case of corruption
git worktree prune
```

### Preventive Measures

1. **Always use ticketflow cleanup:**
   ```bash
   ticketflow cleanup <ticket-id>
   ```
   Don't manually delete worktree directories.

2. **Regular maintenance:**
   ```bash
   # Periodically clean up stale references
   git worktree prune
   ```

3. **Check worktree status:**
   ```bash
   # List all worktrees and their status
   git worktree list --porcelain
   ```

### Advanced Debugging

If basic solutions don't work:

1. **Check git worktree metadata:**
   ```bash
   ls -la .git/worktrees/
   ```

2. **Verify worktree configuration:**
   ```bash
   cat .git/worktrees/<worktree-name>/gitdir
   ```

3. **Force remove problematic worktree:**
   ```bash
   git worktree remove --force <path>
   ```

4. **Last resort - manual cleanup:**
   ```bash
   # Remove worktree directory
   rm -rf <worktree-path>
   
   # Remove git metadata
   rm -rf .git/worktrees/<worktree-name>
   
   # Prune references
   git worktree prune
   ```

### Getting Help

If you encounter issues not covered here:
1. Check the [GitHub Issues](https://github.com/your-org/ticketflow/issues)
2. Run commands with verbose output: `git -c core.trace=1 worktree ...`
3. Include full error messages when reporting issues
```

## Files to Create/Modify
- Create `docs/troubleshooting.md`
- Update `README.md` to link to troubleshooting guide

## Acceptance Criteria
- [ ] Comprehensive documentation covers all common worktree issues
- [ ] Each issue has clear symptoms and solutions
- [ ] Examples use actual command outputs
- [ ] Preventive measures are clearly explained
- [ ] Documentation is linked from main README
- [ ] Writing is clear and accessible to developers

## Notes
This documentation approach aligns with the decision to keep ticketflow focused on ticket management while empowering users to handle git worktree issues themselves. The guide should be practical and solution-oriented.