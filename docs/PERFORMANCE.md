# Thermal Performance & Ingestion Baseline

This document defines the performance baselines, benchmarking methodologies, and hot-path telemetry for **Thermal** (`github.com/jadmadi/thermal`).

Thermal operates under a strict **Zero-Allocation, Sub-10ms Cached Invocation** architectural invariant codified in [`GOVERNANCE.md`](../GOVERNANCE.md). This report establishes the empirical measurements for end-to-end CLI latency, per-tool ingestion microbenchmarks, and sustained live monitor CPU utilization across reproducible synthetic fixtures.

---

## 1. Executive Summary

| Measurement | Result | Target Budget | Status |
| :--- | :--- | :--- | :--- |
| **Warm Cached CLI Latency (p50)** | **8ms - 9ms** | `<10ms` | **PASS (Sub-10ms met)** |
| **Cold CLI Ingestion (p50)** | **72ms - 73ms** | `<100ms` | **PASS** |
| **Devin Warm Cache Hit** | **755 µs/op** | `<2ms` | **PASS (Sub-1ms)** |
| **Codex Warm Cache Hit** | **568 µs/op** | `<1ms` | **PASS (Sub-millisecond)** |
| **Live Monitor CPU (100ms poll)** | **0.9% CPU** | `<5% CPU` | **PASS** |
| **Live Monitor CPU (1s poll)** | **0.7% CPU** | `<2% CPU` | **PASS** |
| **Model Name Canonicalization** | **21 ns/op (0 B/op)** | `0 allocs` | **PASS (Zero-Alloc)** |

---

## 2. Test Environment & Hardware Specification

All benchmarks were captured on a clean, isolated synthetic fixture environment:

* **CPU**: Intel(R) Core(TM) i7-9700 CPU @ 3.00GHz (8 physical cores)
* **Memory**: 32 GiB DDR4 RAM
* **Operating System**: Linux 6.12 (Debian testing/unstable amd64)
* **Toolchain**: Go 1.27.1 (`linux/amd64`)
* **Binary Flags**: Built with standard Go toolchain (`go build -o ./thermal ./cmd/thermal`)

---

## 3. Workload Fixture Specifications

To ensure reproducibility without leaking personal session logs or requiring proprietary cloud credentials, the benchmark suite operates on standardized synthetic fixtures provisioned by [`scripts/benchmark.sh`](../scripts/benchmark.sh):

| Tool Source | File / Store Type | Volume / Scale | Workload Profile |
| :--- | :--- | :--- | :--- |
| **Devin** | SQLite (`sessions.db`) | 25 sessions, 500 message nodes | Joins `sessions`, `message_nodes`, and `prompt_history` with full token telemetry. |
| **Codex** | SQLite + Multi-file JSONL | 20 threads, 100 rollout chunks | Primary SQLite thread records supplemented by parallel JSONL rollout token chunks. |
| **OpenCode** | SQLite (`opencode.db`) | 50 sessions | Schema `session_v2` storing pre-aggregated input, output, reasoning, and cache tokens. |
| **Claude** | Directory Multi-file JSONL | 30 sessions (30 JSONL files) | Assistant message usage and tool call records parsed across session streams. |
| **CodeWhale** | Directory Multi-file JSON | 20 sessions (20 JSON files) | Workspace metadata and session token / cost attribution. |
| **Pricing** | JSON Catalog (`pricing.json`) | 2 models (Sonnet, GPT-4o) | Local disk-cached pricing catalog operating in `--offline` mode. |

**Fixture Verification**: Total lifetime tokens across the 5 synthetic tools equal exactly **1,873,500 tokens**.

---

## 4. End-to-End CLI Latency Profile

Measured over repeated runs using `scripts/benchmark.sh` against the synthetic workload fixture in `--offline` mode:

```
Cold Ingestion: min=72ms, avg=72ms, p50=73ms, p95=73ms, max=73ms
Warm Ingestion: min=8ms,  avg=8ms,  p50=9ms,  p95=9ms,  max=9ms
Changed/Append: min=71ms, avg=72ms, p50=72ms, p95=77ms, max=77ms
```

### Stage Breakdown Analysis
* **Ingestion (Cold)**: ~65ms. Dominated by initial SQLite database opens, index queries, and JSONL directory file stats.
* **Ingestion (Warm)**: ~2ms. Devin's `DevinCache` returns in 755 µs; overall warm scanning skips re-parsing unchanged sources.
* **Pricing (Offline)**: `<1ms`. Fast memory-mapped hash map lookups against cached `pricing.json`.
* **Rendering & JSON Serialization**: ~5ms. CLI argument routing, JSON encoding, and stdout write.
* **Overall Warm Latency**: **8ms - 9ms**, successfully satisfying the governance budget of `<10ms`.

---

## 5. Ingestion Microbenchmarks (`internal/loaders`)

Run via `go test ./internal/loaders -run '^$' -bench 'Benchmark' -benchmem -benchtime 500ms`:

| Benchmark Function | Iterations | Latency (ns/op) | Memory (B/op) | Allocs (allocs/op) | Notes |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `BenchmarkDevin_Cold` | 8 | 71,451,712 ns/op | 139,938 B/op | 7,909 | Full SQLite join & JSON unmarshal |
| `BenchmarkDevin_Warm` | **759** | **755,694 ns/op** | **25,663 B/op** | **766** | **Sub-millisecond cache hit (<0.76ms)** |
| `BenchmarkDevin_AppendDelta` | 8 | 62,778,731 ns/op | 88,161 B/op | 3,718 | Incremental append scan |
| `BenchmarkCodex_Cold` | 644 | 1,066,804 ns/op | 1,435,490 B/op | 1,764 | SQLite + rollout JSONL parsing |
| `BenchmarkCodex_Warm` | **1,970** | **568,462 ns/op** | **53,938 B/op** | **932** | **Sub-millisecond rollout cache hit (96.4% alloc reduction)** |
| `BenchmarkOpenCode_Scan` | 441 | 1,336,883 ns/op | 22,189 B/op | 347 | SQLite `session_v2` scan |
| `BenchmarkClaude_Scan` | 747 | 804,743 ns/op | 3,375,794 B/op | 810 | Multi-file directory scan |
| `BenchmarkModelName_Canonical` | 27,239,860 | **21.11 ns/op** | **0 B/op** | **0** | **Zero-allocation hot-path** |
| `BenchmarkModelName_Alias` | 22,648,272 | **25.81 ns/op** | **0 B/op** | **0** | **Zero-allocation hot-path** |
| `BenchmarkModelName_Router` | 5,264,727 | **112.00 ns/op** | **0 B/op** | **0** | **Zero-allocation router match** |

---

## 6. Live Monitor CPU & Velocity Profile

Sustained measurements of `thermal live --interval <rate> --json --stream --offline` over 2-second windows:

| Polling Interval | Wall Clock Time | CPU Time (User + Sys) | CPU Utilization (%) | Responsiveness |
| :--- | :--- | :--- | :--- | :--- |
| **100ms** (`--interval 100ms`) | 2.005s | 0.019s (0.016u + 0.003s) | **0.9% CPU** | Instantaneous streaming |
| **1000ms** (`--interval 1s`) | 2.010s | 0.015s (0.007u + 0.008s) | **0.7% CPU** | Low-overhead background watch |

Both interval modes consume under 1% of a single core, well within the target budget of `<5%` for rapid live monitoring.

---

## 7. Hotspot Analysis & Optimization Roadmap

Based on the measured execution profile, subsequent caching efforts are prioritized by empirical cost:

1. **Codex Rollout Transcript Caching (`Goal #69: codex-incremental-rollout-cache`) [COMPLETED]**:
   - *Delivered*: Incremental rollout transcript caching via `internal/loaders/codexcache.go`.
   - *Result*: Cut allocations by 96.4% (from 1.44 MB/op down to 53.9 KB/op) and reduced warm latency to **568 µs/op** (46% faster).
2. **OpenCode / MiMoCode Incremental State Tracking**:
   - *Hotspot*: SQLite queries scan the entire `session_v2` table on every invocation.
   - *Target*: Persist the maximum seen `time_updated` timestamp and delta-merge subsequent sessions.
3. **Directory Walk Overhead for Claude & Droid**:
   - *Hotspot*: Globbing and stating dozens of `.jsonl` files in `~/.claude/projects/` incurs filesystem traversal latency.
   - *Target*: Fast directory modification timestamp check before globbing.

---

## 8. Reproducibility & Verification Guide

To run the full performance verification suite locally:

```bash
# 1. Run unit microbenchmarks with memory allocations
go test -v ./internal/loaders -run '^$' -bench 'Benchmark' -benchmem

# 2. Run end-to-end CLI performance baseline
./scripts/benchmark.sh
```
