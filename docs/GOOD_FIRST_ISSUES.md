# Good First Issues — Thermal (`thermal`)

Welcome! If you are looking to contribute to Thermal, this document catalogues curated starter issues designed for newcomers. Each issue has a clear scope, concrete file pointers, test instructions, and a designated maintainer mentor.

To claim an issue, leave a comment on the corresponding GitHub Issue requesting assignment. Please review [CONTRIBUTING.md](../CONTRIBUTING.md) and [GOVERNANCE.md](../GOVERNANCE.md) before starting.

---

### Issue #1: `feat(loaders)`: Add Cline / Roo Code Agent Loader

- **Context & Problem Statement**:
  Cline (formerly Claude Dev) and Roo Code are popular open-source autonomous coding extensions for VS Code. They store task execution history and token usage locally in VS Code global storage (`~/.config/Code/User/globalStorage/rooveterinaryinc.roo-cline/` or `saoudrizwan.claude-dev`). Thermal currently does not ingest their sessions into the leaderboard.
- **Expected Behavior**:
  - Implement a new loader package in `internal/loaders/cline.go` reading local Cline/Roo Code JSON session tasks.
  - Extract disjoint token counts (`Input`, `Output`, `Reasoning`, `CacheRead`, `CacheWrite`) and model identifiers.
  - Normalize model identifiers via `modelName()`.
  - Open files in read-only mode and use bounded worker pools if scanning directories.
  - Wire the loader into `internal/loaders/registry.go` and add short CLI alias `cline` or `roo`.
- **Relevant Files & Functions**:
  - `internal/loaders/cline.go`: Loader ingestion logic.
  - `internal/loaders/registry.go`: Tool registration and alias resolution.
  - `internal/loaders/cline_test.go`: Unit tests with mock JSON fixtures.
- **How to Test & Verify**:
  ```bash
  go test -v -race ./internal/loaders -run TestCline
  ./scripts/check.sh
  ```
- **Prerequisites & Required Skills**:
  - Familiarity with Go JSON parsing and struct decoding.
  - Understanding of Thermal's disjoint token model (see `AGENTS.md`).
- **Maintainer Mentor**: **Jad Madi** ([@jadmadi](https://github.com/jadmadi))

---

### Issue #2: `test(loaders)`: Benchmark Fixture Suite for Grok and Muse Loaders

- **Context & Problem Statement**:
  Thermal requires sub-10ms invocation speed for cached reads and zero allocations on critical hot paths. While `dsh_test.go` and `sqlite_test.go` have microbenchmarks, the Grok JSONL parser (`internal/loaders/grok.go`) and Muse SQLite index loader (`internal/loaders/muse.go`) lack dedicated Go benchmarks.
- **Expected Behavior**:
  - Add benchmark tests `BenchmarkGrokScan` and `BenchmarkMuseScan` in `internal/loaders/loaders_bench_test.go`.
  - Use realistic mock JSONL transcripts (1,000 lines) and in-memory SQLite tables with `testing.B`.
  - Assert zero heap allocation regressions on warm cached scans using `b.ReportAllocs()`.
- **Relevant Files & Functions**:
  - `internal/loaders/grok.go`: Grok scanner logic.
  - `internal/loaders/muse.go`: Muse SQLite index logic.
  - `internal/loaders/loaders_bench_test.go`: New or extended benchmark suite.
- **How to Test & Verify**:
  ```bash
  go test -benchmem -bench='Benchmark(Grok|Muse)' ./internal/loaders
  ./scripts/check.sh
  ```
- **Prerequisites & Required Skills**:
  - Go benchmarking (`testing.B`, `b.ReportAllocs()`).
  - Familiarity with Go memory profiling.
- **Maintainer Mentor**: **Jad Madi** ([@jadmadi](https://github.com/jadmadi))

---

### Issue #3: `docs(contributing)`: Author Step-by-Step Guide for Adding New Agent Loaders

- **Context & Problem Statement**:
  While `AGENTS.md` defines technical rules for loaders, contributors often ask for a beginner-friendly walkthrough explaining how to take a raw agent database or transcript directory and build an ingestion engine step-by-step.
- **Expected Behavior**:
  - Create `docs/ADDING_A_LOADER.md` providing a practical walkthrough:
    1. Inspecting local agent data format (SQLite vs. JSONL vs. JSON).
    2. Registering the tool name and short aliases.
    3. Implementing `LoadOrScanWithCache` delta-caching.
    4. Mapping disjoint tokens and normalizing models via `modelName()`.
    5. Writing mock unit tests and adding fixture verification.
- **Relevant Files & Functions**:
  - `docs/ADDING_A_LOADER.md`: New documentation guide.
  - `CONTRIBUTING.md`: Add cross-reference to the new guide.
- **How to Test & Verify**:
  ```bash
  ./scripts/check.sh
  ```
- **Prerequisites & Required Skills**:
  - Technical writing in clear, concise GitHub-flavored Markdown.
  - Basic understanding of Thermal's loader architecture.
- **Maintainer Mentor**: **Jad Madi** ([@jadmadi](https://github.com/jadmadi))
