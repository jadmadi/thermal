# Thermal Contributor & Maintainer Playbook

This document defines the review philosophy, communication SLAs, review hierarchy, and contributor retention workflows for maintainers and reviewers of **Thermal** (`github.com/jadmadi/thermal`).

---

## 1. Review Philosophy & The Two-Clock Rule

Maintainer attention is Thermal's scarcest resource. To maintain healthy communication without maintainer burnout, Thermal operates under a **Two-Clock review commitment**:

1. **Clock 1 — First Acknowledgment (24 to 48 Hours)**:
   - A maintainer triages the PR or issue, tags relevant labels, and confirms that automated checks (`scripts/check.sh`) have run.
   - If CLA sign-off is missing, the maintainer immediately posts the standard remediation instructions.
2. **Clock 2 — Deep Technical Review (within 72 Hours)**:
   - Aligned with the 72-hour lazy consensus window in [GOVERNANCE.md](../GOVERNANCE.md).
   - Deep review evaluates architectural invariants, correctness, and performance benchmarks.

---

## 2. Review Order Hierarchy

Reviewers must evaluate pull requests in strict hierarchical order to avoid wasting time debating formatting on code that violates architectural invariants:

```
┌─────────────────────────────────────────────────────────────┐
│ Level 1: Scope & Anti-Goals Check (GOVERNANCE.md)            │
└──────────────────────────────┬──────────────────────────────┘
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ Level 2: Correctness, Tests & Verification (scripts/check.sh)│
└──────────────────────────────┬──────────────────────────────┘
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ Level 3: Architectural Invariants (AGENTS.md)                │
└──────────────────────────────┬──────────────────────────────┘
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ Level 4: Machine-Formattable Style (gofmt, SPDX, trailers)  │
└─────────────────────────────────────────────────────────────┘
```

1. **Level 1: Scope & Anti-Goals**: Does the PR introduce cloud telemetry, background daemons, GUI bloat, or non-local network dependencies? If yes, decline immediately.
2. **Level 2: Correctness & Tests**: Does `./scripts/check.sh` pass cleanly with race detection? Are edge cases (empty DBs, corrupted logs) covered by unit tests?
3. **Level 3: Architecture & Invariants**: Does SQLite use `?mode=ro` and `PRAGMA mmap_size`? Are goroutines bounded by worker pools? Are models passed through `modelName()`? Are token types disjoint?
4. **Level 4: Machine Style**: Run `gofmt` and verify conventional commits. Do not bike-shed stylistic preferences that the machine formatter does not enforce.

---

## 3. "Decline Fast and Kindly" Template

When a contribution falls into the Anti-Goals defined in `GOVERNANCE.md`, maintainers must decline promptly and respectfully without leaving the PR hanging in limbo:

> Hi @contributor, thank you for putting effort into this pull request!
>
> In accordance with Thermal's project governance ([GOVERNANCE.md](https://github.com/jadmadi/thermal/blob/main/GOVERNANCE.md)), Thermal's architectural scope is strictly focused on local-first, zero-daemon, terminal-native telemetry with a sub-10ms cached invocation ceiling.
>
> Because [FEATURE_NAME] introduces [EXPLANATION: e.g., cloud network dependency / persistent daemon runner / GUI framework bloat], this falls into our documented Anti-Goals and cannot be accepted into the core repository.
>
> You are very welcome to maintain this as an independent extension or fork under the terms of our AGPL-3.0 Community Reciprocity Covenant. Thank you again for understanding our strict scope boundaries!

---

## 4. Contributor Credit & Merge Workflow

When merging an external pull request:

1. **Preserve Commit Authorship**: Use "Rebase and Merge" or standard merge commit to retain the contributor's Git identity and `Signed-off-by` trailer.
2. **Release Notes Attribution**: `release-please` automatically includes conventional commit subjects in `CHANGELOG.md`. For significant contributions, explicitly credit the contributor by GitHub username.
3. **Prompt Thank You**: Leave a warm merge comment acknowledging the specific impact (e.g., *"Merged! Thanks @contributor for adding the [X] loader and preserving our disjoint token invariants."*).

---

## 5. Second-Contribution Retention Loop

To convert drive-by contributors into recurring maintainers:

> Hi @contributor, your PR for [FEATURE] just landed in `main`! Thank you for the clean implementation and high test coverage.
>
> If you're interested in taking on another piece of the roadmap, we have a few curated tasks in [`docs/GOOD_FIRST_ISSUES.md`](https://github.com/jadmadi/thermal/blob/main/docs/GOOD_FIRST_ISSUES.md) that build directly on what you just did. Let me know if one catches your eye, and I'll gladly assign it to you!
