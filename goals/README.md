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
- **Total Goals**: 70
- 🟢 **Implemented & Verified**: 60
- 🟡 **Ready to Execute (Pending)**: 5 (5 independent ⚡, 0 unblocked 🔗)
- ⛔ **Blocked on Prerequisites**: 5
- 🎯 **By Tier**: 1 immediate (Tier 1), 69 roadmap (Tier 2), 0 nice-to-have (Tier 3), 0 wont-fix (Tier 4)
- 📋 **Execution Tasks Progress**: 258/307 completed (84%)

---

## 🟡 Ready to Execute (Pending Goals)

| Goal Package | Tier | Mode & Sequence | Dependencies | Focus & Description | Launch Command |
| :--- | :---: | :---: | :--- | :--- | :--- |
| [`live-polling-cpu-budget`](live-polling-cpu-budget/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent (#60)` | - | Keep the live monitor responsive without saturating CPU at 100ms — Bound expensive collection work independently of fast display ticks, reuse uncha... | `/goal goals/live-polling-cpu-budget/goal.md` |
| [`receipt-evidence-integrity`](receipt-evidence-integrity/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent (#61)` | - | Require observed command results for verified receipts — Prevent missing, mixed or mismatched command outcomes from becoming verified wor... | `/goal goals/receipt-evidence-integrity/goal.md` |
| [`web-telemetry-parity`](web-telemetry-parity/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent (#63)` | - | Make web totals match canonical reports — Reuse canonical token and cost aggregation in web telemetry and honor supported ... | `/goal goals/web-telemetry-parity/goal.md` |
| [`devin-cache-source-isolation`](devin-cache-source-isolation/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent (#65)` | - | Keep cached Devin totals isolated by source database — Key Devin snapshots by canonical source identity and invalidate on relevant sour... | `/goal goals/devin-cache-source-isolation/goal.md` |
| [`ingestion-performance-baseline`](ingestion-performance-baseline/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent (#68)` | - | Measure cold and warm CLI performance on representative fixtures — Establish reproducible end-to-end latency and allocation evidence for the docume... | `/goal goals/ingestion-performance-baseline/goal.md` |
| [`receipt-accounting-coverage`](receipt-accounting-coverage/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⛔ Blocked (#62)` | Prereqs: receipt-evidence-integrity | Preserve recorded usage and disclose receipt coverage — Remove invented activity tokens and prevent partial transcript coverage from sil... | `/goal goals/receipt-accounting-coverage/goal.md` |
| [`live-metric-parity`](live-metric-parity/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⛔ Blocked (#64)` | Prereqs: web-telemetry-parity, live-polling-cpu-budget | Make live burn metrics reflect recorded tokens and turns — Exclude activity-only counts from token burn and use actual turn and cache-read ... | `/goal goals/live-metric-parity/goal.md` |
| [`localhost-request-boundary`](localhost-request-boundary/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⛔ Blocked (#66)` | Prereqs: web-telemetry-parity | Reject untrusted hosts at the local telemetry API — Validate request authority before loading private local telemetry while preservi... | `/goal goals/localhost-request-boundary/goal.md` |
| [`codex-incremental-rollout-cache`](codex-incremental-rollout-cache/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⛔ Blocked (#69)` | Prereqs: ingestion-performance-baseline | Avoid rescanning unchanged Codex rollout transcripts — Add bounded source-aware rollout caching with exact parity against full scans an... | `/goal goals/codex-incremental-rollout-cache/goal.md` |
| [`hermetic-verification-gates`](hermetic-verification-gates/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⛔ Blocked (#67)` | Prereqs: receipt-accounting-coverage, live-metric-parity, devin-cache-source-isolation, localhost-request-boundary, live-polling-cpu-budget | Make verification independent of host data and color settings — Use isolated fixtures and explicit color environments, and assert metric semanti... | `/goal goals/hermetic-verification-gates/goal.md` |

---

## 🟢 Implemented & Verified

| Goal Package | Mode | Shape | Task / Ref | Commit |
| :--- | :---: | :--- | :--- | :--- |
| [`charting-dep-spike`](charting-dep-spike/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `d1b68ab` |
| [`governance-foundations`](governance-foundations/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `332ad26` |
| [`tiered-precommit-fastpath`](tiered-precommit-fastpath/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `1b5a3c8` |
| [`repo-alignment-and-security`](repo-alignment-and-security/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `bee76ff` |
| [`version-migration-guide`](version-migration-guide/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `12d9833` |
| [`devin-model-attribution`](devin-model-attribution/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `ecc9a92` |
| [`unpriceable-token-sources`](unpriceable-token-sources/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `09eb367` |
| [`jsonl-line-ceilings`](jsonl-line-ceilings/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `86d6ebb` |
| [`acp-commerce-discovery`](acp-commerce-discovery/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `c9ee829` |
| [`agent-link-headers`](agent-link-headers/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `cf43a37` |
| [`agent-skills-discovery`](agent-skills-discovery/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `e087bf4` |
| [`agpl-dual-licensing`](agpl-dual-licensing/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `eb9aeb4` |
| [`ard-capability-manifest`](ard-capability-manifest/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `dbbb856` |
| [`auth-md-agent-registration`](auth-md-agent-registration/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `8483316` |
| [`cache-efficiency-metrics`](cache-efficiency-metrics/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `fa4c8f9` |
| [`claude-message-dedupe`](claude-message-dedupe/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `d39ef40` |
| [`cli-daily-update-check`](cli-daily-update-check/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `2c387fb` |
| [`cli-upgrade-file-integrity`](cli-upgrade-file-integrity/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `15d2ffb` |
| [`database-loader-iteration-safety`](database-loader-iteration-safety/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `c53c32b` |
| [`dns-aid-discovery`](dns-aid-discovery/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `769b9d6` |
| [`docs-parity-and-human-simulation`](docs-parity-and-human-simulation/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `64d1837` |
| [`dsh-loader`](dsh-loader/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `9be2513` |
| [`embedded-web-dashboard`](embedded-web-dashboard/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `dceba19` |
| [`hermes-loader`](hermes-loader/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `ad2d523` |
| [`leaderboard-slice-preallocation`](leaderboard-slice-preallocation/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `f375fdf` |
| [`license-strategy-and-enforcement`](license-strategy-and-enforcement/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `98772b5` |
| [`loader-bounded-concurrency`](loader-bounded-concurrency/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `0c4fffb` |
| [`markdown-for-agents`](markdown-for-agents/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `3de20b6` |
| [`mcp-server-card`](mcp-server-card/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `f7350bf` |
| [`model-identity-and-alias-resolution`](model-identity-and-alias-resolution/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `e460d4a` |
| [`purge-bot-contributors`](purge-bot-contributors/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `ced4751` |
| [`report-period-totals`](report-period-totals/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `b8fac89` |
| [`rfc9727-api-catalog`](rfc9727-api-catalog/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `e657745` |
| [`robots-content-signals`](robots-content-signals/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `fd5ca83` |
| [`setup-health-audit`](setup-health-audit/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `8afd798` |
| [`sila-onboarding`](sila-onboarding/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `b9ccdf9` |
| [`slsa3-release-provenance`](slsa3-release-provenance/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `78b8258` |
| [`stateless-telemetry-sharing`](stateless-telemetry-sharing/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `7557bf9` |
| [`streak-parsing-defensive-guards`](streak-parsing-defensive-guards/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `a2467d7` |
| [`telemetry-and-ui-audit`](telemetry-and-ui-audit/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `8c9847d` |
| [`token-yield-metrics`](token-yield-metrics/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `ef10ceb` |
| [`tui-dense-finops-grid`](tui-dense-finops-grid/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `f720e57` |
| [`verifiable-work-receipts`](verifiable-work-receipts/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `ebc550a` |
| [`webmcp-browser-tools`](webmcp-browser-tools/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `02f3ef6` |
| [`workload-replay`](workload-replay/goal.md) | `⚡ Indep` | 🚢 SHIP | Git Commit | `1ca92ed` |
| [`charting-analytics`](charting-analytics/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `a059d01` |
| [`community-covenant-neutrality`](community-covenant-neutrality/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `9a82384` |
| [`trust-ladder-and-succession`](trust-ladder-and-succession/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `9fa6245` |
| [`cli-staged-deprecation-and-upgrade`](cli-staged-deprecation-and-upgrade/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `2842f1a` |
| [`pricing-catalog-coverage`](pricing-catalog-coverage/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `79bbd75` |
| [`jsonl-scan-diagnostics`](jsonl-scan-diagnostics/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `17b6b59` |
| [`tui-shell-overview`](tui-shell-overview/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `85af7be` |
| [`purge-legacy-mit-releases`](purge-legacy-mit-releases/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `cee5b3b` |
| [`contributor-onboarding`](contributor-onboarding/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `3bd8d0a` |
| [`tui-projects-view`](tui-projects-view/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `50faef3` |
| [`tui-mix-models-views`](tui-mix-models-views/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `43c146e` |
| [`tui-stats-view`](tui-stats-view/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `be1e45f` |
| [`release-communication-and-changelog`](release-communication-and-changelog/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `7b0813c` |
| [`charting-flags-docs-release`](charting-flags-docs-release/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `25163db` |
| [`distribution-coverage-foundation`](distribution-coverage-foundation/goal.md) | `🔗 Dep` | 🚢 SHIP | Git Commit | `80b3e4f` |

