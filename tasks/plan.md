# Thermal promise delivery remediation

Evaluation: [2026-10-14 report](../docs/evaluations/2026-10-14-promise-vs-delivery.md).
Repository/cwd: `/mnt/Jad/github/lab/thermal-streak`.
Tasks are tracked in Sila goal packages, not a second task checklist.

## Decisions and order

Design choice: repair trust before expanding features. Preserve prior implementation commits,
checked historical tasks and acceptance IDs; appended remediation cards govern new acceptance.
No application changes are authorized by this planning document itself.

1. [receipt-evidence-integrity](../goals/receipt-evidence-integrity/remediation-2026-10-14.md), P1.
2. [workload-replay](../goals/workload-replay/remediation-2026-10-14.md), P1, cards R1 then R2.
3. [chatgpt-rolling-burst-replay](../goals/chatgpt-rolling-burst-replay/remediation-2026-10-14.md),
   P1, after workload-replay.
4. [setup-health-audit](../goals/setup-health-audit/remediation-2026-10-14.md), P2.
5. [ingestion-performance-baseline](../goals/ingestion-performance-baseline/remediation-2026-10-14.md), P2.
6. [promise-claim-evidence](../goals/promise-claim-evidence/plan.md), P2, after all five.

Independent readiness means no prerequisite, not permission to edit shared files concurrently.
Execute serially: receipt/replay share tests, migration docs and metadata.

## Common implementation and review contract

All cards are planning only and return `ready_for_review`, never self-approved. Reviewer: Jad Madi
or an independent reviewer he assigns. After receipt/replay and again before public claim changes,
review actual case evidence against this contract. Tracker closure does not establish acceptance.

Keep source ingestion read-only, diagnostics off JSON stdout, and user data out of evidence.
No new dependencies, daemons, provider calls, remote push, or unrelated cleanup. Preserve the
pre-existing README and `.commandcode/` work. Do not edit AGENTS.md or governance under these cards.
For each changed public CLI/JSON/math contract append a six-part `docs/MIGRATION.md` entry.
Update `project.jsonc` capability wording and local `lastEdit` only if affected by that card.
These two documentation files are permitted companion edits to each implementation card.

Commands run from the cwd above, with Go 1.27.2 and a matching GOROOT (unset stale GOROOT).
Use a writable GOCACHE and record color settings. Final candidate checks, expected exit 0:
`env -u NO_COLOR go test -race ./...`, `go vet ./...`,
`go build -o /tmp/thermal-test ./cmd/thermal`, `./scripts/simulated_user_gate.sh`.
A socket-denied environment is blocked verification, not a product failure or a pass.
Run gofmt only on touched Go files. Retain exact command exits and raw output for the final revision.

Return a JSON object with `schema: "thermal/task-result/v1"`, `status` (`ready_for_review` or `blocked`),
`summary`, `files` (relative paths), `cases` (objects: `id`, `result`, `evidence`), `commands`
(objects: `command`, `exit`, `log`), `candidate` (object: `head`, `diff_sha256`), and `blockers`
(array of evidence plus required decision). No missing fields, no fabricated receipts.

If a fixed interface cannot express a required case, stop that card and report the conflict.
After two repeated failures, return to the planner. Do not broaden scope or weaken an assertion.

## Validation debt outside these cards

The pre-existing global Sila validator reports a missing `html-docs-licensing-parity` prerequisite
and missing `goal.md`/`facts.md` under `goals/mcp-server-card/`, plus sequence warnings. Do not invent
historical goal content or remove dependencies merely to obtain a green validator. The new package
and reopened packages must introduce no additional errors. These registry defects remain disclosed.
