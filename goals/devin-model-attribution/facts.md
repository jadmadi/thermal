# Facts: Devin model attribution and honest unestimated reporting

## Architectural Invariants & Constraints
- `thermal.DailyRow.Models` is the only channel the pricer reads. A day with no models is priced at zero and is not reported as unpriced either, so its cost silently disappears from every report.
- The estimator prices only days that name a model: `Catalog.PriceDay` returns `(0, nil)` when `len(day.Models) == 0`. A model it cannot find, or a day with Unclassified tokens, lands in the missing list instead and is named in the footer.
- Token types stay disjoint. A session-level model attribution must add up to the same day total the loader already emits, and a test must assert that rather than trusting the join.
- Model ids pass through `modelName()` before they become map keys, because tools disagree on case and separators.
- The Devin cache records the snapshot shape; `devinCacheVersion` must be bumped so a stale cache is discarded rather than serving days without model attribution.

## Evidence gathered on this machine

Devin's `sessions` table already names the model, and nothing reads it:

```
sqlite> PRAGMA table_info(sessions);
id working_directory backend_type model agent_mode created_at last_activity_at title
main_chain_id shell_last_seen_index cogs_json workspace_dirs hidden metadata

sqlite> SELECT model, COUNT(*) FROM sessions WHERE model != '' GROUP BY model ORDER BY 2 DESC;
glm-5-2|138   glm-5-3-flash-max|11   swe-1-7|5   kimi-k2-7|5
claude-sonnet-5-high|2   swe-1-7-medium|1   penguin-max|1   kimi-k3-high|1
```

A daily join is possible today:

```
SELECT date(s.created_at,'unixepoch','localtime') day, s.model,
       SUM(json_extract(m.chat_message,'$.metadata.metrics.input_tokens'))
FROM message_nodes m JOIN sessions s ON s.id = m.session_id
GROUP BY day, s.model;
```

Each row has `backend_type` and `agent_mode` beside the model, which is enough to
decide whether a model id is a real model or a routing alias.

Scale of what this unlocks: Devin reported 19.2B tokens and 45 active days with
no cost, and 7.76B of them fell inside one 30 day window. Those tokens currently
contribute nothing to any estimate, which is why the projects total reads `$670`
against `$112` recorded.

## File & Interface Contracts
- `internal/loaders/sqlite.go`: the Devin path emits per-model `ModelTokens` on each `DailyRow`, keyed by `modelName(s.model)`.
- `internal/loaders/cache.go`: `devinCacheVersion` is bumped.
- `internal/pricing/pricing.go`: expose what could not be priced in a way a report can print, beyond the existing per-model missing list.
- `internal/render/report.go`, `internal/render/projects.go`: the footer states the tokens that carry no price, so an estimate is never read as complete.
- No change to `--json` shapes beyond additive fields.
