# Plan: Make window and repeated-flag semantics explicit and non-silent

## Execution Steps
- [ ] **Phase 1 — LastSpec and parsing**: Add `LastSpec`, suffix parsing, and
  unit-aware errors; keep bare integers on their current per-report unit.
  Acceptance: `TestParseLast`, `TestLastUnitMatrix`, and
  `TestLastInvalidForms` pass.
  - `7`, `7d`, `4w`, `3m`, case-insensitive.
  - `0`, `-1`, `1.5`, `7x` are usage errors with a range and example.
  - `changelog --last` stays a release count.
- [ ] **Phase 2 — One start-date resolver**: Add `LastWindowStart` and route
  `Aggregate`, `FilterDailyRows`, `AggregateStats`, `AggregateTrend`, and
  `AggregateToolMix` through it; remove grain-specific multiplication.
  Acceptance: `TestWindowStartAllGrains`, `TestWindowTotalsParity`, and all
  existing report tests pass.
  - Frozen-clock table for first/last included day.
  - `weekly --last 14d` equals `daily --last 14` totals.
- [ ] **Phase 3 — Repeated-flag warnings**: Change `present` to a count, emit
  one stderr line per repeated singleton, and keep booleans silent.
  Acceptance: `TestRepeatedScalarFlags`, `TestRepeatedBooleanSilent`, and
  `TestRepeatedFlagJSONStdout` pass.
  - Warning names the flag, the count, and the last value.
  - `--json` stdout is unchanged.
- [ ] **Phase 4 — `--` diagnostics and examples**: Explain flag-like
  positionals after `--`, replace `<path>` with a real path, and add `sh -n`
  coverage. Acceptance: `TestDoubleDashDiagnostics` and
  `TestExamplesShellSafe` pass.
- [ ] **Phase 5 — Docs, gate, evidence, closure**: README, MIGRATION,
  `project.jsonc`, simulated-user gate, evidence, commit, certify, reconcile.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | `weekly --last 7` | `last 7 weeks` |
| AC-002 | `weekly --last 14d` | `last 14 days` |
| AC-003 | `monthly --last 3m` vs `3` | same window and totals |
| AC-004 | `daily --last` | exit 2, unit-aware example |
| AC-005 | `weekly --last 7x` | exit 2, suffix list |
| AC-006 | repeated `--tool` | one stderr warning, last wins |
| AC-007 | repeated `--json` | silent |
| AC-008 | `weekly -- --offline` | `--` explanation, exit 2 |
| AC-009 | all `example:` lines | `sh -n` passes |
| AC-010 | help/error grep | no `<path>` |
| AC-011 | `go test -race ...` | exit 0 |
| AC-012 | simulated user gate | exit 0 |

## Rollback
- Keep the suffix parser additive; if a report cannot honor a suffix, return a
  usage error naming the supported units rather than silently ignoring it.
- If repeated-flag warnings are too noisy, emit them only under `--verbose` in
  a follow-up; do not restore silent discards.
- If `--` diagnostics conflict with a valid tool name that begins with `-`,
  keep the POSIX behavior and adjust the message only.
- Do not change bare `--last N` behavior without a major-version MIGRATION.

## Phase 6 — Attribution & Registry Closure
- [ ] **Phase 6 — Attribution & Registry Closure**: (1) check every task
  above; (2) commit with
  `fix(cli): make window units and repeated flags explicit
  (goals/cli-window-flag-semantics/goal.md)`; (3) certify via
  `sila goals certify cli-window-flag-semantics
  --evidence=goals/cli-window-flag-semantics/evidence.md`; (4) run
  `sila goals` and commit the updated `goals/README.md` and certification
  receipts; verify `sila goals --json` reports `unverified_count: 0`.
