# Goal: Complete Sila Onboarding

## Goal Description
Wire thermal-streak fully into sila and verify it: map the legacy task surfaces found by `sila onboard`, reconcile duplicates, record the verified runbook and directives, then prove the result with `sila onboard check --strict`. The goal closes when the check is green.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship

## Done Condition
1. Every unmanaged surface from `sila onboard` is mapped into a sila goal, deferred with a recorded reason, or frozen in place.
2. The verified runbook commands are recorded and current.
3. Directives and solved-problem lessons from the legacy system are recorded with `sila memo` and `sila lesson`.
4. `sila onboard check --strict` exits clean.
5. Goal closed via commit containing `(goals/sila-onboarding/goal.md)`.
