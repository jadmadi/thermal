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
├── internal/thermal/    # Heatmap algorithms, streak calculation, and time parsing
├── internal/render/     # Terminal formatting, color palettes, dashboard & leaderboard UI
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

### B. Tool-Specific Loader Quirks
* **Devin (`devin.go`)**: Queries the SQLite DB joining `message_nodes` against `sessions`. Always check `metadata.metrics` for true input/output/cache token counts (`input_tokens`, `output_tokens`, `cache_creation_tokens`, `cache_read_tokens`). Check `prompt_history` (`updated_at` fallback to `created_at`) for accurate streak calculations across sessions without messages.
* **OpenCode / MiMoCode (`opencode.go`, `mimocode.go`)**: Read pre-aggregated `session_summary` or `sessions` tables containing token counts, costs, and diff lines.
* **Codex (`codex.go`)**: Reads `state_5.sqlite` (`threads.tokens_used`, reasoning effort, source/model breakdown) as the primary source. Supplements with rollout JSONL logs (`~/.codex/sessions/**/*.jsonl`) for granular token breakdowns when available.
* **codewhale (`codewhale.go`)**: Reads `~/.codewhale/sessions/*.json` files for `metadata.total_tokens` and `metadata.cost.session_cost_usd`.
* **command-code (`commandcode.go`)**: Scans `~/.commandcode/sessions/*/transcript.jsonl` for message activity and `.meta.json` sidecars for model distributions.
* **Agy (`agy.go`)**: Scans `~/.gemini/antigravity-cli/brain/*/overview.txt` (`Step <id>:` step count activity) and `.system_generated/logs/transcript.jsonl` (`MODEL` step entries) for model distribution.

---

## 3. Terminal Rendering & CLI (`internal/render/`)

1. **Leaderboard Categories**:
   * **Token Warriors**: Tools reporting token usage (`Devin`, `OpenCode`, `MiMoCode`, `Codex`, `codewhale`).
   * **Activity Hunters**: Tools reporting actions/messages/steps instead of tokens (`command-code`, `Agy`).
2. **Compact Number Formatting**: Always format large numbers concisely via `CompactNumber()`: `57.2B tok`, `44.0M tok`, `22.7K tok`, `304 step`.
3. **Color & Verbosity Flags**:
   * `--no-color`: Strips all ANSI escape sequences. Always check `colorEnabled` before emitting color codes.
   * `--verbose`: Outputs non-fatal loader diagnostic warnings (`database locked`, `missing directory`) exclusively to `os.Stderr`. Never pollute `os.Stdout` or JSON output (`--json`) with warnings.

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
