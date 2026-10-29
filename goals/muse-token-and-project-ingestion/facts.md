# Facts: Ingest Muse JSONL session tokens, model facets, and project attribution

## Observed Evidence
- Meta Muse writes real-time session transcripts to `~/.local/share/muse/sessions/YYYY/MM/DD/<session-id>/session.jsonl`, `subagent/<sub-id>/session.jsonl`, and `approval-review/<rev-id>.jsonl`, while `session-index.db` lags behind active sessions and only stores `prompt_count`.
- Each `model_completed` event in `sessions/**/*.jsonl` records `input_tokens` (inclusive of `cache_read_tokens` and `cache_write_tokens`) and `output_tokens` (inclusive of `reasoning_tokens`).
- Files inside `approval-review/*.jsonl` run with stream kind `approval.review.runtime` and a deterministic synthetic clock (`1780531500000000 + sequence - 1`, corresponding to `2026-06-04`), and emit their model name on preceding `status` events at `event.details.facets[].detail.model`.
- Muse logs do not record billed USD cost per turn; `~/.local/share/muse/model-catalog/*.json` is a model rate card that must be applied by `internal/pricing` as `EstimatedCost` (`~$`), never stored as `DailyRow.Cost` (`StoredCost`).

## Architectural Invariants & Constraints
- Token fields (`Input`, `Output`, `Reasoning`, `CacheRead`, `CacheWrite`) must be disjoint and sum to `Tokens`.
- `DailyRow.Cost` (`StoredCost`) holds only cost recorded by the source on a turn or session, never a rate-card estimate (`internal/thermal/types.go:124`).

## File & Interface Contracts
- `internal/loaders/muse.go`: `LoadMuseData(path string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error)`
- `internal/loaders/sourcesig.go`: `GetToolSourceSig` case `thermal.ToolMuse`
- `internal/loaders/muse_test.go`: `TestLoadMuseData_SessionIndex`, `TestLoadMuseData_SessionJSONL`, `TestLoadMuseData_Hybrid`, `TestLoadMuseData_SourceSig`

## Verification Contract
```bash
CGO_ENABLED=0 go test -race ./internal/loaders -run 'TestLoadMuseData'
```
