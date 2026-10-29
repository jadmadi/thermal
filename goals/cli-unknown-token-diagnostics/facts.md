# Facts: Classify unknown commands, tools, and extra positionals with actionable suggestions

## Observed Evidence (2026-10-29, current source build `thermal dev`)
| Command | stderr | Exit |
| --- | --- | --- |
| `thermal today` | `thermal: unknown tool: today` | 1 |
| `thermal daly` | `thermal: unknown tool: daly` | 1 |
| `thermal weak` | `thermal: unknown tool: weak` | 1 |
| `thermal streak` | `thermal: unknown tool: streak` | 1 |
| `thermal heatmap` | `thermal: unknown tool: heatmap` | 1 |
| `thermal tools` | `thermal: unknown tool: tools` | 1 |
| `thermal cost` | `thermal: unknown tool: cost` | 1 |
| `thermal stats weekly` | `thermal: unknown tool: weekly` | 1 |
| `thermal projects /tmp` | `thermal: unknown tool: /tmp` | 1 |
| `thermal projects extra --top 3` | `thermal: unknown tool: extra` | 1 |
| `thermal daily --metirc cost` | `thermal: unknown tool: cost` | 1 |
| `thermal --tool nope` | `thermal: unknown tool: nope` | 1 |
| `thermal dashboard` | `thermal: the interactive dashboard has been retired in favor of 'thermal live' and static report commands.` + two lines | 1 |

No invocation produced a suggestion, a valid-command list, or a `thermal --help` pointer.

## Root Cause
- `cmd/thermal/main.go:299-311`: any first positional that is not a report word becomes `opts.Tool`.
- `cmd/thermal/main.go:774-780`, `:923-927`, `:1548-1552`: three separate call sites print `thermal: unknown tool: %s` and exit `1`.
- `internal/loaders/registry.go:345`: `ResolveTool(name string) (thermal.Tool, bool)` is an exact-match map lookup; it has no alias-aware or fuzzy API.
- `cmd/thermal/main.go:330-336`: `isReportWord` is a literal list (`daily`, `weekly`, `monthly`, `projects`, `models`, `trend`, `mix`, `stats`, `replay`, `yield`, `receipt`, `web`, `serve`, `changelog`, `live`, `statusline`, `prompt`) that is not shared with the error layer.
- There is no distinction between a command slot, a tool slot, and an extra argument; every failure collapses into "unknown tool".

## Architectural Invariants & Constraints
- Error messages must name the thing the user actually typed and the slot it occupies.
- A suggestion must be a real, current command/tool; retired names must never be suggested.
- The suggestion engine must be deterministic and dependency-free (no fuzzy library).
- Unknown-token diagnostics are usage errors: exit `2`, stdout empty, stderr only.
- The tool vocabulary must come from `loaders.AllTools()` and its alias map, not a second literal list.
- Zero telemetry: do not record typo tokens anywhere.
- Valid invocations must be untouched; suggestion logic runs only after exact resolution fails.

## File & Interface Contracts
- `cmd/thermal/diagnostics.go` (new):
  ```go
  type tokenSlot int // slotCommand, slotTool, slotReport, slotOther
  type tokenClass int // classKnown, classUnknown, classWrongSlot, classPath, classRetired, classEmpty
  type diagnostic struct {
      Token, Classification, Suggestion string
      Alternatives []string
      Slot tokenSlot
      ExitCode int
  }
  func classifyToken(token string, slot tokenSlot) diagnostic
  func renderDiagnostic(d diagnostic, noColor bool) string
  func suggest(token string, candidates []string) (best string, alternatives []string, ok bool)
  func damerauLevenshtein(a, b string) int
  ```
- `internal/loaders/registry.go`: add `func ToolVocabulary() (canonical []string, aliases map[string]thermal.Tool)` or equivalent, derived from `AllTools`; `ResolveTool` stays exact-match for loading.
- `cmd/thermal/main.go`: replace the three unknown-tool sites with `classifyToken` + `renderDiagnostic`; keep the retirement path.
- `cmd/thermal/diagnostics_test.go`: corpora and property tests.

## Classification Order
1. Empty string → `classEmpty`.
2. Retired verb → `classRetired` (exit `1`, existing message).
3. Exact command/alias in the command slot → `classKnown`.
4. Exact command/alias in a non-command slot → `classWrongSlot`.
5. Path-like token (`/`, `\`, leading `.`, `~`) → `classPath`.
6. Exact tool/alias in the tool slot → `classKnown`.
7. Fuzzy match above threshold → `classUnknown` + suggestion.
8. Otherwise → `classUnknown` + vocabulary list.

## Suggestion Rules
- Distance: Damerau-Levenshtein, case-insensitive, trimmed.
- Threshold: `min(2, max(1, len(token)/3))`.
- Tie-break: smallest distance, then shortest candidate, then alphabetical.
- Synonyms (intent-level, checked before distance):
  `today|yesterday → daily`, `week → weekly`, `month → monthly`,
  `repo|project → projects`, `model → models`, `heatmap|streak → thermal`,
  `chart → daily --chart`, `cost|spend → stats --metric cost`,
  `status → statusline`.
- Candidate pools: command names + aliases; tool canonical names + aliases; report words.
- Never cross slots: a tool-slot failure suggests only tools, a command-slot failure suggests only commands/reports.
- Never suggest a retired verb.

## Message Templates
```
thermal: unknown command "today"
  Did you mean: thermal daily?
  Valid commands: daily, weekly, monthly, projects, models, stats, mix, trend,
  live, statusline, audit, changelog, share, upgrade, version, license
  run thermal --help for the command list

thermal: "weekly" is a report command
  Did you mean: thermal weekly?

thermal: "/tmp" looks like a path
  thermal commands do not take paths; use --db <path> for a data source

thermal: unknown tool "nope"
  Did you mean: "opencode"?
  Valid tools: OpenCode (oc), Codex, Claude, Devin, MiMoCode (mimo), CMD (command-code), ...
  run thermal --help for supported tools
```
- Maximum five lines, 80 columns, two-space continuation indent.
- stdout empty, stderr only.
- Color only when stderr is a TTY and `NO_COLOR` is unset.

## Verification Contract
```bash
# Command-slot classification
thermal today >/tmp/o 2>/tmp/e; test $? -eq 2; test ! -s /tmp/o
grep -q 'unknown command "today"' /tmp/e
grep -q 'Did you mean: thermal daily?' /tmp/e
! grep -q 'unknown tool' /tmp/e

# Wrong slot
thermal stats weekly >/dev/null 2>/tmp/e; test $? -eq 2
grep -q '"weekly" is a report command' /tmp/e
grep -q 'thermal weekly' /tmp/e

# Path
thermal projects /tmp >/dev/null 2>/tmp/e; test $? -eq 2
grep -q 'looks like a path' /tmp/e
grep -q -- '--db' /tmp/e

# Tool slot
thermal --tool nope >/dev/null 2>/tmp/e; test $? -eq 2
grep -q 'unknown tool "nope"' /tmp/e
grep -q 'Valid tools' /tmp/e

# No false positives
thermal opencode >/dev/null 2>&1; test $? -eq 0
thermal --tool opencode >/dev/null 2>&1; test $? -eq 0

# Retirement unchanged
thermal dashboard >/dev/null 2>/tmp/e; test $? -eq 1
grep -q 'thermal live' /tmp/e

CGO_ENABLED=0 go test -race ./cmd/thermal/... ./internal/loaders/...
./scripts/simulated_user_gate.sh
```

## Regression Fixtures
- Command typo corpus (≥40) with expected targets.
- Tool typo corpus (≥20) including every alias.
- Wrong-slot corpus: `stats weekly`, `projects /tmp`, `models mix`, `daily stats`, `--tool daily`.
- False-positive corpus: every valid command and tool, plus `all`, `auto`, `help`, `version`, `license`, `a`, `x`.
- A line-budget snapshot at 80 columns.
- An ANSI-free snapshot under `NO_COLOR=1`.

## Open Questions
- Should unknown commands exit `2` or `1`? Recommendation: `2` (usage), documented in MIGRATION. Retired verbs keep `1`.
- Should `thermal a` suggest `agy`? Recommendation: no; the threshold for a one-character token is 1, but the false-positive cost is high. Require at least 3 characters for fuzzy suggestions.
- Should synonyms be case-sensitive? Recommendation: case-insensitive.
- Should the command list include retired verbs? Recommendation: no; list only current verbs and point to `thermal changelog` for history.
