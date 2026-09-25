<p align="center">
  <img src="thermal.webp" alt="Thermal Gopher Mascot" width="160" />
</p>

# Thermal — Contribution Heatmap for AI Coding Tools

Terminal heatmap, streaks, tokens, and cost for 14 AI coding tools. Thermal reads the session data your agents already wrote. One static binary, zero config.

- **Website**: https://jadmadi.net/projects/thermal/
- **Repository**: https://github.com/jadmadi/thermal
- **License**: Dual-licensed under AGPL-3.0 (Community Edition) and Commercial Enterprise License

---

## Quickstart

Install Thermal with a single command:

```bash
curl -fsSL https://jadmadi.net/thermal/install.sh | sh
```

Or via Homebrew:

```bash
brew install jadmadi/tap/thermal
```

Or build from source with Go 1.23+:

```bash
go install github.com/jadmadi/thermal/cmd/thermal@latest
```

---

## Common Invocations

```bash
# Render contribution leaderboard across all discovered AI coding tools
thermal

# Inspect specific tool activity, token metrics, and heatmaps
thermal opencode
thermal claude
thermal devin
thermal codex

# Multi-view interactive Terminal User Interface (TUI)
thermal dashboard

# Breakdown usage by project and repository
thermal projects

# Model usage distribution and catalog pricing
thermal models

# Analytics, workload mix, and trend charts
thermal mix --chart
thermal stats --chart
thermal stats --dense
thermal trend --chart

# Replay historical workload against subscription caps
thermal replay --against "Claude Pro"
thermal replay --against "ChatGPT Plus"

# Code generation token yield and net line efficiency
thermal yield --sort lines

# Verifiable work outcome proofs (test passes, linter runs)
thermal receipt --sort verified

# Local setup and context health diagnostic
thermal audit

# Stateless zero-database streak share card
thermal share

# Embedded local web dashboard
thermal serve --open

# Self-update binary in place
thermal upgrade
```

---

## Supported Tools

Thermal automatically ingests activity and disjoint token metrics from 14 AI coding environments:

| Tool | Category | Token Telemetry | Project Attribution |
| :--- | :--- | :--- | :--- |
| **Devin** | Token Warrior | Input, Output, Cache Read/Write | Working directory |
| **OpenCode** | Token Warrior | Input, Output, Reasoning, Cache Read/Write | Session directory / git worktree |
| **MiMoCode** | Token Warrior | Input, Output, Reasoning, Cache Read/Write | Session directory |
| **Codex** | Token Warrior | Input, Output, Reasoning, Cache Read/Write | Thread working directory |
| **codewhale** | Token Warrior | Total tokens, Recorded USD cost | Session workspace |
| **ZCode** | Token Warrior | Input, Output, Cache Read | Working directory |
| **Grok** | Token Warrior | Input, Output, Reasoning, Cache Read | Git root directory |
| **Claude** | Token Warrior | Input, Output, Cache Read/Write | Per-message working directory |
| **DeepSeek (DSH)** | Token Warrior | Input, Output, Cache Read/Write | Identity current working directory |
| **Nous Hermes** | Token Warrior | Input, Output, Reasoning, Cache Read/Write | Session working directory |
| **Muse** | Activity Hunter | Prompt & turn counts | Session index |
| **Droid** | Activity Hunter | Turn & message counts | Message logs |
| **command-code** | Activity Hunter | Message counts, model distributions | Workspace directory |
| **Agy** | Activity Hunter | Step activity, model selections | CLI session logs |

---

## Core Capabilities

<p align="center">
  <img src="thermal2.webp" alt="Thermal Go Gopher Mascot with HUD" width="180" />
</p>

### 1. Terminal Heatmaps & Streaks
- Visualizes daily activity and streak continuity directly in your terminal using ANSI 256-color palettes.
- Supports `--no-color` for plain monochrome output and screen readers.
- Computes current streak, longest streak, and active days across all tools.

### 2. Disjoint Token Accounting
- All token types (`Input`, `Output`, `Reasoning`, `CacheRead`, `CacheWrite`) are strictly disjoint and sum to total volume.
- Nested structures from tools (e.g. cache inside input or reasoning inside output) are automatically subtracted during loader ingestion.

### 3. Workload Replay & Capacity Planning (`thermal replay`)
- Replays recorded daily token distributions against commercial subscription ceilings (Claude Pro, Claude Max, ChatGPT Plus/Pro, Cursor Pro).
- Outputs capacity status (`PASS`, `DEGRADED`, `FAIL`) and monthly cost deltas.

### 4. Interactive TUI Dashboard (`thermal dashboard`)
- Fullscreen Bubble Tea terminal interface with keyboard navigation.
- Views include Leaderboard, Projects, Models, Mix, and Trends.

### 5. Zero Telemetry & Privacy Invariant
- 100% local operation: scans local databases and logs in strictly read-only mode (`mode=ro`).
- Zero network telemetry, zero background daemons, zero external analytics.

### 6. Self-Update & Release Checks (`thermal upgrade`)
- Checks GitHub Releases and atomically updates the running binary in place.
- Daily background check caches the latest tag in `~/.cache/thermal/update.json` and notifies users on stderr in interactive sessions.
- Fully controllable via `--no-update-check` or `THERMAL_NO_UPDATE_CHECK=1`.

### 7. Token Yield Telemetry (`thermal yield`)
- Evaluates code generation productivity from git diffs and patches.
- Calculates Token Yield (tokens burned per net line of code added) with efficiency classification ratings (`[HIGH]`, `[BALANCED]`, `[VERBOSE]`, `[EXPLORATORY]`).

### 8. Verifiable Work Receipts (`thermal receipt`)
- Parses tool-level test execution (`go test`, `pytest`, `cargo test`, `npm test`) and linter proofs (`golangci-lint`, `eslint`, `ruff`).
- Provides 3-tier confidence classification (`Tier 1 Verified`, `Tier 2 Claimed`, `Tier 3 Failed`) with zero prompt or privacy leakage.

### 9. High-Density FinOps Grid (`thermal stats --dense`)
- Responsive 9-box Bubble Tea / Lip Gloss terminal grid.
- Decomposes agent activity across Coding, Debugging, Testing, and Exploration, while tracking prompt cache savings and MCP tool overhead.

### 10. Embedded Local Web Dashboard (`thermal serve`)
- Embedded single-page web application served on localhost (`127.0.0.1:8080`) with assets compiled directly into the binary via `go:embed`.
- Exposes read-only JSON telemetry APIs under strict Content-Security-Policy headers with zero cloud tracking.

### 11. Environment Health Audit & Stateless Sharing (`thermal audit` / `share`)
- `thermal audit` inspects local SQLite permissions, 256MB mmap status, and pricing cache integrity.
- `thermal share` encodes streak and telemetry metadata into compact, zero-database stateless share URLs.

---

## Documentation & Project Links

- **Documentation**: https://jadmadi.net/projects/thermal/
- **Repository**: https://github.com/jadmadi/thermal
- **License Terms**: Dual-licensed under GNU AGPL-3.0 (Community Edition) and Commercial Enterprise License
- **Governance & Scope**: [GOVERNANCE.md](https://github.com/jadmadi/thermal/blob/main/GOVERNANCE.md)
- **Migration Guide**: [docs/MIGRATION.md](https://github.com/jadmadi/thermal/blob/main/docs/MIGRATION.md)
- **Contributing & CLA**: [CONTRIBUTING.md](https://github.com/jadmadi/thermal/blob/main/CONTRIBUTING.md)
