# Plan: Self-document the leaderboard ranking key, column abbreviations, and cost markers

## Execution Steps
- [x] **Phase 1 — Ranking caption and window**: Add `rankingCaption(sortKey)` and the all-time window line; render it above the first table; add the activity caption/unit warning. Acceptance: `TestLeaderboardCaption` passes for unset/tokens/cost and the activity table.
  - Default: `Ranked by current streak · all time · --sort tokens|cost to change`.
  - `--sort tokens`: `Ranked by tokens · all time · …`.
  - `--sort cost`: `Ranked by cost (recorded + estimated) · all time · …`.
  - Activity: `Ranked by current streak · activity mixes steps and messages; compare within a unit`.
- [x] **Phase 2 — Legend and cost footer**: Extract the shared cost-marker constant, add the legend, replace the footer with the explicit sum, and add the unpriced-model/unattributed lines. Acceptance: `TestLeaderboardLegend`, `TestCostFooterArithmetic`, `TestUnpricedModels`, and `TestLegendParityWithLive` pass.
  - Recorded-only, estimated-only, mixed, overlapping, `--no-estimate`.
  - One/many/none unpriced models.
  - No `No pricing for:` string.
- [x] **Phase 3 — JSON metadata**: Add `window`, `ranking`, `costBasis`, `unpricedModels`, and `unattributedTokens` to the leaderboard envelope. Acceptance: `TestLeaderboardJSONRanking` and the JSON schema snapshot pass; existing keys unchanged.
  - `ranking.key` matches the effective sort.
  - `costBasis.shown` equals recorded + estimated within rounding.
  - Top 10 unpriced models + count.
- [x] **Phase 4 — Width, gate, docs, and closure**: Coordinate the final layout with `cli-terminal-width-adaptation`, add the simulated-user-gate section, README/MIGRATION/project.jsonc, and close the goal.
  - `CGO_ENABLED=0 go test -race ./...` and `./scripts/simulated_user_gate.sh` pass.
  - Commit with `feat(render): document leaderboard ranking and cost markers (goals/cli-leaderboard-ranking-and-legend/goal.md)`.
  - Certify with `sila goals certify cli-leaderboard-ranking-and-legend --evidence=goals/cli-leaderboard-ranking-and-legend/evidence.md`.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | default leaderboard | streak caption + all time |
| AC-002 | `--sort tokens`/`cost` | matching caption |
| AC-003 | default | legend present |
| AC-004 | activity table | unit warning |
| AC-005 | footer | recorded + estimated = shown |
| AC-006 | unpriced model | new line, no `No pricing for:` |
| AC-007 | `--no-estimate` | estimation disabled |
| AC-008 | `--json` | `window`, `ranking.key` |
| AC-009 | `--sort tokens --json` | `ranking.key=="tokens"` |
| AC-010 | `--json` cost basis | arithmetic holds |
| AC-011 | legend parity | shared constant |
| AC-012 | 80-col check | no long lines |
| AC-013 | `go test -race ...` | exit 0 |
| AC-014 | gate | exit 0 |

## Rollback
- If the extra lines are too noisy, keep the caption and the cost footer, and move the legend to `thermal --help leaderboard` with a one-line pointer. Do not remove the ranking caption or the unpriced-model volume.
- JSON additions are additive; reverting them is safe but should be accompanied by a MIGRATION note.
- Never restore `No pricing for:` without the volume and marker definition.

## Phase 5 — Attribution & Registry Closure
- [x] **Phase 5 — Attribution & Registry Closure**: (1) check all completed tasks in `goals/cli-leaderboard-ranking-and-legend/plan.md`; (2) commit with `(goals/cli-leaderboard-ranking-and-legend/goal.md)` using a `feat(render):` subject; (3) certify via `sila goals certify cli-leaderboard-ranking-and-legend --evidence=goals/cli-leaderboard-ranking-and-legend/evidence.md` (with `Verdict: PASS`); (4) reconcile via `sila goals` and commit the updated `goals/README.md` and certification receipts; verify `sila goals --json` reports `unverified_count: 0` for this goal.
