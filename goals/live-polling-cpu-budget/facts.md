# Facts: Live polling CPU budget

## User observation
The user reports that `thermal live --interval 100ms` burns CPU. Their exact dataset, CPU percentage and hardware were not captured; do not represent the user report as a controlled benchmark.

## Confirmed code path
- `cmd/thermal/main.go:1691-1705`: each poll calls `loadUsage` and traverses daily rows for estimated costs.
- `cmd/thermal/main.go:875-917`: all-tools collection loops over every available tool and calls `LoadToolData` on each poll.
- `internal/tui/live.go:92-101`: a tick calls the collector synchronously in `Update`, then schedules the next tick. The TUI does not spawn overlapping scans today; expensive scans consume CPU and stall message handling.
- `cmd/thermal/main.go:1750-1762`: NDJSON uses a ticker and the same full collector. Slow polls can leave the loop continuously busy, even without concurrent scans.
- `internal/loaders/codex.go:156-185,349+`: bounded workers still rescan rollout files. Several other directory/SQLite paths also lack unchanged-source reuse.

The issue is excessive repeated work and coupling collection to UI updates, not proven unbounded goroutine creation. A 100ms timer does not guarantee 10 completed scans per second when scanning itself is slow.

## Controlled reproduction
On Go 1.27.1 / Linux / Intel i7-9700 (8 logical CPUs), 1,000 unchanged Claude files containing 30,000 records were sampled for six seconds after two seconds of warmup, with pricing/network disabled. PTY 120x40 measured 12.8% of one core at 1s versus 96.2% at 100ms; NDJSON measured 13.5% versus 130.5%. Process user+system time, not whole-machine utilization. These short diagnostic samples do not replace the 30-second acceptance measurement.

[Durable CPU probe](../../docs/evaluations/2026-09-24-live-cpu.py.txt) and [full evaluation](../../docs/evaluations/2026-09-24-thermal.md) record the fixture and commands.

## Verification contract
Use the durable synthetic CPU probe referenced by the evaluation for before/after streaming CPU, and add a real PTY run for the exact interactive command. Report process CPU seconds/wall seconds (100% = one core), startup versus steady state, source size, toolchain, machine and memory/I/O. Use a controlled fixture; no personal logs in artifacts. The original short probe is diagnostic only; closure needs the goal's 30-second measurement.

Focused: `go test -race ./internal/tui ./cmd/thermal -run Live`. Exercise injected slow loaders without timing-flaky sleeps. Where view output changes, follow ui-craft and preserve the required 80x24/120x40 goldens and key-flow tests.

## Ownership and coordination
Own scheduling and collection reuse, not token/turn formula changes. `live-metric-parity` follows this goal so it can rely on the new scheduling contract. `hermetic-verification-gates` must enforce bounded poll counts and cancellation. `ingestion-performance-baseline` adds sustained 100ms/1s measurements without blocking this urgent fix. Required new files in the ownership list are deliverables, not pre-existing artifacts.
