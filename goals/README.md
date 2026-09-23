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
- **Total Goals**: 59
- 🟢 **Implemented & Verified**: 44
- 🔵 **In Progress (Claimed)**: 1
- 🟡 **Ready to Execute (Pending)**: 13 (10 independent ⚡, 3 unblocked 🔗)
- ⛔ **Blocked on Prerequisites**: 1
- 🎯 **By Tier**: 1 immediate (Tier 1), 58 roadmap (Tier 2), 0 nice-to-have (Tier 3), 0 wont-fix (Tier 4)
- 📋 **Execution Tasks Progress**: 231/264 completed (87%)

---

## 🔵 In Progress (Claimed Goals)

| Goal Package | Tier | Mode | Agent | Status & Note | Started |
| :--- | :---: | :---: | :--- | :--- | :--- |
| [`tui-projects-view`](tui-projects-view/goal.md) | `🗺️ Roadmap` | `🔗 Dep` | `@Antigravity` | in progress | 0s ago |

---

## 🟡 Ready to Execute (Pending Goals)

| Goal Package | Tier | Mode & Sequence | Dependencies | Focus & Description | Launch Command |
| :--- | :---: | :---: | :--- | :--- | :--- |
| [`devin-model-attribution`](devin-model-attribution/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent (#8)` | - | Devin model attribution and honest unestimated reporting — Emit Devin per-model daily tokens from sessions.model so the estimator can price... | `/goal goals/devin-model-attribution/goal.md` |
| [`agpl-dual-licensing`](agpl-dual-licensing/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent` | - | AGPL-3.0 Relicensing, Commercial Dual-Licensing & Contributor Gating — Relicense Thermal to GNU AGPLv3, establish commercial dual-licensing track, impl... | `/goal goals/agpl-dual-licensing/goal.md` |
| [`cache-efficiency-metrics`](cache-efficiency-metrics/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent` | - | Cache Efficiency & Disjoint Token Breakdown — Expose disjoint canonical token metrics (uncached input, cache read, cache write... | `/goal goals/cache-efficiency-metrics/goal.md` |
| [`claude-message-dedupe`](claude-message-dedupe/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent` | - | Claude Code Message-Level Deduplication — Implement message-level deduplication in the Claude Code loader (`internal/loade... | `/goal goals/claude-message-dedupe/goal.md` |
| [`dsh-loader`](dsh-loader/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent` | - | Support DeepSeek harness (dsh) in Thermal — Add dsh loader engine, registry, tests, and documentation to thermal | `/goal goals/dsh-loader/goal.md` |
| [`hermes-loader`](hermes-loader/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent` | - | Nous Hermes Agent Loader — Implement an ingestion loader for **Nous Hermes** (`hermes`), an open-source, lo... | `/goal goals/hermes-loader/goal.md` |
| [`report-period-totals`](report-period-totals/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent` | - | Period report totals and table format — Weekly and monthly totals must equal the displayed rows for the requested window... | `/goal goals/report-period-totals/goal.md` |
| [`sila-onboarding`](sila-onboarding/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent` | - | Complete Sila Onboarding — Wire thermal-streak fully into sila and verify it: map the legacy task surfaces ... | `/goal goals/sila-onboarding/goal.md` |
| [`telemetry-and-ui-audit`](telemetry-and-ui-audit/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent` | - | Telemetry Accuracy, Pricing Estimates, and UI Polish (5 Issues) — Fix 5 issues identified during operational audit of Thermal: 1. **Pricing Estima... | `/goal goals/telemetry-and-ui-audit/goal.md` |
| [`workload-replay`](workload-replay/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent` | - | Workload Replay Simulation (`thermal replay`) — Implement the `thermal replay` command to simulate historical developer workload... | `/goal goals/workload-replay/goal.md` |
| [`pricing-catalog-coverage`](pricing-catalog-coverage/goal.md) | `🗺️ Roadmap` | `🚢 SHIP 🔗 Ready (#11)` | Deps met: unpriceable-token-sources | Pricing catalog coverage for tool-native model names — Classify each model id the catalog cannot price, then choose a mechanism with th... | `/goal goals/pricing-catalog-coverage/goal.md` |
| [`tui-mix-models-views`](tui-mix-models-views/goal.md) | `🗺️ Roadmap` | `🚢 SHIP 🔗 Ready (#5)` | Deps met: tui-shell-overview | TUI Mix and Models views — Stacked share over time by tool and by model, a switching panel with dominant to... | `/goal goals/tui-mix-models-views/goal.md` |
| [`tui-stats-view`](tui-stats-view/goal.md) | `🗺️ Roadmap` | `🚢 SHIP 🔗 Ready (#6)` | Deps met: tui-shell-overview | TUI Stats view — Histogram of daily tokens or cost with a log toggle, weekday profile, top days, ... | `/goal goals/tui-stats-view/goal.md` |
| [`charting-flags-docs-release`](charting-flags-docs-release/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⛔ Blocked (#7)` | Prereqs: tui-projects-view, tui-mix-models-views, tui-stats-view | Chart flags, docs, and 0.7.0 release — Add --chart to projects, models and the period reports, print concrete bar rows ... | `/goal goals/charting-flags-docs-release/goal.md` |

---

## 🟢 Implemented & Verified

| Goal Package | Mode | Shape | Task / Ref | Commit |
| :--- | :---: | :--- | :--- | :--- |
| [`charting-dep-spike`](charting-dep-spike/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `d1b68ab` |
| [`governance-foundations`](governance-foundations/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `332ad26` |
| [`repo-alignment-and-security`](repo-alignment-and-security/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `bee76ff` |
| [`version-migration-guide`](version-migration-guide/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `12d9833` |
| [`unpriceable-token-sources`](unpriceable-token-sources/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `09eb367` |
| [`jsonl-line-ceilings`](jsonl-line-ceilings/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `86d6ebb` |
| [`acp-commerce-discovery`](acp-commerce-discovery/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `c9ee829` |
| [`agent-link-headers`](agent-link-headers/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `cf43a37` |
| [`agent-skills-discovery`](agent-skills-discovery/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `e087bf4` |
| [`ard-capability-manifest`](ard-capability-manifest/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `dbbb856` |
| [`auth-md-agent-registration`](auth-md-agent-registration/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `8483316` |
| [`cli-daily-update-check`](cli-daily-update-check/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `2c387fb` |
| [`cli-upgrade-file-integrity`](cli-upgrade-file-integrity/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `15d2ffb` |
| [`database-loader-iteration-safety`](database-loader-iteration-safety/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `c53c32b` |
| [`dns-aid-discovery`](dns-aid-discovery/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `769b9d6` |
| [`docs-parity-and-human-simulation`](docs-parity-and-human-simulation/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `64d1837` |
| [`embedded-web-dashboard`](embedded-web-dashboard/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `dceba19` |
| [`leaderboard-slice-preallocation`](leaderboard-slice-preallocation/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `f375fdf` |
| [`license-strategy-and-enforcement`](license-strategy-and-enforcement/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `98772b5` |
| [`loader-bounded-concurrency`](loader-bounded-concurrency/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `0c4fffb` |
| [`markdown-for-agents`](markdown-for-agents/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `3de20b6` |
| [`mcp-server-card`](mcp-server-card/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `f7350bf` |
| [`model-identity-and-alias-resolution`](model-identity-and-alias-resolution/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `e460d4a` |
| [`purge-bot-contributors`](purge-bot-contributors/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `ced4751` |
| [`rfc9727-api-catalog`](rfc9727-api-catalog/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `e657745` |
| [`robots-content-signals`](robots-content-signals/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `fd5ca83` |
| [`setup-health-audit`](setup-health-audit/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `8afd798` |
| [`slsa3-release-provenance`](slsa3-release-provenance/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `78b8258` |
| [`stateless-telemetry-sharing`](stateless-telemetry-sharing/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `7557bf9` |
| [`streak-parsing-defensive-guards`](streak-parsing-defensive-guards/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `a2467d7` |
| [`token-yield-metrics`](token-yield-metrics/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `ef10ceb` |
| [`tui-dense-finops-grid`](tui-dense-finops-grid/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `f720e57` |
| [`verifiable-work-receipts`](verifiable-work-receipts/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `ebc550a` |
| [`webmcp-browser-tools`](webmcp-browser-tools/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `02f3ef6` |
| [`charting-analytics`](charting-analytics/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `a059d01` |
| [`community-covenant-neutrality`](community-covenant-neutrality/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `9a82384` |
| [`trust-ladder-and-succession`](trust-ladder-and-succession/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `9fa6245` |
| [`cli-staged-deprecation-and-upgrade`](cli-staged-deprecation-and-upgrade/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `2842f1a` |
| [`jsonl-scan-diagnostics`](jsonl-scan-diagnostics/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `17b6b59` |
| [`tui-shell-overview`](tui-shell-overview/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `85af7be` |
| [`purge-legacy-mit-releases`](purge-legacy-mit-releases/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `cee5b3b` |
| [`contributor-onboarding`](contributor-onboarding/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `3bd8d0a` |
| [`release-communication-and-changelog`](release-communication-and-changelog/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `7b0813c` |
| [`distribution-coverage-foundation`](distribution-coverage-foundation/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `80b3e4f` |

