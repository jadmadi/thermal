# Thermal Distribution & Discovery Operations Runbook

This document defines the operational procedures, curated list submission packages, repository topic tags, and quarterly review calendar for **Thermal** (`github.com/jadmadi/thermal`).

---

## 1. Distribution Strategy & Channels

Thermal employs a **Coverage-First** distribution model focused on zero-friction developer adoption, terminal habit formation, and machine discoverability.

### Primary Distribution Channels
1. **Homebrew Tap (`jadmadi/homebrew-tap`)**:
   - Automated formula generation on release via `.goreleaser.yml`.
   - Install: `brew install jadmadi/tap/thermal`
2. **Standalone POSIX Shell Installer**:
   - Zero-dependency script hosted on GitHub Pages: `https://thermal.jadmadi.net/install.sh`.
   - Install: `curl -fsSL https://thermal.jadmadi.net/install.sh | sh`
   - Features: Architecture auto-detection, SHA-256 verification against `checksums.txt`, atomic installation to `/usr/local/bin` or `~/.local/bin`.
3. **Go Toolchain**:
   - Direct module compilation: `go install github.com/jadmadi/thermal/cmd/thermal@latest`.
4. **Machine & Agent Discoverability**:
   - Web standard `llms.txt`: `https://thermal.jadmadi.net/llms.txt`
   - Exhaustive specification: `https://thermal.jadmadi.net/llms-full.txt`

---

## 2. Curated List PR Packages

When submitting Thermal to curated indexes, use these exact, pre-validated snippets.

### A. Awesome Go (`avelino/awesome-go`)
- **Target Category**: `Command Line` -> `Advanced Console UIs` (or `Utilities`)
- **Format**: Markdown alphabetical list entry with license tag
- **Submission Snippet**:
```markdown
* [thermal](https://github.com/jadmadi/thermal) - High-performance terminal contribution heatmap, streak tracker, and leaderboard tool for AI coding assistants. [![AGPL-3.0](https://img.shields.io/badge/License-AGPL%203.0-blue.svg)](https://opensource.org/licenses/AGPL-3.0)
```
- **PR Description**:
```markdown
### What does this PR do?
Adds Thermal to Advanced Console UIs / Command Line.

Thermal is a terminal contribution heatmap, streak tracker, and telemetry leaderboard tool for AI coding assistants (Claude Code, Devin, OpenCode, Codex, Grok, DeepSeek, Hermes, and more). It scans local SQLite databases and transcripts in <10ms, calculates disjoint token accounting, and renders heatmaps and analytics in the terminal.

- Repository: https://github.com/jadmadi/thermal
- License: AGPL-3.0 (with commercial dual-licensing)
- Tests: 100% PASS with -race detector and comprehensive simulated user gates
```

### B. Awesome AI Tools (`mahseema/awesome-ai-tools` or `steven2358/awesome-generative-ai`)
- **Target Category**: `Developer Tools` / `Monitoring & Observability`
- **Submission Snippet**:
```markdown
* [Thermal](https://github.com/jadmadi/thermal?ref=awesome-ai-tools) - Zero-telemetry terminal contribution heatmap and leaderboard tracking developer token usage across Claude Code, Devin, OpenCode, Codex, and Grok.
```

---

## 3. GitHub Repository Topics Synchronization

Maintain the following topic tags on `github.com/jadmadi/thermal` to ensure discoverability across AI agent ecosystems:

```text
ai-agents, anthropic-claude, claude-code, devin, opencode, codex-cli, grok, deepseek,
nous-hermes, llm-metrics, token-tracking, terminal-heatmap, developer-tools, bubbletea, tui,
terminal-streak, go-cli
```

### Verification Command
Run GitHub CLI to synchronize topics:
```bash
gh repo edit jadmadi/thermal --add-topic "claude-code,deepseek,nous-hermes,llm-metrics,token-tracking,terminal-heatmap"
```

---

## 4. Referral Traffic & Attribution Tracking

To measure channel performance without invading user privacy or transmitting runtime cloud telemetry, attribution tracking is strictly confined to inbound web and link referrers on documentation surfaces:

| Channel | URL Parameter | Destination Surface |
| :--- | :--- | :--- |
| Awesome Go | `?ref=awesome-go` | `https://github.com/jadmadi/thermal` |
| Awesome AI Tools | `?ref=awesome-ai` | `https://thermal.jadmadi.net` |
| Homebrew Tap | `?utm_source=homebrew` | `https://thermal.jadmadi.net` |
| Standalone Script | `?utm_source=install-sh` | `https://thermal.jadmadi.net` |

---

## 5. Quarterly Listing & Decay Sweep Schedule

To maintain high directory standing and prevent broken documentation links, execute the following sweep every 90 days:

1. **Link & Release Check**:
   - Verify `https://thermal.jadmadi.net/install.sh` downloads the latest release asset.
   - Verify `jadmadi/homebrew-tap` formula builds and passes `brew test thermal`.
   - Verify `llms.txt` and `llms-full.txt` match all currently supported tools and CLI flags.
2. **Curated Directory Health**:
   - Audit external PRs and star counts on listings.
   - Update descriptions if newly supported AI agents are introduced.
3. **Decay Schedule Calendar**:
   - **Q1 Sweep**: January 15
   - **Q2 Sweep**: April 15
   - **Q3 Sweep**: July 15
   - **Q4 Sweep**: October 15
