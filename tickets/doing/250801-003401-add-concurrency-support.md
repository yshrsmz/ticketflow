---
priority: 3
description: Add safe concurrency for independent git operations and parallel file processing
created_at: "2025-08-01T00:34:01+09:00"
started_at: "2025-08-08T22:55:58+09:00"
closed_at: null
related:
    - parent:250801-003206-add-context-support
---

# Add Concurrency Support

Implement safe concurrency for independent operations using goroutines and channels to improve performance, especially for operations that can be parallelized.

## Context

Many operations in TicketFlow could benefit from concurrency:
- Reading multiple ticket files
- Running independent git operations
- Processing ticket lists for display
- Performing batch operations

Currently all operations are sequential, which can be slow when:
- Working with many tickets
- Running multiple git commands
- Processing large amounts of data

## Tasks

### Concurrent Git Operations
- [ ] Identify independent git operations that can run in parallel
- [ ] Implement goroutine pool for git commands
- [ ] Add proper error handling for concurrent operations
- [ ] Ensure thread-safe access to shared resources

### Parallel File Processing
- [ ] Implement parallel ticket file reading
- [ ] Add worker pool for file operations
- [ ] Batch process ticket updates
- [ ] Implement parallel directory scanning

### TUI Concurrency
- [ ] Add background loading for ticket lists
- [ ] Implement non-blocking UI updates
- [ ] Add progress indicators for long operations
- [ ] Ensure thread-safe state management

### Infrastructure
- [ ] Create worker pool implementation
- [ ] Add semaphore for resource limiting
- [ ] Implement error aggregation for parallel operations
- [ ] Add metrics for concurrent operations

### Quality Assurance
- [ ] Add tests with race detector enabled
- [ ] Test error handling in concurrent scenarios
- [ ] Benchmark concurrent vs sequential operations
- [ ] Run `make test -race`
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation
- [ ] Get developer approval before closing

## Implementation Guidelines

### Worker Pool Pattern
```go
type WorkerPool struct {
    workers int
    tasks   chan func() error
    errors  chan error
    wg      sync.WaitGroup
}

func (wp *WorkerPool) Start(ctx context.Context) {
    for i := 0; i < wp.workers; i++ {
        wp.wg.Add(1)
        go wp.worker(ctx)
    }
}

func (wp *WorkerPool) worker(ctx context.Context) {
    defer wp.wg.Done()
    for {
        select {
        case <-ctx.Done():
            return
        case task := <-wp.tasks:
            if err := task(); err != nil {
                wp.errors <- err
            }
        }
    }
}
```

### Concurrent File Reading
```go
func (m *Manager) ListTicketsConcurrent(ctx context.Context) ([]*Ticket, error) {
    files, err := m.findTicketFiles()
    if err != nil {
        return nil, err
    }
    
    tickets := make([]*Ticket, len(files))
    errors := make(chan error, len(files))
    
    var wg sync.WaitGroup
    for i, file := range files {
        wg.Add(1)
        go func(idx int, path string) {
            defer wg.Done()
            ticket, err := m.loadTicket(ctx, path)
            if err != nil {
                errors <- err
                return
            }
            tickets[idx] = ticket
        }(i, file)
    }
    
    wg.Wait()
    close(errors)
    
    // Check for errors
    for err := range errors {
        if err != nil {
            return nil, err
        }
    }
    
    return tickets, nil
}
```

### Safe State Management
```go
type SafeState struct {
    mu    sync.RWMutex
    data  map[string]*Ticket
}

func (s *SafeState) Get(id string) (*Ticket, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    ticket, ok := s.data[id]
    return ticket, ok
}

func (s *SafeState) Set(id string, ticket *Ticket) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.data[id] = ticket
}
```

## Notes

This ticket depends on context support being implemented first, as proper cancellation is crucial for concurrent operations.

Be careful with concurrency - it adds complexity. Only add it where there's a clear performance benefit. Always use the race detector during development and testing.

Consider the number of goroutines carefully - too many can overwhelm the system. Use worker pools to limit concurrency.