# Thermal Output Gallery

> Generated from a fixed fixture by `scripts/output_gallery.sh`. Capture
> `2026-10-29T20:48:06Z` · fixture `2db489728bc70f4013f5aec8b4fd805ab961414dbfe7f6c52a0619a8a4053e26` · `thermal dev (built from source)`.
> Fixed data date: 2026-01-15. Text is `NO_COLOR=1`; non-TTY default width is
> 100. Width sections use a PTY at 80/120. The live TUI block is synthetic and
> labeled; the other text blocks are captured from the same fixture run.

## Contents

- [Consistency matrix](#consistency-matrix)
- [Text surfaces](#text-surfaces)
  - [Global help](#global-help)
  - [Leaderboard](#leaderboard)
  - [Daily](#daily) · [Weekly](#weekly) · [Monthly](#monthly)
  - [Projects](#projects) · [Models](#models)
  - [Mix](#mix) · [Stats](#stats) · [Trend](#trend)
  - [Share](#share) · [Statusline](#statusline) · [Audit](#audit)
  - [Changelog](#changelog) · [License](#license) · [Live](#live)
- [Width and color contracts](#width-and-color-contracts)
- [Row-level JSON schemas](#row-level-json-schemas)
- [Validation cases](#validation-cases)
- [Appendices](#appendices)

## Consistency matrix

| Surface | Tokens | Recorded | Estimated | Total | Active/costed days | Window/scope | Basis |
|---|---:|---:|---:|---:|---|---|---|
| leaderboard | 759700 | 1.75 | 0.0046 | 1.7546 | 1/— | all tools | total |
| daily | 759700 | 1.75 | 0.0046 | 1.7546 | 1/— | all time (1 active day) | total |
| weekly | 759700 | 1.75 | 0.0046 | 1.7546 | — | all time (1 active week) | total |
| monthly | 759700 | 1.75 | 0.0046 | 1.7546 | — | all time (1 active month) | total |
| projects | 759700 | 1.75 | 0.0046 | 1.7546 | 2 project-days | all tools/projects | total |
| models | 759700 classified + 0 unclassified | 0 | 0.0046 | 0.0046 | 2 model-days | all models | estimated |
| mix (tool cost) | — | — | — | 1.7546 | — | all tools | total |
| stats (tokens) | 759700 | — | — | — | 1/— | all time | n/a (tokens) |
| stats (cost) | — | — | — | 1.7546 | 1/1 | all time | total |
| trend (tokens) | mean 759700 | — | — | — | — | calendar | n/a (tokens) |
| trend (cost) | mean 1.7546 | — | — | — | — | calendar | total |
| share | 759700 | 1.75 | 0.0046 | 1.7546 | 1/— | all · all time (1 active day) | recorded+estimated |
| live | — | — | — | — | — | today snapshot (live, 2026-10-29T23:48:07.100966823+03:00) | recorded+estimated |

Allowed equalities are checked by `scripts/output_gallery_gate.sh`: leaderboard
estimated equals daily estimated, projects total equals daily total, models
classified + unclassified equals daily total, tool-mix cost equals daily total,
and share token/cost fields use one window.

## Text surfaces

### Global help

```text

  Thermal · usage profile & activity heatmaps for AI coding tools

  Usage:
    thermal [options]                    All-tool leaderboard & streak heatmap
    thermal [command] [options]          Analytics, audit, and utility commands
    thermal [tool] [report] [options]    Scoped report (e.g. opencode weekly)
    thermal help [command]               Global help or one command's help page

  Dashboards & Views:
    live           Real-time token burn monitor with velocity, flame & sparkline
    statusline     Instant shell statusline and prompt helper (<3ms, alias: prompt)
    share          Generate a stateless, private URL card for your streak

  Reports & Rankings:
    daily          Daily token volume and cost breakdown
    weekly         Weekly aggregated activity and spend
    monthly        Monthly summary and model breakdown
    projects       Tokens and cost per project, folded to git root
    models         Tokens and estimated cost per model, ranked

  FinOps & Analytics:
    stats          Daily distribution, percentiles, and token composition
    mix            Tool and model market-share & switching frequency
    trend          Daily trend fit with month-end volume & spend projection

  Diagnostics & Utilities:
    help           Global help or one command's help page (e.g. thermal help daily)
    audit          Audit local agent instruction files & MCP context tax
    changelog      Show release history, newly added commands & arguments
    upgrade        Self-upgrade thermal to the latest release
    version        Print SemVer, commit hash, and build timestamp
    license        Display AGPL-3.0 and commercial enterprise licensing

  Common Examples:
    thermal                            All-tool leaderboard & streak heatmap
    thermal opencode weekly --chart    Weekly breakdown with terminal bar chart
    thermal daily --help               Scoped help for the daily report
    thermal help daily                 Same page via the help command
    thermal stats                      Daily distribution and token composition
    thermal audit                      Audit instruction files & MCP context tax
    thermal changelog --top 5          View latest new features & arguments

  Filter & Window Options:
    --tool <name>          Filter to a single tool (default: all)
    --interval <dur>       Polling interval for live monitor (default: 1s)
    --flame-ceiling <N>    Live flame ceiling in active units per minute (e.g. 9M or 60)
    --since <date>         Start date filter (YYYY-MM-DD or YYYYMMDD)
    --until <date>         End date filter (YYYY-MM-DD or YYYYMMDD)
    --last <N>             Limit to last N days, weeks, or months
    --order <asc|desc>     Sort order for report rows (default: desc)
    --weeks <N>            Heatmap width in weeks (default: 52)
    --start-of-week <day>  First day of week: sunday-saturday (default: sunday)

  Ranking & Analytics Options:
    --sort <field>         Leaderboard: streak|tokens|cost; projects: tokens|cost|days|recent;
                           models: tokens|cost
    --top <N>              Cap table rows to top N items (default: all)
    --metric <field>       Analytics metric: tokens or cost (default: tokens)
    --by <tool|model>      Mix breakdown dimension (default: tool)
    --grain <grain>        Mix window: day, week, or month (default: week)

  Display & Formatting Options:
    --chart                Charts for daily, weekly, monthly, projects, and models reports
    --breakdown            Show per-model token breakdown under each period
    --nerd                 Use Nerd Font flame glyph (󰈸) in statusline
    --plain                Plain text statusline output without emoji or glyphs
    --no-estimate          Recorded cost only; disable models.dev pricing estimates
    --format <text|json>   Output format (json is an alias for --json; md/markdown retired)

  Global & Engine Options:
    --db <path>            Read one tool's data from a file or directory (requires --tool)
    --offline              Use cached models.dev pricing catalog without network
    --fresh                Bypass cached statusline data; live sessions always start at launch
    --stream, -f           Stream continuous NDJSON live events (live verb)
    --json                 Output structured JSON (never launches a terminal UI)
    --no-color             Disable all ANSI escape sequences
    -V, --version          Print version, commit, and build timestamp
    --verbose, -v          Output diagnostic warnings to stderr
    -h, --help             Show global or command-scoped help

  Supported Tools:
    Token Warriors:    OpenCode · Devin · Codex · Claude · ZCode · Grok · MiMoCode (mimo) ·
                       codewhale · DeepSeek DSH (dsh) · Nous Hermes (hermes) · CMD
                       (command-code) · Zed
    Activity Hunters:  Agy (agy) · Muse (muse) · Droid (droid)

  Documentation & guides: https://jadmadi.net/project/thermal/ · AGPL-3.0 / Commercial
```

### Leaderboard

```text

  THERMAL  — Don't break the streak.

  ╭─ Token Warriors ────────────────────────────────────── tokens & spend ─╮
  │    #  Tool              Strk    Best    Days       Tokens         Cost │
  │  ───────────────────────────────────────────────────────────────────── │
  │   1.  codewhale           1d      1d      1d   750.0K tok        $1.75 │
  │   2.  Claude              1d      1d      1d     9.7K tok     ~$0.0046 │
  ╰────────────────────────────────────────────────────────────────────────╯

  ╭─ Activity Hunters ────────────────────── messages & steps ─╮
  │    #  Tool              Strk    Best    Days      Activity │
  │  ───────────────────────────────────────────────────────── │
  │   1.  Agy                 1d      1d      1d       12 step │
  ╰────────────────────────────────────────────────────────────╯

  >> codewhale is on fire with a 1-day streak!

  Recorded cost: $1.75 from 1 tools. Estimated: ~$0.0046 across 1 tools (~ prefix).
  12 tokens have no model recorded, so no price is applied to them.

  Keep the heat going. Don't break the streak.
```

Single tool: [codewhale](#single-tool-codewhale) · [Agy](#single-tool-agy)

### Single tool: codewhale

```text

  ╭─ codewhale Activity ─────────────────────────────────────────────────── ~/.codewhale/sessions ─╮
  │  750.0K tokens / 52 weeks   ·   1 active days  |  1 day streak  |  1 best  |  750.0K all-time  │
  │                                                                                                │
  │       Nov  Dec Jan Feb Mar  Apr May  Jun Jul Aug  Sep Oct                                      │
  │       ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░                                     │
  │   Mon ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░                                     │
  │       ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░                                     │
  │   Wed ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░                                     │
  │       ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░                                     │
  │   Fri ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░                                      │
  │       ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░                                      │
  │       Less □░▒▓█ More                                                                          │
  │  ───────────────────────────────────────────────────────────────────────────────────────────── │
  │  $1.75 spent  ·  1 sessions                                                                    │
  │  · agents  chat: 1                                                                             │
  ╰────────────────────────────────────────────────────────────────────────────────────────────────╯
```

### Single tool: Agy

```text

  ╭─ Agy Activity ────────────────────────────────────────── ~/.gemini/antigravity/brain ─╮
  │  12 steps / 52 weeks   ·   1 active days  |  1 day streak  |  1 best  |  12 all-time  │
  │                                                                                       │
  │       Nov  Dec Jan Feb Mar  Apr May  Jun Jul Aug  Sep Oct                             │
  │       ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░                            │
  │   Mon ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░                            │
  │       ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░                            │
  │   Wed ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░                            │
  │       ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░                            │
  │   Fri ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░                             │
  │       ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░                             │
  │       Less □░▒▓█ More                                                                 │
  │  ──────────────────────────────────────────────────────────────────────────────────── │
  │  1 sessions                                                                           │
  │  12 tokens have no model recorded, so no price is applied to them.                    │
  ╰───────────────────────────────────────────────────────────────────────────────────────╯
```

### Daily

```text

  Thermal · daily usage · all tools

  Window: [all time (1 active day) · all tools]  Total: [759.7K tok · ~$1.75]  Cache: [89.1% hit]

  ╭─ Daily Usage ─────────────────────────────────────────────────────────────── all tools ─╮
  │  Period      Models                   Input     Output      Cache     Tokens       Cost │
  │  ────────────────────────────────────────────────────────────────────────────────────── │
  │  2026-01-15  claude-3.5-son… +1        1.0K        500       8.2K     759.7K     ~$1.75 │
  │  ────────────────────────────────────────────────────────────────────────────────────── │
  │  Total                                 1.0K        500       8.2K     759.7K     ~$1.75 │
  ╰─────────────────────────────────────────────────────────────────────────────────────────╯

  Total = $1.75 recorded + ~$0.0046 estimated from pricing data.
```

### Weekly

```text

  Thermal · weekly usage · all tools

  Window: [all time (1 active week) · all tools]  Total: [759.7K tok · ~$1.75]  Cache: [89.1% hit]

  ╭─ Weekly Usage ────────────────────────────────────────────────────────────── all tools ─╮
  │  Period      Models                   Input     Output      Cache     Tokens       Cost │
  │  ────────────────────────────────────────────────────────────────────────────────────── │
  │  2026-01-11  claude-3.5-son… +1        1.0K        500       8.2K     759.7K     ~$1.75 │
  │  ────────────────────────────────────────────────────────────────────────────────────── │
  │  Total                                 1.0K        500       8.2K     759.7K     ~$1.75 │
  ╰─────────────────────────────────────────────────────────────────────────────────────────╯

  Total = $1.75 recorded + ~$0.0046 estimated from pricing data.
```

### Monthly

```text

  Thermal · monthly usage · all tools

  Window: [all time (1 active month) · all tools]  Total: [759.7K tok · ~$1.75]  Cache: [89.1% hit]

  ╭─ Monthly Usage ───────────────────────────────────────────────────────────── all tools ─╮
  │  Period      Models                   Input     Output      Cache     Tokens       Cost │
  │  ────────────────────────────────────────────────────────────────────────────────────── │
  │  2026-01     claude-3.5-son… +1        1.0K        500       8.2K     759.7K     ~$1.75 │
  │  ────────────────────────────────────────────────────────────────────────────────────── │
  │  Total                                 1.0K        500       8.2K     759.7K     ~$1.75 │
  ╰─────────────────────────────────────────────────────────────────────────────────────────╯

  Total = $1.75 recorded + ~$0.0046 estimated from pricing data.
```

### Projects

```text

  Thermal · projects

  ╭─ Projects ─────────────────────────────────────────────────────────────────── attribution & spend ─╮
  │    #  Project                       Tools                    Tokens       Cost  P-Days  Last       │
  │  ───────────────────────────────────────────────────────────────────────────────────────────────── │
  │   1.  repo-a                        codewhale                750.0K      $1.75       1  2026-01-15 │
  │   2.  repo-b                        Claude                     9.7K   ~$0.0046       1  2026-01-15 │
  │  ───────────────────────────────────────────────────────────────────────────────────────────────── │
  │       Total                         Claude · codewhale       759.7K     ~$1.75       2  2026-01-15 │
  ╰────────────────────────────────────────────────────────────────────────────────────────────────────╯
  Total = $1.75 recorded + ~$0.0046 estimated from pricing data.
```

Chart variant (`projects --top 2 --order asc --chart`):

```text

  Thermal · projects

  ╭─ Projects ─────────────────────────────────────────────────────────────────── attribution & spend ─╮
  │    #  Project                       Tools                    Tokens       Cost  P-Days  Last       │
  │  ───────────────────────────────────────────────────────────────────────────────────────────────── │
  │   1.  repo-b                        Claude                     9.7K   ~$0.0046       1  2026-01-15 │
  │   2.  repo-a                        codewhale                750.0K      $1.75       1  2026-01-15 │
  │  ───────────────────────────────────────────────────────────────────────────────────────────────── │
  │       Total                         Claude · codewhale       759.7K     ~$1.75       2  2026-01-15 │
  ╰────────────────────────────────────────────────────────────────────────────────────────────────────╯
  Total = $1.75 recorded + ~$0.0046 estimated from pricing data.


  Projects by tokens  bar = share of the largest charted row, 750.0K tok

  repo-b                    █···············································    9.7K tok
  repo-a                    ████████████████████████████████████████████████  750.0K tok
```

### Models

```text

  Thermal · models

  ╭─ Models ──────────────────────────────────────────────────────────────────────────── tokens & spend ─╮
  │    #  Model                           Tools                    Tokens       Cost  M-Days  Last       │
  │  ─────────────────────────────────────────────────────────────────────────────────────────────────── │
  │   1.  claude-3.5-sonnet               codewhale                750.0K          —       1  2026-01-15 │
  │   2.  claude-3.5-haiku                Claude                     9.7K   ~$0.0046       1  2026-01-15 │
  │  ─────────────────────────────────────────────────────────────────────────────────────────────────── │
  │       Total                           Claude · codewhale       759.7K   ~$0.0046       2  2026-01-15 │
  ╰──────────────────────────────────────────────────────────────────────────────────────────────────────╯
  Cost is estimated from models.dev list prices; recorded session cost is not attributable to one
  model.
```

Chart variant (`models --top 2 --order asc --chart`):

```text

  Thermal · models

  ╭─ Models ──────────────────────────────────────────────────────────────────────────── tokens & spend ─╮
  │    #  Model                           Tools                    Tokens       Cost  M-Days  Last       │
  │  ─────────────────────────────────────────────────────────────────────────────────────────────────── │
  │   1.  claude-3.5-haiku                Claude                     9.7K   ~$0.0046       1  2026-01-15 │
  │   2.  claude-3.5-sonnet               codewhale                750.0K          —       1  2026-01-15 │
  │  ─────────────────────────────────────────────────────────────────────────────────────────────────── │
  │       Total                           Claude · codewhale       759.7K   ~$0.0046       2  2026-01-15 │
  ╰──────────────────────────────────────────────────────────────────────────────────────────────────────╯
  Cost is estimated from models.dev list prices; recorded session cost is not attributable to one
  model.


  Models by tokens  bar = share of the largest charted row, 750.0K tok; ~ = estimated cost

  claude-3.5-haiku              █···············································      9.7K  ~$0.0046
  claude-3.5-sonnet             ████████████████████████████████████████████████    750.0K         —
```

### Mix

```text

  Thermal · mix · by tool · tokens

  ╭─ Tool Mix ───────────────────────────── tokens share ─╮
  │  Period        codewhale     Claude             Total │
  │  ──────────────────────────────────────────────────── │
  │  2026-01       99%           1%                759.7K │
  │  ──────────────────────────────────────────────────── │
  │  Total         99%           1%                759.7K │
  ╰───────────────────────────────────────────────────────╯

  Shares are rounded to the nearest percent.

  Dominant   codewhale
  Switches   0 
  Spread     0.97 (1 = one tool, 0 = even spread)
```

### Stats

Token stats:

```text

  Thermal · stats · tokens

  Active days  1
  Total        759.7K
  Mean         759.7K
  Median       759.7K
  p90          759.7K
  Max          759.7K

  Weekday profile
  Thursday        759.7K  ████████████████     1 day

  Token composition
  Cache read            8.0K  ████████████████   82.5%
  Cache write            200  █···············    2.1%
  Uncached input        1.0K  ██··············   10.3%
  Output                 500  █···············    5.2%
  Reasoning                0  ················    0.0%
  Unclassified             0  ················    0.0%
  Cache hit rate: 87.0% of prompt tokens read from cache

  Distribution
          0 — 76.0K  ························     0
     76.0K — 151.9K  ························     0
    151.9K — 227.9K  ························     0
    227.9K — 303.9K  ························     0
    303.9K — 379.9K  ························     0
    379.9K — 455.8K  ························     0
    455.8K — 531.8K  ························     0
    531.8K — 607.8K  ························     0
    607.8K — 683.7K  ························     0
    683.7K — 759.7K  ████████████████████████     1

  Top days
  2026-01-15       759.7K
```

Cost stats:

```text

  Thermal · stats · cost

  Active days  1
  Costed days  1
  Total        $1.75
  Mean         $1.75
  Median       $1.75
  p90          $1.75
  Max          $1.75

  Weekday profile
  Thursday         $1.75  ████████████████     1 day

  Distribution
     $0.00 — $0.18  ························     0
     $0.18 — $0.35  ························     0
     $0.35 — $0.53  ························     0
     $0.53 — $0.70  ························     0
     $0.70 — $0.88  ························     0
     $0.88 — $1.05  ························     0
     $1.05 — $1.23  ························     0
     $1.23 — $1.40  ························     0
     $1.40 — $1.58  ························     0
     $1.58 — $1.75  ████████████████████████     1

  Top days
  2026-01-15        $1.75

  Cost mixes recorded values with estimates from models.dev list prices.
```

### Trend

Token trend:

```text

  Thermal · trend · tokens

  Trajectory: [FLAT → steady state]  Calendar window: [1 days · 2026-01-15 to 2026-01-15]  Daily Mean: ≈759.7K

  ╭─ Activity Trend ────────────── tokens ─╮
  │  Range        2026-01-15 to 2026-01-15 │
  │  Days         1                        │
  │  Mean         759.7K                   │
  │  Slope        0 per day                │
  │  Direction    flat                     │
  ╰────────────────────────────────────────╯

  Actionable Insights:
  • Workload is holding steady at ≈759.7K daily with consistent utilization.
```

Cost trend:

```text

  Thermal · trend · cost

  Trajectory: [FLAT → steady state]  Calendar window: [1 days · 2026-01-15 to 2026-01-15]  Daily Mean: ≈$1.75

  ╭─ Activity Trend ──────────────── cost ─╮
  │  Range        2026-01-15 to 2026-01-15 │
  │  Days         1                        │
  │  Mean         $1.75                    │
  │  Slope        $0.00 per day            │
  │  Direction    flat                     │
  ╰────────────────────────────────────────╯

  Actionable Insights:
  • Workload is holding steady at ≈$1.75 daily with consistent utilization.

  Cost mixes recorded values with estimates from models.dev list prices.
```

### Share

```text

  Thermal · share · stateless streak card

  ╭─ Stateless Streak Card ── Client-Verified · Zero-Database ─╮
  │  Scope:         All                                        │
  │  Window:        all time (1 active day)                    │
  │  Coverage:      recorded+estimated                         │
  │  Streak:        1 days 🔥 (Longest: 1 days, token-active)  │
  │  Active Days:   1 token-active days                        │
  │  Total Volume:  759.7K tokens                              │
  │  Spend Profile: $1.75 incl. est.                           │
  ╰────────────────────────────────────────────────────────────╯


  Shareable Card URL:
  https://jadmadi.net/project/thermal/share#v1.71687d92.TI9RbsIwEE…ZuVbr9tOJ7BzOc9-IVTIpnSiIjj-DsA
  … full URL (358 chars) is available with --json

  • Privacy Guarantee: 100% client-only URL fragment. Zero database or cloud storage. No file
    paths, repository roots, or prompt telemetry encoded.
```

The full URL is in [the share JSON appendix](#full-share-json).

### Statusline

```text
🔥 1d all tools · 0 tok today · top: codewhale
```

### Audit

```text

  Thermal · audit · local setup & context health

  Score: [95/100 · Grade A+]  Context Tax: ≈344 tokens / turn (Rules: ≈344 tok, MCP: ≈0 tok)
  Score components: rules -5 · MCP -0 · readiness -0. Largest: rule files. Biggest win: Warning:
  AGENTS.md (repo root) is 160 lines (~333 tok). Prune stale rules to save context budget.

  ╭─ Instruction & Rule Files ─────────────────────────────────────────────────── rule overhead ─╮
  │  PASS  CLAUDE.md (repo root)              2 lines  ≈7 tok                                    │
  │  WARN  AGENTS.md (repo root)            160 lines  ≈333 tok                                  │
  │  PASS  CLAUDE.md (~/.claude)              1 lines  ≈4 tok                                    │
  │  ─────────────────────────────────────────────────────────────────────────────────────────── │
  │  Legend: [WARN] >150 lines (moderate context tax)                                            │
  ╰──────────────────────────────────────────────────────────────────────────────────────────────╯

  ╭─ MCP Server Registrations ────────────────────────────────────────────────────── schema tax ─╮
  │  No MCP servers configured.                                                                  │
  ╰──────────────────────────────────────────────────────────────────────────────────────────────╯

  ╭─ Agent Readiness & Repository Hygiene ─────────────────────────────────── environment check ─╮
  │  PASS  Instructions      AGENTS.md verified (160 lines)                                      │
  │  PASS  Gitignore         .gitignore rules properly anchored                                  │
  │  PASS  Delta Cache       Cache directory writable (probe file created and removed)           │
  │  INFO  Tool Databases    No local tool database files found; database access was not         │
  │                          assessed                                                            │
  ╰──────────────────────────────────────────────────────────────────────────────────────────────╯

  ╭─ Actionable Advice ────────────────────────────────────────────────────────── optimizations ─╮
  │  • Warning: AGENTS.md (repo root) is 160 lines (~333 tok). Prune stale rules to save context │
  │    budget.                                                                                   │
  ╰──────────────────────────────────────────────────────────────────────────────────────────────╯
```

### Changelog

```text

  Thermal · changelog · release history & new features

  Current: dev (source build)  ·  Showing 5 of 25 releases

  ╭─ v0.21.0 ─────────────────────────────────────────────────────────────────────── 2026-10-29 ─╮
  │  Features:                                                                                   │
  │   • pricing: add cost provenance reconciliation, model pricing disclosures, and subscription │
  │     usage labeling                                                                           │
  │   • loaders/muse: add Muse token and project day ingestion from session JSONL records        │
  │   • loaders/sqlite: enforce schema object validation, 256MB mmap limits, 60s query timeouts, │
  │     and 64MB store ceilings                                                                  │
  │                                                                                              │
  │  Bug Fixes:                                                                                  │
  │   • test: normalize Windows project path assertions and handle read-only file ACLs in audit  │
  │     tests                                                                                    │
  │   • ci: bump Go baseline to 1.27.3 to resolve crypto/x509 GO-2026-6809 in govulncheck        │
  ╰──────────────────────────────────────────────────────────────────────────────────────────────╯

  ╭─ v0.20.1 ─────────────────────────────────────────────────────────────────────── 2026-10-14 ─╮
  │  Features:                                                                                   │
  │   • loaders: rename command-code canonical display name to CMD                               │
  ╰──────────────────────────────────────────────────────────────────────────────────────────────╯

  ╭─ v0.20.0 ─────────────────────────────────────────────────────────────────────── 2026-10-12 ─╮
  │  Features:                                                                                   │
  │   • ci: add fast-path bypass to pre-push gate for docs and metadata diffs                    │
  │                                                                                              │
  │  Bug Fixes:                                                                                  │
  │   • loaders: exclude sqlite shm files to prevent live monitor CPU thrashing                  │
  │   • tui: clamp projection stats bar and band bounds against negative values                  │
  ╰──────────────────────────────────────────────────────────────────────────────────────────────╯

  ╭─ v0.19.0 ─────────────────────────────────────────────────────────────────────── 2026-10-08 ─╮
  │  Features:                                                                                   │
  │   • build: verify multi-target cross-compilation release gate                                │
  │   • ci: verify cross-platform operating system test matrix                                   │
  │   • docs: update documentation parity and simulated human user audit for 15 tools            │
  │   • gate: ensure hermetic verification and release gate semantic parity                      │
  │   • git: verify zero bot contributor attribution in git history                              │
  │   • loaders: ingest Zed editor assistant threads and tokens                                  │
  │   • sila: complete Sila onboarding and strict verification                                   │
  ╰──────────────────────────────────────────────────────────────────────────────────────────────╯

  ╭─ v0.18.0 ─────────────────────────────────────────────────────────────────────── 2026-10-03 ─╮
  │  Features:                                                                                   │
  │   • loaders: add platform-native application data directory discovery                        │
  │                                                                                              │
  │  Bug Fixes:                                                                                  │
  │   • loaders: normalize SQLite URI paths for cross-platform compatibility                     │
  │   • test: resolve cross-platform paths and line endings on windows and macos                 │
  ╰──────────────────────────────────────────────────────────────────────────────────────────────╯

  … and 20 more releases (use --top 25 to view all, or --json for full history)
```

### License

```text
Thermal — AI Agent Telemetry & Streak Engine
Copyright (C) 2026 Jad Madi. All rights reserved.

Open Source & Community Edition (GNU AGPLv3):
  Thermal is free and open-source software under the terms of the
  GNU Affero General Public License v3.0 (AGPL-3.0).
  Modifications distributed as binaries or served over a network
  (hosted dashboards, team leaderboards, SaaS) require full source
  code disclosure under AGPLv3 Section 13.

Community Reciprocity Covenant:
  Thermal guarantees that the core CLI, loaders, TUI, and telemetry engine
  will remain free software under GNU AGPL-3.0 in perpetuity. Community
  contributions will never be retroactively relicensed into proprietary
  or closed-source products without contributor consent.

Commercial Use License:
  $30 per developer per 12-month term for internal commercial use.
  Includes updates and community support; excludes redistribution,
  embedding, hosted services, and support SLAs.
  Buy:     https://paypal.me/Madise/30
  Page:    https://thermal.jadmadi.net/pricing.html
  Terms:   COMMERCIAL-LICENSE.md
  Invoice: contact@jadmadi.net

Enterprise & OEM:
  Proprietary embedding, closed-source hosted SaaS, fleet/security
  review, custom agent loaders, and support commitments.
  Contact: contact@jadmadi.net
  Terms:   DUAL-LICENSE.md

Licensing Inquiries:
  Email:   contact@jadmadi.net
  GitHub:  https://github.com/jadmadi/thermal

Third-Party Notices:
  Incorporates permissive open-source components (MIT, BSD-3, Public Domain).
  See NOTICES.md in the source repository for complete attributions.
```

### Live

Non-TTY pointer:

```text
thermal live is interactive and needs a terminal.
For scripts, pipelines, and status bars use:

  thermal live --json            instant JSON snapshot
  thermal live --json --stream   stream NDJSON live events
  thermal stats                  daily distribution and token composition
```

Representative synthetic TUI frame (80x24; units and labels are the contract,
values are illustrative):

```text
╭─ 🔥 THERMAL LIVE ─────────────────────────────── ● LIVE (1.0s) · 12:00:00 ─╮
│ BURN VELOCITY           TODAY'S USAGE           LIVE SESSION              │
│ Rate:     3.4M tok/m    Volume:   1.7M tok      Since launch: 1.6M tok   │
│ Speed:    56.7K tok/s   Cost:     ~$1.20        Spend:    ~$1.10          │
│                         Burned: +60 step (session)                     │
│ Burn/hr now: $0.00/hr   Cache hit: 96.1%        State:    IDLE 💤         │
│   Steps: 23 step/m · 0.383 step/s                                         │
│   Top model (est. cost): deepseek-v4.1-flash ~$3.95                       │
│   Today by project: thermal ~$4.43 · scripts ~$0.46                       │
│   Flame: [····················]   0% of 9.0M tok/m  IDLE 💤              │
│   Spark: [···▂▃▂···············]  Peak: 1.0K tok/s (1s bucket)            │
│   Last event: 0s ago · * >60s · ! >3x                                     │
│   CMD×3  1.7M TOK  60 STEP  MODEL · PROJECT  ~$0.30                       │
╰──────────────────────────────────────────────────────────────────────────╯
 • Controls: [Space] Pause · [c] Reset · [s] Sort: time · [?] Help · [q] Quit
```

At 120x40 the same frame keeps the full hybrid rate/speed units and controls;
the 80x24 frame uses the compact `Steps:` line when a KPI cell cannot fit the
secondary value. No unit is truncated.

## Width and color contracts

All captures below use `NO_COLOR=1`; no ANSI escapes are present. The body is
compact; the full fixture tables are small enough to keep inline.

The fixed report grid is 93 cells wide; at 80 it is intentionally not shrunk
because the columns are numeric values, not prose. Responsive surfaces (help,
stats, share, statusline) are captured at both widths.

### 80 columns: help

```text

  Thermal · usage profile & activity heatmaps for AI coding tools

  Usage:
    thermal [options]                    All-tool leaderboard & streak heatmap
    thermal [command] [options]          Analytics, audit, and utility commands
    thermal [tool] [report] [options]    Scoped report (e.g. opencode weekly)
    thermal help [command]               Global help or one command's help page

  Dashboards & Views:
    live           Real-time token burn monitor with velocity, flame & sparkline
    statusline     Instant shell statusline and prompt helper (<3ms, alias:
                   prompt)
    share          Generate a stateless, private URL card for your streak

  Reports & Rankings:
    daily          Daily token volume and cost breakdown
    weekly         Weekly aggregated activity and spend
    monthly        Monthly summary and model breakdown
    projects       Tokens and cost per project, folded to git root
    models         Tokens and estimated cost per model, ranked

  FinOps & Analytics:
    stats          Daily distribution, percentiles, and token composition
    mix            Tool and model market-share & switching frequency
    trend          Daily trend fit with month-end volume & spend projection

  Diagnostics & Utilities:
    help           Global help or one command's help page (e.g. thermal help
                   daily)
    audit          Audit local agent instruction files & MCP context tax
    changelog      Show release history, newly added commands & arguments
    upgrade        Self-upgrade thermal to the latest release
    version        Print SemVer, commit hash, and build timestamp
    license        Display AGPL-3.0 and commercial enterprise licensing

  Common Examples:
    thermal                            All-tool leaderboard & streak heatmap
    thermal opencode weekly --chart    Weekly breakdown with terminal bar chart
    thermal daily --help               Scoped help for the daily report
    thermal help daily                 Same page via the help command
    thermal stats                      Daily distribution and token composition
    thermal audit                      Audit instruction files & MCP context tax
    thermal changelog --top 5          View latest new features & arguments

  Filter & Window Options:
    --tool <name>          Filter to a single tool (default: all)
    --interval <dur>       Polling interval for live monitor (default: 1s)
    --flame-ceiling <N>    Live flame ceiling in active units per minute (e.g.
                           9M or 60)
    --since <date>         Start date filter (YYYY-MM-DD or YYYYMMDD)
    --until <date>         End date filter (YYYY-MM-DD or YYYYMMDD)
    --last <N>             Limit to last N days, weeks, or months
    --order <asc|desc>     Sort order for report rows (default: desc)
    --weeks <N>            Heatmap width in weeks (default: 52)
    --start-of-week <day>  First day of week: sunday-saturday (default: sunday)

  Ranking & Analytics Options:
    --sort <field>         Leaderboard: streak|tokens|cost; projects:
                           tokens|cost|days|recent; models: tokens|cost
    --top <N>              Cap table rows to top N items (default: all)
    --metric <field>       Analytics metric: tokens or cost (default: tokens)
    --by <tool|model>      Mix breakdown dimension (default: tool)
    --grain <grain>        Mix window: day, week, or month (default: week)

  Display & Formatting Options:
    --chart                Charts for daily, weekly, monthly, projects, and
                           models reports
    --breakdown            Show per-model token breakdown under each period
    --nerd                 Use Nerd Font flame glyph (󰈸) in statusline
    --plain                Plain text statusline output without emoji or glyphs
    --no-estimate          Recorded cost only; disable models.dev pricing
                           estimates
    --format <text|json>   Output format (json is an alias for --json;
                           md/markdown retired)

  Global & Engine Options:
    --db <path>            Read one tool's data from a file or directory
                           (requires --tool)
    --offline              Use cached models.dev pricing catalog without network
    --fresh                Bypass cached statusline data; live sessions always
                           start at launch
    --stream, -f           Stream continuous NDJSON live events (live verb)
    --json                 Output structured JSON (never launches a terminal UI)
    --no-color             Disable all ANSI escape sequences
    -V, --version          Print version, commit, and build timestamp
    --verbose, -v          Output diagnostic warnings to stderr
    -h, --help             Show global or command-scoped help

  Supported Tools:
    Token Warriors:    OpenCode · Devin · Codex · Claude · ZCode · Grok ·
                       MiMoCode (mimo) · codewhale · DeepSeek DSH (dsh) · Nous
                       Hermes (hermes) · CMD (command-code) · Zed
    Activity Hunters:  Agy (agy) · Muse (muse) · Droid (droid)

  Documentation & guides: https://jadmadi.net/project/thermal/ · AGPL-3.0 /
  Commercial
```

### 120 columns: help

```text

  Thermal · usage profile & activity heatmaps for AI coding tools

  Usage:
    thermal [options]                    All-tool leaderboard & streak heatmap
    thermal [command] [options]          Analytics, audit, and utility commands
    thermal [tool] [report] [options]    Scoped report (e.g. opencode weekly)
    thermal help [command]               Global help or one command's help page

  Dashboards & Views:
    live           Real-time token burn monitor with velocity, flame & sparkline
    statusline     Instant shell statusline and prompt helper (<3ms, alias: prompt)
    share          Generate a stateless, private URL card for your streak

  Reports & Rankings:
    daily          Daily token volume and cost breakdown
    weekly         Weekly aggregated activity and spend
    monthly        Monthly summary and model breakdown
    projects       Tokens and cost per project, folded to git root
    models         Tokens and estimated cost per model, ranked

  FinOps & Analytics:
    stats          Daily distribution, percentiles, and token composition
    mix            Tool and model market-share & switching frequency
    trend          Daily trend fit with month-end volume & spend projection

  Diagnostics & Utilities:
    help           Global help or one command's help page (e.g. thermal help daily)
    audit          Audit local agent instruction files & MCP context tax
    changelog      Show release history, newly added commands & arguments
    upgrade        Self-upgrade thermal to the latest release
    version        Print SemVer, commit hash, and build timestamp
    license        Display AGPL-3.0 and commercial enterprise licensing

  Common Examples:
    thermal                            All-tool leaderboard & streak heatmap
    thermal opencode weekly --chart    Weekly breakdown with terminal bar chart
    thermal daily --help               Scoped help for the daily report
    thermal help daily                 Same page via the help command
    thermal stats                      Daily distribution and token composition
    thermal audit                      Audit instruction files & MCP context tax
    thermal changelog --top 5          View latest new features & arguments

  Filter & Window Options:
    --tool <name>          Filter to a single tool (default: all)
    --interval <dur>       Polling interval for live monitor (default: 1s)
    --flame-ceiling <N>    Live flame ceiling in active units per minute (e.g. 9M or 60)
    --since <date>         Start date filter (YYYY-MM-DD or YYYYMMDD)
    --until <date>         End date filter (YYYY-MM-DD or YYYYMMDD)
    --last <N>             Limit to last N days, weeks, or months
    --order <asc|desc>     Sort order for report rows (default: desc)
    --weeks <N>            Heatmap width in weeks (default: 52)
    --start-of-week <day>  First day of week: sunday-saturday (default: sunday)

  Ranking & Analytics Options:
    --sort <field>         Leaderboard: streak|tokens|cost; projects: tokens|cost|days|recent;
                           models: tokens|cost
    --top <N>              Cap table rows to top N items (default: all)
    --metric <field>       Analytics metric: tokens or cost (default: tokens)
    --by <tool|model>      Mix breakdown dimension (default: tool)
    --grain <grain>        Mix window: day, week, or month (default: week)

  Display & Formatting Options:
    --chart                Charts for daily, weekly, monthly, projects, and models reports
    --breakdown            Show per-model token breakdown under each period
    --nerd                 Use Nerd Font flame glyph (󰈸) in statusline
    --plain                Plain text statusline output without emoji or glyphs
    --no-estimate          Recorded cost only; disable models.dev pricing estimates
    --format <text|json>   Output format (json is an alias for --json; md/markdown retired)

  Global & Engine Options:
    --db <path>            Read one tool's data from a file or directory (requires --tool)
    --offline              Use cached models.dev pricing catalog without network
    --fresh                Bypass cached statusline data; live sessions always start at launch
    --stream, -f           Stream continuous NDJSON live events (live verb)
    --json                 Output structured JSON (never launches a terminal UI)
    --no-color             Disable all ANSI escape sequences
    -V, --version          Print version, commit, and build timestamp
    --verbose, -v          Output diagnostic warnings to stderr
    -h, --help             Show global or command-scoped help

  Supported Tools:
    Token Warriors:    OpenCode · Devin · Codex · Claude · ZCode · Grok · MiMoCode (mimo) ·
                       codewhale · DeepSeek DSH (dsh) · Nous Hermes (hermes) · CMD
                       (command-code) · Zed
    Activity Hunters:  Agy (agy) · Muse (muse) · Droid (droid)

  Documentation & guides: https://jadmadi.net/project/thermal/ · AGPL-3.0 / Commercial
```

### 80 columns: stats

```text

  Thermal · stats · tokens

  Active days  1
  Total        759.7K
  Mean         759.7K
  Median       759.7K
  p90          759.7K
  Max          759.7K

  Weekday profile
  Thursday        759.7K  ████████████████     1 day

  Token composition
  Cache read            8.0K  ████████████████   82.5%
  Cache write            200  █···············    2.1%
  Uncached input        1.0K  ██··············   10.3%
  Output                 500  █···············    5.2%
  Reasoning                0  ················    0.0%
  Unclassified             0  ················    0.0%
  Cache hit rate: 87.0% of prompt tokens read from cache

  Distribution
          0 — 76.0K  ························     0
     76.0K — 151.9K  ························     0
    151.9K — 227.9K  ························     0
    227.9K — 303.9K  ························     0
    303.9K — 379.9K  ························     0
    379.9K — 455.8K  ························     0
    455.8K — 531.8K  ························     0
    531.8K — 607.8K  ························     0
    607.8K — 683.7K  ························     0
    683.7K — 759.7K  ████████████████████████     1

  Top days
  2026-01-15       759.7K
```

### 120 columns: stats

```text

  Thermal · stats · tokens

  Active days  1
  Total        759.7K
  Mean         759.7K
  Median       759.7K
  p90          759.7K
  Max          759.7K

  Weekday profile
  Thursday        759.7K  ████████████████     1 day

  Token composition
  Cache read            8.0K  ████████████████   82.5%
  Cache write            200  █···············    2.1%
  Uncached input        1.0K  ██··············   10.3%
  Output                 500  █···············    5.2%
  Reasoning                0  ················    0.0%
  Unclassified             0  ················    0.0%
  Cache hit rate: 87.0% of prompt tokens read from cache

  Distribution
          0 — 76.0K  ························     0
     76.0K — 151.9K  ························     0
    151.9K — 227.9K  ························     0
    227.9K — 303.9K  ························     0
    303.9K — 379.9K  ························     0
    379.9K — 455.8K  ························     0
    455.8K — 531.8K  ························     0
    531.8K — 607.8K  ························     0
    607.8K — 683.7K  ························     0
    683.7K — 759.7K  ████████████████████████     1

  Top days
  2026-01-15       759.7K
```

### 80 columns: share

```text

  Thermal · share · stateless streak card

  ╭─ Stateless Streak Card ── Client-Verified · Zero-Database ─╮
  │  Scope:         All                                        │
  │  Window:        all time (1 active day)                    │
  │  Coverage:      recorded+estimated                         │
  │  Streak:        1 days 🔥 (Longest: 1 days, token-active)  │
  │  Active Days:   1 token-active days                        │
  │  Total Volume:  759.7K tokens                              │
  │  Spend Profile: $1.75 incl. est.                           │
  ╰────────────────────────────────────────────────────────────╯


  Shareable Card URL:
  https://jadmadi.net/project/thermal/share#v1.71687…tOJ7BzOc9-IVTIpnSiIjj-DsA
  … full URL (358 chars) is available with --json

  • Privacy Guarantee: 100% client-only URL fragment. Zero database or cloud
    storage. No file paths, repository roots, or prompt telemetry encoded.
```

### 120 columns: share

```text

  Thermal · share · stateless streak card

  ╭─ Stateless Streak Card ── Client-Verified · Zero-Database ─╮
  │  Scope:         All                                        │
  │  Window:        all time (1 active day)                    │
  │  Coverage:      recorded+estimated                         │
  │  Streak:        1 days 🔥 (Longest: 1 days, token-active)  │
  │  Active Days:   1 token-active days                        │
  │  Total Volume:  759.7K tokens                              │
  │  Spend Profile: $1.75 incl. est.                           │
  ╰────────────────────────────────────────────────────────────╯


  Shareable Card URL:
  https://jadmadi.net/project/thermal/share#v1.71687d92.TI9RbsIwEE…ZuVbr9tOJ7BzOc9-IVTIpnSiIjj-DsA
  … full URL (358 chars) is available with --json

  • Privacy Guarantee: 100% client-only URL fragment. Zero database or cloud storage. No file
    paths, repository roots, or prompt telemetry encoded.
```

## Row-level JSON schemas

Canonical lower-camelCase keys; no uppercase aliases. All floats are serialized
at four decimal places by `cmd/thermal/json.go`.

### Leaderboard result

```text
["activeDays","currentStreak","daily","dataPath","estimatedCost","longestStreak","name","summary","tool","totalActivity","unattributedTokens"]
["cacheTokens","cost","filesTouched","inputTokens","lifetimeTokens","linesAdded","linesDeleted","longestSessionMs","outputTokens","reasoningTokens","sessions","tool"]
```

### Daily row

```text
["activeDays","cacheTokens","cost","estimatedCost","inputTokens","models","outputTokens","period","reasoningTokens","storedCost","totalTokens","turns"]
```

### Project row

```text
["activeDays","cacheTokens","cost","firstDay","inputTokens","lastDay","models","outputTokens","project","reasoningTokens","storedCost","toolTokens","tools","totalTokens","turns"]
```

### Model row

```text
["activeDays","cacheReadTokens","cacheWriteTokens","estimatedCost","firstDay","inputTokens","lastDay","model","outputTokens","reasoningTokens","tools","totalTokens","unattributedTokens"]
```

### Live event / rollup

```text
[]
[]
```

### Share snapshot

```text
["activeDays","coverage","currentStreak","days","estimatedCost","longestStreak","models","recordedCost","scope","streakBasis","t","tool","totalCost","totalTokens","v","window"]
```

### Envelope types and timestamps

```text
leaderboard      leaderboard    2026-10-29T20:48:07Z
daily            daily          2026-10-29T20:48:07Z
weekly           weekly         2026-10-29T20:48:07Z
monthly          monthly        2026-10-29T20:48:07Z
projects         projects       2026-10-29T20:48:07Z
models           models         2026-10-29T20:48:07Z
mix              mix            2026-10-29T20:48:07Z
stats-token      stats          2026-10-29T20:48:07Z
stats-cost       stats          2026-10-29T20:48:07Z
trend-token      trend          2026-10-29T20:48:07Z
trend-cost       trend          2026-10-29T20:48:07Z
share            share          2026-10-29T20:48:07Z
audit            audit          2026-10-29T23:48:07.09271347+03:00
changelog        changelog      2026-10-29T20:48:07Z
live             live           2026-10-29T23:48:07.100966823+03:00
version          version        2026-10-29T20:48:07Z
```

## Validation cases

Actual command and exit status; stderr/stdout are combined for readability.

### Report flag without report

```text
$ thermal --since 2026-01-01
exit 2
thermal: report options (--since, --until, --last, --breakdown, --chart, --order, --start-of-week) need a report command: daily, weekly, monthly, projects, or models
  run thermal --help for the flag list
```

### Analytics --chart is rejected

```text
$ thermal mix --chart
exit 2
thermal: --chart does not apply to the mix command; use daily, weekly, monthly, projects, or models
  run thermal --help for the flag list
```

### --top on daily is rejected

```text
$ thermal daily --top 2
exit 2
thermal: --top only applies to the projects and models commands
  run thermal --help for the flag list
```

### --chart on daily succeeds

```text
$ thermal daily --chart
exit 0

  Thermal · daily usage · all tools

  Window: [all time (1 active day) · all tools]  Total: [759.7K tok · ~$1.75]  Cache: [89.1% hit]

  ╭─ Daily Usage ─────────────────────────────────────────────────────────────── all tools ─╮
  │  Period      Models                   Input     Output      Cache     Tokens       Cost │
  │  ────────────────────────────────────────────────────────────────────────────────────── │
  │  2026-01-15  claude-3.5-son… +1        1.0K        500       8.2K     759.7K     ~$1.75 │
  │  ────────────────────────────────────────────────────────────────────────────────────── │
  │  Total                                 1.0K        500       8.2K     759.7K     ~$1.75 │
  ╰─────────────────────────────────────────────────────────────────────────────────────────╯

  Total = $1.75 recorded + ~$0.0046 estimated from pricing data.


  Tokens by period  bar = share of the largest charted period, 759.7K tok

  2026-01-15    ████████████████████████████████████████████████  759.7K tok
```

## Appendices

### Full share JSON

```json
{
  "generatedAt": "2026-10-29T20:48:07Z",
  "snapshot": {
    "activeDays": 1,
    "coverage": "recorded+estimated",
    "currentStreak": 1,
    "days": {
      "2026-01-15": 759700
    },
    "estimatedCost": 0.0046,
    "longestStreak": 1,
    "models": {
      "claude-3.5-haiku": 9700,
      "claude-3.5-sonnet": 750000
    },
    "recordedCost": 1.75,
    "scope": "all",
    "streakBasis": "token-active",
    "t": 1793306887,
    "tool": "all",
    "totalCost": 1.7546,
    "totalTokens": 759700,
    "v": 1,
    "window": "all time (1 active day)"
  },
  "token": "v1.71687d92.TI9RbsIwEESvEu1XqzpoAzUBf7a9QXsBy15RC-Ot7CUIody9itNQPmf8ZmZ9gwFMp0DAdP1-s8HtbtcrEOYIBmyMoKA4_qG7uoTk-TLLRsKJmqeusU7CQI231-cpIJns8c2WUMCA8JFSOxOgwJ1zpiSflanbkdOByqMzwx_2WqoUFhu_ppoCptf7HlFBJsfZk3_nMh2_6rUCKhJOVhYTV4ivWv8V_IPVczxQtofpY0vVyz0PCnxdv8Ea19sWu7bTy_ao4MSeYn120Z49tZuVbr9tOJ7BzOc9-IVTIpnSiIjj-DsA",
  "type": "share",
  "url": "https://jadmadi.net/project/thermal/share#v1.71687d92.TI9RbsIwEESvEu1XqzpoAzUBf7a9QXsBy15RC-Ot7CUIody9itNQPmf8ZmZ9gwFMp0DAdP1-s8HtbtcrEOYIBmyMoKA4_qG7uoTk-TLLRsKJmqeusU7CQI231-cpIJns8c2WUMCA8JFSOxOgwJ1zpiSflanbkdOByqMzwx_2WqoUFhu_ppoCptf7HlFBJsfZk3_nMh2_6rUCKhJOVhYTV4ivWv8V_IPVczxQtofpY0vVyz0PCnxdv8Ea19sWu7bTy_ao4MSeYn120Z49tZuVbr9tOJ7BzOc9-IVTIpnSiIjj-DsA"
}
```

### Fixture

- Home: fixed temp fixture (removed after generation)
- Date: 2026-01-15
- Hash: `2db489728bc70f4013f5aec8b4fd805ab961414dbfe7f6c52a0619a8a4053e26`
- Reports: codewhale 750,000 tokens/$1.75 recorded; Claude 9,700 tokens;
  Agy 12 steps; two projects; two models.

### Normalization

The gallery gate compares normalized totals and schema keys, not capture
timestamps or randomized fixture paths. It fails on retired commands, mixed
totals, aliases, >4-decimal floats, missing basis/unclassified fields, unit
truncation, overlapping heatmap month labels, chart scale/order drift, and empty validation cases.
