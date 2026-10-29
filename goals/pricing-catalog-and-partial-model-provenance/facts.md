# Facts: Price classified tokens on partial days, resolve free-tier suffixes, and refresh model catalog

## Observed Evidence
- In `internal/pricing/pricing.go:156`, `PriceDay` checks `if !ok || t.Unclassified > 0` and skips the entire model for that day. Meanwhile `AggregateModels` (`internal/thermal/models.go:98-104`) constructs `ModelTokens` without `Unclassified`, causing `thermal models` to price `deepseek-v4-flash-0731` at `~$131.07` while `thermal monthly` prints `No pricing for: deepseek-v4-flash-0731`.
- In `internal/pricing/pricing.go:198`, `lookupCandidates` strips `:free` (colon) but does not handle `-free` (hyphen), causing 2.1B tokens of `muse-spark-1.3-contributor-free`, `muse-spark-1.2-contributor-free`, `nemotron-3-ultra-free`, and `x-preview-f-free` in OpenCode to show as missing pricing (`—`).
- `internal/pricing/subscription.go` and `internal/pricing/subscription_test.go` have zero callers outside their own test file since `thermal replay` was retired.

## Architectural Invariants & Constraints
- A model whose tokens on a day are 100% `Unclassified` (e.g. `codewhale-session` with `Unclassified: 1_000_000` and zero `Input`/`Output`) must still return `$0` and appear in `MissingPricing` (`TestPriceDaySkipsUnclassifiedAndEmpty`).
- User overrides in `~/.config/thermal/pricing.json` must always win over `models.dev`, `applyMuseCatalog`, and `-free` defaults.

## File & Interface Contracts
- `internal/pricing/pricing.go`: `Catalog.Lookup(model string) (Price, bool)`, `Catalog.PriceDay(day thermal.DailyRow) (float64, []string)`, `applyMuseCatalog(cat *Catalog, dir string)`
- `internal/pricing/models.go`: `CanonicalIdentities`
- `internal/pricing/modelsdev.go`: `variantAliases`

## Verification Contract
```bash
CGO_ENABLED=0 go test -race ./internal/pricing/...
```
