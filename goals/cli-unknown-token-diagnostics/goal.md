# Goal: Classify unknown commands, tools, and extra positionals with actionable suggestions

## Goal Description
Priority: P1 (error empathy). Every unrecognized positional is currently funnelled into `opts.Tool` and reported by `loaders.ResolveTool`, so the CLI blames a "tool" for tokens that are commands, paths, or stray arguments. The message contains no suggestion, no valid vocabulary, and no next step.

Observed on the current source build (2026-10-29):

- `thermal today` → `thermal: unknown tool: today`, exit `1`.
- `thermal daly`, `thermal weak`, `thermal streak`, `thermal heatmap`, `thermal show`, `thermal tools`, `thermal cost`, `thermal top`, `thermal me` → the same message with the token substituted.
- `thermal stats weekly` → `thermal: unknown tool: weekly`, exit `1`, even though `weekly` is a valid top-level command.
- `thermal projects /tmp` → `thermal: unknown tool: /tmp`, exit `1`.
- `thermal projects extra --top 3` → `thermal: unknown tool: extra`, exit `1`.
- `thermal daily --metirc cost` → `thermal: unknown tool: cost`, exit `1` (the typo flag is silently ignored and its value is blamed).
- `thermal --tool nope` → `thermal: unknown tool: nope`, exit `1`.
- `thermal dashboard` is the counter-example: it prints a specific retirement message with two alternatives and exits `1`.
- There is no `did you mean` path anywhere in the CLI.

The outcome of this goal is a **typed resolution layer**: each token is classified before it is blamed (command, tool, report word in the wrong slot, path-like token, retired verb, extra argument, empty value), and every unknown token gets a nearest-match suggestion, the valid vocabulary for that slot, and a `thermal --help` pointer.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: cli-argument-parser-strictness
- **Sequence**: 245
- **Shape**: ship
- **Tier**: roadmap
- **Dependency Reason**: provides typed parse errors and the positional model
- **Memory Protocol**: required
- **Wave**: 2
- **Related goals**: `cli-help-interception-and-scoped-help` owns `help <unknown>`; this goal owns unknown command/tool/positional diagnostics. `cli-argument-parser-strictness` owns unknown *flags*; this goal must not duplicate that message.

## Files to Touch
- `cmd/thermal/diagnostics.go`
- `cmd/thermal/diagnostics_test.go`
- `cmd/thermal/main.go`
- `internal/loaders/registry.go`
- `internal/render/help.go`
- `scripts/simulated_user_gate.sh`
- `README.md`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- **Shared Understanding & Fact Sheet**: [`goals/cli-unknown-token-diagnostics/facts.md`](facts.md)
- **Execution Plan**: [`goals/cli-unknown-token-diagnostics/plan.md`](plan.md)

## Requirements

1. **Classify before blaming.**
   - Build an ordered classifier for a single unknown token:
     1. empty → `thermal: empty argument is not a command or tool` (exit `2`).
     2. retired verb (`dashboard`, `serve`, `web`, `replay`, `yield`, `receipt`, `stats --dense` aliases) → existing retirement message + `See also`, exit `1`.
     3. known command or command alias in the command slot → valid; no error.
     4. known command or command alias in a non-command slot → `thermal: "weekly" is a report command; run thermal weekly`, exit `2`.
     5. path-like token (contains `/`, `\`, starts with `.`, `~`, or `/`) → `thermal: "…" looks like a path; thermal commands do not take paths. Use --db <path> for a data source.`, exit `2`.
     6. known tool or tool alias in the tool slot → valid; no error.
     7. close match (distance threshold) → suggestion.
     8. unknown → `thermal: unknown command "…"` (command slot) or `unknown tool "…"` (tool slot), plus a suggestion if available and the valid vocabulary, exit `2`.
   - The word "tool" must never appear for a token that is a command, a report word, a path, or an extra positional.
   - `stats weekly` must say "weekly is a report command", not "unknown tool".
   - `projects /tmp` must say "looks like a path" and mention `--db`.
   - `daily --metirc cost` must be handled by the flag parser as `unknown flag --metirc`; this goal only verifies it is not blamed as a tool.

2. **Suggestion engine.**
   - Use Damerau-Levenshtein distance, case-insensitive, trimmed, with a threshold of `min(2, max(1, len(token)/3))`; prefer the smallest distance, then the shortest candidate, then alphabetical.
   - Candidate pools by slot: command names + command aliases; tool canonical names + tool aliases; report words.
   - Add a small synonym map for intent-level guesses: `today`/`yesterday` → `daily`, `week` → `weekly`, `month` → `monthly`, `repo`/`project` → `projects`, `model` → `models`, `chart` → `daily --chart`, `heatmap`/`streak` → the leaderboard (`thermal`), `cost`/`spend` → `stats --metric cost`, `status` → `statusline`.
   - If a synonym maps, use `Did you mean: thermal daily?` even when the edit distance is large.
   - Never suggest a retired verb as a replacement; explicitly say it is retired and point to the replacement.
   - Never suggest a token from a different slot when the current slot is known (e.g. `--tool today` should not suggest the `daily` command; it should say `unknown tool "today"` and list tools).
   - Limit to one primary suggestion and up to three alternatives.

3. **Message anatomy.**
   - Line 1: `thermal: <classification> "<token>"`.
   - Line 2 (optional): `Did you mean: thermal <suggestion>?`.
   - Line 3 (optional): `Valid <commands|tools|reports>: <comma list>` wrapped to 80 columns.
   - Line 4: `run thermal --help for the command list` or `thermal --help for supported tools`.
   - Maximum five lines; no box art; no full manual.
   - Every line begins `thermal: ` only on the first line; continuation lines are indented two spaces.
   - stdout is empty; all diagnostics on stderr.
   - Color only when stderr is a TTY and `NO_COLOR` is unset.

4. **Exit codes.**
   - Unknown command, unknown tool, path-like token, wrong-slot report word, extra positional, empty token: exit `2` (usage).
   - Retired verb: keep exit `1` for compatibility with the existing retirement messages.
   - Valid command with missing data: exit `1` (unchanged).
   - Document the change from `1` to `2` for unknown tools/commands in `docs/MIGRATION.md`.

5. **Vocabulary sources.**
   - Commands: the `commandSpec` registry from `cli-help-interception-and-scoped-help`; until that lands, a single exported list in `cmd/thermal/diagnostics.go` that the help goal can adopt.
   - Tools: `loaders.AllTools()` plus the alias map (`mimo`, `oc`, `cmd`, `whale`, `zc`, `ccode`, `dsh`, `deepseek`, `hermes`, `nous`, `agy`, `muse`, `droid`). Do not hard-code display names twice.
   - Reports: `isReportWord` / the registry; include `daily`, `weekly`, `monthly`, `projects`, `models`, `stats`, `mix`, `trend`, `live`, `statusline`, `audit`, `changelog`, `share`.
   - A test asserts every alias resolves and every canonical name appears in the suggestion pool.

6. **Tests.**
   - Typo corpus (≥40): `daly→daily`, `weeky→weekly`, `monyhly→monthly`, `projcts→projects`, `modles→models`, `stat→stats`, `mxi→mix`, `tren→trend`, `lvie→live`, `statuslin→statusline`, `audti→audit`, `changelog→changelog` (valid), `upgrad→upgrade`, `verison→version`.
   - Tool corpus (≥20): `opnecode→opencode`, `codx→codex`, `clade→claude`, `zcod→zcode`, `gork→grok`, `mimo→MiMoCode`, `cmd→command-code`, `dsh→DeepSeek (DSH)`, `nous→Nous Hermes`.
   - Wrong-slot corpus: `stats weekly`, `projects /tmp`, `/abs/path`, `~/data`, `daily stats`, `models mix`, `--tool daily`.
   - False-positive corpus: valid commands, valid tools, `all`, `auto`, `help`, `version`, `license`, short tokens (`a`, `x`) must not produce a misleading suggestion.
   - Snapshot tests for the five-line maximum and the exact first line.
   - A property test: no message contains `unknown tool` unless the slot is a tool.

7. **Documentation and migration.**
   - MIGRATION entry: unknown commands/tools now exit `2` instead of `1`; messages are different; scripts should treat any non-zero as failure and parse nothing from stderr.
   - README: add a `Troubleshooting` subsection with three examples.
   - `project.jsonc`: mention suggestions.
   - `scripts/simulated_user_gate.sh`: assert `today`, `stats weekly`, `projects /tmp`, `--tool nope`, and `dashboard`.

## Non-Goals
- Fixing unknown *flags* (owned by `cli-argument-parser-strictness`).
- Adding shell completion or dynamic command discovery.
- Changing the retirement messages' wording (only append `See also` if missing).
- Adding telemetry for typo tracking (zero telemetry invariant).

## Acceptance Criteria
- AC-001: `thermal today` exits `2`, stdout empty, stderr contains `unknown command "today"` and `Did you mean: thermal daily?`; it does not contain `unknown tool`.
- AC-002: `thermal daly` exits `2` and suggests `thermal daily`; `thermal weak` suggests `thermal weekly`.
- AC-003: `thermal stats weekly` exits `2`, contains `"weekly" is a report command`, and suggests `thermal weekly`; it does not contain `unknown tool`.
- AC-004: `thermal projects /tmp` exits `2`, contains `looks like a path`, mentions `--db`, and does not suggest a tool.
- AC-005: `thermal --tool nope` exits `2`, contains `unknown tool "nope"`, lists at least five tool names, and does not suggest a command.
- AC-006: `thermal --tool opencode`, `thermal opencode`, `thermal daily`, and `thermal --tool all` still exit `0` (no false positive).
- AC-007: `thermal dashboard` keeps the existing retirement message, exit `1`, and includes `See also: thermal live`.
- AC-008: `thermal a` exits `2` with the command list and no confident suggestion (or a suggestion only if the distance threshold allows).
- AC-009: Every sampled diagnostic is ≤5 lines, starts with `thermal: `, and writes nothing to stdout.
- AC-010: `NO_COLOR=1 thermal today 2>&1 | grep -c $'\x1b'` returns `0`.
- AC-011: `go test -race ./cmd/thermal/... ./internal/loaders/...` exits `0`, including `TestUnknownCommandCorpus`, `TestUnknownToolCorpus`, `TestWrongSlotCorpus`, `TestSuggestionFalsePositives`, and `TestDiagnosticLineBudget`.
- AC-012: `./scripts/simulated_user_gate.sh` exits `0` and includes the diagnostics section.

## Negative Cases & Failure Modes
| Invocation | Required result |
| --- | --- |
| `today` | unknown command, suggest `daily`, exit 2 |
| `daly` | suggest `daily`, exit 2 |
| `weak` | suggest `weekly`, exit 2 |
| `stats weekly` | report word in wrong slot, suggest `thermal weekly`, exit 2 |
| `projects /tmp` | path message, mention `--db`, exit 2 |
| `projects ~/data` | path message, exit 2 |
| `--tool nope` | unknown tool, list tools, exit 2 |
| `--tool daily` | unknown tool, do not suggest the `daily` command |
| `daily --metirc cost` | unknown flag `--metirc` (from the parser goal), not `unknown tool: cost` |
| `dashboard` | retirement message, `See also: thermal live`, exit 1 |
| empty positional | empty-argument message, exit 2 |
| valid token | no suggestion, normal execution |

## Metrics & Regression Guards
- Suggestion precision: the false-positive corpus (valid commands/tools) must produce zero suggestions.
- Suggestion recall: the typo corpus (≥40) must hit the intended target for ≥90% of cases; record the rate in evidence.
- Line budget: every diagnostic ≤5 lines at 80 columns.
- Slot integrity: zero messages containing `unknown tool` for command/path/extra-positional slots.
- Gate: the simulated user gate asserts AC-001, AC-003, AC-004, AC-005, and AC-007.

## Risks & Mitigations
- **Risk:** Aggressive suggestions annoy users ("did you mean `agy`?" for `a`). **Mitigation:** distance threshold and a false-positive corpus; small tokens get no suggestion.
- **Risk:** Changing exit `1` → `2` breaks scripts. **Mitigation:** MIGRATION entry and a changelog note; any non-zero should already be treated as failure.
- **Risk:** A second vocabulary list drifts from the help registry. **Mitigation:** depend on `cli-help-interception-and-scoped-help`'s `commandSpec`; otherwise export one list and add a parity test.
- **Risk:** Path detection misclassifies a project name containing a slash. **Mitigation:** only classify as a path when the token contains a separator or starts with a path prefix, and never for `--tool` values that resolve to a known alias.
- **Risk:** Synonyms become stale as commands change. **Mitigation:** keep the synonym map in the same registry file and test every target.

## Memory Plan
- **Prime**: run `sila prime` at task start.
- **Recall**: run `sila recall "unknown tool message suggestions"` and `sila recall "CLI typo diagnostics"`.
- **Writes**: record a `case` for the "slot-aware token classifier with a synonym map" mechanism, grounded to `cmd/thermal/diagnostics.go`; record a `lesson` for the misattributed `unknown tool` class.
- **Facts**: run `sila facts verify` before certification.
- **Handoff**: run `sila handoff "unknown-token diagnostics shipped" --open="shell completion"`.
- **Evidence**: embed a `sila/memory-writes/v1` block under `## Memory Writes`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`).
2. Implementation is committed with `(goals/cli-unknown-token-diagnostics/goal.md)` under a `feat(cli):` or `fix(cli):` subject.
3. Goal is certified and `evidence.md` contains a standalone `Verdict: PASS` plus the corpus recall/precision numbers.
4. `sila goals --json` reports `status: "implemented"`, `verification_status: "verified"`, and `unverified_count: 0`, with a clean `git status`.
