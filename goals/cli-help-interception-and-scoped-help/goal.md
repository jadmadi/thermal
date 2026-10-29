# Goal: Never swallow -h/--help and add command-scoped help

## Goal Description
Priority: P0 (discoverability, safety-adjacent). Help is the first thing a new user types and the only safe way to explore a destructive command. Today help is honoured only when it is the sole argument: `parseArgs()` special-cases `-h`, `--help`, and `help` inside `if len(os.Args) == 2` (`cmd/thermal/main.go:47-63`). Once a positional is present, Go's `flag` package stops parsing at that positional, the hand-rolled post-positional scanner has no `--help` case, and the command executes.

Observed on the current source build (2026-10-29):

- `thermal daily --help` renders the full daily table and exits `0`.
- `thermal daily -h` does the same.
- `thermal live --help` in a 100x30 PTY launches the interactive TUI; the user must press `q` to escape.
- `thermal upgrade --help` on a disposable binary copy performs a real self-replacement (md5 `43110b4…` → `ced2913…`) because every unknown flag is ignored.
- `thermal help daily` prints `thermal: unknown tool: help` and exits `1`.
- Only `thermal --help` works, and it is one 84-line global page with no per-command scope.

The outcome of this goal is a single help contract: **any help request is intercepted before dispatch, prints a side-effect-free page scoped to the resolved command context, writes only to stdout, exits 0 when the context is valid, and never reads data, opens a TUI, calls the network, writes a cache, or replaces the binary.**

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 242
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required
- **Wave**: 1 (safe to run in parallel with `cli-upgrade-nondestructive-safety-flags`, `cli-argument-parser-strictness`, `cli-db-override-scope-and-errors`)
- **Related implemented goals (do not regress)**: `help-supported-tools-and-flags` owns help *content accuracy* (tool registry, `--sort` values, placeholders); this goal owns help *precedence, scope, and side-effect freedom*. `cli-staged-deprecation-and-upgrade` owns upgrade notifications.

## Files to Touch
- `cmd/thermal/main.go`
- `cmd/thermal/help.go`
- `cmd/thermal/help_test.go`
- `internal/render/help.go`
- `internal/render/help_test.go`
- `cmd/thermal/upgrade.go`
- `scripts/simulated_user_gate.sh`
- `README.md`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- **Shared Understanding & Fact Sheet**: [`goals/cli-help-interception-and-scoped-help/facts.md`](facts.md)
- **Execution Plan**: [`goals/cli-help-interception-and-scoped-help/plan.md`](plan.md)
- **Evidence standard**: every AC below must appear verbatim in `evidence.md` with its observed exit code and output excerpt.

## Requirements

1. **Global interception with explicit precedence.**
   - Scan raw `os.Args[1:]` before `flag.Parse()` and before any dispatch.
   - Precedence, highest first: `help` > `version` > argument validation > command execution.
   - A help request is any of: `-h`, `--help`, or `help` as the first non-flag token, in any position relative to flags and positionals.
   - `thermal --help`, `thermal -h`, and `thermal help` print global help and exit `0`.
   - `thermal <verb> --help`, `thermal <verb> -h`, and `thermal help <verb>` print the same scoped page for `<verb>` and exit `0`. The two spellings must be byte-identical.
   - `thermal <tool> --help` prints tool-scoped help (heatmap, accepted report words, `--weeks`, `--db`).
   - `thermal <tool> <verb> --help` and `thermal <verb> <tool> --help` print the same tool+verb page.
   - `thermal help help` prints a short page describing global, scoped, and `help <verb>` forms.
   - `thermal help <unknown>` and `thermal <unknown> --help` exit `1`, print `thermal: unknown command "<token>"`, a nearest-match suggestion, and `run thermal --help for the command list`; they must not print a misleading scoped page.
   - `-h`/`--help` must never be treated as an unknown flag, a tool name, a report word, or a value for another flag.

2. **Command registry as the single source of truth.**
   - Add a `commandSpec` table with, per verb: canonical name, aliases, one-line summary, usage synopsis, accepted flags with value type/default/valid values, examples (2–3), JSON top-level keys, exit codes, related commands, and whether it accepts a tool positional.
   - Drive `isReportWord`, top-level dispatch, scoped help, and the global command list from this table. `isReportWord` may remain as a thin wrapper for compatibility but must not contain a second hard-coded list.
   - Add a parity test: every `case` in the dispatch switch and every `isReportWord` value has exactly one `commandSpec`, and vice versa. Adding a verb without a help page must fail the test.
   - Cross-check flags against `validateReportFlags`: no help page may list a flag that the command rejects, and no accepted command-specific flag may be omitted.

3. **Scoped page anatomy (every verb).**
   - `Usage:` one-line synopsis including `[tool]` and `[options]` positions.
   - `Description:` two to four sentences stating what the command answers and what it does not do.
   - `Reports:` for positional verbs, the valid report words and tool aliases.
   - `Flags:` command-specific flags first, then a `Global flags:` section; each line shows `<value>` placeholders, the default, and the accepted value set.
   - `Examples:` at least two copy-pasteable invocations, one text and one `--json`.
   - `JSON:` the top-level envelope keys and the row keys, or `not supported` when `--json` is rejected.
   - `Exit codes:` `0` success, `1` validation/data error, `2` flag parse error.
   - `See also:` two related commands.
   - `Minimum width:` the narrowest supported rendering (link to `cli-terminal-width-adaptation`).

4. **Side-effect freedom.**
   - A help path must not call `loadUsage`, any loader, `pricing.Load`, `render.LoadStatuslineCache`, `runUpgrade`, `tea.NewProgram`, the daily update check, or any cache writer.
   - A help path must not create, modify, or delete files. Test with a temporary `HOME`, a read-only `~/.cache/thermal`, and a snapshot of file mtimes before/after.
   - A help path must not perform network I/O. Test by running with an `http_proxy` pointing at a closed port and a 2-second timeout; help must return immediately.
   - `thermal live --help` must not enter the alternate screen or emit `\x1b[?1049h`/`\x1b[?1049l`; assert the byte stream contains no TUI frame marker.
   - `thermal upgrade --help` must not resolve `os.Executable`, download, chmod, rename, or write `.old`/`.new`; verify with a disposable copy and md5 equality.

5. **Output and exit contract.**
   - Help goes to stdout, never stderr, and ends with exactly one trailing newline.
   - Exit `0` for a valid context, `1` for an unknown context, `2` only for a genuine flag parse error that is not accompanied by a help request.
   - Color is emitted only when stdout is a TTY and `NO_COLOR` is unset; `--no-color` wins everywhere.
   - Every page fits 80 columns and also renders cleanly at 120; no line exceeds the requested width.
   - `--json` combined with `--help` prints help, not JSON; `--json` must not suppress help.

6. **Aliases and legacy forms.**
   - `thermal prompt --help` and `thermal status --help` must resolve to the documented `statusline` page while retaining the alias name in `Usage:`.
   - `thermal dashboard --help`, `thermal serve --help`, and other retired verbs must print the existing retirement message plus a `See also: thermal live` line, exit `1`, and must not attempt to execute the retired path.
   - `thermal -h` and `thermal --help` must remain byte-identical; `thermal help` must equal the same global page.

7. **Documentation and migration.**
   - Add a `help` row to the global command list and an `Examples` line: `thermal daily --help`, `thermal help daily`.
   - Add a six-part `docs/MIGRATION.md` entry: what changes (help now wins over execution), who it hits (scripts relying on `daily --help` producing data), what to do (remove `--help` from data pipelines; use `daily --json`), detection command, fix command, escape hatch (none needed; `--` does not disable help).
   - Update `README.md` and `project.jsonc` capability text.
   - `scripts/simulated_user_gate.sh` must fail if any verb listed by `isReportWord` lacks a help page or if `thermal <verb> --help` produces a table/TUI.

## Non-Goals
- Rewriting the wording of the existing global help (owned by `help-supported-tools-and-flags`).
- Adding shell completion generation.
- Changing `version` flag semantics (owned by `cli-version-verbose-flag-contract`).
- Changing upgrade download, verification, or migration-notice logic (owned by `cli-upgrade-nondestructive-safety-flags` and `cli-staged-deprecation-and-upgrade`).

## Acceptance Criteria
- AC-001: For every verb in `daily weekly monthly projects models stats mix trend live statusline share audit changelog upgrade version license`, `thermal <verb> --help` exits `0`, writes to stdout, and contains the verb's `Usage:` line. A loop over the list records all 14 exit codes as `0`.
- AC-002: `thermal daily --help` contains `--since`, `--until`, `--last`, `--breakdown`, `--chart`, and `--json`, and does **not** contain a `Daily Usage` table border or a data row matching `^  │  [0-9]{4}-[0-9]{2}-[0-9]{2}`.
- AC-003: `cmp <(thermal daily --help) <(thermal help daily)` exits `0`; the same holds for `weekly`, `live`, and `upgrade`.
- AC-004: `thermal live --help` in a PTY exits without `\x1b[?1049h`, and no `THERMAL LIVE` frame is captured.
- AC-005: `cp "$(command -v thermal)" /tmp/thermal-help-copy && before=$(md5sum …) && /tmp/thermal-help-copy upgrade --help && after=$(md5sum …)`; `before == after`, exit `0`, and no `.old`/`.new` files are created.
- AC-006: `HOME=$(mktemp -d) thermal --help` and `HOME=$(mktemp -d) thermal upgrade --help` both exit `0` with no `no supported tool data found` text and no files created under the temporary HOME.
- AC-007: `thermal help nope` exits `1`, contains `unknown command "nope"`, a suggestion line, and `thermal --help`; it does not contain a scoped `Usage:` block for `nope`.
- AC-008: `thermal help help` exits `0` and documents both `thermal <verb> --help` and `thermal help <verb>`.
- AC-009: `thermal --json daily --help` and `thermal daily --help --json` both print help and exit `0`; stdout is not valid JSON.
- AC-010: `NO_COLOR=1 thermal daily --help | grep -c $'\x1b'` returns `0`; `thermal daily --help | awk 'length>80'` prints no lines.
- AC-011: `go test -race ./cmd/thermal/... ./internal/render/...` exits `0`, including `TestHelpInterceptionMatrix`, `TestScopedHelpParity`, `TestHelpHasNoSideEffects`, and `TestHelpUnknownCommand`.
- AC-012: `./scripts/simulated_user_gate.sh` exits `0` and includes a help-precedence section that would fail on the pre-fix binary.

## Negative Cases & Failure Modes
| Invocation | Required result |
| --- | --- |
| `thermal daily --help` | Scoped daily help, exit 0, no table |
| `thermal daily -h` | Byte-identical to the above |
| `thermal -h daily` | Scoped daily help (help wins over flag order) |
| `thermal --json daily --help` | Help on stdout, exit 0, not JSON |
| `thermal help` | Global help, exit 0 |
| `thermal help help` | Help-about-help, exit 0 |
| `thermal help nope` | Unknown command + suggestion, exit 1 |
| `thermal nope --help` | Unknown command + suggestion, exit 1 |
| `thermal live --help` in a TTY | Scoped help; no alternate screen, no key handling |
| `thermal upgrade --help` | Scoped help; binary md5 unchanged |
| `thermal upgrade --help --dry-run` | Help only; no download |
| `thermal dashboard --help` | Retirement message + `thermal live`, exit 1 |
| `thermal prompt --help` | `statusline` page, alias preserved |
| `thermal --help --bogus-flag` | Help wins; exit 0 (documented precedence) |
| `thermal daily --help --since not-a-date` | Help wins; exit 0, no validation error |
| `thermal help daily extra` | Scoped daily help, exit 0; extra tokens are ignored because help wins |

## Metrics & Regression Guards
- Help latency: `thermal daily --help` must complete in <50 ms warm and must not touch any tool database. Measure with `strace -f -e trace=openat` or a temp HOME with no data; record in evidence.
- Zero writes: before/after `find "$HOME" -newer "$marker"` must be empty.
- Zero network: run with `HTTPS_PROXY=http://127.0.0.1:9` and a 2 s timeout; exit 0.
- Page parity: a generated table of verb → page hash is stored in `cmd/thermal/testdata/help-pages.sha256` and updated only with a deliberate golden change.
- Gate: the simulated user gate asserts AC-001, AC-003, AC-005, AC-007, and AC-009.

## Risks & Mitigations
- **Risk:** A global pre-scan could break `--help` as a future value for a flag. **Mitigation:** reserve `-h/--help` permanently; document in the registry and add a test that no value-taking flag accepts `help` as a separate token.
- **Risk:** Scoped pages drift from `validateReportFlags`. **Mitigation:** generate flag lists from the same metadata table used by validation, and add a parity test that walks the flagset.
- **Risk:** `thermal upgrade --help` is destructive, so a regression is costly. **Mitigation:** test on a disposable copy in CI; never run the real upgrade path in tests; assert md5 equality.
- **Risk:** Help pages become a maintenance burden. **Mitigation:** one `commandSpec` table, one renderer, golden snapshots, and a gate that fails when a verb is added without a page.
- **Risk:** Changing `daily --help` from data to help surprises a script. **Mitigation:** MIGRATION entry and a one-release deprecation note in the changelog; scripts should use `--json`.

## Memory Plan
- **Prime**: run `sila prime` at task start and record the briefing timestamp.
- **Recall**: run `sila recall "help flag swallowed positional"` and `sila recall "flag package stops at first positional"`; record match counts.
- **Writes**: record a `case` for the "pre-parse help interception with a command registry" mechanism, grounded to `cmd/thermal/help.go`; record a `lesson` if the pre-scan conflicts with flag parsing.
- **Facts**: run `sila facts verify` before certification and record the exit code and violation delta.
- **Handoff**: run `sila handoff "help interception + scoped help shipped" --open="width-aware pages"` at session end.
- **Evidence**: embed a `sila/memory-writes/v1` block under `## Memory Writes`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`); unchecked tasks block the `implemented` transition.
2. Implementation is committed with `(goals/cli-help-interception-and-scoped-help/goal.md)` attribution under a `feat(cli):` subject.
3. Goal is certified (`goals/cli-help-interception-and-scoped-help/certification.json` with `verified: true`) and `evidence.md` contains a standalone `Verdict: PASS` line plus every AC command and observed exit code.
4. `sila goals --json` reports `status: "implemented"`, `verification_status: "verified"`, and `unverified_count: 0`, with a clean `git status`.
