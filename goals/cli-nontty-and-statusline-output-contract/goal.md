# Goal: Define machine-clean non-TTY stdout, exit codes, and silent statusline degradation

## Goal Description
Priority: P1 (automation and prompt integration). Two command surfaces break the Unix contract in opposite directions. `thermal live` in a pipe writes human guidance to **stdout** and exits `0`, so a pipeline treats a non-result as success. `thermal statusline` in an empty environment exits `1` and writes an error to **stderr**, so a shell prompt prints an error on every redraw.

Observed on the current source build (2026-10-29):

- `thermal live | cat` → stdout: `thermal live is interactive and needs a terminal.` plus four guidance lines; stderr empty; exit `0`.
- `thermal live --json` → valid JSON on stdout, exit `0`.
- `thermal live --json --stream` → valid NDJSON, exit `0`; `| head -1` produces no error text.
- `HOME=$(mktemp -d) thermal statusline` → stdout empty, stderr `thermal: no supported tool data found`, exit `1`.
- `HOME=$(mktemp -d) thermal statusline --json` → same, stdout empty.
- `HOME=$(mktemp -d) thermal statusline --plain` → same.
- `thermal statusline` with data → one line (`🔥 80d all tools · 134.3M tok today · top: Muse`), exit `0`; `--plain` drops the emoji; `--json` is a compact object.
- `thermal share` with no data → stderr `thermal: no supported tool data found`, exit `1` (appropriate: it is an explicit command, not a prompt helper).
- Piped static reports already disable color correctly.

The outcome of this goal is a **channel and exit-code contract**: human guidance never contaminates stdout; non-interactive `live` without a machine format fails with exit `2`; `--json`/`--stream` are always valid and pipe-safe; and `statusline`/`prompt` degrade silently with exit `0` when there is no data, while still offering a structured no-data object under `--json`.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 248
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required
- **Wave**: 2
- **Related goals**: `cli-first-run-source-inventory` owns the empty-state inventory and the bare `thermal --json` empty payload; this goal owns `live` and `statusline` behavior. `statusline-and-share-scope-contract` (implemented) owns the statusline content/scope; do not regress its labels. `json-output-contract-hardening` (implemented) owns envelope shapes; this goal adds only the statusline `state` field.

## Files to Touch
- `cmd/thermal/main.go`
- `cmd/thermal/stream_test.go`
- `cmd/thermal/live_test.go`
- `cmd/thermal/statusline_test.go`
- `internal/render/statusline.go`
- `scripts/simulated_user_gate.sh`
- `README.md`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- **Shared Understanding & Fact Sheet**: [`goals/cli-nontty-and-statusline-output-contract/facts.md`](facts.md)
- **Execution Plan**: [`goals/cli-nontty-and-statusline-output-contract/plan.md`](plan.md)

## Requirements

1. **`live` non-TTY contract.**
   - If stdout is not a TTY and neither `--json` nor `--stream` is set, write the existing guidance to **stderr** and exit `2`:
     ```
     thermal: live needs a terminal
       thermal live --json            instant JSON snapshot
       thermal live --json --stream   stream NDJSON live events
       thermal stats                  daily distribution and token composition
     ```
   - stdout must be empty in this case.
   - `--json` (snapshot) and `--json --stream` (NDJSON) keep working in pipes and exit `0`; `--stream` implies NDJSON even if `--json` is omitted, or is rejected with a clear message. Choose one and test both spellings.
   - Human progress, retries, warnings, and loader diagnostics go to stderr; stdout carries only the requested format.
   - A broken pipe (`| head -1`) must terminate cleanly: no `broken pipe` text on stderr, exit `0` or `141` consistently, and no panic.
   - In a TTY, the existing TUI behavior is unchanged.
   - `--no-color` and `NO_COLOR` do not affect JSON/NDJSON.

2. **`statusline`/`prompt` no-data contract.**
   - When no source has data, exit `0`, write nothing to stdout, and write nothing to stderr by default.
   - `--verbose` writes one stderr line: `thermal: statusline: no data (15 sources checked)`; it never writes to stdout.
   - `--json` emits one valid object on stdout and exits `0`:
     ```json
     {"state":"no-data","streak":0,"scope":"all tools","longestStreak":0,"todayTokens":0,"todayTurns":0,"activeTool":"","day":"2026-10-29","updatedAt":1793300000}
     ```
   - With data, `--json` keeps the existing keys plus `"state":"ok"`; text and `--plain` keep their current one-line shape.
   - A missing or unreadable cache must not fail the prompt: fall back to the normal load, and if that fails, degrade as no-data with exit `0`.
   - Do not write a no-data cache that would suppress recovery for more than 60 seconds; a fresh session must appear promptly. Recommend a 30–60 s no-data marker or no cache at all.
   - `thermal statusline` must not trigger the update check (`NoUpdateCheck` already set) or any network call.
   - `thermal prompt` and `thermal status` remain aliases with identical behavior.

3. **`share` and other explicit commands.**
   - `share` with no data remains a deliberate failure: stderr message, exit `1`, stdout empty. It is not a prompt helper and must not silently render an empty card.
   - `thermal --json` with no data is owned by the first-run goal; this goal only ensures `statusline --json` and `live --json` never inherit the generic stderr path.
   - `audit` is unaffected; it does not require agent data.

4. **Channel purity.**
   - Add a table-driven test that runs each surface with stdout and stderr captured separately:
     | Surface | TTY? | Data? | Format | stdout | stderr | Exit |
     | --- | --- | --- | --- | --- | --- | --- |
     | `live` | no | any | text | empty | guidance | 2 |
     | `live` | no | any | `--json` | JSON | empty | 0/1 |
     | `live` | no | any | `--json --stream` | NDJSON | empty | 0 |
     | `statusline` | no | no | text | empty | empty | 0 |
     | `statusline` | no | no | `--json` | JSON | empty | 0 |
     | `statusline` | no | yes | text | one line | empty | 0 |
     | `share` | no | no | text | empty | message | 1 |
     | `daily` | no | yes | text | table | empty | 0 |
     | any | no | any | `--json` error | empty | message | 1/2 |
   - No update-check or pricing notice may appear on stdout.
   - No ANSI on any non-TTY stream.
   - JSON/NDJSON must remain valid when stderr is also written.

5. **Exit-code contract.**
   - `live` non-TTY text: `2` (usage).
   - `live --json` success: `0`; data/load failure: `1`.
   - `live --json --stream` success: `0`; clean broken pipe: `0`.
   - `statusline`/`prompt` no data: `0`; success: `0`; flag error: `2`.
   - `share` no data: `1`.
   - Document the `live` change (`0` → `2`) in `docs/MIGRATION.md` with detection and fix.

6. **Tests.**
   - Pipe harness using `exec.Command` with `StdoutPipe`/`StderrPipe`; assert exact channels and exit codes.
   - A PTY harness for the TTY `live` case to prove it still launches.
   - `live --json --stream | head -1`: assert one valid JSON line and no broken-pipe text.
   - Empty HOME and one-tool fixture for statusline.
   - Cache tests: missing cache, stale cache, unreadable cache, no-data marker expiry.
   - A latency test for the no-data statusline (<10 ms warm).
   - A fuzz/property test that every captured stdout for `--json` parses as JSON and every NDJSON line parses independently.

7. **Docs and migration.**
   - README prompt integration section with `$(thermal statusline)` and a note that it prints nothing when there is no data.
   - MIGRATION six-part entry for `live` non-TTY exit `2` and the statusline no-data change; detection (`thermal live </dev/null; echo $?`), fix (`--json`), escape hatch (`--json --stream`).
   - `project.jsonc` updated.
   - `scripts/simulated_user_gate.sh`: pipe checks for `live`, `statusline`, and `share`.

## Non-Goals
- Changing the live TUI layout, metrics, or polling behavior.
- Changing the statusline content/scope labels (owned by `statusline-and-share-scope-contract`).
- Adding a JSON error envelope for every command (owned by `json-output-contract-hardening` if desired).
- Changing `share` URL encoding or privacy guarantees.
- Adding a daemon or background process.

## Acceptance Criteria
- AC-001: `thermal live | cat` exits `2`, stdout is empty, and stderr contains `live needs a terminal`, `--json`, and `--json --stream`.
- AC-002: `thermal live --json | jq -e .` exits `0` and stdout is valid JSON.
- AC-003: `thermal live --json --stream | head -1 | jq -e .` exits `0`; stderr is empty; no `broken pipe`.
- AC-004: `HOME=$(mktemp -d) thermal statusline` exits `0`, stdout and stderr are both empty.
- AC-005: `HOME=$(mktemp -d) thermal statusline --json | jq -e '.state=="no-data" and .todayTokens==0'` exits `0`.
- AC-006: `HOME=$(mktemp -d) thermal statusline --plain` exits `0` with empty output.
- AC-007: `thermal statusline` with data exits `0` and prints exactly one line; `--json` includes `"state":"ok"`.
- AC-008: `HOME=$(mktemp -d) thermal share` exits `1`, stdout empty, stderr names the missing data.
- AC-009: The channel matrix in requirement 4 is implemented as a table test and passes for every row.
- AC-010: `thermal live | grep -c $'\x1b'` returns `0`; `NO_COLOR=1 thermal statusline` emits no ANSI.
- AC-011: `go test -race ./cmd/thermal/... ./internal/render/...` exits `0`, including `TestLiveNonTTYContract`, `TestStatuslineNoData`, `TestStatuslineJSONState`, `TestChannelMatrix`, and `TestStreamBrokenPipe`.
- AC-012: `./scripts/simulated_user_gate.sh` exits `0` and includes the channel/exit-code section.

## Negative Cases & Failure Modes
| Condition | Required result |
| --- | --- |
| `live` piped, no format | exit 2, stdout empty, guidance on stderr |
| `live --json` piped | exit 0, JSON stdout |
| `live --json --stream` piped | exit 0, NDJSON stdout |
| `live --json --stream \| head -1` | clean exit, no broken-pipe text |
| `statusline` empty HOME | exit 0, silent |
| `statusline --json` empty HOME | exit 0, `state:"no-data"` |
| `statusline` unreadable cache | exit 0, fall back or silent |
| `statusline` flag error | exit 2, stderr |
| `share` empty HOME | exit 1, stderr |
| `daily --json` error | exit 1, stderr, empty stdout |
| TTY `live` | TUI unchanged |
| `NO_COLOR` non-TTY | no ANSI |

## Metrics & Regression Guards
- Channel matrix: 100% of rows asserted; any stdout contamination fails.
- Exit-code table frozen in tests and mirrored in help/MIGRATION.
- Statusline no-data latency <10 ms warm and <100 ms cold; record in evidence.
- Stream line validity: every line parses with `jq -e`; a 100-line sample.
- Gate: simulated user gate asserts AC-001, AC-002, AC-003, AC-004, AC-005, and AC-008.

## Risks & Mitigations
- **Risk:** Changing `live` from exit `0` to `2` breaks pipelines that ignored the output. **Mitigation:** MIGRATION entry; those pipelines were already receiving non-data.
- **Risk:** A silent statusline hides a real configuration problem. **Mitigation:** `--verbose` and `thermal sources` provide the diagnosis; the prompt stays clean.
- **Risk:** `loadUsage` currently calls `os.Exit`, so statusline cannot degrade without a refactor. **Mitigation:** change `loadUsage` to return `(usageSet, error)` and let callers choose exit behavior; coordinate with the first-run goal.
- **Risk:** A no-data cache could mask a newly created session. **Mitigation:** short TTL (≤60 s) or no cache; test recovery.
- **Risk:** SIGPIPE handling differs across platforms. **Mitigation:** treat EPIPE as clean exit `0` on Unix; add a Windows-specific test/skip.

## Memory Plan
- **Prime**: run `sila prime` at task start.
- **Recall**: run `sila recall "non-tty stdout guidance exit zero"` and `sila recall "statusline prompt error"`.
- **Writes**: record a `case` for the "channel matrix + quiet prompt helper" contract, grounded to `cmd/thermal/main.go` and `internal/render/statusline.go`; record a `lesson` for the stdout-guidance contamination.
- **Facts**: run `sila facts verify` before certification.
- **Handoff**: run `sila handoff "non-TTY and statusline contract shipped" --open="JSON error envelope"`.
- **Evidence**: embed a `sila/memory-writes/v1` block under `## Memory Writes`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`).
2. Implementation is committed with `(goals/cli-nontty-and-statusline-output-contract/goal.md)` under a `fix(cli):` subject.
3. Goal is certified and `evidence.md` contains a standalone `Verdict: PASS` plus the channel matrix results.
4. `sila goals --json` reports `status: "implemented"`, `verification_status: "verified"`, and `unverified_count: 0`, with a clean `git status`.
