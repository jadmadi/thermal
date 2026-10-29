# Facts: Self-document the leaderboard ranking key, column abbreviations, and cost markers

## Observed Evidence (2026-10-29, current source build `thermal dev`)
### Default leaderboard (`thermal --no-color`)
```
  THERMAL  — Don't break the streak.

  ╭─ Token Warriors ────────────────────────────────────── tokens & spend ─╮
  │    #  Tool              Strk    Best    Days       Tokens         Cost │
  │   1.  Codex              11d     11d     63d     1.5B tok     ~$221.79 │
  │   2.  Muse                2d      2d      2d   134.3M tok      ~$11.37 │
  │   3.  CMD                 1d     18d     68d     8.5B tok      $120.83 │
  │   4.  OpenCode            1d     11d     41d    11.0B tok      ~$64.66 │
  ...
  ╭─ Activity Hunters ────────────────────── messages & steps ─╮
  │    #  Tool              Strk    Best    Days      Activity │
  │   1.  Agy                18d     44d     84d   315.1K step │
  │   2.  Droid               1d      1d      2d        14 msg │
  ╰────────────────────────────────────────────────────────────╯

  >> Agy is on fire with a 18-day streak!

  Recorded cost: $241.93 from 5 tools. Estimated: ~$261.50 across 7 tools (~ prefix).
  1.3B tokens have no model recorded, so no price is applied to them.
  No pricing for: atria-dawn-preview
```
### Sort comparison
- Default: Codex (1.5B, 11d) before CMD (8.5B, 1d) — ranked by streak.
- `--sort tokens`: OpenCode (11.0B), CMD (8.5B), Codex (1.5B), MiMoCode (1.2B).
- The header text is identical in both cases; only the row order changes.

### Cost markers
- `~$221.79` = estimated; `$120.83` = recorded; `—` = unpriced (seen in models/projects); `≥` = partial known; `$0.00` = recorded zero.
- The only legend is the live `[?]` overlay:
  `Costs ~ = estimated; — = unpriced; ≥ = partial known; $0.00 = recorded zero`.

### JSON
`thermal --json` top-level keys: `generatedAt`, `results`, `totals`, `type`.
`totals` includes `activeDays: 150` and `period: null`; there is no ranking key or window field.

## Root Cause
- `internal/render/leaderboard.go:156` and `:254`: fixed headers `{"#", "Tool", "Strk", "Best", "Days", "Tokens", "Cost"}` and `{"#", "Tool", "Strk", "Best", "Days", "Activity"}`. No caption, no legend.
- `internal/render/leaderboard.go:79` accepts `sortKey string` and sorts accordingly, but never renders the key.
- `cmd/thermal/main.go:740-750` passes `opts.Sort` to `RenderLeaderboard`; `validateReportFlags` defaults an empty sort to streak for the leaderboard, but the renderer does not state that.
- `internal/render/leaderboard.go` footer code emits `Recorded cost: … from N tools. Estimated: … across M tools` without explaining whether the tools overlap.
- `No pricing for: <model>` is emitted by the leaderboard footer without token volume or a marker definition.
- `cmd/thermal/json.go` (`leaderboardJSONPayload`) emits `results`/`totals`/`type`/`generatedAt`; no ranking metadata.
- `internal/render/statusline.go` and the live overlay contain separate cost-marker text; there is no shared constant.

## Architectural Invariants & Constraints
- The leaderboard is all-time; `--since`, `--until`, and `--last` are rejected by validation. The caption must say `all time`.
- Ranking must remain by streak by default and by tokens/cost when `--sort` is set; this goal does not change order.
- Cost arithmetic must remain unchanged; only the explanation changes.
- The `~`, `—`, `≥`, `$0.00` vocabulary is already established; reuse it verbatim.
- The default view must fit 80 columns after the width goal lands; before then, keep the caption/legend within 74–80 where possible.
- JSON additions must be additive; do not rename or retype existing keys.
- No color-only encoding: the legend and captions are plain text.

## File & Interface Contracts
- `internal/render/leaderboard.go`:
  - `RenderLeaderboard(results []thermal.ToolResult, weeks int, noColor bool, sortKey string, estimate bool) string`
  - add `rankingCaption(sortKey string) string`, `activityCaption() string`, `legendLines() []string`, `costFooter(...) string`.
- `internal/render/legend.go` (new): shared `CostMarkerLegend` constant used by the leaderboard and the live help overlay.
- `cmd/thermal/main.go`: pass a `LeaderboardMeta{SortKey, Window}` or extend the renderer signature.
- `cmd/thermal/json.go`: add `window`, `ranking`, `costBasis`, `unpricedModels`, `unattributedTokens` to the leaderboard payload.
- `internal/render/leaderboard_test.go`: caption, legend, footer, JSON tests.

## Text Contract
```
  Ranked by current streak · all time · --sort tokens|cost to change

  ╭─ Token Warriors ... ╮
  ...

  Legend: Strk = current streak · Best = longest streak · Days = active days
          tok = tokens · step = steps · msg = messages
          Cost: ~ estimated · — unpriced · ≥ partial · $0.00 recorded zero

  Cost: $241.93 recorded (5 tools) + ~$261.50 estimated (7 tools) = $503.43 shown.
  Unpriced models: atria-dawn-preview (328.7M tok)
  Unattributed: 1.3B tok have no model recorded, so no price is applied.
```
- Activity table caption: `Ranked by current streak · activity mixes steps and messages; compare within a unit`.
- `--no-estimate`: `Cost: $241.93 recorded (5 tools) · estimation disabled`.
- If no unpriced models: omit the `Unpriced models:` line.

## JSON Contract (additive)
```json
{
  "type": "leaderboard",
  "window": "all-time",
  "ranking": {"key": "streak", "order": "desc"},
  "costBasis": {"recorded": 241.93, "estimated": 261.50, "shown": 503.43},
  "unpricedModels": [{"model": "atria-dawn-preview", "tokens": 328700000}],
  "unattributedTokens": 1300000000,
  "results": [...],
  "totals": {...},
  "generatedAt": "..."
}
```
- `ranking.key` ∈ `streak|tokens|cost`, matching the effective sort.
- `costBasis.shown` equals recorded + estimated within $0.01.
- `unpricedModels` is always an array (possibly empty).
- Existing `results`/`totals`/`type`/`generatedAt` keys remain.

## Verification Contract
```bash
thermal --no-color > /tmp/lb
grep -q 'Ranked by current streak' /tmp/lb
grep -q 'all time' /tmp/lb
grep -q 'Strk = current streak' /tmp/lb
grep -q '~ estimated' /tmp/lb
grep -q 'recorded' /tmp/lb
grep -q 'estimated' /tmp/lb
grep -q 'shown' /tmp/lb
! grep -q 'No pricing for:' /tmp/lb
grep -q 'Unpriced models:' /tmp/lb

thermal --sort tokens --no-color | grep -q 'Ranked by tokens'
thermal --sort cost --no-color | grep -q 'Ranked by cost'

thermal --json | jq -e '.window=="all-time" and .ranking.key=="streak"' >/dev/null
thermal --sort tokens --json | jq -e '.ranking.key=="tokens"' >/dev/null
thermal --json | jq -e '(.costBasis.recorded + .costBasis.estimated - .costBasis.shown) | fabs < 0.02' >/dev/null

CGO_ENABLED=0 go test -race ./internal/render/... ./cmd/thermal/...
./scripts/simulated_user_gate.sh
```

## Regression Fixtures
- A leaderboard fixture with recorded-only, estimated-only, mixed, and overlapping tools.
- One, three, and five unpriced models.
- `--no-estimate` on/off.
- `--sort` unset/tokens/cost.
- An activity-only fixture with one step tool and one message tool.
- Golden captures at 80 and 120 columns.
- A JSON schema snapshot with the additive keys.

## Open Questions
- Should the caption go above the title or below it? Recommendation: below the `THERMAL` title and above the first table, so it reads as context.
- Should `Strk` be renamed to `Streak`? Recommendation: prefer explicit headers at ≥80 columns; keep the abbreviation plus legend at narrow widths.
- Are the recorded/estimated tool counts overlapping? The current data shows 5 recorded and 7 estimated; if any tool has both, state `tools counted per cost source` rather than an overlap number unless it is computed.
- Should `No pricing for:` be removed entirely? Recommendation: replace it with `Unpriced models:` including volume; keep the marker legend.
- Should JSON include the full unpriced list or top 10? Recommendation: top 10 plus `unpricedModelCount`; keep the payload bounded.
