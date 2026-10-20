# Facts: Reconcile report totals, unclassified tokens, cost bases, windows, and share scope

## Observed Evidence
- `docs/OUTPUT-GALLERY.md:261` daily: `38.0B tok`, `~$20596.53`; `:659` models: `36.7B`, `$20606.08`; `:581` projects: `38.0B`, `$20599.42`; `:969` mix cost: `$20606.08`; `:988` stats cost: `$20596.53`.
- `docs/OUTPUT-GALLERY.md:366` daily footer: `$214.82 recorded + ~$20381.71 estimated`; `:185` leaderboard: `$214.82 recorded + ~$20381.54 estimated`; `:583` projects: `$214.82 + ~$20384.60`.
- `docs/OUTPUT-GALLERY.md:988` stats cost: `Active days 80`; `:702` token stats: `Active days 96`; `:261` daily active window: 96 days.
- `docs/OUTPUT-GALLERY.md:1076` share: `Active Days: 145`; `:1078` share spend `~$20596.53`; daily active days 96 at `:261`.
- Current JSON checks: leaderboard sum of `.estimatedCost` 20381.542956675003 vs daily `.totals.estimatedCost` 20381.709905725005; projects `.totals.cost` 20600.191278 vs daily 20597.297059; models `.totals.totalTokens` 36744128150 vs daily 38096937893.
- `internal/thermal/stats.go:129` filters `value <= 0` after building `byDay`, so cost days with zero/unpriced cost are excluded from `Days`; `internal/thermal/stats.go:185-204` builds composition from classified categories only.

## Interpretation
The gaps are structural aggregation-path differences, not mostly live drift. The 1.3B gap is unclassified volume and must be explicit. The cost gaps come from per-tool/per-project/per-model estimators; either unify them or disclose the basis. The stats cost day count is a filtering/label bug. The share card mixes all-time streak/activity fields with a daily-window spend/volume.

## Ownership and Constraints
`internal/thermal` owns canonical totals and basis metadata; `internal/render` prints the reconciliation; `cmd/thermal` wires JSON; share owns its snapshot. Preserve read-only ingestion, disjoint token types, activity-only exclusion, and sub-10ms static commands. No new dependencies.

## Verification Contract
```bash
go test -race ./internal/thermal ./internal/render -run 'Reconcile|Unclassified|StatsCost|ProjectCost|ModelCost|ShareScope'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```

## Open Questions
- Which estimator is canonical for estimated cost: per-tool summary, per-day model map, or per-project/model map? The goal requires one; the implementation must pick and document it.
- Should stats cost include zero-cost days in the distribution or expose `Costed days`? The goal accepts either if disclosed; tests must pin the choice.
