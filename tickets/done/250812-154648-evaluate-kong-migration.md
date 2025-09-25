---
priority: 4
description: "Evaluate migration to Kong CLI library after current refactoring"
created_at: "2025-08-12T15:46:48+09:00"
started_at: null
closed_at: "2025-09-24T09:50:00+09:00"
related:
    - parent:250810-003001-refactor-command-interface
    - "blocks:250812-152824-migrate-help-command"
    - "blocks:250812-152902-migrate-init-command"
    - "blocks:250812-152927-migrate-remaining-commands"
---

# Evaluate Kong CLI Library Migration [ABANDONED]

**Status: Abandoned in favor of pflag migration**
**Reason: After thorough analysis, pflag provides a simpler, drop-in solution for the immediate need of supporting flags after positional arguments**

After completing the current command interface migration, evaluate whether migrating to Kong (github.com/alecthomas/kong) would provide benefits over the custom implementation.

## Background

Based on 2025 Go CLI library trends:
- Cobra is increasingly seen as "bloated" with messy code generation
- Kong is gaining popularity as a cleaner, struct-based alternative
- Developers report migrating from other libraries to Kong in ~40 minutes
- Kong provides a good balance between simplicity and features

## Evaluation Criteria

### Features to Compare
- [ ] Command and subcommand support
- [ ] Flag parsing and validation
- [ ] Help text generation
- [ ] Shell completions (bash, zsh, fish)
- [ ] Error handling and user feedback
- [ ] Testing support
- [ ] Integration with BubbleTea TUI

### Code Quality Metrics
- [ ] Lines of code reduction
- [ ] Testability improvements
- [ ] Maintenance burden
- [ ] Learning curve for team
- [ ] Documentation quality

## Tasks

### Phase 1: Research
- [ ] Review Kong documentation and examples
- [ ] Analyze how Kong handles subcommands (for worktree operations)
- [ ] Check Kong's compatibility with BubbleTea
- [ ] Review migration stories from other projects
- [ ] Evaluate Kong's shell completion capabilities

### Phase 2: Proof of Concept
- [ ] Create a branch for Kong POC
- [ ] Migrate 2-3 commands to Kong as examples
- [ ] Compare code complexity with current implementation
- [ ] Test shell completions
- [ ] Benchmark performance (if relevant)

### Phase 3: Decision
- [ ] Document pros and cons
- [ ] Calculate migration effort for all commands
- [ ] Make recommendation (migrate/don't migrate)
- [ ] If migrating, create detailed migration plan
- [ ] Get team consensus

## Success Criteria

The migration should only proceed if Kong provides:
- Significant code reduction (>30%)
- Automatic shell completions
- Better maintainability
- No loss of current functionality
- Clean integration with existing BubbleTea TUI

## Example Kong Structure

```go
type CLI struct {
    Version VersionCmd `cmd:"" aliases:"v" help:"Show version"`
    List    ListCmd    `cmd:"" help:"List tickets"`
    New     NewCmd     `cmd:"" help:"Create new ticket"`
    
    Worktree struct {
        List  WorktreeListCmd  `cmd:"" help:"List worktrees"`
        Clean WorktreeCleanCmd `cmd:"" help:"Clean worktrees"`
    } `cmd:"" help:"Manage worktrees"`
}

type VersionCmd struct{}

func (v *VersionCmd) Run() error {
    fmt.Printf("ticketflow version %s\n", Version)
    return nil
}
```

## Notes

- This evaluation should happen AFTER the current custom interface migration is complete
- Kong uses struct tags for configuration, which is very different from our current approach
- Consider creating a small separate tool with Kong first to get familiar
- Shell completions alone might justify the migration

## References

- Kong repository: https://github.com/alecthomas/kong
- Kong examples: https://github.com/alecthomas/kong/tree/master/_examples
- Migration guide from Cobra to Kong: Search for community examples
- Current implementation: `internal/command/` directory