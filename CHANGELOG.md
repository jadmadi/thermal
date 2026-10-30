# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.22.0](https://github.com/jadmadi/thermal/compare/v0.21.0...v0.22.0) (2026-10-30)


### Features

* **cli:** add a glossary and namespace-aware help routing (goals/cli-help-glossary-and-routing/goal.md) ([a2dd941](https://github.com/jadmadi/thermal/commit/a2dd941a967fe471e1c6c5fb952dbe518b2012ab))
* **cli:** add a stat-only source inventory and actionable empty state (goals/cli-first-run-source-inventory/goal.md) ([964cf9e](https://github.com/jadmadi/thermal/commit/964cf9e72158af9813033c249714bab8979a2c30))
* **cli:** add slot-aware unknown-token diagnostics (goals/cli-unknown-token-diagnostics/goal.md) ([9903ca2](https://github.com/jadmadi/thermal/commit/9903ca2764405200a285d6e7b1f9f8a33bc0323a))
* **cli:** honor terminal width with column drops and refusals (goals/cli-terminal-width-adaptation/goal.md) ([a936195](https://github.com/jadmadi/thermal/commit/a936195231f0264ecb36ceb2eb41bc37b187925a))
* **cli:** intercept help globally and add scoped help pages (goals/cli-help-interception-and-scoped-help/goal.md) ([e86a107](https://github.com/jadmadi/thermal/commit/e86a1077bf01ba13a72576045eba575a2f02d303))
* **cli:** make no-data a success state with stable JSON (goals/cli-empty-data-success-contract/goal.md) ([6bd8d1c](https://github.com/jadmadi/thermal/commit/6bd8d1cbd5b9ae7ddba02e7707ff02e734a313e8))
* **cli:** make the leaderboard state its ranking and costs (goals/cli-leaderboard-ranking-and-legend/goal.md) ([17ab978](https://github.com/jadmadi/thermal/commit/17ab978c09c5d815bbc4689cb96e4e8f970d5b6a))
* **loaders:** count and report skipped JSONL records (goals/jsonl-partial-read-observability/goal.md) ([3787548](https://github.com/jadmadi/thermal/commit/3787548e7dd29b1a31b884222769ed4fb6c5e768))
* **loaders:** count Codex rollout partial reads (goals/jsonl-partial-read-observability/goal.md) ([40b2726](https://github.com/jadmadi/thermal/commit/40b2726194afafae878a9c2e1fb69e76de7742d1))


### Bug Fixes

* **changelog:** isolate embedded changelog from CWD and strip goal tokens (goals/changelog-integrity-and-cwd-isolation/goal.md) ([674bca0](https://github.com/jadmadi/thermal/commit/674bca03b3ec3e4d60b1e0f725491d3a773574eb))
* **cli:** align non-TTY tests and diagnostics (goals/cli-nontty-and-statusline-output-contract/goal.md) ([a352e43](https://github.com/jadmadi/thermal/commit/a352e43a811eb2224935bbe2e314f34bcbe44a82))
* **cli:** colocate --db diagnostics and tests (goals/cli-db-override-scope-and-errors/goal.md) ([4d97260](https://github.com/jadmadi/thermal/commit/4d97260b1f2c558cb12047fc1de9b8a5349b6e2a))
* **cli:** colocate total-row helpers and tests (goals/cli-report-total-row-semantics/goal.md) ([20e2ace](https://github.com/jadmadi/thermal/commit/20e2ace9711e6a589b00888f7c49b8ef9164c6e6))
* **cli:** document exit codes and accept global flags on upgrade (goals/cli-automation-exit-and-upgrade-parity/goal.md) ([770a13d](https://github.com/jadmadi/thermal/commit/770a13d33779f02ce96412a0b64a124c67025a0e))
* **cli:** give -v, -V, --version, and --verbose one meaning (goals/cli-version-verbose-flag-contract/goal.md) ([743281a](https://github.com/jadmadi/thermal/commit/743281a5627c72b135e1f656f7068c5b2d452ce5))
* **cli:** make non-TTY live and empty statusline channel-clean (goals/cli-nontty-and-statusline-output-contract/goal.md) ([80a533e](https://github.com/jadmadi/thermal/commit/80a533e96761f5e162fa8f66bb81f5eb6049e224))
* **cli:** make project and model totals additive and explicit (goals/cli-report-total-row-semantics/goal.md) ([d39d43b](https://github.com/jadmadi/thermal/commit/d39d43b5e79ad3501d02c226275eb175ff589878))
* **cli:** make upgrade previewable and non-destructive (goals/cli-upgrade-nondestructive-safety-flags/goal.md) ([c0e4087](https://github.com/jadmadi/thermal/commit/c0e4087c91102848873e2375ceb83a6d11f1bcfe))
* **cli:** make version and deprecation warnings truthful (goals/cli-version-and-deprecation-truth/goal.md) ([88de961](https://github.com/jadmadi/thermal/commit/88de961cce5fbea29dc979a8ac18eda83d76357a))
* **cli:** make window units and repeated flags explicit (goals/cli-window-flag-semantics/goal.md) ([aa3c643](https://github.com/jadmadi/thermal/commit/aa3c6438cc26124fccdc846808ef28bb940b4cd6))
* **cli:** name data-access conditions and add next steps (goals/cli-data-access-error-guidance/goal.md) ([a8a3661](https://github.com/jadmadi/thermal/commit/a8a3661baf98a6470a5bdf9e61d22b8768b731d9))
* **cli:** reject malformed and contradictory arguments (goals/cli-argument-parser-strictness/goal.md) ([8ebde49](https://github.com/jadmadi/thermal/commit/8ebde49ac6617d7937ffc1c32a2e210a718f093d))
* **cli:** scope --db to one tool and translate path errors (goals/cli-db-override-scope-and-errors/goal.md) ([44d1325](https://github.com/jadmadi/thermal/commit/44d132554c98970598cb5864fb071313b8913586))
* **cli:** treat SIGINT as a normal live stop (goals/cli-sigint-lifecycle-contract/goal.md) ([7975021](https://github.com/jadmadi/thermal/commit/79750213c169d00e004c60b968df82f73d90b267))
* **cli:** unify the --top contract and chart remediation (goals/cli-top-flag-contract/goal.md) ([4dbf119](https://github.com/jadmadi/thermal/commit/4dbf119a53f05229e3b889a5c0bd09f67d5959ae))
* **cli:** unify width, colour, and sanitisation across renderers (goals/cli-terminal-output-integrity/goal.md) ([c7775aa](https://github.com/jadmadi/thermal/commit/c7775aaad509e1be69e7b6a61b181ffb1b173768))
* **gate:** isolate output_gallery_gate.sh from gate HOME and git hook env (goals/changelog-integrity-and-cwd-isolation/goal.md) ([ad8009c](https://github.com/jadmadi/thermal/commit/ad8009cbcaf0b5f5da29588a85e0f9966cbea2a7))

## [0.21.0](https://github.com/jadmadi/thermal/compare/v0.20.1...v0.21.0) (2026-10-29)

### Features

* **pricing:** add cost provenance reconciliation, model pricing disclosures, and subscription usage labeling
* **loaders/muse:** add Muse token and project day ingestion from session JSONL records
* **loaders/sqlite:** enforce schema object validation, 256MB mmap limits, 60s query timeouts, and 64MB store ceilings

### Bug Fixes

* **test:** normalize Windows project path assertions and handle read-only file ACLs in audit tests
* **ci:** bump Go baseline to 1.27.3 to resolve crypto/x509 GO-2026-6809 in govulncheck

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

* **build:** verify multi-target cross-compilation release gate ([0efc97b](https://github.com/jadmadi/thermal/commit/0efc97b89ebf630046b0582cadc9f206f0eecf06))
* **ci:** verify cross-platform operating system test matrix ([70143cb](https://github.com/jadmadi/thermal/commit/70143cb1e136afa996437d24099928bc4d478f53))
* **docs:** update documentation parity and simulated human user audit for 15 tools ([135b22e](https://github.com/jadmadi/thermal/commit/135b22ef673c49bfea8a32c646ff0fa0d7b13200))
* **gate:** ensure hermetic verification and release gate semantic parity ([1f71cd2](https://github.com/jadmadi/thermal/commit/1f71cd20bd10c90faf05dcabc22c16bcf5423f2b))
* **git:** verify zero bot contributor attribution in git history ([cebda93](https://github.com/jadmadi/thermal/commit/cebda9375d156912adf81a950f5a5ec84aa2455c))
* **loaders:** ingest Zed editor assistant threads and tokens ([9b6a3ef](https://github.com/jadmadi/thermal/commit/9b6a3ef15721bfbbebec4b537a4ebbdc84f00476))
* **sila:** complete Sila onboarding and strict verification ([8e3d3e9](https://github.com/jadmadi/thermal/commit/8e3d3e9afc72d6f3049f316b5e235795c85f068f))

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

* **audit:** add repository AGENTS.md readiness and agent hygiene check ([259d1e0](https://github.com/jadmadi/thermal/commit/259d1e0b408a11f160ddd956dc6220bbfe11b13d))
* **cli:** add instant shell statusline and prompt composable helper ([26a20ae](https://github.com/jadmadi/thermal/commit/26a20ae6c7dbf6c66ac2df486ad41d091dd060a0))
* **models:** add reasoning effort breakdown & cognitive intensity telemetry ([0f3f533](https://github.com/jadmadi/thermal/commit/0f3f533f91307842a0461583e519ac4b01359c03))
* **receipt:** extract Codex sandboxed test runs and exit codes ([198772e](https://github.com/jadmadi/thermal/commit/198772efb99c05aea1eb00864adfb08c0ebbfe63))
* **render:** add PR-ready markdown output for receipt and yield ([c00ef02](https://github.com/jadmadi/thermal/commit/c00ef02ff312df8df8bdfa8ad01af79135aefa13))
* **replay:** add ChatGPT 5-hour rolling burst modeling ([b0759e0](https://github.com/jadmadi/thermal/commit/b0759e00379cd9efc4c0d5b519957ce726a270ee))
* **yield:** analyze session fork lineage and branch churn in code delta yield ([56ec02b](https://github.com/jadmadi/thermal/commit/56ec02bf8efeb8b952eb4ea7fbde80d4f2e8ae4c))


### Bug Fixes

* **upgrade:** increase asset download timeout to 120s for slower networks ([25b9b5f](https://github.com/jadmadi/thermal/commit/25b9b5fcf606f27868422ec1a74bce554847bfe8))

## [0.15.0](https://github.com/jadmadi/thermal/compare/v0.14.0...v0.15.0) (2026-09-25)


### Features

* **analytics:** shared chart analytics, distribution stats and trend projections ([a059d01](https://github.com/jadmadi/thermal/commit/a059d0190a5f99b7ed5ec5efa231d91fe8e1ee62))
* **changelog:** add interactive changelog command with top filtering and JSON export ([bacde98](https://github.com/jadmadi/thermal/commit/bacde988c6bad71f7a19f6db2b6dfb412ebe48c5))
* **chart:** support --chart flag across reports, document dashboard and conventions ([25163db](https://github.com/jadmadi/thermal/commit/25163dbac85f1ea23d03c89bbc5eea307822b2ab))
* **check:** introduce fast-path pre-commit mode with package tests & diff sentry ([1b5a3c8](https://github.com/jadmadi/thermal/commit/1b5a3c88bb640c146c1a41236ab4636073cee6a1))
* **ci:** add OpenSSF SLSA Level 3 build provenance and verification runbook ([78b8258](https://github.com/jadmadi/thermal/commit/78b82584a94f2e64ebec1e40c239892647d46d9d))
* **cli:** add staged deprecation warnings & upgrade migration alerts ([2842f1a](https://github.com/jadmadi/thermal/commit/2842f1a839026349543a10d6a2172b62b20ecef3))
* **cli:** add stateless zero-database url sharing engine ([7557bf9](https://github.com/jadmadi/thermal/commit/7557bf99aceef2d40792072847c44e36bc33c993))
* **cli:** add thermal audit command for local setup and context health ([8afd798](https://github.com/jadmadi/thermal/commit/8afd798a12491a19e7f572160f511241e9902fec))
* **cli:** add thermal receipt command for verifiable work outcomes ([ebc550a](https://github.com/jadmadi/thermal/commit/ebc550afe6ee44437d2e6d019af1ba3f0066a453))
* **cli:** add thermal serve embedded local web dashboard and telemetry APIs ([dceba19](https://github.com/jadmadi/thermal/commit/dceba19f4aa0651c1b7c5443f54046ba2cb14711))
* **cli:** add thermal yield command for code output delta telemetry ([ef10ceb](https://github.com/jadmadi/thermal/commit/ef10cebcf9f112c1f8fa23d7009743716067ee50))
* **contributing:** establish contributor onboarding path, preflight check & retention loop ([3bd8d0a](https://github.com/jadmadi/thermal/commit/3bd8d0a3e92d9de0c78c8dc34a0c15b284ed1d04))
* **devin:** project sessions.model onto DailyRow and ProjectDay for accurate model attribution ([ecc9a92](https://github.com/jadmadi/thermal/commit/ecc9a920ec460ea8e3035ca0665680bef168f886))
* **distribution:** configure homebrew tap, installer script, agent docs & distribution runbook ([80b3e4f](https://github.com/jadmadi/thermal/commit/80b3e4fe566c7b106d14d868615e363ce9dbe2b9))
* **docs:** establish canonical migration guide & governance gate ([12d9833](https://github.com/jadmadi/thermal/commit/12d98338458d89935843468c03df287be753cd22))
* **governance:** establish Written Solo Authority model and scope defense ([332ad26](https://github.com/jadmadi/thermal/commit/332ad262b176f427d2a3535c7db686ba082fa41d))
* **governance:** implement staged trust ladder, succession runbook & absence factor checks ([9fa6245](https://github.com/jadmadi/thermal/commit/9fa62450d31d8814910e5cd05ce857d3d8665647))
* **license:** add Community Reciprocity Covenant and license CLI metadata ([9a82384](https://github.com/jadmadi/thermal/commit/9a8238472f7d5c2b51d455117143bc3681cf972d))
* **license:** AGPL-3.0 relicensing, commercial dual-licensing & contributor gating ([eb9aeb4](https://github.com/jadmadi/thermal/commit/eb9aeb4e79cf31307d979813e5c2a47ed6d35dee))
* **license:** harden licensing architecture, CI enforcement & SPDX headers ([98772b5](https://github.com/jadmadi/thermal/commit/98772b519955e1ec8ec0a0dbde1a8897dbfa98f3))
* **live:** attribute resumed session projects, seed today's baseline, and track activity units ([f463509](https://github.com/jadmadi/thermal/commit/f46350990e48357b9d3b26c508de50108402b552))
* **loaders:** add DeepSeek harness (dsh) loader and registry ([9be2513](https://github.com/jadmadi/thermal/commit/9be2513404ae96090832cc4feef74259c65fc179))
* **loaders:** add Nous Hermes agent ingestion loader for state.db ([ad2d523](https://github.com/jadmadi/thermal/commit/ad2d5238e2866fade6b359adb7a8704a8d6964ef))
* **loaders:** report truncated JSONL scans via Summary.Warnings to verbose stderr ([17b6b59](https://github.com/jadmadi/thermal/commit/17b6b595f3fca038b5d20ec32bf78bdba0b9c8c2))
* **loaders:** support unified Antigravity hub and OpenCode model breakdowns ([2afb9f8](https://github.com/jadmadi/thermal/commit/2afb9f8d1ec5bc87524783f2b5f98e1216d28546))
* **models:** declare canonical model identities and router alias normalization ([e460d4a](https://github.com/jadmadi/thermal/commit/e460d4a769928c28117e2b8b7908be79ceb2bf78))
* **pricing:** catalog coverage and tier suffix lookup for tool-native model variants ([79bbd75](https://github.com/jadmadi/thermal/commit/79bbd7543fc59c4a5d5df841d9e3e420bb01c672))
* **release:** adapt changelog-writing standards and configure release-please sections ([7b0813c](https://github.com/jadmadi/thermal/commit/7b0813ccd44e14d2dd5056f2f5907740dd5f7d78))
* **release:** script legacy release retirement & document AGPL-3.0 release line ([cee5b3b](https://github.com/jadmadi/thermal/commit/cee5b3bd8e74d0bb4fb49118b81c437517d1ce76))
* **render:** establish unified terminal design tokens, adaptive cards, and text wrapping ([a8e65d7](https://github.com/jadmadi/thermal/commit/a8e65d774952b1e656183a59b4be1d4645f7272f))
* **replay:** simulate historical developer workloads against subscription limits and model cards ([1ca92ed](https://github.com/jadmadi/thermal/commit/1ca92edc4c508c65679e75fb7950f2c89b3bafcc))
* **repo:** establish code ownership, maintainers, security policy & PR lifecycle ([bee76ff](https://github.com/jadmadi/thermal/commit/bee76ff55e8a2e1d6ef384d24bf9c30bb0249b36))
* **report:** verify weekly and monthly period totals and table formatting parity ([b8fac89](https://github.com/jadmadi/thermal/commit/b8fac89012194436ef29dbde37fb1e5dcc4b0f83))
* **server:** make web totals match canonical reports and support collection options ([d2cd4a2](https://github.com/jadmadi/thermal/commit/d2cd4a2e67cc803f19291749ad151496c494802b))
* **server:** reject untrusted host authorities and cross-site requests at telemetry API ([d162bf1](https://github.com/jadmadi/thermal/commit/d162bf1f679b436e6b9cab2c158c46705b536765))
* **stats:** expose disjoint token composition and cache hit rate metrics ([fa4c8f9](https://github.com/jadmadi/thermal/commit/fa4c8f9c0de80b6c5b92cebb61b590598b23b60b))
* **tui:** add high-density 9-box finops grid and activity taxonomy view ([f720e57](https://github.com/jadmadi/thermal/commit/f720e5798584a1e4360059243dd98bfa6df472ba))
* **tui:** bubble tea v2 application shell and overview dashboard ([85af7be](https://github.com/jadmadi/thermal/commit/85af7bef073805878594fbadb14611f6bdfc07ac))
* **tui:** elevate codeburn-inspired 9-box FinOps grid & dashboard dense view ([88c24d5](https://github.com/jadmadi/thermal/commit/88c24d510ae8bab65a32ec190f399fd022bbcf36))
* **tui:** interactive projects ranking view and drill-down detail ([50faef3](https://github.com/jadmadi/thermal/commit/50faef37cace029df5f9b3d6284708715324cf79))
* **tui:** interactive statistical distribution, weekday cadence and outlier view ([be1e45f](https://github.com/jadmadi/thermal/commit/be1e45fed317b7dfd1db2b709190448d431f8281))
* **tui:** interactive tool and model mix share views with switching metrics ([43c146e](https://github.com/jadmadi/thermal/commit/43c146e37b2ae87872bea77da5eb51f307e538a4))
* **tui:** spike ntcharts v2 against stock charm stack and record findings ([d1b68ab](https://github.com/jadmadi/thermal/commit/d1b68ab84f9e297fdfd81edb688ff4dccfc1ae91))


### Bug Fixes

* **claude:** deduplicate assistant messages by message id and report in summary warnings ([d39ef40](https://github.com/jadmadi/thermal/commit/d39ef406b54d0196ac716ee5443b45b103c02195))
* **live:** align token burn, turn counts, and cache rates with canonical metrics ([305f589](https://github.com/jadmadi/thermal/commit/305f589714a8b40ef8cf0a870ad2b3ab84422ed8))
* **live:** decouple display cadence and cache unchanged sources to bound 100ms CPU ([b71d464](https://github.com/jadmadi/thermal/commit/b71d46487fd4f1a535aed1f6b4d0a2a047ca58cf))
* **loaders:** enforce SQLite rows iteration error inspection and document standard ([c53c32b](https://github.com/jadmadi/thermal/commit/c53c32b0b241ceaeb3b52707b059d46d931c7ad8))
* **loaders:** ignore empty WAL and volatile SHM locks in Devin source identity ([31b7b84](https://github.com/jadmadi/thermal/commit/31b7b84d1a23ff12ea5ab1d3ce5edcf18fdf8b25))
* **loaders:** isolate Devin cache by source database and validate freshness ([d182cf9](https://github.com/jadmadi/thermal/commit/d182cf9e0f7134733ce94b92725e19a3115087a0))
* **loaders:** raise JSONL scanner line ceiling to 32 MiB for large transcripts ([86d6ebb](https://github.com/jadmadi/thermal/commit/86d6ebb1d2624e9b3c8efd7ae6ba7d6d32cfe810))
* **receipt:** preserve recorded usage, deduplicate messages, and disclose partial coverage ([dd3ab4f](https://github.com/jadmadi/thermal/commit/dd3ab4f987b51f7fe6868af1a70164d1625a4455))
* **receipt:** require observed command results and drop uncompleted invocations ([7a8ce91](https://github.com/jadmadi/thermal/commit/7a8ce91cd22fd928cccd8a75d1470bea36591707))
* **server:** bypass local disk presence check when synthetic loader is injected ([8e47085](https://github.com/jadmadi/thermal/commit/8e47085a1750a54bb8ffbb46f95a32af6c167628))
* **telemetry:** audit pricing estimates, dsh tool naming, isActivityOnly filtering, and stats UI ([8c9847d](https://github.com/jadmadi/thermal/commit/8c9847db412ee0669f5257e153ec888220487722))
* **thermal:** guard streak computation against malformed dates and invalid calendar tokens ([a2467d7](https://github.com/jadmadi/thermal/commit/a2467d79c51d6c35bb89c9726c39e008b0e25eba))
* **thermal:** separate activity steps from token totals and fix model attribution ([09eb367](https://github.com/jadmadi/thermal/commit/09eb367a96fc928ea6f8fada48a8fb09eb3cafcc))
* **upgrade:** check file close errors and deduplicate archive extraction ([15d2ffb](https://github.com/jadmadi/thermal/commit/15d2ffb0a8eccfb5f40ffc31c1cd9b9549fc0cc4))


### Performance Improvements

* **loaders:** add incremental rollout transcript caching and delta parser for codex ([752569e](https://github.com/jadmadi/thermal/commit/752569efa1dec1d86c802d4809ff42c5e80923d6))
* **loaders:** establish end-to-end ingestion performance baseline and benchmark suite ([e2a569f](https://github.com/jadmadi/thermal/commit/e2a569f035ce3405f06bd2c2a3e3e653d41bdf98))
* **render:** preallocate leaderboard slices capacity and add allocation benchmark ([f375fdf](https://github.com/jadmadi/thermal/commit/f375fdf4a93adfed1375dac534ffa002e19fb997))

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

* **replay:** implement thermal replay simulation against subscriptions and APIs (`thermal replay`) ([4f7ff02](https://github.com/jadmadi/thermal/commit/4f7ff02f5e7710deabc5454265330bcbee35057a))
* **testing:** add automated simulated user release gate and update gitignore ([b5651df](https://github.com/jadmadi/thermal/commit/b5651df195f1702078d233ba9fb3d4c5ea98c29c))


### Bug Fixes

* **testing:** ensure simulated user tests use dynamic timestamps and isolated home environment ([2c578a0](https://github.com/jadmadi/thermal/commit/2c578a0fd86e035939805ecd1a7612005d70502e))

## [0.11.0](https://github.com/jadmadi/thermal/compare/v0.10.0...v0.11.0) (2026-09-21)


### Features

* **analytics,render:** expose disjoint token composition and cache efficiency metrics ([6e070bd](https://github.com/jadmadi/thermal/commit/6e070bd87108f8b6994d5adaeba7e5cff1d4cc48))
* **loaders:** add Nous Hermes agent loader and registry ([9e90d01](https://github.com/jadmadi/thermal/commit/9e90d017e24a4a4cdd04bcdd9b32f166a824b941))
* **loaders:** add prompt-level turn deduplication to Grok loader ([e589226](https://github.com/jadmadi/thermal/commit/e5892265a86dd65f07b7ab1dfb5df6e34543bab8))
* **loaders:** implement message-level deduplication in Claude Code loader ([104ba23](https://github.com/jadmadi/thermal/commit/104ba23a39bdaacfe4d16e9e34d4dbbe2ed61da8))

## [0.10.0](https://github.com/jadmadi/thermal/compare/v0.9.1...v0.10.0) (2026-09-21)


### Features

* **docs:** polish layout, motion choreography, and window chrome in html docs ([d9ab799](https://github.com/jadmadi/thermal/commit/d9ab7997b7c5e56be42354f858ac69f72ebc3161))
* **docs:** switch canonical site domain to thermal.jadmadi.net ([a1c9c4d](https://github.com/jadmadi/thermal/commit/a1c9c4dd5002a62b4954ee3a820a5817b9d0fc43))
* **loaders:** add DeepSeek harness (dsh) loader and registry ([414a815](https://github.com/jadmadi/thermal/commit/414a8151615119ca3f043a8beafe8a67f9bfb194))
* **loaders:** report JSONL scans that stop early via Summary.Warnings ([247bb80](https://github.com/jadmadi/thermal/commit/247bb80b2a49ceb455beb3cb76d650a5f43bb8f1))
* **pricing:** map tier variants to the base model, and classify the rest ([b361693](https://github.com/jadmadi/thermal/commit/b361693abb24a963f62b5fc0769288dd9322b5ac))
* **render,analytics:** telemetry accuracy, pricing estimates, and stats UI polish ([662f25a](https://github.com/jadmadi/thermal/commit/662f25a3e9c761f4369cd20425e75c268299dbee))
* **sila:** complete sila onboarding and wire repository ([92ea6ee](https://github.com/jadmadi/thermal/commit/92ea6ee123cb044818dab401ce450f59e83617d3))


### Bug Fixes

* **loaders:** raise JSONL line ceilings to 32 MiB ([ada5b52](https://github.com/jadmadi/thermal/commit/ada5b526722e6f20b3b8281a65b574d616ab7170))

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
* **docs:** add author links, schema, FAQ, and crawler files to the docs site ([ccc7bc1](https://github.com/jadmadi/thermal/commit/ccc7bc1226aba6dc183de09f35dfa4358d31d2ab))
* **docs:** docs site, corrected report totals, and Pages publishing ([7a470d1](https://github.com/jadmadi/thermal/commit/7a470d15c49c290cbe27f9b631e62d11185a60e6))
* **reports:** --chart, dashboard docs, and the rules ([656db32](https://github.com/jadmadi/thermal/commit/656db32151827d2a6afa84160dd5d49201d40574))
* **reports:** add --chart, document the dashboard, and record the rules ([1cb2f50](https://github.com/jadmadi/thermal/commit/1cb2f501825abb429f30940eb3da20e8fa715567))
* **reports:** keep token-bearing periods only, grain-independent totals ([9c27be4](https://github.com/jadmadi/thermal/commit/9c27be46c625036754c9637b447bbe85d43372fd))
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
* **pages:** move the JSON reference inside the reports section and reindent it ([c9c6af8](https://github.com/jadmadi/thermal/commit/c9c6af80892ec4f8fa0e1a284db251ee1e7f3abe))
* **tools:** drop the ambiguous cc alias and point tool links at their sites ([2a81533](https://github.com/jadmadi/thermal/commit/2a81533383ea852f0f274f205a2aa47133a86c02))
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

* **agy:** read live antigravity-cli transcripts for step activity ([#4](https://github.com/jadmadi/thermal/issues/4)) ([c7f4fa6](https://github.com/jadmadi/thermal/commit/c7f4fa631d5020dc2647647bfd26b3ae5ea3d3))

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
