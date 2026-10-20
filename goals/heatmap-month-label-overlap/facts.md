# Facts: Eliminate overlapping heatmap month labels

## Observed Evidence
- `docs/OUTPUT-GALLERY.md:197` OpenCode heatmap header: `ONov  Dec Jan Feb Mar ...`.
- `docs/OUTPUT-GALLERY.md:226` Agy heatmap header: `ONov  Dec Jan Feb Mar ...`.
- `internal/render/heatmap.go:37-48` writes `date.Format("Jan")` into `monthLine[w+i]` without clearing the previous label or checking collisions; `monthLine` starts as spaces but partial old characters remain.
- `internal/render/heatmap_test.go` currently checks heat cells/width but does not include a fixed-date late-October label sequence.

## Interpretation
`ONov` is the visible symptom of a write-without-clear algorithm. If October starts at a late Sunday, November starts at the next week and overwrites `ct` but not `O`. The fix must be algorithmic, not a special-case string replacement.

## Ownership and Constraints
`internal/render/heatmap.go` owns the header; `heatmap_test.go`/`dashboard_test.go` own golden coverage. Preserve `--no-color` and width behavior. No new dependencies.

## Verification Contract
```bash
go test -race ./internal/render -run 'Heatmap|Dashboard'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
/tmp/thermal-test opencode --no-color | grep -E 'Nov|ONov'
./scripts/simulated_user_gate.sh
```

## Open Questions
- Should a month be skipped when two labels cannot fit with one space, or should the header use single-letter labels at narrow widths? The goal prefers skip over partial labels; tests must pin it.
- Is a year marker required for 104-week views? If yes, it must be a separate token; if not, document the omission.
