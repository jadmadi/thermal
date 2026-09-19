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
- **Total Goals**: 8
- 🟢 **Implemented & Verified**: 7
- 🟡 **Ready to Execute (Pending)**: 1 (0 independent ⚡, 1 unblocked 🔗)
- 📋 **Execution Tasks Progress**: 36/42 completed (85%)

---

## 🟡 Ready to Execute (Pending Goals)

| Goal Package | Mode & Sequence | Dependencies | Focus & Description | Launch Command |
| :--- | :---: | :--- | :--- | :--- |
| [`charting-flags-docs-release`](charting-flags-docs-release/goal.md) | `🚢 SHIP 🔗 Ready (#7)` | Deps met: tui-projects-view, tui-mix-models-views, tui-stats-view | Chart flags, docs, and 0.7.0 release — Add --chart to projects, models and the period reports, print concrete bar rows ... | `/goal goals/charting-flags-docs-release/goal.md` |

---

## 🟢 Implemented & Verified

| Goal Package | Mode | Shape | Task / Ref | Commit |
| :--- | :---: | :--- | :--- | :--- |
| [`charting-dep-spike`](charting-dep-spike/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `fa74fe7` |
| [`charting-analytics`](charting-analytics/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `9c28a0a` |
| [`tui-shell-overview`](tui-shell-overview/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `050f4a5` |
| [`tui-projects-view`](tui-projects-view/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `e15056f` |
| [`tui-mix-models-views`](tui-mix-models-views/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `6a591f8` |
| [`tui-stats-view`](tui-stats-view/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `338b8bd` |
| [`report-period-totals`](report-period-totals/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `c9c6af8` |

