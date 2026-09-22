# Thermal Migration Guide

This guide documents breaking changes, removed aliases, flag validation updates, and schema migrations across versions of **Thermal** (`github.com/jadmadi/thermal`).

Thermal follows Semantic Versioning (`vMAJOR.MINOR.PATCH`). In accordance with our CLI stability contract, every breaking change conforms to a strict 6-part entry format with exact detection signals, search patterns, before/after examples, fixes, and escape hatches.

---

## Table of Contents
- [Staged Deprecations (v0.14.0)](#staged-deprecations-v0140)
  - [Deprecated: `--license` CLI Flag](#deprecated---license-cli-flag)
  - [Deprecated: `nous` Tool Alias](#deprecated-nous-tool-alias)
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
