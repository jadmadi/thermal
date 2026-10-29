# Goal: Give --top one documented contract and make every remediation hint valid

## Goal Description
Priority: P1 (consistency). `--top` has three incompatible stories. The help says "Cap table rows to top N items (default: all)". Validation rejects it on the leaderboard and every period report with "only applies to the projects and models commands". The changelog accepts it anyway. And the daily chart tells the user to "raise --top" on a command that rejects the flag.

Observed on the current source build (2026-10-29):

- `thermal --top 5` → `thermal: --top only applies to the projects and models commands`, exit `1`.
- `thermal daily --top 5` → the same message, exit `1`.
- `thermal daily --chart --top 100` → the same message, exit `1`.
- `thermal daily --chart` ends with `… and 84 more rows (raise --top to see them in the table)`; the table already shows all 96 active days, and the chart is what is capped.
- `thermal changelog --top 5` works and prints `Showing 5 of 25 releases`, exit `0`.
- `thermal changelog` prints `Showing 3 of 25 releases` with no hint that `--top` reveals more.
- `thermal projects --top 3` and `thermal models --top 3` work.
- Validation sources: `cmd/thermal/main.go:450-452`, `:471-473`, `:487-489`, `:522-524`, `:553-555`; the changelog case at `:496-502` never rejects `Top`.
- The chart footer is hard-coded at `internal/render/chart.go:68`.

The outcome of this goal is a **single `--top` contract**: one applicability table, one validation message that names every accepting command, one row-count policy shared by tables, charts, and JSON, and a changelog header that tells the user how to see more. No remediation hint may suggest a flag that the command rejects.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: cli-argument-parser-strictness
- **Sequence**: 250
- **Shape**: ship
- **Tier**: roadmap
- **Dependency Reason**: provides strict `--top` parsing and exit `2` usage errors
- **Memory Protocol**: required
- **Wave**: 3
- **Related implemented goals (do not regress)**: `chart-scale-order-and-top-contract` owns chart scaling and row-cap alignment; this goal resolves the remaining flag-scope contradiction and the invalid footer advice. `help-supported-tools-and-flags` owns the general flag reference; this goal supplies the final `--top` wording.

## Files to Touch
- `cmd/thermal/main.go`
- `cmd/thermal/args.go`
- `internal/render/chart.go`
- `internal/render/leaderboard.go`
- `internal/render/report.go`
- `internal/render/projects.go`
- `internal/render/models.go`
- `internal/changelog/render.go`
- `internal/render/help.go`
- `cmd/thermal/top_test.go`
- `scripts/simulated_user_gate.sh`
- `README.md`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- **Shared Understanding & Fact Sheet**: [`goals/cli-top-flag-contract/facts.md`](facts.md)
- **Execution Plan**: [`goals/cli-top-flag-contract/plan.md`](plan.md)

## Requirements

1. **One applicability table.**
   | Command | `--top` accepted | Meaning | Default |
   | --- | --- | --- | --- |
   | leaderboard (`thermal`) | yes | cap ranked tools | all |
   | `daily` | yes | cap period rows after ordering | all |
   | `weekly` | yes | cap week rows | all |
   | `monthly` | yes | cap month rows | all |
   | `projects` | yes | cap project rows | all |
   | `models` | yes | cap model rows | all |
   | `changelog` | yes | cap releases | 3 |
   | `stats` | no | — | — |
   | `mix` | no | — | — |
   | `trend` | no | — | — |
   | `audit` | no | — | — |
   | `share` | no | — | — |
   | `live` | no | — | — |
   | `statusline`/`prompt`/`status` | no | — | — |

2. **Validation and exit contract.**
   - Rejecting commands exit `2` (usage), not `1`, with:
     ```
     thermal: --top does not apply to the stats command
       it applies to: leaderboard, daily, weekly, monthly, projects, models, changelog
       run thermal stats --help for accepted flags
     ```
   - No message may say "only applies to the projects and models commands".
   - `--top -1` exits `2` with `--top cannot be negative`.
   - `--top 0` means "all rows" for every accepting command, including changelog (whose default is 3 when the flag is absent).
   - The parser must distinguish "flag absent" from "flag explicitly 0" for changelog; use `Present["top"]` or a sentinel.
   - `--top` is accepted in any position (parser goal) and follows the strict parser's value rules.

3. **Row-count policy across surfaces.**
   - The table, the chart, and the JSON `data` array use the same cap: `effectiveRows = min(totalRows, top)` when `top > 0`, else `totalRows`.
   - Charts keep their own maximum of 12 rows by default; `--top N` may raise the chart up to `min(N, totalRows)`. The chart never shows more rows than the table.
   - Totals always reflect the full window, not the capped rows. Add a footer/JSON field if needed to distinguish `returned` from `total`; do not change existing totals.
   - `--breakdown` rows are emitted only for the capped table rows.
   - `--order asc|desc` applies before the cap; `--top` never re-sorts.
   - Leaderboard `--top` caps after `--sort`.
   - Period reports `--top` caps after the default chronological order; define whether it keeps the newest or oldest rows (recommendation: the first N in the displayed order, so `--order asc` keeps the oldest).
   - `--json` `data` length equals the effective rows; document this and test it. If the JSON envelope has a row count, set it to the effective count and keep `totals` for the full window.

4. **Chart footer contract.**
   - If `--top` is accepted by the command and the chart omits rows:
     `… and N more rows (chart shows 12 of M; use --top M for all)`
   - If `--top` is accepted and already applied:
     `chart shows 12 of N rows (use --top N for all)`
   - If `--top` is not accepted (hypothetical future chart), the footer must not mention `--top`.
   - The omitted count in the table footer and the chart footer must agree for the same invocation.
   - The footer must name the chart cap (12) so the user understands why the table has more rows.
   - Unit tests assert the exact footer for `--top 0`, unset, 5, 12, 13, and 100.

5. **Changelog affordance.**
   - Header: `Current: … · Showing 3 of 25 releases · thermal changelog --top 25 for all`.
   - `thermal changelog --top 0` shows all releases (or the documented maximum).
   - `thermal changelog --top 5` shows `Showing 5 of 25`.
   - `--top` is listed in the changelog help page with its default.
   - The JSON changelog exposes the effective and total release counts (or a `showing`/`total` pair) without breaking existing keys.

6. **Help accuracy.**
   - Global `--top` description: `--top <N>  Cap rows: leaderboard, daily, weekly, monthly, projects, models, changelog (0 = all)`.
   - Per-command pages list `--top` only when accepted and state the default (all, or 3 for changelog).
   - The help test from `help-supported-tools-and-flags` must be extended with the applicability matrix.
   - No help text may promise `--top` for stats/mix/trend/audit/share/live/statusline.

7. **Tests and documentation.**
   - Acceptance matrix for all 14 commands.
   - Footer matrix (unset, 0, 5, 12, 13, 100) for daily and projects.
   - JSON `data` length and totals invariance.
   - Changelog default/0/5.
   - Leaderboard 5 rows and sort interaction.
   - Breakdown with top.
   - MIGRATION six-part entry: `--top` now works on leaderboard/daily/weekly/monthly; rejecting commands exit `2`; message changed; changelog default remains 3.
   - README and `project.jsonc`.

## Non-Goals
- Changing chart scaling, bar lengths, or the 12-row default cap (owned by `chart-scale-order-and-top-contract`).
- Adding pagination or `--page`.
- Changing the changelog parser or release ordering.
- Changing leaderboard sorting semantics.

## Acceptance Criteria
- AC-001: `thermal --top 5` exits `0` and prints exactly 5 tool rows (plus the rule/total rows).
- AC-002: `thermal daily --top 5` exits `0` and prints exactly 5 period rows; the `Total` row is unchanged from the uncapped run.
- AC-003: `thermal weekly --top 3` and `thermal monthly --top 3` each print 3 period rows; `thermal projects --top 3` and `thermal models --top 3` keep working.
- AC-004: `thermal changelog --top 5` exits `0` and prints `Showing 5 of 25 releases`; `thermal changelog --top 0` prints all releases or the documented maximum.
- AC-005: `thermal stats --top 5` exits `2`, the message lists all seven accepting commands, and it does not contain `only applies to the projects and models`.
- AC-006: `thermal daily --chart` footer contains `chart shows 12 of` and `--top` only with a valid value; it does not contain the old `raise --top to see them in the table` sentence.
- AC-007: `thermal daily --chart --top 5` exits `0`, the table and chart each show 5 rows, and the footers agree on the omitted count.
- AC-008: `thermal --top 5 --json | jq '.data|length'` returns `5`; `totals` equals the uncapped run's `totals`.
- AC-009: `thermal --top -1` exits `2` with `--top cannot be negative`.
- AC-010: `thermal --help` describes `--top` with the full applicability list; `thermal stats --help` does not list `--top`.
- AC-011: `go test -race ./cmd/thermal/... ./internal/render/...` exits `0`, including `TestTopApplicabilityMatrix`, `TestTopFooterMatrix`, `TestTopJSON`, `TestChangelogTop`, and `TestTopOrderInteraction`.
- AC-012: `./scripts/simulated_user_gate.sh` exits `0` and includes the `--top` section.

## Negative Cases & Failure Modes
| Invocation | Required result |
| --- | --- |
| `--top 5` | 5 leaderboard rows |
| `daily --top 5` | 5 period rows, totals unchanged |
| `daily --top 0` | all rows |
| `daily --top -1` | exit 2 |
| `stats --top 5` | exit 2, accurate list |
| `mix --top 5` | exit 2, accurate list |
| `live --top 5` | exit 2, accurate list |
| `changelog --top 5` | 5 releases |
| `changelog` | 3 releases + affordance |
| `changelog --top 0` | all releases |
| `daily --chart` | valid footer, no invalid advice |
| `daily --chart --top 5` | 5 rows table + chart |
| `daily --top 5 --json` | `data` length 5, totals full |
| `daily --top 5 --breakdown` | breakdown only for 5 rows |
| `daily --order asc --top 5` | oldest 5 rows |

## Metrics & Regression Guards
- Applicability matrix frozen in a test and in help; changing it requires docs.
- Footer tests: no `--top` mention for a command that rejects it.
- JSON/totals invariance: `totals` must be byte-identical with and without `--top`.
- Changelog header always states `Showing N of M` and the command to see more.
- Gate: simulated user gate asserts AC-001, AC-002, AC-004, AC-005, AC-006, and AC-008.

## Risks & Mitigations
- **Risk:** Enabling `--top` on more commands changes output for scripts. **Mitigation:** it is opt-in; no default behavior changes except the changelog header text.
- **Risk:** Capping JSON `data` while totals stay full could confuse consumers. **Mitigation:** document clearly; add `returned`/`total` counts if the envelope allows; keep totals semantics unchanged.
- **Risk:** Chart and table row counts drift again. **Mitigation:** one `effectiveRows` helper used by both; footer tests.
- **Risk:** `--top 0` vs unset for changelog is ambiguous. **Mitigation:** track flag presence in the parser; test both.
- **Risk:** The leaderboard already has a natural small row count; `--top` may be unnecessary. **Mitigation:** it is still consistent and cheap; the alternative is a documented exclusion, but the chart footer needs it.

## Memory Plan
- **Prime**: run `sila prime` at task start.
- **Recall**: run `sila recall "--top only applies to projects and models"` and `sila recall "chart footer raise top"`.
- **Writes**: record a `case` for the "single applicability table + effectiveRows helper" mechanism, grounded to `cmd/thermal/main.go` and `internal/render/chart.go`; record a `lesson` for invalid remediation hints.
- **Facts**: run `sila facts verify` before certification.
- **Handoff**: run `sila handoff "--top contract shipped" --open="JSON row-count fields"`.
- **Evidence**: embed a `sila/memory-writes/v1` block under `## Memory Writes`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`).
2. Implementation is committed with `(goals/cli-top-flag-contract/goal.md)` under a `fix(cli):` subject.
3. Goal is certified and `evidence.md` contains a standalone `Verdict: PASS` plus the applicability and footer matrices.
4. `sila goals --json` reports `status: "implemented"`, `verification_status: "verified"`, and `unverified_count: 0`, with a clean `git status`.
