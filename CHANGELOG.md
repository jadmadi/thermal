# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **`build.sh` build automation script** — versatile build script supporting `--release` (`-ldflags="-s -w" -trimpath`), `--dev`, and `--upx` modes, cross-compilation (`--target` and `--all`), and detailed `--help` documentation explaining build trade-offs and binary sizes.
- **`--verbose` / `-v` diagnostic flag** — outputs non-fatal tool loading failures, missing databases, and parsing warnings directly to `os.Stderr`.
- **`thermal upgrade` command** — self-upgrade by checking GitHub Releases, downloading the matching binary for your OS/arch, and atomically replacing the running executable.
- **`thermal version` command** — prints version, commit hash, and build date (injected via ldflags at build time).
- **`internal/version` package** — build-time version embedding via `-ldflags`.
- **Token Warriors / Activity Hunters split** — the leaderboard now separates tools with real token consumption data from tools that only track activity (messages, steps). Tools are auto-categorized based on whether token data is available.
- **Cost column** in the leaderboard for tools that track cost (OpenCode, MiMoCode, codewhale). Sub-cent formatting shows costs like `$0.0032`.
- **Devin loader** — extracts real token metrics from `message_nodes.metadata.metrics` (input, output, cache, reasoning tokens), session duration from `sessions` table, and engagement from `prompt_history`. Devin moved from Activity Hunters to Token Warriors.
- **Codex loader rewrite** — reads `state_5.sqlite` (`threads.tokens_used`, `model`, `source`, `reasoning_effort`, `agent_role`, session duration) as the primary source, supplements with rollout JSONL for the input/output/reasoning/cache token breakdown. Codex moved from Activity Hunters to Token Warriors.
- **codewhale loader rewrite** — extracts `metadata.total_tokens`, `metadata.cost.session_cost_usd`, `metadata.model`, `metadata.mode`, and `metadata.cumulative_turn_secs` from session JSON files. codewhale moved from Activity Hunters to Token Warriors.
- **OpenCode loader rewrite** — leverages pre-aggregated `session.tokens_*` columns, `cost`, `agent`, and `summary_additions/deletions/files` for faster and more comprehensive data extraction.
- **MiMoCode loader** — aggregates from `message.data.tokens` and surfaces `session.summary_additions/deletions/files` and agent distribution.
- **Agy loader rewrite** — reads `overview.txt` from all brain sessions (was only reading `transcript.jsonl` from 1 of 14 sessions). Now surfaces 304 steps across 14 sessions instead of 57 steps from 1 session. Extracts model info from `transcript.jsonl` USER_SETTINGS_CHANGE events.
- **Session duration tracking** — `LongestSessionMs` field added to Summary, surfaced in dashboard for all tools that have session timestamps.
- **Model breakdown** — `ModelBreakdown` field shows which models were used per tool (e.g., Codex: gpt-5.4-mini:6, gpt-5.4:1).
- **Agent/source breakdown** — `AgentBreakdown` field shows source distribution (e.g., Codex: cli:7, vscode:3, exec:1).
- **Lines added/deleted and files touched** — `LinesAdded`, `LinesDeleted`, `FilesTouched` fields added to Summary for tools that track code diffs (OpenCode, MiMoCode).
- **Progress bar** — new `internal/render/progress.go` for loading indicator, hardened with `atomic.Bool` and `sync.Once` for thread safety.
- **Session cache** — `internal/loaders/cache.go` for caching loader results.
- **GitHub Actions release workflow** and **GoReleaser config** for cross-platform binary builds with version injection.

### Performance & Optimization

- **Devin SQLite database optimization** — added incremental delta scanning (`row_id > c.MaxRowID`), memory-mapped I/O (`PRAGMA mmap_size`), and byte-string pre-filtering (`chat_message LIKE '%"assistant"%'`), reducing query times on large (7GB+) databases from ~11s down to `< 2ms`.
- **SQLite memory-mapped I/O across all loaders** — `LoadOpenCodeData` and `LoadMiMoCodeData` benefit from `PRAGMA cache_size = -64000` and `PRAGMA mmap_size = 30GB` alongside raw string pre-filters.
- **Parallel multi-worker loaders** — `internal/loaders/agy.go`, `commandcode.go`, and `codex.go` use bounded parallel worker pools (limit of 8 concurrent workers) to scan multi-file directories and JSONL rollouts simultaneously.
- **Binary size reduction** — configured `-trimpath` and `-ldflags="-s -w"` in `.goreleaser.yml` and `build.sh` to strip DWARF symbol tables and host filepaths, reducing release binaries from ~15MB to ~9.8MB (35% reduction).

### Changed

- **Leaderboard format** — now split into two sections (Token Warriors, Activity Hunters) with a Cost column. Old format was a single flat list.
- **Tool registry architecture (`DataSubdir`)** — refactored tool registry path resolution (`internal/loaders/registry.go`) to use clean `DataSubdir` property mappings instead of hardcoded switch cases.
- **README** — updated supported tools table, leaderboard example, build instructions (`build.sh`), and How It Works section to reflect the new data sources, metrics, and optimizations.
- **`.goreleaser.yml`** — added `-trimpath` and ldflags (`-s -w`) to inject version, commit, and date into the binary while minimizing size.

### Fixed

- **Daylight Saving Time (DST) streak accuracy** — replaced 24-hour time subtraction (`24*time.Hour`) in streak logic (`internal/thermal/streak.go`) with calendar date arithmetic (`AddDate(0, 0, 1)`) to guarantee accuracy during spring-forward (23h) and fall-back (25h) transitions. Ensured strictly ascending quartile thresholds.
- **Self-upgrader network and decompression guards** — hardened `thermal upgrade` (`cmd/thermal/upgrade.go`) with package-wide 30-second HTTP timeouts, context propagation across requests, exact OS/arch asset checking, and `io.LimitReader(tr, 250*1024*1024)` decompression limits against disk starvation attacks.
- **Codex token over-counting (260x)** — the old parser summed ALL cumulative `total_token_usage` values from every `token_count` event in every rollout file, producing 11.6B instead of the correct 43.9M. The new loader uses `state_5.sqlite` as the source of truth and only reads the LAST `token_count` event per rollout file for the breakdown.
- **command-code session inflation** — the old glob `*.jsonl` matched 115 files (54 sessions + 54 checkpoints + 7 hooks-audit), inflating counts to 61 sessions and 6,303 messages. Now correctly filters to 54 sessions and 5,614 messages.
- **Agy session under-counting** — the old loader only read `transcript.jsonl` (1 of 14 sessions = 57 steps). Now reads `overview.txt` from all 14 sessions (304 steps).
- **SQLite NULL handling** — Codex `state_5.sqlite` has NULL values in `model`, `reasoning_effort`, and `agent_role` columns. Switched to `sql.NullString` to avoid scan errors.

## [v0.1.0] - 2026-06-12

### Added

- Initial open source release.
- GitHub-style contribution heatmap for AI coding tools: MiMoCode, OpenCode, Codex, Agy, command-code, codewhale.
- Leaderboard mode (`thermal`) showing all installed tools ranked by activity.
- Per-tool heatmap mode (`thermal --tool <name>`).
- JSON output (`thermal --json`).
- Configurable weeks range (`thermal --weeks 26`).
- Auto-detect mode (`thermal --tool auto`).
- MIT license.

[Unreleased]: https://github.com/jadmadi/thermal/compare/v0.1.0...HEAD
[v0.1.0]: https://github.com/jadmadi/thermal/releases/tag/v0.1.0
