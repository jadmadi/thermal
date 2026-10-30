# Plan: Make version output and deprecation warnings factually true

## Execution Steps
- [ ] **Phase 1 — Version Info and payload**: Add `Info`, `Current`, and
  `IsPast`; always print `commit`/`built`; add `stamped`/`semver` to JSON.
  Acceptance: `TestVersionPayloadMatrix` and `TestStampedBuild` pass.
- [ ] **Phase 2 — Help truth and verbose signal**: Rewrite the version help
  summary/description and call `announceVerbose` once when `-v` is active.
  Acceptance: `TestVersionHelpTruth`,
  `TestVerboseAnnouncement`, and `TestVerboseJSONStdout` pass.
- [ ] **Phase 3 — Deprecation table and guard**: Replace the inline
  `0.14.0` strings with one table, absolute URLs, and a current-version guard;
  fall back to the no-version form at runtime. Acceptance:
  `TestDeprecationTruth`, `TestNoExpiredWarningVersion`, and
  `TestDeprecationRuntimeFallback` pass.
- [ ] **Phase 4 — Docs, gate, evidence, closure**: README version section,
  MIGRATION entries, `docs/PROMISES.md`, `project.jsonc`, gate; record
  evidence; commit, certify, reconcile.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | `version` | fallback + unknown fields |
| AC-002 | `version --json` | `stamped:false`, unknowns |
| AC-003 | stamped test build | real version/commit/date |
| AC-004 | `version --help` | fallback sentence, no promise |
| AC-005 | `-v` | one stderr announcement |
| AC-006 | `-v --json` | valid stdout, stderr announcement |
| AC-007 | `--license` | no v0.14.0, no repo path, URL |
| AC-008 | `nous` | no expired version, `hermes` |
| AC-009 | deprecation table test | rejects past versions |
| AC-010 | source grep | no v0.14.0 deprecation |
| AC-011 | `go test -race ...` | exit 0 |
| AC-012 | simulated user gate | exit 0 |

## Rollback
- Keep the `stamped` field additive; if a consumer rejects it, document the
  minor schema version rather than reverting the truth.
- If the `-v` announcement is too noisy, emit it only when a warning is also
  printed; do not return to zero signal.
- Keep the deprecation table even if the URL changes; update one constant.
- Do not reintroduce a removal version without a current-version test.

## Phase 5 — Attribution & Registry Closure
- [ ] **Phase 5 — Attribution & Registry Closure**: (1) check every task
  above; (2) commit with
  `fix(cli): make version and deprecation warnings truthful
  (goals/cli-version-and-deprecation-truth/goal.md)`; (3) certify via
  `sila goals certify cli-version-and-deprecation-truth
  --evidence=goals/cli-version-and-deprecation-truth/evidence.md`; (4) run
  `sila goals` and commit the updated `goals/README.md` and certification
  receipts; verify `sila goals --json` reports `unverified_count: 0`.
