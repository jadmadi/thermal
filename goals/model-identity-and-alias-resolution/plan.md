# Plan: Declared Model Identity, Router Alias Normalization & Canonical Catalog

## Execution Steps
- [x] Phase 1 — Schema & Canonical Model Catalog: Define canonical model ID types and alias mappings in internal/loaders/modelname.go.
- [x] Phase 2 — Harness-Scoped Normalization: Implement alias normalization handling router prefixes (openrouter/, deepseek/) and date tags.
- [x] Phase 3 — Unit Testing & Benchmarking: Add comprehensive table-driven tests and zero-alloc benchmark tests in internal/loaders/modelname_test.go.
