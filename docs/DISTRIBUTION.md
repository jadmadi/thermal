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
5. **Cryptographic Release Provenance (SLSA Level 3)**:
   - Built and attested via OpenSSF generic SLSA3 generator inside ephemeral GitHub Actions runners.
   - Non-falsifiable in-toto provenance (`checksums.txt.intoto.jsonl`) generated using Sigstore keyless OIDC (Fulcio + Rekor transparency log).
   - Independent verification via `slsa-verifier` without private keys or vendor lock-in.

---

## 2. Supply Chain Security & SLSA Level 3 Verification

Thermal releases provide cryptographic build provenance conforming to **SLSA Build Level 3** (Supply-chain Levels for Software Artifacts) via the OpenSSF SLSA GitHub Generator (`slsa-framework/slsa-github-generator`). This satisfies enterprise OSPO compliance requirements, NIST SSDF guidelines, and US Executive Order 14028.

### Transitive Trust Architecture

```text
GitHub Actions Runner (Isolated / Ephemeral)
  ├── 1. GoReleaser builds binaries & generates checksums.txt
  └── 2. OpenSSF generic_slsa3 generator signs checksums.txt via Sigstore OIDC
        └── Produces checksums.txt.intoto.jsonl (Fulcio certificate + Rekor transparency entry)

User Verification:
  slsa-verifier (verifies checksums.txt against checksums.txt.intoto.jsonl)
    └── sha256sum verifies release archive / binary against verified checksums.txt
```

### Independent Verification Runbook

Anyone can cryptographically verify that a downloaded Thermal release binary or archive was built from the official source repository and workflow without tampering:

```bash
# 1. Install official slsa-verifier
go install github.com/slsa-framework/slsa-verifier/v2/cli/slsa-verifier@latest

# 2. Download release archive, checksums, and in-toto provenance attestation
VERSION="0.6.0"
curl -fsSLO "https://github.com/jadmadi/thermal/releases/download/v${VERSION}/thermal_${VERSION}_linux_amd64.tar.gz"
curl -fsSLO "https://github.com/jadmadi/thermal/releases/download/v${VERSION}/checksums.txt"
curl -fsSLO "https://github.com/jadmadi/thermal/releases/download/v${VERSION}/checksums.txt.intoto.jsonl"

# 3. Verify checksums.txt against SLSA Level 3 provenance
slsa-verifier verify-artifact checksums.txt \
  --provenance-path checksums.txt.intoto.jsonl \
  --source-uri github.com/jadmadi/thermal \
  --source-tag "v${VERSION}"

# 4. Verify downloaded archive against the verified checksums
sha256sum --check --ignore-missing checksums.txt
```

When verified, `slsa-verifier` confirms that the artifacts were built on GitHub Actions runners initiated from a release tag on `github.com/jadmadi/thermal` and have not been altered.

---

## 3. Curated List PR Packages

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

## 4. GitHub Repository Topics Synchronization

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

## 5. Referral Traffic & Attribution Tracking

To measure channel performance without invading user privacy or transmitting runtime cloud telemetry, attribution tracking is strictly confined to inbound web and link referrers on documentation surfaces:

| Channel | URL Parameter | Destination Surface |
| :--- | :--- | :--- |
| Awesome Go | `?ref=awesome-go` | `https://github.com/jadmadi/thermal` |
| Awesome AI Tools | `?ref=awesome-ai` | `https://thermal.jadmadi.net` |
| Homebrew Tap | `?utm_source=homebrew` | `https://thermal.jadmadi.net` |
| Standalone Script | `?utm_source=install-sh` | `https://thermal.jadmadi.net` |

---

## 6. Quarterly Listing & Decay Sweep Schedule

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
