# Facts: Coverage-First Distribution: Homebrew Tap, Standalone Installer, Agent Docs & Curated Showcases

## 1. Context & Interview Profile

* **Ecosystem**: Standalone Go CLI binary (cross-platform Linux, macOS, Windows via GoReleaser), `go install github.com/jadmadi/thermal/cmd/thermal@latest`, GitHub Releases.
* **Adopter Type**: **Both** — individual developers bottom-up for local terminal habit, leading to team/enterprise telemetry & tracking.
* **Primary Objective**: Active individual users & terminal footprint (organic word-of-mouth & daily habit).
* **Maintainer Capacity**: **2–4 hours/month** (lean, set-once foundation, automated release hygiene).
* **Horizon**: **Open horizon** (sustainable compounding engine with no arbitrary deadline).
* **Refused Channels**: High-maintenance recurring channels (manual social grinding, paid ads, messy multi-distro upstream packaging).

## 2. Invariants & Conflict Avoidance with Existing Ready Goals

1. **`repo-alignment-and-security` Boundary**:
   - `SECURITY.md` and `.github/workflows/ci.yml` vulnerability scanning are **strictly owned** by `goals/repo-alignment-and-security/`.
   - This distribution goal MUST NOT touch `SECURITY.md` or `.github/workflows/ci.yml`. Trust posture is delegated cleanly to `repo-alignment-and-security`.

2. **`purge-legacy-mit-releases` Coordination**:
   - `goals/purge-legacy-mit-releases/` retires pre-AGPL-3.0 releases from GitHub and resets the release line to AGPL-3.0.
   - The GoReleaser Homebrew tap formula and the standalone `install.sh` script must download and verify assets strictly against the active, post-purge release line (e.g. `v0.12.0`+) and enforce the `AGPL-3.0-only` license identifier in the Homebrew formula.

3. **`community-covenant-neutrality` Harmony**:
   - Licensing described in distribution surfaces (`llms.txt`, `docs/DISTRIBUTION.md`, Homebrew formula) must harmonize with the dual-licensing scope established in `DUAL-LICENSE.md`: GNU AGPLv3 for community use and Commercial Enterprise for proprietary embedding.

4. **`version-migration-guide` Concurrency**:
   - `version-migration-guide` owns `docs/MIGRATION.md` and updates `docs/pages/index.html`.
   - To avoid file collisions, this distribution goal isolates its installer to `docs/pages/install.sh`, machine specs to `docs/pages/llms.txt` and `docs/pages/llms-full.txt`, and distribution runbooks to `docs/DISTRIBUTION.md`. Any edits to `README.md` or `docs/pages/index.html` are coordinated and non-destructive.

5. **Resource Budget Safety Rail**:
   - Recurring maintenance must not exceed **1.5 hours/month** (<60% of the 2–4 hr/mo capacity), reserving the remaining hours for core tool loader additions and bug fixes.

## 3. File & Interface Contracts

* **`.goreleaser.yml`**: Adds `brews` block targeting `jadmadi/homebrew-tap` (or `jadmadi/tap`), generating formula with binary installation, test execution (`thermal version`), and AGPL-3.0 license declaration.
* **`docs/pages/install.sh`**: Pure POSIX `/bin/sh` shell script. Auto-detects OS (`darwin`, `linux`, `windows/mingw`), architecture (`amd64`, `arm64`), downloads matching release from GitHub Releases API, verifies SHA-256 against `checksums.txt`, and installs to `/usr/local/bin` (fallback to `~/.local/bin`).
* **`docs/pages/llms.txt`**: Concise markdown standard for AI coding agents and search engines, summarizing Thermal's purpose, terminal UI features, commands (`thermal`, `thermal <tool>`, `stats`, `mix`, `trend`, `replay`, `projects`, `models`), supported tools (Devin, Claude, OpenCode, Codex, Grok, Agy, etc.), and installation routes.
* **`docs/pages/llms-full.txt`**: Exhaustive specification with full CLI flags, loader architectures, token disjoint calculation rules, and JSON schema outputs.
* **`docs/DISTRIBUTION.md`**: Maintainer reference document containing:
  - Exact ready-to-submit PR diffs for `awesome-go` and `awesome-ai-tools`.
  - GitHub repo metadata checklist (updating topics: `claude`, `grok`, `deepseek`, `hermes`, `llm-metrics`).
  - Quarterly channel review sweep schedule and signal tracking metrics.
