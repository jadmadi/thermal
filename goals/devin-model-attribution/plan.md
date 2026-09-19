# Plan: Devin model attribution and honest unestimated reporting

## Execution Steps
- [x] **Task 1 — Devin per-model days** (`internal/loaders/sqlite.go`): join `message_nodes` to `sessions` and emit `DailyRow.Models` keyed by `modelName(s.model)`, with the token types disjoint and summing to the day total the loader already reports. Handle a session whose model is empty by leaving that day's contribution unattributed rather than inventing a name.
- [x] **Task 2 — Cache bump** (`internal/loaders/cache.go`): raise `devinCacheVersion` so a cached snapshot without model attribution is discarded on first run.
- [x] **Task 3 — Parity test** (`internal/loaders/sqlite_test.go`): assert that the sum of a day's model tokens equals the day's total, that two sessions naming the same model on one day merge into one entry, and that a session with no model leaves the day's totals unchanged.
- [ ] **Task 4 — Unestimated disclosure** (next; the attribution above is the larger half and is verified) (`internal/pricing`, `internal/render`): report the tokens whose source names no model, so the footer can state them instead of leaving them silent. The existing per-model missing list keeps its meaning.
- [ ] **Task 5 — Report footers** (follows Task 4) (`internal/render/report.go`, `internal/render/projects.go`, `docs/pages/index.html`): state the unestimated tokens beside the estimated split, and update the guide's cost section to match.
- [x] **Task 6 — Verification**: `go test -race ./...`, build, and a hand check that a Devin day now prices and that the footer names the remainder. Commit with `Goal-Ref: devin-model-attribution` under a `feat:` subject, then open a pull request into `main`.
