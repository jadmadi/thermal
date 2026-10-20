# Facts: Fix chart scaling, ordering, row counts, and no-op --chart flags

## Observed Evidence
- `docs/OUTPUT-GALLERY.md:881-890`: `Projects by tokens  bar = share of the largest row, 3.9B`; `tree.waqf.app 3.9B` and `mahak-bench 7.1B` both render 48 filled cells.
- `internal/render/chart.go:137` hard-codes `rep.Rows[0].Tokens` in the caption; `:148` uses `maxVal := float64(rep.Rows[0].Tokens)`; `:181` repeats it for models.
- `docs/OUTPUT-GALLERY.md:870` table says `and 106 more`; `:892` chart says `and 104 more`; models table/chart show `44` vs `42` at `:921/:948`.
- `docs/OUTPUT-GALLERY.md:950` mix `--chart`, `:976` stats `--chart`, `:1028` trend `--chart` produce no appended chart, while daily `:835` and projects `:881` do.

## Interpretation
The caption and bars currently describe different row sets. The max bug is independent of sorting, but cost sorting exposes it. The no-op flag is a contract bug: users cannot tell whether the chart is missing because of width, data, or implementation.

## Ownership and Constraints
`internal/render/chart.go` owns bars/captions; `cmd/thermal` owns flag validation; help owns applicability. Preserve chart conventions in AGENTS.md: no color-only encoding, named scale, width-aware, capped with remainder. No new dependencies.

## Verification Contract
```bash
go test -race ./internal/render -run 'Chart|Project|Model'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```

## Open Questions
- For analytics commands, is a chart additive value or is the existing internal bar/sparkline enough? If not, reject the flag; do not leave a no-op.
- Should chart rows follow table sort or bar metric? The goal accepts either if the caption is explicit and tested.
