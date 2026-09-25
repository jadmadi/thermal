# Thermal Migration Guide

This guide documents breaking changes, removed aliases, flag validation updates, and schema migrations across versions of **Thermal** (`github.com/jadmadi/thermal`).

Thermal follows Semantic Versioning (`vMAJOR.MINOR.PATCH`). In accordance with our CLI stability contract, every breaking change conforms to a strict 6-part entry format with exact detection signals, search patterns, before/after examples, fixes, and escape hatches.

---

## Table of Contents
- [Staged Deprecations (v0.14.0)](#staged-deprecations-v0140)
  - [Deprecated: `--license` CLI Flag](#deprecated---license-cli-flag)
  - [Deprecated: `nous` Tool Alias](#deprecated-nous-tool-alias)
- [Upgrading to v0.14.x](#upgrading-to-v014x)
  - [Changed: Local Web Dashboard Request Authority Boundary](#changed-localhost-request-boundary)
  - [Changed: Live Monitor Metric Parity and Burn Accounting](#changed-live-metric-parity)
  - [Changed: Web Telemetry Parity and Filtering Options](#changed-web-telemetry-parity)
  - [Changed: Work Receipt Usage Accounting and Partial Coverage Disclosure](#changed-receipt-accounting-coverage)
  - [Changed: Verifiable Work Receipt Evidence Classification](#changed-receipt-evidence-classification)
  - [Changed: Default Stats View to 9-Box FinOps Grid](#changed-default-stats-view)
- [Upgrading to v0.12.x](#upgrading-to-v012x)
  - [Changed: Workload Replay JSON Payload Schema](#changed-workload-replay-json-payload-schema)
- [Upgrading to v0.11.x](#upgrading-to-v011x)
  - [Changed: Disjoint Token Metrics Accounting](#changed-disjoint-token-metrics-accounting)
- [Upgrading to v0.7.x](#upgrading-to-v07x)
  - [Removed: Ambiguous `cc` Tool Alias](#removed-ambiguous-cc-tool-alias)
  - [Changed: Strict Report Option Validation](#changed-strict-report-option-validation)

---

## Staged Deprecations (v0.14.0)

### Deprecated: `--license` CLI Flag
<a id="deprecated-license-flag"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal.*--license' .`
- **Detection signal**: Warning on stderr: `thermal: warning: --license is deprecated and will be removed in v0.14.0; use thermal license (see docs/MIGRATION.md#deprecated-license-flag)`

**What changed and why**:
Thermal previously accepted `--license` as a top-level flag to display licensing and commercial terms. In accordance with standard CLI command hierarchy, licensing information is now a dedicated subcommand `thermal license`. The `--license` flag is staged for deprecation in v0.13.x and will be removed in v0.14.0.

**Before / After**:
```bash
# Before (Deprecated syntax)
thermal --license

# After (v0.13.0+)
thermal license
```

**The fix**:
Replace `--license` with the `thermal license` subcommand in scripts, CI jobs, and documentation:
```bash
thermal license
```

**Escape hatch**:
The `--license` flag continues to function during the deprecation window with exit code 0.

---

### Deprecated: `nous` Tool Alias
<a id="deprecated-tool-nous"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal.*[[:space:]]nous[[:space:]]' .` or scripts running `thermal --tool nous`.
- **Detection signal**: Warning on stderr: `thermal: warning: nous is deprecated and will be removed in v0.14.0; use hermes (see docs/MIGRATION.md#deprecated-tool-nous)`

**What changed and why**:
The tool alias `nous` was previously used to refer to the Nous Hermes telemetry source. To align tool naming with the canonical project name `hermes`, the `nous` and `nous-hermes` aliases are staged for deprecation in v0.13.x and will be removed in v0.14.0. The canonical identifier `hermes` should be used instead.

**Before / After**:
```bash
# Before (Deprecated alias)
thermal nous weekly

# After (v0.13.0+)
thermal hermes weekly
```

**The fix**:
Use `hermes` for Nous Hermes sessions across all reports and analytics commands:
```bash
thermal hermes weekly
```

**Escape hatch**:
The `nous` alias continues to resolve to Hermes during the deprecation window with exit code 0.

## Upgrading to v0.14.x

### Changed: Local Web Dashboard Request Authority Boundary
<a id="changed-localhost-request-boundary"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal.*serve' .` or reverse proxy configurations forwarding to Thermal's local web dashboard or telemetry API.
- **Detection signal**: HTTP 403 Forbidden with payload `{"error":"forbidden: untrusted request authority"}` when accessing `http://<host>:<port>/` or `/api/*` with a foreign or unconfigured Host header, cross-site origin, or invalid authority port.

**What changed and why**:
Thermal's embedded web dashboard (`thermal serve`) previously accepted incoming HTTP requests regardless of the `Host` or `Origin` header values, leaving local telemetry endpoints susceptible to DNS rebinding and cross-site browser exfiltration. The server now strictly validates request authority against loopback interfaces (`127.0.0.1`, `localhost`, `[::1]`) and explicitly configured listening hosts (`--host`), structurally parses authorities, and rejects foreign names, invalid ports, and cross-site browser requests with HTTP 403 Forbidden before loading telemetry data. Reverse proxies and custom host setups that forward foreign Host headers must now preserve loopback authority or explicitly specify their target hostname using `--host`.

**Before / After**:
```bash
# Before (Wildcard or custom hostname without --host allowed arbitrary foreign Host headers)
curl -H "Host: untrusted.example" http://127.0.0.1:8080/api/telemetry
# HTTP/1.1 200 OK (returned private local telemetry)

# After (Untrusted host authorities and foreign origins rejected before data collection)
curl -H "Host: untrusted.example" http://127.0.0.1:8080/api/telemetry
# HTTP/1.1 403 Forbidden
# {"error":"forbidden: untrusted request authority"}
```

**The fix**:
For standard local browsing, connect directly to `http://localhost:8080` or `http://127.0.0.1:8080`. When running behind a reverse proxy or accessing across an intranet via a custom domain name, pass the exact domain via `--host`:
```bash
thermal serve --host dashboard.internal.net --port 8080
```
Alternatively, configure your reverse proxy (e.g. Nginx or Caddy) to preserve or rewrite the Host header to `localhost` or `127.0.0.1:8080` when forwarding upstream.

**Escape hatch**:
Explicitly specify your custom domain or network interface using `--host <hostname-or-ip>`. Wildcard binding (`--host 0.0.0.0`) allows direct IP access from LAN devices while strictly continuing to block arbitrary DNS rebinding names.

### Changed: Live Monitor Metric Parity and Burn Accounting
<a id="changed-live-metric-parity"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal.*live' .` or automated consumers of `thermal live --json` or `thermal live --json --stream`.
- **Detection signal**: `todayTokens` and `sessionTokens` exclude activity-only counts (e.g. Agy steps) and match canonical daily totals; `todayCacheHit` correctly computes `CacheRead / (Input + CacheRead + CacheWrite)` rather than inflating on cache-write-only events; turns delta increments across turns within existing sessions instead of tracking session count; baseline initialization, counter resets, and new sources no longer emit spurious historical token bursts.

**What changed and why**:
Thermal's real-time live monitor previously summed all daily rows directly, which inadvertently converted activity-only step counts (such as Agy actions) into token burn and inflated volume. Cache hit rate combined cache write into the numerator, reporting non-zero hit rate on cache-write-only prompts. Completed turns tracked session counts (`Summary.Sessions`), failing to observe turns added inside existing sessions, and newly appearing sources or truncated counters emitted historical spikes. Live tracking now enforces canonical token-only filtering, computes cache hit rate from `CacheRead / (Input + CacheRead + CacheWrite)`, tracks observed turns across daily rows, handles resets/first appearance safely, and honors `--no-estimate`.

**Before / After**:
```bash
# Before (Agy steps counted as tokens in todayTokens and burst deltas)
thermal live --json
# Output: {"todayTokens": 7, "todayTurns": 7, "todayCacheHit": 50.0}

# After (Activity-only tools generate 0 token burn; turns track observed turns; cache hit rate reflects reads)
thermal live --json
# Output: {"todayTokens": 0, "todayTurns": 7, "todayCacheHit": 0.0}
```

**The fix**:
Update monitoring integrations and automated parsers of `thermal live --json` and `thermal live --json --stream` to expect exact agreement between `todayTokens` and canonical `thermal daily` or `thermal stats` token totals. Rely on `todayTurns` for tracking agent prompt turn volume across both token and activity-only tools.

**Escape hatch**:
To inspect raw step counts for activity-only tools, use `thermal agy daily` or inspect `toolTotals` in the live JSON snapshot which preserves per-tool lifetime activity counts.

---

### Changed: Web Telemetry Parity and Filtering Options
<a id="changed-web-telemetry-parity"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal.*\(serve\|web\)' .` or automated clients consuming `/api/telemetry` or `thermal serve --json`.
- **Detection signal**: Activity-only steps (e.g. Agy) are excluded from `totalTokens` in web telemetry matching `thermal stats`; unsupported flags (e.g. `--stream`, unknown tools) or HTTP query parameters (e.g. `?tool=bad`, `?unsupported=1`) return HTTP 400 Bad Request or CLI exit code 1; `recordedCost`, `estimatedCost`, and `unpricedTokens` are distinguished in payloads.

**What changed and why**:
Thermal previously summed summary tokens and daily rows independently for the web dashboard, erroneously inflating `totalTokens` by including activity-only step counts. Additionally, collection options like `--no-estimate`, `--tool`, and date windows were ignored by the web server. Web telemetry collection has been unified through the canonical `thermal.AggregateTelemetry` adapter, ensuring exact parity with `thermal stats` and `thermal projects`, threading collection options cleanly, and rejecting unsupported options explicitly.

**Before / After**:
```bash
# Before (Agy step count inflated totalTokens; --no-estimate was ignored)
thermal serve --json
# Output: totalTokens included activity steps; estimatedCost always calculated

# After (totalTokens strictly matches thermal stats; options honored)
thermal serve --json --no-estimate
# Output: totalTokens reflects only token tools; estimatedCost: 0
```

**The fix**:
Update HTTP consumers and scripts targeting `thermal serve` or `/api/telemetry` to expect exact metric parity with `thermal stats`. If filtering telemetry, pass supported query parameters (`?tool=...`, `?since=...`, `?until=...`, `?last=...`, `?no-estimate=true`) and remove any unsupported query parameters.

**Escape hatch**:
To inspect raw step counts for activity-only tools, use `thermal agy daily` or query `/api/leaderboard` which preserves per-tool lifetime activity.

---

### Changed: Work Receipt Usage Accounting and Partial Coverage Disclosure
<a id="changed-receipt-accounting-coverage"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal.*receipt' .` or automated parsers consuming `thermal receipt --json`.
- **Detection signal**: Activity-only sessions (such as Agy) report 0 tokens instead of synthesized 500-tokens-per-step estimates; tools without session transcripts (OpenCode, Devin, etc.) are included as `Tier 3 (UNVERIFIED)` aggregate-only rows; `summary.unpricedTokens` is populated when models lack pricing catalog entries.

**What changed and why**:
Thermal previously synthesized 500 equivalent tokens per step for activity-only tools and suppressed all aggregate-only sources whenever any transcript file was detected on disk. To maintain strict usage accounting and audit fidelity, activity steps are preserved as zero tokens, uncovered tools are explicitly merged as `Tier 3 (UNVERIFIED)` rows, and Claude message tokens are deduplicated by message ID. In addition, `summary.unpricedTokens` is exposed in JSON payloads to prevent unpriced models from falsely implying free compute.

**Before / After**:
```bash
# Before (Agy synthesized 500 tokens/step; OpenCode was suppressed if Claude existed)
thermal receipt --json
# Output: { "tool": "Agy", "tokens": 500 }, OpenCode missing from report

# After (Accurate recorded tokens; all active tools merged into receipts)
thermal receipt --json
# Output: { "tool": "Agy", "tokens": 0 }, { "tool": "OpenCode", "status": "UNVERIFIED", "tier": "Tier 3 (Unverified)" }
```

**The fix**:
Update consumers of `thermal receipt` to expect zero token volume for step-only tools and inspect `unpricedTokens` when evaluating model pricing completeness. Use canonical tool aliases (`thermal ccode receipt`, `thermal claude receipt`) with confidence that aliases resolve consistently across all reports.

**Escape hatch**:
To inspect only verified sessions, pass `--sort verified` or filter for `"status": "VERIFIED"` in JSON output. To view step counts for activity tools, use `thermal agy daily` or `thermal leaderboard`.

---

### Changed: Verifiable Work Receipt Evidence Classification
<a id="changed-receipt-evidence-classification"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal receipt' .` or automated CI pipelines verifying factual session outcomes with `thermal receipt --json`.
- **Detection signal**: Sessions with interrupted or trailing commands at EOF or mixed test outputs (e.g. `1 passed, 2 failed`) no longer qualify as `Tier 1 (VERIFIED)` and instead produce `Tier 2 (CLAIMED)` or `Tier 3 (FAILED / UNVERIFIED)`.

**What changed and why**:
Thermal previously defaulted unresolved or trailing commands to exit status 0 and checked positive substring indicators ahead of failure markers. To maintain strict factual evidence integrity, command execution outcomes must be definitively observed and correlated against specific tool invocation IDs. Mixed test results prioritize failures (`exit 1`), non-execution commands (`cat`, `echo`, `git tag`) are excluded from evidence, and aborted commands without tool results at EOF are rejected from Tier 1 verification.

**Before / After**:
```bash
# Before (Interrupted command at EOF was optimistically marked VERIFIED)
thermal receipt --json
# Output: { "session_id": "...", "status": "VERIFIED", "tier": "Tier 1 Verified", "tests_passed": 1 }

# After (Observed result required; uncompleted command produces UNVERIFIED or CLAIMED)
thermal receipt --json
# Output: { "session_id": "...", "status": "UNVERIFIED", "tier": "Tier 3 Unverified", "tests_passed": 0 }
```

**The fix**:
Ensure AI agent sessions run test suites and linters to completion with explicit exit code telemetry or clean passing logs. If reviewing interrupted sessions or agent prose claims without executed tests, inspect `Tier 2 (CLAIMED)` and `Tier 3 (UNVERIFIED)` rows via `thermal receipt`.

**Escape hatch**:
Sessions with verified git commits or zero-exit test runs continue to qualify as `Tier 1 (VERIFIED)`. Unverified or speculative sessions can still be inspected across all tiers using `thermal receipt` without filtering.

---

### Changed: Default Stats View to 9-Box FinOps Grid
<a id="changed-default-stats-view"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal stats' .` or automated parsers scraping `thermal stats` output without flags.
- **Detection signal**: Terminal or JSON output begins with the 9-box FinOps grid (`Today`, `Taxonomy`, `Yield`, `MCP`, etc.) instead of the distribution histogram.

**What changed and why**:
Thermal previously rendered a 12-line token/cost distribution histogram as the default output for `thermal stats`, requiring the explicit `--dense` flag to view the high-density FinOps telemetry grid. Based on user feedback and FinOps analytics priorities, `thermal stats` now defaults to the interactive 9-box FinOps grid with reactive multi-period navigation (Today, 7D, 30D, Month, 6M, Lifetime) and project filtering. Scripts expecting the classical distribution histogram can pass `--distribution` (or `--dist`) to preserve the historical output.

**Before / After**:
```bash
# Before (v0.13.x and earlier)
thermal stats         # rendered 12-line distribution histogram
thermal stats --dense # rendered 9-box FinOps grid

# After (v0.14.0+)
thermal stats                # renders 9-box FinOps grid by default
thermal stats --distribution # renders classical distribution histogram (also --dist)
```

**The fix**:
If scripts or dashboards rely on the legacy token/cost distribution histogram, append `--distribution` (or shorthand `--dist`) to the invocation:
```bash
thermal stats --distribution
thermal stats --dist --no-color
```

**Escape hatch**:
Use `--distribution`, `--dist`, or `--sparse` with `thermal stats` (e.g. `thermal stats --dist` or `thermal stats --distribution --json`) to restore the legacy distribution histogram format and JSON schema.

---

### Renamed: `thermal serve` to `thermal web`
<a id="renamed-thermal-serve-to-web"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal.*serve' .`
- **Detection signal**: Warning or documentation referencing `thermal serve`.

**What changed and why**:
`thermal serve` has been renamed to `thermal web` to better communicate its role as an embedded localhost dashboard and live SSE token telemetry interface. In addition, `thermal web` now provides `/api/stream` for real-time telemetry updates and can be exited gracefully via `q` as well as `Ctrl+C`. `thermal serve` remains supported as an exact alias for full backwards compatibility.

**Before / After**:
```bash
# Before (v0.13.x and earlier)
thermal serve --port 8080

# After (v0.14.0+)
thermal web --port 8080
```

**The fix**:
Update CLI invocations and documentation from `thermal serve` to `thermal web`:
```bash
thermal web
```

**Escape hatch**:
`thermal serve` continues to be supported as an alias with identical flag parsing, exit codes, and behavior.

---

### Replaced: `cursor-pro` with Modern AI Subscriptions in Replay
<a id="replay-modern-subscriptions"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'cursor-pro' .` or invocations using `--compare ...cursor-pro...`.
- **Detection signal**: `thermal: error: unknown plan "cursor-pro"`.

**What changed and why**:
The obsolete `cursor-pro` subscription tier has been removed from `thermal replay`. It has been replaced with modern subscription models: `agy-pro` (Antigravity Pro, $20/mo, 50M limit), `opencode-pro` (OpenCode Pro, $20/mo, 40M limit), `zcode-pro` (ZCode Pro, $20/mo, 35M limit), and `kimi-k2` (Moonshot Kimi K2, $15/mo, 60M limit). Default comparison runs across `claude-pro`, `agy-pro`, `opencode-pro`, `kimi-k2`, and `deepseek-api`.

**Before / After**:
```bash
# Before
thermal replay --compare claude-pro,cursor-pro

# After
thermal replay --compare claude-pro,agy-pro,opencode-pro,kimi-k2
```

**The fix**:
Replace `cursor-pro` in `--compare` or `--against` flags with any of the active subscriptions: `agy-pro`, `opencode-pro`, `zcode-pro`, or `kimi-k2`.

**Escape hatch**:
Custom subscription limits can be simulated using `--against` with specific model price cards.

---

### Unified: `thermal receipt` Outcome Column Replacing Redundant Status and Tier
<a id="receipt-outcome-column"></a>

**Affected if**:
- **Search pattern**: Automated scrapers matching exact column names `Status` and `Tier` from `thermal receipt`.
- **Detection signal**: Output table displays `Outcome` header with semantic badges (`[PASS] Verified`, `[CLAIM] Untested`, `[FAIL] Broken`).

**What changed and why**:
`thermal receipt` previously rendered two separate, duplicative columns: `Status` (e.g. `[VERIFIED]`) and `Tier` (e.g. `Tier 1`). This forced users to cross-reference multiple columns to understand verification state. These have been consolidated into a single semantic `Outcome` column: `[PASS] Verified` (Tier 1: tests/linters passed with exit 0), `[CLAIM] Untested` (Tier 2: session claimed work, no tests found), and `[FAIL] Broken` (Tier 3: tests/linters failed). In addition, project paths in the table are now automatically resolved to clean directory slugs via `thermal.ProjectDisplayNames`.

**Before / After**:
```bash
# Before: Table columns "Status" and "Tier"
...  Evidence / Tests       Status        Tier
...  go test (exit 0)       [VERIFIED]    Tier 1

# After: Consolidated semantic "Outcome" column
...  Evidence / Tests       Outcome
...  go test (exit 0)       [PASS] Verified
```

**The fix**:
Update terminal output scrapers to expect the `Outcome` column or use `thermal receipt --json` for structured verification metrics.

**Escape hatch**:
`thermal receipt --json` preserves full structured `status` and `tier` fields in JSON objects.

---

### Decoupled: `thermal stats` and `thermal dashboard`
<a id="decoupled-stats-dashboard"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal stats' .`
- **Detection signal**: `thermal stats` completes immediately without launching an interactive Bubble Tea terminal buffer.

**What changed and why**:
Previously, running `thermal stats` launched an interactive Bubble Tea fullscreen session, competing with `thermal dashboard`. In accordance with Unix pipeline philosophy, `thermal dashboard` is now the single interactive TUI entry point (with support for sub-views like `thermal dashboard stats`), while `thermal stats` outputs a non-interactive, zero-allocation FinOps grid directly to `stdout` in `<10ms`.

**Before / After**:
```bash
# Before: launched interactive TUI
thermal stats

# After:
thermal stats           # outputs static FinOps grid to stdout in <10ms
thermal dashboard stats # launches interactive TUI focused on stats tab
```

**The fix**:
Use `thermal dashboard` (or `thermal dashboard stats`) when an interactive terminal interface is desired.

**Escape hatch**:
None required; `thermal stats` continues to provide full FinOps telemetry without terminal hijacking.

---

## Upgrading to v0.12.x

### Changed: Workload Replay JSON Payload Schema
<a id="v0-12-replay-schema"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal replay.*--json' .` or scripts parsing `plan_results` in `thermal replay --json`.
- **Error signal**: Scripts expecting a flat cost array crash or receive `null` for old replay top-level keys.

**What changed and why**:
`thermal replay --json` introduced commercial subscription tier simulation (Claude Pro/Max, ChatGPT Plus/Pro, Cursor Pro) with multi-model rate-card evaluations. To support capacity assessments, the JSON output schema was upgraded to expose structured capacity verdicts, headroom percentages, and monthly cost differences per tier.

**Before / After**:
```bash
// Before (v0.11.0 schema mock)
{
  "summary": { "total_tokens": 1500000 },
  "targets": [
    { "name": "claude-pro", "estimated_cost_usd": 20.00 }
  ]
}
```

```bash
// After (v0.12.0 schema)
{
  "summary": { "total_tokens": 1500000, "active_days": 28 },
  "targets": [
    {
      "name": "Claude Pro ($20/mo)",
      "target_type": "subscription",
      "verdict": "PASS",
      "monthly_cost_usd": 20.00,
      "monthly_delta_usd": -12.50,
      "headroom_pct": 34.2
    }
  ]
}
```

**The fix**:
Update downstream JSON processing scripts (such as `jq`) to query `.verdict`, `.monthly_delta_usd`, and `.headroom_pct`:
```bash
thermal replay --json | jq -r '.targets[] | "\(.name): \(.verdict) (\(.headroom_pct)% headroom)"'
```

**Escape hatch**:
None. The upgraded JSON schema is backward-additive for existing keys (`total_tokens`, `name`) while enriching subscription assessment fields.

---

## Upgrading to v0.11.x

### Changed: Disjoint Token Metrics Accounting
<a id="v0-11-disjoint-tokens"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'total_tokens.*cache' .` or custom tooling computing `Input + Cache = Total`.
- **Error signal**: Mathematical discrepancies where `Input + Output + CacheRead` exceeded reported `TotalTokens` by the cache amount.

**What changed and why**:
Certain AI tools (Codex, Grok, ZCode) nest reasoning tokens inside output tokens and cache-read tokens inside input tokens in their raw telemetry. Starting in v0.11.0, Thermal enforces strict token disjointness at the loader boundary: nested tokens are subtracted so that `Input`, `Output`, `Reasoning`, `CacheRead`, and `CacheWrite` are mutually disjoint and add up precisely to `Total`.

**Before / After**:
```bash
// Before (Nested accounting in raw logs)
Input: 10,000 (includes 8,000 cache read)
Output: 2,000 (includes 500 reasoning)
Sum: 12,000 (disagrees with 4,000 net new tokens)
```

```bash
// After (Disjoint accounting in v0.11.0+)
Input (uncached): 2,000
CacheRead:        8,000
Output:           1,500
Reasoning:          500
Sum:             12,000 (exact mathematical partition)
```

**The fix**:
Do not subtract cache read tokens from input tokens in external analysis scripts; Thermal now presents them pre-separated.

**Escape hatch**:
None. Strict disjointness is an architectural invariant required for accurate catalog cost estimation.

---

## Upgrading to v0.7.x

### Removed: Ambiguous `cc` Tool Alias
<a id="v0-7-removed-cc-alias"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal.*[[:space:]]cc[[:space:]]' .`
- **Error signal**: `thermal: unknown tool "cc". Run "thermal --help" to see available tools.`

**What changed and why**:
The short alias `cc` was previously ambiguous between `claude` (Anthropic Claude Code CLI) and `command-code`. To prevent conflicting telemetry attribution and unexpected scans, `cc` was retired.

**Before / After**:
```bash
// Before (v0.6.0)
thermal cc daily --last 7
```

```bash
// After (v0.7.0+)
thermal claude daily --last 7
# OR for command-code:
thermal commandcode daily --last 7
```

**The fix**:
Use `claude` for Claude Code transcripts or `commandcode` (`ccode`, `cmd`) for command-code transcripts:
```bash
thermal claude daily --last 7
```

**Escape hatch**:
Shell alias in your `.bashrc` or `.zshrc`:
```bash
alias "thermal cc"="thermal claude"
```

---

### Changed: Strict Report Option Validation
<a id="v0-7-strict-report-flags"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal[[:space:]].*--last' .` or scripts running `thermal --last 7` without a report subcommand.
- **Error signal**: `thermal: error: --last is only applicable to reports (daily, weekly, monthly, projects, models). Exit code 1.`

**What changed and why**:
In v0.6.0 and earlier, report flags passed to the root leaderboard command (`thermal --last 7` or `thermal --breakdown`) were silently ignored while still printing the all-time leaderboard. In v0.7.0+, Thermal strictly validates CLI arguments to prevent users from believing a filtered view is displayed when it is not.

**Before / After**:
```bash
// Before (v0.6.0: silently printed default all-time leaderboard)
thermal --last 7
```

```bash
// After (v0.7.0+: exits with code 1)
thermal daily --last 7
```

**The fix**:
Provide an explicit report subcommand (`daily`, `weekly`, `monthly`, `projects`, `models`) when passing filtering or grouping flags:
```bash
thermal daily --last 7
```

**Escape hatch**:
None. Strict flag validation protects against deceptive metrics and silent misconfiguration.
