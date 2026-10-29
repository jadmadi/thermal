# Plan: Give -v, -V, --version, and --verbose one unambiguous contract

## Execution Steps
- [ ] **Phase 1 — Remove the overloaded special case**: Delete the `len(os.Args)==2` version branch; add `versionRequested` for `-V`, `--version`, and `version`; ensure `-v` always maps to verbose. Acceptance: `TestVerboseNeverPrintsVersion` and `TestVersionFlagMatrix` pass.
  - `-V` is a known flag in every position.
  - `-v` is not consumed as a value.
  - `-V`/`--version`/`version` produce identical output.
- [ ] **Phase 2 — Precedence resolver**: Integrate with the help resolver: help > version > validation > execution. Document `-V daily`, `daily -V`, `-v -V`, and `-V --bogus`. Acceptance: `TestVersionPrecedence` passes for all rows in `facts.md`.
  - `-h -V` and `-V -h` print help.
  - `-V daily` prints version and does not execute daily.
  - `-v --json` remains command JSON.
- [ ] **Phase 3 — JSON and help**: Make `-V --json`, `--version --json`, and `version --json` identical (deterministic or clock-injected); add both flag lines and the precedence note to help. Acceptance: `TestVersionJSONParity`, `TestHelpVersionFlags`, and the no-side-effect test pass.
  - Version JSON has `type`, `version`, `commit`, `date`.
  - No data/network/cache access.
  - Help does not map `-v` to version.
- [ ] **Phase 4 — Gate, docs, and closure**: Add the simulated-user-gate version section, README flag table, MIGRATION for the `-v` change, project.jsonc, and close the goal.
  - Record the invocation matrix in `evidence.md`.
  - `CGO_ENABLED=0 go test -race ./...` and `./scripts/simulated_user_gate.sh` pass.
  - Commit with `fix(cli): disambiguate -v, -V, --version, and --verbose (goals/cli-version-verbose-flag-contract/goal.md)`.
  - Certify with `sila goals certify cli-version-verbose-flag-contract --evidence=goals/cli-version-verbose-flag-contract/evidence.md`.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | `-v` | leaderboard, no version |
| AC-002 | `-v daily`, `daily -v` | daily report |
| AC-003 | `-V` | version, exit 0 |
| AC-004 | `--version`, `version` | identical to `-V` |
| AC-005 | `-V --json` | version JSON |
| AC-006 | `-v --json` | command JSON |
| AC-007 | help | both flag lines |
| AC-008 | `-v -V` | version |
| AC-009 | `-V daily` | version, no table |
| AC-010 | `-V --bogus` | version, exit 0 |
| AC-011 | empty HOME `-v` | empty-state error, exit 1 |
| AC-012 | `go test -race ...` | exit 0 |
| AC-013 | gate | exit 0 |

## Rollback
- If removing the `-v` version path breaks an internal script, update the script to `-V`; do not restore context-dependent semantics.
- Keep `thermal version` as the long-form command even if a flag is renamed.
- The version JSON schema is additive/deterministic; keep it stable.

## Phase 5 — Attribution & Registry Closure
- [ ] **Phase 5 — Attribution & Registry Closure**: (1) check all completed tasks in `goals/cli-version-verbose-flag-contract/plan.md`; (2) commit with `(goals/cli-version-verbose-flag-contract/goal.md)` using a `fix(cli):` subject; (3) certify via `sila goals certify cli-version-verbose-flag-contract --evidence=goals/cli-version-verbose-flag-contract/evidence.md` (with `Verdict: PASS`); (4) reconcile via `sila goals` and commit the updated `goals/README.md` and certification receipts; verify `sila goals --json` reports `unverified_count: 0` for this goal.
