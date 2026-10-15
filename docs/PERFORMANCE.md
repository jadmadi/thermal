# Thermal Performance & Ingestion Baseline

This document defines the performance baselines, benchmarking methodology, and hot-path telemetry for **Thermal** (`github.com/jadmadi/thermal`).

Thermal operates under a **Sub-10ms Cached Invocation** architectural target codified in [`GOVERNANCE.md`](../GOVERNANCE.md). This report is bounded evidence for that target on one machine and one synthetic fixture family — it is not a universal latency guarantee. Every number below is reproducible with [`scripts/benchmark.sh`](../scripts/benchmark.sh), and correctness failures stop the run before any performance number is measured.

---

## 1. Evidence binding

| Field | Value |
| :--- | :--- |
| Source revision | `e2cdc72` (`e2cdc72080debb591869da26383bd70d19d7c168`) |
| Working tree | dirty (pre-existing documentation edits only; diff sha256 `6000d420004d5a1e4ae7b30eccf4a54d0369bb148b178ed48b6dfb2e201f3a95`) |
| Toolchain | Go 1.27.2 (`linux/amd64`) |
| CPU | Intel(R) Core(TM) i7-9700 CPU @ 3.00GHz (8 cores) |
| OS | Linux (`linux/amd64`) |
| Command | `./scripts/benchmark.sh --out /tmp/thermal-bench-final` |
| Raw artifacts | `bench-report.json` and `bench-report.txt` written by the run (JSON mirrors this document) |

The benchmark binary and all fixtures are built under the operating system's temporary directory. No binary is written to the repository root, and no user data store is read or written: the run uses an isolated `HOME` plus every tool-root override (`GROK_HOME`, `DSH_HOME`, `HERMES_HOME`, `CODEX_HOME`) and XDG path the loaders consult.

---

## 2. Executive summary — per-machine verdicts

| Measurement | Result (this machine) | Budget | Verdict |
| :--- | :--- | :--- | :--- |
| **Warm cached CLI latency, small fixture (n=100)** | **p50 7.8ms · p95 9.1ms** | `<10ms` | **PASS** |
| **Warm cached CLI latency, 1,000-file fixture (n=20)** | p50 42.4ms · p95 52.9ms | `<10ms` | **MISSED — disclosed scale gap** |
| Cold ingestion (n=30, all snapshots reset) | p50 11.5ms · p95 13.8ms | none published | OBSERVED |
| Live monitor CPU at 100ms (60s window) | 1.32% CPU (0.857s CPU / 65.0s process wall) | `<5%` (previously published on a 30-file fixture over 2s) | OBSERVED — not directly comparable |
| Live monitor CPU at 1s (60s window) | 0.67% CPU (0.436s CPU / 65.0s process wall) | `<2%` (same caveat) | OBSERVED — not directly comparable |
| Source freshness at 100ms | append reflected after 133ms | none published | OBSERVED |
| Source freshness at 1s | append reflected after 909ms | none published | OBSERVED |
| Cancellation latency (SIGINT) | 1.1ms both intervals | none published | OBSERVED |
| Model-name helper allocations | 21.3 / 26.1 / 106.5 ns/op, **0 B/op, 0 allocs/op** | `0 allocs` | PASS (helpers only) |

The governance target holds for the five-source baseline fixture (73 workload files). At 1,000-file scale the warm p50 is ~5.4× the budget; that gap is measured, disclosed here, and left open — this measurement card does not change loaders to make the number pass.

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

Method: each sample is a full `thermal --json --offline` subprocess with stdout sunk, timed with Go's monotonic clock. **Cold** removes every Thermal snapshot artifact in the fixture cache (`devin/`, `codex/`, `zed/`, snapshots) while preserving the cached pricing catalog — not just the Devin cache; the fixture home is the only cache location touched. **Warm** runs leave snapshots in place.

```
Cold  n=30  min 10.8ms  mean 11.7ms  p50 11.5ms  p95 13.8ms  max 15.2ms
Warm  n=100 min  6.7ms  mean  7.9ms  p50  7.8ms  p95  9.1ms  max 10.5ms
Large n=20  min 39.1ms  mean 42.8ms  p50 42.4ms  p95 52.9ms  max 52.9ms
```

Raw samples (milliseconds):

```text
cold (30): 13.813 11.959 11.994 12.137 11.512 13.258 12.362 11.430 10.891 11.768
           10.975 10.885 11.324 11.924 10.851 11.253 11.698 11.017 11.320 11.017
           11.107 11.481 11.050 10.836 11.133 15.193 11.484 11.906 11.885 11.252

warm (100): 8.027 7.832 7.755 7.374 7.475 7.462 7.472 8.324 7.995 7.249
            7.610 7.498 7.982 7.799 7.749 7.772 8.912 6.743 8.128 7.367
            7.557 8.173 7.518 6.955 7.618 7.744 9.069 7.793 7.288 7.307
            7.879 6.989 7.850 9.995 7.558 7.809 7.738 7.531 8.006 10.287
            7.386 7.832 7.734 7.688 7.362 7.570 8.075 8.094 7.604 7.379
            7.105 8.082 7.258 8.137 7.385 7.846 7.904 7.404 8.709 8.164
            8.423 9.040 10.501 7.919 7.968 8.059 7.488 7.648 9.285 8.306
            7.778 8.179 7.940 7.877 7.750 7.641 7.490 7.594 8.256 8.358
            7.803 7.591 7.982 8.160 8.537 8.321 8.438 8.199 8.850 7.928
            8.347 8.344 7.888 6.942 8.832 7.221 7.526 7.395 8.510 7.553

large n=20: 39.143 40.997 41.161 42.883 41.423 42.886 39.062 42.410 42.178 47.314
            42.632 40.407 41.569 42.809 46.856 42.462 42.190 39.693 52.915 44.485
```

**Interpretation.** On the five-source fixture the warm p50 (7.8ms) meets the 10ms governance target and the cold p50 (11.5ms) is close to it. On the 1,000-file fixture the warm p50 (42.4ms) exceeds the target: directory-backed sources (Claude JSONL, CodeWhale JSON) are re-walked per invocation and have no incremental snapshot cache, while SQLite-backed sources (Devin, Codex) do. That is the disclosed scale gap, not a claim that the target is universally met or universally missed.

---

## 5. Sustained Live Monitor (NDJSON)

Method: `thermal live --json --stream --interval <rate> --offline` against the 1,000-file fixture for a 5s warm-up plus a 60s measured window. One source file is appended a third of the way into the window (with a 137ms offset so the sample does not land on a poll-tick boundary); freshness is the time from that append until the change event line arrives. CPU is the child's user+system rusage over the process lifetime (warm-up included, disclosed). Termination is a deliberate SIGINT; a stream that exits on its own before cancellation, emits malformed NDJSON, or never reflects the appended source fails the run with a nonzero exit.

| Interval | Window | CPU | Change events | Invalid lines | Freshness | Cancellation |
| :--- | ---: | ---: | ---: | ---: | ---: | ---: |
| 100ms | 60s | 0.857s = **1.32%** of one core | 1 | 0 | **133ms** | 1.1ms |
| 1s | 60s | 0.436s = **0.67%** of one core | 1 | 0 | **909ms** | 1.1ms |

The stream emits the initial snapshot line and then one line per detected source change; the measured runs therefore recorded 2 NDJSON lines each (initial snapshot + the appended source's change event).

**Not measured:** TTY responsiveness. These runs exercise NDJSON output on a pipe; NDJSON is not TTY evidence, and no interactive terminal frame rate or input-latency claim is made here.

---

## 6. Allocations: whole process vs helpers

The zero-allocation invariant applies to specific hot-path helpers, **not** to the whole CLI process. `go test ./internal/loaders -run '^$' -bench Benchmark -benchmem -benchtime 500ms` on the same machine:

| Benchmark | Iterations | ns/op | B/op | allocs/op | Note |
| :--- | ---: | ---: | ---: | ---: | :--- |
| `BenchmarkModelName_Canonical` | 28,116,436 | 21.33 | **0** | **0** | zero-allocation helper |
| `BenchmarkModelName_Alias` | 23,090,486 | 26.08 | **0** | **0** | zero-allocation helper |
| `BenchmarkModelName_Router` | 5,647,664 | 106.5 | **0** | **0** | zero-allocation helper |
| `BenchmarkDevin_Cold` | 64 | 8,182,605 | 138,355 | 7,893 | whole-process loader work |
| `BenchmarkDevin_Warm` | 769 | 870,425 | 25,772 | 769 | cache hit still allocates |
| `BenchmarkDevin_AppendDelta` | 94 | 6,158,701 | 95,551 | 4,267 | incremental append |
| `BenchmarkCodex_Cold` | 542 | 1,159,146 | 1,488,701 | 2,025 | SQLite + rollout JSONL |
| `BenchmarkCodex_Warm` | 930 | 651,674 | 58,728 | 1,008 | rollout cache hit |
| `BenchmarkOpenCode_Scan` | 385 | 1,502,478 | 24,058 | 368 | `session_v2` scan |
| `BenchmarkClaude_Scan` | 706 | 834,156 | 3,376,845 | 812 | directory scan |

Whole-process invocations allocate (JSON decoding, token maps, rendering); the loaders' warm paths allocate tens of KB per scan, not zero. Any public claim of "zero allocations" must be read as scoped to the model-name helpers above.

---

## 7. Hotspot analysis and bounded follow-ups

Follow-ups below are derived from the measured costs in this report. No loader change ships with this measurement card.

1. **Directory-backed sources at scale (largest measured gap).** Claude/CodeWhale directory scans are re-walked on every invocation, which is the dominant cost at 1,000 files (large-fixture warm p50 42.4ms vs 7.8ms small fixture). A bounded follow-up worth designing is a fast directory modification check or an incremental scan cache for directory loaders, with the same read-only and exactness guarantees as the SQLite snapshot caches.
2. **Unchanged-source re-use in the live collector.** At 100ms cadence the collector re-scans directory sources on its ~200ms budget; measured sustained CPU is 1.32% (100ms) vs 0.67% (1s) at 1,000-file scale. Any change here must preserve source freshness (measured 133ms / 909ms) and cancellation behavior.
3. **Warm SQLite path allocations.** Cache-hit scans still allocate tens of KB (`BenchmarkDevin_Warm` 25,772 B/op); this is acceptable today and only worth revisiting against a measured regression.

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
