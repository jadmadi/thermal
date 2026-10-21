# Facts: Sila Mandatory Preflight Gate

## Current Evidence (2026-10-21)
- `sila salam` reports `heal: 1 automatic fix pending` and `sila heal --dry-run` says `Provision sila wiring in this repository`, affecting `AGENTS.md`.
- `sila install --help` lists `claude`, `cursor`, and `devin` only; there is no `command-code` target.
- `sila agents check` passes on the managed block hash, but the full nine-verb protocol is only in `.cursor/rules/000-sila-principles.mdc`.
- `.githooks/` contains only `pre-commit` and `pre-push`; no session-start or session-end hook exists.
- `.commandcode/mcp.json` registers `sila`, but the MCP surface exposes only continuity tools such as `prime_context`, `recall`, `save_lesson`, `save_memo`, `save_handoff`, and `get_status`.
- `sila prime` surfaces an active handoff and `fact_c831b36c`, a prior bot-attribution lesson, proving the skipped preflight had concrete value.
- The current ledger is 114 goals, 518/518 tasks, but only 4 verified, 22 stale, and 88 unverified.

## Architectural Invariants
- No implementation edit without a fresh preflight artifact.
- The artifact cannot be forged by creating an empty file; it must come from the gate.
- A skip is always visible.

## File & Interface Contracts
- `scripts/sila-preflight.sh`: Returns 0 only after the required checks and writes a JSON stamp.
- `AGENTS.md`: Contains the gate as a top-of-file MUST, not a footnote.
- ``.commandcode/config.json``: Invokes the gate on session start and before edit-capable tools.

## Failure Modes
- Sila CLI missing.
- MCP unavailable.
- Goal has no ready item.
- Preflight stamp stale after midnight.
- User explicitly asks for a read-only answer.

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
- Can Command Code block a tool call with a pre-tool-use hook?
- Should the gate run before goal claim or after?
- What is the exact freshness window?
