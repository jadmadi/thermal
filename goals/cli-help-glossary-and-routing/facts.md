# Facts: Route every unknown word to the right namespace and explain domain vocabulary

## Observed Evidence (2026-10-30, `/tmp/thermal-cli-review`, `thermal dev`)

### Jargon is un-lookupable
```
$ thermal help streak
exit 1; stdout empty
thermal: unknown command "streak"
  run thermal --help for the command list

$ thermal help token
exit 1; stdout empty
thermal: unknown command "token"
  run thermal --help for the command list
```

### Beginner intent is blamed on the tool namespace
```
$ thermal doctor
exit 2; stdout empty
thermal: unknown tool "doctor"
  Valid tools: OpenCode, Devin, Codex, Claude, ZCode, Grok, MiMoCode,
  codewhale, DeepSeek (DSH), Nous Hermes, CMD, Zed, Agy, Muse, Droid
  run thermal --help for supported tools

$ thermal config
exit 2; the same shape.
```

### Suggestions can be worse than none
```
$ thermal top
exit 2; stdout empty
thermal: unknown tool "top"
  did you mean: "oc"?
  Valid tools: ... (15 names, none of them "oc" as a word)

$ thermal --tool open
exit 2; stdout empty
thermal: unknown tool "open"
  Valid tools: ... 15 names ...        # no suggestion

$ thermal --tool code
exit 2; stdout empty
thermal: unknown tool "code"
  did you mean: "ccode"?               # "ccode" absent from the list
  Valid tools: ... 15 names ...
```

### The populated leaderboard has no next step
```
$ thermal
...
  >> codewhale is on fire with a 2-day streak!
  Cost: $7.00 recorded + ~$0.02 estimated = ~$7.02 shown

  Keep the heat going. Don't break the streak.
```
No hint for `--sort`, `--last`, `--chart`, `sources`, or `--help`.

## Root Cause
- `cmd/thermal/help.go:564-573` (`helpOutput`) prints
  `unknown command %q` and the command-list pointer for any unresolved help
  target. It has no glossary lookup and no namespace fallback.
- `cmd/thermal/help.go:442-455` (`helpSynonyms`) covers `today`, `week`,
  `month`, `repo`, `project`, `model`, `cost`, `spend`, `status`, `prompt`,
  but not `streak`, `token`, `warriors`, `activity`, `doctor`, `config`, or
  `setup`.
- `cmd/thermal/diagnostics.go:161-183` (`classifyToken`) tries `helpSynonyms`,
  then `nearestName` over command names, then tool names only when the slot is
  `slotTool`. An unknown command-slot token with no close match ends as
  `classUnknown` with the tool label (the default) and the 15-tool list.
- `cmd/thermal/args.go:312-330` (`nearestName`) rejects keys shorter than 3
  and accepts any candidate within edit distance 2. It does not know canonical
  names versus aliases.
- `cmd/thermal/diagnostics.go:79-89` (`diagnosticToolNames`) includes aliases
  but the rendered valid list uses `toolHelpEntries` display names, so an
  alias suggestion can be absent from the list.
- `cmd/thermal/diagnostics.go:219-230` renders the suggestion without a
  canonical name for aliases.
- `internal/render/leaderboard.go` ends with the cheer and cost footer; there
  is no `Next:` line.

## Architectural Invariants & Constraints
- Help is an interception layer: `help` wins over validation and execution,
  and never runs a command.
- A token has exactly one namespace. A word that is both a glossary term and a
  command resolves to the command when executed and to the glossary when
  passed to `help`.
- Suggestions must be actionable: a command, a flag, or a canonical tool name
  with its alias.
- Never print the 15-tool list for a word that is not close to a tool.
- Exit codes: glossary `0`, unknown `2`, retired `1`; do not invent `3`.
- No colour-only encoding: the namespace word is always present.
- The glossary is one table, used by help, tests, and the README.
- The leaderboard hint is additive; `--json` is unchanged.

## File & Interface Contracts
- `cmd/thermal/help.go`:
  ```go
  type GlossaryEntry struct {
      Term       string
      Definition string
      SeeAlso    []string
  }

  var glossary = []GlossaryEntry{...}
  func lookupGlossary(term string) (GlossaryEntry, bool)
  func renderGlossary(term string) string
  ```
- `cmd/thermal/diagnostics.go`:
  ```go
  type tokenNamespace int
  const (
      nsCommand tokenNamespace = iota
      nsReport
      nsTool
      nsFlag
      nsGlossary
      nsPath
      nsRetired
      nsUnknown
  )

  type tokenDiagnostic struct {
      Token      string
      Namespace  tokenNamespace
      Canonical  string
      Alias      string
      Suggestion string
      Next       []string
      ExitCode   int
  }
  ```
- `cmd/thermal/args.go`:
  ```go
  func nearestName(key string, names []string) (string, bool) // minLen 3, maxDist 2
  func nearestCanonical(key string, entries []ToolHelpEntry) (canonical, alias string, ok bool)
  ```
- `cmd/thermal/main.go`: help dispatch calls `lookupGlossary` before the
  unknown-command path.
- `internal/render/leaderboard.go`: add `NextHint string` to the render
  options and print one line after the cheer.

## Namespace Matrix
| Token | Namespace | Exit | Next |
| --- | --- | --- | --- |
| `daily` | command | 0 | execute |
| `streak` | glossary | 0 | definition + see also |
| `opencode`, `oc` | tool | 0 | execute |
| `--json` | flag | 0 | execute |
| `--jsno` | flag (typo) | 2 | `did you mean --json?` |
| `dashboard` | retired | 1 | `thermal live` |
| `./db.sqlite` | path | 2 | `--db` hint |
| `doctor` | unknown command | 2 | `thermal sources`, `--help` |
| `top` | unknown command | 2 | `--top <N>` |
| `open` (tool slot) | tool (prefix) | 2 | `OpenCode (alias "oc")` |
| `code` (tool slot) | tool (ambiguous) | 2 | Claude, OpenCode, CMD |
| `zzzz` | unknown | 2 | `--help` |

## Verification Contract
```bash
B=/tmp/thermal-cli-review
H=<fixture-home>

# Glossary
$B help streak >/tmp/o 2>/tmp/e; test $? -eq 0
grep -q 'streak' /tmp/o; grep -q 'See also' /tmp/o
for t in token warriors activity glossary; do
  $B help "$t" >/tmp/o 2>/tmp/e || exit 1
  test -s /tmp/o
done
$B help glossary | grep -c '^  [a-z]' | awk '$1>=15'

# Unknown help
$B help nosuchterm >/tmp/o 2>/tmp/e; test $? -eq 2
grep -q 'help glossary' /tmp/e

# Intent words
$B doctor >/tmp/o 2>/tmp/e; test $? -eq 2
grep -q 'thermal sources' /tmp/e
! grep -q 'Valid tools:' /tmp/e
$B top >/tmp/o 2>/tmp/e; test $? -eq 2
grep -q -- '--top' /tmp/e
! grep -q '"oc"' /tmp/e

# Tool suggestions
$B --tool open >/tmp/o 2>/tmp/e; test $? -eq 2
grep -q 'OpenCode' /tmp/e; grep -q 'oc' /tmp/e
$B --tool code >/tmp/o 2>/tmp/e; test $? -eq 2
grep -q 'Claude' /tmp/e; grep -q 'OpenCode' /tmp/e; grep -q 'CMD' /tmp/e

# Leaderboard hint
$B >/tmp/o 2>/dev/null
grep -q '^  Next:' /tmp/o
$B --json >/tmp/j 2>/dev/null
! grep -q 'Next:' /tmp/j

CGO_ENABLED=0 go test -race ./cmd/thermal/... ./internal/render/...
./scripts/simulated_user_gate.sh
```

## Regression Fixtures
- A glossary table with at least 15 terms and see-also links.
- One token per namespace and per exit code.
- Tool prefix cases: `open`, `code`, `codex`, `dsh`, `ccode`, `oc`.
- A `help` term that collides with a command (`cost`, `model`).
- A leaderboard with data at 80 and 120 columns.
- An empty HOME to prove help does not require data.

## Open Questions
- Should `thermal help streak` exit `0` or `1`? Recommendation: `0`; the term
  exists and the command succeeded.
- Should `dashboard` become a glossary term or stay retired? Recommendation:
  stay retired with `thermal live`; a glossary entry would imply it is a
  concept, not a removed command.
- Should suggestions ever include an alias alone? Recommendation: no; always
  print `Canonical (alias "x")`.
- Should the leaderboard hint be configurable? Recommendation: no; one line,
  no setting.
- Should `help` list glossary terms in the global page? Recommendation: one
  pointer line, not the full table, to keep help under 105 lines.
