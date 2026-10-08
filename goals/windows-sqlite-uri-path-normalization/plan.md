# Plan: Windows SQLite URI and Path Normalization

## Execution Steps
- [x] Phase 1: Shared Helper: Define `sqliteDSN(path string, pragma string)` in `internal/loaders` converting path via `filepath.ToSlash` and building the URI.
- [x] Phase 2: Loader Adoption: Refactor `sqlite.go`, `hermes.go`, `agy.go`, `codex.go`, `zcode.go`, and `muse.go` to use the shared helper.
- [x] Phase 3: Unit Testing: Add unit tests verifying Windows paths (`C:\Users\foo\db.sqlite`, paths with spaces) parse into valid SQLite connection strings.
- [x] Phase 4: Attribution Commit: Commit with `(goals/windows-sqlite-uri-path-normalization/goal.md)` using a `fix:` or `feat:` subject, then run `sila goals`.

