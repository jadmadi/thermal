# Plan: Make no-data a successful, schema-stable state across text, JSON, and statusline

## Execution Steps
- [ ] **Phase 1 — Contract types and flags**: Add `RequireData` to
  `thermal.Options` and `parseArgsFrom`, define `NoDataNext`, and add the
  `state`/`next` fields to the report envelopes without changing payloads that
  have data. Acceptance: `TestRequireDataFlag`, `TestNoDataEnvelopeKeys`, and
  all existing JSON contract tests pass.
  - `--require-data` accepted by every read-only report; rejected by `live`,
    `statusline`, and `share` if it has no meaning there (or accepted as a
    no-op with a documented note; choose one and test).
  - `state` and `next` are `omitempty`; populated payloads are byte-identical
    to today.
- [ ] **Phase 2 — Text onboarding**: Replace `printEmptyInventory` with
  `printNoData`, print the three-line block to stdout, keep the inventory for
  `sources` and `--verbose`, and honor the one-empty-source and unreadable
  variants. Acceptance: `TestNoDataText`, `TestNoDataOneEmptySource`, and
  `TestNoDataUnreadableSource` pass.
  - Exit `0` by default, `1` with `--require-data`.
  - stderr empty unless `--verbose` or `--require-data`.
  - No ANSI under `NO_COLOR`.
- [ ] **Phase 3 — JSON and statusline**: Make every report return its own
  no-data envelope; keep `sources --json` unchanged; add `state:"no-data"` to
  statusline JSON; document statusline silence in help. Acceptance:
  `TestNoDataJSONSchema` (table over every verb),
  `TestSourcesOnlyEnvelope`, `TestStatuslineNoDataContract`, and
  `TestStatuslineCacheFallback` pass.
  - `jq` schema checks for each verb.
  - `--json --verbose` stays valid; warnings on stderr only.
- [ ] **Phase 4 — Migration, docs, gate**: Add the six-part MIGRATION entry,
  update `docs/JSON-CONTRACT.md`, README first-run, `project.jsonc`, and the
  simulated-user gate. Acceptance: MIGRATION lint passes; gate exits `0`.
- [ ] **Phase 5 — Evidence and closure**: Record the no-data matrix, JSON
  samples, statusline channels, and gate output in `evidence.md`; commit with
  goal attribution; certify and reconcile.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | empty HOME `thermal` | exit 0, onboarding stdout, empty stderr |
| AC-002 | empty HOME `thermal --json` | leaderboard, `state:"no-data"`, exit 0 |
| AC-003 | each report `--json` | verb `type`, not `sources`, exit 0 |
| AC-004 | empty HOME `--require-data` | exit 1, same stdout, stderr hint |
| AC-005 | empty HOME `sources` / `--json` | exit 0, table / `type:"sources"` |
| AC-006 | empty HOME `statusline` | exit 0, both channels empty |
| AC-007 | empty HOME `statusline --json` | `state:"no-data"` |
| AC-008 | empty HOME `statusline --verbose` | one stderr line, stdout empty |
| AC-009 | empty HOME `share` | exit 1, stdout empty |
| AC-010 | one empty Claude dir | "installed but has no sessions yet" |
| AC-011 | sentinel prompt | absent from all output |
| AC-012 | `--json` env/width variants | schema and bytes stable |
| AC-013 | `go test -race ...` | exit 0 |
| AC-014 | simulated user gate | exit 0 |
| AC-015 | MIGRATION entry | six parts, detection command |

## Rollback
- Keep `--require-data` as the compatibility switch; never restore the
  `type:"sources"` swap for `thermal --json`.
- If the onboarding block is too noisy for an internal caller, add `--quiet`
  (empty stdout, exit 0) rather than returning to stderr.
- If a report envelope cannot carry `state` without a breaking schema change,
  pin the new keys behind a documented minor version and update MIGRATION.
- Do not remove `sources`; it is the diagnostic inventory, not the onboarding
  payload.

## Phase 6 — Attribution & Registry Closure
- [ ] **Phase 6 — Attribution & Registry Closure**: (1) check every task
  above; (2) commit with
  `feat(cli): make no-data a success state with stable JSON
  (goals/cli-empty-data-success-contract/goal.md)`; (3) certify via
  `sila goals certify cli-empty-data-success-contract
  --evidence=goals/cli-empty-data-success-contract/evidence.md`;
  (4) run `sila goals` and commit the updated `goals/README.md` and
  certification receipts; verify `sila goals --json` reports
  `unverified_count: 0`.
