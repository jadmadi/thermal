# Facts: Unify output formatting semantics across every Thermal surface

## Observed Evidence
- `docs/OUTPUT-GALLERY.md:185` uses `~$20381.54` for estimated cost; `:366` uses `~$20381.71`; `:583` uses `$20384.60`; `:659` uses `$20606.08`; these are the same total family rendered with different markers.
- `docs/OUTPUT-GALLERY.md:261` prints `~$20596.53` for a total whose footer is `$214.82 recorded + ~$20381.71 estimated`; the tilde does not distinguish the recorded portion.
- `docs/OUTPUT-GALLERY.md:762` prints `Daily Mean: ~97.5M` for a token mean; `:1100` prints `Context Tax: ~45.4K tokens / turn`; `:1078` prints `Spend Profile: ~$20596.53`; the same symbol means approximate volume, approximate token tax, and mixed money.
- `docs/OUTPUT-GALLERY.md:460` and `:600` render project/model tool lists as `OpenCode,Devin +8` and `CMD,OpenCode,ZCode` while leaderboard/live use `·`/`×` conventions.
- `docs/OUTPUT-GALLERY.md:261` prints `Window: [96 days · all tools]` for active days; `:762` prints `Window: [390 Days ...]` for calendar days; `:988` prints `Active days 80` for the cost metric while `:702` prints 96 for tokens.
- `docs/OUTPUT-GALLERY.md:1326` shows `Speed: 56.7K · 0.383 step…` and `:1325` shows `Burned: +1.7M tok (+60 s)`; live's unit abbreviations are not the same as report units.
- `internal/render/report.go:242` defines one local `formatCost`; `internal/render/leaderboard.go:213-221` defines separate cost-prefix logic; projects/models/analytics have their own formatting paths.

## Interpretation
The inconsistencies are presentation-policy bugs, not data bugs. A single shared set of helpers plus a documented legend will remove most of them and give the reconciliation/JSON goals a stable vocabulary.

## Ownership and Constraints
`internal/render` owns text formatting; `internal/thermal/format.go` owns count/path formatting; tests in `internal/render` and `internal/thermal` own the contract. Preserve read-only ingestion, stdout/stderr separation, `--no-color`, and sub-10ms static commands. No new dependencies.

## Verification Contract
```bash
go test -race ./internal/render ./internal/thermal -run 'Format|Cost|Window|Tool|Percent'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```

## Open Questions
- Should `≈` be acceptable in `--plain`/ASCII-only statusline output, or should those modes spell `approx.`?
- Should the mixed-total legend live in every table footer or in one README legend plus a single footer? The goal chooses explicit two-part totals; tests will enforce it.
