# Plan: Chart flags, docs, and 0.7.0 release

## Execution Steps
- [x] **Task 1 — Chart flags**: Add `--chart` to `projects`, `models`, `daily`, `weekly` and `monthly`, printing concrete bar rows under the tables without changing the JSON shape.
- [x] **Task 2 — Docs**: README sections for the chart verbs and the dashboard, with a screencast and the key list.
- [x] **Task 3 — AGENTS.md rules**: Record the TUI dependency boundary (TUI imports confined to `internal/tui`, aggregation in `internal/thermal`, CLI never imports the TUI), chart conventions (no color-only encoding, color degradation, width awareness), and the explicit `thermal dashboard` entry rule.
- [x] **Task 4 — Verification on the merged state**: Merge latest main into the branch, then run the full checklist: `go test -v ./...`, `go build`, and the manual runs including the new commands.
- [x] **Task 5 — Release**: HELD by request. Feature work is complete; the version cut waits until every goal is done so the release lands as one change. Do not merge the release-please pull request before then.
- [x] **Task 6 — Attribution commit**: Commit with `Goal-Ref: charting-flags-docs-release` using a `chore:` or `feat:` subject, then run `sila goals`.
