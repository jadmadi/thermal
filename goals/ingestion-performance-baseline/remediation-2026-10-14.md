# Performance evidence remediation PERF-2026

## Context

P2. Ready, no prerequisites. Cwd `/mnt/Jad/github/lab/thermal-streak`.
Read `scripts/benchmark.sh`, `internal/loaders/benchmark_test.go`, `docs/PERFORMANCE.md`,
`internal/loaders/sourcesig.go` and this audit's benchmark log. Common rules: `tasks/plan.md`.
Why: a small two-second measurement cannot establish a universal latency/CPU promise.
Historical timing is not fabricated or overwritten; record new candidate-bound measurements.

## P1: Correct the measurement runner

Allowed: `scripts/benchmark.sh`, `internal/loaders/benchmark_test.go`,
create `scripts/benchmark/main.go` and `scripts/benchmark/main_test.go` for stdlib-only fixture
assertions and measurement orchestration as needed. The shell remains the public entrypoint.
This Go helper is repository-specific, not a new general go-tools utility.
Build binary and fixtures under a temporary directory. Validate exactly 1,873,500 baseline tokens
and exactly the five intended source names, with per-source totals calculated from fixture rows.
Assert expected token deltas after a valid Devin metadata.metrics append and a valid OpenCode
append, plus rewrite/truncate/delete cases. Cold means every relevant Thermal snapshot removed,
not just Devin; never remove user caches. Isolate every discovered tool-root override and XDG path.
Timing must stop with nonzero exit on fixture mismatch, failed command, malformed stream or
unexpected timeout. Only a deliberately terminated healthy streaming sample accepts timeout.
Use Go monotonic timing for repeated subprocess measurements. Record 30 cold and 100 warm samples,
stdout sink, source revision/diff hash, Go version, OS, CPU, fixture dimensions, p50 and p95.
No hard universal 10ms assertion across machines. Report whether the documented budget was met
on this machine. Add a relative-regression comparison only against matching hardware/fixture data.

- PERF-201: helper tests deliberately perturb totals and process exit; runner must reject them.
  Exact baseline/delta checks pass. `go test ./scripts/benchmark ./internal/loaders` exits 0 if
  helper is added; no helper means equivalent assertion tests remain in loaders benchmarks.
- PERF-202: cold/warm cache counters demonstrate distinct states; no root binary or user-store
  mutation. `./scripts/benchmark.sh` exits 0 only after correctness checks pass.

## P2: Publish bounded performance evidence

Depends on P1. Allowed: `scripts/benchmark.sh`, `scripts/benchmark/main.go`,
`docs/PERFORMANCE.md`, `README.md`. Common companion metadata edits permitted.
Use both the small fixture and 1000 Claude files with 30 records each. Measure idle and changing
sources at 100ms and 1s for at least 60 seconds after warm-up. Include CPU seconds/wall seconds,
refresh counts, source freshness latency, NDJSON validity, and cancellation latency. Keep the
100ms interface. State whether TTY responsiveness was measured; NDJSON is not TTY evidence.
Document nonzero whole-process allocations separately from zero-allocation helper benchmarks.
Do not change loaders to make the benchmark pass; create a bounded follow-up only from measured
hotspots. Preserve the governance target while qualifying unsupported blanket README promises.

- PERF-203: published report has raw logs tied to the final source hash and exact runnable commands;
  warm latency, freshness and CPU have separate verdicts and sample sizes. All evidence is synthetic.
- PERF-204: common final checks pass; independent review confirms no gate can PASS on bad fixture
  totals or a dead live process. A missed performance target stays a disclosed gap, not a fake PASS.

Stop/return/review contract: `tasks/plan.md`. P2 required before closure.
