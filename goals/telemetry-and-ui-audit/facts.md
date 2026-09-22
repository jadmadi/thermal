# Facts: Telemetry Accuracy, Pricing Estimates, and UI Polish

## Architectural Invariants & Root Causes

### 1. Cost Estimates on Leaderboard & Single-Tool Dashboard
- **Root Cause**:
  - `internal/render/leaderboard.go` only checked `r.Summary.Cost > 0`. If 0, it rendered `—`.
  - `internal/render/dashboard.go` only printed `summary.Cost > 0`. If 0, no cost information was printed.
- **Invariant**:
  - Recorded cost always wins over estimated cost.
  - When recorded cost is 0, estimate cost using `internal/pricing` based on `r.Daily` model breakdown.
  - Prefix estimated costs with `~` (e.g. `~$0.07`).
  - Honor `--no-estimate` to omit estimates if requested.
  - On the single-tool dashboard, show `~$0.07 spent (est)` in the extra summary line.

### 2. `--help` Listing Missing `dsh`
- **Root Cause**:
  - `cmd/thermal/main.go` has a hardcoded text block in `printUsage()` and flag text in `flag.StringVar(&opts.Tool, ...)` that omitted `dsh`.
- **Invariant**:
  - Every tool registered in `loaders.AllTools()` must be listed in `Supported tools` in `--help`.

### 3. Display Name: `DeepSeek (DSH)`
- **Root Cause**:
  - `internal/loaders/registry.go` set `Name: "dsh"`.
- **Invariant**:
  - Tool display names are title-cased and recognizable (e.g., `OpenCode`, `MiMoCode`, `Codex`, `Devin`, `Agy`).
  - Canonical display name for `thermal.ToolDsh` is `DeepSeek (DSH)`.

### 4. Activity-Only Tools Leaking Into Token Analytics (`mix`, `stats`, `trend`)
- **Root Cause**:
  - In `AggregateToolMix`, `AggregateStats`, and `AggregateTrend`, when `metric == "tokens"`, `float64(day.Tokens)` was summed without checking `isActivityOnly(day)`.
  - Tools like Agy (steps) and Droid (messages) record counts into `day.Tokens == int64(day.Turns)` with zero token breakdown fields.
  - In April 2026, 417 steps from an activity tool appeared as "417 tokens" in `thermal mix` under "other (8)".
  - In `thermal stats`, 56 days of activity-only steps created a giant spike at bucket 0.
- **Invariant**:
  - Per AGENTS.md: *"Activity-only rows, where Tokens equals Turns and every other field is empty, are isActivityOnly and never enter token totals: a step count is not a token."*
  - `mix`, `stats`, and `trend` must filter out `isActivityOnly(day)` when `metric == "tokens"`.

### 5. `thermal stats` UI & Layout Polish
- **Root Cause**:
  - `RenderStats` used `strings.Repeat("#", barWidth)` with wide awkward padding.
  - `Weekday profile` was plain text with no visual bars or relative weight.
  - `Top days` and `Outliers` printed duplicate lists if top days met the outlier threshold.
- **Invariant**:
  - Thermal UI standard: Unicode block bars (`█` / `·`), high-design typography, clean alignment.
  - Weekday profile should feature mini-bars representing relative volume.
  - Outliers should be presented cleanly without verbatim duplication of Top Days.
