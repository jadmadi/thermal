# Goal: Bounded Worker Concurrency in Multi-File Loaders

## Goal Description
Multi-file session loaders (`agy`, `claude`, `codex`, `command-code`, `droid`, `dsh`, `grok`) scan local directories containing hundreds to tens of thousands of JSON/JSONL session transcripts. Previously, semaphore tokens were acquired *inside* spawned goroutines (`go func() { sem <- struct{}{} ... }`), causing the main scanning loop to spawn unbounded goroutines on the Go runtime heap before waiting on the semaphore.

This goal moves semaphore acquisition immediately *before* goroutine invocation (`sem <- struct{}{}; wg.Add(1); go func() { ... }()`), ensuring that at most $N$ (8 or 16) goroutines exist concurrently in memory at any point during directory scans.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship

## Files to Touch
- `internal/loaders/agy.go`
- `internal/loaders/claude.go`
- `internal/loaders/codex.go`
- `internal/loaders/commandcode.go`
- `internal/loaders/droid.go`
- `internal/loaders/dsh.go`
- `internal/loaders/grok.go`
- `AGENTS.md`

## References
- **Shared Understanding & Fact Sheet**: [`goals/loader-bounded-concurrency/facts.md`](facts.md)
- **Execution Plan**: [`goals/loader-bounded-concurrency/plan.md`](plan.md)

## Done Condition
1. All 7 multi-file session loaders acquire their worker limit token before calling `go func`.
2. Memory allocation during large directory scans is bounded by the semaphore channel capacity.
3. Unit tests pass cleanly with race detection: `go test -race ./internal/loaders/...`.
4. Release gate passes: `./scripts/simulated_user_gate.sh`.
