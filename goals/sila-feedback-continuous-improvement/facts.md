# Facts: Sila Feedback Continuous Improvement

## Current Evidence (2026-10-21)
- `sila salam` reports `heal: 1 automatic fix pending` and `sila heal --dry-run` says `Provision sila wiring in this repository`, affecting `AGENTS.md`.
- `sila install --help` lists `claude`, `cursor`, and `devin` only; there is no `command-code` target.
- `sila agents check` passes on the managed block hash, but the full nine-verb protocol is only in `.cursor/rules/000-sila-principles.mdc`.
- `.githooks/` contains only `pre-commit` and `pre-push`; no session-start or session-end hook exists.
- `.commandcode/mcp.json` registers `sila`, but the MCP surface exposes only continuity tools such as `prime_context`, `recall`, `save_lesson`, `save_memo`, `save_handoff`, and `get_status`.
- `sila prime` surfaces an active handoff and `fact_c831b36c`, a prior bot-attribution lesson, proving the skipped preflight had concrete value.
- The current ledger is 114 goals, 518/518 tasks, but only 4 verified, 22 stale, and 88 unverified.

## Architectural Invariants
- Every feedback item has an owner and status.
- Upstream submission is explicit.
- Resolution links to a goal or a recorded decision.

## File & Interface Contracts
- `feedback record`: ID, type, severity, surface, linked goal, status.
- `resolution`: Evidence, decision, and timestamp.

## Failure Modes
- Duplicate feedback.
- Accidental public share.
- Feedback closed without a fix.
- Sensitive content leaked.

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
- Should `fb_heal-contract-protocol` be filed upstream?
- Who owns anonymous sharing?
