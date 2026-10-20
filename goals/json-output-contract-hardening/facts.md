# Facts: Standardize JSON envelopes, row schemas, precision, and alias fields

## Observed Evidence
- `docs/OUTPUT-GALLERY.md:1379` `todayCost: 1.9774259520000008`; `:1381` `todayCacheHit: 96.77737513675537`; `:1404` project cost `1.3979846039999992`; `:1856` mix share `0.5038565852808624`.
- `docs/OUTPUT-GALLERY.md:1354-1366` JSON envelope table: leaderboard has `generatedAt`/`results` and no `type`; reports have `type`/`data`/`totals`/`generatedAt`; mix/stats/trend differ; audit/changelog/license/live lack `generatedAt`.
- Command evidence: `/tmp/thermal-gallery --json --offline | jq '.results[0] | keys'` returns `Name`, `Summary`, `Tool`, `activeDays`, `currentStreak`, `daily`, `dataPath`, `estimatedCost`, `longestStreak`, `name`, `summary`, `tool`, `totalActivity`.
- `share --json | jq 'keys'` returns `snapshot`, `token`, `url`; `snapshot.cost` and `snapshot.scope` are null while the text card prints `Spend Profile` and `Scope: All`.
- `stats --json` exposes `mean`/weekday means as raw floats; `mix --json` exposes `concentration`/`share` as raw floats; models/stats lack the unclassified token field while daily/projects have `unattributedTokens`.

## Interpretation
The JSON is valid but not a stable contract. Consumers must branch on capitalized aliases, re-round floats, and guess why some envelopes have timestamps/type and others do not. A single documented envelope and serialization policy is needed before more surfaces are added.

## Ownership and Constraints
`internal/thermal` owns payload structs and tags; `cmd/thermal` owns envelopes/`generatedAt`; `internal/share` owns share output. Preserve stdout-only JSON, stderr diagnostics, and no new dependencies. No aggregation changes beyond rounding.

## Verification Contract
```bash
go test -race ./cmd/thermal ./internal/thermal ./internal/share -run 'JSON|Envelope|Precision|Alias'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```

## Open Questions
- Is the `--json` schema covered by a compatibility promise? If yes, aliases need a deprecation window; if no, remove them now and document the break.
- Should JSON costs use a fixed four decimals for all values, or mirror the text 2/4 policy? The goal requires one documented choice.
