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
- **Total Goals**: 11
- 🟢 **Implemented & Verified**: 9
- 🔵 **In Progress (Claimed)**: 1
- 🟡 **Ready to Execute (Pending)**: 0 (0 independent ⚡, 0 unblocked 🔗)
- ⛔ **Blocked on Prerequisites**: 1
- 📋 **Execution Tasks Progress**: 48/60 completed (80%)

---

## 🔵 In Progress (Claimed Goals)

| Goal Package | Mode | Agent | Status & Note | Started |
| :--- | :---: | :--- | :--- | :--- |
| [`unpriceable-token-sources`](unpriceable-token-sources/goal.md) | `⚡ Indep` | `@opencode` | Codex day-to-rollout mapping, then the activity-count rule | 0s ago |

---

## 🟡 Ready to Execute (Pending Goals)

| Goal Package | Mode & Sequence | Dependencies | Focus & Description | Launch Command |
| :--- | :---: | :--- | :--- | :--- |
| [`pricing-catalog-coverage`](pricing-catalog-coverage/goal.md) | `🚢 SHIP ⛔ Blocked (#11)` | Prereqs: unpriceable-token-sources | Pricing catalog coverage for tool-native model names — Classify each model id the catalog cannot price, then choose a mechanism with th... | `/goal goals/pricing-catalog-coverage/goal.md` |

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
| [`report-period-totals`](report-period-totals/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `c9c6af8` |

