# Goals Status & Pragmatic Execution Roadmap

This index tracks all goal packages under `goals/`, their execution mode (Independent ⚡ vs Dependent 🔗), dependencies, and progress status.

## 🤖 Agent Guide: How to Create, Claim & Execute Goals

1. **Create a Goal**: Scaffold a new goal package conforming to standards:
   ```bash
   sila goals create <slug> --title="..." [--depends-on="..."] [--independent]
   ```
2. **Claim a Goal**: Lock the goal so no other agent duplicates work:
   ```bash
   sila goals claim <slug> --agent=<name> [--note="evaluating..."]
   ```
3. **Launch & Implement**: Follow `facts.md` and `plan.md` until all tests pass:
   ```bash
   /goal goals/<slug>/goal.md
   ```
4. **Progress Updates**: Report status updates during execution:
   ```bash
   sila goals report <slug> "Running e2e test suite"
   ```
5. **Commit with Goal Reference**: Include the goal path in your commit message:
   ```bash
   git commit -m "feat(scope): implement description (goals/<slug>/goal.md)"
   ```
6. **Auto-Reconciliation**: Run `sila goals` to scan commits, mark 🟢 **Implemented**, and clear the claim lock.

## Summary
- **Total Goals**: 20
- 🟢 **Implemented & Verified**: 19
- 🟡 **Ready to Execute (Pending)**: 1 (1 independent ⚡, 0 unblocked 🔗)
- 📋 **Execution Tasks Progress**: 108/121 completed (89%)

---

## 🟡 Ready to Execute (Pending Goals)

| Goal Package | Mode & Sequence | Dependencies | Focus & Description | Launch Command |
| :--- | :---: | :--- | :--- | :--- |
| [`workload-replay`](workload-replay/goal.md) | `🚢 SHIP ⚡ Independent` | - | Workload Replay Simulation (`thermal replay`) — Implement the `thermal replay` command to simulate historical developer workload... | `/goal goals/workload-replay/goal.md` |

---

## 🟢 Implemented & Verified

| Goal Package | Mode | Shape | Task / Ref | Commit |
| :--- | :---: | :--- | :--- | :--- |
| [`charting-dep-spike`](charting-dep-spike/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `fa74fe7` |
| [`charting-analytics`](charting-analytics/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `9c28a0a` |
| [`tui-shell-overview`](tui-shell-overview/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `050f4a5` |
| [`tui-projects-view`](tui-projects-view/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `0619190` |
| [`tui-mix-models-views`](tui-mix-models-views/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `6a591f8` |
| [`tui-stats-view`](tui-stats-view/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `338b8bd` |
| [`charting-flags-docs-release`](charting-flags-docs-release/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `66a7466` |
| [`devin-model-attribution`](devin-model-attribution/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `e4d542c` |
| [`unpriceable-token-sources`](unpriceable-token-sources/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `ce31aaa` |
| [`pricing-catalog-coverage`](pricing-catalog-coverage/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `774f634` |
| [`jsonl-line-ceilings`](jsonl-line-ceilings/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `ada5b52` |
| [`jsonl-scan-diagnostics`](jsonl-scan-diagnostics/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `247bb80` |
| [`cache-efficiency-metrics`](cache-efficiency-metrics/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `6e070bd` |
| [`claude-message-dedupe`](claude-message-dedupe/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `104ba23` |
| [`dsh-loader`](dsh-loader/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `414a815` |
| [`hermes-loader`](hermes-loader/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `9e90d01` |
| [`report-period-totals`](report-period-totals/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `c9c6af8` |
| [`sila-onboarding`](sila-onboarding/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `92ea6ee` |
| [`telemetry-and-ui-audit`](telemetry-and-ui-audit/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `662f25a` |

