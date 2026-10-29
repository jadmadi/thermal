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
- **Total Goals**: 134
- 🟡 **Ready to Execute (Pending)**: 2 (2 independent ⚡, 0 unblocked 🔗)
- 🟢 **Implemented**: 132 (15 verified 📜, 29 stale ⚠️, 88 unverified)
- 🎯 **By Tier**: 1 immediate (Tier 1), 133 roadmap (Tier 2), 0 nice-to-have (Tier 3), 0 wont-fix (Tier 4)
- 📋 **Execution Tasks Progress**: 597/608 completed (98%)

---

## 🟡 Ready to Execute (Pending Goals)

| Goal Package | Tier | Mode & Sequence | Dependencies | Focus & Description | Launch Command |
| :--- | :---: | :---: | :--- | :--- | :--- |
| [`cli-report-total-row-semantics`](cli-report-total-row-semantics/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent (#252)` | - | Make report total rows contain only additive, dimension-consistent values — Priority: P2 (output semantics). The projects and models reports reuse the row s... | `/goal goals/cli-report-total-row-semantics/goal.md` |
| [`cli-analytics-and-audit-prose-clarity`](cli-analytics-and-audit-prose-clarity/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent (#253)` | - | Name denominators and intervals and remove run-on or zero-noise report prose — Priority: P2 (output clarity). The analytics commands are numerically correct bu... | `/goal goals/cli-analytics-and-audit-prose-clarity/goal.md` |

---

## 🟢 Implemented

| Goal Package | Mode | Shape | Verification | Task / Ref | Commit |
| :--- | :---: | :--- | :---: | :--- | :--- |
| [`charting-dep-spike`](charting-dep-spike/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `31fdf3e` |
| [`governance-foundations`](governance-foundations/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `644303c` |
| [`tiered-precommit-fastpath`](tiered-precommit-fastpath/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `e6fc9a8` |
| [`repo-alignment-and-security`](repo-alignment-and-security/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `4ce10ff` |
| [`version-migration-guide`](version-migration-guide/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `e64b8f0` |
| [`devin-model-attribution`](devin-model-attribution/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `e1c75b1` |
| [`unpriceable-token-sources`](unpriceable-token-sources/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `42bb6bb` |
| [`jsonl-line-ceilings`](jsonl-line-ceilings/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `f6d44cc` |
| [`live-polling-cpu-budget`](live-polling-cpu-budget/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `76e442f` |
| [`web-telemetry-parity`](web-telemetry-parity/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `6251e5d` |
| [`devin-cache-source-isolation`](devin-cache-source-isolation/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `e120350` |
| [`reasoning-effort-telemetry`](reasoning-effort-telemetry/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `4da74c0` |
| [`codex-sandboxed-receipts`](codex-sandboxed-receipts/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `aabfe19` |
| [`markdown-pr-receipts`](markdown-pr-receipts/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `50eccc4` |
| [`session-fork-tree-analytics`](session-fork-tree-analytics/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `2d2b582` |
| [`terminal-statusline-helper`](terminal-statusline-helper/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `a0123ca` |
| [`agent-readiness-audit`](agent-readiness-audit/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `695d71c` |
| [`live-baseline-separation`](live-baseline-separation/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `2d66157` |
| [`receipt-evidence-integrity`](receipt-evidence-integrity/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `fe2efdc` |
| [`workload-replay`](workload-replay/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `9a58fb7` |
| [`setup-health-audit`](setup-health-audit/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `b56e2e1` |
| [`ingestion-performance-baseline`](ingestion-performance-baseline/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `0086194` |
| [`directory-scan-incremental-cache`](directory-scan-incremental-cache/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `b252697` |
| [`test-env-isolation`](test-env-isolation/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `3fbc0b6` |
| [`replay-against-plan-resolution`](replay-against-plan-resolution/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `e03a7f9` |
| [`docs-page-claim-parity`](docs-page-claim-parity/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `431ede5` |
| [`release-provenance-verification`](release-provenance-verification/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `83c8251` |
| [`output-formatting-primitives`](output-formatting-primitives/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `06568f3` |
| [`help-supported-tools-and-flags`](help-supported-tools-and-flags/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `630ccfa` |
| [`heatmap-month-label-overlap`](heatmap-month-label-overlap/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `60cb42a` |
| [`sqlite-engine-security-upgrade`](sqlite-engine-security-upgrade/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `a7034d1` |
| [`sqlite-schema-substitution-guard`](sqlite-schema-substitution-guard/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `a7034d1` |
| [`muse-token-and-project-ingestion`](muse-token-and-project-ingestion/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `a7034d1` |
| [`changelog-integrity-and-cwd-isolation`](changelog-integrity-and-cwd-isolation/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `ad8009c` |
| [`cli-help-interception-and-scoped-help`](cli-help-interception-and-scoped-help/goal.md) | `⚡ Indep` | 🚢 SHIP | verified 📜 | Git Commit | `e86a107` |
| [`cli-upgrade-nondestructive-safety-flags`](cli-upgrade-nondestructive-safety-flags/goal.md) | `⚡ Indep` | 🚢 SHIP | verified 📜 | Git Commit | `c0e4087` |
| [`cli-argument-parser-strictness`](cli-argument-parser-strictness/goal.md) | `⚡ Indep` | 🚢 SHIP | verified 📜 | Git Commit | `8ebde49` |
| [`cli-db-override-scope-and-errors`](cli-db-override-scope-and-errors/goal.md) | `⚡ Indep` | 🚢 SHIP | verified 📜 | Git Commit | `4d97260` |
| [`cli-first-run-source-inventory`](cli-first-run-source-inventory/goal.md) | `⚡ Indep` | 🚢 SHIP | verified 📜 | Git Commit | `964cf9e` |
| [`cli-nontty-and-statusline-output-contract`](cli-nontty-and-statusline-output-contract/goal.md) | `⚡ Indep` | 🚢 SHIP | verified 📜 | Git Commit | `a352e43` |
| [`cli-terminal-width-adaptation`](cli-terminal-width-adaptation/goal.md) | `⚡ Indep` | 🚢 SHIP | verified 📜 | Git Commit | `a936195` |
| [`cli-leaderboard-ranking-and-legend`](cli-leaderboard-ranking-and-legend/goal.md) | `⚡ Indep` | 🚢 SHIP | verified 📜 | Git Commit | `17ab978` |
| [`cli-version-verbose-flag-contract`](cli-version-verbose-flag-contract/goal.md) | `⚡ Indep` | 🚢 SHIP | verified 📜 | Git Commit | `743281a` |
| [`acp-commerce-discovery`](acp-commerce-discovery/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `0c83f16` |
| [`agent-link-headers`](agent-link-headers/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `7dd5f9a` |
| [`agent-skills-discovery`](agent-skills-discovery/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `383ec34` |
| [`agpl-dual-licensing`](agpl-dual-licensing/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `41caac9` |
| [`ard-capability-manifest`](ard-capability-manifest/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `980ca1f` |
| [`auth-md-agent-registration`](auth-md-agent-registration/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `d203c1c` |
| [`cache-efficiency-metrics`](cache-efficiency-metrics/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `d6d8235` |
| [`claude-message-dedupe`](claude-message-dedupe/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `9671a0b` |
| [`cli-daily-update-check`](cli-daily-update-check/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `a53141b` |
| [`cli-upgrade-file-integrity`](cli-upgrade-file-integrity/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `2c13a20` |
| [`cross-platform-ci-matrix`](cross-platform-ci-matrix/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `972b37f` |
| [`database-loader-iteration-safety`](database-loader-iteration-safety/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `1cecbb5` |
| [`dns-aid-discovery`](dns-aid-discovery/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `4ce6090` |
| [`docs-parity-and-human-simulation`](docs-parity-and-human-simulation/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `6f7117f` |
| [`dsh-loader`](dsh-loader/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `8ed687f` |
| [`embedded-web-dashboard`](embedded-web-dashboard/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `4777f76` |
| [`hermes-loader`](hermes-loader/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `45060c2` |
| [`leaderboard-slice-preallocation`](leaderboard-slice-preallocation/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `2a2185a` |
| [`license-strategy-and-enforcement`](license-strategy-and-enforcement/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `278d1d3` |
| [`live-grid-turns-truncation`](live-grid-turns-truncation/goal.md) | `⚡ Indep` | 🚢 SHIP | verified 📜 | Audit Receipt | `70ad214cdaad7c65a1971f1ad57bcc8cf74e4adc` |
| [`live-speed-fractional-display`](live-speed-fractional-display/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `20461a4` |
| [`live-tracker-delta-clamping`](live-tracker-delta-clamping/goal.md) | `⚡ Indep` | 🚢 SHIP | verified 📜 | Audit Receipt | `70ad214cdaad7c65a1971f1ad57bcc8cf74e4adc` |
| [`live-tracker-midnight-drift`](live-tracker-midnight-drift/goal.md) | `⚡ Indep` | 🚢 SHIP | verified 📜 | Audit Receipt | `70ad214cdaad7c65a1971f1ad57bcc8cf74e4adc` |
| [`live-tracker-phantom-bursts`](live-tracker-phantom-bursts/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `6ab164c` |
| [`loader-bounded-concurrency`](loader-bounded-concurrency/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `33e9005` |
| [`markdown-for-agents`](markdown-for-agents/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `a6f415b` |
| [`mcp-server-card`](mcp-server-card/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `16d05c2` |
| [`model-identity-and-alias-resolution`](model-identity-and-alias-resolution/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `536d8a2` |
| [`multi-target-cross-compile-ci-gate`](multi-target-cross-compile-ci-gate/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `3b1279f` |
| [`platform-native-data-directories`](platform-native-data-directories/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `cdd1711` |
| [`purge-bot-contributors`](purge-bot-contributors/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `d3fc320` |
| [`report-period-totals`](report-period-totals/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `9eb30ce` |
| [`retire-dashboard-tui`](retire-dashboard-tui/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `6789fc6` |
| [`retire-token-yield`](retire-token-yield/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `eda8f0f` |
| [`retire-web-server`](retire-web-server/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `35f9fc6` |
| [`retire-work-receipts`](retire-work-receipts/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `4f085a9` |
| [`retire-workload-replay`](retire-workload-replay/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `f7b96f1` |
| [`rfc9727-api-catalog`](rfc9727-api-catalog/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `49a1d15` |
| [`robots-content-signals`](robots-content-signals/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `c71ec00` |
| [`sila-onboarding`](sila-onboarding/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `22b2530` |
| [`slsa3-release-provenance`](slsa3-release-provenance/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `47ba67d` |
| [`stateless-telemetry-sharing`](stateless-telemetry-sharing/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `de9bbdb` |
| [`streak-parsing-defensive-guards`](streak-parsing-defensive-guards/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `e3cf52c` |
| [`telemetry-and-ui-audit`](telemetry-and-ui-audit/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `97b6631` |
| [`token-yield-metrics`](token-yield-metrics/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `0b773e6` |
| [`tui-dense-finops-grid`](tui-dense-finops-grid/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `8fe7e6d` |
| [`verifiable-work-receipts`](verifiable-work-receipts/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `ea09db0` |
| [`webmcp-browser-tools`](webmcp-browser-tools/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `2ec434c` |
| [`windows-sqlite-uri-path-normalization`](windows-sqlite-uri-path-normalization/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `6a8fa87` |
| [`zed-editor-loader`](zed-editor-loader/goal.md) | `⚡ Indep` | 🚢 SHIP | stale ⚠️ | Git Commit | `e59d54b` |
| [`charting-analytics`](charting-analytics/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `bec2a9d` |
| [`community-covenant-neutrality`](community-covenant-neutrality/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `079ff1f` |
| [`trust-ladder-and-succession`](trust-ladder-and-succession/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `6679528` |
| [`cli-staged-deprecation-and-upgrade`](cli-staged-deprecation-and-upgrade/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `311e1a1` |
| [`pricing-catalog-coverage`](pricing-catalog-coverage/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `cce1ebe` |
| [`jsonl-scan-diagnostics`](jsonl-scan-diagnostics/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `7d84d9a` |
| [`receipt-accounting-coverage`](receipt-accounting-coverage/goal.md) | `🔗 Dep` | 🚢 SHIP | stale ⚠️ | Git Commit | `fe4d2ba` |
| [`live-metric-parity`](live-metric-parity/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `a655e01` |
| [`localhost-request-boundary`](localhost-request-boundary/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `ad78cfe` |
| [`codex-incremental-rollout-cache`](codex-incremental-rollout-cache/goal.md) | `🔗 Dep` | 🚢 SHIP | stale ⚠️ | Git Commit | `ef07150` |
| [`live-second-bucket-accounting`](live-second-bucket-accounting/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `72f9e10` |
| [`chatgpt-rolling-burst-replay`](chatgpt-rolling-burst-replay/goal.md) | `🔗 Dep` | 🚢 SHIP | stale ⚠️ | Git Commit | `21ba55f` |
| [`live-ticker-unit-and-cost-clarity`](live-ticker-unit-and-cost-clarity/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `2d66157` |
| [`output-report-totals-reconciliation`](output-report-totals-reconciliation/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `a279a5f` |
| [`chart-scale-order-and-top-contract`](chart-scale-order-and-top-contract/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `0cfdd5f` |
| [`audit-path-and-score-clarity`](audit-path-and-score-clarity/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `9ae5494` |
| [`sqlite-loader-resource-bounds`](sqlite-loader-resource-bounds/goal.md) | `🔗 Dep` | 🚢 SHIP | stale ⚠️ | Git Commit | `a7034d1` |
| [`pricing-catalog-and-partial-model-provenance`](pricing-catalog-and-partial-model-provenance/goal.md) | `🔗 Dep` | 🚢 SHIP | stale ⚠️ | Git Commit | `a7034d1` |
| [`cli-unknown-token-diagnostics`](cli-unknown-token-diagnostics/goal.md) | `🔗 Dep` | 🚢 SHIP | verified 📜 | Git Commit | `9903ca2` |
| [`cli-top-flag-contract`](cli-top-flag-contract/goal.md) | `🔗 Dep` | 🚢 SHIP | verified 📜 | Git Commit | `4dbf119` |
| [`tui-shell-overview`](tui-shell-overview/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `4d0f81a` |
| [`purge-legacy-mit-releases`](purge-legacy-mit-releases/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `4925b76` |
| [`contributor-onboarding`](contributor-onboarding/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `e0530fe` |
| [`hermetic-verification-gates`](hermetic-verification-gates/goal.md) | `🔗 Dep` | 🚢 SHIP | stale ⚠️ | Git Commit | `b4f045d` |
| [`live-flame-scale-contract`](live-flame-scale-contract/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `36eccea` |
| [`live-ticker-rollup-and-cost-sort`](live-ticker-rollup-and-cost-sort/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `4c5dd1b` |
| [`promise-claim-evidence`](promise-claim-evidence/goal.md) | `🔗 Dep` | 🚢 SHIP | verified 📜 | Audit Receipt | `c610c97fc84bf0d0b18416a5c137f01b758757db` |
| [`json-output-contract-hardening`](json-output-contract-hardening/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `7fe97c8` |
| [`live-ticker-unit-and-rollup-contract`](live-ticker-unit-and-rollup-contract/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `3b7cb85` |
| [`cost-provenance-display-and-reconciliation`](cost-provenance-display-and-reconciliation/goal.md) | `🔗 Dep` | 🚢 SHIP | stale ⚠️ | Git Commit | `a7034d1` |
| [`tui-projects-view`](tui-projects-view/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `f662fef` |
| [`tui-mix-models-views`](tui-mix-models-views/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `c849d30` |
| [`tui-stats-view`](tui-stats-view/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `ed0143c` |
| [`release-communication-and-changelog`](release-communication-and-changelog/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `e4e7857` |
| [`statusline-and-share-scope-contract`](statusline-and-share-scope-contract/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `f0fc71e` |
| [`charting-flags-docs-release`](charting-flags-docs-release/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `c994d87` |
| [`distribution-coverage-foundation`](distribution-coverage-foundation/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `266557c` |
| [`output-gallery-regression-contract`](output-gallery-regression-contract/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `6920439` |
| [`html-docs-features-and-analytics`](html-docs-features-and-analytics/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `c1da729` |
