# CLI UX Hardening Program (post-0.21.0)

Status: **registered, not implemented** · Date: 2026-10-29 · Owner: Thermal maintainer
Source audit: persona stress test of the current source build (`thermal dev`) and the
v0.21.0 release binary, using a real PTY, an isolated `HOME`, and disposable binary copies
for destructive paths.

## Verdict

The data layer and JSON contract are strong: JSON is valid on all ten report surfaces,
ANSI is absent under `--json`, `--no-color`, `NO_COLOR`, and `TERM=dumb`, empty windows are
polite, and `statusline` is 3–5 ms warm. The **help, error, and geometry layers** are the
weak seam. They answer a different question than the one asked, and three paths return a
confident wrong answer with exit `0`.

This program turns the audit's 20 friction points into **13 Sila goal packages**, ordered by
severity. Every package has `goal.md` (requirements, acceptance criteria, negative cases,
risks, memory plan), `facts.md` (observed evidence, root cause, contracts, verification),
and `plan.md` (phases, acceptance ledger, rollback, closure).

**Do the four P0 goals first.** They remove paths that can silently return the wrong data
or replace the binary. Everything else is clarity, geometry, or automation polish.

## Goal packages

| # | Goal slug | Priority | Wave | Logical deps | Primary audit findings |
| --- | --- | --- | --- | --- | --- |
| 1 | `cli-upgrade-nondestructive-safety-flags` | P0 safety | 1 | none | `upgrade --help/--check/--dry-run/--json` perform a real self-replacement |
| 2 | `cli-argument-parser-strictness` | P0 trust | 1 | none | invalid values after a positional silently return all-time data; conflicts ignored; reversed dates return empty |
| 3 | `cli-db-override-scope-and-errors` | P0 scope | 1 | none | `--db` ignored in aggregate views; raw SQLite codes leak |
| 4 | `cli-help-interception-and-scoped-help` | P0 discoverability | 2 | none | `daily --help` renders data; `live --help` opens the TUI; `help daily` is an unknown tool |
| 5 | `cli-version-verbose-flag-contract` | P1 clarity | 2 | none | `-v` means version only when alone; help says verbose; `-V` errors |
| 6 | `cli-unknown-token-diagnostics` | P1 empathy | 2 | parser | every unknown token is blamed as an "unknown tool", no suggestions |
| 7 | `cli-first-run-source-inventory` | P1 activation | 3 | none | empty HOME prints one line, exit 1, no paths or next steps |
| 8 | `cli-nontty-and-statusline-output-contract` | P1 automation | 3 | none | `live` prose on stdout/exit 0; statusline errors on every prompt with no data |
| 9 | `cli-terminal-width-adaptation` | P1 geometry | 3 | none | reports 98–119 cols at any width; live clips below 80; `COLUMNS` ignored |
| 10 | `cli-top-flag-contract` | P1 consistency | 3 | parser | `--top` rejected on daily/leaderboard, accepted on changelog, recommended by the chart anyway |
| 11 | `cli-leaderboard-ranking-and-legend` | P1 clarity | 4 | none | ranking key invisible; `Strk/Best/Days` and `~ / — / ≥` undefined outside live help |
| 12 | `cli-report-total-row-semantics` | P2 semantics | 4 | none | total rows show a tool union and sum P-Days/M-Days across rows |
| 13 | `cli-analytics-and-audit-prose-clarity` | P2 prose | 4 | none | two "Mean" values with different denominators; audit `Biggest win: Critical:`; zero bins |

`Logical deps` are what Sila records. They are not the whole execution story: most goals
touch `cmd/thermal/main.go` and `internal/render/help.go`, so the **merge order** below is
the practical order. `--independent` means "no logical prerequisite", not "no file overlap".

## Coverage matrix: audit finding → goal

### Impatient Beginner
| Finding | Goal |
| --- | --- |
| Empty state is a dead end (`no supported tool data found`) | 7 |
| Unknown verbs (`today`, `daly`, `weak`) blamed as tools | 6 |
| Extra positionals (`stats weekly`, `projects /tmp`) blamed as tools | 6 |
| `daily --help` renders data; `live --help` opens the TUI; `help daily` fails | 4 |
| `-v` documented as verbose but prints version | 5 |
| Leaderboard ranking key invisible; column/cost legend missing | 11 |
| `--top` rejected on the tables users try; changelog truncation has no affordance | 10 |
| Changelog `Showing 3 of 25` with no way to see more | 10 |

### Clumsy User
| Finding | Goal |
| --- | --- |
| `upgrade --help/--check/--dry-run/--json` mutates the binary | 1 |
| `daily --last abc`, `--weeks abc`, `--top xyz`, `--interval abc` silently ignored | 2 |
| `--db` ignored in aggregate views; invalid paths return default data | 3 |
| Raw SQLite errors (`disk I/O error (10)`, `file is not a database (26)`) | 3 |
| Reversed `--since`/`--until` returns "No activity" | 2 |
| `opencode --tool codex` silently picks the positional | 2 |
| `--tool=` prints `unknown tool: ` | 2 + 6 |
| `--metirc cost` blames the value, not the flag | 2 |
| Flag parse errors dump the 84-line manual and lack the `thermal:` prefix | 2 |
| `-V` is an unknown flag | 5 |

### Script Automator
| Finding | Goal |
| --- | --- |
| `thermal live` in a pipe writes prose to stdout and exits 0 | 8 |
| `statusline` exits 1 with stderr noise in an empty environment | 8 |
| Reports are 98–119 columns wide; `COLUMNS` ignored; live clips below 80 | 9 |
| `stats` and `trend` both say "Mean" with different denominators | 13 |
| `projects` total row shows a tool union and 505 P-Days vs 96 active days | 12 |
| `audit` run-on `Biggest win: Critical:`; stats outlier self-reference; zero bins | 13 |
| `upgrade --json` prints prose and mutates | 1 |

### Checked and already good (do not regress)
- JSON validity on `leaderboard, daily, projects, models, stats, mix, trend, live, audit, weekly`.
- stderr/stdout separation for data errors; exit `1`/`2` conventions on the healthy paths.
- Zero ANSI under `--json`, `--no-color`, `NO_COLOR=1`, `TERM=dumb`; color auto-disables in pipes.
- Empty date windows print `No activity in the selected window.` and exit `0`.
- `share` discloses its truncated URL and points to `--json`.
- `statusline` warm latency 3–5 ms, matching the advertised `<3 ms`.
- The live 80x24 and 120x40 frames are clean; the clipping starts below 80.
- The live `[?]` overlay's cost legend is the right vocabulary; it should be shared, not lost.

## Program-level contracts

These are cross-cutting invariants; each goal's ACs instantiate them.

1. **Precedence:** help > version > argument validation > execution. `--help` and `--version`
   never execute a command, never mutate a binary, and never read data.
2. **Strict parsing:** every value-taking flag has one parser, one error message, and exit
   `2` on malformed input. No invalid value may be silently replaced by a default.
3. **One meaning per spelling:** `-v` = verbose, `-V`/`--version` = version, `-h`/`--help` =
   help. No flag changes meaning based on argv length or position.
4. **Scope honesty:** `--db` requires one tool and never falls back to defaults; `--top` has
   one applicability table; totals describe the full window, not the capped rows.
5. **Channel purity:** stdout is data; stderr is human guidance and diagnostics. Prompt
   helpers degrade silently with exit `0`; deliberate commands may fail loudly.
6. **Geometry:** every line fits the resolved width or the command refuses with a one-line
   minimum-width message. `--json` is width-independent.
7. **No remediation lies:** a message may only suggest a flag/command that the same
   invocation would accept.
8. **Additive JSON:** new keys are additive; semantic changes to existing keys (e.g.
   `totals.activeDays`) are documented in `docs/MIGRATION.md`.

### Exit codes (program-wide)
| Code | Meaning |
| --- | --- |
| `0` | Success, help, version, or a prompt helper with no data |
| `1` | Data, path, network, or semantic failure after a successful parse |
| `2` | Usage/parse/scope error, including minimum-width refusals and non-TTY `live` without a machine format |
| `3` | Upgrade declined or non-TTY mutation without `--yes` |

## Recommended merge order

Sila's dependency graph has 11 independent and 2 dependent goals (both parser-dependent). Because `cmd/thermal/main.go` and `internal/render/help.go`
are shared, use this order to avoid rebasing the same hunks:

1. **`cli-argument-parser-strictness`** — establishes `cmd/thermal/args.go`, the flag table,
   and the `usageError` voice. Everything else builds on it.
2. **`cli-help-interception-and-scoped-help`** — adds `cmd/thermal/help.go` and the
   `commandSpec` registry; parser errors and help pages then share one vocabulary.
3. **`cli-version-verbose-flag-contract`** — resolves the version/verbose half of the
   precedence table and updates the same help pages.
4. **`cli-upgrade-nondestructive-safety-flags`** — uses `upgradeOptions`, its own parser, and
   the upgrade help page; lands after help so `--help` is guaranteed non-destructive.
5. **`cli-db-override-scope-and-errors`** — uses the parser's `--db` value type and the help
   page.
6. **`cli-unknown-token-diagnostics`** — consumes the typed parse errors and the command
   registry.
7. **`cli-top-flag-contract`** — consumes the parser's `TopSet` and the help applicability
   table.
8. **`cli-first-run-source-inventory`** — adds the stat-only inventory and the `sources`
   verb; refactors `loadUsage` to return errors.
9. **`cli-nontty-and-statusline-output-contract`** — uses the `loadUsage` error and the
   inventory summary for quiet degradation.
10. **`cli-terminal-width-adaptation`** — adds the width resolver and threads it through the
    renderers; do this before content-heavy render changes to avoid double layout work.
11. **`cli-leaderboard-ranking-and-legend`** — content for the leaderboard on the new width
    engine.
12. **`cli-report-total-row-semantics`** — totals data model and typed total cells.
13. **`cli-analytics-and-audit-prose-clarity`** — final prose pass on the settled layouts.

If two goals must land in parallel, branch from the same base and keep the shared changes
(`Options`, `commandSpec`, `reportWidth`) in the earlier goal.

## Verification strategy

Every goal must satisfy its own ACs **and** the program gates:

```bash
# Fast local gate
./scripts/check.sh --pre-commit

# Full gate: race, coverage, vet, build, simulated user flows
./scripts/check.sh --full

# Release-facing user simulation
./scripts/simulated_user_gate.sh

# Sila closure
sila goals --json
```

Program-level evidence requirements:

- **Exit-code matrix:** every goal records the command, exit code, stdout, and stderr for its
  ACs in `goals/<slug>/evidence.md`.
- **Channel matrix:** `live`, `statusline`, `share`, and a report JSON error are captured with
  separate stdout/stderr files.
- **Width matrix:** every command at `COLUMNS=40, 50, 56, 60, 80, 100, 120, 160`.
- **PTY captures:** `live` at 50x20, 60x20, 80x24, 100x30, 120x40, 160x40.
- **Binary integrity:** md5 before/after for every non-mutating upgrade path.
- **No-fallback proof:** a sentinel default HOME for `--db`; a sentinel transcript for the
  source inventory privacy test.
- **JSON snapshots:** additive keys and deterministic envelopes; no ANSI, no aliases.
- **Gate diff:** the simulated user gate must fail on the pre-fix binary for at least one
  assertion per goal.

## Documentation and migration obligations

Each goal that changes a user-visible contract must add a six-part
`docs/MIGRATION.md` entry (stable heading, detection regex/command, 3-sentence rationale,
before/after, fix command, escape hatch) and update `README.md` and `project.jsonc` in the
same commit. The changes that definitely need entries:

- `-v` now means verbose; use `-V`/`--version`.
- `--db` now requires one tool; invalid paths fail.
- `--top` now works on the leaderboard/period reports; rejecting commands exit `2`.
- `live` non-TTY without `--json/--stream` exits `2` instead of `0`.
- `statusline` with no data exits `0` and prints nothing.
- `totals.activeDays` now counts distinct calendar days; `projectDays`/`modelDays` added.
- Unknown commands/tools exit `2` with suggestions.
- Help now wins over execution; `daily --help` no longer prints data.

## Non-goals and governance

- No new cloud telemetry, daemons, or background processes (GOVERNANCE.md anti-goals).
- No new third-party dependencies; the parser, suggestions, and width logic are stdlib-only
  and CGO-free.
- No aggregation or pricing math changes in the UX goals; totals goals add semantics, not
  new formulas.
- No width/layout work in the content goals beyond what the width contract defines.
- No commit or push without the maintainer's instruction; goals are registered for execution,
  not executed by this program document.
- Every implementation commit uses Conventional Commits and a goal attribution token:
  `fix(cli): ... (goals/<slug>/goal.md)` or `[goal: <slug>]`.

## Registration evidence (2026-10-29)

- `sila goals --sync` → 134 total goals, 121 implemented, **13 pending**, 11 ready,
  2 blocked on `cli-argument-parser-strictness`, 66 open tasks.
- `sila goals --json` reports the 13 pending slugs listed in the table above and
  `pending_count: 13`.
- `./scripts/check.sh --pre-commit` cleared in 1 s (docs/goals-only diff; no Go packages
  modified).
- `sila memo fact_64a4f44b` records this program and points at this document.
- No goal is claimed, implemented, or certified by this program document.

## Start here

```bash
sila goals ready
sila goals claim cli-argument-parser-strictness --agent=<name>
/goal goals/cli-argument-parser-strictness/goal.md
```

After each goal:

```bash
sila goals report <slug> "phase N complete"
sila goals certify <slug> --evidence=goals/<slug>/evidence.md
sila goals            # reconcile README and receipts
```

The program is complete when `sila goals stats` reports 100% for these 13 slugs,
`./scripts/check.sh --full` and `./scripts/simulated_user_gate.sh` exit `0`, and the P0/P1
findings above have a regression test that fails on the pre-fix binary.
