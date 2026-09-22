# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.13.0](https://github.com/jadmadi/thermal/compare/v0.12.0...v0.13.0) (2026-09-22)


### Features

* **init:** initial release of Thermal under GNU AGPL-3.0 with commercial dual-licensing ([23a7733](https://github.com/jadmadi/thermal/commit/23a7733a8ce62fba5a8f52a0eb3e17c0581f806a))

## [0.12.0](https://github.com/jadmadi/thermal/compare/v0.11.0...v0.12.0) (2026-09-21)


### Features

* **replay:** implement thermal replay simulation against subscriptions and APIs (goals/workload-replay/goal.md) ([4f7ff02](https://github.com/jadmadi/thermal/commit/4f7ff02f5e7710deabc5454265330bcbee35057a))
* **testing:** add automated simulated user release gate and update gitignore ([b5651df](https://github.com/jadmadi/thermal/commit/b5651df195f1702078d233ba9fb3d4c5ea98c29c))


### Bug Fixes

* **testing:** ensure simulated user tests use dynamic timestamps and isolated home environment ([2c578a0](https://github.com/jadmadi/thermal/commit/2c578a0fd86e035939805ecd1a7612005d70502e))

## [Unreleased]

### Features

* **replay:** simulate historical developer workloads against subscription tiers and API models (`thermal replay`)
* **testing:** add automated simulated user testing release gate verifying all CLI commands, formatting, semantics, and language

## [0.11.0](https://github.com/jadmadi/thermal/compare/v0.10.0...v0.11.0) (2026-09-21)


### Features

* **analytics,render:** expose disjoint token composition and cache efficiency metrics (goals/cache-efficiency-metrics/goal.md) ([6e070bd](https://github.com/jadmadi/thermal/commit/6e070bd87108f8b6994d5adaeba7e5cff1d4cc48))
* **loaders:** add Nous Hermes agent loader and registry (goals/hermes-loader/goal.md) ([9e90d01](https://github.com/jadmadi/thermal/commit/9e90d017e24a4a4cdd04bcdd9b32f166a824b941))
* **loaders:** add prompt-level turn deduplication to Grok loader ([e589226](https://github.com/jadmadi/thermal/commit/e5892265a86dd65f07b7ab1dfb5df6e34543bab8))
* **loaders:** implement message-level deduplication in Claude Code loader (goals/claude-message-dedupe/goal.md) ([104ba23](https://github.com/jadmadi/thermal/commit/104ba23a39bdaacfe4d16e9e34d4dbbe2ed61da8))

## [0.10.0](https://github.com/jadmadi/thermal/compare/v0.9.1...v0.10.0) (2026-09-21)


### Features

* **docs:** polish layout, motion choreography, and window chrome in html docs ([d9ab799](https://github.com/jadmadi/thermal/commit/d9ab7997b7c5e56be42354f858ac69f72ebc3161))
* **docs:** switch canonical site domain to thermal.jadmadi.net ([a1c9c4d](https://github.com/jadmadi/thermal/commit/a1c9c4dd5002a62b4954ee3a820a5817b9d0fc43))
* **loaders:** add DeepSeek harness (dsh) loader and registry (goals/dsh-loader/goal.md) ([414a815](https://github.com/jadmadi/thermal/commit/414a8151615119ca3f043a8beafe8a67f9bfb194))
* **loaders:** report JSONL scans that stop early via Summary.Warnings (goals/jsonl-scan-diagnostics/goal.md) ([247bb80](https://github.com/jadmadi/thermal/commit/247bb80b2a49ceb455beb3cb76d650a5f43bb8f1))
* **pricing:** map tier variants to the base model, and classify the rest ([b361693](https://github.com/jadmadi/thermal/commit/b361693abb24a963f62b5fc0769288dd9322b5ac))
* **render,analytics:** telemetry accuracy, pricing estimates, and stats UI polish (goals/telemetry-and-ui-audit/goal.md) ([662f25a](https://github.com/jadmadi/thermal/commit/662f25a3e9c761f4369cd20425e75c268299dbee))
* **sila:** complete sila onboarding and wire repository (goals/sila-onboarding/goal.md) ([92ea6ee](https://github.com/jadmadi/thermal/commit/92ea6ee123cb044818dab401ce450f59e83617d3))


### Bug Fixes

* **loaders:** raise JSONL line ceilings to 32 MiB (goals/jsonl-line-ceilings/goal.md) ([ada5b52](https://github.com/jadmadi/thermal/commit/ada5b526722e6f20b3b8281a65b574d616ab7170))

## [0.9.1](https://github.com/jadmadi/thermal/compare/v0.9.0...v0.9.1) (2026-09-20)


### Bug Fixes

* **codex:** scan large rollout lines, and keep step counts out of token totals ([249b1ee](https://github.com/jadmadi/thermal/commit/249b1ee0df083268986fec5f3eb56b1a3639594b))
* **codex:** scan large rollout lines, and keep step counts out of token totals ([ce31aaa](https://github.com/jadmadi/thermal/commit/ce31aaac0c4a78fba84961372943f39ba0400643))

## [0.9.0](https://github.com/jadmadi/thermal/compare/v0.8.0...v0.9.0) (2026-09-20)


### Features

* **devin:** attribute models to project rows so projects can price them ([5741f5f](https://github.com/jadmadi/thermal/commit/5741f5fc98289f3712858833391d71324489ea48))
* **devin:** attribute models to project rows so projects can price them ([e4d542c](https://github.com/jadmadi/thermal/commit/e4d542cfce7f6df0166711b04299592a231f9b25))

## [0.8.0](https://github.com/jadmadi/thermal/compare/v0.7.0...v0.8.0) (2026-09-19)


### Features

* **devin:** attribute tokens to the session model so they can be priced ([0508a4e](https://github.com/jadmadi/thermal/commit/0508a4ea15c68d5718ba4a3aa694ebdd22cbde12))
* **devin:** attribute tokens to the session model so they can be priced ([818975d](https://github.com/jadmadi/thermal/commit/818975d0bfac021b5b0daad2f25373d4bb44c05c))


### Bug Fixes

* **reports:** count the tokens no price covers, and stop dropping token days ([49d6766](https://github.com/jadmadi/thermal/commit/49d676643bfdef21d531e396aa4147ead4ac6b3d))
* **reports:** count unpriceable tokens and stop dropping token days ([b130221](https://github.com/jadmadi/thermal/commit/b130221746f843f926627587acc4b5705a8a6135))

## [0.7.0](https://github.com/jadmadi/thermal/compare/v0.6.0...v0.7.0) (2026-09-19)


### Features

* **analytics:** add mix, stats, and trend verbs ([#22](https://github.com/jadmadi/thermal/issues/22)) ([9c28a0a](https://github.com/jadmadi/thermal/commit/9c28a0af26a610ebaff49a8b02cc9635e8120e85))
* **docs:** add author links, schema, FAQ, and crawler files to the docs site (goals/report-period-totals/goal.md) ([ccc7bc1](https://github.com/jadmadi/thermal/commit/ccc7bc1226aba6dc183de09f35dfa4358d31d2ab))
* **docs:** docs site, corrected report totals, and Pages publishing ([7a470d1](https://github.com/jadmadi/thermal/commit/7a470d15c49c290cbe27f9b631e62d11185a60e6))
* **reports:** --chart, dashboard docs, and the rules ([656db32](https://github.com/jadmadi/thermal/commit/656db32151827d2a6afa84160dd5d49201d40574))
* **reports:** add --chart, document the dashboard, and record the rules ([1cb2f50](https://github.com/jadmadi/thermal/commit/1cb2f501825abb429f30940eb3da20e8fa715567))
* **reports:** keep token-bearing periods only, grain-independent totals (goals/report-period-totals/goal.md) ([9c27be4](https://github.com/jadmadi/thermal/commit/9c27be46c625036754c9637b447bbe85d43372fd))
* **tui:** add the dashboard shell and Overview view ([050f4a5](https://github.com/jadmadi/thermal/commit/050f4a5bcf2cf4f27f08a42807f6b233285720d5))
* **tui:** add the Mix and Models views ([6a591f8](https://github.com/jadmadi/thermal/commit/6a591f8d72fb895bf374d50f98acf929792d073c))
* **tui:** add the Projects view with drill-down ([e15056f](https://github.com/jadmadi/thermal/commit/e15056f8cf5c4c4c40c1bbe38bec0b279b45683c))
* **tui:** add the Stats view ([338b8bd](https://github.com/jadmadi/thermal/commit/338b8bd833ab14bd45fb77ef8ed1ec758904cdc2))
* **tui:** dashboard shell and Overview view ([e8a5dbb](https://github.com/jadmadi/thermal/commit/e8a5dbb8436baed6829d004766104e9e574d02df))
* **tui:** Mix and Models views ([70c6239](https://github.com/jadmadi/thermal/commit/70c62399ad1daf7a32fc585bb243112b0e4684af))
* **tui:** Projects view with drill-down ([c5bb23c](https://github.com/jadmadi/thermal/commit/c5bb23c21486c5e61ea27781d96b8b8ff27c8a6d))
* **tui:** Stats view ([76b4fba](https://github.com/jadmadi/thermal/commit/76b4fba65c99b78427328e530c15f97dba835328))


### Bug Fixes

* **cost:** show the recorded and estimated split everywhere the total appears ([b76b9a5](https://github.com/jadmadi/thermal/commit/b76b9a538ea1c0ad6bf6b0507ebe825c7b019226))
* **cost:** show the recorded and estimated split everywhere the total appears ([66a7466](https://github.com/jadmadi/thermal/commit/66a74664fab3b6d5b6d58912da62ad340e3cbc11))
* **pages:** move the JSON reference inside the reports section and reindent it (goals/report-period-totals/goal.md) ([c9c6af8](https://github.com/jadmadi/thermal/commit/c9c6af80892ec4f8fa0e1a284db251ee1e7f3abe))
* **tools:** drop the ambiguous cc alias and point tool links at their sites (goals/report-period-totals/goal.md) ([2a81533](https://github.com/jadmadi/thermal/commit/2a81533383ea852f0f274f205a2aa47133a86c02))
* **tui:** keep the cursor on screen and the frame inside the terminal ([b1df32b](https://github.com/jadmadi/thermal/commit/b1df32bae06bf0eaec9a45db9c6a59ac309a0e78))
* **tui:** keep the cursor on screen and the frame inside the terminal ([0619190](https://github.com/jadmadi/thermal/commit/0619190fa2bba9f8486274318687771c36ab11ff))

## [0.6.0](https://github.com/jadmadi/thermal/compare/v0.5.0...v0.6.0) (2026-09-16)


### Features

* add daily, weekly, monthly reports, projects, and models ([#15](https://github.com/jadmadi/thermal/issues/15)) ([2097a62](https://github.com/jadmadi/thermal/commit/2097a62a258fea485abf5cc84b64ede17f764ac8))
* rank tokens and cost per project across tools ([#18](https://github.com/jadmadi/thermal/issues/18)) ([8eb3c27](https://github.com/jadmadi/thermal/commit/8eb3c278ab091993162a98717511d0c42e808532))

## [0.5.0](https://github.com/jadmadi/thermal/compare/v0.4.0...v0.5.0) (2026-09-15)


### Features

* **claude,droid:** add Claude Code token loader and Droid activity loader ([#13](https://github.com/jadmadi/thermal/issues/13)) ([e3254d1](https://github.com/jadmadi/thermal/commit/e3254d10a3741013f9dbae74510749fb45e88a57))

## [0.4.0](https://github.com/jadmadi/thermal/compare/v0.3.0...v0.4.0) (2026-09-15)


### Features

* **grok:** add Grok CLI loader from session turn telemetry ([#10](https://github.com/jadmadi/thermal/issues/10)) ([85af773](https://github.com/jadmadi/thermal/commit/85af773d6da6504f028b253d0b6a284a9f917c76))
* **muse:** add activity-only Muse loader from session index ([#11](https://github.com/jadmadi/thermal/issues/11)) ([f15f943](https://github.com/jadmadi/thermal/commit/f15f9431a1f62512fab0d3d25a0f95905e725be6))

## [0.3.0](https://github.com/jadmadi/thermal/compare/v0.2.3...v0.3.0) (2026-09-15)


### Features

* **zcode:** add ZCode loader with model usage telemetry ([#6](https://github.com/jadmadi/thermal/issues/6)) ([54cd9d7](https://github.com/jadmadi/thermal/commit/54cd9d7bee17fec079c9b0999af202994d5410b4))

## [0.2.3](https://github.com/jadmadi/thermal/compare/v0.2.2...v0.2.3) (2026-09-13)


### Bug Fixes

* **agy:** read live antigravity-cli transcripts for step activity ([#4](https://github.com/jadmadi/thermal/issues/4)) ([c7f4fa6](https://github.com/jadmadi/thermal/commit/c7f4fa631d5020dc2647644a7bfd26b3ae5ea3d3))

## [0.2.2](https://github.com/jadmadi/thermal/compare/v0.2.1...v0.2.2) (2026-09-12)


### Bug Fixes

* **opencode:** read session_v2 table written by OpenCode v2 ([#2](https://github.com/jadmadi/thermal/issues/2)) ([037a073](https://github.com/jadmadi/thermal/commit/037a0735d4662109651b6ba920b7142f461897b6))

## [0.2.1](https://github.com/jadmadi/thermal/compare/v0.2.0...v0.2.1) (2026-09-11)

Documentation and repository chores only; no user-facing changes. Tagged outside release-please, recorded here to keep the manifest in sync.

## [0.2.0](https://github.com/jadmadi/thermal/compare/v0.1.0...v0.2.0) (2026-07-18)


### Features

* add open source release plan and config ([61c9449](https://github.com/jadmadi/thermal/commit/61c9449d6baab5945e7a1a540217e5405cbc41b8))
* optimize loaders with memory-mapped SQLite and parallel workers, add build.sh automation, and integrate release-please ([bdf5b2d](https://github.com/jadmadi/thermal/commit/bdf5b2da63d0270b466802250804df5b21e0a330))


### Bug Fixes

* update author name in readme ([19e2c54](https://github.com/jadmadi/thermal/commit/19e2c5448949c306a7cb65b815aa3ac2f17031cd))

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
