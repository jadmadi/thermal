# Thermal improvement plan — 2026-09-24

The [evaluation](../docs/evaluations/2026-09-24-thermal.md) records working-tree evidence and verification limits. Tasks are tracked exclusively in the ten Sila goal packages; this file is their ordered index, not a second checklist.

## Direction

Bound live-monitor CPU at 100ms first. Restore trustworthy receipts and consistent metrics before expanding features. Isolate Devin caches by source. Protect local HTTP request authority. Use deterministic verification and measure ingestion before optimizing it. Preserve the current architecture and all unrelated work.

## Task index

60. [Keep 100ms live monitoring within a CPU budget](../goals/live-polling-cpu-budget/goal.md) — P1; ready independently; first recommended.

61. [Require observed command results for verified receipts](../goals/receipt-evidence-integrity/goal.md) — P1; ready independently.
62. [Preserve recorded usage and disclose receipt coverage](../goals/receipt-accounting-coverage/goal.md) — P1; depends on receipt-evidence-integrity.
63. [Make web totals match canonical reports](../goals/web-telemetry-parity/goal.md) — P1; ready independently.
64. [Make live burn metrics reflect recorded tokens and turns](../goals/live-metric-parity/goal.md) — P1; depends on web-telemetry-parity, live-polling-cpu-budget.
65. [Keep cached Devin totals isolated by source database](../goals/devin-cache-source-isolation/goal.md) — P1; ready independently.
66. [Reject untrusted hosts at the local telemetry API](../goals/localhost-request-boundary/goal.md) — P1; depends on web-telemetry-parity.
67. [Make verification independent of host data and color settings](../goals/hermetic-verification-gates/goal.md) — P2; depends on receipt-accounting-coverage, live-metric-parity, devin-cache-source-isolation, localhost-request-boundary, live-polling-cpu-budget.
68. [Measure cold and warm CLI performance on representative fixtures](../goals/ingestion-performance-baseline/goal.md) — P2; ready independently.
69. [Avoid rescanning unchanged Codex rollout transcripts](../goals/codex-incremental-rollout-cache/goal.md) — P2; depends on ingestion-performance-baseline.

## Checkpoints and ownership

After live polling: 100ms display remains usable, unchanged history is reused, at most one collection is in flight, pause/quit remain responsive under a slow collector, and TTY/NDJSON CPU measurements meet the documented fixture budget.

After receipt evidence and accounting: unfinished commands remain unverified, source coverage is explicit, and no tokens are invented.

After web/live parity and Devin isolation: the same fixture has identical metric totals across views, and switching database paths cannot reuse stale totals.

After request-boundary work: allowed local access works while foreign authorities are rejected before data collection. After the dependent gate goal: these behaviors are enforced through user-level assertions in an isolated environment.

Performance baseline and Codex caching form a separate sequence. Benchmark evidence must support the optimization and compare exact output parity.

Each package lists files, acceptance criteria, tests and exclusions. Shared docs/MIGRATION.md edits require coordination even between independent goals. A changed public contract gets the mandated migration entry. A planning commit does not close implementation goals.

## Status

Ten newly registered goals, 50 unchecked tasks/checkpoints, five ready and five dependency-blocked after the CPU follow-up. None claimed or implemented by this evaluation. Refer to `sila goals` for subsequent live state.
