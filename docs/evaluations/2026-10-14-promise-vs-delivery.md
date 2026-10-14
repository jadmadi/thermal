# Thermal: promise versus delivery

Evaluated 2026-10-14 in `/mnt/Jad/github/lab/thermal-streak` at source revision
`302501b17f96a26a13b7b5cbaf4e7654199f7d5c`. The README already had 26 added lines;
they were included as current promises and left untouched. No application fixes were made.

## Verdict

Inference: Thermal delivers a useful local usage aggregator, but its strongest decision-support
and performance promises exceed the evidence. The highest-value improvement is trust in results,
not more commands. A passing release gate currently misses important semantic counterexamples.

| Promise | Delivery assessed | Evidence and limit |
| --- | --- | --- |
| Local multi-tool usage and cost views | Substantially delivered on tested fixtures | Five-tool benchmark ingests exactly 1,873,500 tokens. Existing gate passes 107 checks. This is not every schema/version or every tool. |
| Same numbers across views | Delivered for tested paths | Existing static/web/live fixture parity is 759,700 tokens; TUI parity tests pass. Browser appearance and all real histories were not assessed. |
| Verified work receipts | Partially delivered, P1 defect | Skipped compound command becomes test-pass evidence; some successful output becomes failure. Verified outcomes do not establish source-bound task correctness. |
| Complete replay and honest pricing | Not delivered for missing/partial telemetry, P1 | Partial attribution loses 900 of 1000 tokens; unclassified 1000 becomes zero; unavailable rate becomes recommended $0 plan. |
| Rolling five-hour capacity simulation | Not delivered by daily input, P1 | SimulatePlan compares each entire daily sum to RollingWindowLimit; no intraday timestamps participate. |
| Instant, zero-allocation execution | Narrow helper claims supported; blanket claim unsupported | Warm fixture p50 10ms, p95 16ms under concurrent checks. Warm Devin 769 allocations/op; warm Codex 1009. Timing is not isolated regression proof. |
| Setup audit verifies database/mmap health | Overstated, P2 | Plain text named state_5.sqlite is reported PASS, “verified (mmap-ready)”. The code only opens the file. |
| Private read-only ingestion | Supported by inspected architecture, not exhaustively certified | Local sources and read-only SQLite paths exist; no claim of whole-program security audit. Pricing/update network traffic differs from cloud telemetry. |
| Yield measures useful work | Recorded volume ratio delivered; quality inference unsupported | CalculateYield divides tokens by net added lines and applies thresholds. It does not measure correctness, maintainability or business value. |
| Release provenance proves distribution | Workflow exists; released artifacts unverified here | Local release workflow generates provenance for checksums.txt. No release download or verifier was run. |

No current provider prices or subscription entitlements were verified. This is a software contract
assessment, not a recommendation to purchase a plan. No browser UX or legal assessment is claimed.

## Reproduced defects

1. Receipt, `internal/thermal/receipt.go:ParseCommandEvidence`: input
   `false && go test ./... || true`, outer exit 0, empty output becomes `go test (exit 0)`.
   The test invocation was short-circuited. Conversely, an unknown exit with the completed line
   `The command exited with code 0`, or pytest summary `10 passed, 0 failed`, becomes exit 1.
   Positive/negative controls confirm explicit successful single commands pass and empty missing
   results remain unknown. The September unfinished-result defect is not simply assumed to persist.
2. Replay, `internal/thermal/replay.go:BuildWorkloadSnapshot`: canonical day total 1000,
   Input 800/Output 200, model map covering only 100 -> snapshot total 100 but peak 1000.
   A model with 1000 Unclassified tokens -> total zero. The latter shape is emitted by
   `internal/loaders/sessions.go` for codewhale, not just a hypothetical type construction.
3. Replay, `AggregateReplay`/`SimulateModel`: missing target price, 1000 source tokens and $2
   recorded spend on a one-day window -> MonthlyCost 0, CostDelta -60, CostDeltaPercent -100,
   CapacityVerdict PASS, IsRecommended true. Missing price is presented as free.
4. Audit, `internal/audit/audit.go:checkToolDatabaseAccess`: a readable plain-text database path
   returns “1 local tool databases verified (mmap-ready)”. This proves readability only. Writable
   cache similarly becomes “sub-10ms ready”, which is not a latency measurement.

Source finding: `SimulatePlan` compares `dayRow.Tokens` to `RollingWindowLimit` and counts burst
throttles. Daily totals cannot distinguish the same volume spread across 24 hours from one burst,
or resolve cross-midnight windows. The bounded remediation discloses UNKNOWN, rather than inventing
fine-grained data or building a new telemetry system.

## Performance evidence quality

`scripts/benchmark.sh` exited 0: cold p50 74ms, warm p50 10ms/p95 16ms, changed p50 74ms.
The five-tool fixture happened to match its documented 1,873,500 total, but the runner only asserts
positive total and at least five tools. “Cold” clears only Devin snapshots. Changed-input timing
adds a differently shaped metrics record and does not assert a token delta. CPU samples last two
seconds and tolerate any live process failure via `|| true`. No latency threshold is enforced.

Observed 100ms/1s idle samples were 1.0%/0.6% CPU. These short, small-fixture measurements neither
reproduce the old large-history CPU defect nor prove it solved at scale. No CPU regression is
asserted. Historical docs/PERFORMANCE.md numbers remain historical, not current guarantees.

## Verification actually run

- CLI build to `/tmp/thermal-promise`: exit 0 after unsetting stale Go 1.27.1 GOROOT for Go 1.27.2.
- `env -u NO_COLOR go test -race ./...`: exit 0 with loopback access; thermal 94.573s, TUI 182.055s.
- `go vet ./...`: exit 0.
- Existing simulated gate: exit 0, 107/107. First sandbox run failed only the listener check;
  the permitted rerun passed. Scripts were copied with binary output redirected to /tmp only.
- Existing benchmark runner: exit 0; timings were collected alongside verification, not in a
  controlled idle-machine regression experiment.
- New one-off Go overlay probes: exit 1, intentionally exposing seven semantic failures and three
  passing controls. No failing probe was installed in the normal application test suite.

[Raw logs and runnable probes](2026-10-14-promise/README.md) bind findings to this candidate.
Tests passing and tracker attribution are kept separate from acceptance of the stronger promises.

## Improvement goals

Design choice: repair existing outcomes in place, preserving historical commits and checked tasks.
No goal was claimed or certified. Added unchecked remediation tasks and current evidence links.

| Order | Goal | Status and purpose |
| --- | --- | --- |
| 1 | [receipt-evidence-integrity](../../goals/receipt-evidence-integrity/remediation-2026-10-14.md) | Reopened P1: observed single-command evidence, no skipped-test success. |
| 2 | [workload-replay](../../goals/workload-replay/remediation-2026-10-14.md) | Reopened P1: volume conservation and explicit price coverage. |
| 3 | [chatgpt-rolling-burst-replay](../../goals/chatgpt-rolling-burst-replay/remediation-2026-10-14.md) | Reopened P1: disclose unknown rolling capacity; depends on replay. |
| 4 | [setup-health-audit](../../goals/setup-health-audit/remediation-2026-10-14.md) | Reopened P2: diagnostic wording matches the checks performed. |
| 5 | [ingestion-performance-baseline](../../goals/ingestion-performance-baseline/remediation-2026-10-14.md) | Reopened P2: exact fixtures, sustained measurements, scoped claims. |
| 6 | [promise-claim-evidence](../../goals/promise-claim-evidence/plan.md) | New P2: claim ledger and semantic gate after all remediation. |

[Execution plan and handoff contract](../../tasks/plan.md) define scope, acceptance, return fields
and independent review. Recommended next action: implement receipt-evidence-integrity.

## Tooling and registry limits

Initial live Sila inventory: 83 implemented goals, zero current verified, eight stale and 75
unverified. The checked-in index had reported seven verified; live verification state differed.
This is evidence freshness, not proof that all those implementations are defective.

Sila heal attempted setup but refused an AGENTS.md over its 24KB ceiling; rules were preserved and
feedback recorded. Goal inventory and mutation commands still worked using the workspace store.
Global validator has pre-existing errors: missing html-docs-licensing-parity dependency and missing
goal.md/facts.md for mcp-server-card, plus 20 sequence warnings. New/reopened packages are checked
separately and registry synchronization does not erase those unrelated defects.

Final Sila readback: 84 packages, five explicitly reopened and the new package pending. Dependency
propagation also taints downstream completed packages; therefore the global pending count is nine,
not six. Validation reports the same three pre-existing errors and 23 warnings: the original 20
sequence warnings plus three checked-task warnings exposed by prerequisite reopening. None targets
a new/remediated package. Dependencies were read back from `sila goals --no-sync --json`.
