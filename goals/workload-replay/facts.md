# Facts & Invariants: Workload Replay Simulation (`thermal replay`)

## 1. Context & Motivation
Developers using AI coding assistants (Claude Code, Devin, Codex, OpenCode, Cursor, Grok, etc.) spend between $20/mo to $500+/mo across various subscriptions and API keys. Because each tool logs usage locally in different schemas and units, developers have no objective way to answer:
1. **Capacity Feasibility**: *"Can subscription X (e.g. Claude Pro $20/mo) actually handle my daily workload, or will it throttle me?"*
2. **Economic Simulation**: *"What would my actual historical coding workload have cost if executed through alternative models (e.g. Claude 3.7 Sonnet vs DeepSeek V3 vs GPT-4o) or API providers?"*
3. **Plan Recommendations**: *"Am I paying for subscriptions whose limits I constantly exceed, forcing me onto secondary fallback tools?"*

Thermal already scans 14 tools locally and parses disjoint token metrics (`UncachedInput`, `Output`, `Reasoning`, `CacheRead`, `CacheWrite`) and daily timestamps. `thermal replay` leverages this ground-truth data to simulate historical workloads against target rate cards and subscription ceilings.

---

## 2. Invariants & Calculation Rules

### A. Disjoint Token Pricing
Any API rate card simulation MUST use Thermal's disjoint canonical token types:
$$\text{SimulatedCost} = I_{\text{uncached}} \times R_{\text{in}} + O \times R_{\text{out}} + R_{\text{reason}} \times R_{\text{reason}} + C_{\text{read}} \times R_{\text{cache\_read}} + C_{\text{write}} \times R_{\text{cache\_write}}$$
- Because 95%+ of agent tokens are cache reads, ignoring cache discounts yields 10x–20x cost overestimations.
- If a target model does not support prompt caching, all prompt tokens ($I_{\text{uncached}} + C_{\text{read}} + C_{\text{write}}$) are priced at the base input rate.

### B. Subscription Ceilings & Throttle Estimation
- Subscription plans impose rate limits rather than raw token bills.
- Limits are modeled as daily effective throughput thresholds (e.g., Claude Pro ~15M tokens/day; Claude Max ~75M tokens/day; ChatGPT Plus ~20M tokens/day).
- On any historical day where `DailyRow.Tokens > TierCapacity`, that day is marked as **Throttled**.
- **Throttle Rate** is calculated as:
  $$\text{ThrottleRate} = \frac{\text{Days Where } \text{Tokens} > \text{Ceiling}}{\text{Total Active Days}} \times 100\%$$
- A plan with $\text{ThrottleRate} > 20\%$ receives a `FAIL (Throttled)` verdict.

### C. Local-First & Zero Telemetry
- Replay simulations run 100% locally.
- Price catalogs are fetched via existing `internal/pricing` caching mechanisms (`models.dev`).
- Custom price cards or subscription definitions can be passed via CLI or local configuration.
