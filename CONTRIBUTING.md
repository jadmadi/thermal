# Contributing to Thermal

Thank you for your interest in contributing to **Thermal** (`github.com/jadmadi/thermal`)! We welcome bug reports, new agent loaders, performance optimizations, and documentation fixes.

---

## Code Standards & Philosophy

Thermal is designed for extreme terminal speed, zero-allocation data ingestion, and rich terminal visuals:

1. **Read `AGENTS.md`**: All architectural invariants (read-only SQLite connections with `PRAGMA mmap_size`, bounded concurrency, canonical model naming, JSONL line ceilings) are documented in [`AGENTS.md`](AGENTS.md).
2. **Quality Gates**: Every pull request must pass:
   - `go test -v -race ./...` (unit tests and race detection)
   - `./scripts/simulated_user_gate.sh` (comprehensive simulated user release gate)

---

## Contributor License Agreement (CLA)

Thermal is a **dual-licensed** project:
- The public codebase is released under the **GNU Affero General Public License v3.0 (AGPL-3.0)** to protect against closed-source wrapper services and enforce open-source reciprocity.
- The project maintainer (**Jad Madi**) also provides **Commercial & Enterprise Licenses** to companies requiring proprietary deployment, closed SaaS integration, or relief from copyleft requirements.

To maintain the legal authority to offer commercial licenses without copyright fragmentation, **all external contributions require a Contributor License Agreement (CLA)**.

### Contributor Agreement Terms

By submitting a pull request or code contribution to Thermal, you agree to the following terms:

1. **Copyright Ownership**: You retain copyright ownership of your contribution.
2. **Open Source Grant**: You license your contribution to the public under the terms of the GNU Affero General Public License v3.0 (AGPL-3.0).
3. **Commercial & Dual-License Grant**: You grant **Jad Madi** a perpetual, worldwide, non-exclusive, royalty-free, transferable, and irrevocable license to reproduce, prepare derivative works of, publicly display, sublicense, and distribute your contribution (and derivative works thereof) under terms of Jad Madi's choice, including under commercial, proprietary, or dual-license agreements.
4. **Original Work Warranty**: You represent that you are legally entitled to grant the above license. If your employer has rights to intellectual property that you create, you represent that you have received permission to make contributions on behalf of that employer.

### Community Reciprocity Covenant

We believe contributor trust requires mutual commitment. Thermal is bound by the **Community Reciprocity Covenant** documented in [DUAL-LICENSE.md](DUAL-LICENSE.md):
- **Permanent Open Source**: The core CLI, loaders, TUI, and telemetry engine will remain licensed under OSI-approved copyleft (GNU AGPL-3.0) in perpetuity.
- **No Unilateral Relicensing**: Community contributions will never be retroactively relicensed into closed-source or proprietary products without contributor consent.
- **Commercial Demarcation**: Commercial enterprise licenses are strictly demarcated for proprietary embedding and enterprise fleet deployments, not for gating local terminal features.

### How to Sign Off

Include a standard Git `Signed-off-by` trailer in every commit message using the `-s` flag:

```bash
git commit -s -m "feat(loaders): add support for new agent harness"
```

This trailer certifies that you authored the commit and accept the Thermal Contributor License Agreement terms.

### Automated CI Enforcement & PR Recovery

Our GitHub Actions CI pipeline strictly verifies that **every commit** in a pull request carries a valid `Signed-off-by` trailer. Pull requests containing unsigned commits cannot be merged.

If your PR check fails due to missing sign-off trailers, fix your commits locally and update your PR branch:

```bash
# To fix a single recent commit:
git commit --amend -s
git push --force-with-lease

# To fix multiple commits in your feature branch:
git rebase --signoff origin/main
git push --force-with-lease
```

---

## Out of Scope & Anti-Goals

Before investing significant engineering time, please review Thermal's architectural scope and Anti-Goals codified in **[GOVERNANCE.md](GOVERNANCE.md)**. The following are explicitly out of scope:

- **Cloud Telemetry**: Features requiring external server pings, cloud accounts, or tracking. Thermal is strictly local-first and offline-capable.
- **Background Daemons**: Persistent system daemons or background runners. Thermal is an on-demand CLI tool.
- **Heavy Web/GUI Frameworks**: Electron, Tauri, or browser embeddings in the core binary.
- **Mutating Operations on User Data**: Writing or modifying local agent session databases.
- **Fuzzy Pricing**: Price catalog lookups that guess or average pricing models.

---

## Finding Work & Claiming Issues

Looking for somewhere to start?
1. Check **[docs/GOOD_FIRST_ISSUES.md](docs/GOOD_FIRST_ISSUES.md)** for curated, beginner-friendly starter tasks.
2. Browse our [GitHub Issues](https://github.com/jadmadi/thermal/issues) labeled `good first issue` or `help wanted`.

### Claiming Convention & Inactivity Policy
- To claim an issue, leave a comment on the issue asking for assignment.
- **7-Day Inactivity Timeout**: If an assigned issue sees no pull request or status update within **7 days**, the assignment is released so other contributors can proceed.

---

## Response Targets & Review Expectations

We value your time and effort. Thermal maintainers adhere to the **Two-Clock review commitment** documented in [docs/CONTRIBUTOR_PLAYBOOK.md](docs/CONTRIBUTOR_PLAYBOOK.md):
- **Clock 1 (24 to 48 hours)**: Initial acknowledgment, triage labeling, and automated checks confirmation.
- **Clock 2 (within 72 hours)**: Deep technical review and architectural feedback (aligned with the 72-hour lazy consensus in [GOVERNANCE.md](GOVERNANCE.md)).

---

## Local Pre-Flight Verification

To ensure a seamless, 100% first-pass green build in CI, run the unified pre-flight script locally before pushing or opening a pull request:

```bash
./scripts/check.sh
```

This single command verifies CLA commit trailers, `gofmt` code formatting, `go vet` static analysis, unit tests with race detection (`go test -race ./...`), binary compilation, and the simulated user release gate.


