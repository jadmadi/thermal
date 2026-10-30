# Goal: Document exit codes and give upgrade global-flag parity

## Goal Description
Priority: P1 (scripting). A wrapper that prepends `--offline` — the documented
"never touch the network" flag accepted by every report — breaks on the one
subcommand that actually reaches the network:
`thermal upgrade --check --offline` exits `2` with
`unknown flag for upgrade: --offline`. Separately, the exit-code scheme is
clean and consistent (`0` success, `1` data/IO, `2` usage, `3` refused) but only
`3` is documented, in `upgrade --help`; a script author must reverse-engineer
the rest from failures.

Observed on `2026-10-30` with `/tmp/thermal-cli-review` (`thermal dev`):

- `thermal upgrade --check --offline` -> exit `2`, stdout empty:
  `thermal: unknown flag for upgrade: --offline`
  `  run thermal upgrade --help`
  `  run thermal --help for the flag list`
- `thermal upgrade --check --json` -> exit `0` with a release object.
- `thermal upgrade --json` (non-interactive) -> exit `3`,
  `{"action":"declined","error":"confirmation required for --json; ..."}`.
- `thermal version` -> exit `0`; `thermal opencode` -> exit `1`;
  `thermal --tool open` -> exit `2`.
- `thermal --help` has no "Exit codes" block; `upgrade --help` mentions only
  the non-interactive `3`.

The outcome of this goal is an **automation contract**:

1. A known global flag is never reported as "unknown" on any subcommand. It is
   either honored or explicitly not applicable, with a working alternative.
2. `upgrade --check` and `--dry-run` accept `--offline` and
   `--no-update-check`; `--offline` fails closed with a clear message when no
   cached release index exists.
3. The `0/1/2/3` exit-code scheme is documented in global help and every
   command page that can return `1` or `2`.
4. The contract is frozen in a table test and mirrored in README/MIGRATION.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: cli-upgrade-nondestructive-safety-flags, cli-nontty-and-statusline-output-contract
- **Sequence**: 262
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required
- **Wave**: 3
Related goals:
- `cli-upgrade-file-integrity` owns the atomic replacement; this goal owns flag
  parsing and exit codes.
- `cli-daily-update-check` owns the background check; this goal owns the
  explicit `upgrade` command.
- `cli-empty-data-success-contract` owns the empty-data exit code; this goal
  documents the scheme, not the individual surfaces.

## Files to Touch
- `cmd/thermal/upgrade_flow.go`
- `cmd/thermal/upgrade_test.go`
- `cmd/thermal/args.go`
- `cmd/thermal/args_strict_test.go`
- `cmd/thermal/help.go`
- `cmd/thermal/main.go`
- `cmd/thermal/simulated_user_test.go`
- `scripts/simulated_user_gate.sh`
- `README.md`
- `docs/MIGRATION.md`
- `docs/CLI-CONTRACT.md`
- `project.jsonc`

## References
- **Shared Understanding & Fact Sheet**:
  [`goals/cli-automation-exit-and-upgrade-parity/facts.md`](facts.md)
- **Execution Plan**:
  [`goals/cli-automation-exit-and-upgrade-parity/plan.md`](plan.md)
- Prior contracts:
  [`goals/cli-upgrade-nondestructive-safety-flags/goal.md`](../cli-upgrade-nondestructive-safety-flags/goal.md),
  [`goals/cli-nontty-and-statusline-output-contract/goal.md`](../cli-nontty-and-statusline-output-contract/goal.md)

## Requirements

1. **Global flag applicability.**
   - Define one table for every command group:
     ```go
     type FlagDisposition string
     const (
         FlagHonored   FlagDisposition = "honored"
         FlagNoOp      FlagDisposition = "no-op"
         FlagRejected  FlagDisposition = "not applicable"
     )

     type FlagRule struct {
         Flag        string
         Disposition FlagDisposition
         Note        string
         Alternative string
     }
     ```
   - Known global flags: `--json`, `--no-color`, `--verbose`, `--offline`,
     `--no-update-check`, `--width`, `--fresh`.
   - `upgrade`:
     - `--json`, `--no-color`, `--verbose` honored;
     - `--offline` honored for `--check`/`--dry-run`: no network; use a cached
       release index, else exit `1` with
       `thermal: offline: no cached release index; run thermal upgrade --check`;
     - `--no-update-check` no-op (the command is explicit);
     - `--width` not applicable, but the message must say so and offer
       `--json`;
     - `--fresh` not applicable.
   - A rejected global flag exits `2` with:
     ```
     thermal: --width is not applicable to upgrade
       upgrade output is fixed-width; use --json for structured output
     ```
     Never `unknown flag for upgrade`.
   - An actually unknown flag (`--bogus`) keeps `unknown flag`.

2. **Upgrade parsing.**
   - Parse flags in any position, before/after `upgrade`, with `=` or a space.
   - `--offline` and `--no-update-check` are accepted on `upgrade`,
     `upgrade --check`, and `upgrade --dry-run`.
   - `--offline` must not silently fall through to the network. If a cached
     release index exists, use it and mark the JSON `"source":"cache"`;
     otherwise exit `1` with the message above.
   - Mutating upgrade still requires `--yes` or a TTY; `--offline` does not
     change that.
   - `upgrade --json` emits exactly one object on stdout, including on error;
     the human fix goes to stderr.
   - `upgrade --help` lists every accepted flag and the exit codes.

3. **Exit-code contract.**
   | Code | Meaning | Typical surfaces |
   | --- | --- | --- |
   | `0` | success | any completed read/write/check |
   | `1` | data, IO, or network failure | missing data, unreadable DB, failed download |
   | `2` | usage error | unknown flag, invalid value, wrong slot, width refusal |
   | `3` | refused in non-interactive mode | `upgrade` without `--yes`/TTY |
   - Define constants in one place:
     ```go
     const (
         ExitOK       = 0
         ExitFailure  = 1
         ExitUsage    = 2
         ExitRefused  = 3
     )
     ```
   - Replace magic numbers at `os.Exit` sites where practical; keep the
     behavior identical.
   - Add a four-line block to global help:
     ```
     Exit codes: 0 ok · 1 data/IO error · 2 usage error · 3 refused
     ```
   - Add the same block to `upgrade --help` and to every command page that can
     return `1` or `2` (all reports, `live`, `statusline`, `share`, `sources`).
   - `docs/CLI-CONTRACT.md` (new or existing) is the canonical table with one
     example command per code.
   - `--json` error objects keep their existing keys; do not add `exitCode`
     unless the JSON contract already has a place for it. Recommendation: do
     not add it; the process exit code is the contract.

4. **Channel and non-interactive rules.**
   - `upgrade` non-interactive without `--yes`: exit `3`, JSON object on
     stdout when `--json`, human message on stderr otherwise.
   - `upgrade --check --json --offline` with no cache: exit `1`, one JSON
     object with `"error"` and `"source":"none"`; stderr empty.
   - `upgrade --check --offline` without JSON: exit `1`, one stderr line, no
     network attempt. A test uses a dead API URL to prove no request.
   - `upgrade --json` never prints the human fix on stdout.
   - `--no-color`/`NO_COLOR` keep the warning plain.

5. **Tests.**
   - Flag disposition table for every command group and every global flag.
   - `upgrade` matrix: `--offline`, `--no-update-check`, `--width`,
     `--fresh`, `--bogus`, `--check`, `--dry-run`, `--yes`, `--json`, in
     before/after positions.
   - Dead-API test: `--offline --check` makes zero HTTP requests.
   - Exit-code table: one command per code, asserted.
   - Help test: the block appears in global and upgrade help.
   - JSON test: one object on stdout for success and error; stderr carries the
     human message.
   - Gate: `simulated_user_gate.sh` covers `upgrade --check --offline`,
     `upgrade --json`, and the exit-code block.

6. **Documentation.**
   - README "Exit codes" section with the table.
   - `docs/CLI-CONTRACT.md` canonical table and examples.
   - MIGRATION entry for accepting `--offline` (additive) and for the explicit
     "not applicable" message replacing "unknown flag" (text change).
   - `project.jsonc` updated.

## Non-Goals
- Adding a new global flag.
- Changing the mutating upgrade safety rules.
- Adding an exit code for "no update available" (it stays `0`).
- Changing the JSON upgrade schema beyond `"source"` if a cache is used.
- Making `--width` work on upgrade.

## Acceptance Criteria
- AC-001: `thermal upgrade --check --offline` exits `0` when a cached release
  index exists, or `1` with `offline: no cached release index`; it never
  prints `unknown flag`.
- AC-002: A dead-API test proves `--offline --check` makes zero HTTP requests.
- AC-003: `thermal upgrade --no-update-check --check` exits `0` as a no-op.
- AC-004: `thermal upgrade --width 100` exits `2` with `not applicable` and
  `--json` as the alternative, not `unknown flag`.
- AC-005: `thermal upgrade --bogus` still exits `2` with `unknown flag`.
- AC-006: `thermal upgrade --json` non-interactive exits `3` with one JSON
  object on stdout.
- AC-007: `thermal --help` contains the four-line exit-code block.
- AC-008: `thermal upgrade --help` contains the exit-code block and the
  `--offline` semantics.
- AC-009: `thermal version` = 0, `thermal opencode` = 1,
  `thermal --tool open` = 2, `thermal upgrade --json` = 3.
- AC-010: `docs/CLI-CONTRACT.md` has one example per code and matches the
  tests.
- AC-011: `go test -race ./cmd/thermal/...` exits `0`, including
  `TestUpgradeGlobalFlags`, `TestOfflineNoNetwork`, `TestExitCodeMatrix`, and
  `TestExitCodeHelp`.
- AC-012: `./scripts/simulated_user_gate.sh` exits `0` and includes the
  upgrade/exit-code cases.

## Negative Cases & Failure Modes
| Invocation | Required result |
| --- | --- |
| `upgrade --check --offline`, cache | exit 0, `source:"cache"` |
| `upgrade --check --offline`, no cache | exit 1, offline message |
| `upgrade --check --offline` dead API | zero HTTP requests |
| `upgrade --no-update-check` | no-op, exit 0 |
| `upgrade --width 100` | not applicable + `--json` |
| `upgrade --fresh` | not applicable + `--json` |
| `upgrade --bogus` | unknown flag, exit 2 |
| `upgrade --json` no TTY | exit 3, one object |
| `version` | exit 0 |
| missing data | exit 1 |
| bad flag/value | exit 2 |
| `upgrade --yes --offline` | mutating path unchanged |

## Metrics & Regression Guards
- Flag parity: every known global flag has a disposition for every command
  group; a test enumerates the table.
- No `unknown flag` for a known global; a grep test over captured stderr.
- Exit codes: the table test asserts one command per code.
- Offline: a dead-API test asserts zero requests.
- Gate: simulated user gate asserts AC-001, AC-004, and AC-006.

## Risks & Mitigations
- **Risk:** `--offline` on upgrade silently uses stale release data.
  **Mitigation:** mark `"source":"cache"` and print the cache age in text.
- **Risk:** A cached index is absent on first run.
  **Mitigation:** fail closed with a clear `--check` instruction; do not
  contact the network.
- **Risk:** Adding exit-code text bloats help.
  **Mitigation:** four lines, one block, reused on every page.
- **Risk:** Magic-number replacement changes a subtle path.
  **Mitigation:** constants equal the current values; tests assert the matrix.
- **Risk:** `--width` rejection breaks a wrapper that passes it blindly.
  **Mitigation:** explicit `not applicable` + `--json`; MIGRATION note.

## Memory Plan
- **Prime**: run `sila prime` at task start.
- **Recall**: run `sila recall "unknown flag for upgrade offline"` and
  `sila recall "exit code 3"`.
- **Writes**: record a `case` for the "global flag disposition matrix"
  mechanism; record a `lesson` for a known flag reported as unknown.
- **Facts**: run `sila facts verify` before certification.
- **Handoff**: run `sila handoff "automation exit and upgrade parity shipped"
  --open="cache index format"`.
- **Evidence**: embed a `sila/memory-writes/v1` block under
  `## Memory Writes`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`).
2. Implementation is committed with
   `(goals/cli-automation-exit-and-upgrade-parity/goal.md)` under a
   `fix(cli):` or `feat(cli):` subject.
3. Goal is certified and `evidence.md` contains a standalone `Verdict: PASS`
   plus the flag matrix, offline/no-network, and exit-code evidence.
4. `sila goals --json` reports `status: "implemented"`,
   `verification_status: "verified"`, and `unverified_count: 0`, with a clean
   `git status`.
