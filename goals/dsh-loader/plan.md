# Plan: Support DeepSeek Harness (dsh) in Thermal

## Execution Steps

- [x] **Phase 1 — Schema & Core Loader Implementation**:
  - Add `thermal.ToolDsh Tool = "dsh"` in `internal/thermal/types.go`.
  - Implement `LoadDshData(dataDir string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error)` in `internal/loaders/dsh.go`.
  - Parse session files in `storages/session_projcache/sessions/*.json` and fallback to `storages/session_projcache.json`.
  - Map `uncachedInputTokens`, `outputTokens`, `cacheReadTokens`, `cacheWriteTokens`, `modelSelection`, `turns`, `cwd`, and millisecond `createdAt`.
  - Register `ToolDsh` in `internal/loaders/registry.go` with aliases (`dsh`, `deepseek`, `deepseek-harness`).

- [x] **Phase 2 — Unit & Edge Case Testing**:
  - Implement `internal/loaders/dsh_test.go` covering:
    - Normal multi-session directory.
    - Legacy single-file `session_projcache.json`.
    - Zero-token sessions and nil `modelSelection.val` / `lastUsed`.
    - Malformed / corrupted JSON files.
    - Missing directory error handling.
    - Project attribution resolution with `thermal.ProjectKey`.

- [x] **Phase 3 — Verification & Live Smoke Test**:
  - Run `go test -v -race ./...`.
  - Build test binary `go build -o /tmp/thermal-test ./cmd/thermal`.
  - Run live smoke test against local `~/.dsh` (`/tmp/thermal-test --tool dsh --verbose`).
  - Verify formatting and linter pass (`gofmt`, `golangci-lint`).

- [x] **Phase 4 — Documentation Updates**:
  - Update `README.md`: 13 tools supported, add dsh to Token Warriors table, document aliases (`dsh`, `deepseek`).
  - Update `AGENTS.md`: Document `dsh` quirks in Section 2.D.
  - Update `docs/pages/index.html`: Bump tool counter from 12 to 13, add dsh to table and JSON-LD FAQ/software schema.
  - Update `docs/pages/404.html`: Maintain 100% parity with `index.html`.

- [x] **Phase 5 — Goal Closure & Pre-Release Checklist Execution**:
  - Execute pre-release checklist.
  - Commit atomically with `feat(loaders): add DeepSeek harness (dsh) loader and registry (goals/dsh-loader/goal.md)`.
  - Run `sila goals` to confirm automated goal attribution.
