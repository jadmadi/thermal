# AGENTS.md — Thermal Coding Standards & Architecture Guide

This document defines the architectural conventions, coding standards, loader requirements, build quirks, and historical lessons for **Thermal** (`github.com/jadmadi/thermal`). Any AI coding assistant (Claude Code, Gemini CLI, Devin, OpenCode, Cursor, etc.) working on this repository MUST read and adhere to these standards.

---

## 1. Project Purpose & Architecture

Thermal (`thermal`) is a high-performance, zero-allocation terminal contribution heatmap and leaderboard tool for AI coding assistants. It scans local databases and session transcripts across multiple AI tools, calculates activity and token metrics, and renders beautiful terminal heatmaps.

Thermal operates under a **Written Solo Authority** model led by Jad Madi. Architectural invariants, in-scope domains, and explicit anti-goals (e.g., zero cloud telemetry, no background daemons, sub-10ms cached invocation target) are codified in [`GOVERNANCE.md`](GOVERNANCE.md). All AI agents and human contributors must strictly observe these scope boundaries.

### Directory Structure

```
thermal/
├── cmd/thermal/         # Main CLI entrypoint (main.go, upgrade.go for self-update mechanism)
├── internal/loaders/    # Tool-specific ingestion engines and incremental delta-cache
├── internal/thermal/    # Heatmap algorithms, streak calculation, time parsing, report and project aggregation
├── internal/pricing/    # models.dev price catalog, cache, and cost estimation
├── internal/render/     # Terminal formatting, color palettes, report, project & leaderboard UI
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
4. **Incremental Delta Caching (`cache.go`, `codexcache.go`, `dircache.go`)**: Large databases and multi-file scanners use disk-backed delta snapshot caches in `~/.cache/thermal/<tool>/` (`DevinCache`, `codexRolloutCachePath`, and `DirScanCache` keyed on canonical root + relative path, byte size, and nanosecond `mod_time`) so subsequent invocations take `<10ms`.
5. **Concurrency Safety**: Multi-file, directory, and JSONL log scanners (`command-code`, `Codex` rollout logs, `Agy` overview logs) must use bounded worker pools (`sync.WaitGroup` or semaphore channels capped at ~16 workers) with thread-safe aggregation (`sync.Mutex`). Never spawn unbounded goroutines over thousands of files.
6. **Disjoint Token Types**: Token type fields (`Input`, `Output`, `Reasoning`, `CacheRead`, `CacheWrite`) MUST be disjoint and add up to the recorded total. Sources that nest them (Codex and Grok put reasoning inside output and cache inside input; ZCode and command-code put cache inside input) must subtract the nested parts in the loader, never in render or pricing.
7. **Project Attribution**: Loaders whose source records where a session ran MUST emit `thermal.ProjectDay` rows beside the daily rows, normalized through `thermal.ProjectKey` (resolving symlinks first and folding subdirectories into their nearest git root; never invent a project name from a hash).
8. **Canonical Model Names**: Model ids MUST pass through `modelName()` before they become map keys (`DailyRow.Models`, `ProjectDay.Models`, `Summary.ModelBreakdown`), because tools disagree on case (`GLM-5.3-Flash` in ZCode versus `glm-5.3-flash` in OpenCode). One model is one row everywhere, including `thermal models`.
9. **JSONL Scanner Line Ceilings**: Every JSONL scan MUST set an explicit ceiling that fits real data, using the shared `newJSONLScanner()` constructor (32 MiB ceiling today).
10. **JSONL Scan Diagnostics**: Every JSONL scan loop MUST check `scanner.Err()` and report errors through `Summary.Warnings` using `formatScanWarning()` (`json:"-"`, `--verbose` stderr only; never stdout or `--json`).
11. **SQLite Rows Iteration Safety**: Every database query iterating over `rows.Next()` MUST explicitly check `rows.Err()` immediately after loop termination.

### B. TUI Boundary (`internal/tui/`)
1. **Imports stay inside the package**: `internal/tui` may import `internal/thermal`, `internal/loaders`, `internal/pricing`, and the charm stack. Nothing outside it imports `internal/tui` except `cmd/thermal`, and the CLI must never render a TUI frame itself.
2. **Aggregation belongs to `internal/thermal`**: the live view calls the tracker and existing aggregators and never re-derives a total. Live totals must equal the matching static command for the same window, and the parity check is a test, not an inspection.
3. **Every frame is tested at two sizes**: a golden `View()` at 80x24 and 120x40 with colour forced off, plus a key-flow test for the keys the view owns.
4. **`thermal live` is the only TUI entry point**: plain `thermal` keeps printing the leaderboard, a non-TTY stdout gets a pointer to the static commands, and `--json` never opens a UI. The historical `thermal dashboard` and `stats --dense` surfaces are retired; their replacement is `thermal live` plus static reports.

### C. Chart Conventions (`internal/render/chart.go`)
1. **No colour-only encoding**: a bar always prints its number beside it, and a legend or caption carries the meaning, so a monochrome terminal loses nothing.
2. **The scale is named**: a caption states what the bar is measured against.
3. **Width aware**: charts size themselves against the terminal and shorten the bar rather than wrapping a line. A chart that overshoots the frame stops being pasteable.
4. **Capped at twelve rows** with a stated remainder, because a bar chart with eighty rows duplicates the table instead of summarising it.
5. **`--chart` never touches `--json`**: it only adds rows to stdout.

### D. Tool-Specific Loader Quirks
* **Devin (`devin.go`)**: Joins `message_nodes` + `sessions` (`metadata.metrics` for true input/output/cache tokens; `prompt_history` for streaks). Projects `sessions.model` onto `DailyRow.Models` and `ProjectDay.Models` (empty model stays unattributed).
* **OpenCode / MiMoCode (`sqlite.go`)**: Probes `session_v2` first and folds in legacy `session` rows (`tokens_input/output/reasoning/cache_read/write`, `cost`).
* **Codex (`codex.go`)**: Reads `state_5.sqlite` (`threads.tokens_used`) supplemented by `~/.codex/sessions/**/*.jsonl` rollout logs.
* **codewhale (`codewhale.go`)**: Reads `~/.codewhale/sessions/*.json` (`metadata.total_tokens`, `metadata.cost.session_cost_usd`).
* **command-code (`commandcode.go`)**: Scans `~/.commandcode/projects/*/*.jsonl` with bounded pool; subtracts `cacheReadTokens`/`cacheWriteTokens` from `inputTokens`; `.meta.json` model fallback; legacy flat transcripts count as turns only.
* **Agy (`agy.go`)**: Scans `~/.gemini/antigravity-cli/brain/*/.system_generated/logs/transcript.jsonl` (fallback `overview.txt` and `~/.gemini/antigravity`) for step activity and `Model Selection`.
* **ZCode (`zcode.go`)**: Reads `~/.zcode/cli/db/db.sqlite` (`model_usage` with `status = 'completed'`, `turn_usage`); Cost is 0.
* **Grok (`grok.go`)**: Scans `<grok-home>/sessions/*/*/updates.jsonl` for deduplicated `turn_completed` events (`prompt_id` -> `turn_id` -> `_meta.eventId`); `costUsdTicks` at `1e-10` USD; subtracts cache from input and reasoning from output.
* **Muse (`muse.go`) & Droid (`droid.go`)**: Activity-only (`~/.local/share/muse/session-index.db` prompt counts; `~/.factory/sessions/*/*.jsonl` message records).
* **Claude (`claude.go`)**: Scans `~/.claude/projects/*/*.jsonl` for `message.usage` and model ids; Cost is 0.
* **DeepSeek (`dsh.go`)**: Scans `~/.dsh/storages/session_projcache/sessions/*.json` + monolithic `session_projcache.json` (`DSH_HOME`), mapping `uncachedInputTokens`, `outputTokens`, `cacheReadTokens`, `cacheWriteTokens`, and `identity.cwd`.
* **Nous Hermes (`hermes.go`)**: Reads `~/.hermes/state.db` (`HERMES_HOME`) `sessions` and `session_model_usage` (`cwd` -> `ProjectKey`).

---

## 3. Terminal Rendering & CLI (`internal/render/`)

1. **Leaderboard Categories**:
   * **Token Warriors**: Tools reporting token usage (`Devin`, `OpenCode`, `MiMoCode`, `Codex`, `CMD` (command-code), `codewhale`, `ZCode`, `Grok`, `Claude`, `DeepSeek (DSH)`, `Nous Hermes`, `Zed`).
   * **Activity Hunters**: Tools reporting actions/messages/steps instead of tokens (`Agy`, `Muse`, `Droid`).
2. **Compact Number Formatting**: Always format large numbers concisely via `CompactNumber()`: `57.2B tok`, `44.0M tok`, `22.7K tok`, `304 step`.
3. **Color & Verbosity Flags**:
   * `--no-color`: Strips all ANSI escape sequences. Always check `colorEnabled` before emitting color codes.
   * `--verbose`: Outputs non-fatal loader diagnostic warnings (`database locked`, `missing directory`) exclusively to `os.Stderr`. Never pollute `os.Stdout` or JSON output (`--json`) with warnings.
4. **Report, Project, Analytics, and Extended Commands**:
   * Reserved positional words: `daily`, `weekly`, `monthly`, `projects`, `models`, `mix`, `stats`, `trend`, `audit`, `share`, `live`, `license`, `changelog`, `version`, `upgrade`.
   * Report flags (`--since`, `--until`, `--last`, `--order`, `--breakdown`, `--start-of-week`, `--offline`, `--no-estimate`) error when no report word is present. `--sort`/`--top` validate per command (`streak|tokens|cost` on leaderboard, `tokens|cost|days|recent` on projects, `tokens|cost` on models). `--interval`, `--stream`, and `--flame-ceiling` apply to `live`.
   * Pricing (`internal/pricing`): Estimates only days without recorded cost that name a model (prefixed with `~`, disabled via `--no-estimate`). Unattributed tokens (`PeriodRow.UnattributedTokens`, `ProjectRow.UnattributedTokens`) and unknown catalog models are priced at zero and disclosed in the footer. Catalog lookup strips one tier suffix (`-high`, `-medium`, `-max`, `-mini`, `-fast`, `-eco`) and never uses fuzzy/stem matching. Activity-only rows (`isActivityOnly`) never enter token totals.
5. **Dual-Licensing & Contributor Gating**: Thermal is dual-licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)** (community edition) and a **Commercial Enterprise License**. All external contributions require acceptance of the Contributor License Agreement (CLA) in `CONTRIBUTING.md` (via git `Signed-off-by:` trailers or PR sign-off) granting commercial sublicensing rights to Jad Madi so external contributions never deadlock commercial dual-licensing rights. Third-party permissive notices are preserved in `NOTICES.md`.
6. **Breaking Change & Migration Governance Gate**: Any pull request or commit that modifies CLI flag semantics, removes an alias, alters `--json` schema keys, or changes token aggregation behavior MUST include a matching entry in [`docs/MIGRATION.md`](docs/MIGRATION.md) following the 6-part Entry Anatomy (stable verb-prefixed heading, detection regex with literal error output, 3-sentence rationale, before/after snippets, fix command, and escape hatch). PRs altering user-visible CLI contracts without migration documentation will not be approved.

---

## 4. Build, Versioning & Release Guidelines

### A. `.gitignore` Path Rules
* **IMPORTANT**: Never put bare binary names like `thermal` directly in `.gitignore`. Because `cmd/thermal/` and `internal/thermal/` share the name `thermal`, a bare `thermal` rule will silently ignore source files inside those directories (`e.g., cmd/thermal/upgrade.go`)!
* **Rule**: Always use root-anchored paths for build artifacts: `/thermal`, `/dist/`, `*.exe`.
* **Repository Asset & Build Hygiene**: Keep the repository root clean and professional. Place all screenshots, diagrams, and branding media inside `assets/` or `docs/pages/`. Never leave loose build binaries in the repository root; build to `/tmp` (e.g. `go build -o /tmp/thermal-test ./cmd/thermal`) or use `build.sh`. Archive temporary scratch files into `rm/` (`mkdir -p rm && mv <path> rm/`).

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
1. **Conventional Commits & Sila Goal Attribution**:
   * All commit messages MUST follow Conventional Commits (`feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `test:`, `perf:`).
   * **Impact-First Subjects**: The commit subject MUST describe user-observable CLI or TUI behavior rather than internal file paths, internal packages, or test harness mechanics.
   * **Goal Token Positioning**: Sila goal closure tokens `(goals/<slug>/goal.md)` MUST be appended at the very end of the commit subject (e.g. `feat(cli): add staged deprecation warnings & upgrade migration alerts (goals/cli-staged-deprecation-and-upgrade/goal.md)`). This ensures Sila's closed-loop goal attribution verifies the commit while preserving clean, readable changelog generation.
   * `release-please` (`googleapis/release-please-action@v5`) parses commits since the last release tag (`vx.y.z`) to calculate SemVer bumps and automatically generate PRs updating `CHANGELOG.md` and `.release-please-manifest.json`.
   * **Changelog Filtering (`release-please-config.json`)**: Internal maintenance, documentation, test updates, and CI commits (`chore`, `test`, `ci`, `docs`, `refactor`, `style`, `build`) are configured as hidden in `release-please-config.json` so public changelogs focus purely on user-facing features, bug fixes, performance gains, and breaking changes.
2. **Editorial Release PR Polish**:
   * When `release-please` opens an automated release pull request (`chore(main): release x.y.z`), apply editorial changelog curation before merging:
     - Review generated bullet points for clarity, narrative flow, and user impact.
     - Strip raw internal scaffolding references (such as `(goals/<slug>/goal.md)` tokens) from the final PR description.
     - Ensure any breaking changes adhere to the 3-part form: **What breaks**, **Who it hits**, and **What to do**, cross-linking directly to `docs/MIGRATION.md`.
3. **Release Execution Flow**:
   * When the automated `release-please` PR (`chore(main): release x.y.z`) is merged into `main`, `release-please` tags the commit (`vx.y.z`) and creates the GitHub Release entry.
   * The `.github/workflows/release.yml` workflow triggers on `push: tags: ["v*"]`. It installs `upx` (`sudo apt-get install -y upx`) and runs GoReleaser (`goreleaser/goreleaser-action@v6`) to cross-compile across 5 OS/Arch targets, apply UPX compression to Linux/Windows binaries, and upload `.tar.gz` / `.zip` assets directly to the GitHub Release.

---

## 5. Testing & Verification Checklist

When modifying loaders, ensure unit tests inside `internal/loaders/*_test.go` cover edge cases (missing databases, corrupted files, zero-token sessions, timestamp variations, and schema differences) using mock temporary files or SQLite in-memory databases.

* **Tiered Gate Verification (`scripts/check.sh` & `./scripts/output_gallery_gate.sh`)**:
  - **Fast-Path Pre-Commit (`./scripts/check.sh --pre-commit` / `-q`)**: Runs in `<3s`. Audits active diff for secrets/merge markers/debug lines via Diff Sentry, runs `gofmt -l` strictly on modified files, and executes unit tests ONLY for modified Go packages. Install via `./scripts/check.sh --install-hooks`.
  - **Full Pre-Push & CI Gate (`./scripts/check.sh` / `--full`)**: Runs full race detector (`go test -race ./...`), per-package statement coverage table (`go test -cover ./...`), Go report card (`go vet ./...`), binary compilation (`/tmp/thermal-test`), and the simulated user release gate.
* **Simulated User Release Gate**: Before announcing any feature or release, run `./scripts/simulated_user_gate.sh` (also enforced in CI). It runs end-to-end user flows across every CLI command and flag combination, checking exit codes, ANSI stripping under `--no-color`, numerical sanity (no NaN, Inf, or unformatted raw floats), JSON schema validity, token math, and language/terminology integrity.
* **gofmt scope**: run `gofmt` only on files the change touches. Several test files carry pre-existing drift, and a repo-wide sweep adds unrelated noise to feature diffs.

<!-- project-sync:begin -->
### Continuous Project Metadata Maintenance (`project.jsonc`)
This repository maintains an authoritative `project.jsonc` at its root describing its title, tagline, version, capabilities, features, installation commands, tech stack, and FAQs for jadmadi.net synchronization.
Whenever you implement a new feature, add a CLI command or flag, bump the version, or modify installation instructions:
- You MUST update `project.jsonc` at the repository root in the same commit.
- Update `lastEdit` to the current local timestamp (`YY-MM-DD-HH-MM`).
- Keep `version`, `features`, `stats`, and `install_command` strictly synchronized with the code.
<!-- project-sync:end -->

<!-- sila:begin (managed by sila; content inside is regenerated by `sila install` / `sila agents sync`) -->
<!-- sila:meta rev=1 scope=core hash=a57df412ba43bc74 -->
## 0. Safety Rails (NEVER Do This)
- No unprompted remote push: Local commits are autonomous; remote push requires user consent.
- No unauthorized repo creation or re-homing: NEVER run 'gh repo create' or push external specs to personal namespaces; fail closed on auth hurdles.

## 1. Working Relationship & Discipline
- Concise and direct.
- Terminal Markdown: Keep visible line and table width under 80 chars; no hard breaks inside bullets or links.
- Decision Brief & `/adoo` Receipts: Lead with 1-line verdict ("X happens because Y, confirmed by Z"), compact receipt/risk table, file links, and next action. No 1-5 scores.
- Verified Repo Runbook Primacy: Never guess commands; verify exit code 0 locally.
- Closed-Loop Goal Attribution (6 Steps) for `goals/<slug>/`:
  1. Check every completed task `- [x]` in `goals/<slug>/plan.md`.
  2. Write markdown evidence with a standalone `Verdict: PASS` line.
  3. Commit with `(goals/<slug>/goal.md)` or `[goal: <slug>]` under `feat:`, `fix:`, `refactor:`, `perf:`, or `chore:` (`docs:`/`test:` ignored).
  4. Certify via `sila goals certify <slug> --evidence=goals/<slug>/evidence.md` or `sila goals audit <slug>`.
  5. Run `sila goals` to reconcile `goals/README.md` and commit receipts.
  6. Verify `sila goals --json` reports `status: "implemented"`, `verification_status: "verified"`, `unverified_count: 0`, and clean `git status`.

## Sila Trigger Contract

Memory writes are required by the goal contract and enforced by `sila gate`.
- `sila prime` (start) → briefing + goals.
- `sila recall "<symptom>"` (stuck) → prior resolutions.
- `sila memo "<directive>" --files=...` (decision) → grounded directive.
- `sila lesson "<symptom>" --cause=... --fix=...` (solved) → grounded lesson.
- `sila case "<intent>" --approach=... --invariant=...` (pattern) → case record.
- `sila facts verify` (grounding) → file/ledger evidence.
- `sila gate` (before commit) → deterministic quality receipt.
- `sila handoff "<summary>" --open=...` (end) → handoff record.
- `sila sync` (after changes) → cross-harness continuity.
- `sila salam`, `sila guide`, `docs/reference.md` → detail.
- Prior research/specs (BEFORE grep/find): `sila search "<q>" --omni` (`omni_search`), `sila search expand <id>` (`get_message_context`), `sila doc search "<q>"` (`search_docs`).
- Activity/domains/review/skills: `sila standup last 24h`, `sila digest --since=24h`, `sila recap`; `sila guide` (`get_operational_guidebook`), `sila learn`, `sila persona`, `sila hire --for-goal=<slug>`, `sila audit` (`audit_goal`), `sila skill` (`skill_search`, `skill_load`).
- Symbols/decisions: `sila edit-check --symbol=<sym>`, `sila safe-delete --symbol=<sym>`, `sila note --symbol=<sym>`, `sila decision` (`create_decision`), `sila vision triage`.

---

## Stack-Specific Standards (Go)

- 1. Code Review & Hygiene: maintain pure Go standard library with zero third-party dependencies, Pure-Go CGO-Free Portability with CGO_ENABLED=0, and hard file ceilings under 1,000 LOC.
- 2. Goroutine & Context Safety: require context.Context on all background routines; declare defer cancel() immediately after timeout context creation; ensure goroutine leak prevention.
- 3. Goal Scaffolding: break findings into discrete single-outcome packages under goals/<slug>/ with goal.md, facts.md, plan.md, and explicit dependencies.
- 4. Test-Driven Implementation: claim goal locks before editing; write table-driven tests verifying both acceptance assertions and negative rejection paths.
- 5. Quality Gates & Race Detector Discipline: verify changes locally with linter scripts and full race detection ('go test -race ./...') with zero tolerance for data races.
- 6. Attributed Commits: use semantic prefixes containing the goal slug: feat(scope): desc [goal: <slug>]; run sila goals audit to certify receipts.
- 7. Git Branch Discipline: develop on dedicated feature branches; never push unprompted to remotes; re-verify race tests on merged main.
- 8. Security & Trust Boundaries: validate all CLI flags and stdin readers; enforce UID ownership; use mode 0600 on sensitive temporary storage.
- 9. Release Lifecycle & Value Changelog: implement a 24-hour cached daily version check, offer direct upgrades, provide a human-readable benefits-focused changelog, and credit contributors.

---
<!-- sila:end -->
