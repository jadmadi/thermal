# Plan: Complete Sila Onboarding

## Execution Steps

- [x] **Phase 1 — Review the inventory**
  - Run `sila onboard --json` and list every unmanaged surface: brain stores, markdown task lists, drafts, instructions, and the runbook.
  - For each surface, decide: map into sila goals, defer with a recorded reason, or freeze.

- [x] **Phase 2 — Map legacy work**
  - Convert open markdown tasks with `sila onboard --convert-tasks --from=<file> --review --yes` when the list is real work; otherwise record the decision with `sila memo`.
  - Reconcile duplicate or stale goals with `sila goals` and close what is already done.

- [x] **Phase 3 — Record what the project knows**
  - Record the verified runbook commands and any architectural directives with `sila memo`.
  - Record solved problems with `sila lesson` so the next session does not rediscover them.

- [x] **Phase 4 — Verify**
  - Run `sila onboard check --strict`. Resolve every critical failure and review the warnings.
  - Commit and run `sila goals`.
