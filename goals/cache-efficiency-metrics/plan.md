# Plan: Cache Efficiency & Disjoint Token Breakdown

## Tasks
- [ ] **Phase 1 — Token Composition Aggregator**: Add helper in `internal/thermal/stats.go` to aggregate global disjoint token counts (`Input`, `Output`, `Reasoning`, `CacheRead`, `CacheWrite`) and compute cache hit rate.
- [ ] **Phase 2 — `thermal stats` UI Enhancement**: Render a "Token composition" block in `internal/render/analytics.go` with Unicode block bars (`█` / `·`) and percentages for each bucket.
- [ ] **Phase 3 — Cache Efficiency Metrics**: Compute and display prompt cache hit rate (`XX.X%`) and estimated savings compared to uncached pricing.
- [ ] **Phase 4 — Tests & Parity Verification**: Add unit tests in `stats_test.go` and `analytics_test.go` covering 0% cache, 100% cache, mixed, and empty inputs.
- [ ] **Phase 5 — Documentation Updates**: Document the token composition section in `README.md` and HTML site.
