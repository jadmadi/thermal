# Plan: Chart flags, docs, and 0.7.0 release

## Execution Steps
- [ ] **Task 1 — Chart flags**: Add `--chart` to `projects`, `models`, `daily`, `weekly` and `monthly`, printing concrete bar rows under the tables without changing the JSON shape.
- [ ] **Task 2 — Docs**: README sections for the chart verbs and the dashboard, with a screencast and the key list.
- [ ] **Task 3 — AGENTS.md rules**: Record the TUI dependency boundary (TUI imports confined to `internal/tui`, aggregation in `internal/thermal`, CLI never imports the TUI), chart conventions (no color-only encoding, color degradation, width awareness), and the explicit `thermal dashboard` entry rule.
- [ ] **Task 4 — Verification on the merged state**: Merge latest main into the branch, then run the full checklist: `go test -v ./...`, `go build`, and the manual runs including the new commands.
- [ ] **Task 5 — Release**: Land through a pull request and cut 0.7.0 through release-please, then verify the tag, assets, and the released binary.
- [ ] **Task 6 — Attribution commit**: Commit with `Goal-Ref: charting-flags-docs-release` using a `chore:` or `feat:` subject, then run `sila goals`.
