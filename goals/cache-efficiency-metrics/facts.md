# Facts: Cache Efficiency & Disjoint Token Breakdown

1. **The 28.7 Billion Token Paradox**:
   - In multi-turn AI coding tools, agents resend conversation history, repo file trees, and tool schemas on every turn.
   - For an agent session with 50 turns and 150k context, the session consumes `50 * 150k = 7.5M` tokens, even though the repo is only 100k tokens.
   - Analysis of local developer histories (e.g. StackReplay across 96k events) shows:
     - 98.14% Cache Read
     - 0.98% Cache Write
     - 0.53% Uncached Input
     - 0.31% Output
     - 0.04% Reasoning
2. **Current Thermal Token Modeling**:
   - `internal/thermal/types.go` has `ModelTokens` with `Input`, `Output`, `Reasoning`, `CacheRead`, `CacheWrite`, and `Unclassified`.
   - AGENTS.md Rule A.6 strictly mandates that these 5 token buckets are disjoint and add up to total tokens.
   - Loaders for Codex, Grok, OpenCode, Claude, and DeepSeek already populate these disjoint buckets.
3. **Current User-Facing Gap**:
   - In `thermal` CLI output, tables and summaries display only a single total `Tokens` column.
   - Users cannot see their cache hit rate or understand why their token counts reach into millions or billions.
4. **Cache Hit Rate Formula**:
   - `PromptTokens = Input + CacheRead + CacheWrite`
   - `CacheHitRate = CacheRead / PromptTokens` (when `PromptTokens > 0`)
