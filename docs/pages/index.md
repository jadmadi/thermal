# Thermal — Contribution Heatmap for AI Coding Tools

Terminal heatmap, streaks, tokens, and cost for 14 AI coding tools. Thermal reads the session data your agents already wrote. One static binary, zero config.

- **Website**: https://thermal.jadmadi.net/
- **Repository**: https://github.com/jadmadi/thermal
- **License**: Dual-licensed under AGPL-3.0 (Community Edition) and Commercial Enterprise License

---

## Quickstart

Install Thermal with a single command:

```bash
curl -fsSL https://thermal.jadmadi.net/install.sh | sh
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
thermal trend --chart

# Replay historical workload against subscription caps
thermal replay --against "Claude Pro"
thermal replay --against "ChatGPT Plus"
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

---

## Agent & Machine Discovery Endpoints

Thermal publishes standards-compliant discovery endpoints for autonomous AI agents:

- **API Catalog (RFC 9727)**: `/.well-known/api-catalog`
- **Agent Skills Discovery (RFC v0.2.0)**: `/.well-known/agent-skills/index.json`
- **MCP Server Card**: `/.well-known/mcp/server-card.json`
- **ARD Capability Manifest**: `/.well-known/ai-catalog.json`
- **ACP Commerce Discovery**: `/.well-known/acp.json`
- **OAuth Authorization Server**: `/.well-known/oauth-authorization-server`
- **OAuth Protected Resource (RFC 9728)**: `/.well-known/oauth-protected-resource`
- **Universal Commerce Protocol (UCP)**: `/.well-known/ucp`
- **x402 Payment Protocol**: `/.well-known/x402`
- **Agent Registration Documentation**: `/auth.md`
- **DNS for AI Discovery**: `/DNS-AID.md`
- **Markdown for Agents**: Content negotiation via `Accept: text/markdown` on `/` returns this document.
