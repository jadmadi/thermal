# Thermal Public Promises & Evidence Ledger

This ledger maps Thermal's strong public claims to the behavior that backs them, a
reproducible command or artifact, the source revision it was verified at, and — just as
important — the limits that are *not* covered. It exists so users can tell observed
local activity apart from inference, and so a passing test suite is never mistaken for
untested semantics.

**Status legend**

| Status | Meaning |
| :--- | :--- |
| **verified** | The behavior is implemented and has reproducible evidence at the recorded revision. |
| **partial** | The claim holds only within a stated scope (fixture, machine, source subset) or with qualifiers. |
| **unverified** | No reproducible evidence was produced; the claim must be read as intent or workflow definition. |

**Evidence base (2026-10-15, revision `abc2169`)**

- [2026-10-14 promise-vs-delivery evaluation](evaluations/2026-10-14-promise-vs-delivery.md)
- Remediation receipts and signed audits:
  [receipt evidence](../goals/receipt-evidence-integrity/remediation-result.json),
  [workload replay](../goals/workload-replay/remediation-result.json),
  [rolling-window capacity](../goals/chatgpt-rolling-burst-replay/remediation-result.json),
  [setup audit](../goals/setup-health-audit/remediation-result.json),
  [ingestion performance](../goals/ingestion-performance-baseline/remediation-result.json)
- [Performance baseline report](PERFORMANCE.md) · [Migration guide](MIGRATION.md)

---

## 1. Local, private, read-only ingestion

**Wording:** "All local, zero cloud telemetry" (README), "Thermal never modifies your data"
(README), zero-telemetry anti-goals in [GOVERNANCE.md](../GOVERNANCE.md).

**Behavior:** every SQLite source is opened `file:…?mode=ro` with `PRAGMA mmap_size`; file
scanners only read; no telemetry endpoint exists; the daily update check is opt-out and
`--offline` keeps pricing on disk.

**Evidence:** `rg -n 'mode=ro' internal/loaders` · `go test ./internal/loaders` ·
gate offline checks (`--offline` pricing) · source-tree absence of any upload path.

**Limits:** the update check and pricing refresh contact GitHub/models.dev unless disabled;
session *content* stays on disk, but the network calls themselves are not "offline by default".

**Outcome: verified** for read-only, no-telemetry ingestion.

## 2. Token and cost coverage

**Wording:** "Unified cross-agent disjoint token accounting … with recorded vs estimated cost
breakdown" (README, project.jsonc); disjointness migration entry
[MIGRATION.md#changed-disjoint-token-metrics-accounting](MIGRATION.md#changed-disjoint-token-metrics-accounting).

**Behavior:** loaders normalize per-tool schemas into disjoint input/output/reasoning/cache
read/cache write categories, canonicalize model ids before they become keys, prefer recorded
cost, and mark estimates with `~`; unpriced models are named in the footer instead of being
treated as free.

**Evidence:** `go test ./internal/thermal ./internal/loaders` · gate token parity assertion
(`759,700` on the synthetic fixture) · pricing tests (`go test ./internal/pricing`).

**Limits:** coverage depends on what each tool version records; activity-only tools carry
steps, not tokens; tokens with no recorded model are counted but only estimated (or flagged
unattributed), never silently priced.

**Outcome: verified** for covered sources; **partial** across all schema/version combinations.

## 3. Cross-surface totals

**Wording:** "Four surfaces render the same numbers" (README Why Thermal).

**Behavior:** CLI reports, JSON export, embedded web telemetry, live monitor and TUI views
aggregate through the same canonical functions; a test asserts TUI totals equal the matching
static command.

**Evidence:** gate check "Static/Web/Live token parity verified (759,700 tokens)" ·
`go test ./internal/tui` · `go test ./internal/server`.

**Limits:** parity is verified on synthetic fixtures; browser-rendered appearance and every
real history are not assessed.

**Outcome: verified** (fixture-scoped).

## 4. Verified work receipts (`thermal receipt`)

**Wording:** "verifies test and lint outcomes with tiered confidence" (README).

**Behavior:** receipts classify **observed command outcomes** from session transcripts —
token-boundary runner recognition, explicit nonzero exits win, compound shell commands stay
unknown, completed result lines can resolve an unrecorded exit. Tier 1 means an observed
zero-exit test/linter/commit result, never an agent claim.

**Evidence:** `go test -race ./internal/thermal -run 'Receipt|Evidence|Outcome'` ·
[receipt remediation result](../goals/receipt-evidence-integrity/remediation-result.json) and
its [signed audit](../goals/receipt-evidence-integrity/audit-receipt.json) ·
[MIGRATION.md#changed-receipt-evidence-token-boundary](MIGRATION.md#changed-receipt-evidence-token-boundary).

**Limits:** receipt sources are Claude, Agy, codewhale and Codex — a subset of the full
ingestion registry listed in the README. A Tier 1 receipt is evidence that a command was
observed to succeed; it is **not** proof that the task was correct, complete, or committed.

**Outcome: verified** as observed command outcomes; **partial** as a coverage claim.

## 5. Token yield (`thermal yield`)

**Wording:** "measures tokens burned per net line of code" (README).

**Behavior:** parses recorded diff/patch volume from transcripts and computes tokens per net
line added, with session fork/mainline lineage where the source provides it.

**Evidence:** `go test ./internal/thermal -run Yield` · render tests `go test ./internal/render -run Yield`.

**Limits:** yield is a **recorded diff-volume ratio**. It does not measure correctness,
maintainability, review effort, or business productivity, and deleted/rewritten code is only
as well recorded as the transcript's diff data.

**Outcome: verified** as a volume ratio; **unverified** as a productivity claim (not claimed).

## 6. Workload replay (`thermal replay`)

**Wording:** "simulates your real workload against subscription ceilings and API rate cards" (README).

**Behavior:** a scenario comparison: canonical volume is conserved, unclassified tokens are
disclosed, target pricing is labeled `complete` / `partial` / `missing` (rendered as `Unpriced`
or `Partial estimate`), incomplete rows never carry a savings claim, and rolling-window
capacity that daily telemetry cannot assess reports `UNKNOWN`.

**Evidence:** `go test -race ./internal/thermal ./internal/render -run Replay` ·
[replay remediation result](../goals/workload-replay/remediation-result.json) ·
[capacity remediation result](../goals/chatgpt-rolling-burst-replay/remediation-result.json)
and their signed audits ·
[MIGRATION.md#changed-replay-volume-and-price-coverage](MIGRATION.md#changed-replay-volume-and-price-coverage).

**Limits:** plan limits and rate cards are **configured assumptions**, not verified provider
entitlements; replay cannot establish code quality or actual bills. Exact five-hour simulation
is deferred until intraday timestamp coverage has its own design.

**Outcome: verified** as a scenario comparison with explicit coverage labels.

## 7. Performance (sub-10ms cached execution)

**Wording:** "sub-10ms cached execution on the baseline fixture" (README, qualified);
sub-10ms target in [GOVERNANCE.md](../GOVERNANCE.md).

**Behavior and evidence:** [PERFORMANCE.md](PERFORMANCE.md) publishes a full run at revision
`e2cdc72`: warm p50 **7.8ms** / p95 9.1ms (n=100) on the five-source fixture — **within** the
10ms target — and warm p50 **42.4ms** (n=20) at 1,000-file directory scale, which **exceeds**
the target and is disclosed as a scale gap. `./scripts/benchmark.sh` reproduces the run and
fails nonzero on fixture mismatch, malformed output, or a dead live stream.

**Limits:** one machine (i7-9700, Linux, Go 1.27.2), synthetic fixtures; not a universal
latency guarantee. Live-monitor CPU and source freshness have their own per-machine verdicts
in the same report. TTY responsiveness was not measured.

**Outcome: partial** — verified on the baseline fixture, disclosed gap at directory-scan scale.

## 8. Setup & context health audit (`thermal audit`)

**Wording:** "setup and permissions audit", "database mmap-readiness" (historical README text).

**Behavior:** the audit measures rule-file/MCP context tax, repository instruction hygiene,
**filesystem readability** of tool database paths (SQLite integrity and mmap are *not*
checked), and a temporary writability probe in Thermal's own cache directory.

**Evidence:** `go test -race ./internal/audit -run Audit` ·
[audit remediation result](../goals/setup-health-audit/remediation-result.json) and its
[signed audit](../goals/setup-health-audit/audit-receipt.json).

**Limits:** readability is not database validity; the audit writes a temporary probe only to
Thermal's cache, never to agent stores. Pricing-cache integrity and transcript-ceiling checks
are not performed and are no longer advertised.

**Outcome: verified** under the narrowed scope.

## 9. Distribution and build provenance

**Wording:** SLSA Level 3 badge and "All official releases provide cryptographic build
provenance … generated via OpenSSF" (README).

**Behavior:** the release workflow (`.github/workflows/release.yml`) generates checksums and
OpenSSF in-toto provenance on `v*` tags.

**Evidence:** workflow definition in-tree; the publication runbook in
[DISTRIBUTION.md](DISTRIBUTION.md).

**Limits:** a workflow definition is not a verified artifact. In this audit no released
artifact was downloaded and verified with `slsa-verifier`, so provenance is **unverified**
here; the README now points to this ledger for that status.

**Outcome: unverified** (workflow-defined, artifact-level verification pending).

## 10. Allocation claims

**Wording:** "zero allocations" (historical docs); performance invariant.

**Behavior and evidence:** [PERFORMANCE.md](PERFORMANCE.md) measures the model-name helper
benchmarks at **0 B/op, 0 allocs/op**, and records whole-process loader paths that **do**
allocate (e.g. `BenchmarkDevin_Warm` 25,772 B/op, `BenchmarkCodex_Warm` 58,728 B/op).

**Limits:** zero-allocation claims apply to the measured helpers only, never to the whole CLI
process.

**Outcome: partial** — verified for the named helpers, explicitly not a whole-binary claim.

---

**Maintenance:** when a public claim changes, update its row here in the same commit as the
wording change. A claim with no evidence command and no stated limit is a documentation bug.
Claims whose verification depends on a released artifact (provenance, installers) stay
**unverified** until an artifact is independently verified, and the README must say so.
