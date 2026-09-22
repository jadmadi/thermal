# Facts: Bounded Worker Concurrency in Multi-File Loaders

## Architectural Facts & Grounding

1. **Unbounded Goroutine Spawning Pitfall**:
   In Go, spawning a goroutine (`go func()`) immediately allocates a stack frame (typically 2KB–4KB). When iterating over thousands of files in a directory, spawning `len(files)` goroutines where each goroutine blocks on `sem <- struct{}{}` creates all goroutines simultaneously on the Go scheduler queue, causing memory spikes and GC overhead.

2. **Canonical Semaphore Scheduling Pattern**:
   According to `golang-concurrency` and `AGENTS.md` Rule A.5, bounded worker pools using semaphore channels must acquire the slot token *prior* to `go func()`:
   ```go
   sem <- struct{}{}
   wg.Add(1)
   go func(p string) {
       defer wg.Done()
       defer func() { <-sem }()
       // do work
   }(path)
   ```
   This guarantees that the producer loop pauses whenever `cap(sem)` workers are active, limiting the number of alive goroutines strictly to the pool capacity.

3. **Affected Loaders**:
   - `internal/loaders/agy.go` (16 workers)
   - `internal/loaders/claude.go` (8 workers)
   - `internal/loaders/codex.go` (8 workers via `workerLimit`)
   - `internal/loaders/commandcode.go` (8 workers)
   - `internal/loaders/droid.go` (8 workers)
   - `internal/loaders/dsh.go` (16 workers)
   - `internal/loaders/grok.go` (8 workers)

4. **Safety & Deadlock Freedom**:
   - Each worker releases its slot via `defer func() { <-sem }()` before calling `wg.Done()`.
   - Results channels are buffered or read after completion, preventing deadlock between workers and the collector.
