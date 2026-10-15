# Goal: Independently verify a released artifact's SLSA provenance

## Goal Description
[`docs/PROMISES.md`](../../docs/PROMISES.md) records release provenance as **unverified**: the release workflow and runbook exist in-tree, but no downloaded artifact has been checked with `slsa-verifier`. Close that gap by downloading a published release (`checksums.txt`, `checksums.txt.intoto.jsonl`, one platform archive), running the documented verification, and recording the exact commands, outputs, and exit codes — or, if the verification fails or the network is unavailable, recording a disclosed failure that keeps the claim unverified.

Priority: **P2**. Follow-up of the promise-claim-evidence ledger (section 9).

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 111
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `docs/PROMISES.md` (outcome update for section 9)
- `docs/DISTRIBUTION.md` (verification runbook corrections if the documented steps drift from reality)
- `docs/evaluations/2026-10-15-release-provenance/` (raw verifier output and command log)

## References
- [Evidence and constraints](facts.md)
- [Execution plan](plan.md)

## Done Condition
1. A published release asset is downloaded and verified with `slsa-verifier` against its `checksums.txt.intoto.jsonl`, and `sha256sum --check` passes for the archive; raw output (commands, exit codes, verifier version) is recorded under `docs/evaluations/2026-10-15-release-provenance/`.
2. `docs/PROMISES.md` section 9 moves to **verified** with the artifact, version, source revision and commands, or stays **unverified** with a disclosed failure (missing provenance, signature mismatch, network unavailable) — a failure is evidence, not a pass.
3. No publishing, tagging, release creation, or workflow change; verification is read-only against GitHub Releases.
4. Verification commands exit with recorded, reproducible results; implementation-attributed commit before closure.

## Scope Boundary
Verification only. Do not modify `.github/workflows/release.yml`, do not create or re-publish releases, and do not weaken the ledger wording if verification cannot complete. If the environment cannot reach GitHub, record a blocked run with evidence and leave the claim unverified.
