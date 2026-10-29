# Evidence: cli-leaderboard-ranking-and-legend

Date: 2026-10-30
Project: thermal-streak
Goal: cli-leaderboard-ranking-and-legend

## Implementation Summary
- Added `Ranked by <streak|tokens|cost> · all time` caption driven by `--sort`.
- Added a legend expanding `Strk` (current streak), `Best` (longest streak), and `Days` (active days), plus the shared cost markers `$`, `~$`, `≥$`, `—`, `≈`, and `$0.00`.
- Added the activity caption that steps/messages are not comparable to token totals.
- Replaced the recorded/estimated prose with one arithmetic footer: `Cost: $X recorded + ~$Y estimated = ~$Z shown`; `--no-estimate` prints `estimation disabled` and makes shown equal recorded.
- Replaced `No pricing for:` with `Unpriced models: <name> (<volume> tok)`.
- Added shared cost-marker constants and `CostLegendText()` used by both the leaderboard and the live help overlay.
- Added `window`, `ranking`, `costBasis`, and `unpricedModels` to the leaderboard JSON envelope; the cost basis is rounded once and sums to `shown`.
- Updated README, migration guide, project metadata, gallery, and simulated gate.

## Commands Executed
- `go test ./...` → PASS.
- `go test -race ./internal/render/... ./cmd/thermal/...` → PASS.
- `./scripts/output_gallery_gate.sh` → PASS.
- `./scripts/simulated_user_gate.sh` → PASS (146/146 checks).
- `thermal --no-color` → `Ranked by current streak · all time`; legend/cost/activity lines present.
- `thermal --sort tokens --no-color` → `Ranked by tokens · all time`.
- `thermal --sort cost --no-color` → `Ranked by cost · all time`.
- `thermal --no-color` footer → `Cost: $243.90 recorded + ~$261.50 estimated = ~$505.39 shown`.
- `thermal --no-estimate --no-color` → `Cost: $243.90 recorded + estimation disabled = $243.90 shown`.
- `thermal --no-color` unpriced line → `Unpriced models: atria-dawn-preview (2.5K tok)`; no `No pricing for:`.
- `thermal --json` → `.window=="all-time"`, `.ranking.key=="streak"`, `.ranking.order=="desc"`, `.costBasis` sums, `.unpricedModels` array.
- `thermal --sort tokens --json` → `.ranking.key=="tokens"`.
- `COLUMNS=80 thermal --no-color` → no line over 80.

## Acceptance Criteria

- AC-001: Default text contains `Ranked by current streak` and `all time`; it does not imply a token ranking. **PASS**
- AC-002: `--sort tokens` contains `Ranked by tokens`; `--sort cost` contains `Ranked by cost`. **PASS**
- AC-003: The legend defines `Strk`, `Best`, `Days`, and `~$`, `—`, `≥$`, `$0.00`. **PASS**
- AC-004: The activity caption states steps/messages are not comparable to token totals. **PASS**
- AC-005: The footer contains `recorded`, `estimated`, `shown`, `+`, `=`; recorded + estimated equals shown. **PASS**
- AC-006: An unpriced model renders `Unpriced models: <name> (<volume> tok)`; `No pricing for:` is absent. **PASS**
- AC-007: `--no-estimate` says `estimation disabled` and shown equals recorded. **PASS**
- AC-008: JSON `.window=="all-time"`, `.ranking.key=="streak"`, `.ranking.order=="desc"`. **PASS**
- AC-009: `--sort tokens --json` has `.ranking.key=="tokens"`. **PASS**
- AC-010: `.costBasis.recorded + .estimated == .shown` within 4-decimal rounding; `.unpricedModels` is an array. **PASS**
- AC-011: `TestCostMarkersShared` and `TestLiveHelpUsesSharedCostLegend` share one legend between the leaderboard and live overlay. **PASS**
- AC-012: At `COLUMNS=80`, no leaderboard line exceeds 80 columns. **PASS**
- AC-013: `go test -race ./internal/render/... ./cmd/thermal/...` passes, including `TestLeaderboardCaption`, `TestLeaderboardLegend`, `TestCostFooterArithmetic`, `TestUnpricedModels`, and `TestLeaderboardJSONRanking`. **PASS**
- AC-014: `./scripts/simulated_user_gate.sh` passes 146/146 and includes the leaderboard self-documentation section. **PASS**

## Regression Notes
- Table sorting behavior is unchanged; only the caption/metadata are new.
- JSON additions are additive; `results` and `totals` remain.
- At ≥96 columns the established card layout is preserved.
- `--no-estimate` disables estimated costs everywhere, including the JSON cost basis.

Verdict: PASS

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "cli-leaderboard-ranking-and-legend",
  "project": "thermal-streak",
  "prime": {"ran": true, "command": "sila prime"},
  "recall": [{"query": "Make the leaderboard state its ranking, legend, and cost arithmetic", "matches": 5}],
  "memos": [],
  "lessons": [],
  "cases": [{"id": "fact_1bc3edd9", "intent": "Make the leaderboard state its ranking, legend, and cost arithmetic"}],
  "facts_verify": {"exit_code": 0},
  "handoff": {"id": "hoff_6b0f7b33b96c"}
}
```
