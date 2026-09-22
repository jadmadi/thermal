# Plan: Retire Legacy MIT Releases and Reset Release Line to AGPL-3.0

## Execution Steps

- [x] **Phase 1 — Release Inventory & Historical Audit**:
  - Run `gh release list` to inventory all published legacy releases (e.g. `v0.7.0`, `v0.6.0`, etc.).
  - Record the commit hashes, release dates, and changelogs in `docs/RELEASE.md` under "Historical MIT Release Archive".

- [x] **Phase 2 — Scripted Retirement Tool with Safety Guards**:
  - Implement `scripts/retire_legacy_releases.sh` supporting:
    - `--dry-run` (default): lists target releases and tags without modifying remote.
    - `--confirm`: executes deletion via `gh release delete <tag> --cleanup-tag -y`.
    - Safety rail: verifies caller is authorized and checks repository state before execution.

- [x] **Phase 3 — Release Pipeline Preparation**:
  - Verify `.github/workflows/release.yml` and `.goreleaser.yml` to ensure future release builds inject AGPL-3.0 metadata, `DUAL-LICENSE.md`, and `NOTICES.md` into all release tarballs.
  - Document the clean reset procedure in `docs/RELEASE.md`.

- [x] **Phase 4 — Execution & Verification**:
  - Run `scripts/retire_legacy_releases.sh --dry-run` to audit.
  - Upon explicit user approval, run retirement on remote (preserved in dry-run mode per safety directives).
  - Verify `gh release list` returns empty or only the target release set.

- [x] **Phase 5 — Attribution Commit**:
  - Commit changes with message `feat(release): script legacy release retirement & document AGPL-3.0 release line (goals/purge-legacy-mit-releases/goal.md)`.
  - Run `sila goals` to mark goal completed.
