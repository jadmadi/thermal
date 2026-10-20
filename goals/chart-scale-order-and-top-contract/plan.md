# Plan: Fix chart scaling, ordering, row counts, and no-op --chart flags

## Execution Steps
- [x] **Task 1 — Fix chart max and caption**: Compute the max over `rep.Rows[:shown]` and use it in bars/caption. Acceptance: synthetic cost-sorted project report with first row 3.9B and later row 7.1B renders the later row at full width and caption `7.1B`.
- [x] **Task 2 — Define order/row-count policy**: Choose table-order with an explicit caption or independent chart sort; align cap/remainder with `--top` or label the chart cap. Acceptance: tests for `--top 5/10/12/0` and projects/models.
- [x] **Task 3 — Resolve --chart no-ops**: Implement a concrete chart for mix/stats/trend or reject `--chart` in validation and help. Acceptance: commands with `--chart` either show the named chart or exit 1 with a clear message; help matches.
- [x] **Task 4 — Width and golden tests**: Add 80/120/160 tests with long labels and monochrome output. Acceptance: no line wraps/exceeds width; `go test -race ./internal/render -run TestChart` passes.
- [x] **Task 5 — Docs and attribution commit**: Add migration/help/README/project.jsonc updates; run full gate; commit with the goal token. Acceptance: gate clean; commit ends in `(goals/chart-scale-order-and-top-contract/goal.md)`.

## Dependencies and Handoff
The gallery goal captures the corrected charts. The formatting goal owns legends/cost markers; reconciliation owns totals; do not change them here.

## Verification
```bash
go test -race ./internal/render -run 'Chart|Project|Model'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```
