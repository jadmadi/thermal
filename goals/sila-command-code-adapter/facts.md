# Facts: Sila Command Code Adapter

## Current Evidence (2026-10-21)
- `sila salam` reports `heal: 1 automatic fix pending` and `sila heal --dry-run` says `Provision sila wiring in this repository`, affecting `AGENTS.md`.
- `sila install --help` lists `claude`, `cursor`, and `devin` only; there is no `command-code` target.
- `sila agents check` passes on the managed block hash, but the full nine-verb protocol is only in `.cursor/rules/000-sila-principles.mdc`.
- `.githooks/` contains only `pre-commit` and `pre-push`; no session-start or session-end hook exists.
- `.commandcode/mcp.json` registers `sila`, but the MCP surface exposes only continuity tools such as `prime_context`, `recall`, `save_lesson`, `save_memo`, `save_handoff`, and `get_status`.
- `sila prime` surfaces an active handoff and `fact_c831b36c`, a prior bot-attribution lesson, proving the skipped preflight had concrete value.
- The current ledger is 114 goals, 518/518 tasks, but only 4 verified, 22 stale, and 88 unverified.

## Architectural Invariants
- AGENTS.md remains the single Command Code rule surface; Cursor rules may be generated separately but cannot be required.
- An adapter update must be idempotent and reversible.
- No hook may silently no-op when Sila is missing.

## File & Interface Contracts
- `AGENTS.md`: Managed block contains the seven-step loop, the nine verbs, and explicit trigger/when-not-to-use guidance.
- `.commandcode/mcp.json`: Contains exactly one Sila server entry and preserves all other servers.
- `docs/SILA-COMMAND-CODE.md`: Names install, verify, repair, and rollback commands.

## Failure Modes
- Command Code config is absent.
- AGENTS.md has a conflicting user block.
- MCP server command path is stale.
- Install runs without write permission.

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
- Should the adapter live upstream in Sila core or as a repo-local generator?
- Should AGENTS.md contain the full persona/audit/skill trigger table, or link to a generated appendix?
- What is the compatibility contract for older Command Code versions?
