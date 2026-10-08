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
- **Total Goals**: 83
- 🟡 **Ready to Execute (Pending)**: 16 (14 independent ⚡, 2 unblocked 🔗)
- ⛔ **Blocked on Prerequisites**: 1
- 🟢 **Implemented**: 66 (1 verified 📜, 65 unverified)
- 🎯 **By Tier**: 1 immediate (Tier 1), 82 roadmap (Tier 2), 0 nice-to-have (Tier 3), 0 wont-fix (Tier 4)
- 📋 **Execution Tasks Progress**: 298/353 completed (84%)

---

## 🟡 Ready to Execute (Pending Goals)

| Goal Package | Tier | Mode & Sequence | Dependencies | Focus & Description | Launch Command |
| :--- | :---: | :---: | :--- | :--- | :--- |
| [`devin-cache-source-isolation`](devin-cache-source-isolation/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent (#65)` | - | Keep cached Devin totals isolated by source database — Key Devin snapshots by canonical source identity and invalidate on relevant sour... | `/goal goals/devin-cache-source-isolation/goal.md` |
| [`chatgpt-rolling-burst-replay`](chatgpt-rolling-burst-replay/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent (#72)` | - | ChatGPT Rolling 5-Hour Burst Capacity Replay — Enhance thermal replay to model dynamic 5-hour rolling message and token burst l... | `/goal goals/chatgpt-rolling-burst-replay/goal.md` |
| [`codex-sandboxed-receipts`](codex-sandboxed-receipts/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent (#73)` | - | Codex Sandboxed Execution Verification Receipts — Extract local test runner and linter execution commands and exit codes from Code... | `/goal goals/codex-sandboxed-receipts/goal.md` |
| [`markdown-pr-receipts`](markdown-pr-receipts/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent (#74)` | - | Zero-Leak Markdown PR Contribution Receipts — Add --format md (and --format markdown) to thermal receipt and thermal yield to ... | `/goal goals/markdown-pr-receipts/goal.md` |
| [`session-fork-tree-analytics`](session-fork-tree-analytics/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent (#75)` | - | Session Fork & Branch Tree Yield Analytics — Ingest parent-child thread relationships and session forks from Codex (and branc... | `/goal goals/session-fork-tree-analytics/goal.md` |
| [`terminal-statusline-helper`](terminal-statusline-helper/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent (#76)` | - | Instant Terminal Statusline & Prompt Composable Helper — Introduce an ultra-fast (<3ms) thermal statusline (or thermal prompt) command de... | `/goal goals/terminal-statusline-helper/goal.md` |
| [`agent-readiness-audit`](agent-readiness-audit/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent (#77)` | - | Repository Agent Readiness and AGENTS.md Audit — Extend thermal audit with an agent-readiness check verifying local repository in... | `/goal goals/agent-readiness-audit/goal.md` |
| [`cross-platform-ci-matrix`](cross-platform-ci-matrix/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent` | - | Cross-Platform Operating System Test Matrix in CI — Add macOS and Windows test runners to GitHub Actions CI to verify platform behav... | `/goal goals/cross-platform-ci-matrix/goal.md` |
| [`docs-parity-and-human-simulation`](docs-parity-and-human-simulation/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent` | - | Comprehensive Documentation Parity, Full Scan & Simulated Human User Audit — Update all repository Markdown documentation (README.md, AGENTS.md, docs/*.md) a... | `/goal goals/docs-parity-and-human-simulation/goal.md` |
| [`multi-target-cross-compile-ci-gate`](multi-target-cross-compile-ci-gate/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent` | - | Multi-Target Cross-Compilation Verification Gate in CI — Add multi-target cross-compilation check (darwin/arm64, darwin/amd64, windows/am... | `/goal goals/multi-target-cross-compile-ci-gate/goal.md` |
| [`platform-native-data-directories`](platform-native-data-directories/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent` | - | Platform-Native Application Data Directory Resolution — Add Windows APPDATA/LOCALAPPDATA and macOS Library/Application Support directory... | `/goal goals/platform-native-data-directories/goal.md` |
| [`purge-bot-contributors`](purge-bot-contributors/goal.md) | `🎯 Immediate` | `🚢 SHIP ⚡ Independent` | - | Purge Devin Bot Contributor Attribution & Git History Rewrite — Eliminate devin-ai-integration[bot] from GitHub's contributor graph. Audit and r... | `/goal goals/purge-bot-contributors/goal.md` |
| [`sila-onboarding`](sila-onboarding/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent` | - | Complete Sila Onboarding — Wire thermal-streak fully into sila and verify it: map the legacy task surfaces ... | `/goal goals/sila-onboarding/goal.md` |
| [`windows-sqlite-uri-path-normalization`](windows-sqlite-uri-path-normalization/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⚡ Independent` | - | Windows SQLite URI and Path Normalization — Add shared SQLite connection URI helper in internal/loaders to clean Windows dri... | `/goal goals/windows-sqlite-uri-path-normalization/goal.md` |
| [`live-metric-parity`](live-metric-parity/goal.md) | `🗺️ Roadmap` | `🚢 SHIP 🔗 Ready (#64)` | Deps met: web-telemetry-parity, live-polling-cpu-budget | Make live burn metrics reflect recorded tokens and turns — Exclude activity-only counts from token burn and use actual turn and cache-read ... | `/goal goals/live-metric-parity/goal.md` |
| [`localhost-request-boundary`](localhost-request-boundary/goal.md) | `🗺️ Roadmap` | `🚢 SHIP 🔗 Ready (#66)` | Deps met: web-telemetry-parity | Reject untrusted hosts at the local telemetry API — Validate request authority before loading private local telemetry while preservi... | `/goal goals/localhost-request-boundary/goal.md` |
| [`hermetic-verification-gates`](hermetic-verification-gates/goal.md) | `🗺️ Roadmap` | `🚢 SHIP ⛔ Blocked (#67)` | Prereqs: live-metric-parity, devin-cache-source-isolation, localhost-request-boundary | Make verification independent of host data and color settings — Use isolated fixtures and explicit color environments, and assert metric semanti... | `/goal goals/hermetic-verification-gates/goal.md` |

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
| [`receipt-evidence-integrity`](receipt-evidence-integrity/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `7a8ce91` |
| [`web-telemetry-parity`](web-telemetry-parity/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `d2cd4a2` |
| [`ingestion-performance-baseline`](ingestion-performance-baseline/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `e2a569f` |
| [`reasoning-effort-telemetry`](reasoning-effort-telemetry/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `0f3f533` |
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
| [`database-loader-iteration-safety`](database-loader-iteration-safety/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `c53c32b` |
| [`dns-aid-discovery`](dns-aid-discovery/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `769b9d6` |
| [`dsh-loader`](dsh-loader/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `9be2513` |
| [`embedded-web-dashboard`](embedded-web-dashboard/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `dceba19` |
| [`hermes-loader`](hermes-loader/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `ad2d523` |
| [`leaderboard-slice-preallocation`](leaderboard-slice-preallocation/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `f375fdf` |
| [`license-strategy-and-enforcement`](license-strategy-and-enforcement/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `98772b5` |
| [`loader-bounded-concurrency`](loader-bounded-concurrency/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `0c4fffb` |
| [`markdown-for-agents`](markdown-for-agents/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `3de20b6` |
| [`mcp-server-card`](mcp-server-card/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `f7350bf` |
| [`model-identity-and-alias-resolution`](model-identity-and-alias-resolution/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `e460d4a` |
| [`report-period-totals`](report-period-totals/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `b8fac89` |
| [`rfc9727-api-catalog`](rfc9727-api-catalog/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `e657745` |
| [`robots-content-signals`](robots-content-signals/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `fd5ca83` |
| [`setup-health-audit`](setup-health-audit/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `8afd798` |
| [`slsa3-release-provenance`](slsa3-release-provenance/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `78b8258` |
| [`stateless-telemetry-sharing`](stateless-telemetry-sharing/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `7557bf9` |
| [`streak-parsing-defensive-guards`](streak-parsing-defensive-guards/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `a2467d7` |
| [`telemetry-and-ui-audit`](telemetry-and-ui-audit/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `8c9847d` |
| [`token-yield-metrics`](token-yield-metrics/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `ef10ceb` |
| [`tui-dense-finops-grid`](tui-dense-finops-grid/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `f720e57` |
| [`verifiable-work-receipts`](verifiable-work-receipts/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `ebc550a` |
| [`webmcp-browser-tools`](webmcp-browser-tools/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `02f3ef6` |
| [`workload-replay`](workload-replay/goal.md) | `⚡ Indep` | 🚢 SHIP | unverified | Git Commit | `1ca92ed` |
| [`zed-editor-loader`](zed-editor-loader/goal.md) | `⚡ Indep` | 🚢 SHIP | verified 📜 | Git Commit | `aa81edd` |
| [`charting-analytics`](charting-analytics/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `a059d01` |
| [`community-covenant-neutrality`](community-covenant-neutrality/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `9a82384` |
| [`trust-ladder-and-succession`](trust-ladder-and-succession/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `9fa6245` |
| [`cli-staged-deprecation-and-upgrade`](cli-staged-deprecation-and-upgrade/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `2842f1a` |
| [`pricing-catalog-coverage`](pricing-catalog-coverage/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `79bbd75` |
| [`jsonl-scan-diagnostics`](jsonl-scan-diagnostics/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `17b6b59` |
| [`receipt-accounting-coverage`](receipt-accounting-coverage/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `dd3ab4f` |
| [`codex-incremental-rollout-cache`](codex-incremental-rollout-cache/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `752569e` |
| [`tui-shell-overview`](tui-shell-overview/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `85af7be` |
| [`purge-legacy-mit-releases`](purge-legacy-mit-releases/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `cee5b3b` |
| [`contributor-onboarding`](contributor-onboarding/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `3bd8d0a` |
| [`tui-projects-view`](tui-projects-view/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `50faef3` |
| [`tui-mix-models-views`](tui-mix-models-views/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `43c146e` |
| [`tui-stats-view`](tui-stats-view/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `be1e45f` |
| [`release-communication-and-changelog`](release-communication-and-changelog/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `7b0813c` |
| [`charting-flags-docs-release`](charting-flags-docs-release/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `25163db` |
| [`distribution-coverage-foundation`](distribution-coverage-foundation/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `80b3e4f` |
| [`html-docs-features-and-analytics`](html-docs-features-and-analytics/goal.md) | `🔗 Dep` | 🚢 SHIP | unverified | Git Commit | `de17df6` |
