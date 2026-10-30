# Thermal Migration Guide

This guide documents breaking changes, removed aliases, flag validation updates, and schema migrations across versions of **Thermal** (`github.com/jadmadi/thermal`).

Thermal follows Semantic Versioning (`vMAJOR.MINOR.PATCH`). In accordance with our CLI stability contract, every breaking change conforms to a strict 6-part entry format with exact detection signals, search patterns, before/after examples, fixes, and escape hatches.

---

## Table of Contents
- [Staged Deprecations (v0.14.0)](#staged-deprecations-v0140)
  - [Deprecated: `--license` CLI Flag](#deprecated---license-cli-flag)
- [Upgrading to v0.22.0](#upgrading-to-v0220)
  - [Changed: Strict Argument Parsing and Usage Exit Codes](#changed-strict-argument-parsing)
  - [Changed: Help Is Intercepted and Command-Scoped](#changed-help-precedence)
  - [Changed: `-v` Means Verbose; Use `-V` or `--version`](#changed-version-verbose-flags)
  - [Changed: Upgrade Preview, Confirmation, and Exit Codes](#changed-upgrade-safety)
  - [Changed: `--db` Requires One Tool and Validates the Path](#changed-db-scope)
  - [Changed: Unknown Commands and Tools Exit 2 with Suggestions](#changed-unknown-token-diagnostics)
  - [Changed: `--top` Caps Leaderboard and Period Rows](#changed-top-contract)
  - [Added: `thermal sources` and an Actionable Empty State](#added-sources-inventory)
  - [Changed: Non-TTY `live` Exit Code and Silent Statusline Degradation](#changed-nontty-statusline-contract)
  - [Changed: Reports Drop Columns Below 80 and Live Refuses Below 60](#changed-width-contract)
  - [Changed: Leaderboard States Its Ranking, Legend, and Cost Arithmetic](#changed-leaderboard-ranking-legend)
  - [Changed: Project and Model Totals Use Distinct-Day Denominators](#changed-total-row-semantics)
  - [Changed: Analytics Denominators and Audit Prose](#changed-analytics-prose)
  - [Changed: No-Data Is a Successful State and JSON Keeps Its Envelope](#changed-no-data-success)
  - [Changed: Data-Access Errors Name the Condition and Next Commands](#changed-data-access-errors)
  - [Changed: Explicit `--last` Units, Repeated-Flag Warnings, and `--` Diagnostics](#changed-window-units)
  - [Changed: Live SIGINT Is a Clean Stop; SIGTERM Exits 143](#changed-live-signals)
  - [Changed: Version Metadata and Deprecation Warnings Are Truthful](#changed-version-deprecation-truth)
  - [Changed: Exit Codes Are Documented and `upgrade` Accepts Global Flags](#changed-exit-codes-upgrade-flags)
- [Upgrading to v0.20.1](#upgrading-to-v0201)
  - [Changed: Renamed `command-code` Canonical Display Name to `CMD`](#changed-rename-command-code-to-cmd)
- [Upgrading to v0.20.2](#upgrading-to-v0202)
  - [Changed: Token-Boundary Receipt Evidence Classification](#changed-receipt-evidence-token-boundary)
  - [Changed: Workload Replay Volume Conservation and Price Coverage](#changed-replay-volume-and-price-coverage)
  - [Changed: Rolling-Window Capacity Verdicts from Daily Telemetry](#changed-replay-rolling-capacity-unknown)
  - [Changed: `thermal replay --against` Resolves Built-In Plan Identifiers](#changed-replay-against-plan-resolution)
  - [Changed: Live Monitor Project Baseline Preservation and Delta Clamping](#changed-live-project-baseline-clamping)
- [Upgrading to v0.21.0](#upgrading-to-v0210)
  - [Changed: Command Code Token and Project Attribution](#changed-command-code-token-attribution)
  - [Added: Commercial Use License Pricing in `thermal license`](#added-commercial-license-pricing)
- [Upgrading to v0.16.x](#upgrading-to-v016x)
  - [Added: GitHub Flavored Markdown Output for Receipt and Yield Reports](#added-markdown-pr-receipts-and-yield)
  - [Added: Instant Shell Statusline and Prompt Helper (`thermal statusline`)](#added-terminal-statusline-helper)
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

## Upgrading to v0.22.0

### Changed: No-Data Is a Successful State and JSON Keeps Its Envelope
<a id="changed-no-data-success"></a>

**Affected if**:
- **Search pattern**: `grep -rnE 'thermal.*(--json|\| *jq|\| *grep)' .`
- **Detection signal**: On a machine with no agent sessions, `thermal --json` now returns `"type":"leaderboard"` with `"state":"no-data"` and exits `0`; it no longer returns `"type":"sources"` with exit `1`.

**What changed and why**:
Thermal treated "no sessions yet" as a command failure and replaced the requested JSON document with a sources inventory. That made `jq .ranking` return `null` and forced every consumer to special-case a foreign schema. Read-only reports now exit `0`, return the envelope they were asked for, and add `state:"no-data"` plus a `next` object; `thermal sources --json` remains the only `type:"sources"` producer.

**Before / After**:
```bash
# Before: foreign envelope, exit 1
thermal --json | jq -r .type    # sources

# After: requested envelope, exit 0
thermal --json | jq -r .type    # leaderboard
thermal --json | jq -r .state   # no-data
```

**The fix**:
Check the state or the empty collection, and keep using the requested keys:
```bash
thermal --json | jq -e '.state=="no-data" or (.results|length)>0'
thermal sources --json | jq -e '.summary.found > 0'
```
Add `--require-data` when empty data must remain a non-zero exit; stdout keeps the same payload.

**Escape hatch**:
`--require-data` restores exit `1` for scripts that treat empty data as a failure. `thermal sources --json` remains the exit-0 inventory for setup checks.

---

### Changed: Data-Access Errors Name the Condition and Next Commands
<a id="changed-data-access-errors"></a>

**Affected if**:
- **Search pattern**: `grep -rnE 'database not found|data directory not found|cannot read .*check permissions' .`
- **Detection signal**: `thermal opencode weekly` on a machine without OpenCode now prints `thermal: OpenCode has no data at ~/.local/share/opencode/opencode.db` followed by inventory, all-tools, and `--db` lines, instead of one bare `database not found` line.

**What changed and why**:
A missing default path is now a typed data-access error that names the condition, the tool, and the display path. Symlink loops and permission failures no longer share one "check permissions" sentence: a loop points at `readlink -f`, a permission failure names the mode and `chmod`, and a wrong shape keeps its existing precise fix. Explicit `--db` errors keep their wording and never fall back to a default; `--verbose` adds one raw OS line, never a stack trace.

**Before / After**:
```bash
# Before: one dead-end line
thermal: database not found: /home/you/.local/share/opencode/opencode.db

# After: condition, inventory, all-tools report, and override
thermal: OpenCode has no data at ~/.local/share/opencode/opencode.db
  inventory: thermal sources
  all tools: thermal weekly
  override:  thermal --tool opencode --db <path>
```

**The fix**:
Run `thermal sources` to see which source exists, use the all-tools report, or pass the exact path with `--tool <name> --db <path>`. Scripts that parsed the old `database not found` text should match `has no data at` or rely on exit `1`.

**Escape hatch**:
Explicit `--db` validation keeps its existing messages (`--db path does not exist`, `expects a directory`, `expects <file>`), so scripts that supplied a path are unaffected.

---

### Changed: Explicit `--last` Units, Repeated-Flag Warnings, and `--` Diagnostics
<a id="changed-window-units"></a>

**Affected if**:
- **Search pattern**: `grep -rnE 'thermal .*--last [0-9]+' .` or `grep -rnE -- '--tool .*--tool ' .`
- **Detection signal**: `thermal weekly --last 14d` now prints `Window: [last 14 days ...]`; `thermal --tool claude --tool codewhale` now writes a stderr warning naming both values; `thermal weekly -- --offline` explains that `--offline` became a positional.

**What changed and why**:
A bare `--last N` still uses the report's unit, but a `d`, `w`, or `m` suffix now pins the window unit and labels it explicitly. Repeated scalar flags keep the last value and warn on stderr with the discarded value, instead of silently changing scope. The `--` terminator now explains that following tokens are positional, and the `--db` examples use a real path instead of a shell-breaking `<path>` placeholder.

**Before / After**:
```bash
# Before: 7 meant weeks on weekly, with no way to say 7 days
thermal weekly --last 7

# After: both units are expressible
thermal weekly --last 7        # 7 weeks
thermal weekly --last 14d      # 14 days
thermal monthly --last 3m      # 3 months

# Before: claude silently discarded
thermal --tool claude --tool codewhale

# After: same result, but the discard is named
thermal: warning: --tool given 2 times; using "codewhale" (last), ignoring "claude"
```

**The fix**:
Add the unit you mean to `--last`, pass one value per scalar flag, and put flags before `--` (or drop `--`). Boolean repeats such as `--json --json` remain silent and idempotent.

**Escape hatch**:
Bare `--last N` behavior is unchanged. `--since`/`--until` remain the exact-date form, and `--last` with a suffix cannot be combined with them.

---

### Changed: Live SIGINT Is a Clean Stop; SIGTERM Exits 143
<a id="changed-live-signals"></a>

**Affected if**:
- **Search pattern**: `grep -rnE 'thermal live' .` in wrappers that treat any non-zero exit as a crash.
- **Detection signal**: After a single Ctrl+C, `thermal live` now exits `0` and prints no `live monitor failed` line; SIGTERM exits `143`.

**What changed and why**:
Bubble Tea reports a user interrupt as an error, and Thermal treated every run error as a renderer failure. The monitor now classifies interrupts and context cancellation as a normal stop, restores the alternate screen, cursor, and bracketed paste, and exits `0`. Real renderer errors still print one failure line and exit `1`; SIGTERM is tracked separately at `143`, and a second interrupt during a hanging shutdown exits `130`.

**Before / After**:
```bash
# Before: Ctrl+C looked like a crash
thermal: live monitor failed: program was killed: program was interrupted
exit 1

# After: Ctrl+C is a normal stop
exit 0
```

**The fix**:
Treat exit `0` after Ctrl+C as success. For scripts, prefer `thermal live --json` or `thermal live --json --stream`; neither depends on an interactive stop.

**Escape hatch**:
A second Ctrl+C during shutdown forces exit `130`. SIGTERM remains `143` for supervisors that use it as a stop signal.

---

### Changed: Version Metadata and Deprecation Warnings Are Truthful
<a id="changed-version-deprecation-truth"></a>

**Affected if**:
- **Search pattern**: `grep -rnE 'thermal (version|--version|-V)' .` or scripts parsing the version JSON.
- **Detection signal**: An unstamped build now prints `version not stamped`, `commit: unknown`, and `built: unknown`; `version --json` adds `"stamped": false` and `"semver": ""`.

**What changed and why**:
The help page promised SemVer, commit, and build time unconditionally, while a source build omitted all three. Version output now states whether the build is stamped and prints the unknown values explicitly. Deprecation warnings no longer name the expired `v0.14.0` deadline or a repository-relative `docs/MIGRATION.md` path; they name the replacement and an absolute migration URL.

**Before / After**:
```bash
# Before: a source build implied nothing and a warning expired in v0.14.0
thermal version
thermal dev (built from source)
thermal: warning: --license is deprecated and will be removed in v0.14.0; ...

# After
thermal version
thermal dev (built from source; version not stamped)
  commit: unknown
  built:  unknown
thermal: warning: --license is deprecated; use thermal license
  details: https://jadmadi.net/project/thermal/migration#deprecated-license-flag
```

**The fix**:
Read `stamped` and `semver` in JSON; do not infer a release from the presence of the `version` key. Parse `thermal version` output only when `stamped` is true. Warning consumers should match `deprecated; use` rather than a version number.

**Escape hatch**:
The `--license` flag and `nous` alias still work during the deprecation window. Version JSON keeps its existing keys and adds only `stamped` and `semver`.

---

### Changed: Exit Codes Are Documented and `upgrade` Accepts Global Flags
<a id="changed-exit-codes-upgrade-flags"></a>

**Affected if**:
- **Search pattern**: `grep -rnE 'thermal upgrade.*--(offline|no-update-check|width|fresh)' .` or scripts that infer exit codes.
- **Detection signal**: `thermal upgrade --check --offline` now exits `1` with `offline: release metadata is not cached`; `thermal upgrade --width 100` exits `2` with `not applicable`, not `unknown flag`.

**What changed and why**:
`upgrade` used a private parser that called every unrecognized global flag "unknown". The command now accepts `--offline` and `--no-update-check`, fails offline closed because release metadata is not cached, and rejects `--width`/`--fresh` with an explicit alternative. The `0/1/2/3` exit contract is printed in global help, `upgrade --help`, and the README.

**Before / After**:
```bash
# Before
thermal upgrade --check --offline
thermal: unknown flag for upgrade: --offline
exit 2

# After
thermal upgrade --check --offline
thermal: offline: release metadata is not cached; run thermal upgrade --check without --offline
exit 1
```

**The fix**:
Drop `--offline` from `upgrade` if you want the network check; use `--check` or `--dry-run` for non-mutating previews. Treat exit `1` as a data/network failure, `2` as usage, and `3` as a refused non-interactive mutation.

**Escape hatch**:
`--no-update-check` is accepted as a no-op on `upgrade`; `--width` and `--fresh` are not applicable and return a runnable alternative.

---

### Changed: Strict Argument Parsing and Usage Exit Codes
<a id="changed-strict-argument-parsing"></a>

**Affected if**:
- **Search pattern**: `grep -rnE 'thermal[^|;]*--(last|weeks|top|interval)[ =]([^0-9-]|$)' .`
- **Detection signal**: A command that previously returned a full report with exit `0` now prints `thermal: --last wants a whole number of days, got "abc"` and exits `2`.

**What changed and why**:
Thermal used Go's `flag` parser only until the first positional and a hand-rolled scanner after it; the scanner silently ignored invalid numeric values and unknown flags. A typo such as `daily --last abc` rendered the all-time report with exit `0`, and `opencode --tool codex` silently chose the positional. The parser now accepts flags in any position, validates every value exactly once, rejects contradictory arguments, and exits `2` for usage errors. `--last` must be at least `1`, `--since` may not be after `--until`, `--tool` may not conflict with a positional tool, and unknown flags are rejected with a nearest-match suggestion.

**Before / After**:
```bash
# Before: silently all-time, exit 0
thermal daily --last abc

# After: exit 2, no report
thermal: --last wants a whole number of days, got "abc"
  example: thermal daily --last 7
```

**The fix**:
Use a valid value or remove the flag:
```bash
thermal daily --last 7
thermal daily --last 7 --json
```

**Escape hatch**:
There is no silent fallback. To request all time, omit `--last`; to scope a report, pass `--since`/`--until`. Scripts should treat exit `2` as a usage error and exit `1` as a data or semantic error.

---

### Changed: Help Is Intercepted and Command-Scoped
<a id="changed-help-precedence"></a>

**Affected if**:
- **Search pattern**: `grep -rnE 'thermal [a-z-]+ +--help' .`
- **Detection signal**: `thermal daily --help` now prints `Thermal · daily · Daily token volume and cost breakdown` and exits `0`; it no longer prints the daily table.

**What changed and why**:
Help used to be honoured only when it was the sole argument. After a command word, `--help` was ignored and the command executed; `thermal live --help` opened the TUI and `thermal upgrade --help` replaced the binary. Help is now intercepted before dispatch, wins over validation and execution, and prints a page scoped to the command or tool. `thermal help <command>` is equivalent to `thermal <command> --help`.

**Before / After**:
```bash
# Before: daily --help rendered data; upgrade --help upgraded
thermal daily --help
thermal upgrade --help

# After: exit 0, read-only scoped pages
thermal daily --help
thermal help daily
```

**The fix**:
Remove `--help` from data pipelines and use `--json` for machine output. Use `thermal <command> --help` or `thermal help <command>` to inspect accepted flags.

**Escape hatch**:
None is needed; help is always read-only. `--` does not disable help, and a help request wins over an invalid flag.

---

### Changed: `-v` Means Verbose; Use `-V` or `--version`
<a id="changed-version-verbose-flags"></a>

**Affected if**:
- **Search pattern**: `grep -rnE 'thermal +-v([^a-zA-Z]|$)' .`
- **Detection signal**: `thermal -v` now renders the leaderboard and exits `0` (or exits `1` with the empty-state message when no agent data exists); it no longer prints `thermal dev (built from source)`.

**What changed and why**:
`-v` used to print the version only when it was the sole argument and meant verbose everywhere else, while the help documented it as verbose. The overload is removed: `-v`/`--verbose` enable diagnostics, and `-V`/`--version`/`thermal version` print build metadata. `-V` wins in any position and over validation, so `thermal -V daily` prints the version and ignores the command.

**Before / After**:
```bash
# Before: context-dependent
thermal -v          # version
thermal -v daily    # verbose daily report

# After: one meaning
thermal -v          # verbose leaderboard
thermal -V          # version
thermal --version   # version
```

**The fix**:
Replace `thermal -v` with `thermal -V` or `thermal --version` in scripts and aliases:
```bash
thermal -V
thermal -V --json
```

**Escape hatch**:
`thermal version` remains the long-form command and always prints metadata.

---

### Changed: Upgrade Preview, Confirmation, and Exit Codes
<a id="changed-upgrade-safety"></a>

**Affected if**:
- **Search pattern**: `grep -rnE 'thermal +upgrade([[:space:]]|$)' .`
- **Detection signal**: `thermal upgrade </dev/null` now exits `3` with `thermal: refusing to replace the binary without a terminal; re-run with --yes or use --check`; it no longer upgrades.

**What changed and why**:
`thermal upgrade` used to download and replace the binary immediately and ignored every flag, so `--help`, `--check`, `--dry-run`, and `--json` all authorized a real upgrade. The command now parses its own flags, rejects unknown flags with exit `2`, and separates preview from mutation. `--check` and `--dry-run` never download or mutate; a mutation requires a TTY confirmation or `--yes`; non-interactive mutation without `--yes` exits `3`. `--json` emits one object on stdout, and a failed download, extraction, or rename leaves the existing binary intact.

**Before / After**:
```bash
# Before: every flag authorized a real replacement
thermal upgrade --dry-run

# After: preview only, no download
thermal upgrade --dry-run
thermal upgrade --check --json

# After: explicit non-interactive mutation
thermal upgrade --yes
```

**The fix**:
Add `--yes` to automated upgrades, or use `--check`/`--dry-run` when the goal is inspection:
```bash
thermal upgrade --check
thermal upgrade --yes
```

**Escape hatch**:
`--check` and `--dry-run` are always read-only. Failed upgrades leave the original binary in place, and the previous binary is restored if the final rename fails.

---

### Changed: `--db` Requires One Tool and Validates the Path
<a id="changed-db-scope"></a>

**Affected if**:
- **Search pattern**: `grep -rnE 'thermal[^|;]*--db' .`
- **Detection signal**: `thermal --db /nonexistent daily` now exits `2` with `thermal: --db applies to one tool; add --tool <name>`, and `thermal --tool opencode --db /tmp daily` exits `1` with `thermal: /tmp is a directory; OpenCode expects opencode.db inside it`.

**What changed and why**:
All-tool views used to ignore `--db` and render the default machine-wide data, so a user who thought they had scoped the read could receive unrelated data with exit `0`. `--db` now requires a single tool, validates the path before opening it, and never falls back to a default. SQLite-file tools accept a database file or a directory containing the expected filename; directory tools accept a directory. Raw SQLite driver text is replaced by an actionable message and remains available behind `--verbose`.

**Before / After**:
```bash
# Before: silently machine-wide, exit 0
thermal --db /nonexistent daily

# After: exit 2, no report
thermal: --db applies to one tool; add --tool <name>
  example: thermal --tool opencode --db <path> daily

# After: actionable path error, exit 1
thermal --tool opencode --db /tmp daily
thermal: /tmp is a directory; OpenCode expects opencode.db inside it
```

**The fix**:
Add `--tool <name>` and pass a valid file or directory for that tool:
```bash
thermal --tool opencode --db ~/.local/share/opencode/opencode.db daily
thermal --tool codex --db ~/.codex daily
```

**Escape hatch**:
Omit `--db` to use the default paths. `--verbose` prints the raw driver error beneath the user-facing message.

---

### Changed: Unknown Commands and Tools Exit 2 with Suggestions
<a id="changed-unknown-token-diagnostics"></a>

**Affected if**:
- **Search pattern**: `grep -rnE 'unknown tool:|unknown command:' .`
- **Detection signal**: `thermal today` now exits `2` with `thermal: unknown command "today"` and `did you mean: thermal daily?`; it no longer prints `unknown tool: today` with exit `1`.

**What changed and why**:
Every unrecognized positional used to be treated as a tool, so a command typo (`today`), a report word in the wrong slot (`stats weekly`), and a path (`projects /tmp`) all produced `thermal: unknown tool: …` with no suggestion. The CLI now classifies the token, suggests the nearest command or tool, names the wrong slot explicitly, and recognizes path-like tokens. Unknown commands/tools, wrong-slot tokens, paths, and empty tokens exit `2`; retired verbs keep exit `1`.

**Before / After**:
```bash
# Before
thermal today
thermal: unknown tool: today

# After
thermal today
thermal: unknown command "today"
  did you mean: thermal daily?
  Valid commands: daily, weekly, monthly, projects, models, stats, mix, trend,
  live, statusline, share, audit, changelog, upgrade, version, license
  run thermal --help for the command list
```

**The fix**:
Use the suggested command/tool, or run `thermal --help` for the vocabulary. Scripts should treat any non-zero exit as failure and not parse stderr.

**Escape hatch**:
There is no compatibility mode. `thermal --help` and scoped help remain the authoritative vocabularies.

---

### Changed: `--top` Caps Leaderboard and Period Rows
<a id="changed-top-contract"></a>

**Affected if**:
- **Search pattern**: `grep -rnE 'thermal[^|;]*--top' .`
- **Detection signal**: `thermal daily --top 5` now exits `0` with five period rows, and `thermal stats --top 5` exits `2` with `thermal: --top does not apply to the stats command` followed by the valid command set.

**What changed and why**:
`--top` was accepted only on projects, models, and changelog, while the help described it as a general row cap and chart footers told readers to raise it on commands that rejected it. The flag now has one applicability table: leaderboard, daily, weekly, monthly, projects, models, and changelog. `--top 0` means all rows; an absent changelog `--top` keeps the three-release default. Tables, charts, and JSON data use the same cap, totals still describe the full window, and chart footers state `chart shows X of Y; use --top Y for all`.

**Before / After**:
```bash
# Before: rejected on daily, accepted on changelog, recommended by charts
thermal daily --top 5
thermal: --top only applies to the projects and models commands

# After: five rows, full-window totals
thermal daily --top 5

# After: rejecting commands name the valid set
thermal stats --top 5
thermal: --top does not apply to the stats command
  it applies to: leaderboard, daily, weekly, monthly, projects, models, changelog
```

**The fix**:
Use `--top N` on the supported commands; use `--top 0` for all rows. Remove `--top` from stats/mix/trend/audit/share/live/statusline invocations.

**Escape hatch**:
`--json` remains available for the full payload on supported commands; chart footers always suggest a valid `--top` value.

---

### Added: `thermal sources` and an Actionable Empty State
<a id="added-sources-inventory"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'no supported tool data found' .`
- **Detection signal**: `HOME=$(mktemp -d) thermal` now prints `Thermal · sources · 15 supported tools` with a `found/empty/missing/unreadable/wrong-shape` table, and `HOME=$(mktemp -d) thermal --json` emits a `type:"sources"` object on stdout with exit `1`.

**What changed and why**:
The empty state used to be one line (`thermal: no supported tool data found`) with no paths, states, or next steps. Thermal now prints a stat-only inventory of all 15 supported sources: expected file or directory, state, and a next action. `thermal sources` exposes the same inventory at any time and exits `0`; `thermal sources --json` is the machine-readable form with 15 stable entries. The inventory reads path metadata only — no database, transcript, prompt, or session content is opened.

**Before / After**:
```bash
# Before
thermal
thermal: no supported tool data found

# After
thermal
Thermal · sources · 15 supported tools
TOOL ... PATH ... STATE ... DETAIL
...
Next: run one prompt in any supported tool, then re-run thermal.

# Script form (exit 0)
thermal sources --json | jq '.summary'
```

**The fix**:
Use `thermal sources --json` when a script needs an exit-0 inventory. Treat `thermal --json` on an empty machine as a structured no-data signal (`type:"sources"`, exit `1`).

**Escape hatch**:
`thermal sources` always exits `0` once it can produce the inventory; the human empty-state table still exits `1` so a data check remains meaningful.

---

### Changed: Non-TTY `live` Exit Code and Silent Statusline Degradation
<a id="changed-nontty-statusline-contract"></a>

**Affected if**:
- **Search pattern**: `grep -rnE 'thermal +live([^a-zA-Z]|$)' .` or `grep -rn 'thermal statusline' .`
- **Detection signal**: `thermal live </dev/null` now exits `2`, writes `thermal: live needs a terminal` to stderr, and leaves stdout empty. `HOME=$(mktemp -d) thermal statusline` exits `0` with no stdout or stderr.

**What changed and why**:
`thermal live` in a pipe wrote human guidance to stdout and exited `0`, so a pipeline treated a non-result as success. `statusline` in an empty environment exited `1` and wrote an error to stderr, so a shell prompt printed an error on every redraw. The channel contract is now explicit: `live` guidance is stderr-only and non-TTY text exits `2`; `--json` and `--json --stream` remain machine-clean. `statusline` degrades silently with exit `0` when no source has data, and `--json` emits `state:"no-data"` (or `state:"ok"` with data). `share` remains a deliberate stderr failure with exit `1`.

**Before / After**:
```bash
# Before: pipeline success, prose on stdout
thermal live | jq .
# stdout: "thermal live is interactive and needs a terminal."; exit 0

# After: usage error, clean stdout
thermal live
# stderr: "thermal: live needs a terminal"; exit 2
thermal live --json | jq .type
# "live"

# Before: prompt error
HOME=$(mktemp -d) thermal statusline
# stderr: "thermal: no supported tool data found"; exit 1

# After: silent prompt
HOME=$(mktemp -d) thermal statusline
# no output; exit 0
```

**The fix**:
Use `thermal live --json` or `thermal live --json --stream` in scripts; use `thermal statusline --json` when a caller needs a value. Empty prompt environments no longer require a special case.

**Escape hatch**:
`thermal sources` diagnoses missing paths; `thermal statusline --verbose` prints one `statusline: no data (15 sources checked)` line when a caller needs to know why the prompt is empty. `thermal live --json --stream | head -1` exits cleanly after the first snapshot.

---

### Changed: Reports Drop Columns Below 80 and Live Refuses Below 60
<a id="changed-width-contract"></a>

**Affected if**:
- **Search pattern**: `grep -rnE 'COLUMNS|--width|--weeks' .`
- **Detection signal**: `COLUMNS=60 thermal daily` now renders a compact Period/Tokens/Cost table; `COLUMNS=50 thermal daily` exits `2` with `thermal: daily needs at least 56 columns (terminal is 50)`; `thermal --width 39 daily` exits `2` with `--width wants a number between 40 and 240`.

**What changed and why**:
Reports used fixed 100+ column layouts regardless of the terminal, so an 80-column terminal wrapped or clipped tables and charts hinted at a `--top` flag that did not exist. Width now resolves once per invocation (`--width`, then `COLUMNS`, then the TTY, then 100), caps at 96 for pasteability, and drops low-priority columns in a documented order. Each command has a minimum width and refuses below it with a fix. Heatmaps auto-fit their week count and print a caption. `thermal live` uses a stacked layout from 60–79 columns and refuses below 60; `--json` ignores width entirely.

**Before / After**:
```bash
# Before: 104-column table at 80 columns
COLUMNS=80 thermal daily

# After: compact table, ≤80 columns
COLUMNS=80 thermal daily

# After: explicit minimum refusal
COLUMNS=50 thermal daily
thermal: daily needs at least 56 columns (terminal is 50)
  widen the terminal, pipe to a file, or set --width 100
```

**The fix**:
Set `--width 100` (or `COLUMNS=100`) for the full layout, use fewer columns for compact output, or use `--json` for machine data. For live scripts use `thermal live --json`.

**Escape hatch**:
`--json` is byte-stable regardless of width. `--width 100` restores the full report layout in a narrow terminal.

---

### Changed: Leaderboard States Its Ranking, Legend, and Cost Arithmetic
<a id="changed-leaderboard-ranking-legend"></a>

**Affected if**:
- **Search pattern**: `grep -rnE 'Recorded cost:|Estimated:|No pricing for:|Strk|Best|Days' .`
- **Detection signal**: `thermal --no-color` now starts with `Ranked by current streak · all time`, includes a `Strk/Best/Days` and cost-marker legend, and ends with `Cost: $X recorded + ~$Y estimated = ~$Z shown`. The old `Recorded cost:` and `No pricing for:` lines are gone.

**What changed and why**:
The leaderboard ranked by streak but never said so, and its `Strk`/`Best`/`Days` headers and `$`/`~$`/`—`/`≥$` cost markers were undocumented. A reader could not tell whether a row was ranked by tokens or streak, or why an `~$` value differed from a report total. The caption now names the ranking and window, the legend defines every header and marker, the footer states recorded + estimated = shown, and unpriced models are named with their volume. The JSON envelope adds `window`, `ranking`, `costBasis`, and `unpricedModels`.

**Before / After**:
```bash
# Before: ranking and legend implicit
thermal
# ... Strk Best Days ... ~$ ...

# After: self-documenting
thermal
Ranked by current streak · all time
Legend: Strk = current streak · Best = longest streak · Days = active days
Cost: $243.90 recorded + ~$261.50 estimated = ~$505.39 shown
Unpriced models: atria-dawn-preview (2.5K tok)
```

**The fix**:
Text consumers can keep parsing the table, or switch to the JSON fields:
```bash
thermal --json | jq '.ranking, .costBasis, .unpricedModels'
```

**Escape hatch**:
`--no-estimate` states `estimation disabled` and makes `shown` equal `recorded`. `--sort tokens`/`--sort cost` change the named ranking.

---

### Changed: Project and Model Totals Use Distinct-Day Denominators
<a id="changed-total-row-semantics"></a>

**Affected if**:
- **Search pattern**: `grep -rnE 'P-Days|M-Days|projectDays|modelDays|tools total|Total +[A-Z]' .`
- **Detection signal**: `thermal projects --no-color` no longer prints `Agy · CMD +10` in the total row; it prints `N tools`, and the footer states `P-Days are project-days (N); distinct calendar days are M`. `thermal projects --json` now has `.totals.projectDays`, `.totals.toolCount`, and distinct `.totals.activeDays`; models has `.totals.modelDays` and `.totals.toolCount`.

**What changed and why**:
The projects total row summed per-project tool names and per-project active days, so a 13-tool machine printed `Agy · CMD +10` and `505` project-days next to a leaderboard that honestly said `96` calendar days. The totals now contain only additive values: tokens, cost, and project-days/model-days are sums; `activeDays` is the distinct calendar-day count; `toolCount` is the distinct tool count. The footer names the denominator so the two numbers cannot be confused. `--top` still caps rows only; totals are full-window.

**Before / After**:
```bash
# Before: non-additive tool list and ambiguous day sum
Total   Agy · CMD +10   23.7B   ~$505    505

# After: additive columns with named denominators
Total   12 tools        23.7B   ~$505    507
Tools are counted per project; P-Days are project-days (507); distinct calendar days are 129.
```

**The fix**:
Use the explicit metadata:
```bash
thermal projects --json | jq '.totals | {activeDays, projectDays, toolCount}'
thermal models --json | jq '.totals | {activeDays, modelDays, toolCount}'
```

**Escape hatch**:
Totals stay full-window under `--top`; per-project and per-model details remain in `data`.

---

### Changed: Analytics Denominators and Audit Prose
<a id="changed-analytics-prose"></a>

**Affected if**:
- **Search pattern**: `grep -rnE 'Daily Mean|Biggest win: Critical|all [0-9]+ top days above' .`
- **Detection signal**: `thermal stats --no-color` now shows `Mean / active day` and `Basis: N active days`; `thermal trend --no-color` shows `Mean / calendar day` and `Basis: N calendar days`; audit prints `Score components:`, `Largest penalty:`, and `Biggest win: Address …` on separate lines.

**What changed and why**:
`stats` and `trend` both labelled a value `Mean`, but stats divided by active days while trend divided by calendar days, so the same data could show values four times apart with no explanation. The labels and JSON now carry `meanBasis` (`active-day`/`calendar-day`), trend includes `activeDays`, and the projection exposes `bandMethod:"residual-sd-1.96"`. Outlier prose is one sentence plus a MAD definition; empty histogram bins are omitted with a count; audit splits its score components, largest penalty, and verb-led biggest win, and defines PASS/WARN/CRIT.

**Before / After**:
```bash
# Before
Mean                 242.7M
Biggest win: Critical: AGENTS.md ...
Outliers above 226.1M (median + 2 MAD): all 5 top days above, plus 22 more

# After
Mean / active day    242.7M
Basis: 97 active days
Outliers: 29 days above 227.1M (median + 2 MAD); top 5 shown.
MAD = median absolute deviation.
Biggest win: Address AGENTS.md ...
```

**The fix**:
Use the basis labels or JSON:
```bash
thermal stats --json | jq '.meanBasis'
thermal trend --json | jq '{meanBasis, activeDays, band: .projection.bandMethod}'
```

**Escape hatch**:
The underlying values are unchanged; only the denominator labels and prose are new. `--json` carries the machine-readable basis.

---

## Upgrading to v0.20.1

### Changed: Renamed `command-code` Canonical Display Name to `CMD`
<a id="changed-rename-command-code-to-cmd"></a>

**Affected if**:
- **Search pattern**: `grep -rn '"command-code"' .` or scripts parsing `.name` / `.toolTotals["command-code"]` from `thermal --json` or `thermal live --json`.
- **Detection signal**: In `thermal --json`, the `.results[].name` field for the command-code tool is now `"CMD"` instead of `"command-code"`. In `thermal live`, the tool column displays `"CMD"` instead of truncated `"command-c…"`.

**What changed and why**:
Thermal previously used `"command-code"` as both the CLI tool identifier and its canonical display name. Because `"command-code"` is 12 characters long, terminal tables with constrained column widths (such as the 10-column tool field in `thermal live`) truncated the name with an ellipsis (`command-c…`). The canonical display name is now `"CMD"`, allowing it to fit without truncation across all views.

**Before / After**:
```bash
# Before (Truncated tool name in live table)
# TIME      TOOL               STEPS  MODEL         PROJECT         COST
# 08:02:43  command-c…     +104 step  deepseek-v4…  —              $0.00

# After (v0.20.1+)
# TIME      TOOL               STEPS  MODEL         PROJECT         COST
# 08:02:43  CMD            +104 step  deepseek-v4…  —              $0.00
```

**The fix**:
If external tooling or scripts assert on `.name == "command-code"` in JSON payloads, update the match to check for `"CMD"` or use the stable tool identifier field `.tool == "command-code"` (which remains unchanged):
```bash
# Match by tool identifier rather than display name
thermal --json | jq '.results[] | select(.tool == "command-code")'
```

**Escape hatch**:
Both `command-code` and `cmd` CLI arguments continue to be accepted identically by all commands (`thermal cmd`, `thermal command-code`, `thermal live cmd`).

## Upgrading to v0.20.2

### Changed: Token-Boundary Receipt Evidence Classification
<a id="changed-receipt-evidence-token-boundary"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal receipt' .` or automated consumers of `thermal receipt --json` that classify session outcomes from local agent transcripts.
- **Detection signal**: Sessions whose only test evidence is a compound shell command (`cd repo && go test ./...`, `false && go test ./... || true`) now report `Tier 3 (UNVERIFIED)` or `Tier 2 (CLAIMED)` instead of `Tier 1 (VERIFIED)`; conversely, a single `go test ./...` invocation with no recorded exit code but a completed `The command exited with code 0` line or a clean pytest summary (`10 passed, 0 failed`) now reports `Tier 1 (VERIFIED)` instead of `Tier 3 (FAILED)`.

**What changed and why**:
Receipt evidence previously recognized runner names by substring presence anywhere in the command text, so a short-circuited mention like `false && go test ./... || true` inherited the outer exit status and certified a test that never ran. At the same time, generic failure words were matched before complete success lines, so a finished `The command exited with code 0` result or a clean pytest summary was misclassified as a failure. Evidence now requires one recognized invocation, matched on token boundaries, whose exit status is either observed directly or established by a complete result line; compound shell lists, pipelines, substitutions, and control operators remain unknown unless the source adapter records separate command/result pairs.

**Before / After**:
```bash
# Before (v0.20.1: skipped compound test certified; completed success read as failure)
thermal receipt --json
# Output: { "sessionId": "s-skip", "status": "VERIFIED",   "tier": "Tier 1 (Verified)",   "testsPassed": 1 }
#         { "sessionId": "s-done", "status": "FAILED",     "tier": "Tier 3 (Failed)",     "testsFailed": 1 }
```

```bash
# After (v0.20.2+: single observed invocations only)
thermal receipt --json
# Output: { "sessionId": "s-skip", "status": "UNVERIFIED", "tier": "Tier 3 (Unverified)", "testsPassed": 0 }
#         { "sessionId": "s-done", "status": "VERIFIED",   "tier": "Tier 1 (Verified)",   "testsPassed": 1 }
```

**The fix**:
Run verification commands as single explicit invocations (`go test ./...` rather than `cd repo && go test ./...`) so the recorded exit status belongs to the test run itself, and prefer completed result lines or recorded exit codes over narrative text. When auditing receipts, filter on `"status": "VERIFIED"` in `thermal receipt --json` to review certified sessions.

**Escape hatch**:
None. Compound commands and unmatched result records remain visible as `Tier 2 (CLAIMED)` or `Tier 3 (UNVERIFIED)` rows, and no configuration restores exit-status inheritance across shell lists.

### Changed: Workload Replay Volume Conservation and Price Coverage
<a id="changed-replay-volume-and-price-coverage"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal.*replay' .` or automated parsers consuming `thermal replay --json`.
- **Detection signal**: `workload.unclassifiedTokens` and `plans[].pricingStatus` appear in replay JSON; `workload.totalTokens` now equals the canonical in-window daily sum even when a day's model map covers only part of the volume; a target with no catalog rate reports `monthlyCost: 0` with `costDelta: 0` and no `isRecommended` flag instead of a −100% savings claim.

**What changed and why**:
Workload replay previously derived its volume from whichever model map happened to be present, so a partially attributed day could shrink the simulated workload (for example 1000 source tokens becoming 100) and unclassified model tokens could vanish entirely. It also treated a missing target rate as free, printing a full savings recommendation against a $0.00 comparison. Replay now always conserves canonical volume, discloses unclassified tokens, and labels each comparison's pricing completeness so an incomplete estimate can never read as a savings claim.

**Before / After**:
```bash
# Before (v0.20.1: partial attribution dropped volume; missing rate became a $0 plan)
thermal replay --against not-in-catalog --json
# workload.totalTokens: 100 · plans[0].monthlyCost: 0 · costDeltaPercent: -100 · isRecommended: true
```

```bash
# After (v0.20.2+: volume conserved; incomplete pricing labeled, never recommended)
thermal replay --against not-in-catalog --json
# workload.totalTokens: 1000 · workload.unclassifiedTokens: 900
# plans[0].pricingStatus: "missing" · monthlyCost: 0 (numeric sentinel) · costDelta: 0 · isRecommended: absent
```

**The fix**:
When parsing replay JSON, key on `pricingStatus` before trusting `monthlyCost` or `costDelta`: `complete` rows carry real comparisons, `partial` rows carry the known-category subtotal, and `missing` rows use zero only as a numeric sentinel (the terminal view prints `Unpriced` or `Partial estimate` instead of a dollar figure).

**Escape hatch**:
None. Volume conservation and price-coverage disclosure are correctness requirements; complete comparisons keep their existing numeric keys and values unchanged.

### Changed: Rolling-Window Capacity Verdicts from Daily Telemetry
<a id="changed-replay-rolling-capacity-unknown"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'capacityVerdict' .` or scripts consuming `thermal replay --json` and parsing the Capacity Verdict column.
- **Detection signal**: subscription rows with a rolling-window limit and no daily-ceiling exceedance now report `"UNKNOWN"` instead of `"PASS"`; `burstThrottles` is always `0`; verdict details say `rolling spikes unasserted` (daily failures) or `intraday evidence unavailable` (UNKNOWN). Consumers matching only `PASS`/`DEGRADED`/`FAIL` see those plans fall through.

**What changed and why**:
A day's token sum is not a five-hour rolling peak, yet `SimulatePlan` compared daily totals against `RollingWindowLimit` and counted the exceedances as "burst" spikes. The same daily total could therefore become a fabricated burst count for a workload actually spread across the day, or a `PASS` that could not exclude a cross-midnight peak. Plans with a rolling-window limit are now reported as `UNKNOWN` unless a configured daily ceiling is exceeded, rolling spikes are never asserted from daily rows, and `UNKNOWN` rows are never recommended.

**Before / After**:
```bash
# Before (v0.20.1: daily sum treated as a rolling peak)
thermal replay --against chatgpt-plus --json
# plans[0].capacityVerdict: "DEGRADED" · verdictDetail: "throttled 1/28 days (4%) · 1 burst limit spikes"
```

```bash
# After (v0.20.2+: rolling capacity disclosed as unassessable from daily telemetry)
thermal replay --against chatgpt-plus --json
# plans[0].capacityVerdict: "UNKNOWN"
# plans[0].verdictDetail: "rolling-window capacity cannot be assessed from daily totals; intraday evidence unavailable"
# plans[0].burstThrottles: 0
```

**The fix**:
Treat `UNKNOWN` as "cannot be assessed from daily telemetry": do not rank or recommend those rows (recommendation selection requires `PASS` with complete pricing), and for capacity comparisons in scripts prefer plans without rolling limits (`agy-pro`, `opencode-pro`, `kimi-k2`) or compare against the configured daily ceilings. A daily-ceiling exceedance still reports `DEGRADED`/`FAIL` with rolling spikes explicitly unasserted.

**Escape hatch**:
None. Daily rows cannot observe intraday bursts; exact five-hour simulation waits for a separate timestamp-coverage design. UNKNOWN rows remain visible with their daily figures so no capacity claim silently disappears.

### Changed: `thermal replay --against` Resolves Built-In Plan Identifiers
<a id="changed-replay-against-plan-resolution"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'replay.*--against' .` or scripts parsing `thermal replay --against <id> --json`.
- **Detection signal**: `--against claude-pro` (or any built-in plan id or exact plan name) now renders the subscription plan row — `plans[0].id` is `claude-pro`, `type` is `"sub"`, `pricingStatus` is `"complete"` — instead of a pay-as-you-go `"claude-pro (API)"` row with `pricingStatus: "missing"`. Values that are not exact plan ids or names keep the model rate-card behavior unchanged.

**What changed and why**:
`--against` targeted model rate cards only, so naming a built-in plan produced an unpriced pseudo-model row with no hint that `--compare <plan>` was intended. A value that exactly matches a plan id or name (case-insensitive, no prefix matching) now resolves to that plan, making `--against claude-pro` and `--compare claude-pro` identical for that plan; model targeting is unchanged for every other value.

**Before / After**:
```bash
# Before (v0.20.2-: the plan id became an unpriced model)
thermal replay --against claude-pro --json
# plans[0].id: "claude-pro" · name: "claude-pro (API)" · type: "payg" · pricingStatus: "missing"
```

```bash
# After (v0.20.2+: the plan id resolves to the plan itself)
thermal replay --against claude-pro --json
# plans[0].id: "claude-pro" · name: "Claude Pro ($20)" · type: "sub" · pricingStatus: "complete"
```

**The fix**:
If a script depended on `--against <plan>` producing an unpriced model row, switch to `--compare <plan>` for plan comparisons or use the target model's rate-card id with `--against`. Resolution is exact-only: `--against claude` remains a model lookup and never captures `claude-pro`.

**Escape hatch**:
None needed for plan ids; model targeting is unchanged for any value that is not an exact plan id or name.

### Changed: Live Monitor Project Baseline Preservation and Delta Clamping
<a id="changed-live-project-baseline-clamping"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal.*live' .` or automated consumers of `thermal live --json` or
  `thermal live --stream`.
- **Detection signal**: `RecentEvents` no longer emits spurious bursts with lifetime project totals on
  idle ticks or midnight crossings; `recentEvents[].turns` within an observation window strictly
  equals `sessionTurns`; project token and turn deltas are clamped to tool-level increments.

**What changed and why**:
Thermal's real-time live monitor previously collapsed project baseline maps to single-day slices on
idle ticks and wiped them during midnight day crossings, causing subsequent polls to re-emit cumulative
lifetime project volume as new bursts. In addition, project deltas were not clamped to tool-level
increments, allowing anomalous project telemetry to generate events exceeding actual tool burn. The
tracker now accumulates project baselines across all days on both active and idle ticks, preserves
cumulative baselines across midnight, and unconditionally clamps project deltas to tool totals so
recent events strictly match session metrics.

**Before / After**:
```bash
# Before (Idle tick collapsed project baselines; next turn replayed lifetime volume)
thermal live --json
# Output: {"sessionTurns": 1, "recentEvents": [{"tool": "Agy", "turns": 125600, "project": "lab"}]}

# After (Cumulative baselines preserved; only incremental deltas emitted)
thermal live --json
# Output: {"sessionTurns": 1, "recentEvents": [{"tool": "Agy", "turns": 1, "project": "lab"}]}
```

**The fix**:
Update monitoring consumers and automation expecting inflated project burst deltas to rely on accurate
incremental deltas. Verify that the sum of `recentEvents[].turns` matches `sessionTurns` across
observation windows.

**Escape hatch**:
None. Cumulative baseline preservation and delta clamping are correctness requirements; raw lifetime
totals remain accessible in `toolTotals` and per-tool daily reports (`thermal <tool> daily`).

---

## Upgrading to v0.21.0

### Changed: Command Code Token and Project Attribution
<a id="changed-command-code-token-attribution"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal.*\(cmd\|command-code\)' .` or automated consumers of `thermal cmd --json`, `thermal projects --json`, and `thermal live`.
- **Detection signal**: `CMD` moves from the Activity Hunters table to Token Warriors; `.summary.lifetimeTokens` changes from a timestamped-line count to real input+output tokens; `.summary.cost` becomes the recorded `costUsd` sum; `.daily[].input`, `.daily[].output`, `.daily[].cache`, and `.daily[].models` are populated; `thermal cmd projects` gains rows attributed to the session header `cwd`; `thermal live` reports CMD token burn instead of message steps.

**What changed and why**:
Thermal previously treated Command Code as an activity-only source and counted every timestamped transcript line as a "token", even though modern transcripts already carry per-assistant `usage` (`inputTokens`, `outputTokens`, `cacheReadTokens`, `cacheWriteTokens`, `costUsd`), a per-message model, and the session working directory. The loader now reads those fields, subtracts cache read/write from `inputTokens` (they were nested) to keep token types disjoint, counts assistant messages as turns, and emits `ProjectDay` rows keyed on the header `cwd`. Legacy flat transcripts without usage still contribute assistant turns but never enter token totals. This removes the double count of user and assistant lines as tokens and exposes recorded spend without estimation.

**Before / After**:
```bash
# Before (v0.20.x: line counts exposed as tokens, no cost or project attribution)
thermal cmd --json
# lifetimeTokens: 25800 · inputTokens: 0 · outputTokens: 0 · cacheTokens: 0 · cost: 0
# daily[].tokens: line count
thermal cmd projects
# no project rows
```

```bash
# After (v0.21.0+: real usage, recorded cost, per-project rows)
thermal cmd --json
# lifetimeTokens: 3500000000 · inputTokens: ... · outputTokens: ... · cacheTokens: ... · cost: 89.42
# daily[].tokens: ... · daily[].turns: assistant messages
thermal cmd projects
# Project · Tools: CMD · Tokens · Cost · Days
```

**The fix**:
Update dashboards, alerts, and scripts that compared Command Code's `lifetimeTokens` to message counts or expected it under Activity Hunters. Read `.summary.inputTokens`, `.summary.outputTokens`, `.summary.cacheTokens`, and `.summary.cost` for spend; use `.daily[].turns` for assistant-call counts and `thermal cmd projects` for per-repository FinOps. Compare recorded `costUsd` rather than token sums when measuring spend, because input tokens include re-sent context and cache is priced separately.

**Escape hatch**:
None. The old line-count metric conflated user and assistant messages with billable tokens and cannot be restored; legacy transcripts without usage remain visible as assistant turns with zero tokens.

### Added: Commercial Use License Pricing in `thermal license`
<a id="added-commercial-license-pricing"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal license' .` or parsers of `thermal license --json`.
- **Detection signal**: `thermal license --json` now contains `commercial_licensing.commercial_use` (`price_usd: 30`, `unit: "developer/year"`, `url`, `pricing_page`, `terms`, `includes`, `excludes`) and `commercial_licensing.enterprise_oem`; `commercial_licensing.use_cases` no longer lists `dedicated_support_sla`. Text output now prints the $30 Commercial Use License and a separate Enterprise & OEM contact.

**What changed and why**:
The commercial offer was previously contact-only, with no price, terms, or buyer-facing path. Thermal now publishes a **$30 per developer per 12-month term** Commercial Use License for internal commercial use, with full terms in [`COMMERCIAL-LICENSE.md`](../COMMERCIAL-LICENSE.md), and keeps embedding, hosted services, fleet/security review, custom loaders, and support commitments under separate Enterprise & OEM terms. The JSON adds the offer fields and moves SLA language out of the general commercial list so the command no longer promises an SLA in the $30 tier.

**Before / After**:
```bash
# Before (v0.20.x: contact-only, no price; SLA listed in the general commercial cases)
thermal license --json
# commercial_licensing.use_cases: ["proprietary_embedding","closed_source_saas","enterprise_ospo_compliance","dedicated_support_sla"]
```

```bash
# After (v0.21.0+: published $30 offer and separate enterprise terms)
thermal license --json
# commercial_licensing.commercial_use.price_usd: 30
# commercial_licensing.commercial_use.unit: "developer/year"
# commercial_licensing.commercial_use.pricing_page: "https://thermal.jadmadi.net/pricing.html"
# commercial_licensing.enterprise_oem.contact: "contact@jadmadi.net"
```

**The fix**:
Read `commercial_licensing.commercial_use` for the published offer and `commercial_licensing.enterprise_oem` for embedding, hosted-service, or SLA discussions. If a script depended on `dedicated_support_sla` in `use_cases`, treat support commitments as absent from the $30 tier and use the enterprise contact.

**Escape hatch**:
The legacy `commercial_licensing.available`, `contact`, `github`, and `use_cases` keys remain. Only the SLA claim was removed from `use_cases`; no AGPL rights or community-edition behavior changed.

---

## Upgrading to v0.16.x

### Added: GitHub Flavored Markdown Output for Receipt and Yield Reports
<a id="added-markdown-pr-receipts-and-yield"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal.*--format' .`
- **Detection signal**: Error output `thermal: --format must be text, json, md, or markdown` when an invalid format is specified, or `thermal: --format md only applies to the receipt and yield commands` when `--format md` is combined with unsupported commands, or `thermal: --json cannot be combined with --format md`.

**What changed and why**:
Thermal now supports `--format md` and `--format markdown` for `thermal receipt` and `thermal yield`, rendering copy-paste ready GitHub Flavored Markdown blocks with summary blockquotes, structured metric tables, and collapsible details for PR descriptions and issue audits. Previously, script automation requiring Markdown had to transform raw ASCII box-drawing characters or parse `--json` output manually. The `--format` flag strictly validates input options (`text`, `json`, `md`, `markdown`), rejects invalid commands, and prevents conflicting `--json` combinations.

**Before / After**:
```bash
# Before (Only ASCII terminal tables or JSON available)
thermal receipt
# Output: ASCII box-drawing border table

# After (GitHub Flavored Markdown ready for PR summaries)
thermal receipt --format md
# Output: ### 🧾 Verifiable Work Receipts ...
```

**The fix**:
Use `--format md` or `--format markdown` directly on `thermal receipt` or `thermal yield`:
```bash
thermal receipt --format md
thermal yield --format md
```

**Escape hatch**:
Omit `--format` or specify `--format text` to retain the default ANSI/ASCII terminal output, or use `--format json` / `--json` for machine-readable payloads.

---

### Added: Instant Shell Statusline and Prompt Helper (`thermal statusline`)
<a id="added-terminal-statusline-helper"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal.*\(statusline\|prompt\)' .`
- **Detection signal**: Error output `thermal: report options do not apply to the statusline command` when report flags like `--since`, `--until`, `--last`, or `--chart` are passed to `statusline`, or `thermal: --sort does not apply to the statusline command`.

**What changed and why**:
Thermal introduces an ultra-fast (<3ms) `thermal statusline` (alias: `thermal prompt`) command purpose-built for shell prompt integration (Starship, Zsh, Bash, Tmux). Unlike the full leaderboard or live monitor, `statusline` outputs a single composable line with the current streak, today's token volume, and active tool using an atomic disk-backed snapshot (`~/.cache/thermal/statusline.json`), avoiding expensive disk and SQLite rescans on every prompt draw.

**Before / After**:
```bash
# Before (Shell prompts had to parse full JSON or run slow leaderboard)
thermal --json | jq -r '...' # ~1.1s latency, too slow for shell prompts

# After (v0.16.0+)
thermal statusline           # <3ms warm cached draw
thermal prompt --plain       # Plain ASCII without emojis
thermal statusline --nerd    # Nerd Font flame glyph (󰈸)
thermal statusline --json    # Structured JSON snapshot for custom prompt engines
```

**The fix**:
Integrate `thermal statusline` directly into your shell configuration or prompt manager (e.g. Starship custom command module):
```toml
# ~/.config/starship.toml
[custom.thermal]
command = "thermal statusline"
when = "command -v thermal"
format = "[$output]($style) "
style = "bold yellow"
```

**Escape hatch**:
Pass `--fresh` to bypass the 120-second snapshot cache and force an immediate recomputation across all tool databases:
```bash
thermal statusline --fresh
```

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

### Changed: Live Session Starts at Launch and History Is Carried In
<a id="changed-live-baseline-separation"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'sessionTokens\|todayTokens\|thermal live' .` or consumers of `thermal live --json` and `thermal live --json --stream`.
- **Detection signal**: On startup, `sessionTokens` is 0 instead of equal to `todayTokens`; `recentEvents` no longer contains a synthetic row whose tokens equal the day total; the snapshot now includes `carriedInTokens`, `carriedInTurns`, `carriedInCost`, and `carriedInAt`. `--fresh` is still accepted but no longer changes live session seeding.

**What changed and why**:
The first live poll used to copy the current day's persisted totals into `sessionTokens`, `sessionTurns`, and `sessionCost`, set a warm flame, and emit one event per tool at the current clock. That made the opening history read look like a live burst, duplicated Today and Session, and inflated rate/flame aggregates. The first poll now records that opening balance as `carriedIn*`, leaves session counters at zero, and emits no burst events; later polls increment session only from observed deltas. `--fresh` remains valid for statusline cache bypass and live compatibility but no longer selects seeded live counters.

**Before / After**:
```bash
# Before (startup copied today's 651.4M into Session and emitted a 651.4M burst)
thermal live --json
# {"todayTokens":651400000,"sessionTokens":651400000,"recentEvents":[{"tokens":651400000,...}]}

# After (opening balance is carried in; Session starts at zero)
thermal live --json
# {"todayTokens":651400000,"carriedInTokens":651400000,"sessionTokens":0,"recentEvents":[]}
```

**The fix**:
Consume `carriedIn*` for the opening balance and `session*` for launch-relative deltas. Do not treat a startup `recentEvents` row as live activity; the TUI renders a `Carried in:` line instead. If a script needs the full local-day total, continue to use `todayTokens`, `todayCost`, and `todayTurns`.

**Escape hatch**:
For persisted daily totals independent of live tracking, use `thermal daily`, `thermal stats`, or the `today*` fields. To force a fresh process baseline, restart `thermal live`; `ResetSession` (`[c]`) only clears observed session counters and preserves the carried-in line.

---

### Changed: Live Ticker Units, Cost Status, and Day Rollups
<a id="changed-live-ticker-units"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'TOK/STEP\|recentEvents\|costKind\|todayProjects' .` or consumers of `thermal live --json`.
- **Detection signal**: ticker headers now expose `TOKENS` and `STEPS` separately plus `AGE`; `TOK/STEP` is gone. Event JSON includes `costKind` (`recorded`, `estimated`, `unpriced`) and snapshots include `unpricedTokens`, `unpricedEvents`, `todayProjects`, and `todayModels`. The KPI grid says `Cache hit:` and no longer prints a redundant `Turns:` line.

**What changed and why**:
The old `TOK/STEP` cell mixed token and step units, so rows could not be compared, and cost values collapsed to `<$0.01` or `$0.00`, conflating tiny spend, free, and unpriced sources. The live view now uses separate numeric columns, an explicit cost status with four decimals below one cent, `—` for unpriced, and `~` only for estimates. It also adds age/outlier markers and today project/model cost rollups so an idle screen still answers where the money went.

**Before / After**:
```text
# Before
TOK/STEP       COST
+651.4M tok    ~$4.43
+11.5K step    $0.00

# After
TOKENS    STEPS    AGE    COST
+651.4M   —        29s    ~$4.43
—         +11.5K   29s    —
```

**The fix**:
Read `tokens` and `turns` as separate fields; use `costKind` to format `recorded` as `$`, `estimated` as `~`, and `unpriced` as `—`. Use `todayProjects` and `todayModels` for the idle decision view instead of re-deriving costs from ticker rows.

**Escape hatch**:
For the old combined display, derive your own string from `tokens`/`turns`. For raw persisted totals, use `thermal daily`, `thermal projects`, or `thermal models`.

---

### Changed: Live Peak Uses One-Second Buckets
<a id="changed-live-one-second-peak"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'peakRollingTok\|peakRollingTurn\|rollingTokens\|thermal.*live' .`.
- **Detection signal**: `peakRollingTok` and `peakRollingTurn` are replaced by `peakTokensPerSec` and `peakStepsPerSec`; snapshots now include `bucketSeconds: 1` and `windowSeconds: 60`. `rollingTokens` and `rollingTurns` are one-second buckets, so a multi-second poll delta is split across the seconds it spans instead of becoming one oversized peak.

**What changed and why**:
The old tracker advanced empty buckets and then added the whole poll delta to the latest bucket, so a single event's token count was displayed as a per-second peak. Peak is now the maximum one-second bucket, and Rate and Speed both derive from the same 60-second ring. Sub-second polls accumulate in the current second; gaps longer than 60 seconds contribute only their last 60 seconds to the rolling window, while session counters keep the full delta. Displayed sub-one-per-second speeds use three decimals so multiplying by 60 matches the displayed per-minute rate within 0.05.

**Before / After**:
```text
# Before: 20K tokens observed over 2 seconds
Peak: 20K tok/s
Rate: 20K/min
Speed: 333 tok/s

# After: the same delta split into two one-second buckets
Peak: 10K tok/s (1s bucket)
Rate: 20K/min
Speed: 333 tok/s
```

**The fix**:
Consume `peakTokensPerSec` and `peakStepsPerSec` as one-second maxima, and use `bucketSeconds`/`windowSeconds` to describe the series. Do not infer a rate from a single event; use `burnTokensPerMin`/`burnTokensPerSec` and `burnTurnsPerMin`/`burnTurnsPerSec`.

**Escape hatch**:
For poll-sized deltas, read `recentEvents[].tokens` and `recentEvents[].turns`. For static period totals, use `thermal daily`, `thermal stats`, or `thermal models`.

---

### Changed: Live Flame Ceiling and Metric
<a id="changed-live-flame-ceiling"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'activityIntensity\|flameIntensity\|flame-ceiling' .` or consumers of `thermal live --json`.
- **Detection signal**: `activityIntensity` is removed. `flameIntensity` is now `rate / flameCeiling` for one metric, and snapshots include `flameMetric` (`tokens` or `steps`) and `flameCeiling`. `--flame-ceiling` is available on `thermal live`.

**What changed and why**:
The old flame added 0.5 per token poll and 0.4 per turn poll, decayed by 0.85 per second, and the TUI took the larger of the two values. That mixed token and step events and made 23 step/min look hotter than 3.4M tok/min. Flame now uses one metric at a time: tokens when token burn exists, otherwise steps. Intensity is `60-second rate / ceiling`, with defaults of 9,000,000 tok/min and 60 step/min, and the state thresholds remain 60% BURNING, 20% WARM, and 2% COOLING. The ceiling is configurable with `--flame-ceiling=N`, where `N` is a positive number in the active metric's units and accepts `K`, `M`, and `B` suffixes.

**Before / After**:
```text
# Before: event-count flame, mixed metrics
Flame: [██████████░░░░]  72%  BURNING   # 23 step/min
Flame: [█████░░░░░░░░░]  37%  WARM      # 3.4M tok/min

# After: rate / ceiling, one metric
Flame: [█████░░░░░░░░░]  38% of 60 step/m  WARM
Flame: [█████░░░░░░░░░]  38% of 9.0M tok/m WARM
```

**The fix**:
Read `flameMetric` and `flameCeiling` to interpret `flameIntensity`; do not read `activityIntensity`. To change the scale, pass `--flame-ceiling=9M` or `--flame-ceiling=60`; the value applies to the active metric and is reflected in both the TUI and JSON.

**Escape hatch**:
For raw rates independent of the gauge, use `burnTokensPerMin`, `burnTokensPerSec`, `burnTurnsPerMin`, and `burnTurnsPerSec`. The flame is a presentation ratio and does not change token or cost totals.

---

### Changed: Live Ticker Burst Rollup and Cost Sort
<a id="changed-live-ticker-rollup"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'Top spenders\|sortMode\|LiveEventGroup\|×' internal/tui`.
- **Detection signal**: the live ticker now collapses repeated events sharing tool, model, and project within 60 seconds into one `TOOL×N` row with summed tokens, steps, and cost. Pressing `[s]` toggles time/cost ordering; cost mode adds a `Top spenders (60s)` project line. `thermal live --json` is unchanged; `recentEvents` remains the raw event stream.

**What changed and why**:
A project/model that emitted several near-identical bursts used to occupy several rows, pushing the useful decision data off screen. The ticker now groups those bursts, reports the count in the TOOL cell, and sums the group's values. Cost mode answers where the money went without requiring the user to add rows mentally. Mixed priced/unpriced groups show `≥$X`; all-unpriced groups show `—`.

**Before / After**:
```text
# Before
23:59:43  CMD  +1.0M  deepseek  thermal-streak  $0.10
23:59:41  CMD  +500K  deepseek  thermal-streak  $0.10
23:59:40  CMD  +200K  deepseek  thermal-streak  $0.10

# After, time sort
23:59:43  CMD×3  +1.7M  deepseek  thermal-streak  $0.30

# After, cost sort ([s])
Top spenders (60s): thermal-streak $0.30 · scripts —
```

**The fix**:
No JSON migration is required. The grouping and ordering are TUI projections over the existing `recentEvents` array. If you need raw rows, read `recentEvents`; if you need today's decision totals, use `todayProjects` and `todayModels`.

**Escape hatch**:
Use `thermal live --json` for the ungrouped event stream, or `thermal projects` and `thermal models` for persisted period totals.

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

---

### Removed: Interactive Dashboard and Dense FinOps Grid
<a id="removed-dashboard-tui"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal dashboard\|stats --dense\|--distribution\|internal/tui/dense' .`.
- **Detection signal**: `thermal dashboard` exits 1 with a migration notice. `--dense`, `--distribution`, `--dist`, and `--sparse` exit 1 with a retirement notice. `internal/tui` now contains only the live monitor files, and `thermal stats` prints the distribution report.

**What changed and why**:
The five-tab `thermal dashboard` duplicated static report commands, and its 9-box dense grid depended on the yield, receipt, and replay surfaces that are being retired. Maintaining a parallel TUI added substantial complexity without unique data. `thermal live` remains the only interactive terminal monitor, while static commands remain the scriptable and JSON-friendly surface.

**Before / After**:
```bash
# Before
thermal dashboard
thermal stats --dense
thermal stats --distribution

# After: exits 1 with migration guidance
thermal dashboard
# thermal: the interactive dashboard has been retired...
# Use: thermal live, thermal, thermal projects, thermal stats

thermal stats
```

**The fix**:
Use `thermal live` for real-time monitoring and static commands (`thermal`, `daily`, `weekly`, `monthly`, `projects`, `models`, `mix`, `stats`, `trend`) for historical queries. `thermal stats` itself now prints the distribution report.

**Escape hatch**:
For machine-readable data, use the static commands with `--json`. For live streaming, use `thermal live --json --stream`.

---

### Removed: Embedded Localhost Web Server
<a id="removed-web-server"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal serve\|thermal web\|--port\|--host\|--open\|internal/server' .`.
- **Detection signal**: `thermal web` and `thermal serve` exit 1 with a retirement notice. `--port`, `--host`, `--open`, and `--no-open` exit 1 when passed after a command. `internal/server/` no longer exists, and `thermal live --json --stream` is the supported headless telemetry path.

**What changed and why**:
The embedded HTTP server, SSE hub, and bundled HTML dashboard conflicted with Thermal's terminal-native anti-goals. The same telemetry is available as a snapshot or a continuous NDJSON stream from `thermal live`, which can be consumed by external tools without embedding a web stack in the binary. Static report commands also continue to support `--json`.

**Before / After**:
```bash
# Before
thermal serve --port 3000 --open
curl http://127.0.0.1:3000/api/telemetry

# After
thermal live --json
thermal live --json --stream
```

**The fix**:
Replace `thermal serve`/`thermal web` with `thermal live --json` for a snapshot or `thermal live --json --stream` for continuous events. For historical reports, use the static commands with `--json`.

**Escape hatch**:
Pipe the NDJSON stream into your own local dashboard, log collector, or status bar. No network listener is required.

---

### Removed: Subscription Replay Simulation
<a id="removed-workload-replay"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal replay\|--against\|--compare\|internal/thermal/replay' .`.
- **Detection signal**: `thermal replay` exits 1 with a retirement notice. `--against` and `--compare` exit 1. `internal/thermal/replay.go` and `internal/render/replay.go` no longer exist, and `thermal replay --json` no longer emits a replay payload.

**What changed and why**:
The replay simulator hardcoded subscription ceilings and rate-card assumptions that change unpredictably, and its capacity verdicts could not be verified against provider entitlements. Maintaining that scenario engine inside a local CLI added drag and implied precision the data did not support. The static reports already expose the recorded volume and model pricing needed for a user's own scenario analysis.

**Before / After**:
```bash
# Before
thermal replay --against claude-pro
thermal replay --compare all --json

# After: exits 1 with migration guidance
thermal replay
# thermal: subscription replay simulation has been retired.
# For monthly volume and model pricing: thermal monthly, thermal models
```

**The fix**:
Use `thermal monthly` for period volume and `thermal models` for model-level token and cost breakdowns. Export either with `--json` for your own scenario model.

**Escape hatch**:
Combine `thermal daily --json` or `thermal monthly --json` with your own subscription limits and rate cards outside the CLI.

---

### Removed: Transcript Command Receipts
<a id="removed-work-receipts"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal receipt\|internal/thermal/receipt\|internal/render/receipt' .`.
- **Detection signal**: `thermal receipt` and tool-prefixed forms such as `thermal opencode receipt` exit 1 with a retirement notice. `internal/thermal/receipt.go` and `internal/render/receipt.go` no longer exist, and `thermal receipt --json` no longer emits a receipt payload. Codex token ingestion is unchanged.

**What changed and why**:
Scraping shell commands inside AI tool call blocks across many transcript formats was fragile, slow, and duplicated native test runners and CI gates without providing verifiable token accounting. The same session tokens remain available through the normal loaders and static reports. Removing the receipt engine also removes its tiered confidence claims, which could not be independently verified.

**Before / After**:
```bash
# Before
thermal receipt --sort verified
thermal opencode receipt --json

# After: exits 1 with migration guidance
thermal receipt
# thermal: transcript command receipt auditing has been retired.
# Verify test outcomes using native test runners: go test, pytest, cargo test
```

**The fix**:
Run tests and linters through their native tools or CI. Use `thermal daily`, `thermal projects`, or `thermal models` for token and cost accounting.

**Escape hatch**:
Export token and project data with `--json` and join it to CI results outside the CLI.

---

### Removed: Token Yield and Diff Parsing
<a id="removed-token-yield"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'thermal yield\|--sort yield\|--format md\|internal/thermal/yield' .`.
- **Detection signal**: `thermal yield` and tool-prefixed forms such as `thermal codewhale yield` exit 1 with a retirement notice. `internal/thermal/yield.go` and `internal/render/yield.go` no longer exist. `--format md` and `--format markdown` exit 1; `--format text` and `--format json` remain. Codex rollout caching no longer parses `FileChange` unified diffs; token caching and line metrics from databases that record them are unchanged.

**What changed and why**:
Net lines added is a misleading proxy for AI code efficiency: it penalizes refactoring, debugging, and deletions, and it coupled git-diff parsing to token telemetry. The yield badges implied a quality ranking the data could not support. Token burn and model cost remain available in the live monitor and static reports.

**Before / After**:
```bash
# Before
thermal yield --sort yield
thermal yield --json
thermal daily --format md

# After: exits 1 with migration guidance
thermal yield
# thermal: token yield efficiency reporting has been retired.
# For token burn rates and velocity: thermal live, thermal stats

thermal daily --json
```

**The fix**:
Use `thermal live` for real-time burn rates, `thermal stats` for distribution and token composition, and `thermal models` for model-level cost. Use `git diff --numstat` or your repository host for line-change data.

**Escape hatch**:
Join `thermal daily --json` or `thermal models --json` with `git diff --numstat` outside the CLI.

---

### Changed: Output Formatting Semantics
<a id="changed-output-formatting-semantics"></a>

**Affected if**:
- **Search pattern**: `grep -rn '~97.5M\|Window: \\[96 days\\]\|OpenCode,Devin\|P-Days\|M-Days\|incl. est.' .` or scripts parsing cost markers/table headers.
- **Detection signal**: `~` now means estimated money only; approximate non-cost values use `≈` (trend mean, audit tax). Report windows say `all time (N active days/weeks/months)` unless an explicit `--last` is used. Projects/models day columns are `P-Days`/`M-Days`. Tool lists use ` · ` instead of commas. Mixed report totals are printed without `~` and explained in the footer; all-estimated totals keep `~`. Share spend is `$X incl. est.` when estimates are included.

**What changed and why**:
One symbol and several local formatters were doing different jobs across surfaces. The result was a trust problem: `~$10` could mean estimated cost, a mixed total, or an approximate volume; `OpenCode,Devin` differed from the `·` convention; `Window: 96 days` looked like a calendar window when it was the number of active days. The formatting contract centralizes money/percent/window/tool-list rendering so a reader can compare surfaces without learning a different legend for each one.

**Before / After**:
```text
# Before
Window: [96 days · all tools]  Total: [38.0B tok · ~$20596.53]
Tools: OpenCode,Devin +8
Daily Mean: ~97.5M
Context Tax: ~45.4K tokens / turn

# After
Window: [all time (96 active days) · all tools]  Total: [38.0B tok · $20596.53]
Tools: OpenCode · Devin +8
Daily Mean: ≈97.5M
Context Tax: ≈45.4K tokens / turn
```

**The fix**:
Consume the shared formatting helpers rather than local `formatCost` copies. `$` is recorded money, `~$` includes estimated pricing, `—` is absent/unpriced, `≥$` is a partial-known live group, and `≈` is an approximate non-cost value. Use `--json` for machine values; do not parse the text cost marker as a numeric type.

**Escape hatch**:
For raw values, use `thermal <command> --json`. For a custom legend, map `$`/`~$`/`—`/`≥$` explicitly and use `≈` only for non-money approximations.

---

### Changed: Report Totals, Unclassified Tokens, and Cost Basis
<a id="changed-report-totals-and-cost-basis"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'Costed days\|Unclassified\|unattributedTokens\|estimatedCost\|costedDays' .` or consumers comparing leaderboard/daily/projects/models/mix/stats/share totals.
- **Detection signal**: `thermal models` now prints a `Classified + Unclassified = total volume` line and its JSON total exposes `unattributedTokens`; `thermal stats --metric cost` reports `Active days` and `Costed days` (with the excluded count); `thermal stats` includes `ClassifiedTotal`/`Unclassified` and labels the composition denominator. Reports expose `costBasis` (`total` for daily/projects/stats/tool-mix; `estimated` for models/model-mix). Leaderboard estimated cost, daily/project totals, and share spend use the same daily recorded+estimated basis. The share snapshot exposes numeric `recordedCost`, `estimatedCost`, `totalCost`, plus `scope`, `window`, and `coverage`; share active days are token-active days in that window.

**What changed and why**:
The same data previously produced 38.0B tokens on daily, 36.7B on models, 38.0B on projects, 20596.53 on daily and 20606.08 on models/mix, and 145 share active days vs 96 daily active days. Most of the token gap was unclassified volume; some cost gaps came from leaderboard/project estimators and from counting zero-cost days as active cost days. Report totals are now either equal to the canonical daily basis or explicitly labeled as a different basis (models/mix estimates). Unclassified volume is counted in composition percentages rather than silently folded into uncached input.

**Before / After**:
```text
# Before
models: 36.7B tok, no unclassified line
stats cost: Active days 80
stats composition: Cache read + Uncached input + ... (no unclassified row)
leaderboard estimated: ~$20381.54
daily estimated:        ~$20381.71
projects estimated:     ~$20384.60
share: Active Days 145, Total Volume 38.0B

# After
models: 36.7B classified + 1.3B unclassified = 38.0B total volume
stats cost: Active days 96, Costed days 80
stats composition: includes Unclassified; percentages include it
leaderboard estimated == daily estimated
projects total == daily recorded+estimated
share: Active Days 96, Total Volume 38.0B
```

**The fix**:
Use `thermal daily --json` (`totals.storedCost`, `estimatedCost`, `cost`, `unattributedTokens`) as the canonical cost/volume basis. Treat `thermal models` and model mixes as **estimate-only** because recorded cost is not attributable to a model. Treat `Costed days` as the denominator for cost mean/median/percentile/distribution. Use `unattributedTokens`/`Unclassified` for tokens without a model.

**Escape hatch**:
For a custom allocation, consume daily/project/model JSON and apply your own cost basis; do not compare `models.estimatedCost` to `daily.cost` as if they were the same basis.

---

### Changed: Chart Scale, Order, and `--chart` Applicability
<a id="changed-chart-scale-order-and-top"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'Projects by tokens\|Models by tokens\|--chart' .` or scripts comparing chart rows to `thermal projects --top`.
- **Detection signal**: `thermal mix --chart`, `thermal stats --chart`, and `thermal trend --chart` now fail with a message naming the supported commands. Project/model charts draw the same first `--top` rows as their table, scale against the largest **charted** row, preserve table order, and cap at twelve bars with a matching remainder count. Model chart costs use the shared `~$` estimate marker.

**What changed and why**:
Charts were described as a view of the table, but project/model charts scaled from `Rows[0]` and ignored `--top`; a sorted or ascending table could produce misleading bars, and analytics commands accepted `--chart` while doing nothing. A chart that silently disagrees with its table is worse than no chart.

**Before / After**:
```text
# Before
thermal projects --top 5 --chart   # table 5 rows, chart 12 bars, scale = first row
thermal mix --chart                # accepted, no chart

# After
thermal projects --top 5 --chart   # table 5 rows, chart 5 bars, scale = largest charted row
thermal mix --chart                # error: --chart does not apply to the mix command
```

**The fix**:
Use `--chart` only with `daily`, `weekly`, `monthly`, `projects`, or `models`. Pass `--top` through to project/model charts; the chart follows the table's order and scales against its largest displayed row. When the twelve-row cap omits bars, the remainder is printed.

**Escape hatch**:
`thermal <command> --json` always returns the full data regardless of `--chart`. Analytics commands remain table/text only.

---

### Changed: Audit Finding Scope and Score Components
<a id="changed-audit-finding-scope-and-score-components"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'Score components\|rulePenalty\|mcpPenalty\|readinessPenalty\|"scope"' .` or consumers of `thermal audit --json`.
- **Detection signal**: rule findings now carry `scope` (`repo root`, `~/.claude`, `repo .cursor/rules`, etc.) and the text table renders `AGENTS.md (repo root)` / `AGENTS.md (~/.agents)`. `thermal audit` prints a `Score components:` line naming rule/MCP/readiness penalties, the largest contributor, and the first mapped recommendation. Readiness rows use one status/check width and wrapped messages align under the message column. Long MCP names truncate with an ellipsis.

**What changed and why**:
Duplicate basenames (`CLAUDE.md`, `AGENTS.md`) were indistinguishable, so a recommendation could not be acted on. `PASS Tool Databases` and wrapped readiness messages broke the table's columns, and `58/100` did not explain its deductions. Scope and penalty components make the finding and the score addressable without changing what the audit checks.

**Before / After**:
```text
# Before
WARN  CLAUDE.md   204 lines  ≈6.8K tok
WARN  AGENTS.md   204 lines  ≈6.8K tok
PASS Tool Databases 3 local tool database...
Score: [58/100 · Grade D]

# After
WARN  CLAUDE.md (repo root)   204 lines  ≈6.8K tok
WARN  AGENTS.md (repo root)   204 lines  ≈6.8K tok
PASS  Tool Databases          3 local tool database...
Score: [58/100 · Grade D]
Score components: rules -30 · MCP -12 · readiness -0. Largest: rule files. Biggest win: ...
```

**The fix**:
Use `rules[].scope` and `rules[].path` in JSON, and the `(scope)` label in text. Use `rulePenalty`, `mcpPenalty`, and `readinessPenalty` to explain the score; recommendations carry the same scope as their finding.

**Escape hatch**:
Read `thermal audit --json` for full paths and the penalty fields; text intentionally abbreviates scopes for width.

---

### Changed: JSON Envelopes, Aliases, and Numeric Precision
<a id="changed-json-envelopes-aliases-precision"></a>

**Affected if**:
- **Search pattern**: `grep -rn '"Tool"\|"Name"\|"Summary"\|estimatedCost\|todayCost\|share\|mean' .` or consumers of any `thermal ... --json` payload.
- **Detection signal**: leaderboard JSON now has `type:"leaderboard"`, `generatedAt`, `results`, and `totals`; results no longer contain uppercase `Tool`, `Name`, or `Summary` aliases. Audit/changelog/license/live/share/version have explicit types. Every serialized float is rounded to four decimal places. `docs/JSON-CONTRACT.md` is the canonical key table.

**What changed and why**:
Four envelope shapes and duplicated leaderboard keys made JSON consumers branch by command and inflated the leaderboard payload. Raw floats such as `1.9774259520000008` were an artifact of binary arithmetic, not data. The contract now has one precision rule and a documented envelope table; aliases are removed rather than kept as a second full copy.

**Before / After**:
```text
# Before
{ "generatedAt": "...", "results": [{ "Tool": "opencode", "tool": "opencode",
  "Name": "OpenCode", "name": "OpenCode", "Summary": {...}, "summary": {...} }] }
"todayCost": 1.9774259520000008
"share": 0.5038565852808624

# After
{ "type": "leaderboard", "generatedAt": "...", "results": [{ "tool": "opencode",
  "name": "OpenCode", "summary": {...} }], "totals": {...} }
"todayCost": 1.9774
"share": 0.5039
```

**The fix**:
Use lower camelCase keys from `docs/JSON-CONTRACT.md`. Do not read uppercase aliases; they are removed. If you need raw precision for a calculation, compute from the source rather than parsing the rounded JSON. The four-decimal rule is serialization-only and does not change aggregation values.

**Escape hatch**:
Text output keeps its display precision; use JSON for machine values. Audit/live use `timestamp` by documented exception; all other envelopes use `generatedAt`.

---

### Changed: Live Ticker Units, Sparkline Glyphs, and Model Cost Basis
<a id="changed-live-ticker-units-and-model-cost-basis"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'step/min\|tok/min\|/min\|step…\|tok…\|(+60 s)\|Top model:' .` or consumers of live `todayModels`.
- **Detection signal**: live rate units are `tok/m` and `step/m`; speed units are `tok/s` and `step/s`; `Burned` uses `step`, not `s`. A compact `Steps:` line keeps both step rates visible when a narrow KPI cell drops the secondary value. Sparkline zero cells and the all-zero state both use `·`. Model rollups render `Top model (est. cost)` and JSON retains `costKind:"estimated"`.

**What changed and why**:
The old hybrid line omitted the token unit and could truncate `step/s` into `step…`; `s` meant both seconds and steps. The sparkline used spaces for leading zeros but dots for the all-zero state, which looked like missing data. Model cost was an estimate while Today/project cost was recorded, so the two dollar totals appeared to disagree without a basis label.

**Before / After**:
```text
# Before
Rate: 3.4M/m · 23 step/m
Speed: 56.7K · 0.383 step…
Burned: +1.7M tok (+60 s)
Spark: [   ▂▃ ...] / all-zero [····]
Top model: deepseek ~$3.95

# After
Rate: 3.4M tok/m · 23 step/m
Speed: 56.7K tok/s · 0.383 step/s
Burned: +1.7M tok · +60 step
Steps: 23 step/m · 0.383 step/s   # when the KPI cell is narrow
Spark: [···▂▃ ...] / all-zero [····]
Top model (est. cost): deepseek ~$3.95
```

**The fix**:
Use the canonical `/m` and `/s` units. Drop the secondary value before truncating a unit; the compact `Steps:` line preserves it. Treat model rollup cost as estimate-only (`costKind`) and never compare it directly to Today's reconciled cost.

**Escape hatch**:
`thermal live --json` exposes `todayCost`, `todayModels[].cost`, and `todayModels[].costKind` separately. Use the recorded/blended Today total for reconciliation and the model estimate only for attribution.

---

### Changed: Statusline and Share-Card Scope, and Share URL Display
<a id="changed-statusline-share-scope-url"></a>

**Affected if**:
- **Search pattern**: `grep -rn 'statusline\|Shareable Card URL\|Scope:.*All\|Active Days\|streakBasis' .` or consumers of `thermal share --json`.
- **Detection signal**: the statusline reads like `🔥 91d all tools · 220.0M tok today · top: CMD`; the share card prints `Window`, `Streak ... token-active`, and `Active Days ... token-active days`; `share --json` includes `scope`, `window`, `streakBasis`, `activeDays`, `totalTokens`, `recordedCost`, `estimatedCost`, `totalCost`, and `coverage`. The default text URL is shortened in the middle with a note; the full URL remains in JSON.

**What changed and why**:
`91d`, `220.0M tok`, and `CMD` were three different scopes in one line, while share `Active Days: 145` counted activity-only days but its volume/spend were token-only. A 1.5KB URL was presented as a compact card line. Scope labels and one snapshot now make the differences explainable, and the full URL remains machine-recoverable.

**Before / After**:
```text
# Before
🔥 91d · 220.0M tok · CMD
Scope: All · Active Days: 145 days · Total Volume: 38.0B tokens
Shareable Card URL: https://...<1500 chars>

# After
🔥 91d all tools · 220.0M tok today · top: CMD
Scope: All · Window: all time (96 active days)
Streak: 20 days (Longest: 37 days, token-active)
Active Days: 96 token-active days
Shareable Card URL: https://...<middle ellipsis>...
(full URL (1517 chars) is available with --json)
```

**The fix**:
Use the statusline's `scope`/`today`/`top` labels. Use share `scope`, `window`, and `streakBasis` together; do not mix all-activity streaks with token-active active days without the label. Copy the full URL from `thermal share --json`; the text card intentionally abbreviates it.

**Escape hatch**:
`thermal share --json` returns the full `url` and `token` plus the snapshot. Statusline JSON includes `scope`, `todayTokens`, and `activeTool`.
