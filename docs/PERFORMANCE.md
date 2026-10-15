# Thermal Performance & Ingestion Baseline

This document defines the performance baselines, benchmarking methodology, and hot-path telemetry for **Thermal** (`github.com/jadmadi/thermal`).

Thermal operates under a **Sub-10ms Cached Invocation** architectural target codified in [`GOVERNANCE.md`](../GOVERNANCE.md). This report is bounded evidence for that target on one machine and one synthetic fixture family — it is not a universal latency guarantee. Every number below is reproducible with [`scripts/benchmark.sh`](../scripts/benchmark.sh), and correctness failures stop the run before any performance number is measured.

---

## 1. Evidence binding

| Field | Value |
| :--- | :--- |
| Source revision | `78b5d70` (`78b5d70627f89245e751a828df8ebb31518d2879`) |
| Working tree | dirty (pre-existing documentation edits only; diff sha256 `d46be152f87d7ab8d21900b1dbc902d287bec77e27f98064fa4146a3d537c984`) |
| Toolchain | Go 1.27.2 (`linux/amd64`) |
| CPU | Intel(R) Core(TM) i7-9700 CPU @ 3.00GHz (8 cores) |
| OS | Linux (`linux/amd64`) |
| Command | `./scripts/benchmark.sh --out /tmp/thermal-bench-dircache-final` |
| Raw artifacts | `bench-report.json` and `bench-report.txt` written by the run (JSON mirrors this document) |

This run measures the directory scan cache introduced in this revision (`DirScanCache`, `internal/loaders/dircache.go`). The previous published run at `e2cdc72` is the comparison point in §4 and §7; comparisons are only meaningful against matching hardware, fixture dimensions, and method.

The benchmark binary and all fixtures are built under the operating system's temporary directory. No binary is written to the repository root, and no user data store is read or written: the run uses an isolated `HOME` plus every tool-root override (`GROK_HOME`, `DSH_HOME`, `HERMES_HOME`, `CODEX_HOME`) and XDG path the loaders consult.

---

## 2. Executive summary — per-machine verdicts

| Measurement | Result (this machine) | Budget | Verdict |
| :--- | :--- | :--- | :--- |
| **Warm cached CLI latency, small fixture (n=100)** | **p50 5.9ms · p95 6.7ms** | `<10ms` | **PASS** |
| **Warm cached CLI latency, 1,000-file fixture (n=20)** | **p50 5.5ms · p95 5.8ms** | `<10ms` | **PASS** (was MISSED at 42.4ms before the directory scan cache) |
| Cold ingestion (n=30, all snapshots reset) | p50 11.6ms · p95 12.1ms | none published | OBSERVED |
| Live monitor CPU at 100ms (60s window) | 1.08% CPU (0.703s CPU / 65.0s process wall) | `<5%` (previously published on a 30-file fixture over 2s) | OBSERVED — not directly comparable |
| Live monitor CPU at 1s (60s window) | 0.46% CPU (0.296s CPU / 65.0s process wall) | `<2%` (same caveat) | OBSERVED — not directly comparable |
| Source freshness at 100ms | append reflected after 81ms | none published | OBSERVED |
| Source freshness at 1s | append reflected after 886ms | none published | OBSERVED |
| Cancellation latency (SIGINT) | 1.5ms / 0.9ms | none published | OBSERVED |
| Model-name helper allocations | 21.1 / 29.4 / 106.7 ns/op, **0 B/op, 0 allocs/op** | `0 allocs` | PASS (helpers only) |

The governance target now holds on both fixture scales. The directory scan cache reuses an untouched tree from a per-file signature instead of re-reading file bodies, cutting the 1,000-file warm p50 from 42.4ms to 5.5ms without changing results (cache-vs-full-scan parity is a test, and freshness below shows live detection still works).

---

## 3. Synthetic fixtures and correctness assertions

All fixtures are generated deterministically by `scripts/benchmark` under a temp directory; only synthetic data is used.

### Small fixture — exact `1,873,500` tokens across five sources

| Tool source | Store | Scale | Expected tokens |
| :--- | :--- | :--- | ---: |
| Devin | SQLite `sessions.db` | 25 sessions × 20 message nodes | 425,000 |
| Codex | SQLite `state_5.sqlite` + rollout JSONL | 20 threads × 5 rollouts | 70,000 |
| OpenCode | SQLite `opencode.db` (`session_v2`) | 50 sessions | 87,500 |
| Claude | directory JSONL | 30 session files | 291,000 |
| CodeWhale | directory JSON | 20 session files | 1,000,000 |
| **Total** | | **73 workload files** | **1,873,500** |

The run fails with a nonzero exit unless **every** source total matches exactly, exactly the five intended sources are present, and the grand total equals 1,873,500. A non-zero-token unexpected source, a missing source, or a one-token perturbation is rejected (unit-tested in `scripts/benchmark/main_test.go`).

### Delta behavior — tracked expectation, exact match required

| Case | Mutation | Expected Δ | Observed Δ | Status |
| :--- | :--- | ---: | ---: | :--- |
| devin append | valid `metadata.metrics` message node | +80 | +80 | PASS |
| opencode append | valid `session_v2` row | +1,750 | +1,750 | PASS |
| devin rewrite | update one message's token metrics | −750 | −750 | PASS |
| devin truncate | delete one message node | −850 | −850 | PASS |
| codewhale delete | remove one session file | −50,000 | −50,000 | PASS |

### Large fixture — `1,000` Claude files × 30 records = `291,000,000` tokens

Used for the sustained live-monitor measurements and for warm-latency evidence at directory-scan scale. Its total is asserted exactly in the same way.

---

## 4. End-to-End CLI Latency

Method: each sample is a full `thermal --json --offline` subprocess with stdout sunk, timed with Go's monotonic clock. **Cold** removes every Thermal snapshot artifact in the fixture cache (`devin/`, `codex/`, `zed/`, and every directory snapshot under `claude/`, `codewhale/`, `droid/`, `command-code/`, `agy/`, `grok/`, `dsh/`) while preserving the cached pricing catalog; the fixture home is the only cache location touched. **Warm** runs leave snapshots in place.

```
Cold  n=30  min 10.9ms  mean 11.5ms  p50 11.6ms  p95 12.1ms  max 12.5ms
Warm  n=100 min  5.6ms  mean  6.0ms  p50  5.9ms  p95  6.7ms  max  7.6ms
Large n=20  min  5.1ms  mean  5.5ms  p50  5.5ms  p95  5.8ms  max  5.8ms
```

Raw samples (milliseconds):

```text
cold (30): 11.414 11.093 11.454 11.217 11.102 11.602 11.638 10.860 11.272 11.692
           11.091 11.456 11.136 11.606 11.507 11.203 12.148 12.483 11.962 11.782
           11.815 11.635 11.627 11.537 11.741 11.559 11.328 11.726 11.648 11.292

warm (100): 5.902 5.706 5.803 5.656 5.650 5.713 5.600 5.858 5.854 5.815
            5.642 5.865 6.175 6.225 5.742 5.857 5.730 6.188 5.842 5.581
            6.062 6.024 6.011 6.628 5.588 6.680 6.652 6.122 6.391 5.684
            6.015 5.696 5.690 5.862 6.195 5.996 5.719 5.759 5.631 6.399
            6.622 6.333 5.783 6.446 6.897 6.247 7.575 6.226 6.142 6.687
            6.353 5.884 6.248 6.043 5.616 5.877 5.793 6.236 6.137 6.138
            6.011 5.728 6.103 5.763 5.766 5.731 6.012 5.707 5.940 5.865
            5.616 6.190 5.922 5.907 6.153 6.620 5.943 6.667 6.205 6.489
            5.943 6.001 5.999 6.140 6.122 5.906 6.040 5.614 5.781 5.851
            5.655 5.637 5.955 6.180 6.111 5.889 5.936 5.934 5.791 5.968

large n=20: 5.684 5.570 5.346 5.676 5.400 5.192 5.299 5.253 5.483 5.617
            5.589 5.609 5.362 5.759 5.407 5.549 5.308 5.556 5.479 5.133
```

**Interpretation.** The small fixture's warm p50 of 5.9ms and the 1,000-file fixture's 5.5ms both meet the 10ms governance target. For comparison, the `e2cdc72` run measured warm p50 7.8ms (small) and 42.4ms (large): the 1,000-file gap was dominated by re-walking and re-parsing directory sources on every invocation, and the directory scan cache removed it while keeping cold latency unchanged (the cache is reset before each cold sample, so cold still pays the full scan).

---

## 5. Sustained Live Monitor (NDJSON)

Method: `thermal live --json --stream --interval <rate> --offline` against the 1,000-file fixture for a 5s warm-up plus a 60s measured window. One source file is appended a third of the way into the window (with a 137ms offset so the sample does not land on a poll-tick boundary); freshness is the time from that append until the change event line arrives, which also proves the directory cache does not mask source changes. CPU is the child's user+system rusage over the process lifetime (warm-up included, disclosed). Termination is a deliberate SIGINT; a stream that exits on its own before cancellation, emits malformed NDJSON, or never reflects the appended source fails the run with a nonzero exit.

| Interval | Window | CPU | Change events | Invalid lines | Freshness | Cancellation |
| :--- | ---: | ---: | ---: | ---: | ---: | ---: |
| 100ms | 60s | 0.703s = **1.08%** of one core | 1 | 0 | **81ms** | 1.5ms |
| 1s | 60s | 0.296s = **0.46%** of one core | 1 | 0 | **886ms** | 0.9ms |

The stream emits the initial snapshot line and then one line per detected source change; the measured runs therefore recorded 2 NDJSON lines each (initial snapshot + the appended source's change event).

**Not measured:** TTY responsiveness. These runs exercise NDJSON output on a pipe; NDJSON is not TTY evidence, and no interactive terminal frame rate or input-latency claim is made here.

---

## 6. Allocations: whole process vs helpers

The zero-allocation invariant applies to specific hot-path helpers, **not** to the whole CLI process. `go test ./internal/loaders -run '^$' -bench Benchmark -benchmem -benchtime 500ms` on the same machine (these benchmarks call the loaders directly; the `DirScanCache` wrapper in `LoadToolData` is not part of them):

| Benchmark | Iterations | ns/op | B/op | allocs/op | Note |
| :--- | ---: | ---: | ---: | ---: | :--- |
| `BenchmarkModelName_Canonical` | 28,488,267 | 21.14 | **0** | **0** | zero-allocation helper |
| `BenchmarkModelName_Alias` | 19,984,441 | 29.44 | **0** | **0** | zero-allocation helper |
| `BenchmarkModelName_Router` | 5,549,145 | 106.7 | **0** | **0** | zero-allocation helper |
| `BenchmarkDevin_Cold` | 82 | 7,035,564 | 137,947 | 7,892 | whole-process loader work |
| `BenchmarkDevin_Warm` | 796 | 743,036 | 25,745 | 769 | cache hit still allocates |
| `BenchmarkDevin_AppendDelta` | 100 | 5,811,322 | 96,218 | 4,306 | incremental append |
| `BenchmarkCodex_Cold` | 580 | 1,067,603 | 1,489,010 | 2,024 | SQLite + rollout JSONL |
| `BenchmarkCodex_Warm` | 968 | 615,275 | 58,904 | 1,009 | rollout cache hit |
| `BenchmarkOpenCode_Scan` | 404 | 1,477,957 | 24,052 | 368 | `session_v2` scan |
| `BenchmarkClaude_Scan` | 872 | 713,122 | 3,371,861 | 800 | directory scan |

Whole-process invocations allocate (JSON decoding, token maps, rendering); the loaders' warm paths allocate tens of KB per scan, not zero. Any public claim of "zero allocations" must be read as scoped to the model-name helpers above.

---

## 7. Hotspot analysis and bounded follow-ups

1. **Directory-backed sources at scale — resolved.** The `e2cdc72` run measured warm p50 42.4ms at 1,000 files; `DirScanCache` (`internal/loaders/dircache.go`, wired into `LoadToolData`) now reuses an untouched tree from a per-file signature (relative path, size, nanosecond mtime), measuring 5.5ms at the same scale. Boundary: a rewrite that preserves both byte size and the recorded modification time is not observable by a stat-based signature — the same contract the SQLite snapshot caches use.
2. **Live collector re-scans.** With the cache in place, sustained CPU at 1,000 files is 1.08% (100ms) and 0.46% (1s), both improved from 1.32% / 0.67%; any further change must preserve the measured freshness (81ms / 886ms) and cancellation behavior.
3. **Warm SQLite path allocations.** Cache-hit scans still allocate tens of KB (`BenchmarkDevin_Warm` 25,745 B/op); acceptable today and only worth revisiting against a measured regression.
4. **OpenCode / MiMoCode incremental state tracking — open.** These loaders still scan their full `session_v2`/session tables per invocation (measured 1.48ms per scan on the fixture); no cross-invocation delta exists yet.

---

## 8. Reproducibility

```bash
# 1. Full end-to-end baseline (fixtures, correctness gates, latency, live CPU)
./scripts/benchmark.sh --out /tmp/thermal-bench-final

# 2. Helper unit tests (reject perturbed totals and failing/malformed subprocesses)
go test ./scripts/benchmark

# 3. Loader microbenchmarks with allocations
go test ./internal/loaders -run '^$' -bench 'Benchmark' -benchmem -benchtime 500ms
```

`bench-report.json` embeds the exact command, revision, diff hash, environment, per-source baseline, delta cases, every latency sample, the stream measurements, and the verdicts. Comparisons are only meaningful against runs on matching hardware, fixture dimensions, window lengths, and measurement method; the previously published 100ms/1s CPU budgets were measured on a 30-file fixture over 2-second windows and are reported above as context only.
