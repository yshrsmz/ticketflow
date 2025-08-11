# TicketFlow Benchmark Results

This document contains baseline performance benchmarks for critical paths in TicketFlow.

## Overview

Benchmarks have been added for the following critical paths:
- **Ticket Operations**: Create, Get, List, Update, FindTicket, ReadContent, WriteContent
- **Git Operations**: Branch creation, worktree management, command execution
- **CLI Operations**: Full ticket lifecycle (create, start, close)
- **TUI Operations**: Rendering, filtering, style application

## Running Benchmarks

### Run all benchmarks:
```bash
make bench
```

### Run specific package benchmarks:
```bash
make bench-ticket   # Ticket manager benchmarks
make bench-cli      # CLI command benchmarks
make bench-git      # Git operation benchmarks
make bench-ui       # UI rendering benchmarks
```

### Run verbose benchmarks with longer duration:
```bash
make bench-verbose
```

## Baseline Performance Results

> Note: These results are from initial implementation. Run benchmarks on your system for accurate measurements.

### Ticket Manager Operations

| Operation | ns/op | B/op | allocs/op | Notes |
|-----------|-------|------|-----------|-------|
| Create | ~2.5ms | ~550KB | ~6500 | Creates ticket file with frontmatter |
| Get | TBD | TBD | TBD | Reads single ticket |
| **List Sequential (10 tickets)** | 607µs | 127KB | 951 | Original implementation |
| **List Concurrent (10 tickets)** | 401µs | 121KB | 984 | 34% faster |
| **List Sequential (50 tickets)** | 2.8ms | 578KB | 4640 | Original implementation |
| **List Concurrent (50 tickets)** | 1.4ms | 593KB | 4881 | 51% faster |
| **List Sequential (100 tickets)** | 5.6ms | 1.16MB | 9246 | Original implementation |
| **List Concurrent (100 tickets)** | 2.7ms | 1.19MB | 9764 | 52% faster |
| **List Sequential (200 tickets)** | 11ms | 2.33MB | 18450 | Original implementation |
| **List Concurrent (200 tickets)** | 5.2ms | 2.37MB | 19550 | 53% faster |
| **List Sequential (500 tickets)** | 28ms | 5.94MB | 46067 | Original implementation |
| **List Concurrent (500 tickets)** | 14ms | 5.93MB | 48922 | 50% faster |
| Update | TBD | TBD | TBD | Updates ticket content |
| FindTicket | TBD | TBD | TBD | Searches for ticket by ID |

### Git Operations

| Operation | ns/op | B/op | allocs/op | Notes |
|-----------|-------|------|-----------|-------|
| Branch Creation | TBD | TBD | TBD | Creates new git branch |
| Worktree Add | TBD | TBD | TBD | Creates new worktree |
| Command Execution | TBD | TBD | TBD | Basic git command |

### CLI Operations

| Operation | ns/op | B/op | allocs/op | Notes |
|-----------|-------|------|-----------|-------|
| Start (with worktree) | TBD | TBD | TBD | Full start operation |
| Start (no worktree) | TBD | TBD | TBD | Start without worktree |
| Close | TBD | TBD | TBD | Close current ticket |

### UI Rendering

| Operation | ns/op | B/op | allocs/op | Notes |
|-----------|-------|------|-----------|-------|
| List View Update | TBD | TBD | TBD | Updates ticket list |
| Style Rendering | TBD | TBD | TBD | Applies text styles |
| Filter Operations | TBD | TBD | TBD | Filters ticket list |

## Performance Considerations

1. **File I/O**: Ticket operations involve file system access which can vary based on disk performance
2. **Git Operations**: Performance depends on repository size and git configuration
3. **Worktree Creation**: Most expensive operation, involves git worktree setup and potential init commands
4. **Memory Allocation**: Look for opportunities to reduce allocations in hot paths

## Optimization Opportunities

Based on initial benchmarks, potential areas for optimization:
1. ~~Reduce allocations in ticket list operations~~ ✅ Implemented concurrent loading with 50%+ improvement for 50+ tickets
2. Cache frequently accessed ticket metadata
3. Optimize file path operations
4. Consider connection pooling for git operations

### Implemented Optimizations

#### Concurrent Directory Operations (Task 1.2)
- **Implementation**: Added concurrent file loading using `errgroup` and `semaphore`
- **Automatic switching**: Uses concurrent loading for 10+ tickets, sequential for smaller sets
- **Worker management**: Uses `runtime.NumCPU()` with max 8 workers to avoid excessive file handles
- **Results**: Achieved 50-53% performance improvement for 100+ tickets
- **Memory impact**: Minimal increase in allocations (~5%) with similar memory usage
- **Debug logging**: Set log level to debug to see strategy selection and performance metrics
  ```bash
  ticketflow --log-level debug list
  ```

## How to Interpret Results

- **ns/op**: Nanoseconds per operation (lower is better)
- **B/op**: Bytes allocated per operation (lower is better)
- **allocs/op**: Number of allocations per operation (lower is better)

To generate updated results:
```bash
make bench > benchmark-results.txt
```

## Best Practices for Benchmarking

1. **Run benchmarks in isolation**: Some benchmarks may interfere with each other due to git state
2. **Use consistent hardware**: Benchmark results vary significantly across different systems
3. **Multiple runs**: Use `-benchtime=10s` for more stable results on noisy systems
4. **Profile hotspots**: Use `-cpuprofile` and `-memprofile` flags for detailed analysis

## Future Enhancements

Based on the review, consider these enhancements:

1. **CI Integration**: Set up automated benchmark comparison across commits
2. **Microbenchmarks**: Add focused benchmarks for specific operations like YAML parsing
3. **Error Path Benchmarks**: Test performance of error handling scenarios
4. **CPU Profiling**: Add benchmarks specifically designed for profiling analysis