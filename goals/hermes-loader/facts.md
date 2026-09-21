# Facts: Nous Hermes Agent Loader

1. **Tool Identity & Scope**:
   - **Name**: Nous Hermes (`ToolHermes = "hermes"`)
   - **Harness**: Nous Research's persistent local agent.
   - **Storage Engine**: SQLite 3.
   - **Default Path**: `~/.hermes/state.db` (honors `HERMES_HOME` env or default `~/.hermes`).
2. **Schema Invariants**:
   - Must open with URI read-only flag: `?mode=ro`.
   - Must apply `PRAGMA mmap_size=268435456` immediately upon opening.
   - Support table probing defensively to handle version schema migrations gracefully.
3. **Attribution & Disjoint Tokens**:
   - Working directories mapped through `thermal.ProjectKey(cwd)` to resolve symlinks and identify git repository boundaries.
   - Token columns mapped to disjoint buckets: `Input`, `Output`, `Reasoning`, `CacheRead`, `CacheWrite`.
   - Models normalized via `modelName()` for canonical case-insensitive pricing and aggregation.
