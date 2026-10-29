# Plan: Replace the empty-state dead end with a per-source inventory and next steps

## Execution Steps
- [ ] **Phase 1 — Inventory model**: Add `internal/loaders/inventory.go` with `SourceInfo`, `SourceState`, `Inventory`, and `InventorySummary`, reusing the registry paths and `ToolPathShape`. Acceptance: `TestInventoryStableOrder`, `TestInventoryStates`, and `TestInventoryEmptyHome` pass.
  - All 15 tools, stable order, canonical name + alias.
  - `stat`/`ReadDir` only; no database open.
  - `HERMES_HOME`/`DSH_HOME` and platform paths respected.
  - Timeout/slow-path handling.
- [ ] **Phase 2 — Renderers and command**: Add `internal/render/sources.go`, the `sources` verb, text table at 80/120, JSON schema, and filter by tool. Acceptance: `TestSourcesText`, `TestSourcesJSON`, and `TestSourcesNarrowWidth` pass.
  - Human table to stdout for `sources`, stderr for the empty-state path.
  - Legend and summary.
  - No ANSI under `NO_COLOR`.
- [ ] **Phase 3 — Empty-state integration and JSON**: Replace the three generic messages with `renderEmptyState`, add the command-specific hints, the `thermal --json` empty payload, and the one-tool empty wording. Acceptance: `TestEmptyStateBare`, `TestEmptyStateJSON`, and `TestEmptyStatePerCommand` pass.
  - Bare `thermal`: stderr, exit 1.
  - `thermal --json`: `type:"sources"` on stdout, exit 1.
  - `thermal sources`/`--json`: exit 0.
- [ ] **Phase 4 — Privacy, performance, gate, docs, and closure**: Add the sentinel privacy test, the <100 ms benchmark, the simulated-user-gate section, README/MIGRATION/project.jsonc, and close the goal.
  - Record the AC matrix, fixture outputs, and latency in `evidence.md`.
  - `CGO_ENABLED=0 go test -race ./...` and `./scripts/simulated_user_gate.sh` pass.
  - Commit with `feat(cli): add a per-source inventory and actionable empty state (goals/cli-first-run-source-inventory/goal.md)`.
  - Certify with `sila goals certify cli-first-run-source-inventory --evidence=goals/cli-first-run-source-inventory/evidence.md`.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | empty HOME `thermal` | exit 1, 15-row table on stderr |
| AC-002 | empty HOME `thermal --json` | exit 1, sources JSON, 15 entries |
| AC-003 | empty HOME `thermal sources` | exit 0, stdout table |
| AC-004 | `sources --json | jq` | valid, 15 entries |
| AC-005 | one-tool fixture | `found==1` |
| AC-006 | empty Claude dir | `empty` |
| AC-007 | permission denied | `unreadable` (skip as root) |
| AC-008 | `HERMES_HOME` fixture | resolved path shown |
| AC-009 | sentinel prompt | absent from output |
| AC-010 | latency | <100 ms, no DB open |
| AC-011 | 80-column check | no long lines |
| AC-012 | `go test -race ...` | exit 0 |
| AC-013 | `./scripts/simulated_user_gate.sh` | exit 0 |

## Rollback
- Keep `sources` as an additive command; if the empty-state table is too noisy, gate it behind `--verbose` and keep the two next-step lines.
- Do not revert the `thermal --json` empty payload without a MIGRATION note; scripts may rely on the new shape.
- Never read file contents to improve the inventory; that would violate the privacy and performance invariants.

## Phase 5 — Attribution & Registry Closure
- [ ] **Phase 5 — Attribution & Registry Closure**: (1) check all completed tasks in `goals/cli-first-run-source-inventory/plan.md`; (2) commit with `(goals/cli-first-run-source-inventory/goal.md)` using a `feat(cli):` subject; (3) certify via `sila goals certify cli-first-run-source-inventory --evidence=goals/cli-first-run-source-inventory/evidence.md` (with `Verdict: PASS`); (4) reconcile via `sila goals` and commit the updated `goals/README.md` and certification receipts; verify `sila goals --json` reports `unverified_count: 0` for this goal.
