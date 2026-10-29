# Evidence: Price classified tokens on partial days, resolve free-tier suffixes, and refresh model catalog

Date: 2026-10-29
Project: thermal-streak
Goal: pricing-catalog-and-partial-model-provenance

## Changes

- `internal/pricing/pricing.go`:
  - `Catalog.Lookup`: added resolution for `-free` and `:free` suffixes as zero-cost tiers (`Price{}, true`), preventing false missing-price warnings for models such as `muse-spark-1.3-contributor-free`, `nemotron-3-ultra-free`, and `x-preview-f-free`.
  - `lookupCandidates`: strips `-free` suffix so base models with list prices resolve.
  - `Catalog.PriceDay`: prices classified tokens whenever `classified > 0`, reporting a model in `MissingPricing` only if it is missing from the catalog or has zero classified tokens (`classified == 0`).
  - `applyMuseCatalog`: layered Meta Muse's local `model-catalog/*.json` (and default Muse catalog rates) into `pricing.Load` before user overrides.
- `internal/pricing/modelsdev.go`:
  - Added variant aliases for `deepseek-v4.1-flash`, `deepseek-v4-flash`, `deepseek-v4-flash-0731`, `deepseek-v4-pro-0813`, `deepseek-flash`, `gpt-6-astra`, `codex-auto-review`.
- `internal/pricing/models.go`:
  - Registered canonical model identities for `muse-wonder-1.0`, `muse-spark-1.3`, `muse-spark-1.3-contributor`, `muse-spark-1.2`, `muse-spark-1.2-contributor`.
- `internal/pricing/subscription.go` and `internal/pricing/subscription_test.go`:
  - Archived dead replay simulation code to `rm/`.
- `internal/pricing/pricing_test.go`:
  - Added `TestPriceDayPartialClassifiedTokens` and `TestLookupFreeTierSuffix`.

## Verification

```text
$ CGO_ENABLED=1 go test -race ./internal/pricing/... -v
--- PASS: TestLookupCandidates (0.00s)
--- PASS: TestLookupPrefersOriginalName (0.00s)
--- PASS: TestPriceFallbacks (0.00s)
--- PASS: TestPriceDayTypesAndMissing (0.00s)
--- PASS: TestPriceDaySkipsUnclassifiedAndEmpty (0.00s)
--- PASS: TestPriceDayPartialClassifiedTokens (0.00s)
--- PASS: TestLookupFreeTierSuffix (0.00s)
--- PASS: TestReadCacheRejectsOldVersion (0.00s)
--- PASS: TestOverridesWin (0.00s)
--- PASS: TestBuildCatalogProviderPreference (0.00s)
--- PASS: TestBuildCatalogSkipsPlanOnlyModel (0.00s)
--- PASS: TestIsPlanProvider (0.00s)
--- PASS: TestLoadWithMissingCacheAndOverrides (0.00s)
--- PASS: TestBuildCatalogAlias (0.00s)
--- PASS: TestLookupTierSuffix (0.00s)
--- PASS: TestFindModelIdentity (0.00s)
PASS
ok  	github.com/jadmadi/thermal/internal/pricing	1.008s
```

Verdict: PASS

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "pricing-catalog-and-partial-model-provenance",
  "project": "thermal-streak",
  "prime": {"ran": true, "command": "sila prime"},
  "recall": [{"query": "model pricing unclassified tokens and free tier suffix", "matches": 3}],
  "memos": [{"id": "fact_b44066b4", "text": "Price classified tokens on partial days even when Unclassified > 0, and resolve -free / :free suffixes as $0.00."}],
  "lessons": [{"id": "fact_cce022f6", "symptom": "Models with partial unclassified tokens were discarded entirely into MissingPricing by PriceDay"}],
  "cases": [{"id": "fact_e1659a15", "intent": "Resolve free-tier model suffixes and price classified tokens on partial days"}],
  "facts_verify": {"exit_code": 2},
  "handoff": {"id": "hoff_cd63055e31a4"},
  "waivers": [{"verb": "facts", "reason": "Pre-existing grounding drift reporting missing files from earlier relocated projects."}]
}
```
