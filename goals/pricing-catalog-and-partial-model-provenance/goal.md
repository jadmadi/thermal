# Goal: Price classified tokens on partial days, resolve free-tier suffixes, and refresh model catalog

## Chain of Custody (Isnad)
- **Registered By**: Pending Registration
- **Spec Audited By**: Pending Spec Audit
- **Implemented By**: Pending Implementation
- **Mystery Audited By**: @mystery-auditor | Report: `receipt_pricing-catalog-and-partial-model-provenance_1793263600`
## Goal Description
Harden `internal/pricing` so that (1) known catalog models with both classified and unclassified tokens on the same day price their classified tokens instead of being dropped into `MissingPricing`, (2) `-free` and `:free` model variants resolve cleanly without false missing-price warnings, (3) Meta Muse's local `model-catalog/*.json` rates are layered into `pricing.Load` before user overrides so Muse usage is priced as `EstimatedCost` (`~$`), and (4) dead `internal/pricing/subscription.go` files from the retired `replay` command are archived to `rm/`.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: muse-token-and-project-ingestion
- **Sequence**: 230
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required

## Files to Touch
- Edit: `internal/pricing/pricing.go`
- Edit: `internal/pricing/pricing_test.go`
- Edit: `internal/pricing/models.go`
- Edit: `internal/pricing/modelsdev.go`
- Edit: `internal/pricing/models_test.go`

## Requirements
1. **Partial Classified Token Pricing (`internal/pricing/pricing.go`)**:
   - In `PriceDay(day thermal.DailyRow)`, compute `classified := t.Input + t.Output + t.Reasoning + t.CacheRead + t.CacheWrite`.
   - When `p, ok := c.Lookup(model)` succeeds and `classified > 0`, price the classified tokens (`Input`, `Output+Reasoning`, `CacheRead`, `CacheWrite`) even if `t.Unclassified > 0`, rather than rejecting the entire model into `MissingPricing`. Only append `model` to `MissingPricing` when `!ok` or when `classified == 0 && t.Unclassified > 0`.
2. **Free-Tier Suffix Resolution (`internal/pricing/pricing.go`)**:
   - In `Catalog.Lookup(model string)`, after checking exact/alias/candidate keys, recognize models ending in `:free` or `-free` as zero-cost (`Price{}`, `true`) when the base model or free suffix indicates an explicitly free tier, so free-tier models (`muse-spark-1.3-contributor-free`, `muse-spark-1.2-contributor-free`, `nemotron-3-ultra-free`, `x-preview-f-free`) do not trigger false `No pricing for:` warnings.
3. **Muse Local Catalog Layering & Model Identities (`internal/pricing/pricing.go`, `internal/pricing/models.go`, `internal/pricing/modelsdev.go`)**:
   - Layer `~/.local/share/muse/model-catalog/*.json` (and default Muse catalog prices) in `pricing.Load` via `applyMuseCatalog` before `applyOverrides`.
   - Add current 2026 model identities and aliases to `CanonicalIdentities` and `variantAliases`.
   - Archive dead `internal/pricing/subscription.go` and `internal/pricing/subscription_test.go` to `rm/`.

## Acceptance Criteria
- AC-001: `CGO_ENABLED=0 go test -race ./internal/pricing/...` exits 0.
- AC-002: A model in the catalog with `Input: 1_000_000, Unclassified: 500` on a day has its 1M input tokens priced and is not listed in `MissingPricing`, while a model with only `Unclassified: 1_000_000` stays unpriced and listed in `MissingPricing`.
- AC-003: Models with `-free` or `:free` suffixes resolve to `Price{}` (`$0.00`) without appearing in `MissingPricing`.

## References
- **Shared Understanding & Fact Sheet**: [`goals/pricing-catalog-and-partial-model-provenance/facts.md`](facts.md)
- **Execution Plan**: [`goals/pricing-catalog-and-partial-model-provenance/plan.md`](plan.md)

## Memory Plan
- **Prime**: run `sila prime` (CLI) or `prime_context` (MCP) at task start and record the briefing timestamp.
- **Recall**: run `sila recall "<symptom or design question>"` before implementation; record each query and match count.
- **Writes**: record at least one `memo`, `lesson`, or `case` grounded to a file in this goal; use a `case` when the goal introduces a reusable mechanism.
- **Facts**: run `sila facts verify` before certification and record the exit code and violation delta.
- **Handoff**: run `sila handoff "<summary>" --open="..."` at session end and record the handoff ID.
- **Evidence**: embed a `sila/memory-writes/v1` block under `## Memory Writes`, per `memory-protocol-goal-contract`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`); unchecked tasks block the `implemented` transition.
2. Implementation is committed with `(goals/pricing-catalog-and-partial-model-provenance/goal.md)` or `[goal: pricing-catalog-and-partial-model-provenance]` attribution under a `feat:`, `fix:`, `refactor:`, `perf:`, or `chore:` prefix.
3. Goal is certified (`goals/pricing-catalog-and-partial-model-provenance/certification.json` with `verified: true` or a signed `audit-receipt.json`) and any markdown evidence contains an explicit standalone `Verdict: PASS` declaration.
4. `sila goals --json` reports `status: "implemented"`, `verification_status: "verified"`, and `unverified_count: 0` for delivered goals, with a clean `git status`.
