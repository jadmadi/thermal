# Facts: Support DeepSeek Harness (dsh) in Thermal

## Architectural Invariants & Constraints
- **Read-Only Data Ingestion**: Open and read files in strictly read-only mode (`os.ReadFile` or read-only handles). Never write to `~/.dsh`.
- **Zero Allocation & Bounded Worker Pools**: Multi-file directory scan over `session-*.json` must be bounded or efficient; pre-aggregated files are fast to decode.
- **Disjoint Token Types**:
  - `uncachedInputTokens` maps to `Input`
  - `outputTokens` maps to `Output`
  - `cacheReadTokens` maps to `CacheRead`
  - `cacheWriteTokens` maps to `CacheWrite`
  - All disjoint token types add up to `TotalTokens`.
- **Model Normalization**: Pass model ids through `modelName()` for canonical keys (e.g. `deepseek-v4-flash`).
- **Project Attribution**: Use `thermal.ProjectKey(identity.Cwd)` to resolve nearest git root. Symlinks resolved first.
- **Robust Time Parsing**: Parse `identity.CreatedAt` as Unix millisecond timestamp (`time.UnixMilli`). Handle RFC3339 or second timestamps if string or different unit.
- **Safe Nil/Null Handling**: In `record.rows`, handle cases where `modelSelection.val` is null/nil, `lastUsed` is null/nil, or `tokenUsage.val` is null/nil.
- **Category**: `dsh` belongs to **Token Warriors** because it records granular token consumption.

## Source Locations
- Default root: `~/.dsh` (or `DSH_HOME` env)
- Primary sessions directory: `~/.dsh/storages/session_projcache/sessions/*.json`
- Fallback monolithic cache: `~/.dsh/storages/session_projcache.json` (`tables.sessions`)
- Fallback transcript directory: `~/.dsh/sessions/*/*/session*.jsonl.zstd`

## Pre-Deploy & Pre-Release Checklist
1. **Unit & Edge Case Tests**: Mock session files with zero-tokens, nil modelSelection, missing cwd, corrupt JSON, legacy single-file fallback.
2. **Real Data Smoke Test**: Run `thermal --tool dsh` against `~/.dsh` and verify live terminal output.
3. **Race Detection**: Run `go test -v -race ./...` across all packages.
4. **Gofmt Scope**: Run `gofmt` only on touched files.
5. **Linter Gate**: Run `golangci-lint run --disable-all -E gofmt,ineffassign,misspell,revive`.
6. **Build Verification**: Run `go build -o /tmp/thermal-test ./cmd/thermal`.
7. **Documentation Parity**:
   - `README.md`: Update tool count (13 tools), add `dsh` to Token Warriors table, add aliases.
   - `AGENTS.md`: Add dsh quirks to Section 2.D.
   - `docs/pages/index.html`: Update tool count, table rows, hero stats, and FAQ schema.
   - `docs/pages/404.html`: Synchronize with index.html to maintain 100% parity.
8. **Autonomous Local Commit**: Commit with Conventional Commits linking goal.md `(goals/dsh-loader/goal.md)`.
9. **Verification Loop Closure**: Run `sila goals` to confirm goal implementation.
10. **Prompt User Before Remote Push**: Remote push requires explicit consent.
