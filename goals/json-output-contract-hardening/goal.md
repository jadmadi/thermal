# Goal: Standardize JSON envelopes, row schemas, precision, and alias fields

## Goal Description
Priority: P1 (machine contract). The gallery's JSON table shows four different envelope shapes: leaderboard has `generatedAt`/`results` but no `type` or totals; reports have `type`/`data`/`totals`/`generatedAt`; mix/stats/trend have `type` plus metric-specific fields; audit/license/changelog/live have no `generatedAt` (audit/live use `timestamp`). A leaderboard result exposes both `Tool`/`tool`, `Name`/`name`, and `Summary`/`summary`, inflating a 140KB payload. Costs leak raw floats such as `1.9774259520000008`, `96.77737513675537`, and `396960162.5520833`. Standardize the envelope and row schemas, round noisy numbers, remove aliases, and test the contract.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: output-report-totals-reconciliation
- **Sequence**: 202
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/thermal/types.go`
- `internal/thermal/live.go`
- `internal/thermal/stats.go`
- `internal/thermal/models.go`
- `internal/thermal/projects.go`
- `internal/share/share.go`
- `internal/thermal/*_test.go`
- `cmd/thermal/main.go`
- `cmd/thermal/simulated_user_test.go`
- `docs/JSON-CONTRACT.md`
- `docs/MIGRATION.md`
- `README.md`
- `project.jsonc`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)


## Requirements
1. **Envelope contract**: Every `--json` command returns an object with `type` (`leaderboard`, `daily`, `weekly`, `monthly`, `projects`, `models`, `mix`, `stats`, `trend`, `audit`, `changelog`, `license`, `live`, `share`, `version`), a timestamp (`generatedAt` or `timestamp` with a documented rule), and an explicit payload key (prefer `data`/`totals` for reports; metric-specific keys for analytics is allowed if documented). Leaderboard must gain `type` and a `totals` object or a documented `results`-only exception. Add a schema version if the contract must evolve.
2. **Remove aliases**: Delete the duplicate `Tool`/`tool`, `Name`/`name`, and `Summary`/`summary` fields from leaderboard results (and any equivalent duplicates elsewhere), or move them behind an explicit deprecated compatibility object. The canonical keys are lower camelCase. Update `jq` examples and tests; assert `json.Marshal` has exactly one of each key and `results[0]|keys` contains no uppercase aliases.
3. **Rounding policy**: Add JSON-only rounding helpers. Costs: four decimals for sub-cent, two decimals for >=0.01 (or a documented fixed precision such as 4 decimals for all); percentages: one decimal; shares: four decimals; means: four decimals or integer token values; timestamps unchanged. No serialized number may have more than the documented decimal places. `todayCost` must not serialize as `1.9774259520000008`; `todayCacheHit` not as `96.77737513675537`; mix `share` not as `0.5038565852808624`; stats `mean` not as `396960162.5520833`.
4. **Row schemas**: Document and test the row schema for leaderboard results, daily rows, project rows, model rows, live recent events/projects/models, and share snapshot. Each row must expose the fields consumers need: cost kind/estimated flags, classified/unclassified tokens, window/scope, and no duplicate aliases. Add `unattributedTokens` to models/stats JSON; add `cost`/`scope`/`window` to share JSON.
5. **Envelope consistency**: `audit`, `changelog`, `license`, and `live` must either include `generatedAt` or document why they use `timestamp`/no timestamp. `leaderboard` must not be the only report-like object without `type`; `audit` must not be the only diagnostic without a generation timestamp. Prefer one shared envelope helper in `cmd/thermal` or `internal/thermal`.
6. **Share JSON**: `share --json` must include `scope`, `window`, `activeDays`, `recordedCost`, `estimatedCost`, `totalCost`, and the URL/token; no displayed text field may be `null` while the text card prints a value. `cost:null`/`scope:null` in the current snapshot are contract bugs.
7. **Backward compatibility**: If aliases must remain for one release, add a `deprecated` object or a migration note with a removal version; do not silently keep two full copies. Add a six-part migration entry for any renamed/removed key and for the rounding change.
8. **Contract tests**: Add `TestJSONEnvelopeContract`, `TestJSONNoAliasKeys`, `TestJSONPrecision`, and `TestJSONShareFields`. Decode every command's JSON and assert envelope keys, row keys, numeric decimal places, and no null displayed fields. Add a `docs/JSON-CONTRACT.md` table used by the tests.

## Done Condition
1. Every `--json` surface has a documented envelope; leaderboard has `type` and a totals contract; audit/license/changelog/live timestamps are documented or added.
2. Leaderboard results contain one canonical key for tool/name/summary; `jq '.results[0]|keys'` has no uppercase aliases; payload size drops accordingly.
3. No JSON monetary/percentage/share/mean value exceeds the documented decimal precision; the gallery examples are replaced with stable values.
4. Models/stats expose unclassified tokens; share exposes cost/scope/window/active days; no displayed field is null.
5. Contract tests, full race suite, build, and simulated gate pass; docs/JSON-CONTRACT.md, docs/MIGRATION.md, README, and project.jsonc are updated; the attribution commit ends in `(goals/json-output-contract-hardening/goal.md)`.

## Scope Boundary
JSON shape/precision/schema only. Text formatting is owned by output-formatting-primitives; totals/reconciliation by output-report-totals-reconciliation; share URL layout by statusline-and-share-scope-contract. Do not change aggregation values except rounding at serialization.
