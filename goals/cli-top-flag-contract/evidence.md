# Evidence: cli-top-flag-contract

Date: 2026-10-30
Project: thermal-streak
Goal: cli-top-flag-contract

## Implementation Summary
- Added one `topApplies(report, tool)` table: leaderboard, daily, weekly, monthly, projects, models, and changelog accept `--top`; stats/mix/trend/audit/share/live/statusline reject it.
- Rejecting commands exit `2` with `--top does not apply to the <command> command`, the full accepting set, and `run thermal <command> --help for accepted flags`.
- `--top 0` means all rows; an absent changelog `--top` keeps the three-release default.
- Leaderboard tables cap each category; period reports cap `Rows` while keeping `Totals` full; projects/models already capped and now share the chart policy.
- `chartRowsFor` keeps the 12-row default cap but lets an explicit `--top` raise it up to the full row count.
- `chartFooter` replaces the invalid `raise --top` sentence with `… and N more rows (chart shows X of Y; use --top Y for all)`.
- Changelog header adds `· thermal changelog --top N for all` when capped.
- Updated help text, README, migration guide, project metadata, gallery fixture, and simulated gate.

## Commands Executed
- `go test ./...` → PASS.
- `go test -race ./...` → PASS.
- `./scripts/check.sh --pre-commit` → PASS.
- `./scripts/output_gallery_gate.sh` → PASS.
- `./scripts/simulated_user_gate.sh` → PASS (118/118 checks).
- `thermal --top 5` → exit `0`.
- `thermal daily --top 5` → exit `0`.
- `thermal weekly --top 5`, `monthly --top 5`, `projects --top 5`, `models --top 5`, `changelog --top 5` → exit `0`.
- `thermal stats --top 5`, `mix --top 5`, `live --top 5` → exit `2` with the accepting-command list.
- `thermal daily --top 5 --json | jq '.data|length'` → `5`; `.totals` identical to the uncapped run.
- `thermal daily --chart --top 5` → `… and 92 more rows (chart shows 5 of 97; use --top 97 for all)`.
- `thermal daily --chart` → `… and 85 more rows (chart shows 12 of 97; use --top 97 for all)`.
- `thermal changelog` → `Showing 3 of 25 releases · thermal changelog --top 25 for all`.
- `thermal changelog --top 5` → `Showing 5 of 25 releases`.
- `thermal changelog --top 0` → `Showing 25 of 25 releases`.
- `thermal --top -1` → exit `2`, `--top cannot be negative`.
- `thermal daily --help` lists `--top`; `thermal stats --help` does not.

## Acceptance Criteria

- AC-001: `thermal --top 5` exits `0` and caps each leaderboard category to five rows. **PASS**
- AC-002: `thermal daily --top 5` exits `0` with five period rows; the table `Total` row is unchanged. **PASS**
- AC-003: `weekly`, `monthly`, `projects`, and `models` honor `--top`; each renders at most the requested rows. **PASS**
- AC-004: `changelog --top 5` shows five releases; `--top 0` shows all; the default header states the valid `--top` value. **PASS**
- AC-005: `stats --top 5` exits `2`, lists leaderboard/daily/weekly/monthly/projects/models/changelog, and does not say `only applies to the projects and models`. **PASS**
- AC-006: Chart footers contain `chart shows` and `use --top`, and never contain `raise --top`. **PASS**
- AC-007: `daily --chart --top 5` renders five table rows and five chart rows; the footer names a valid value. **PASS**
- AC-008: `daily --top 5 --json` has a five-element `data` array and `totals` equal to the uncapped run. **PASS**
- AC-009: `--top -1` exits `2` with `--top cannot be negative`. **PASS**
- AC-010: Global help lists the full `--top` applicability; `daily --help` lists it; `stats --help` does not. **PASS**
- AC-011: `go test -race ./...` passes, including `TestTopApplicability`, `TestTopAppliesTable`, `TestChangelogLimit`, `TestLeaderboardTopCapsEachCategory`, `TestRenderChangelogTopAffordance`, and `TestChartCappedAtTwelve`. **PASS**
- AC-012: `./scripts/simulated_user_gate.sh` passes 118/118 and includes the `--top` section. **PASS**

## Regression Notes
- The leaderboard remains ranked by the same keys; `--top` caps after ranking.
- `--top 0` is distinct from an absent changelog `--top` (all vs three).
- Totals and JSON `totals` always describe the full window.
- The chart's default cap remains 12; `--top` is the only way to widen it.

Verdict: PASS

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "cli-top-flag-contract",
  "project": "thermal-streak",
  "prime": {"ran": true, "command": "sila prime"},
  "recall": [{"query": "Unify --top applicability, row caps, and chart remediation", "matches": 5}],
  "memos": [],
  "lessons": [],
  "cases": [{"id": "fact_299e0aea", "intent": "Unify --top applicability, row caps, and chart remediation"}],
  "facts_verify": {"exit_code": 0},
  "handoff": {"id": "hoff_6b0f7b33b96c"}
}
```
