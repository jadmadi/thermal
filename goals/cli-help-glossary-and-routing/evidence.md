# Evidence: Route every unknown word to the right namespace

Date: 2026-10-30
Binary: `/tmp/thermal-g6` (source build)
Goal: `cli-help-glossary-and-routing`

## Acceptance Matrix

| AC | Command | Result |
| --- | --- | --- |
| AC-001 | `thermal help streak` | exit 0; definition and See also |
| AC-002 | `help token`, `help warriors`, `help activity`, `help glossary` | exit 0 |
| AC-003 | `help glossary` | 19 terms listed |
| AC-004 | `help nosuchterm` | exit 2; nearest glossary/command suggestion |
| AC-005 | `thermal doctor` | exit 2; `thermal sources` pointer, no vendor list |
| AC-006 | `thermal top` | exit 2; `--top <N>`, no `oc` |
| AC-007 | `thermal --tool open` | exit 2; `OpenCode (alias "oc")` |
| AC-008 | `thermal --tool code` | Claude, OpenCode, and CMD suggestions |
| AC-009 | suggestion table | no alias-only suggestion |
| AC-010 | populated leaderboard | one `Next:` hint line; `--json` unchanged |
| AC-011 | `thermal --help` | includes `thermal help glossary` |
| AC-012 | `go test ./cmd/thermal/... ./internal/render/...` | ok |
| AC-013 | `simulated_user_gate.sh` | 171 passed, 0 failed |

## Observed Output

```text
$ thermal help streak
  thermal glossary · streak
  Consecutive calendar days with activity for a tool. The default
  leaderboard ranks by current streak; Best is the longest streak.

  See also: thermal daily · thermal stats
exit 0
```

```text
$ thermal doctor
exit 2
thermal: unknown command "doctor"
  checking what thermal can read:  thermal sources
  all commands:                    thermal --help
```

```text
$ thermal top
exit 2
thermal: unknown command "top"
  did you mean: --top <N>   (e.g. thermal models --top 5)
  run thermal --help for the flag list
```

```text
$ thermal --tool open
exit 2
thermal: unknown tool "open"
  did you mean: OpenCode (alias "oc")?
  ...
```

```text
$ thermal --tool code
exit 2
thermal: unknown tool "code"
  did you mean: Claude (alias "ccode"), OpenCode (alias "oc"), or CMD?
  ...
```

```text
$ thermal --no-color | grep '^  Next:'
  Next: thermal weekly --chart   weekly bars · thermal --help   all flags
```

## Implementation Notes

- Added a 19-term glossary and `glossaryPage`/`nearestGlossary` in
  `cmd/thermal/help.go`; `help <term>` exits 0 for exact terms and 2 for
  unknown terms with a nearest suggestion.
- Added intent-word routing (`doctor`, `config`, `setup`, `history`, `usage`)
  to `thermal sources`; `top` routes to `--top <N>`.
- Added canonical tool suggestions for prefixes (`open` -> OpenCode) and
  ambiguous names (`code` -> Claude, OpenCode, CMD).
- Added one leaderboard `Next:` hint; `--json` is unchanged.
- Global help lists `thermal help glossary` in Common Examples.

## Verification Commands

```bash
go test ./cmd/thermal/... ./internal/render/...
./scripts/check.sh --pre-commit
./scripts/simulated_user_gate.sh
```

All passed. The gate reported `171 passed, 0 failed`.

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "cli-help-glossary-and-routing",
  "prime": {"ran": true},
  "recall": [
    {"query": "help streak unknown command", "matches": 5},
    {"query": "did you mean oc", "matches": 5}
  ],
  "memos": [{"id": "fact_d5a23a3f"}],
  "lessons": [{"id": "fact_c4ef59da"}],
  "cases": [{"id": "fact_ba7acb6d"}],
  "facts_verify": {"exit_code": 0},
  "handoff": {"id": "hoff_c551c09046d6"}
}
```

Verdict: PASS
