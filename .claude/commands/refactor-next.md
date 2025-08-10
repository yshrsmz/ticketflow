# /refactor-next

Analyzes the refactoring ticket progress and intelligently selects the next ticket to work on based on dependencies and current progress.

## What it does

1. **Analyzes Progress**: Scans all refactoring tickets to determine their current status (todo/doing/done)
2. **Checks Dependencies**: Ensures dependencies are met before suggesting a ticket
3. **Selects Best Next Ticket**: Prioritizes by phase and complexity
4. **Provides Full Context**: Shows ticket details, implementation notes, and quick-start commands

## Usage

```
/refactor-next
```

## Implementation

```bash
bash ./scripts/refactor-next.sh
```

## Features

- **Smart Selection**: Prioritizes tickets by phase (1-5) and complexity (Low/Medium/High)
- **Dependency Checking**: Only suggests tickets whose dependencies are complete
- **Progress Tracking**: Shows completed, in-progress, and available tickets
- **Full Context**: Displays ticket content, implementation notes, and documentation links
- **Quick Commands**: Provides ready-to-use commands to start working immediately

## Example Output

```
TicketFlow Refactoring Progress Analyzer
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📈 Progress Summary:
  ✓ Completed: 0 tickets
  ⚡ In Progress: 0 tickets
  ○ Todo: 20 tickets
  ◆ Available (deps met): 4 tickets

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🎯 Selected Ticket: refactor-benchmark-infrastructure
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📊 Metadata:
  Phase: 1
  Complexity: Low
  Duration: 0.5 days
  Dependencies: None

📄 Ticket Content:
────────────────────────────────────────────────────────────
[Ticket content displayed here]
────────────────────────────────────────────────────────────

📚 Related Documentation:
  • Full refactoring discussion: docs/20250810-refactor-discussion.md
  • Executive summary: docs/20250810-refactor-summary.md
  • Ticket overview: docs/20250810-refactor-tickets.md

💡 Implementation Notes:
  • Use testing.B with b.ReportAllocs() to track allocations
  • Implement b.StopTimer()/b.StartTimer() for setup exclusion
  • Create comparison benchmarks: BenchmarkListSequential vs BenchmarkListConcurrent
  • Use benchstat for statistical analysis
  • Set up pprof integration for CPU and memory profiling
  • Key files: Create benchmark_test.go files in relevant packages
  • Reference: golang.org/x/perf/cmd/benchstat

🚀 Quick Start Commands:
  # Start working on this ticket:
  ticketflow start 250810-002848-refactor-benchmark-infrastructure
  cd ../ticketflow.worktrees/250810-002848-refactor-benchmark-infrastructure

  # View the ticket:
  ticketflow show 250810-002848-refactor-benchmark-infrastructure

  # After completing work:
  ticketflow close  # Run from within the worktree
  git push

📋 Other Available Tickets:
  • refactor-command-interface (Phase 2, Medium complexity, 1 day)
  • refactor-parallel-test-execution (Phase 1, Low complexity, 0.5 days)
  • refactor-configuration-caching (Phase 3, Low complexity, 1 day)
```