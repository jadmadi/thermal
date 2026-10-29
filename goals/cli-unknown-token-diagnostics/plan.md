# Plan: Classify unknown commands, tools, and extra positionals with actionable suggestions

## Execution Steps
- [x] **Phase 1 — Vocabulary and classifier**: Export the command/tool/report vocabularies from one source, add `ToolVocabulary`, and implement `classifyToken` + `damerauLevenshtein` + `suggest`. Acceptance: `TestVocabularyParity`, `TestClassifierOrder`, and `TestSuggestionDistance` pass.
  - Command pool: registry/`isReportWord` + aliases.
  - Tool pool: `AllTools()` canonical names + aliases.
  - Report pool: report words.
  - Retired pool: dashboard, serve, web, replay, yield, receipt.
- [x] **Phase 2 — Message templates and wiring**: Implement `renderDiagnostic` with the five-line budget and replace the three unknown-tool sites. Add the synonym map and slot isolation. Acceptance: `TestUnknownCommandCorpus`, `TestUnknownToolCorpus`, `TestWrongSlotCorpus`, and `TestDiagnosticLineBudget` pass.
  - `today`/`yesterday` → daily; `week` → weekly; `month` → monthly; `repo`/`project` → projects; `model` → models; `heatmap`/`streak` → leaderboard; `cost`/`spend` → stats --metric cost.
  - Path detection before tool resolution.
  - Retired verbs keep their message and exit `1`.
- [x] **Phase 3 — False positives and exit codes**: Add the valid-token corpus, the small-token rule (no fuzzy suggestion under 3 characters), and the exit-code mapping (`2` usage, `1` retired/data). Acceptance: `TestSuggestionFalsePositives` and `TestExitCodeMatrix` pass; `opencode`, `--tool opencode`, `all`, `auto`, `help`, `version`, `license` all behave unchanged.
- [x] **Phase 4 — Gate, docs, and closure**: Add the simulated-user-gate diagnostics section, README troubleshooting, MIGRATION entry, project.jsonc, and close the goal.
  - Record recall/precision numbers and every AC command in `evidence.md`.
  - `CGO_ENABLED=0 go test -race ./...` and `./scripts/simulated_user_gate.sh` pass.
  - Commit with `feat(cli): add slot-aware unknown-token diagnostics (goals/cli-unknown-token-diagnostics/goal.md)`.
  - Certify with `sila goals certify cli-unknown-token-diagnostics --evidence=goals/cli-unknown-token-diagnostics/evidence.md`.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | `thermal today` | exit 2, `unknown command`, `Did you mean: thermal daily?` |
| AC-002 | `thermal daly`, `thermal weak` | suggestions `daily`, `weekly` |
| AC-003 | `thermal stats weekly` | wrong-slot message, `thermal weekly` |
| AC-004 | `thermal projects /tmp` | path message, `--db` |
| AC-005 | `thermal --tool nope` | unknown tool, tool list |
| AC-006 | valid tokens | exit 0, no suggestion |
| AC-007 | `thermal dashboard` | retirement message, exit 1 |
| AC-008 | `thermal a` | no confident suggestion |
| AC-009 | line budget | ≤5 lines, stdout empty |
| AC-010 | `NO_COLOR=1` | zero ANSI |
| AC-011 | `go test -race ./cmd/thermal/... ./internal/loaders/...` | exit 0 |
| AC-012 | `./scripts/simulated_user_gate.sh` | exit 0 |

## Rollback
- Keep the existing retirement path untouched; it is already correct.
- If suggestions cause false positives, disable fuzzy matching (keep the vocabulary list and the wrong-slot/path classification) rather than reverting the whole goal.
- Never restore the unconditional `unknown tool` message for command-slot tokens.

## Phase 5 — Attribution & Registry Closure
- [x] **Phase 5 — Attribution & Registry Closure**: (1) check all completed tasks in `goals/cli-unknown-token-diagnostics/plan.md`; (2) commit with `(goals/cli-unknown-token-diagnostics/goal.md)` using a `feat(cli):` subject; (3) certify via `sila goals certify cli-unknown-token-diagnostics --evidence=goals/cli-unknown-token-diagnostics/evidence.md` (with `Verdict: PASS`); (4) reconcile via `sila goals` and commit the updated `goals/README.md` and certification receipts; verify `sila goals --json` reports `unverified_count: 0` for this goal.
