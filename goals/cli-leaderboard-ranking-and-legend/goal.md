# Goal: Self-document the leaderboard ranking key, column abbreviations, and cost markers

## Goal Description
Priority: P1 (clarity). The default leaderboard is the first screen a user sees, and it does not explain itself. It is ranked by current streak, but the header only says "Token Warriors · tokens & spend". A reader sees Codex (1.5B tok, 11d streak) above CMD (8.5B tok, 1d streak) and concludes the token math is broken. `Strk`, `Best`, and `Days` are unexpanded, the `~`/`—`/`≥` cost markers are defined only inside the live `[?]` overlay, and the footer's tool counts look additive when they are not.

Observed on the current source build (2026-10-29):

- Default order: Codex (11d, 1.5B tok, ~$221.79), Muse (2d, 134.3M), CMD (1d, 8.5B, $120.83), OpenCode (1d, 11.0B).
- `thermal --sort tokens` reorders to OpenCode (11.0B), CMD (8.5B), Codex (1.5B), MiMoCode (1.2B).
- Header rows: `#  Tool  Strk  Best  Days  Tokens  Cost` and `#  Tool  Strk  Best  Days  Activity`.
- Footer: `Recorded cost: $241.93 from 5 tools. Estimated: ~$261.50 across 7 tools (~ prefix).`
- Footer: `1.3B tokens have no model recorded, so no price is applied to them.`
- Footer: `No pricing for: atria-dawn-preview`.
- Live `[?]` overlay documents: `Costs ~ = estimated; — = unpriced; ≥ = partial known; $0.00 = recorded zero`.
- The leaderboard has no window label and no ranking caption.
- Leaderboard JSON has `generatedAt`, `results`, `totals` (with `activeDays:150`, `period:null`), and `type`, but no ranking key or window.

The outcome of this goal is a **self-documenting default view**: a one-line caption states the ranking key and window, a legend defines every abbreviation and cost marker, the footer states cost provenance as an explicit sum and names unpriced models with their volume, and the JSON exposes the same ranking metadata without breaking existing keys.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 251
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required
- **Wave**: 3 (coordinate the final layout with `cli-terminal-width-adaptation`)
- **Related implemented goals (do not regress)**: `output-formatting-primitives` owns symbol semantics; `live-ticker-unit-and-cost-clarity` owns the live cost markers; `output-report-totals-reconciliation` owns numeric reconciliation. This goal owns the leaderboard's explanatory text and JSON metadata.

## Files to Touch
- `internal/render/leaderboard.go`
- `internal/render/leaderboard_test.go`
- `cmd/thermal/main.go`
- `cmd/thermal/json.go`
- `internal/render/help.go`
- `cmd/thermal/simulated_user_test.go`
- `scripts/simulated_user_gate.sh`
- `README.md`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- **Shared Understanding & Fact Sheet**: [`goals/cli-leaderboard-ranking-and-legend/facts.md`](facts.md)
- **Execution Plan**: [`goals/cli-leaderboard-ranking-and-legend/plan.md`](plan.md)

## Requirements

1. **Ranking caption.**
   - Immediately below the `THERMAL` title or above the first table, print one line:
     - default: `Ranked by current streak · all time · --sort tokens|cost to change`
     - `--sort tokens`: `Ranked by tokens · all time · --sort streak|cost to change`
     - `--sort cost`: `Ranked by cost (recorded + estimated) · all time · --sort streak|tokens to change`
   - The caption must state the window explicitly (`all time`), because the leaderboard ignores `--since`/`--until`/`--last`.
   - For Activity Hunters, state the actual key and warn that units differ:
     `Ranked by current streak · activity mixes steps and messages; compare within a unit`.
   - The caption is text, not color-only; it must survive `--no-color`.
   - The caption must not exceed the width contract; at narrow widths, shorten to `Ranked by streak · --sort tokens|cost`.

2. **Column labels and legend.**
   - At widths ≥80, prefer explicit headers: `#  Tool  Streak  Best  Active  Tokens  Cost` (or keep `Strk` if the width calculator requires it, but add the legend).
   - Add a legend footer:
     ```
     Legend: Strk = current streak · Best = longest streak · Days = active days
             tok = tokens · step = steps · msg = messages
             Cost: ~ estimated · — unpriced · ≥ partial · $0.00 recorded zero
     ```
   - The cost-marker wording must match the live `[?]` overlay exactly so there is one vocabulary.
   - The `Activity` column must state that steps and messages are not comparable:
     `Activity: units differ per row; read the suffix (step/msg)`.
   - The legend is omitted or condensed only under the width goal's documented drop order; it must never be dropped without a replacement caption.

3. **Cost provenance footer.**
   - Replace the current two-sentence footer with an explicit sum:
     ```
     Cost: $241.93 recorded (5 tools) + ~$261.50 estimated (7 tools) = $503.43 shown.
     ```
   - If the counts overlap (a tool with both recorded and estimated spend), say so:
     `5 recorded + 7 estimated tools (3 overlap)` or state `tools are counted per cost source`.
   - Unpriced models: replace `No pricing for: <name>` with
     `Unpriced models: atria-dawn-preview (328.7M tok)`; list up to three and append `+N more`.
     If none: omit the line.
   - Unattributed tokens: keep the existing sentence but make it a separate line:
     `Unattributed: 1.3B tok have no model recorded, so no price is applied.`
   - If `--no-estimate` is set, replace the estimated term with `estimation disabled` and keep the recorded total.
   - The footer must not imply that `recorded` and `estimated` tool counts are additive unless the text says they are.

4. **JSON parity.**
   - Add to the leaderboard envelope, without removing existing keys:
     ```json
     "window": "all-time",
     "ranking": {"key": "streak", "order": "desc"}
     ```
   - `ranking.key` must match the effective sort (`streak` when `--sort` is unset).
   - Add `"costBasis": {"recorded": 241.93, "estimated": 261.50, "shown": 503.43}` if the totals object does not already distinguish them; do not rename existing fields.
   - Include `unpricedModels: [{"model":"atria-dawn-preview","tokens":328700000}]` (top 10 or all) and `unattributedTokens` if absent.
   - JSON must remain valid and stable across widths.
   - Additive keys are documented in MIGRATION; no consumer should break.

5. **Consistency with the live help.**
   - Extract the cost-marker legend into one shared constant used by the leaderboard footer and the live `[?]` overlay.
   - Test that the two renderings are identical for the marker definitions.
   - Do not change the live overlay's other content.

6. **Tests and documentation.**
   - Caption tests for `--sort` unset/tokens/cost and the activity table.
   - Legend tests at 80 and 120; condensed form at narrow widths.
   - Footer arithmetic tests with recorded-only, estimated-only, mixed, overlapping-tool, `--no-estimate`, one unpriced model, many unpriced models, and none.
   - JSON tests for `window`, `ranking.key`, `costBasis`, `unpricedModels`.
   - A golden default view at 80 columns.
   - README leaderboard section with the legend; MIGRATION for the additive JSON keys and the footer wording; project.jsonc.
   - `scripts/simulated_user_gate.sh`: assert the caption, legend, footer arithmetic, and JSON fields.

## Non-Goals
- Changing ranking semantics or cost calculations.
- Changing the live TUI layout beyond reusing the marker legend.
- Changing the width/layout engine (owned by `cli-terminal-width-adaptation`).
- Adding interactive sorting in the TUI.
- Changing the leaderboard's all-time window.

## Acceptance Criteria
- AC-001: `thermal --no-color` contains `Ranked by current streak` and `all time`; it does not imply a tokens ranking.
- AC-002: `thermal --sort tokens --no-color` contains `Ranked by tokens`; `--sort cost` contains `Ranked by cost`.
- AC-003: The default output contains a legend defining `Strk`, `Best`, `Days` (or `Active`), and the `~`, `—`, `≥`, `$0.00` cost markers.
- AC-004: The activity table caption or legend states that steps and messages are not comparable.
- AC-005: The footer contains `recorded`, `estimated`, `shown`, and an explicit `+`/`=` arithmetic line; the numbers add up.
- AC-006: An unpriced model is rendered as `Unpriced models: <name> (<volume> tok)`; the string `No pricing for:` is absent.
- AC-007: With `--no-estimate`, the footer says `estimation disabled` and the shown total equals the recorded total.
- AC-008: `thermal --json | jq -e '.window=="all-time" and .ranking.key=="streak" and .ranking.order=="desc"'` exits `0`.
- AC-009: `thermal --sort tokens --json | jq -e '.ranking.key=="tokens"'` exits `0`.
- AC-010: `thermal --json | jq -e '.costBasis.recorded + .costBasis.estimated == .costBasis.shown'` exits `0` (within rounding), and `unpricedModels` is an array.
- AC-011: The leaderboard and live overlay use the same cost-marker definitions (a shared-constant test passes).
- AC-012: At `COLUMNS=80`, `thermal --no-color | awk 'length>80'` prints no lines (coordinated with the width goal).
- AC-013: `go test -race ./internal/render/... ./cmd/thermal/...` exits `0`, including `TestLeaderboardCaption`, `TestLeaderboardLegend`, `TestCostFooterArithmetic`, `TestUnpricedModels`, and `TestLeaderboardJSONRanking`.
- AC-014: `./scripts/simulated_user_gate.sh` exits `0` and includes the leaderboard self-documentation section.

## Negative Cases & Failure Modes
| Condition | Required result |
| --- | --- |
| Default sort | `Ranked by current streak` |
| `--sort tokens` | `Ranked by tokens` |
| `--sort cost` | `Ranked by cost (recorded + estimated)` |
| `--no-estimate` | `estimation disabled`, shown = recorded |
| Recorded only | no estimated term or `~$0.00`; explicit |
| Estimated only | `recorded $0.00` + estimated |
| Overlapping tools | counts labeled per cost source |
| One unpriced model | `Unpriced models: name (volume tok)` |
| Many unpriced models | first three + `+N more` |
| No unpriced models | line omitted |
| Activity table | unit warning present |
| JSON unset sort | `ranking.key=="streak"` |
| JSON `--sort tokens` | `ranking.key=="tokens"` |
| Narrow width | condensed caption/legend, never silently absent |

## Metrics & Regression Guards
- Caption/legend presence: 100% of default renders at ≥80 columns.
- Footer arithmetic: recorded + estimated = shown within $0.01; test with fixtures.
- Vocabulary parity: the leaderboard and live overlay marker strings are identical.
- JSON additive keys: a contract test locks `window`, `ranking`, `costBasis`, `unpricedModels`.
- Gate: simulated user gate asserts AC-001, AC-003, AC-005, AC-006, AC-008, and AC-009.

## Risks & Mitigations
- **Risk:** More explanatory lines make the first screen busier. **Mitigation:** one caption + a compact two-line legend; keep the table unchanged; the width goal handles narrow terminals.
- **Risk:** The estimated/recorded tool counts overlap and the new sentence is wrong. **Mitigation:** compute the overlap explicitly and test it; if indeterminate, say "tools are counted per cost source" instead of a number.
- **Risk:** Adding JSON keys breaks strict consumers. **Mitigation:** additive only; document in MIGRATION; keep existing keys and types.
- **Risk:** The legend duplicates the live overlay and drifts. **Mitigation:** one shared constant and a parity test.
- **Risk:** `--sort cost` mixes recorded and estimated costs in a misleading way. **Mitigation:** label the caption `recorded + estimated` and keep the `~` marker.

## Memory Plan
- **Prime**: run `sila prime` at task start.
- **Recall**: run `sila recall "leaderboard ranking streak tokens"` and `sila recall "cost marker legend"`.
- **Writes**: record a `case` for the "shared legend constant + ranking caption" mechanism, grounded to `internal/render/leaderboard.go`; record a `lesson` if the recorded/estimated tool counts overlap.
- **Facts**: run `sila facts verify` before certification.
- **Handoff**: run `sila handoff "leaderboard self-documentation shipped" --open="JSON cost basis"`.
- **Evidence**: embed a `sila/memory-writes/v1` block under `## Memory Writes`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`).
2. Implementation is committed with `(goals/cli-leaderboard-ranking-and-legend/goal.md)` under a `feat(render):` or `fix(render):` subject.
3. Goal is certified and `evidence.md` contains a standalone `Verdict: PASS` plus the caption/legend/footer/JSON captures.
4. `sila goals --json` reports `status: "implemented"`, `verification_status: "verified"`, and `unverified_count: 0`, with a clean `git status`.
