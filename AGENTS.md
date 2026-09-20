# AGENTS.md — Thermal Coding Standards & Architecture Guide

This document defines the architectural conventions, coding standards, loader requirements, build quirks, and historical lessons for **Thermal** (`github.com/jadmadi/thermal`). Any AI coding assistant (Claude Code, Gemini CLI, Devin, OpenCode, Cursor, etc.) working on this repository MUST read and adhere to these standards.

---

## 1. Project Purpose & Architecture

Thermal (`thermal`) is a high-performance, zero-allocation terminal contribution heatmap and leaderboard tool for AI coding assistants. It scans local databases and session transcripts across multiple AI tools, calculates activity and token metrics, and renders beautiful terminal heatmaps.

### Directory Structure

```
thermal/
├── cmd/thermal/         # Main CLI entrypoint (main.go, upgrade.go for self-update mechanism)
├── internal/loaders/    # Tool-specific ingestion engines and incremental delta-cache
├── internal/thermal/    # Heatmap algorithms, streak calculation, time parsing, report and project aggregation
├── internal/pricing/    # models.dev price catalog, cache, and cost estimation
├── internal/render/     # Terminal formatting, color palettes, dashboard, report, project & leaderboard UI
├── internal/version/    # SemVer constants and ldflag injection targets
├── build.sh             # Local multi-target build and UPX compression script
└── .goreleaser.yml      # Automated GitHub Actions release configuration
```

---

## 2. Core Coding & Architectural Standards

### A. Read-Only Data Ingestion (`internal/loaders/`)
1. **Never Modify User Data**: All SQLite connections MUST open in strictly read-only mode using URI parameters:
   ```go
   db, err := sql.Open("sqlite", "file:"+dbPath+"?mode=ro")
   ```
2. **Memory-Mapped I/O**: Immediately after opening any SQLite database, execute `PRAGMA mmap_size` to enable memory mapping for instant scanning of multi-gigabyte files:
   ```go
   db.Exec("PRAGMA mmap_size=268435456") // 256MB mmap window
   ```
3. **Robust Time Parsing (`parseSessionTime`)**: AI tool timestamps vary wildly across tools and versions (RFC3339 strings, Unix seconds, milliseconds, microseconds, SQLite strings). Always use robust multi-format parsing or `time_test.go` utilities. Never assume a single fixed format.
4. **Incremental Delta Caching (`cache.go`)**: For large databases (`Devin`, `OpenCode`, `MiMoCode`, `Codex`) and multi-file scanners (`command-code`, `Agy`), use the `LoadOrScanWithCache` mechanism (`~/.cache/thermal/<tool>.json`). Store exact file modification times (`mod_time`), sizes (`size`), or max seen transaction IDs so subsequent invocations take under `10ms`.
5. **Concurrency Safety**: Multi-file, directory, and JSONL log scanners (`command-code`, `Codex` rollout logs, `Agy` overview logs) must use bounded worker pools (`sync.WaitGroup` or semaphore channels capped at ~16 workers) with thread-safe aggregation (`sync.Mutex`). Never spawn unbounded goroutines over thousands of files.
6. **Disjoint Token Types**: Token type fields (`Input`, `Output`, `Reasoning`, `CacheRead`, `CacheWrite`) MUST be disjoint and add up to the recorded total. Sources that nest them (Codex and Grok put reasoning inside output and cache inside input; ZCode puts cache inside input) must subtract the nested parts in the loader, never in render or pricing.
7. **Project Attribution**: Loaders whose source records where a session ran MUST emit `thermal.ProjectDay` rows beside the daily rows, normalized through `thermal.ProjectKey` so subdirectories fold into their nearest git root. Sources with no project data return an empty slice. `ProjectKey` resolves symlinks first, because the same directory reached through a symlinked parent (a symlinked home, for example) must count once, and never invent a project name from a hash or an encoded directory name; it falls back to the recorded path only when no repository marker exists.
8. **Canonical Model Names**: Model ids MUST pass through `modelName()` before they become map keys (`DailyRow.Models`, `ProjectDay.Models`, `Summary.ModelBreakdown`), because tools disagree on case (`GLM-5.3-Flash` in ZCode versus `glm-5.3-flash` in OpenCode). One model is one row everywhere, including `thermal models`.

### B. TUI Boundary (`internal/tui/`)
1. **Imports stay inside the package**: `internal/tui` may import `internal/thermal`, `internal/loaders`, `internal/pricing`, and the charm stack. Nothing outside it imports `internal/tui` except `cmd/thermal`, and the CLI must never render a TUI frame itself.
2. **Aggregation belongs to `internal/thermal`**: a view calls an existing aggregator (`AggregateModels`, `AggregateProjects`, `AggregateToolMix`, `AggregateModelMix`, `AggregateStats`, `AggregateTrend`) and never re-derives a total. View totals must equal the matching static command for the same window, and the parity check is a test, not an inspection.
3. **Every frame is tested at two sizes**: a golden `View()` at 80x24 and 120x40 with colour forced off, plus a key-flow test for the keys the view owns.
4. **`thermal dashboard` is the only TUI entry point**: plain `thermal` keeps printing the leaderboard, a non-TTY stdout gets a pointer to the static commands, and `--json` never opens a UI.

### C. Chart Conventions (`internal/render/chart.go`)
1. **No colour-only encoding**: a bar always prints its number beside it, and a legend or caption carries the meaning, so a monochrome terminal loses nothing.
2. **The scale is named**: a caption states what the bar is measured against.
3. **Width aware**: charts size themselves against the terminal and shorten the bar rather than wrapping a line. A chart that overshoots the frame stops being pasteable.
4. **Capped at twelve rows** with a stated remainder, because a bar chart with eighty rows duplicates the table instead of summarising it.
5. **`--chart` never touches `--json`**: it only adds rows to stdout.

### D. Tool-Specific Loader Quirks
* **Devin (`devin.go`)**: Queries the SQLite DB joining `message_nodes` against `sessions`. Always check `metadata.metrics` for true input/output/cache token counts (`input_tokens`, `output_tokens`, `cache_creation_tokens`, `cache_read_tokens`). Check `prompt_history` (`updated_at` fallback to `created_at`) for accurate streak calculations across sessions without messages. The `sessions.model` column MUST be projected onto `DailyRow.Models` and `ProjectDay.Models`, because pricing needs a model and a day without one is priced at zero and disclosed as unattributed. Attribution is session-level carried down to the day a message was created, which is the finest granularity the source offers, and a session whose model is empty leaves its tokens unattributed rather than inventing a name.
* **OpenCode / MiMoCode (`internal/loaders/sqlite.go`)**: Read pre-aggregated token columns (`tokens_input`, `tokens_output`, `tokens_reasoning`, `tokens_cache_read/write`) plus `cost` and diff summaries. OpenCode v2 writes sessions to `session_v2` and abandons the legacy `session` table; probe for `session_v2` first and fold in legacy-only rows. Never assume a table stays the primary source across tool upgrades.
* **Codex (`codex.go`)**: Reads `state_5.sqlite` (`threads.tokens_used`, reasoning effort, source/model breakdown) as the primary source. Supplements with rollout JSONL logs (`~/.codex/sessions/**/*.jsonl`) for granular token breakdowns when available.
* **codewhale (`codewhale.go`)**: Reads `~/.codewhale/sessions/*.json` files for `metadata.total_tokens` and `metadata.cost.session_cost_usd`.
* **command-code (`commandcode.go`)**: Scans `~/.commandcode/sessions/*/transcript.jsonl` for message activity and `.meta.json` sidecars for model distributions.
* **Agy (`agy.go`)**: Scans `~/.gemini/antigravity-cli/brain/*/.system_generated/logs/transcript.jsonl` for step activity and model distribution (`Model Selection` entries); falls back to legacy `overview.txt` (same schema) and legacy root `~/.gemini/antigravity` when the `-cli` root is absent.
* **ZCode (`zcode.go`)**: Reads `~/.zcode/cli/db/db.sqlite` `model_usage` (per-request tokens, model, agent; `status = 'completed'` only) and `turn_usage` for turn counts. The DB records no cost, so Cost stays 0.
* **Grok (`grok.go`)**: Scans `<grok-home>/sessions/*/*/updates.jsonl` (`GROK_HOME` env or `~/.grok`) for `turn_completed` usage only. `costUsdTicks` converts at 1e-10 USD; `inputTokens` already includes cache, `reasoningTokens` is a subset of output, lifetime uses recorded `totalTokens`. Bounded worker pool like command-code.
* **Muse (`muse.go`)**: Reads `~/.local/share/muse/session-index.db` (sessions, prompt counts, model ids, microsecond timestamps). Activity-only: no token or cost columns exist. Prompt counts stand in for activity; upgrade to per-session `session.jsonl` when model-call frames accumulate.
* **Claude (`claude.go`)**: Scans `~/.claude/projects/*/*.jsonl` for assistant `message.usage` (input/output/cache tokens) and model ids; one file per session. No cost fields, so Cost stays 0. Bounded worker pool like command-code.
* **Droid (`droid.go`)**: Scans `~/.factory/sessions/*/*.jsonl` message records for activity; session files carry no token or cost telemetry. Bounded worker pool like command-code.
* **Project sources (token tools)**: OpenCode `session_v2.directory` (fall back to the `project` table worktree), MiMoCode `session.directory`, ZCode `session.directory`, Codex `threads.cwd`, Devin `sessions.working_directory`, Claude per-line `cwd`, Grok `summary.json.git_root_dir` (fall back to the URL-decoded session directory), codewhale `metadata.workspace`. Muse, Droid, and command-code carry a workspace or encoded directory but are deferred; Agy records no project at all.

---

## 3. Terminal Rendering & CLI (`internal/render/`)

1. **Leaderboard Categories**:
   * **Token Warriors**: Tools reporting token usage (`Devin`, `OpenCode`, `MiMoCode`, `Codex`, `codewhale`).
   * **Activity Hunters**: Tools reporting actions/messages/steps instead of tokens (`command-code`, `Agy`).
2. **Compact Number Formatting**: Always format large numbers concisely via `CompactNumber()`: `57.2B tok`, `44.0M tok`, `22.7K tok`, `304 step`.
3. **Color & Verbosity Flags**:
   * `--no-color`: Strips all ANSI escape sequences. Always check `colorEnabled` before emitting color codes.
   * `--verbose`: Outputs non-fatal loader diagnostic warnings (`database locked`, `missing directory`) exclusively to `os.Stderr`. Never pollute `os.Stdout` or JSON output (`--json`) with warnings.
4. **Report and Project Commands**: `daily`, `weekly`, `monthly`, `projects`, and `models` are reserved positional words (`thermal opencode weekly`, `thermal projects`). Report options (`--since`, `--until`, `--last`, `--order`, `--breakdown`, `--start-of-week`, `--offline`, `--no-estimate`) must error when no report word is present instead of being silently ignored, and `--sort`/`--top` must error where they do not apply: `streak|tokens|cost` for the leaderboard, `tokens|cost|days|recent` for projects, `tokens|cost` for models. Cost always comes from the source when recorded; `internal/pricing` estimates only days that have no recorded cost and that name a model. A day with no model is priced at zero, counted in `PeriodRow.UnattributedTokens` and `ProjectRow.UnattributedTokens`, and stated in the footer, so an estimated total never reads as complete when part of it could not be priced. Activity-only rows, where `Tokens` equals `Turns` and every other field is empty, are `isActivityOnly` and never enter token totals: a step count is not a token. Models the catalog cannot find are listed in the footer by name rather than treated as free. Model cost is always an estimate because recorded cost is not attributable to a single model. Project rows merge across tools by project key, tools rank by contributed tokens, and the table prints every row unless `--top` limits it.

---

## 4. Build, Versioning & Release Guidelines

### A. `.gitignore` Path Rules
* **IMPORTANT**: Never put bare binary names like `thermal` directly in `.gitignore`. Because `cmd/thermal/` and `internal/thermal/` share the name `thermal`, a bare `thermal` rule will silently ignore source files inside those directories (`e.g., cmd/thermal/upgrade.go`)!
* **Rule**: Always use root-anchored paths for build artifacts: `/thermal`, `/dist/`, `*.exe`.

### B. GoReleaser Configuration (`.goreleaser.yml`)
1. **Template Variables**: GoReleaser v2 strictly enforces `{{ .ShortCommit }}`. Never use `.Short_commit` (causes fatal build errors).
2. **UPX Compression**:
   ```yaml
   upx:
     - enabled: true
       compress: "9"
       goos:
         - linux
         - windows
   ```
   * **Rule**: Enable UPX compression **ONLY** for `linux` and `windows` binaries.
   * **Do NOT compress `darwin` (macOS) binaries with UPX**: Stripped `darwin/amd64` and `darwin/arm64` binaries must remain uncompressed so macOS Gatekeeper, code signing, and binary format verifiers do not reject the executable.

### C. Automated Release Pipeline (`release-please` + `goreleaser`)
1. **Conventional Commits**: All commit messages MUST follow Conventional Commits (`feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `test:`).
   * `release-please` (`googleapis/release-please-action@v5`) parses commits since the last release tag (`vx.y.z`) to calculate SemVer bumps and automatically generate PRs updating `CHANGELOG.md` and `.release-please-manifest.json`.
2. **Release Execution Flow**:
   * When the automated `release-please` PR (`chore(master): release x.y.z`) is merged into `main`, `release-please` tags the commit (`vx.y.z`) and creates the GitHub Release entry.
   * The `.github/workflows/release.yml` workflow triggers on `push: tags: ["v*"]`. It installs `upx` (`sudo apt-get install -y upx`) and runs GoReleaser (`goreleaser/goreleaser-action@v6`) to cross-compile across 5 OS/Arch targets, apply UPX compression to Linux/Windows binaries, and upload `.tar.gz` / `.zip` assets directly to the GitHub Release.

---

## 5. Testing & Verification Checklist

Before committing or submitting a pull request, run the verification suite:

```bash
# 1. Run all unit tests with race detection and verbose output
go test -v ./...

# 2. Verify compilation of the CLI binary
go build -o /tmp/thermal-test ./cmd/thermal

# 3. Test alias resolution and heatmap output locally
/tmp/thermal-test --tool auto
/tmp/thermal-test --tool devin
/tmp/thermal-test --tool agy --verbose
```

When modifying loaders, ensure unit tests inside `internal/loaders/*_test.go` cover edge cases (missing databases, corrupted files, zero-token sessions, timestamp variations, and schema differences) using mock temporary files or SQLite in-memory databases.

* **gofmt scope**: run `gofmt` only on files the change touches. Several test files carry pre-existing drift, and a repo-wide sweep adds unrelated noise to feature diffs.
