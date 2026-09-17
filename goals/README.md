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
- 🟢 **Implemented & Verified**: 2
- 🔵 **In Progress (Claimed)**: 1
- 🟡 **Ready to Execute (Pending)**: 1 (0 independent ⚡, 1 unblocked 🔗)
- ⛔ **Blocked on Prerequisites**: 4
- 📋 **Execution Tasks Progress**: 10/42 completed (23%)

---

## 🔵 In Progress (Claimed Goals)

| Goal Package | Mode | Agent | Status & Note | Started |
| :--- | :---: | :--- | :--- | :--- |
| [`report-period-totals`](report-period-totals/goal.md) | `⚡ Indep` | `@opencode` | fix weekly/monthly totals and empty-row formatting | 9m0s ago |

---

## 🟡 Ready to Execute (Pending Goals)

| Goal Package | Mode & Sequence | Dependencies | Focus & Description | Launch Command |
| :--- | :---: | :--- | :--- | :--- |
| [`tui-shell-overview`](tui-shell-overview/goal.md) | `🚢 SHIP 🔗 Ready (#3)` | Deps met: charting-analytics | TUI shell and Overview view — Add the thermal dashboard command: a Bubble Tea v2 application shell with tab ro... | `/goal goals/tui-shell-overview/goal.md` |
| [`tui-projects-view`](tui-projects-view/goal.md) | `🚢 SHIP ⛔ Blocked (#4)` | Prereqs: tui-shell-overview | TUI Projects view with drill-down — Ranked project table with token and cost bars, sortable columns, and filters for... | `/goal goals/tui-projects-view/goal.md` |
| [`tui-mix-models-views`](tui-mix-models-views/goal.md) | `🚢 SHIP ⛔ Blocked (#5)` | Prereqs: tui-shell-overview | TUI Mix and Models views — Stacked share over time by tool and by model, a switching panel with dominant to... | `/goal goals/tui-mix-models-views/goal.md` |
| [`tui-stats-view`](tui-stats-view/goal.md) | `🚢 SHIP ⛔ Blocked (#6)` | Prereqs: tui-shell-overview | TUI Stats view — Histogram of daily tokens or cost with a log toggle, weekday profile, top days, ... | `/goal goals/tui-stats-view/goal.md` |
| [`charting-flags-docs-release`](charting-flags-docs-release/goal.md) | `🚢 SHIP ⛔ Blocked (#7)` | Prereqs: tui-projects-view, tui-mix-models-views, tui-stats-view | Chart flags, docs, and 0.7.0 release — Add --chart to projects, models and the period reports, print concrete bar rows ... | `/goal goals/charting-flags-docs-release/goal.md` |

---

## 🟢 Implemented & Verified

| Goal Package | Mode | Shape | Task / Ref | Commit |
| :--- | :---: | :--- | :--- | :--- |
| [`charting-dep-spike`](charting-dep-spike/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `fa74fe7` |
| [`charting-analytics`](charting-analytics/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `9c28a0a` |

