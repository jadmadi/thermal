# Goal: Make no-data a successful, schema-stable state across text, JSON, and statusline

## Goal Description
Priority: P0 (activation and automation). Thermal currently treats "the user has
not run an agent yet" as an error and swaps the JSON document type under the
caller. Both were deliberate choices in
`cli-first-run-source-inventory` and
`cli-nontty-and-statusline-output-contract`; this goal supersedes those two
specific contracts with a success state that stays inside the requested schema.

Observed on `2026-10-30` with `/tmp/thermal-cli-review` (`thermal dev`):

- `HOME=<empty> thermal` -> exit `1`, stdout empty, a 28-line sources table on
  stderr, ending with three `Next:` lines.
- `HOME=<empty> thermal --json` -> exit `1`, stdout is a `type:"sources"`
  inventory. `jq '.ranking'` returns `null`.
- `HOME=<empty> thermal sources` -> exit `0`, the same inventory on stdout.
- `HOME=<empty> thermal sources --json` -> exit `0`, `type:"sources"`.
- `HOME=<empty> thermal statusline` -> exit `0`, stdout empty, stderr empty.
- `HOME=<empty> thermal statusline --json` -> exit `0`,
  `{"state":"no-data",...}`.
- With data, `thermal --json` -> exit `0`, `type:"leaderboard"`.

The outcome of this goal is one **no-data success contract**:

1. A read-only report that has no data exits `0` and returns the envelope the
   caller asked for, with `state:"no-data"` and empty collections.
2. Text mode prints a short onboarding block to stdout, not a 28-line diagnostic
   dump to stderr.
3. `thermal sources` remains the explicit inventory command, and
   `thermal sources --json` remains the only `type:"sources"` payload.
4. `statusline` stays silent in text mode for prompt safety, but documents that
   contract and exposes the state under `--json`.
5. `--require-data` restores the old exit `1` for callers that need strictness,
   without changing the payload.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: cli-first-run-source-inventory, cli-nontty-and-statusline-output-contract
- **Sequence**: 255
- **Shape**: ship
- **Tier**: roadmap
- **Supersedes in part**: the empty-state exit-`1` and bare `--json`
  `type:"sources"` rules in `cli-first-run-source-inventory`; the silent
  no-data statusline rule in `cli-nontty-and-statusline-output-contract`
  (silence is preserved, but the contract and `--json` state are extended).
- **Memory Protocol**: required
- **Wave**: 3 (after the activation and channel foundations)
Related goals:
- `cli-db-override-scope-and-errors` owns `--db` path validation; this goal
  owns the absence of data, not a bad explicit path.
- `cli-terminal-output-integrity` owns the sources table layout; this goal
  only decides when the table is printed.
- `json-output-contract-hardening` owns envelope conventions; this goal adds
  the `state:"no-data"` variant to those conventions.

## Files to Touch
- `cmd/thermal/main.go`
- `cmd/thermal/sources.go`
- `cmd/thermal/json.go`
- `cmd/thermal/help.go`
- `cmd/thermal/statusline_test.go`
- `cmd/thermal/sources_test.go`
- `cmd/thermal/json_contract_test.go`
- `cmd/thermal/simulated_user_test.go`
- `internal/render/statusline.go`
- `internal/render/sources.go`
- `scripts/simulated_user_gate.sh`
- `README.md`
- `docs/MIGRATION.md`
- `docs/JSON-CONTRACT.md`
- `project.jsonc`

## References
- **Shared Understanding & Fact Sheet**:
  [`goals/cli-empty-data-success-contract/facts.md`](facts.md)
- **Execution Plan**:
  [`goals/cli-empty-data-success-contract/plan.md`](plan.md)
- Prior contracts:
  [`goals/cli-first-run-source-inventory/goal.md`](../cli-first-run-source-inventory/goal.md),
  [`goals/cli-nontty-and-statusline-output-contract/goal.md`](../cli-nontty-and-statusline-output-contract/goal.md)
- Migration policy: `docs/MIGRATION.md`, Entry Anatomy section.

## Requirements

1. **No-data matrix.**
   Define and implement this table. "Report" means `daily`, `weekly`,
   `monthly`, `projects`, `models`, `mix`, `stats`, `trend`, and the bare
   leaderboard.

   | Surface | Text | `--json` | Exit (default) | Exit (`--require-data`) |
   | --- | --- | --- | --- | --- |
   | bare `thermal` | onboarding block, stdout | leaderboard envelope | 0 | 1 |
   | report word | onboarding block, stdout | report envelope | 0 | 1 |
   | `sources` | inventory, stdout | sources envelope | 0 | 0 |
   | `statusline` | empty | no-data object | 0 | 0 |
   | `share` | error, stderr | error object | 1 | 1 |
   | `live` non-TTY | guidance, stderr | snapshot/NDJSON | 2 | 2 |

   `--require-data` is a boolean flag accepted by every read-only report. It
   changes only the exit code and the stderr line; it never changes stdout.

2. **Text contract.**
   - On no data, print exactly this shape to **stdout**, exit `0`, stderr
     empty:
     ```
       No AI coding sessions found yet.
       Run one prompt in OpenCode, Claude, Codex, or any supported tool, then
       re-run thermal.

       Next: thermal sources   inspect all 15 source paths and states
             thermal --help    supported tools and flags
     ```
   - Do not print the 15-row inventory by default. The inventory is a
     diagnostic surface, not onboarding.
   - If exactly one source is `empty` (path exists, no sessions), say
     `OpenCode is installed but has no sessions yet.` and keep the same shape.
   - If a source is `unreadable`, add one line:
     `1 source is unreadable: run thermal sources --verbose.` Never print the
     raw `stat` error here.
   - `--verbose` appends the sources table to **stderr** after the onboarding
     block; stdout stays the three-line block.
   - `--no-color` and `NO_COLOR=1` produce zero ANSI.

3. **JSON contract.**
   - Every `--json` surface returns the envelope for the requested verb, never
     a foreign success envelope.
   - Add `"state":"no-data"` and a `"next"` object to the empty form:
     ```json
     {
       "type": "leaderboard",
       "generatedAt": "2026-10-30T00:00:00Z",
       "state": "no-data",
       "ranking": {"key": "streak", "order": "desc"},
       "results": [],
       "totals": {"tokens": 0, "cost": 0, "sessions": 0},
       "unpricedModels": [],
       "next": {
         "inventory": "thermal sources --json",
         "help": "thermal --help"
       }
     }
     ```
   - The same `state` and `next` keys appear on `daily`, `weekly`, `monthly`,
     `projects`, `models`, `mix`, `stats`, and `trend`; their existing keys
     keep their types (`data`, `rows`, `totals`, `ranking`, etc.).
   - `sources --json` keeps `type:"sources"` and never gains `state`.
   - `statusline --json` keeps `type` absent (existing contract) and returns
     `state:"no-data"`; do not invent a new envelope for it.
   - JSON on no data exits `0`; `--require-data` exits `1` but still emits the
     complete envelope.
   - `--width`, `--chart`, `--no-color`, `--verbose`, and `NO_COLOR` never
     change the JSON bytes apart from `generatedAt`.

4. **Statusline contract.**
   - Text mode with no data: stdout and stderr empty, exit `0`. This is
     required for shell prompts; do not print a placeholder by default.
   - `--verbose` writes exactly one stderr line:
     `thermal: statusline: no data (15 sources checked)`.
   - `--json` returns the existing keys with `state:"no-data"`, zero totals,
     `scope:"all tools"`, and `updatedAt` set.
   - `thermal statusline --help` states: "Prints nothing when no source has
     data; use --json for a structured no-data value." This addresses the
     "indistinguishable from broken" finding without polluting prompts.
   - `prompt` and `status` aliases behave identically.
   - A cache read error must fall through to the no-data contract, not exit
     non-zero.

5. **Share and live remain deliberate.**
   - `share` with no data keeps exit `1` and a stderr message; it is an
     explicit command and must not render an empty card.
   - `live` non-TTY guidance keeps exit `2` and stderr-only output.
   - Neither surface inherits the onboarding block.

6. **Compatibility and migration.**
   - `thermal` and `thermal --json` change from exit `1` to `0`. This is a
     breaking change for scripts that used the exit code as a data check.
   - Add a six-part MIGRATION entry:
     - Heading: `thermal --json` no longer swaps to `type:"sources"`.
     - Detection: `thermal --json | jq -r .type` returns `leaderboard`.
     - Rationale (three sentences).
     - Before/after snippets.
     - Fix: `jq -e '.ranking|length>0'`, or add `--require-data`.
     - Escape hatch: `--require-data` restores exit `1`.
   - Update `docs/JSON-CONTRACT.md` with the `state`/`next` keys and the
     "requested envelope always wins" rule.
   - Update README first-run section and `project.jsonc`.

7. **Tests.**
   - Table-driven test for the no-data matrix with separate stdout/stderr
     buffers and exit codes.
   - JSON schema test: for each verb, unmarshal and assert `type`, `state`,
     empty collection, zero totals, and `next`.
   - Statusline test: text silent, JSON no-data, verbose one line, cache error
     fallback.
   - `--require-data` test for each surface.
   - Sentinel privacy test: a prompt string in a fixture never appears in the
     onboarding block or JSON.
   - A test that `sources --json` is the only `type:"sources"` producer.
   - Width test: the onboarding block is <= 80 columns and has no ANSI.
   - Gate: `./scripts/simulated_user_gate.sh` covers empty HOME for bare,
     `--json`, each report, statusline, and `sources`.

8. **Documentation.**
   - README: "First run" shows the three-line block and states exit `0`.
   - `docs/JSON-CONTRACT.md`: table of `type`, `state`, and `next` per verb.
   - `docs/MIGRATION.md`: the breaking entry above.
   - `project.jsonc`: update the first-run feature bullet and `lastEdit`.

## Non-Goals
- Removing the `sources` command or its inventory.
- Changing loader math, pricing, or the meaning of an active day.
- Adding telemetry or an activation event.
- Making `share` succeed with no data.
- Changing `live` non-TTY behavior.
- Adding a pager or interactive onboarding.

## Acceptance Criteria
- AC-001: `HOME=$(mktemp -d) thermal` exits `0`, stdout contains
  `No AI coding sessions found yet.` and `thermal sources`, stderr is empty,
  and stdout has <= 12 lines.
- AC-002: `HOME=$(mktemp -d) thermal --json | jq -e '.type=="leaderboard" and
  .state=="no-data" and (.results|length)==0'` exits `0`.
- AC-003: For every report word, `--json` returns that verb's `type`, not
  `sources`, and exits `0`.
- AC-004: `HOME=$(mktemp -d) thermal --require-data` exits `1`, stdout still
  contains the onboarding block, and stderr has one line naming
  `--require-data`.
- AC-005: `HOME=$(mktemp -d) thermal sources` exits `0` with the inventory on
  stdout; `sources --json` is `type:"sources"`.
- AC-006: `HOME=$(mktemp -d) thermal statusline` exits `0` with stdout and
  stderr empty.
- AC-007: `HOME=$(mktemp -d) thermal statusline --json | jq -e
  '.state=="no-data" and .todayTokens==0'` exits `0`.
- AC-008: `HOME=$(mktemp -d) thermal statusline --verbose` exits `0`, stdout
  empty, stderr exactly one line.
- AC-009: `HOME=$(mktemp -d) thermal share` exits `1` with stdout empty.
- AC-010: A fixture with one empty Claude directory prints the
  "installed but has no sessions yet" line, not the generic block.
- AC-011: A sentinel prompt string in a fixture never appears in text or JSON.
- AC-012: `--width 40`, `--chart`, `--no-color`, `--verbose`, and
  `NO_COLOR=1` do not change the JSON schema or bytes (except `generatedAt`).
- AC-013: `go test -race ./cmd/thermal/... ./internal/render/...` exits `0`,
  including `TestNoDataMatrix`, `TestNoDataJSONSchema`,
  `TestStatuslineNoDataContract`, `TestRequireData`, and
  `TestSourcesOnlyEnvelope`.
- AC-014: `./scripts/simulated_user_gate.sh` exits `0` and includes the
  no-data matrix.
- AC-015: MIGRATION contains the six-part entry and a detection command.

## Negative Cases & Failure Modes
| Condition | Required result |
| --- | --- |
| Empty HOME, bare `thermal` | onboarding on stdout, exit 0 |
| Empty HOME, `thermal --json` | leaderboard envelope, `state:"no-data"`, exit 0 |
| Empty HOME, `thermal daily --json` | daily envelope, `state:"no-data"` |
| Empty HOME, `--require-data` | same payload, exit 1, one stderr line |
| Empty HOME, `sources` | inventory, exit 0 |
| Empty HOME, `sources --json` | `type:"sources"`, exit 0 |
| Empty HOME, `statusline` | silent, exit 0 |
| Empty HOME, `statusline --json` | `state:"no-data"`, exit 0 |
| Empty HOME, `share` | stderr, exit 1 |
| Empty HOME, `live` piped | stderr guidance, exit 2 |
| One empty source | source-specific wording |
| One unreadable source | one hint line, no raw error |
| Sentinel prompt | never printed |
| `--json --verbose` | valid JSON stdout, warnings stderr |
| `--json --chart` | unchanged JSON |

## Metrics & Regression Guards
- Schema stability: 100% of no-data `--json` surfaces return the requested
  `type`; a test fails on any `sources` leak.
- Exit stability: the matrix is frozen in one table test.
- Output size: onboarding block <= 12 lines and <= 80 columns.
- Privacy: sentinel absent from every text and JSON surface.
- Prompt safety: statusline no-data writes zero bytes to stdout and stderr.
- Gate: `simulated_user_gate.sh` asserts AC-001, AC-002, AC-005, AC-006, and
  AC-009.

## Risks & Mitigations
- **Risk:** Scripts relying on exit `1` silently start succeeding.
  **Mitigation:** MIGRATION entry, `--require-data`, and a changelog callout.
- **Risk:** A caller expects `type:"sources"` from `thermal --json` on empty
  data. **Mitigation:** document the switch; `sources --json` remains stable.
- **Risk:** Onboarding on stdout pollutes a pipeline.
  **Mitigation:** the caller asked for text output; `--json` and `--quiet`-style
  surfaces remain stable. Do not add progress or warnings to stdout.
- **Risk:** Two empty-state implementations drift.
  **Mitigation:** one `renderNoData(verb, inventory)` function and one
  `noDataEnvelope(verb)` helper; both tested.
- **Risk:** `statusline` silence hides a misconfiguration.
  **Mitigation:** `--verbose` and `sources`; document the contract in help.

## Memory Plan
- **Prime**: run `sila prime` at task start.
- **Recall**: run `sila recall "no supported tool data found"` and
  `sila recall "empty json sources payload"`.
- **Writes**: record a `case` for the "requested-envelope no-data contract"
  grounded to `cmd/thermal/json.go`; record a `lesson` for the schema swap.
- **Facts**: run `sila facts verify` before certification.
- **Handoff**: run `sila handoff "no-data success contract shipped"
  --open="empty-state activation metrics"`.
- **Evidence**: embed a `sila/memory-writes/v1` block under
  `## Memory Writes`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`).
2. Implementation is committed with
   `(goals/cli-empty-data-success-contract/goal.md)` under a `feat(cli):` or
   `fix(cli):` subject.
3. Goal is certified and `evidence.md` contains a standalone `Verdict: PASS`
   plus the matrix, JSON schema, and statusline evidence.
4. `sila goals --json` reports `status: "implemented"`,
   `verification_status: "verified"`, and `unverified_count: 0`, with a clean
   `git status`.
