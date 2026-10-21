# Facts: Sila Session Lifecycle Hooks

## Current Evidence (2026-10-21)
- `sila salam` reports `heal: 1 automatic fix pending` and `sila heal --dry-run` says `Provision sila wiring in this repository`, affecting `AGENTS.md`.
- `sila install --help` lists `claude`, `cursor`, and `devin` only; there is no `command-code` target.
- `sila agents check` passes on the managed block hash, but the full nine-verb protocol is only in `.cursor/rules/000-sila-principles.mdc`.
- `.githooks/` contains only `pre-commit` and `pre-push`; no session-start or session-end hook exists.
- `.commandcode/mcp.json` registers `sila`, but the MCP surface exposes only continuity tools such as `prime_context`, `recall`, `save_lesson`, `save_memo`, `save_handoff`, and `get_status`.
- `sila prime` surfaces an active handoff and `fact_c831b36c`, a prior bot-attribution lesson, proving the skipped preflight had concrete value.
- The current ledger is 114 goals, 518/518 tasks, but only 4 verified, 22 stale, and 88 unverified.

## Upstream Goal Ownership
- This goal is Thermal adoption only.
- Sila core owner: `harness-session-memory-bridge`, `active-lifecycle-hook-bridge`, and `project-onboarding` in `sila-cli`.
- Do not duplicate or re-register those upstream goals here.

## Architectural Invariants
- Every session has at most one successful start artifact and one successful end handoff.
- A hook failure is observable and actionable.
- Hooks never mutate user source files.

## File & Interface Contracts
- `.commandcode/config.json`: Session-start and session-end entries point to idempotent scripts.
- `scripts/sila-preflight.sh`: Exits 0 only when the required start checks pass; emits JSON on demand.
- `scripts/sila-session-end.sh`: Records one handoff and one sync, or a single explicit failure.

## Failure Modes
- Sila CLI absent from PATH.
- Sila store locked.
- MCP server starts but hangs.
- Session ends twice.
- Hook runs inside a subshell without PATH.

## Baseline Program Metrics
- 114 existing goals; 518/518 tasks; 4 verified, 22 stale, 88 unverified.
- `sila onboard check --strict`: 11 pass, 1 warn (`Agent rules` would update `AGENTS.md`).
- `sila skills list`: no project skills under `.agents/skills/`.
- `sila persona list --json` exposes persona groups; `sila audit` exposes `web`, `ui`, `squad`, and `gap` modes.
- `sila feedback` contains `fb_heal-contract-protocol`, an open issue describing this program's original gap.

## Constraints
- Do not change Thermal product behavior from this goal unless the goal explicitly owns that surface.
- Do not introduce silent fallbacks that make a skipped Sila step look successful.
- Do not use `--force` claims or bot co-author trailers to get past lifecycle checks.
- Keep all verification reproducible from a fixed fixture or a documented live command.
- Do not push remote changes without explicit user consent.

## Open Questions
- Which Command Code hook event is closest to session start?
- Should heal run every session or only when stale?
- Should handoff be automatic or require a summary argument?
