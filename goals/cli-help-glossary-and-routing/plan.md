# Plan: Route every unknown word to the right namespace and explain domain vocabulary

## Execution Steps
- [x] **Phase 1 — Glossary table and renderer**: Add `GlossaryEntry`,
  `lookupGlossary`, and `renderGlossary`; wire `help <term>` before the
  unknown-command path. Acceptance: `TestGlossary` and
  `TestGlossaryExitCodes` pass.
  - At least 15 terms, each with a definition and see-also.
  - `help glossary` lists them; `help <unknown>` exits 2 with pointers.
- [x] **Phase 2 — Namespace classifier and suggestion quality**: Extend
  `tokenDiagnostic` with a namespace, add prefix/canonical matching, and
  render canonical-plus-alias suggestions. Acceptance:
  `TestTokenClassification`, `TestSuggestionQuality`, and
  `TestNoAliasOnlySuggestion` pass.
  - `doctor`/`config` route to commands, not tools.
  - `top` suggests `--top`, not `oc`.
  - `open` -> OpenCode; `code` -> Claude/OpenCode/CMD.
- [x] **Phase 3 — Leaderboard hint**: Add one `Next:` line after the cheer,
  width-aware, JSON-unchanged. Acceptance: `TestLeaderboardNextHint` at 80 and
  120, `TestLeaderboardJSONUnchanged`.
- [x] **Phase 4 — Help grouping and docs**: Add the glossary pointer to global
  help, update README, MIGRATION, `project.jsonc`, and the gate. Acceptance:
  help line count <= 105 at 80 columns; gate exits 0.
- [x] **Phase 5 — Evidence and closure**: Record the glossary, classification,
  suggestion, and leaderboard outputs; commit, certify, reconcile.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | `help streak` | definition + see-also, exit 0 |
| AC-002 | `help token/warriors/activity/glossary` | exit 0 |
| AC-003 | `help glossary` | >=15 terms |
| AC-004 | `help nosuchterm` | exit 2, glossary + help |
| AC-005 | `doctor` | sources pointer, no tool list |
| AC-006 | `top` | `--top`, no `oc` |
| AC-007 | `--tool open` | OpenCode + alias |
| AC-008 | `--tool code` | Claude, OpenCode, CMD |
| AC-009 | suggestion table | no alias-only |
| AC-010 | leaderboard | one `Next:` line; JSON unchanged |
| AC-011 | global help | glossary pointer |
| AC-012 | `go test -race ...` | exit 0 |
| AC-013 | simulated user gate | exit 0 |

## Rollback
- Keep the glossary as an additive `help` path; if a term collides with a
  command, the command still wins on execution.
- If canonical suggestions are too long, print the canonical name only and
  put the alias in the valid-tools list.
- If the leaderboard hint is noisy, gate it behind `--verbose`; do not remove
  the cheer.
- Do not return to alias-only suggestions; that was the defect.

## Phase 6 — Attribution & Registry Closure
- [x] **Phase 6 — Attribution & Registry Closure**: (1) check every task
  above; (2) commit with
  `feat(cli): add a glossary and namespace-aware help routing
  (goals/cli-help-glossary-and-routing/goal.md)`; (3) certify via
  `sila goals certify cli-help-glossary-and-routing
  --evidence=goals/cli-help-glossary-and-routing/evidence.md`; (4) run
  `sila goals` and commit the updated `goals/README.md` and certification
  receipts; verify `sila goals --json` reports `unverified_count: 0`.
