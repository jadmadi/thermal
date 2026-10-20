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

**Evidence base (2026-10-19, revision `70ad214`)**

- [2026-10-14 promise-vs-delivery evaluation](evaluations/2026-10-14-promise-vs-delivery.md)
- Remediation receipts and signed audits:
  [setup audit](../goals/setup-health-audit/remediation-result.json),
  [ingestion performance](../goals/ingestion-performance-baseline/remediation-result.json)
- [Performance baseline report](PERFORMANCE.md) · [Migration guide](MIGRATION.md)
- [Release provenance verification (v0.20.0)](evaluations/2026-10-15-release-provenance/commands.log)

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

**Wording:** "CLI reports and the live monitor render the same numbers" (README Why Thermal).

**Behavior:** CLI reports, JSON export, and the live monitor aggregate through the same
canonical functions; a test asserts live totals equal the matching static command.

**Evidence:** gate check "Static/Live token parity verified (759,700 tokens)" ·
`go test ./internal/tui` · `./scripts/output_gallery_gate.sh` (fixed fixture
matrix: leaderboard/projects/share equal daily, models classified+unclassified
equal daily, tool-mix cost equal daily).

**Limits:** parity is verified on synthetic fixtures; every real history is not assessed.
The gallery gate's equality checks are fixture-scoped and do not prove a user's
mixed-source history is free of attribution gaps.

**Outcome: verified** (fixture-scoped).

## 4. Performance (sub-10ms cached execution)

**Wording:** "sub-10ms cached execution on the baseline fixture" (README, qualified);
sub-10ms target in [GOVERNANCE.md](../GOVERNANCE.md).

**Behavior and evidence:** [PERFORMANCE.md](PERFORMANCE.md) publishes a full run at revision
`78b5d70`: warm p50 **5.9ms** / p95 6.7ms (n=100) on the five-source fixture and warm p50
**5.5ms** / p95 5.8ms (n=20) at 1,000-file directory scale — **within** the 10ms target on
both scales. The earlier `e2cdc72` run had measured 42.4ms at directory scale and disclosed
it as a scale gap; the directory scan cache closed that gap without changing results (freshness
still measures 81ms/886ms in the live monitor). `./scripts/benchmark.sh` reproduces the run and
fails nonzero on fixture mismatch, malformed output, or a dead live stream.

**Limits:** one machine (i7-9700, Linux, Go 1.27.2), synthetic fixtures; not a universal
latency guarantee. Live-monitor CPU and source freshness have their own per-machine verdicts
in the same report. TTY responsiveness was not measured. The directory cache's stat-based
signature cannot observe a rewrite that preserves both byte size and modification time.

**Outcome: verified** on both fixture scales (per-machine, fixture-scoped).

## 5. Setup & context health audit (`thermal audit`)

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

## 6. Distribution and build provenance

**Wording:** SLSA Level 3 badge and "All official releases provide cryptographic build
provenance … generated via OpenSSF" (README).

**Behavior:** the release workflow (`.github/workflows/release.yml`) generates checksums and
OpenSSF in-toto provenance on `v*` tags.

**Evidence:** independently verified on 2026-10-15 against the published `v0.20.0` release:
`slsa-verifier` (installed via `go install github.com/slsa-framework/slsa-verifier/v2/cli/slsa-verifier@latest`)
verified `checksums.txt` against `checksums.txt.intoto.jsonl` — builder
`slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@refs/tags/v2.1.0`,
source commit `3f5fa6d9c047dd093f32eaa21651f0c858cded5a` — and `sha256sum --check` verified the
downloaded `thermal_0.20.0_linux_amd64.tar.gz`. Raw commands and outputs:
[evaluations/2026-10-15-release-provenance/commands.log](evaluations/2026-10-15-release-provenance/commands.log).

**Limits:** the verification covers the artifacts downloaded in that run (the Linux/amd64
archive and the shared `checksums.txt`); the other platform archives share the same checksums
file and provenance but were not individually downloaded. Releases published before this
verification have not been re-checked.

**Outcome: verified** for `v0.20.0` (artifact-level, reproducible).

## 7. Commercial Use License ($30 per developer per year)

**Wording:** "Commercial Use License — $30 per developer per year" (README, `COMMERCIAL-LICENSE.md`, `thermal license`).

**Behavior:** the license grants internal commercial use by named developers for a 12-month term. It excludes redistribution, embedding, hosted services, and support SLAs. Payment is by PayPal or invoice; a certificate is issued by email. Enterprise & OEM terms are separate.

**Evidence:** `COMMERCIAL-LICENSE.md` · [live pricing page](https://thermal.jadmadi.net/pricing.html) · `thermal license` and `thermal license --json` · simulated user gate license checks.

**Limits:** fulfillment is manual; there is no automated checkout or license-key service. The license includes no SLA and no redistribution or OEM rights. It does not change AGPL-3.0 rights or the Community Reciprocity Covenant.

**Outcome: verified** as a published offer and terms; **unverified** as a demand or conversion claim.

## 8. Allocation claims

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
The commercial offer is a public promise: price or scope changes must update
`COMMERCIAL-LICENSE.md`, the README, the landing page, and `thermal license` together.
The hosted documentation under `docs/pages/` mirrors these qualified claims (performance,
audit, and caching wording aligned after this ledger was introduced); the
docs-page-claim-parity goal tracks that in lockstep with any future claim change.
