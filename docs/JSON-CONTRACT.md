# Thermal JSON Contract

Every `--json` surface is an object. The table below is the contract used by
`cmd/thermal/json_contract_test.go`; payload keys are lower camelCase.

## Envelopes

| Command | `type` | timestamp | payload keys |
|---|---|---|---|
| `thermal --json` | `leaderboard` | `generatedAt` (RFC3339 UTC) | `results`, `totals` |
| `thermal <tool> --json` | `tool` | `generatedAt` | `tool`, `dataPath`, `summary`, `daily` |
| `daily` / `weekly` / `monthly` | `daily` / `weekly` / `monthly` | `generatedAt` | `data`, `totals`, `window`, `costBasis` |
| `projects` | `projects` | `generatedAt` | `data`, `totals`, `costBasis` |
| `models` | `models` | `generatedAt` | `data`, `totals`, `costBasis` (`estimated`) |
| `mix` | `mix` | `generatedAt` | `series`, `data`, metric-specific switching fields, `costBasis` |
| `stats` | `stats` | `generatedAt` | `metric`, `activeDays`, `costedDays`, distribution/percentiles, `composition`, `costBasis` |
| `trend` | `trend` | `generatedAt` | trend series/projection fields |
| `audit` | `audit` | `timestamp` (RFC3339 local) | score/findings/recommendations |
| `changelog` | `changelog` | `generatedAt` | `currentVersion`, `totalReleases`, `releases` |
| `license` | `license` | `generatedAt` | license/commercial fields |
| `live --json` | `live` | `timestamp` (RFC3339 local) | snapshot fields |
| `live --json --stream` | `snapshot` / `event` | `timestamp` | `snapshot` / `event` |
| `share` | `share` | `generatedAt` | `url`, `token`, `snapshot` |
| `version --json` | `version` | `generatedAt` | `version`, `commit`, `date` |

`audit` and `live` keep their existing `timestamp` fields because those are the
observation times; every other envelope uses `generatedAt`. `live --json
--stream` is NDJSON, one object per line.

## Precision

`cmd/thermal/json.go` rounds every serialized floating-point value to **four
decimal places**. This is the single JSON-only boundary rule; internal values
keep full precision for calculations. Timestamps and strings are unchanged.

Examples:

| Field | Raw value | Serialized |
|---|---:|---:|
| `todayCost` | `1.9774259520000008` | `1.9774` |
| `todayCacheHit` | `96.77737513675537` | `96.7774` |
| `share` | `0.5038565852808624` | `0.5039` |
| `mean` | `396960162.5520833` | `396960162.5521` |

## Leaderboard rows

`results[]` is a `thermal.ToolResult` with exactly these top-level keys:

```text
activeDays, currentStreak, daily, dataPath, estimatedCost, longestStreak,
name, summary, tool, totalActivity
```

There are no uppercase aliases (`Tool`, `Name`, `Summary`) and no second copy of
the summary. `totals` contains `tools`, `activeDays`, `totalTokens`,
`storedCost`, `estimatedCost`, `cost`, and `costBasis`.

## Report rows

- `daily`/`weekly`/`monthly` `data[]`: `period`, `models`, token fields,
  `turns`, `activeDays`, `storedCost`, `estimatedCost`, `cost`,
  `missingPricing`, `unattributedTokens`.
- `projects` `data[]`: `project`, `tools`, token fields, `turns`,
  `activeDays`, `firstDay`, `lastDay`, `storedCost`, `estimatedCost`, `cost`,
  `missingPricing`, `unattributedTokens`, `models`.
- `models` `data[]`: `model`, `tools`, token fields, `totalTokens`,
  `unattributedTokens` (totals), `activeDays`, `firstDay`, `lastDay`,
  `estimatedCost`, `missingPricing`, `reasoningEffort`.
- `stats.composition`: `uncachedInput`, `output`, `reasoning`, `cacheRead`,
  `cacheWrite`, `total`, `classifiedTotal`, `unclassified`, `cacheHitRate`.
- `share.snapshot`: `scope`, `window`, `activeDays`, `totalTokens`,
  `recordedCost`, `estimatedCost`, `totalCost`, `coverage`, plus the URL/token
  envelope fields.

All displayed text values have a JSON counterpart; no displayed field is
`null`. Warnings are stderr-only and never appear in JSON.
