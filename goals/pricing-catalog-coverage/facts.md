# Facts: Pricing catalog coverage for tool-native model names

## What is uncovered today

The `No pricing for` footer names models the catalog cannot price. On this machine, after model attribution landed, that list grew because Devin's models are now visible:

```
projects --no-color --offline
  No pricing for: accounts/fireworks/models/deepseek-v4-flash-0731,
                  claude-sonnet-5-high, codex-auto-review,
                  glm-5-3-flash-max, +5 more
```

Devin's own ids from its `sessions.model` column:

```
glm-5-2   glm-5-3-flash-max   swe-1-7   kimi-k2-7
claude-sonnet-5-high   swe-1-7-medium   penguin-max   kimi-k3-high
```

Of those, `glm-5-2` resolves against the catalog and the rest do not. `swe-1-7` and `swe-1-7-medium` are Cognition's own model and will never appear in a third-party catalog. `glm-5-3-flash-max` and `kimi-k2-7` look like vendor variants whose naming differs from the catalog, so they are a mapping question rather than a missing price.

The same question applies to `accounts/fireworks/models/deepseek-v4-flash-0731`, which is a provider path rather than a model id.

## Why this matters

An unpriced model is named rather than counted as free, which is honest, but it still understates the reported total. The gap is not visible in the number, only in the footer.

## Architectural Invariants & Constraints
- Recorded cost always wins; this goal only affects days that need an estimate.
- A model with no price must keep appearing in the footer rather than silently contributing zero.
- Prices belong in the catalog or in `~/.config/thermal/pricing.json`, never hardcoded in the loaders or the renderer.
- A mapping from a tool-native id to a catalog id must be explicit and testable, not fuzzy matching that could merge two different models.

## File & Interface Contracts
- `internal/pricing/`: catalog lookup and the user override path at `~/.config/thermal/pricing.json`.
- `internal/loaders/modelname.go`: canonicalisation, which folds case only, not provider prefixes or vendor suffixes.
- Guard: a test for whatever mapping is added, asserting that two distinct models never collapse into one price.

## Open question for the user

Whether to fix this by adding prices for the specific ids, by a documented alias map for vendor variants, or by leaving it and documenting the workflow for a user to add their own prices. The first is data, the second is code, and the third is documentation only.


## Classification result

Measured on this machine after the tier rule landed. The footer dropped from ten unnamed models to six, and the estimated total rose by about $253 because four models began pricing.

| id | class | action |
| --- | --- | --- |
| `claude-sonnet-5-high` | tier variant of `claude-sonnet-5` | fixed by the tier rule |
| `glm-5-3-flash-max` | tier variant of `glm-5-3-flash` | fixed by the tier rule |
| `kimi-k3-high` | tier variant of `kimi-k3` | fixed by the tier rule |
| `swe-1-7` | proprietary, Cognition's own model | stays unpriced; document the override |
| `swe-1-7-medium` | same model, tier variant of an absent base | stays unpriced |
| `codex-auto-review` | a routing or reviewer alias, no public price | stays unpriced |
| `penguin-max` | routing alias seen only in Codex sessions | stays unpriced |
| `kimi-k2-7` | no exact catalog id; `kimi-k2-7-code` is a different SKU | stays unpriced rather than guessed |
| `accounts/fireworks/models/deepseek-v4-flash-0731` | priced, but one codewhale day carries it with unclassified tokens | left unpriced by the no-half-pricing rule |
| `accounts/fireworks/models/deepseek-v4-pro-0813` | priced | resolved |

## Decision

The mechanism is a narrow tier-suffix rule in `lookupCandidates`, plus the documented user override for ids with no public price. Fuzzy matching was rejected: it would fold distinct models onto one price, which is a silent error, while an unpriced model is a visible one.
