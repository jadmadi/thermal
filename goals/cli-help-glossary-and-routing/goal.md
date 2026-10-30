# Goal: Route every unknown word to the right namespace and explain domain vocabulary

## Goal Description
Priority: P1 (discoverability). Thermal's banner says "Ranked by current
streak" and its leaderboard is titled "Token Warriors", but `thermal help
streak` and `thermal help token` answer "unknown command". Beginner words such
as `doctor` and `config` are diagnosed as unknown *tools* with a 15-name vendor
list, while `thermal top` suggests `oc` — a two-letter alias that appears
nowhere in the prose. `thermal --tool open` gets no suggestion at all, and
`thermal --tool code` suggests `ccode` without saying it means Claude.

Observed on `2026-10-30` with `/tmp/thermal-cli-review` (`thermal dev`):

- `thermal help streak` -> exit `1`:
  `thermal: unknown command "streak"` / `run thermal --help for the command list`
- `thermal help token` -> exit `1`: the same shape.
- `thermal doctor` -> exit `2`:
  `thermal: unknown tool "doctor"` plus the 15-tool list.
- `thermal config` -> exit `2`: the same shape.
- `thermal top` -> exit `2`:
  `thermal: unknown tool "top"` / `did you mean: "oc"?` plus the 15-tool list.
- `thermal --tool open` -> exit `2`: no suggestion, the 15-tool list.
- `thermal --tool code` -> exit `2`:
  `did you mean: "ccode"?`; `ccode` is absent from the visible list.
- The populated leaderboard ends with a cheer and no next command.

The outcome of this goal is a **help glossary and routing contract**:

1. Every unknown token is classified into one namespace: command, report,
   tool, flag, glossary, path, retired, or truly unknown.
2. Domain vocabulary is readable in-terminal: `help streak`, `help token`,
   `help warriors`, `help cost`, `help activity` explain the terms.
3. Suggestions have a quality floor, prefer canonical names, and always show
   the canonical form with any alias.
4. Beginner intent words route to the command that answers them (`sources`,
   `--help`) instead of a vendor list.
5. The leaderboard ends with one next-step hint.
6. Help examples are grouped and state their prerequisites.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: cli-unknown-token-diagnostics, cli-help-interception-and-scoped-help, cli-db-override-scope-and-errors
- **Sequence**: 260
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required
- **Wave**: 3
Related goals:
- `cli-data-access-error-guidance` owns a resolved path that cannot be read;
  this goal owns a word that was never resolved.
- `cli-empty-data-success-contract` owns the no-data onboarding; this goal owns
  the populated leaderboard's next step.
- `cli-window-flag-semantics` owns `--` and repeated flags; this goal owns the
  unknown-token message they produce.

## Files to Touch
- `cmd/thermal/help.go`
- `cmd/thermal/diagnostics.go`
- `cmd/thermal/args.go`
- `cmd/thermal/main.go`
- `internal/render/help.go`
- `internal/render/leaderboard.go`
- `cmd/thermal/help_test.go`
- `cmd/thermal/diagnostics_test.go`
- `cmd/thermal/args_strict_test.go`
- `scripts/simulated_user_gate.sh`
- `README.md`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- **Shared Understanding & Fact Sheet**:
  [`goals/cli-help-glossary-and-routing/facts.md`](facts.md)
- **Execution Plan**:
  [`goals/cli-help-glossary-and-routing/plan.md`](plan.md)
- Prior contracts:
  [`goals/cli-unknown-token-diagnostics/goal.md`](../cli-unknown-token-diagnostics/goal.md),
  [`goals/cli-help-interception-and-scoped-help/goal.md`](../cli-help-interception-and-scoped-help/goal.md)

## Requirements

1. **Glossary.**
   - Add a glossary with at least these terms:
     `streak`, `token`, `tokens`, `warriors`, `token warriors`,
     `activity hunters`, `active day`, `best`, `cost`, `estimated`,
     `recorded`, `unpriced`, `cache hit`, `scope`, `window`, `session`,
     `turn`.
   - `thermal help <term>` prints:
     ```
       thermal glossary · streak
       Consecutive calendar days with activity for a tool. The default
       leaderboard ranks by current streak; Best is the longest streak.

       See also: thermal --help · thermal daily · thermal stats
     ```
     exit `0`.
   - `thermal help glossary` lists every term with a one-line definition.
   - Glossary terms may overlap tool or command names (`cost`, `model`);
     the namespace is `help`, so the glossary wins.
   - `thermal <term> --help` for a glossary-only term prints the glossary page
     (help interception wins over execution), exit `0`.
   - The global help page adds one line:
     `Glossary: thermal help streak · thermal help glossary`.

2. **Namespace classification.**
   - Extend `tokenDiagnostic` with a class:
     `command`, `report`, `tool`, `flag`, `glossary`, `path`, `retired`,
     `unknown`.
   - Classify in this order:
     1. flag (`--...`) -> flag hint (`run thermal --help for flags`);
     2. exact command/report -> execute;
     3. exact tool/alias -> execute;
     4. glossary term -> glossary page;
     5. retired -> retirement message;
     6. path-like -> `--db` hint;
     7. near command -> command suggestion;
     8. near tool -> tool suggestion;
     9. beginner intent (`doctor`, `config`, `setup`, `history`, `usage`,
        `cost`, `top`) -> command namespace with a `sources`/`--help` next
        step;
     10. unknown -> command namespace with `thermal --help`.
   - A token beginning with `-` is never reported as an unknown tool.
   - Exit codes: command/tool/flag/path unknown = `2`; glossary = `0`;
     retired = `1` (existing contract); help unknown after the glossary
     fallback = `2`.

3. **Suggestion quality.**
   - `nearestName` keeps a minimum key length of 3 and a maximum edit distance
     of 2.
   - Prefer canonical names over aliases; if the best match is an alias, render
     `did you mean: Claude (alias "ccode")?` and include the canonical name in
     the valid list.
   - Never suggest an alias that is absent from the printed valid list unless
     its canonical name is printed on the same line.
   - Prefix matches count: `open` -> `OpenCode`, `codex` -> `Codex`,
     `dsh` -> `DeepSeek (DSH)`.
   - For `code`, offer `Claude (alias "ccode")`, `OpenCode (alias "oc")`, and
     `CMD`; do not choose a single low-confidence alias.
   - A suggestion must be a runnable command or a `--tool` value. Do not
     suggest `oc` alone.

4. **Beginner intent routing.**
   - `doctor`, `config`, `setup` ->:
     ```
     thermal: unknown command "doctor"
       checking what thermal can read:  thermal sources
       all commands:                    thermal --help
     ```
     exit `2`.
   - `history`, `usage` -> `thermal daily`/`thermal stats` suggestion.
   - `top` -> `--top <N>` flag suggestion, not `oc`.
   - `today`, `week`, `month` keep their existing synonym suggestions.
   - Never print the 15-tool list for a word that is not close to a tool.

5. **Leaderboard next step.**
   - Replace the trailing cheer-only block with a cheer plus one hint line:
     ```
     Next: thermal weekly --chart   weekly bars · thermal --help   all flags
     ```
   - Keep the existing "Keep the heat going" line; do not remove the
     personality.
   - At 80 columns the hint wraps to at most two lines.
   - `--no-color` keeps it plain; `--json` is unchanged.

6. **Help examples and grouping.**
   - Global help groups commands by workflow (already true); add a
     "Glossary" group or a single line under "Common Examples".
   - Every example either runs on a clean machine or states its prerequisite
     in the same block (coordinated with
     `cli-data-access-error-guidance`).
   - Add `thermal sources` and `thermal help glossary` to the examples.
   - Do not add a wall of text; the help must remain <= 105 lines at 80
     columns.

7. **Tests.**
   - Glossary table: every term returns exit `0`, contains the term, and a
     `See also` line.
   - `help streak`, `help token`, `help warriors`, `help activity`, and
     `help glossary` pass.
   - Classification table for one token per class.
   - Suggestion table: `top`, `doctor`, `config`, `open`, `code`, `ccode`,
     `oc`, `dsh`, `zzzz`.
   - Assert no suggestion is an alias without its canonical name.
   - Assert exit codes per class.
   - Leaderboard hint test at 80/120; JSON unchanged.
   - Gate: `simulated_user_gate.sh` covers the glossary and intent words.

8. **Documentation.**
   - README adds a "Glossary" subsection with the same definitions.
   - MIGRATION entry for `help <term>` exit `1` -> `0` for glossary terms and
     exit `2` for truly unknown terms; detection and fix.
   - `project.jsonc` updated if help output changes.

## Non-Goals
- Adding a `glossary` top-level command; it lives under `help`.
- Changing tool detection or loader data.
- Adding fuzzy matching below the quality floor.
- Changing the leaderboard ranking or columns.
- Replacing the personality text; the hint is additive.

## Acceptance Criteria
- AC-001: `thermal help streak` exits `0` and prints a definition plus
  `See also`.
- AC-002: `thermal help token`, `help warriors`, `help activity`, and
  `help glossary` exit `0`.
- AC-003: `thermal help glossary` lists at least 15 terms.
- AC-004: `thermal help nosuchterm` exits `2` and offers `thermal help
  glossary` and `thermal --help`.
- AC-005: `thermal doctor` exits `2`, names `doctor` as an unknown command,
  and points at `thermal sources`, not a 15-tool list.
- AC-006: `thermal top` exits `2` and suggests `--top <N>`, not `oc`.
- AC-007: `thermal --tool open` exits `2` and suggests `OpenCode (alias
  "oc")`.
- AC-008: `thermal --tool code` offers Claude (`ccode`), OpenCode, and CMD,
  each with a canonical name.
- AC-009: No suggestion prints an alias without its canonical name.
- AC-010: The populated leaderboard ends with one `Next:` hint line; `--json`
  is unchanged.
- AC-011: `thermal --help` contains `thermal help glossary`.
- AC-012: `go test -race ./cmd/thermal/... ./internal/render/...` exits `0`,
  including `TestGlossary`, `TestTokenClassification`,
  `TestSuggestionQuality`, and `TestLeaderboardNextHint`.
- AC-013: `./scripts/simulated_user_gate.sh` exits `0` and includes the
  glossary and intent-word cases.

## Negative Cases & Failure Modes
| Token | Required result |
| --- | --- |
| `help streak` | glossary, exit 0 |
| `help nosuchterm` | exit 2, glossary + help pointers |
| `doctor` | command namespace, sources pointer |
| `config` | command namespace, sources pointer |
| `top` | `--top` suggestion, no tool list |
| `--tool open` | OpenCode suggestion |
| `--tool code` | three canonical suggestions |
| `--tool oc` | works (alias) |
| `--tool zzzz` | no suggestion, valid list |
| `--jsno` | `--json` suggestion |
| `-x` | flag namespace, no tool list |
| `foo/bar` | path hint, `--db` |
| `dashboard` | retired message, exit 1 |

## Metrics & Regression Guards
- Glossary coverage: every term in the glossary table is reachable via
  `help <term>`; a test enumerates the table.
- Suggestion quality: zero alias-only suggestions across a fixture token set.
- Namespace correctness: one test per class; no token begins with `-` in the
  tool class.
- Leaderboard hint: exactly one `Next:` line at 80 and 120.
- Gate: simulated user gate asserts AC-001, AC-005, AC-006, and AC-010.

## Risks & Mitigations
- **Risk:** Glossary terms collide with commands (`cost`).
  **Mitigation:** glossary only applies to the `help` path; execution still
  resolves commands first.
- **Risk:** More suggestions become noise.
  **Mitigation:** a quality floor, canonical-first ordering, and a maximum of
  three suggestions.
- **Risk:** `help <term>` exit `0` hides a typo.
  **Mitigation:** only exact glossary terms exit `0`; near terms suggest the
  closest glossary entry and exit `2`.
- **Risk:** The leaderboard hint lengthens the footer.
  **Mitigation:** one line at 80, two maximum; keep the cheer.
- **Risk:** The glossary drifts from the product vocabulary.
  **Mitigation:** one table used by help and tests; add a test that every
  bold term in the README glossary exists in the table.

## Memory Plan
- **Prime**: run `sila prime` at task start.
- **Recall**: run `sila recall "help streak unknown command"` and
  `sila recall "did you mean oc"`.
- **Writes**: record a `case` for the "namespace classifier + glossary
  fallback" mechanism; record a `lesson` for alias-only suggestions.
- **Facts**: run `sila facts verify` before certification.
- **Handoff**: run `sila handoff "help glossary and routing shipped"
  --open="glossary coverage review"`.
- **Evidence**: embed a `sila/memory-writes/v1` block under
  `## Memory Writes`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`).
2. Implementation is committed with
   `(goals/cli-help-glossary-and-routing/goal.md)` under a `feat(cli):` or
   `fix(cli):` subject.
3. Goal is certified and `evidence.md` contains a standalone `Verdict: PASS`
   plus the glossary, classification, suggestion, and leaderboard evidence.
4. `sila goals --json` reports `status: "implemented"`,
   `verification_status: "verified"`, and `unverified_count: 0`, with a clean
   `git status`.
