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
- **Total Goals**: 89
- 🟡 **Ready to Execute (Pending)**: 4 (4 independent ⚡, 0 unblocked 🔗)
- 🟢 **Implemented**: 85 (2 verified 📜, 15 stale ⚠️, 68 unverified)
- 🎯 **By Tier**: 1 immediate (Tier 1), 88 roadmap (Tier 2), 0 nice-to-have (Tier 3), 0 wont-fix (Tier 4)
- 📋 **Execution Tasks Progress**: 372/391 completed (95%)

---

## 🟡 Ready to Execute (Pending Goals)

| Goal Package | Tier | Mode & Sequence | Dependencies | Focus & Description | Launch Command |
| :--- | :---: | :---: | :--- | :--- | :--- |
| [`directory-scan-incremental-cache`](directory-scan-incremental-cache/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent (#107)` | - | Reuse unchanged directory transcripts across invocations — Close the 1,000-file warm-latency gap measured in [`docs/PERFORMANCE.md`](../../... | `/goal goals/directory-scan-incremental-cache/goal.md` |
| [`replay-against-plan-resolution`](replay-against-plan-resolution/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent (#109)` | - | Resolve plan identifiers passed to replay --against — `thermal replay --against claude-pro` silently produces a pay-as-you-go row name... | `/goal goals/replay-against-plan-resolution/goal.md` |
| [`docs-page-claim-parity`](docs-page-claim-parity/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent (#110)` | - | Align hosted documentation claims with the promises ledger — The hosted documentation page still carries the blanket promise: > …computing ... | `/goal goals/docs-page-claim-parity/goal.md` |
| [`release-provenance-verification`](release-provenance-verification/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent (#111)` | - | Independently verify a released artifact's SLSA provenance — [`docs/PROMISES.md`](../../docs/PROMISES.md) records release provenance as **unv... | `/goal goals/release-provenance-verification/goal.md` |

---

## 🟢 Implemented

| Goal Package | Mode | Shape | Verification | Task / Ref | Commit |
| :--- | :---: | :--- | :---: | :--- | :--- |
| [`charting-dep-spike`](charting-dep-spike/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `d1b68ab` |
| [`governance-foundations`](governance-foundations/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `332ad26` |
| [`tiered-precommit-fastpath`](tiered-precommit-fastpath/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `1b5a3c8` |
| [`repo-alignment-and-security`](repo-alignment-and-security/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `bee76ff` |
| [`version-migration-guide`](version-migration-guide/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `12d9833` |
| [`devin-model-attribution`](devin-model-attribution/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `ecc9a92` |
| [`unpriceable-token-sources`](unpriceable-token-sources/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `09eb367` |
| [`jsonl-line-ceilings`](jsonl-line-ceilings/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `86d6ebb` |
| [`live-polling-cpu-budget`](live-polling-cpu-budget/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `b71d464` |
| [`web-telemetry-parity`](web-telemetry-parity/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `d2cd4a2` |
| [`devin-cache-source-isolation`](devin-cache-source-isolation/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `d182cf9` |
| [`reasoning-effort-telemetry`](reasoning-effort-telemetry/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `0f3f533` |
| [`codex-sandboxed-receipts`](codex-sandboxed-receipts/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `198772e` |
| [`markdown-pr-receipts`](markdown-pr-receipts/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `c00ef02` |
| [`session-fork-tree-analytics`](session-fork-tree-analytics/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `56ec02b` |
| [`terminal-statusline-helper`](terminal-statusline-helper/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `26a20ae` |
| [`agent-readiness-audit`](agent-readiness-audit/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `259d1e0` |
| [`receipt-evidence-integrity`](receipt-evidence-integrity/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `6713335` |
| [`workload-replay`](workload-replay/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `81901a8` |
| [`setup-health-audit`](setup-health-audit/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `0d42b32` |
| [`ingestion-performance-baseline`](ingestion-performance-baseline/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `e2a569f` |
| [`test-env-isolation`](test-env-isolation/goal.md) | `⚡ Indep` | 🚢 SHIP | verified 📜 | Audit Receipt | `cd5600c565e5cad966373629c09044df68101b3e` |
| [`acp-commerce-discovery`](acp-commerce-discovery/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `c9ee829` |
| [`agent-link-headers`](agent-link-headers/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `cf43a37` |
| [`agent-skills-discovery`](agent-skills-discovery/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `e087bf4` |
| [`agpl-dual-licensing`](agpl-dual-licensing/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `eb9aeb4` |
| [`ard-capability-manifest`](ard-capability-manifest/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `dbbb856` |
| [`auth-md-agent-registration`](auth-md-agent-registration/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `8483316` |
| [`cache-efficiency-metrics`](cache-efficiency-metrics/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `fa4c8f9` |
| [`claude-message-dedupe`](claude-message-dedupe/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `d39ef40` |
| [`cli-daily-update-check`](cli-daily-update-check/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `2c387fb` |
| [`cli-upgrade-file-integrity`](cli-upgrade-file-integrity/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `15d2ffb` |
| [`cross-platform-ci-matrix`](cross-platform-ci-matrix/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `70143cb` |
| [`database-loader-iteration-safety`](database-loader-iteration-safety/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `c53c32b` |
| [`dns-aid-discovery`](dns-aid-discovery/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `769b9d6` |
| [`docs-parity-and-human-simulation`](docs-parity-and-human-simulation/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `135b22e` |
| [`dsh-loader`](dsh-loader/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `9be2513` |
| [`embedded-web-dashboard`](embedded-web-dashboard/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `dceba19` |
| [`hermes-loader`](hermes-loader/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `ad2d523` |
| [`leaderboard-slice-preallocation`](leaderboard-slice-preallocation/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `f375fdf` |
| [`license-strategy-and-enforcement`](license-strategy-and-enforcement/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `98772b5` |
| [`loader-bounded-concurrency`](loader-bounded-concurrency/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `0c4fffb` |
| [`markdown-for-agents`](markdown-for-agents/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `3de20b6` |
| [`mcp-server-card`](mcp-server-card/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `f7350bf` |
| [`model-identity-and-alias-resolution`](model-identity-and-alias-resolution/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `e460d4a` |
| [`multi-target-cross-compile-ci-gate`](multi-target-cross-compile-ci-gate/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `0efc97b` |
| [`platform-native-data-directories`](platform-native-data-directories/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `598c3d1` |
| [`purge-bot-contributors`](purge-bot-contributors/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `cebda93` |
| [`report-period-totals`](report-period-totals/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `b8fac89` |
| [`rfc9727-api-catalog`](rfc9727-api-catalog/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `e657745` |
| [`robots-content-signals`](robots-content-signals/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `fd5ca83` |
| [`sila-onboarding`](sila-onboarding/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `8e3d3e9` |
| [`slsa3-release-provenance`](slsa3-release-provenance/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `78b8258` |
| [`stateless-telemetry-sharing`](stateless-telemetry-sharing/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `7557bf9` |
| [`streak-parsing-defensive-guards`](streak-parsing-defensive-guards/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `a2467d7` |
| [`telemetry-and-ui-audit`](telemetry-and-ui-audit/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `8c9847d` |
| [`token-yield-metrics`](token-yield-metrics/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `ef10ceb` |
| [`tui-dense-finops-grid`](tui-dense-finops-grid/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `f720e57` |
| [`verifiable-work-receipts`](verifiable-work-receipts/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `ebc550a` |
| [`webmcp-browser-tools`](webmcp-browser-tools/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `02f3ef6` |
| [`windows-sqlite-uri-path-normalization`](windows-sqlite-uri-path-normalization/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `e30bd4d` |
| [`zed-editor-loader`](zed-editor-loader/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `9b6a3ef` |
| [`charting-analytics`](charting-analytics/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `a059d01` |
| [`community-covenant-neutrality`](community-covenant-neutrality/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `9a82384` |
| [`trust-ladder-and-succession`](trust-ladder-and-succession/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `9fa6245` |
| [`cli-staged-deprecation-and-upgrade`](cli-staged-deprecation-and-upgrade/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `2842f1a` |
| [`pricing-catalog-coverage`](pricing-catalog-coverage/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `79bbd75` |
| [`jsonl-scan-diagnostics`](jsonl-scan-diagnostics/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `17b6b59` |
| [`receipt-accounting-coverage`](receipt-accounting-coverage/goal.md) | `🔗 Dep` | 🚢 SHIP | stale ⚠️ | Git Commit | `dd3ab4f` |
| [`live-metric-parity`](live-metric-parity/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `305f589` |
| [`localhost-request-boundary`](localhost-request-boundary/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `d162bf1` |
| [`codex-incremental-rollout-cache`](codex-incremental-rollout-cache/goal.md) | `🔗 Dep` | 🚢 SHIP | stale ⚠️ | Git Commit | `752569e` |
| [`chatgpt-rolling-burst-replay`](chatgpt-rolling-burst-replay/goal.md) | `🔗 Dep` | 🚢 SHIP | stale ⚠️ | Git Commit | `8c7402c` |
| [`tui-shell-overview`](tui-shell-overview/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `85af7be` |
| [`purge-legacy-mit-releases`](purge-legacy-mit-releases/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `cee5b3b` |
| [`contributor-onboarding`](contributor-onboarding/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `3bd8d0a` |
| [`hermetic-verification-gates`](hermetic-verification-gates/goal.md) | `🔗 Dep` | 🚢 SHIP | stale ⚠️ | Git Commit | `1f71cd2` |
| [`promise-claim-evidence`](promise-claim-evidence/goal.md) | `🔗 Dep` | 🚢 SHIP | verified 📜 | Audit Receipt | `c610c97fc84bf0d0b18416a5c137f01b758757db` |
| [`tui-projects-view`](tui-projects-view/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `50faef3` |
| [`tui-mix-models-views`](tui-mix-models-views/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `43c146e` |
| [`tui-stats-view`](tui-stats-view/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `be1e45f` |
| [`release-communication-and-changelog`](release-communication-and-changelog/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `7b0813c` |
| [`charting-flags-docs-release`](charting-flags-docs-release/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `25163db` |
| [`distribution-coverage-foundation`](distribution-coverage-foundation/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `80b3e4f` |
| [`html-docs-features-and-analytics`](html-docs-features-and-analytics/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `de17df6` |
