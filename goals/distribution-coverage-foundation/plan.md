# Plan: Coverage-First Distribution: Homebrew Tap, Standalone Installer, Agent Docs & Curated Showcases

## Execution Tasks

1. [x] **Phase 1 — Automated Homebrew Tap in `.goreleaser.yml`**:
   - Add `brews:` section to `.goreleaser.yml` targeting repository `jadmadi/homebrew-tap` (or `jadmadi/tap`).
   - Define formula metadata: `name: "thermal"`, `homepage: "https://thermal.jadmadi.net"`, `description: "Terminal contribution heatmap and leaderboard for AI coding assistants"`, `license: "AGPL-3.0-only"`.
   - Specify installation block: install binary `thermal`, test block verifying `thermal version`.
   - Add commit author and commit message template following Conventional Commits (`chore(brew): bump formula to {{ .Tag }}`).

2. [x] **Phase 2 — Zero-Dependency POSIX Installer (`docs/pages/install.sh`)**:
   - Write standard POSIX `/bin/sh` shell installer supporting `curl -fsSL https://thermal.jadmadi.net/install.sh | sh`.
   - Auto-detect operating system (`linux`, `darwin`) and machine architecture (`x86_64` -> `amd64`, `aarch64`/`arm64` -> `arm64`).
   - Query GitHub Releases API for latest release tag (fallback to explicit `VERSION` env var).
   - Download the matching archive (`thermal_${VERSION}_${OS}_${ARCH}.tar.gz`) and `checksums.txt`.
   - Verify SHA-256 checksum using `sha256sum`, `shasum -a 256`, or `openssl dgst -sha256`.
   - Extract binary and atomically install to destination directory (checks write permissions on `/usr/local/bin`, falls back to `${HOME}/.local/bin` and warns if not in `$PATH`).
   - Provide flags: `--dry-run`, `--version <v>`, `--prefix <dir>`.
   - Add local automated test script verifying shell syntax with `shellcheck` (or bash POSIX mode).

3. [x] **Phase 3 — Machine & Agent Documentation (`docs/pages/llms.txt` & `llms-full.txt`)**:
   - Implement `docs/pages/llms.txt` according to the `/llms.txt` proposed web standard:
     - Project title, tagline, canonical URLs.
     - Quickstart install options (`brew install jadmadi/tap/thermal`, `curl ... | sh`, `go install ...`).
     - Available commands (`thermal`, `daily`, `weekly`, `monthly`, `projects`, `models`, `mix`, `stats`, `trend`, `replay`, `dashboard`).
     - Supported AI coding tools and their telemetry sources (Devin, OpenCode, Codex, Grok, Claude, Agy, etc.).
   - Implement `docs/pages/llms-full.txt` providing complete command line flag references, JSON output specifications, and loader data models.

4. [x] **Phase 4 — Curated Showcases & Distribution Operations (`docs/DISTRIBUTION.md`)**:
   - Create `docs/DISTRIBUTION.md` documenting Thermal's ongoing distribution runbook:
     - Ready-to-copy PR submission format for `awesome-go` (under "Command Line" / "Advanced Console UIs").
     - Ready-to-copy PR submission format for `awesome-ai-tools`.
     - GitHub topic synchronization guide to ensure discoverability for modern agent harnesses (`claude`, `grok`, `deepseek`, `hermes`, `llm-metrics`).
     - Referral traffic tracking parameters (`?ref=awesome-go`, `?utm_source=brew`, etc.).
     - Quarterly listing and decay sweep schedule (checking for broken links or outdated descriptions).

5. [x] **Phase 5 — Verification & Gate Release Checklist**:
   - Test `.goreleaser.yml` syntax using `goreleaser check`.
   - Execute `docs/pages/install.sh` in dry-run mode and verify checksum verification logic on Linux and macOS test fixtures.
   - Run `go test -race ./...` and `./scripts/simulated_user_gate.sh`.

6. [x] **Phase 6 — Attribution Commit**:
   - Commit with message: `feat(distribution): configure homebrew tap, installer script, agent docs & distribution runbook (goals/distribution-coverage-foundation/goal.md)`.
   - Run `sila goals` to confirm goal status.
