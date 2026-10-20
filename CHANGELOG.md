# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.21.0](https://github.com/jadmadi/thermal/compare/v0.20.1...v0.21.0) (2026-10-20)


### Features

* **analytics:** shared chart analytics, distribution stats and trend projections (goals/charting-analytics/goal.md) ([bec2a9d](https://github.com/jadmadi/thermal/commit/bec2a9d6feb63aa1fed672a37612486c7dd1d261))
* **audit:** add repository AGENTS.md readiness and agent hygiene check (goals/agent-readiness-audit/goal.md) ([695d71c](https://github.com/jadmadi/thermal/commit/695d71c9d401a04a3942b7f47407169b92242733))
* **build:** verify multi-target cross-compilation release gate (goals/multi-target-cross-compile-ci-gate/goal.md) ([3b1279f](https://github.com/jadmadi/thermal/commit/3b1279f180a3c16f64a2d325fe9b567d7a26e40b))
* **changelog:** add interactive changelog command with top filtering and JSON export ([e06910b](https://github.com/jadmadi/thermal/commit/e06910b62a8ac3e6acfc231303d428a408e6e81d))
* **chart:** support --chart flag across reports, document dashboard and conventions (goals/charting-flags-docs-release/goal.md) ([c994d87](https://github.com/jadmadi/thermal/commit/c994d87d321b4ad500b250edff94fb9d6303f4de))
* **check:** introduce fast-path pre-commit mode with package tests & diff sentry (goals/tiered-precommit-fastpath/goal.md) ([e6fc9a8](https://github.com/jadmadi/thermal/commit/e6fc9a81595fed155e73e0d572a5eb58011c211b))
* **ci:** add fast-path bypass to pre-push gate for docs and metadata diffs ([8400b42](https://github.com/jadmadi/thermal/commit/8400b424667968980a58ba9ef6bdc4809c592c50))
* **ci:** add OpenSSF SLSA Level 3 build provenance and verification runbook (goals/slsa3-release-provenance/goal.md) ([47ba67d](https://github.com/jadmadi/thermal/commit/47ba67d322160d1b878afae8984ed46598cd9dbf))
* **ci:** verify cross-platform operating system test matrix (goals/cross-platform-ci-matrix/goal.md) ([972b37f](https://github.com/jadmadi/thermal/commit/972b37f3140cb14d948923b01690b4720e80805a))
* **cli:** add build.sh install mode with shadow verification ([df54345](https://github.com/jadmadi/thermal/commit/df5434509270a83b90d259bda835d712d48e2aff))
* **cli:** add daily cached background update check and upgrade notification (goals/cli-daily-update-check/goal.md) ([a53141b](https://github.com/jadmadi/thermal/commit/a53141bf4db4a6f3d7e6395fa2b862009a12e41c))
* **cli:** add instant shell statusline and prompt composable helper (goals/terminal-statusline-helper/goal.md) ([a0123ca](https://github.com/jadmadi/thermal/commit/a0123ca8d42708f296cf44e3d0491ccf94d481cd))
* **cli:** add staged deprecation warnings & upgrade migration alerts (goals/cli-staged-deprecation-and-upgrade/goal.md) ([311e1a1](https://github.com/jadmadi/thermal/commit/311e1a130497212928a5fa7ec2bde576fe5ac42c))
* **cli:** add stateless zero-database url sharing engine (goals/stateless-telemetry-sharing/goal.md) ([de9bbdb](https://github.com/jadmadi/thermal/commit/de9bbdb060580a1375544aa7b09af57f7a5957ae))
* **cli:** add thermal audit command for local setup and context health (goals/setup-health-audit/goal.md) ([a9409a6](https://github.com/jadmadi/thermal/commit/a9409a67a78cd1de15ccf42f0dc0656adfc594da))
* **cli:** add thermal receipt command for verifiable work outcomes (goals/verifiable-work-receipts/goal.md) ([ea09db0](https://github.com/jadmadi/thermal/commit/ea09db0280c25ef50e8db413b0e37a5b3e71aed7))
* **cli:** add thermal serve embedded local web dashboard and telemetry APIs (goals/embedded-web-dashboard/goal.md) ([4777f76](https://github.com/jadmadi/thermal/commit/4777f76aa1d6250e6120af17c08c2ebec3c63159))
* **cli:** add thermal yield command for code output delta telemetry (goals/token-yield-metrics/goal.md) ([0b773e6](https://github.com/jadmadi/thermal/commit/0b773e66714f87b70956e7b4baa9308712d5f3e2))
* **cli:** make help tool and flag reference exhaustive (goals/help-supported-tools-and-flags/goal.md) ([630ccfa](https://github.com/jadmadi/thermal/commit/630ccfa08c9e65324898f867d96b5f8c4a07fa53))
* **cli:** retire dashboard TUI and dense grid (goals/retire-dashboard-tui/goal.md) ([6789fc6](https://github.com/jadmadi/thermal/commit/6789fc6f93ef7f1de34aa74c78d68afb336fd01d))
* **cli:** retire embedded web server (goals/retire-web-server/goal.md) ([35f9fc6](https://github.com/jadmadi/thermal/commit/35f9fc6d1250b9d727268dfeb4d3ba0b0ae2ea91))
* **cli:** retire subscription replay simulation (goals/retire-workload-replay/goal.md) ([f7b96f1](https://github.com/jadmadi/thermal/commit/f7b96f19c9a3337397ab43624dc4b41133f9e34b))
* **cli:** retire token yield and diff parsing (goals/retire-token-yield/goal.md) ([eda8f0f](https://github.com/jadmadi/thermal/commit/eda8f0f6121140d5beeb80f4060eb86d631babb9))
* **cli:** retire transcript command receipts (goals/retire-work-receipts/goal.md) ([4f085a9](https://github.com/jadmadi/thermal/commit/4f085a9bb2eb9cb804bf5622bacb6f783f493de4))
* **contributing:** establish contributor onboarding path, preflight check & retention loop (goals/contributor-onboarding/goal.md) ([e0530fe](https://github.com/jadmadi/thermal/commit/e0530fe998b0177bf97efc31ca284d5568292d62))
* **devin:** project sessions.model onto DailyRow and ProjectDay for accurate model attribution (goals/devin-model-attribution/goal.md) ([e1c75b1](https://github.com/jadmadi/thermal/commit/e1c75b174df9c3a9f34a70100d7578f252e1414d))
* **discovery:** declare Content-Signal directives in robots.txt (goals/robots-content-signals/goal.md) ([c71ec00](https://github.com/jadmadi/thermal/commit/c71ec002520bdf76da731bb8b78e64fd7d8ce1b6))
* **discovery:** include RFC 8288 Link headers and discovery link relations (goals/agent-link-headers/goal.md) ([7dd5f9a](https://github.com/jadmadi/thermal/commit/7dd5f9aa9c5f9c9d1b945cabd84d6f10ea30f0a4))
* **discovery:** publish ACP commerce discovery document (goals/acp-commerce-discovery/goal.md) ([0c83f16](https://github.com/jadmadi/thermal/commit/0c83f16ca602b08cdd8cb62ed00d0d7ba214fd7f))
* **discovery:** publish agent skills discovery index per RFC v0.2.0 (goals/agent-skills-discovery/goal.md) ([383ec34](https://github.com/jadmadi/thermal/commit/383ec34dee57b5afd9dfd92aedf39c3ac24f8f8d))
* **discovery:** publish ARD capability manifest ai-catalog.json (goals/ard-capability-manifest/goal.md) ([980ca1f](https://github.com/jadmadi/thermal/commit/980ca1f51a4be26adae1c354b87309631acd8fca))
* **discovery:** publish Auth.md metadata for agent registration (goals/auth-md-agent-registration/goal.md) ([d203c1c](https://github.com/jadmadi/thermal/commit/d203c1c61f838ecd2f89877e57192ed73a83f9ad))
* **discovery:** publish DNS-AID zone records and DNSSEC documentation (goals/dns-aid-discovery/goal.md) ([4ce6090](https://github.com/jadmadi/thermal/commit/4ce609014120952ff5854aff9fbf2b0780f8c288))
* **discovery:** publish MCP server card for agent discovery (goals/mcp-server-card/goal.md) ([16d05c2](https://github.com/jadmadi/thermal/commit/16d05c227aa8ba1d033f49ec0e35ea000bbb4889))
* **discovery:** publish RFC 9727 API catalog for automated API discovery (goals/rfc9727-api-catalog/goal.md) ([49a1d15](https://github.com/jadmadi/thermal/commit/49a1d1545650bec3cd22968bc1f52903c35490fe))
* **discovery:** support Markdown for Agents with content negotiation (goals/markdown-for-agents/goal.md) ([a6f415b](https://github.com/jadmadi/thermal/commit/a6f415bc54e2e46a0ed7d8aa9ef8f58f7f0a67fe))
* **discovery:** support WebMCP browser tools in documentation homepage (goals/webmcp-browser-tools/goal.md) ([2ec434c](https://github.com/jadmadi/thermal/commit/2ec434cd7ac221ff344e5f201e88bef021145cad))
* **distribution:** configure homebrew tap, installer script, agent docs & distribution runbook (goals/distribution-coverage-foundation/goal.md) ([266557c](https://github.com/jadmadi/thermal/commit/266557c90381a3bed2fa869fbfc05afb6b58de88))
* **docs:** add analytics suite and charting showcase to GitHub Pages (goals/html-docs-features-and-analytics/goal.md) ([c1da729](https://github.com/jadmadi/thermal/commit/c1da729e0ebc439ed10ddbf20978c85e92ce3f3e))
* **docs:** establish canonical migration guide & governance gate (goals/version-migration-guide/goal.md) ([e64b8f0](https://github.com/jadmadi/thermal/commit/e64b8f024578cc664d28b24712ac794acaf04b33))
* **docs:** update documentation parity and simulated human user audit for 15 tools (goals/docs-parity-and-human-simulation/goal.md) ([6f7117f](https://github.com/jadmadi/thermal/commit/6f7117f6463129c5c15e4d8468be64f0a150b6d4))
* **docs:** update GitHub Pages licensing and governance parity (goals/html-docs-licensing-parity/goal.md) ([2f81dea](https://github.com/jadmadi/thermal/commit/2f81deaa439bd9ceb294b08710e9d8f5d5492277))
* **gate:** ensure hermetic verification and release gate semantic parity (goals/hermetic-verification-gates/goal.md) ([b4f045d](https://github.com/jadmadi/thermal/commit/b4f045d8eeac937e8891e4abb6e113914709a38e))
* **git:** verify zero bot contributor attribution in git history (goals/purge-bot-contributors/goal.md) ([d3fc320](https://github.com/jadmadi/thermal/commit/d3fc320f5d59dfb4c996358e9b31c990c75e7dce))
* **governance:** establish Written Solo Authority model and scope defense (goals/governance-foundations/goal.md) ([644303c](https://github.com/jadmadi/thermal/commit/644303c200f25b342cd7c384450c336e79bd7813))
* **governance:** implement staged trust ladder, succession runbook & absence factor checks (goals/trust-ladder-and-succession/goal.md) ([6679528](https://github.com/jadmadi/thermal/commit/6679528890ed3831835de966435d345f6c0e7890))
* **init:** initial release of Thermal under GNU AGPL-3.0 with commercial dual-licensing ([679902e](https://github.com/jadmadi/thermal/commit/679902e45f0febcf945457d4b0113a840bdb169d))
* **json:** standardize envelopes, aliases, precision, and share fields (goals/json-output-contract-hardening/goal.md) ([7fe97c8](https://github.com/jadmadi/thermal/commit/7fe97c87fc279c37590d4cfe9721691c5a5e5fbd))
* **license:** add Community Reciprocity Covenant and license CLI metadata (goals/community-covenant-neutrality/goal.md) ([079ff1f](https://github.com/jadmadi/thermal/commit/079ff1f5e8fb304a05b51446d8922b95100b5c07))
* **license:** AGPL-3.0 relicensing, commercial dual-licensing & contributor gating (goals/agpl-dual-licensing/goal.md) ([41caac9](https://github.com/jadmadi/thermal/commit/41caac9864f1f9678d27278eeed3e1483682fa20))
* **license:** harden licensing architecture, CI enforcement & SPDX headers (goals/license-strategy-and-enforcement/goal.md) ([278d1d3](https://github.com/jadmadi/thermal/commit/278d1d3af4449beb46fac2dedb027757a7fc7832))
* **license:** publish $30 Commercial Use License and fix public claims ([7965ff6](https://github.com/jadmadi/thermal/commit/7965ff63ccbd22a426cecc958f3fbc313dec5de7))
* **live:** account deltas in one-second buckets and unify Peak/Rate (goals/live-second-bucket-accounting/goal.md) ([72f9e10](https://github.com/jadmadi/thermal/commit/72f9e10aa7712bb1d662d4f5863ce7b78b278b39))
* **live:** attribute resumed session projects, seed today's baseline, and track activity units ([e2238b6](https://github.com/jadmadi/thermal/commit/e2238b67aa322b383f626c11ac1ce0c663ee59c6))
* **live:** define a metric-specific flame ceiling (goals/live-flame-scale-contract/goal.md) ([36eccea](https://github.com/jadmadi/thermal/commit/36eccea259655ce4dd65077c94d5f82ea98826a2))
* **live:** roll up burst rows and add cost-first sorting (goals/live-ticker-rollup-and-cost-sort/goal.md) ([4c5dd1b](https://github.com/jadmadi/thermal/commit/4c5dd1b871ca0d545dec5267f3e10161d1920d2e))
* **live:** separate baseline carry-in and normalize ticker cost/feed (goals/live-baseline-separation/goal.md) (goals/live-ticker-unit-and-cost-clarity/goal.md) ([2d66157](https://github.com/jadmadi/thermal/commit/2d66157291a7b14e8208894f7e846d9d4fe5de4e))
* **loaders:** add DeepSeek harness (dsh) loader and registry (goals/dsh-loader/goal.md) ([8ed687f](https://github.com/jadmadi/thermal/commit/8ed687f822af87b043aa13421ddd5d592488347a))
* **loaders:** add Nous Hermes agent ingestion loader for state.db (goals/hermes-loader/goal.md) ([45060c2](https://github.com/jadmadi/thermal/commit/45060c20ce06c948bc6109e040d3b368278ff266))
* **loaders:** add platform-native application data directory discovery (goals/platform-native-data-directories/goal.md) ([cdd1711](https://github.com/jadmadi/thermal/commit/cdd1711e5148d226055bff75388637796389ea83))
* **loaders:** ingest Zed editor assistant threads and tokens (goals/zed-editor-loader/goal.md) ([e59d54b](https://github.com/jadmadi/thermal/commit/e59d54b51b750fcf659e7bdfe7b92efaab1b7e76))
* **loaders:** rename command-code display name to CMD ([8eb4675](https://github.com/jadmadi/thermal/commit/8eb46757f47dd2e4f5ec207fbba57fce89c773b3))
* **loaders:** report command-code tokens, spend, and per-project usage ([e8b45c0](https://github.com/jadmadi/thermal/commit/e8b45c0feabbe4d31516797d8aee27186e890204))
* **loaders:** report truncated JSONL scans via Summary.Warnings to verbose stderr (goals/jsonl-scan-diagnostics/goal.md) ([7d84d9a](https://github.com/jadmadi/thermal/commit/7d84d9aca752d5cc0dbc44c62718eb7db3496fdd))
* **loaders:** support unified Antigravity hub and OpenCode model breakdowns ([b1831e5](https://github.com/jadmadi/thermal/commit/b1831e530503539d85c1f73108c3126201ae622d))
* **models:** add reasoning effort breakdown & cognitive intensity telemetry (goals/reasoning-effort-telemetry/goal.md) ([4da74c0](https://github.com/jadmadi/thermal/commit/4da74c0f42dc91049aeb599bb26891bb3cbbcd83))
* **models:** declare canonical model identities and router alias normalization (goals/model-identity-and-alias-resolution/goal.md) ([536d8a2](https://github.com/jadmadi/thermal/commit/536d8a283259c4a08584af6e7448696110516f6d))
* **pricing:** catalog coverage and tier suffix lookup for tool-native model variants (goals/pricing-catalog-coverage/goal.md) ([cce1ebe](https://github.com/jadmadi/thermal/commit/cce1ebe7f6b729539061b312b1b89655759eef0d))
* **pricing:** publish a live $30 commercial pricing page ([a2b7c20](https://github.com/jadmadi/thermal/commit/a2b7c206c3a127480f3f17db4f10ddb20a16af35))
* **receipt:** extract Codex sandboxed test runs and exit codes (goals/codex-sandboxed-receipts/goal.md) ([aabfe19](https://github.com/jadmadi/thermal/commit/aabfe19702422a48bd4729de6fa4f189b8a22fbb))
* **release:** adapt changelog-writing standards and configure release-please sections (goals/release-communication-and-changelog/goal.md) ([e4e7857](https://github.com/jadmadi/thermal/commit/e4e785718fb095e790bf39806648ac379b2938f8))
* **release:** script legacy release retirement & document AGPL-3.0 release line (goals/purge-legacy-mit-releases/goal.md) ([4925b76](https://github.com/jadmadi/thermal/commit/4925b76e04241fed5969bfda7cc2fc706de2b16b))
* **render:** add PR-ready markdown output for receipt and yield (goals/markdown-pr-receipts/goal.md) ([50eccc4](https://github.com/jadmadi/thermal/commit/50eccc484da9c4e381cfb50f284df0e273286882))
* **render:** establish unified terminal design tokens, adaptive cards, and text wrapping ([30d8e86](https://github.com/jadmadi/thermal/commit/30d8e86531ef3c462f1878f4d68bee124905d615))
* **render:** unify output formatting semantics (goals/output-formatting-primitives/goal.md) ([06568f3](https://github.com/jadmadi/thermal/commit/06568f3694f5e0855280b41d31365f501581cd91))
* **replay:** add ChatGPT 5-hour rolling burst modeling (goals/chatgpt-rolling-burst-replay/goal.md) ([bca3e9f](https://github.com/jadmadi/thermal/commit/bca3e9fce019b2331b5f1f7d48ca0b1a11aaf8a0))
* **replay:** simulate historical developer workloads against subscription limits and model cards (goals/workload-replay/goal.md) ([a4e463d](https://github.com/jadmadi/thermal/commit/a4e463d3ef07db9e51118f1eee3318c81e19c620))
* **repo:** establish code ownership, maintainers, security policy & PR lifecycle (goals/repo-alignment-and-security/goal.md) ([4ce10ff](https://github.com/jadmadi/thermal/commit/4ce10fffcc4e5a256c0a6ebca537c5a4ca881e75))
* **report:** reconcile totals, cost bases, and unclassified tokens (goals/output-report-totals-reconciliation/goal.md) ([42a3589](https://github.com/jadmadi/thermal/commit/42a3589bb16aaaa545c59453c9ce9730e73f11f8))
* **report:** verify weekly and monthly period totals and table formatting parity (goals/report-period-totals/goal.md) ([9eb30ce](https://github.com/jadmadi/thermal/commit/9eb30ce6f0427d0f15e70ba900b6cbe0631c42b0))
* **server:** craft dense operator dashboard with dark/light themes, inline sparklines & keyboard navigation ([36e1f74](https://github.com/jadmadi/thermal/commit/36e1f74dad3674e3e20a3336a698f283f8ca11c0))
* **server:** make web totals match canonical reports and support collection options (goals/web-telemetry-parity/goal.md) ([6251e5d](https://github.com/jadmadi/thermal/commit/6251e5d58db6413bca74946563ba31f7630e7826))
* **server:** reject untrusted host authorities and cross-site requests at telemetry API (goals/localhost-request-boundary/goal.md) ([ad78cfe](https://github.com/jadmadi/thermal/commit/ad78cfedf71110ead43cbc54a6b7803782346e06))
* **sila:** complete Sila onboarding and strict verification (goals/sila-onboarding/goal.md) ([22b2530](https://github.com/jadmadi/thermal/commit/22b2530b6d1f9c5e7535f3b25389c8916afed40d))
* **stats:** expose disjoint token composition and cache hit rate metrics (goals/cache-efficiency-metrics/goal.md) ([d6d8235](https://github.com/jadmadi/thermal/commit/d6d8235de53d5a82ee2a806e69369ed73b6f6075))
* **tui:** add high-density 9-box finops grid and activity taxonomy view (goals/tui-dense-finops-grid/goal.md) ([8fe7e6d](https://github.com/jadmadi/thermal/commit/8fe7e6df8ed6063e886d75c67f54a2129d0eeafa))
* **tui:** bubble tea v2 application shell and overview dashboard (goals/tui-shell-overview/goal.md) ([4d0f81a](https://github.com/jadmadi/thermal/commit/4d0f81a3c9e95de4a1bf8af4c6ffc0fcf16b5e74))
* **tui:** elevate codeburn-inspired 9-box FinOps grid & dashboard dense view ([baecc17](https://github.com/jadmadi/thermal/commit/baecc17d9f31d028d0c865eb8ed0a2eb1a16eb8e))
* **tui:** interactive projects ranking view and drill-down detail (goals/tui-projects-view/goal.md) ([f662fef](https://github.com/jadmadi/thermal/commit/f662fef1c7f15d474b73b9b7113a571a2485f821))
* **tui:** interactive statistical distribution, weekday cadence and outlier view (goals/tui-stats-view/goal.md) ([ed0143c](https://github.com/jadmadi/thermal/commit/ed0143c19740c050c51ae5ddd5768ad34a778cee))
* **tui:** interactive tool and model mix share views with switching metrics (goals/tui-mix-models-views/goal.md) ([c849d30](https://github.com/jadmadi/thermal/commit/c849d30e41d59f9e18ce978856e907306064bac7))
* **tui:** spike ntcharts v2 against stock charm stack and record findings (goals/charting-dep-spike/goal.md) ([31fdf3e](https://github.com/jadmadi/thermal/commit/31fdf3e4861e543401ff3b9bdb3bfc254cb914d0))
* **yield:** analyze session fork lineage and branch churn in code delta yield (goals/session-fork-tree-analytics/goal.md) ([2d2b582](https://github.com/jadmadi/thermal/commit/2d2b58265090c265ecda1321018fe32768816c2d))


### Bug Fixes

* **audit:** make findings addressable and score explainable (goals/audit-path-and-score-clarity/goal.md) ([9ae5494](https://github.com/jadmadi/thermal/commit/9ae5494c49874b6a8bd718a842c14fbf90ced177))
* **audit:** report database readability only and make cache failures visible (goals/setup-health-audit/goal.md) ([b56e2e1](https://github.com/jadmadi/thermal/commit/b56e2e1fc3989c7e8df2f3d87d96fd291d1ae80f))
* **claude:** deduplicate assistant messages by message id and report in summary warnings (goals/claude-message-dedupe/goal.md) ([9671a0b](https://github.com/jadmadi/thermal/commit/9671a0bdcde6f67422522a04f1381810e9f6f5bb))
* **cli:** align chart scale, order, top, and --chart applicability (goals/chart-scale-order-and-top-contract/goal.md) ([0cfdd5f](https://github.com/jadmadi/thermal/commit/0cfdd5febafb1fd64864327a44b56925cf08ffc6))
* **discovery:** remove hallucinated payment, OAuth, and openapi references ([b5fa6ee](https://github.com/jadmadi/thermal/commit/b5fa6eeaa275cd330b56c119ca1703169f93b2d4))
* **docs:** qualify hosted performance and audit claims against the promises ledger (goals/docs-page-claim-parity/goal.md) ([431ede5](https://github.com/jadmadi/thermal/commit/431ede5de239c030b9579cb7ac5d8be76a8ef04a))
* **docs:** verify the v0.20.0 release provenance with slsa-verifier (goals/release-provenance-verification/goal.md) ([83c8251](https://github.com/jadmadi/thermal/commit/83c8251c02a56568d080f61e73c190be93e3b83f))
* **live:** align token burn, turn counts, and cache rates with canonical metrics (goals/live-metric-parity/goal.md) ([a655e01](https://github.com/jadmadi/thermal/commit/a655e0196f94654cbba00bc83e9543693cc7c0b5))
* **live:** decouple display cadence and cache unchanged sources to bound 100ms CPU (goals/live-polling-cpu-budget/goal.md) ([76e442f](https://github.com/jadmadi/thermal/commit/76e442f5c7183ed836969c1946b51877d8f9001b))
* **live:** format fractional sub-1.0 speeds in live monitor (goals/live-speed-fractional-display/goal.md) ([20461a4](https://github.com/jadmadi/thermal/commit/20461a484fb133feba42d4dcde899069d8b2989d))
* **live:** make ticker units complete and model cost basis explicit (goals/live-ticker-unit-and-rollup-contract/goal.md) ([3b7cb85](https://github.com/jadmadi/thermal/commit/3b7cb85357601093f39033ffa4bdec12ab5a9e38))
* **live:** prevent idle project baseline collapse in live monitor (goals/live-tracker-phantom-bursts/goal.md) ([6ab164c](https://github.com/jadmadi/thermal/commit/6ab164c828210c12622f850b2bb142384ef831dd))
* **loaders:** enforce SQLite rows iteration error inspection and document standard (goals/database-loader-iteration-safety/goal.md) ([1cecbb5](https://github.com/jadmadi/thermal/commit/1cecbb583a3b06cac7cb965703f78f014c31a507))
* **loaders:** exclude sqlite shm files to prevent live monitor CPU thrashing ([55a32ec](https://github.com/jadmadi/thermal/commit/55a32ec65d30ef4831a61f3e9e9f3244baec8e85))
* **loaders:** ignore empty WAL and volatile SHM locks in Devin source identity ([e82a486](https://github.com/jadmadi/thermal/commit/e82a486f13901c7a2d02a016ef95feddd6072028))
* **loaders:** isolate Devin cache by source database and validate freshness (goals/devin-cache-source-isolation/goal.md) ([e120350](https://github.com/jadmadi/thermal/commit/e1203503dfddefd51dc215d1a57f1f5c89379d1b))
* **loaders:** normalize SQLite URI paths for cross-platform compatibility (goals/windows-sqlite-uri-path-normalization/goal.md) ([6a8fa87](https://github.com/jadmadi/thermal/commit/6a8fa87fd0ce2277657fabd772f3a097e7bc3c07))
* **loaders:** raise JSONL scanner line ceiling to 32 MiB for large transcripts (goals/jsonl-line-ceilings/goal.md) ([f6d44cc](https://github.com/jadmadi/thermal/commit/f6d44cc611bec3ebc42f81d37dfc123f3762beb6))
* **loaders:** support relative ANTIGRAVITY_APP_DATA_DIR, unified sibling folding & ~/.config/thermal/thermal.config paths ([544a3d5](https://github.com/jadmadi/thermal/commit/544a3d5808b316efd28a67c8a6d76c679155abb6))
* **receipt:** preserve recorded usage, deduplicate messages, and disclose partial coverage (goals/receipt-accounting-coverage/goal.md) ([fe4d2ba](https://github.com/jadmadi/thermal/commit/fe4d2baba70b81bf2706e253d557c6a6a30e9975))
* **receipt:** require observed command results and drop uncompleted invocations (goals/receipt-evidence-integrity/goal.md) ([478bb07](https://github.com/jadmadi/thermal/commit/478bb07d6fe6d04c02ca1938e714674f3ddf5030))
* **receipt:** require single observed command invocations for verified evidence (goals/receipt-evidence-integrity/goal.md) ([fe2efdc](https://github.com/jadmadi/thermal/commit/fe2efdc604ebb56a3760f067af8ade28303d2833))
* **render:** prevent overlapping heatmap month labels (goals/heatmap-month-label-overlap/goal.md) ([60cb42a](https://github.com/jadmadi/thermal/commit/60cb42a2b16d0b830aa2c3c4c69b66eb449dcdad))
* **replay:** conserve workload volume and disclose incomplete pricing (goals/workload-replay/goal.md) ([9a58fb7](https://github.com/jadmadi/thermal/commit/9a58fb7d5c99d7a2b1d66f8ff57715b312fd6eba))
* **replay:** report rolling-window capacity as UNKNOWN without intraday evidence (goals/chatgpt-rolling-burst-replay/goal.md) ([21ba55f](https://github.com/jadmadi/thermal/commit/21ba55fb08d7a89846d91f94a786bf8efccca8e8))
* **replay:** resolve built-in plan identifiers passed to --against (goals/replay-against-plan-resolution/goal.md) ([e03a7f9](https://github.com/jadmadi/thermal/commit/e03a7f9ebc46b910ea6a6b3410f5aaad62ace8b5))
* **report:** expose zero unclassified fields and reconcile leaderboard totals (goals/output-report-totals-reconciliation/goal.md) ([a279a5f](https://github.com/jadmadi/thermal/commit/a279a5f74d925ac2369729fd287fc985a39d6459))
* **server:** bypass local disk presence check when synthetic loader is injected ([12be731](https://github.com/jadmadi/thermal/commit/12be731bf28d49be470eed673b187570d60911da))
* **server:** eliminate CPU thrashing via streaming timeout fix, file signature change detection & delta cache ([f4dd510](https://github.com/jadmadi/thermal/commit/f4dd510309376010fa782ca9e523b2a221e9d3a5))
* **server:** resolve undefined web telemetry fields, add JSON tags & craft polished animated dashboard ([3628029](https://github.com/jadmadi/thermal/commit/3628029be3d41b4dbf757b969834780982055716))
* **share:** label statusline/share scope and shorten the default URL (goals/statusline-and-share-scope-contract/goal.md) ([f0fc71e](https://github.com/jadmadi/thermal/commit/f0fc71ea0180a08537981b2c3d2bb4cd6126c3c3))
* **telemetry:** audit pricing estimates, dsh tool naming, isActivityOnly filtering, and stats UI (goals/telemetry-and-ui-audit/goal.md) ([97b6631](https://github.com/jadmadi/thermal/commit/97b6631df37aedf644c30082531f636590613e08))
* **test:** normalize expected Zed project path for cross-platform symlinks ([c5562e0](https://github.com/jadmadi/thermal/commit/c5562e0e0e9fd61ad730aaa19f401db69e307299))
* **test:** resolve cross-platform paths and line endings on windows and macos ([6543c34](https://github.com/jadmadi/thermal/commit/6543c34ab4ccf8104f20f0cdf5f0edcd5372f036))
* **thermal:** guard streak computation against malformed dates and invalid calendar tokens (goals/streak-parsing-defensive-guards/goal.md) ([e3cf52c](https://github.com/jadmadi/thermal/commit/e3cf52cbb510b8822b15dd5c431f9405c36b358f))
* **thermal:** separate activity steps from token totals and fix model attribution (goals/unpriceable-token-sources/goal.md) ([42bb6bb](https://github.com/jadmadi/thermal/commit/42bb6bb26f4da3829105d23a57cc9126decae36d))
* **tui:** clamp projection stats bar and band bounds against negative values ([2e95d42](https://github.com/jadmadi/thermal/commit/2e95d42314d8bc5a694be5bcd85e6f5ed885fea0))
* **upgrade:** check file close errors and deduplicate archive extraction (goals/cli-upgrade-file-integrity/goal.md) ([2c13a20](https://github.com/jadmadi/thermal/commit/2c13a201464447a0db500b7c60929c670c15386d))
* **upgrade:** increase asset download timeout to 120s for slower networks ([e43faf7](https://github.com/jadmadi/thermal/commit/e43faf7dbe9daa9c04fe386b9320065aa4139a11))


### Performance Improvements

* **loaders/agy:** add in-memory session cache and cap live monitor at 10 FPS ([196465e](https://github.com/jadmadi/thermal/commit/196465e537435dd3b18e02f20063731537594803))
* **loaders:** add incremental rollout transcript caching and delta parser for codex (goals/codex-incremental-rollout-cache/goal.md) ([ef07150](https://github.com/jadmadi/thermal/commit/ef0715072c37ba7fe869f8a561977a99a3fdc591))
* **loaders:** establish end-to-end ingestion performance baseline and benchmark suite (goals/ingestion-performance-baseline/goal.md) ([b713827](https://github.com/jadmadi/thermal/commit/b7138271d0f4178cf153311cae628d7b9f24b23c))
* **loaders:** reuse unchanged directory transcripts via a signature cache (goals/directory-scan-incremental-cache/goal.md) ([b252697](https://github.com/jadmadi/thermal/commit/b252697dd1a12472aecee0eb642320e4b3131954))
* **render:** preallocate leaderboard slices capacity and add allocation benchmark (goals/leaderboard-slice-preallocation/goal.md) ([2a2185a](https://github.com/jadmadi/thermal/commit/2a2185af10d4a1f7937b8775814d9e47b5d1d28b))

## [0.20.1](https://github.com/jadmadi/thermal/compare/v0.20.0...v0.20.1) (2026-10-14)


### Features

* **loaders:** rename command-code canonical display name to CMD ([0ab8458](https://github.com/jadmadi/thermal/commit/0ab845885c00773493dcaeb922818b6274136ddf))


## [0.20.0](https://github.com/jadmadi/thermal/compare/v0.19.0...v0.20.0) (2026-10-12)


### Features

* **ci:** add fast-path bypass to pre-push gate for docs and metadata diffs ([4ab6ebb](https://github.com/jadmadi/thermal/commit/4ab6ebba3c68e025a4fc5e08ff175fc985f31200))


### Bug Fixes

* **loaders:** exclude sqlite shm files to prevent live monitor CPU thrashing ([dd41a46](https://github.com/jadmadi/thermal/commit/dd41a4626a0c9ace26607b185177f9daf5b6ce94))
* **tui:** clamp projection stats bar and band bounds against negative values ([92b1f10](https://github.com/jadmadi/thermal/commit/92b1f100ef22c8590d2de19b63cd3ddf4dbaa277))


### Performance Improvements

* **loaders/agy:** add in-memory session cache and cap live monitor at 10 FPS ([0af717f](https://github.com/jadmadi/thermal/commit/0af717fbf3b28beba9e54cdb3ac2193388525472))
## [0.19.0](https://github.com/jadmadi/thermal/compare/v0.18.0...v0.19.0) (2026-10-08)


### Features

* **build:** verify multi-target cross-compilation release gate (goals/multi-target-cross-compile-ci-gate/goal.md) ([0efc97b](https://github.com/jadmadi/thermal/commit/0efc97b89ebf630046b0582cadc9f206f0eecf06))
* **ci:** verify cross-platform operating system test matrix (goals/cross-platform-ci-matrix/goal.md) ([70143cb](https://github.com/jadmadi/thermal/commit/70143cb1e136afa996437d24099928bc4d478f53))
* **docs:** update documentation parity and simulated human user audit for 15 tools (goals/docs-parity-and-human-simulation/goal.md) ([135b22e](https://github.com/jadmadi/thermal/commit/135b22ef673c49bfea8a32c646ff0fa0d7b13200))
* **gate:** ensure hermetic verification and release gate semantic parity (goals/hermetic-verification-gates/goal.md) ([1f71cd2](https://github.com/jadmadi/thermal/commit/1f71cd20bd10c90faf05dcabc22c16bcf5423f2b))
* **git:** verify zero bot contributor attribution in git history (goals/purge-bot-contributors/goal.md) ([cebda93](https://github.com/jadmadi/thermal/commit/cebda9375d156912adf81a950f5a5ec84aa2455c))
* **loaders:** ingest Zed editor assistant threads and tokens (goals/zed-editor-loader/goal.md) ([9b6a3ef](https://github.com/jadmadi/thermal/commit/9b6a3ef15721bfbbebec4b537a4ebbdc84f00476))
* **sila:** complete Sila onboarding and strict verification (goals/sila-onboarding/goal.md) ([8e3d3e9](https://github.com/jadmadi/thermal/commit/8e3d3e9afc72d6f3049f316b5e235795c85f068f))

## [0.18.0](https://github.com/jadmadi/thermal/compare/v0.17.0...v0.18.0) (2026-10-03)


### Features

* **loaders:** add platform-native application data directory discovery ([598c3d1](https://github.com/jadmadi/thermal/commit/598c3d163b703e432be86c2c3f8f2c8732b8a3d7))


### Bug Fixes

* **loaders:** normalize SQLite URI paths for cross-platform compatibility ([e30bd4d](https://github.com/jadmadi/thermal/commit/e30bd4daa49c61fd6ed657f6b4429dc3f06eacf0))
* **test:** resolve cross-platform paths and line endings on windows and macos ([2b27996](https://github.com/jadmadi/thermal/commit/2b27996648ace992b39a4bddec725e93515c875b))

## [0.17.0](https://github.com/jadmadi/thermal/compare/v0.16.0...v0.17.0) (2026-09-30)


### Features

* **server:** craft dense operator dashboard with dark/light themes, inline sparklines & keyboard navigation ([a2bfb9f](https://github.com/jadmadi/thermal/commit/a2bfb9f67bb4a4836e700cd8186278494e7d6560))


### Bug Fixes

* **loaders:** support relative ANTIGRAVITY_APP_DATA_DIR, unified sibling folding & ~/.config/thermal/thermal.config paths ([8024634](https://github.com/jadmadi/thermal/commit/8024634a44a3a9e5b610f0ff9ff34f3975b4c03e))
* **server:** eliminate CPU thrashing via streaming timeout fix, file signature change detection & delta cache ([f8e7759](https://github.com/jadmadi/thermal/commit/f8e77592ff9b307c5ff310359425be1f6746c5cf))
* **server:** resolve undefined web telemetry fields, add JSON tags & craft polished animated dashboard ([30fb8ff](https://github.com/jadmadi/thermal/commit/30fb8ff6ebbd06a5728cc9f5aac9c92bfb03ac4f))

## [0.16.0](https://github.com/jadmadi/thermal/compare/v0.15.0...v0.16.0) (2026-09-29)


### Features

* **audit:** add repository AGENTS.md readiness and agent hygiene check (goals/agent-readiness-audit/goal.md) ([259d1e0](https://github.com/jadmadi/thermal/commit/259d1e0b408a11f160ddd956dc6220bbfe11b13d))
* **cli:** add instant shell statusline and prompt composable helper (goals/terminal-statusline-helper/goal.md) ([26a20ae](https://github.com/jadmadi/thermal/commit/26a20ae6c7dbf6c66ac2df486ad41d091dd060a0))
* **models:** add reasoning effort breakdown & cognitive intensity telemetry (goals/reasoning-effort-telemetry/goal.md) ([0f3f533](https://github.com/jadmadi/thermal/commit/0f3f533f91307842a0461583e519ac4b01359c03))
* **receipt:** extract Codex sandboxed test runs and exit codes (goals/codex-sandboxed-receipts/goal.md) ([198772e](https://github.com/jadmadi/thermal/commit/198772efb99c05aea1eb00864adfb08c0ebbfe63))
* **render:** add PR-ready markdown output for receipt and yield (goals/markdown-pr-receipts/goal.md) ([c00ef02](https://github.com/jadmadi/thermal/commit/c00ef02ff312df8df8bdfa8ad01af79135aefa13))
* **replay:** add ChatGPT 5-hour rolling burst modeling (goals/chatgpt-rolling-burst-replay/goal.md) ([b0759e0](https://github.com/jadmadi/thermal/commit/b0759e00379cd9efc4c0d5b519957ce726a270ee))
* **yield:** analyze session fork lineage and branch churn in code delta yield (goals/session-fork-tree-analytics/goal.md) ([56ec02b](https://github.com/jadmadi/thermal/commit/56ec02bf8efeb8b952eb4ea7fbde80d4f2e8ae4c))


### Bug Fixes

* **upgrade:** increase asset download timeout to 120s for slower networks ([25b9b5f](https://github.com/jadmadi/thermal/commit/25b9b5fcf606f27868422ec1a74bce554847bfe8))

## [0.15.0](https://github.com/jadmadi/thermal/compare/v0.14.0...v0.15.0) (2026-09-25)


### Features

* **analytics:** shared chart analytics, distribution stats and trend projections (goals/charting-analytics/goal.md) ([a059d01](https://github.com/jadmadi/thermal/commit/a059d0190a5f99b7ed5ec5efa231d91fe8e1ee62))
* **changelog:** add interactive changelog command with top filtering and JSON export ([bacde98](https://github.com/jadmadi/thermal/commit/bacde988c6bad71f7a19f6db2b6dfb412ebe48c5))
* **chart:** support --chart flag across reports, document dashboard and conventions (goals/charting-flags-docs-release/goal.md) ([25163db](https://github.com/jadmadi/thermal/commit/25163dbac85f1ea23d03c89bbc5eea307822b2ab))
* **check:** introduce fast-path pre-commit mode with package tests & diff sentry (goals/tiered-precommit-fastpath/goal.md) ([1b5a3c8](https://github.com/jadmadi/thermal/commit/1b5a3c88bb640c146c1a41236ab4636073cee6a1))
* **ci:** add OpenSSF SLSA Level 3 build provenance and verification runbook (goals/slsa3-release-provenance/goal.md) ([78b8258](https://github.com/jadmadi/thermal/commit/78b82584a94f2e64ebec1e40c239892647d46d9d))
* **cli:** add staged deprecation warnings & upgrade migration alerts (goals/cli-staged-deprecation-and-upgrade/goal.md) ([2842f1a](https://github.com/jadmadi/thermal/commit/2842f1a839026349543a10d6a2172b62b20ecef3))
* **cli:** add stateless zero-database url sharing engine (goals/stateless-telemetry-sharing/goal.md) ([7557bf9](https://github.com/jadmadi/thermal/commit/7557bf99aceef2d40792072847c44e36bc33c993))
* **cli:** add thermal audit command for local setup and context health (goals/setup-health-audit/goal.md) ([8afd798](https://github.com/jadmadi/thermal/commit/8afd798a12491a19e7f572160f511241e9902fec))
* **cli:** add thermal receipt command for verifiable work outcomes (goals/verifiable-work-receipts/goal.md) ([ebc550a](https://github.com/jadmadi/thermal/commit/ebc550afe6ee44437d2e6d019af1ba3f0066a453))
* **cli:** add thermal serve embedded local web dashboard and telemetry APIs (goals/embedded-web-dashboard/goal.md) ([dceba19](https://github.com/jadmadi/thermal/commit/dceba19f4aa0651c1b7c5443f54046ba2cb14711))
* **cli:** add thermal yield command for code output delta telemetry (goals/token-yield-metrics/goal.md) ([ef10ceb](https://github.com/jadmadi/thermal/commit/ef10cebcf9f112c1f8fa23d7009743716067ee50))
* **contributing:** establish contributor onboarding path, preflight check & retention loop (goals/contributor-onboarding/goal.md) ([3bd8d0a](https://github.com/jadmadi/thermal/commit/3bd8d0a3e92d9de0c78c8dc34a0c15b284ed1d04))
* **devin:** project sessions.model onto DailyRow and ProjectDay for accurate model attribution (goals/devin-model-attribution/goal.md) ([ecc9a92](https://github.com/jadmadi/thermal/commit/ecc9a920ec460ea8e3035ca0665680bef168f886))
* **distribution:** configure homebrew tap, installer script, agent docs & distribution runbook (goals/distribution-coverage-foundation/goal.md) ([80b3e4f](https://github.com/jadmadi/thermal/commit/80b3e4fe566c7b106d14d868615e363ce9dbe2b9))
* **docs:** establish canonical migration guide & governance gate (goals/version-migration-guide/goal.md) ([12d9833](https://github.com/jadmadi/thermal/commit/12d98338458d89935843468c03df287be753cd22))
* **governance:** establish Written Solo Authority model and scope defense (goals/governance-foundations/goal.md) ([332ad26](https://github.com/jadmadi/thermal/commit/332ad262b176f427d2a3535c7db686ba082fa41d))
* **governance:** implement staged trust ladder, succession runbook & absence factor checks (goals/trust-ladder-and-succession/goal.md) ([9fa6245](https://github.com/jadmadi/thermal/commit/9fa62450d31d8814910e5cd05ce857d3d8665647))
* **license:** add Community Reciprocity Covenant and license CLI metadata (goals/community-covenant-neutrality/goal.md) ([9a82384](https://github.com/jadmadi/thermal/commit/9a8238472f7d5c2b51d455117143bc3681cf972d))
* **license:** AGPL-3.0 relicensing, commercial dual-licensing & contributor gating (goals/agpl-dual-licensing/goal.md) ([eb9aeb4](https://github.com/jadmadi/thermal/commit/eb9aeb4e79cf31307d979813e5c2a47ed6d35dee))
* **license:** harden licensing architecture, CI enforcement & SPDX headers (goals/license-strategy-and-enforcement/goal.md) ([98772b5](https://github.com/jadmadi/thermal/commit/98772b519955e1ec8ec0a0dbde1a8897dbfa98f3))
* **live:** attribute resumed session projects, seed today's baseline, and track activity units ([f463509](https://github.com/jadmadi/thermal/commit/f46350990e48357b9d3b26c508de50108402b552))
* **loaders:** add DeepSeek harness (dsh) loader and registry (goals/dsh-loader/goal.md) ([9be2513](https://github.com/jadmadi/thermal/commit/9be2513404ae96090832cc4feef74259c65fc179))
* **loaders:** add Nous Hermes agent ingestion loader for state.db (goals/hermes-loader/goal.md) ([ad2d523](https://github.com/jadmadi/thermal/commit/ad2d5238e2866fade6b359adb7a8704a8d6964ef))
* **loaders:** report truncated JSONL scans via Summary.Warnings to verbose stderr (goals/jsonl-scan-diagnostics/goal.md) ([17b6b59](https://github.com/jadmadi/thermal/commit/17b6b595f3fca038b5d20ec32bf78bdba0b9c8c2))
* **loaders:** support unified Antigravity hub and OpenCode model breakdowns ([2afb9f8](https://github.com/jadmadi/thermal/commit/2afb9f8d1ec5bc87524783f2b5f98e1216d28546))
* **models:** declare canonical model identities and router alias normalization (goals/model-identity-and-alias-resolution/goal.md) ([e460d4a](https://github.com/jadmadi/thermal/commit/e460d4a769928c28117e2b8b7908be79ceb2bf78))
* **pricing:** catalog coverage and tier suffix lookup for tool-native model variants (goals/pricing-catalog-coverage/goal.md) ([79bbd75](https://github.com/jadmadi/thermal/commit/79bbd7543fc59c4a5d5df841d9e3e420bb01c672))
* **release:** adapt changelog-writing standards and configure release-please sections (goals/release-communication-and-changelog/goal.md) ([7b0813c](https://github.com/jadmadi/thermal/commit/7b0813ccd44e14d2dd5056f2f5907740dd5f7d78))
* **release:** script legacy release retirement & document AGPL-3.0 release line (goals/purge-legacy-mit-releases/goal.md) ([cee5b3b](https://github.com/jadmadi/thermal/commit/cee5b3bd8e74d0bb4fb49118b81c437517d1ce76))
* **render:** establish unified terminal design tokens, adaptive cards, and text wrapping ([a8e65d7](https://github.com/jadmadi/thermal/commit/a8e65d774952b1e656183a59b4be1d4645f7272f))
* **replay:** simulate historical developer workloads against subscription limits and model cards (goals/workload-replay/goal.md) ([1ca92ed](https://github.com/jadmadi/thermal/commit/1ca92edc4c508c65679e75fb7950f2c89b3bafcc))
* **repo:** establish code ownership, maintainers, security policy & PR lifecycle (goals/repo-alignment-and-security/goal.md) ([bee76ff](https://github.com/jadmadi/thermal/commit/bee76ff55e8a2e1d6ef384d24bf9c30bb0249b36))
* **report:** verify weekly and monthly period totals and table formatting parity (goals/report-period-totals/goal.md) ([b8fac89](https://github.com/jadmadi/thermal/commit/b8fac89012194436ef29dbde37fb1e5dcc4b0f83))
* **server:** make web totals match canonical reports and support collection options (goals/web-telemetry-parity/goal.md) ([d2cd4a2](https://github.com/jadmadi/thermal/commit/d2cd4a2e67cc803f19291749ad151496c494802b))
* **server:** reject untrusted host authorities and cross-site requests at telemetry API (goals/localhost-request-boundary/goal.md) ([d162bf1](https://github.com/jadmadi/thermal/commit/d162bf1f679b436e6b9cab2c158c46705b536765))
* **stats:** expose disjoint token composition and cache hit rate metrics (goals/cache-efficiency-metrics/goal.md) ([fa4c8f9](https://github.com/jadmadi/thermal/commit/fa4c8f9c0de80b6c5b92cebb61b590598b23b60b))
* **tui:** add high-density 9-box finops grid and activity taxonomy view (goals/tui-dense-finops-grid/goal.md) ([f720e57](https://github.com/jadmadi/thermal/commit/f720e5798584a1e4360059243dd98bfa6df472ba))
* **tui:** bubble tea v2 application shell and overview dashboard (goals/tui-shell-overview/goal.md) ([85af7be](https://github.com/jadmadi/thermal/commit/85af7bef073805878594fbadb14611f6bdfc07ac))
* **tui:** elevate codeburn-inspired 9-box FinOps grid & dashboard dense view ([88c24d5](https://github.com/jadmadi/thermal/commit/88c24d510ae8bab65a32ec190f399fd022bbcf36))
* **tui:** interactive projects ranking view and drill-down detail (goals/tui-projects-view/goal.md) ([50faef3](https://github.com/jadmadi/thermal/commit/50faef37cace029df5f9b3d6284708715324cf79))
* **tui:** interactive statistical distribution, weekday cadence and outlier view (goals/tui-stats-view/goal.md) ([be1e45f](https://github.com/jadmadi/thermal/commit/be1e45fed317b7dfd1db2b709190448d431f8281))
* **tui:** interactive tool and model mix share views with switching metrics (goals/tui-mix-models-views/goal.md) ([43c146e](https://github.com/jadmadi/thermal/commit/43c146e37b2ae87872bea77da5eb51f307e538a4))
* **tui:** spike ntcharts v2 against stock charm stack and record findings (goals/charting-dep-spike/goal.md) ([d1b68ab](https://github.com/jadmadi/thermal/commit/d1b68ab84f9e297fdfd81edb688ff4dccfc1ae91))


### Bug Fixes

* **claude:** deduplicate assistant messages by message id and report in summary warnings (goals/claude-message-dedupe/goal.md) ([d39ef40](https://github.com/jadmadi/thermal/commit/d39ef406b54d0196ac716ee5443b45b103c02195))
* **live:** align token burn, turn counts, and cache rates with canonical metrics (goals/live-metric-parity/goal.md) ([305f589](https://github.com/jadmadi/thermal/commit/305f589714a8b40ef8cf0a870ad2b3ab84422ed8))
* **live:** decouple display cadence and cache unchanged sources to bound 100ms CPU (goals/live-polling-cpu-budget/goal.md) ([b71d464](https://github.com/jadmadi/thermal/commit/b71d46487fd4f1a535aed1f6b4d0a2a047ca58cf))
* **loaders:** enforce SQLite rows iteration error inspection and document standard (goals/database-loader-iteration-safety/goal.md) ([c53c32b](https://github.com/jadmadi/thermal/commit/c53c32b0b241ceaeb3b52707b059d46d931c7ad8))
* **loaders:** ignore empty WAL and volatile SHM locks in Devin source identity ([31b7b84](https://github.com/jadmadi/thermal/commit/31b7b84d1a23ff12ea5ab1d3ce5edcf18fdf8b25))
* **loaders:** isolate Devin cache by source database and validate freshness (goals/devin-cache-source-isolation/goal.md) ([d182cf9](https://github.com/jadmadi/thermal/commit/d182cf9e0f7134733ce94b92725e19a3115087a0))
* **loaders:** raise JSONL scanner line ceiling to 32 MiB for large transcripts (goals/jsonl-line-ceilings/goal.md) ([86d6ebb](https://github.com/jadmadi/thermal/commit/86d6ebb1d2624e9b3c8efd7ae6ba7d6d32cfe810))
* **receipt:** preserve recorded usage, deduplicate messages, and disclose partial coverage (goals/receipt-accounting-coverage/goal.md) ([dd3ab4f](https://github.com/jadmadi/thermal/commit/dd3ab4f987b51f7fe6868af1a70164d1625a4455))
* **receipt:** require observed command results and drop uncompleted invocations (goals/receipt-evidence-integrity/goal.md) ([7a8ce91](https://github.com/jadmadi/thermal/commit/7a8ce91cd22fd928cccd8a75d1470bea36591707))
* **server:** bypass local disk presence check when synthetic loader is injected ([8e47085](https://github.com/jadmadi/thermal/commit/8e47085a1750a54bb8ffbb46f95a32af6c167628))
* **telemetry:** audit pricing estimates, dsh tool naming, isActivityOnly filtering, and stats UI (goals/telemetry-and-ui-audit/goal.md) ([8c9847d](https://github.com/jadmadi/thermal/commit/8c9847db412ee0669f5257e153ec888220487722))
* **thermal:** guard streak computation against malformed dates and invalid calendar tokens (goals/streak-parsing-defensive-guards/goal.md) ([a2467d7](https://github.com/jadmadi/thermal/commit/a2467d79c51d6c35bb89c9726c39e008b0e25eba))
* **thermal:** separate activity steps from token totals and fix model attribution (goals/unpriceable-token-sources/goal.md) ([09eb367](https://github.com/jadmadi/thermal/commit/09eb367a96fc928ea6f8fada48a8fb09eb3cafcc))
* **upgrade:** check file close errors and deduplicate archive extraction (goals/cli-upgrade-file-integrity/goal.md) ([15d2ffb](https://github.com/jadmadi/thermal/commit/15d2ffb0a8eccfb5f40ffc31c1cd9b9549fc0cc4))


### Performance Improvements

* **loaders:** add incremental rollout transcript caching and delta parser for codex (goals/codex-incremental-rollout-cache/goal.md) ([752569e](https://github.com/jadmadi/thermal/commit/752569efa1dec1d86c802d4809ff42c5e80923d6))
* **loaders:** establish end-to-end ingestion performance baseline and benchmark suite (goals/ingestion-performance-baseline/goal.md) ([e2a569f](https://github.com/jadmadi/thermal/commit/e2a569f035ce3405f06bd2c2a3e3e653d41bdf98))
* **render:** preallocate leaderboard slices capacity and add allocation benchmark (goals/leaderboard-slice-preallocation/goal.md) ([f375fdf](https://github.com/jadmadi/thermal/commit/f375fdf4a93adfed1375dac534ffa002e19fb997))

## [0.14.0](https://github.com/jadmadi/thermal/compare/v0.13.0...v0.14.0) (2026-09-22)


### Features

* **cli:** add daily cached background update check and self-update command (`thermal upgrade`) ([2c387fb](https://github.com/jadmadi/thermal/commit/2c387fb58f9452000d0f6d4cbfdcb1c1947105cc))
* **analytics:** add visual analytics and terminal charting showcase (`--chart`) ([de17df6](https://github.com/jadmadi/thermal/commit/de17df6387b70e7cbc2ad9d03b3d395c1609fb44))
* **discovery:** declare Content-Signal directives and Markdown for Agents ([fd5ca83](https://github.com/jadmadi/thermal/commit/fd5ca8395d31dbcfe8f46fd7120883562ae1315d))


### Bug Fixes

* **docs:** sanitize project names and paths across documentation and test suites ([345afbe](https://github.com/jadmadi/thermal/commit/345afbe8e7887aa85bf5644784a0d92375a0c0ad))
* **docs:** prune unbacked discovery manifests and ensure strictly authentic documentation ([5ebccda](https://github.com/jadmadi/thermal/commit/5ebccdaae3ee5f41dc68cb7297e6840742f9b1c7))
* **discovery:** remove hallucinated payment, OAuth, and openapi references ([33895c6](https://github.com/jadmadi/thermal/commit/33895c634c38547503e11209c022f73d9704f40a))

## [0.13.0](https://github.com/jadmadi/thermal/compare/v0.12.0...v0.13.0) (2026-09-22)


### Features

* **init:** initial release of Thermal under GNU AGPL-3.0 with commercial dual-licensing ([23a7733](https://github.com/jadmadi/thermal/commit/23a7733a8ce62fba5a8f52a0eb3e17c0581f806a))

## [0.12.0](https://github.com/jadmadi/thermal/compare/v0.11.0...v0.12.0) (2026-09-21)


### Features

* **replay:** implement thermal replay simulation against subscriptions and APIs (`thermal replay`) (goals/workload-replay/goal.md) ([4f7ff02](https://github.com/jadmadi/thermal/commit/4f7ff02f5e7710deabc5454265330bcbee35057a))
* **testing:** add automated simulated user release gate and update gitignore ([b5651df](https://github.com/jadmadi/thermal/commit/b5651df195f1702078d233ba9fb3d4c5ea98c29c))


### Bug Fixes

* **testing:** ensure simulated user tests use dynamic timestamps and isolated home environment ([2c578a0](https://github.com/jadmadi/thermal/commit/2c578a0fd86e035939805ecd1a7612005d70502e))

## [Unreleased]

### Features

* **changelog:** add release history and features viewer command (`thermal changelog --top <N>`)
* **audit:** add local context health and rule tax audit command (`thermal audit`)
* **replay:** simulate historical developer workloads against subscription tiers and API models (`thermal replay --against <plan>`)
* **yield:** measure code delta velocity and token yield efficiency (`thermal yield --sort lines|yield`)
* **receipt:** evaluate verifiable test and linter outcomes (`thermal receipt --sort verified`)
* **share:** generate zero-database stateless compressed URL cards (`thermal share`)
* **formatting:** enforce MD013 terminal line-length bounds and hanging indent word wrapping across all views

## [0.11.0](https://github.com/jadmadi/thermal/compare/v0.10.0...v0.11.0) (2026-09-21)


### Features

* **analytics,render:** expose disjoint token composition and cache efficiency metrics (goals/cache-efficiency-metrics/goal.md) ([6e070bd](https://github.com/jadmadi/thermal/commit/6e070bd87108f8b6994d5adaeba7e5cff1d4cc48))
* **loaders:** add Nous Hermes agent loader and registry (goals/hermes-loader/goal.md) ([9e90d01](https://github.com/jadmadi/thermal/commit/9e90d017e24a4a4cdd04bcdd9b32f166a824b941))
* **loaders:** add prompt-level turn deduplication to Grok loader ([e589226](https://github.com/jadmadi/thermal/commit/e5892265a86dd65f07b7ab1dfb5df6e34543bab8))
* **loaders:** implement message-level deduplication in Claude Code loader (goals/claude-message-dedupe/goal.md) ([104ba23](https://github.com/jadmadi/thermal/commit/104ba23a39bdaacfe4d16e9e34d4dbbe2ed61da8))

## [0.10.0](https://github.com/jadmadi/thermal/compare/v0.9.1...v0.10.0) (2026-09-21)


### Features

* **docs:** polish layout, motion choreography, and window chrome in html docs ([d9ab799](https://github.com/jadmadi/thermal/commit/d9ab7997b7c5e56be42354f858ac69f72ebc3161))
* **docs:** switch canonical site domain to thermal.jadmadi.net ([a1c9c4d](https://github.com/jadmadi/thermal/commit/a1c9c4dd5002a62b4954ee3a820a5817b9d0fc43))
* **loaders:** add DeepSeek harness (dsh) loader and registry (goals/dsh-loader/goal.md) ([414a815](https://github.com/jadmadi/thermal/commit/414a8151615119ca3f043a8beafe8a67f9bfb194))
* **loaders:** report JSONL scans that stop early via Summary.Warnings (goals/jsonl-scan-diagnostics/goal.md) ([247bb80](https://github.com/jadmadi/thermal/commit/247bb80b2a49ceb455beb3cb76d650a5f43bb8f1))
* **pricing:** map tier variants to the base model, and classify the rest ([b361693](https://github.com/jadmadi/thermal/commit/b361693abb24a963f62b5fc0769288dd9322b5ac))
* **render,analytics:** telemetry accuracy, pricing estimates, and stats UI polish (goals/telemetry-and-ui-audit/goal.md) ([662f25a](https://github.com/jadmadi/thermal/commit/662f25a3e9c761f4369cd20425e75c268299dbee))
* **sila:** complete sila onboarding and wire repository (goals/sila-onboarding/goal.md) ([92ea6ee](https://github.com/jadmadi/thermal/commit/92ea6ee123cb044818dab401ce450f59e83617d3))


### Bug Fixes

* **loaders:** raise JSONL line ceilings to 32 MiB (goals/jsonl-line-ceilings/goal.md) ([ada5b52](https://github.com/jadmadi/thermal/commit/ada5b526722e6f20b3b8281a65b574d616ab7170))

## [0.9.1](https://github.com/jadmadi/thermal/compare/v0.9.0...v0.9.1) (2026-09-20)


### Bug Fixes

* **codex:** scan large rollout lines, and keep step counts out of token totals ([249b1ee](https://github.com/jadmadi/thermal/commit/249b1ee0df083268986fec5f3eb56b1a3639594b))
* **codex:** scan large rollout lines, and keep step counts out of token totals ([ce31aaa](https://github.com/jadmadi/thermal/commit/ce31aaac0c4a78fba84961372943f39ba0400643))

## [0.9.0](https://github.com/jadmadi/thermal/compare/v0.8.0...v0.9.0) (2026-09-20)


### Features

* **devin:** attribute models to project rows so projects can price them ([5741f5f](https://github.com/jadmadi/thermal/commit/5741f5fc98289f3712858833391d71324489ea48))
* **devin:** attribute models to project rows so projects can price them ([e4d542c](https://github.com/jadmadi/thermal/commit/e4d542cfce7f6df0166711b04299592a231f9b25))

## [0.8.0](https://github.com/jadmadi/thermal/compare/v0.7.0...v0.8.0) (2026-09-19)


### Features

* **devin:** attribute tokens to the session model so they can be priced ([0508a4e](https://github.com/jadmadi/thermal/commit/0508a4ea15c68d5718ba4a3aa694ebdd22cbde12))
* **devin:** attribute tokens to the session model so they can be priced ([818975d](https://github.com/jadmadi/thermal/commit/818975d0bfac021b5b0daad2f25373d4bb44c05c))


### Bug Fixes

* **reports:** count the tokens no price covers, and stop dropping token days ([49d6766](https://github.com/jadmadi/thermal/commit/49d676643bfdef21d531e396aa4147ead4ac6b3d))
* **reports:** count unpriceable tokens and stop dropping token days ([b130221](https://github.com/jadmadi/thermal/commit/b130221746f843f926627587acc4b5705a8a6135))

## [0.7.0](https://github.com/jadmadi/thermal/compare/v0.6.0...v0.7.0) (2026-09-19)


### Features

* **analytics:** add mix, stats, and trend verbs ([#22](https://github.com/jadmadi/thermal/issues/22)) ([9c28a0a](https://github.com/jadmadi/thermal/commit/9c28a0af26a610ebaff49a8b02cc9635e8120e85))
* **docs:** add author links, schema, FAQ, and crawler files to the docs site (goals/report-period-totals/goal.md) ([ccc7bc1](https://github.com/jadmadi/thermal/commit/ccc7bc1226aba6dc183de09f35dfa4358d31d2ab))
* **docs:** docs site, corrected report totals, and Pages publishing ([7a470d1](https://github.com/jadmadi/thermal/commit/7a470d15c49c290cbe27f9b631e62d11185a60e6))
* **reports:** --chart, dashboard docs, and the rules ([656db32](https://github.com/jadmadi/thermal/commit/656db32151827d2a6afa84160dd5d49201d40574))
* **reports:** add --chart, document the dashboard, and record the rules ([1cb2f50](https://github.com/jadmadi/thermal/commit/1cb2f501825abb429f30940eb3da20e8fa715567))
* **reports:** keep token-bearing periods only, grain-independent totals (goals/report-period-totals/goal.md) ([9c27be4](https://github.com/jadmadi/thermal/commit/9c27be46c625036754c9637b447bbe85d43372fd))
* **tui:** add the dashboard shell and Overview view ([050f4a5](https://github.com/jadmadi/thermal/commit/050f4a5bcf2cf4f27f08a42807f6b233285720d5))
* **tui:** add the Mix and Models views ([6a591f8](https://github.com/jadmadi/thermal/commit/6a591f8d72fb895bf374d50f98acf929792d073c))
* **tui:** add the Projects view with drill-down ([e15056f](https://github.com/jadmadi/thermal/commit/e15056f8cf5c4c4c40c1bbe38bec0b279b45683c))
* **tui:** add the Stats view ([338b8bd](https://github.com/jadmadi/thermal/commit/338b8bd833ab14bd45fb77ef8ed1ec758904cdc2))
* **tui:** dashboard shell and Overview view ([e8a5dbb](https://github.com/jadmadi/thermal/commit/e8a5dbb8436baed6829d004766104e9e574d02df))
* **tui:** Mix and Models views ([70c6239](https://github.com/jadmadi/thermal/commit/70c62399ad1daf7a32fc585bb243112b0e4684af))
* **tui:** Projects view with drill-down ([c5bb23c](https://github.com/jadmadi/thermal/commit/c5bb23c21486c5e61ea27781d96b8b8ff27c8a6d))
* **tui:** Stats view ([76b4fba](https://github.com/jadmadi/thermal/commit/76b4fba65c99b78427328e530c15f97dba835328))


### Bug Fixes

* **cost:** show the recorded and estimated split everywhere the total appears ([b76b9a5](https://github.com/jadmadi/thermal/commit/b76b9a538ea1c0ad6bf6b0507ebe825c7b019226))
* **cost:** show the recorded and estimated split everywhere the total appears ([66a7466](https://github.com/jadmadi/thermal/commit/66a74664fab3b6d5b6d58912da62ad340e3cbc11))
* **pages:** move the JSON reference inside the reports section and reindent it (goals/report-period-totals/goal.md) ([c9c6af8](https://github.com/jadmadi/thermal/commit/c9c6af80892ec4f8fa0e1a284db251ee1e7f3abe))
* **tools:** drop the ambiguous cc alias and point tool links at their sites (goals/report-period-totals/goal.md) ([2a81533](https://github.com/jadmadi/thermal/commit/2a81533383ea852f0f274f205a2aa47133a86c02))
* **tui:** keep the cursor on screen and the frame inside the terminal ([b1df32b](https://github.com/jadmadi/thermal/commit/b1df32bae06bf0eaec9a45db9c6a59ac309a0e78))
* **tui:** keep the cursor on screen and the frame inside the terminal ([0619190](https://github.com/jadmadi/thermal/commit/0619190fa2bba9f8486274318687771c36ab11ff))

## [0.6.0](https://github.com/jadmadi/thermal/compare/v0.5.0...v0.6.0) (2026-09-16)


### Features

* add daily, weekly, monthly reports, projects, and models ([#15](https://github.com/jadmadi/thermal/issues/15)) ([2097a62](https://github.com/jadmadi/thermal/commit/2097a62a258fea485abf5cc84b64ede17f764ac8))
* rank tokens and cost per project across tools ([#18](https://github.com/jadmadi/thermal/issues/18)) ([8eb3c27](https://github.com/jadmadi/thermal/commit/8eb3c278ab091993162a98717511d0c42e808532))

## [0.5.0](https://github.com/jadmadi/thermal/compare/v0.4.0...v0.5.0) (2026-09-15)


### Features

* **claude,droid:** add Claude Code token loader and Droid activity loader ([#13](https://github.com/jadmadi/thermal/issues/13)) ([e3254d1](https://github.com/jadmadi/thermal/commit/e3254d10a3741013f9dbae74510749fb45e88a57))

## [0.4.0](https://github.com/jadmadi/thermal/compare/v0.3.0...v0.4.0) (2026-09-15)


### Features

* **grok:** add Grok CLI loader from session turn telemetry ([#10](https://github.com/jadmadi/thermal/issues/10)) ([85af773](https://github.com/jadmadi/thermal/commit/85af773d6da6504f028b253d0b6a284a9f917c76))
* **muse:** add activity-only Muse loader from session index ([#11](https://github.com/jadmadi/thermal/issues/11)) ([f15f943](https://github.com/jadmadi/thermal/commit/f15f9431a1f62512fab0d3d25a0f95905e725be6))

## [0.3.0](https://github.com/jadmadi/thermal/compare/v0.2.3...v0.3.0) (2026-09-15)


### Features

* **zcode:** add ZCode loader with model usage telemetry ([#6](https://github.com/jadmadi/thermal/issues/6)) ([54cd9d7](https://github.com/jadmadi/thermal/commit/54cd9d7bee17fec079c9b0999af202994d5410b4))

## [0.2.3](https://github.com/jadmadi/thermal/compare/v0.2.2...v0.2.3) (2026-09-13)


### Bug Fixes

* **agy:** read live antigravity-cli transcripts for step activity ([#4](https://github.com/jadmadi/thermal/issues/4)) ([c7f4fa6](https://github.com/jadmadi/thermal/commit/c7f4fa631d5020dc2647644a7bfd26b3ae5ea3d3))

## [0.2.2](https://github.com/jadmadi/thermal/compare/v0.2.1...v0.2.2) (2026-09-12)


### Bug Fixes

* **opencode:** read session_v2 table written by OpenCode v2 ([#2](https://github.com/jadmadi/thermal/issues/2)) ([037a073](https://github.com/jadmadi/thermal/commit/037a0735d4662109651b6ba920b7142f461897b6))

## [0.2.1](https://github.com/jadmadi/thermal/compare/v0.2.0...v0.2.1) (2026-09-11)

Documentation and repository chores only; no user-facing changes. Tagged outside release-please, recorded here to keep the manifest in sync.

## [0.2.0](https://github.com/jadmadi/thermal/compare/v0.1.0...v0.2.0) (2026-07-18)


### Features

* add open source release plan and config ([61c9449](https://github.com/jadmadi/thermal/commit/61c9449d6baab5945e7a1a540217e5405cbc41b8))
* optimize loaders with memory-mapped SQLite and parallel workers, add build.sh automation, and integrate release-please ([bdf5b2d](https://github.com/jadmadi/thermal/commit/bdf5b2da63d0270b466802250804df5b21e0a330))


### Bug Fixes

* update author name in readme ([19e2c54](https://github.com/jadmadi/thermal/commit/19e2c5448949c306a7cb65b815aa3ac2f17031cd))

## [Unreleased]

### Added

- **`build.sh` build automation script** — versatile build script supporting `--release` (`-ldflags="-s -w" -trimpath`), `--dev`, and `--upx` modes, cross-compilation (`--target` and `--all`), and detailed `--help` documentation explaining build trade-offs and binary sizes.
- **`--verbose` / `-v` diagnostic flag** — outputs non-fatal tool loading failures, missing databases, and parsing warnings directly to `os.Stderr`.
- **`thermal upgrade` command** — self-upgrade by checking GitHub Releases, downloading the matching binary for your OS/arch, and atomically replacing the running executable.
- **`thermal version` command** — prints version, commit hash, and build date (injected via ldflags at build time).
- **`internal/version` package** — build-time version embedding via `-ldflags`.
- **Token Warriors / Activity Hunters split** — the leaderboard now separates tools with real token consumption data from tools that only track activity (messages, steps). Tools are auto-categorized based on whether token data is available.
- **Cost column** in the leaderboard for tools that track cost (OpenCode, MiMoCode, codewhale). Sub-cent formatting shows costs like `$0.0032`.
- **Devin loader** — extracts real token metrics from `message_nodes.metadata.metrics` (input, output, cache, reasoning tokens), session duration from `sessions` table, and engagement from `prompt_history`. Devin moved from Activity Hunters to Token Warriors.
- **Codex loader rewrite** — reads `state_5.sqlite` (`threads.tokens_used`, `model`, `source`, `reasoning_effort`, `agent_role`, session duration) as the primary source, supplements with rollout JSONL for the input/output/reasoning/cache token breakdown. Codex moved from Activity Hunters to Token Warriors.
- **codewhale loader rewrite** — extracts `metadata.total_tokens`, `metadata.cost.session_cost_usd`, `metadata.model`, `metadata.mode`, and `metadata.cumulative_turn_secs` from session JSON files. codewhale moved from Activity Hunters to Token Warriors.
- **OpenCode loader rewrite** — leverages pre-aggregated `session.tokens_*` columns, `cost`, `agent`, and `summary_additions/deletions/files` for faster and more comprehensive data extraction.
- **MiMoCode loader** — aggregates from `message.data.tokens` and surfaces `session.summary_additions/deletions/files` and agent distribution.
- **Agy loader rewrite** — reads `overview.txt` from all brain sessions (was only reading `transcript.jsonl` from 1 of 14 sessions). Now surfaces 304 steps across 14 sessions instead of 57 steps from 1 session. Extracts model info from `transcript.jsonl` USER_SETTINGS_CHANGE events.
- **Session duration tracking** — `LongestSessionMs` field added to Summary, surfaced in dashboard for all tools that have session timestamps.
- **Model breakdown** — `ModelBreakdown` field shows which models were used per tool (e.g., Codex: gpt-5.4-mini:6, gpt-5.4:1).
- **Agent/source breakdown** — `AgentBreakdown` field shows source distribution (e.g., Codex: cli:7, vscode:3, exec:1).
- **Lines added/deleted and files touched** — `LinesAdded`, `LinesDeleted`, `FilesTouched` fields added to Summary for tools that track code diffs (OpenCode, MiMoCode).
- **Progress bar** — new `internal/render/progress.go` for loading indicator, hardened with `atomic.Bool` and `sync.Once` for thread safety.
- **Session cache** — `internal/loaders/cache.go` for caching loader results.
- **GitHub Actions release workflow** and **GoReleaser config** for cross-platform binary builds with version injection.

### Performance & Optimization

- **Devin SQLite database optimization** — added incremental delta scanning (`row_id > c.MaxRowID`), memory-mapped I/O (`PRAGMA mmap_size`), and byte-string pre-filtering (`chat_message LIKE '%"assistant"%'`), reducing query times on large (7GB+) databases from ~11s down to `< 2ms`.
- **SQLite memory-mapped I/O across all loaders** — `LoadOpenCodeData` and `LoadMiMoCodeData` benefit from `PRAGMA cache_size = -64000` and `PRAGMA mmap_size = 30GB` alongside raw string pre-filters.
- **Parallel multi-worker loaders** — `internal/loaders/agy.go`, `commandcode.go`, and `codex.go` use bounded parallel worker pools (limit of 8 concurrent workers) to scan multi-file directories and JSONL rollouts simultaneously.
- **Binary size reduction** — configured `-trimpath` and `-ldflags="-s -w"` in `.goreleaser.yml` and `build.sh` to strip DWARF symbol tables and host filepaths, reducing release binaries from ~15MB to ~9.8MB (35% reduction).

### Changed

- **Leaderboard format** — now split into two sections (Token Warriors, Activity Hunters) with a Cost column. Old format was a single flat list.
- **Tool registry architecture (`DataSubdir`)** — refactored tool registry path resolution (`internal/loaders/registry.go`) to use clean `DataSubdir` property mappings instead of hardcoded switch cases.
- **README** — updated supported tools table, leaderboard example, build instructions (`build.sh`), and How It Works section to reflect the new data sources, metrics, and optimizations.
- **`.goreleaser.yml`** — added `-trimpath` and ldflags (`-s -w`) to inject version, commit, and date into the binary while minimizing size.

### Fixed

- **Daylight Saving Time (DST) streak accuracy** — replaced 24-hour time subtraction (`24*time.Hour`) in streak logic (`internal/thermal/streak.go`) with calendar date arithmetic (`AddDate(0, 0, 1)`) to guarantee accuracy during spring-forward (23h) and fall-back (25h) transitions. Ensured strictly ascending quartile thresholds.
- **Self-upgrader network and decompression guards** — hardened `thermal upgrade` (`cmd/thermal/upgrade.go`) with package-wide 30-second HTTP timeouts, context propagation across requests, exact OS/arch asset checking, and `io.LimitReader(tr, 250*1024*1024)` decompression limits against disk starvation attacks.
- **Codex token over-counting (260x)** — the old parser summed ALL cumulative `total_token_usage` values from every `token_count` event in every rollout file, producing 11.6B instead of the correct 43.9M. The new loader uses `state_5.sqlite` as the source of truth and only reads the LAST `token_count` event per rollout file for the breakdown.
- **command-code session inflation** — the old glob `*.jsonl` matched 115 files (54 sessions + 54 checkpoints + 7 hooks-audit), inflating counts to 61 sessions and 6,303 messages. Now correctly filters to 54 sessions and 5,614 messages.
- **Agy session under-counting** — the old loader only read `transcript.jsonl` (1 of 14 sessions = 57 steps). Now reads `overview.txt` from all 14 sessions (304 steps).
- **SQLite NULL handling** — Codex `state_5.sqlite` has NULL values in `model`, `reasoning_effort`, and `agent_role` columns. Switched to `sql.NullString` to avoid scan errors.

## [v0.1.0] - 2026-06-12

### Added

- Initial open source release.
- GitHub-style contribution heatmap for AI coding tools: MiMoCode, OpenCode, Codex, Agy, command-code, codewhale.
- Leaderboard mode (`thermal`) showing all installed tools ranked by activity.
- Per-tool heatmap mode (`thermal --tool <name>`).
- JSON output (`thermal --json`).
- Configurable weeks range (`thermal --weeks 26`).
- Auto-detect mode (`thermal --tool auto`).
- MIT license.

[Unreleased]: https://github.com/jadmadi/thermal/compare/v0.1.0...HEAD
[v0.1.0]: https://github.com/jadmadi/thermal/releases/tag/v0.1.0
