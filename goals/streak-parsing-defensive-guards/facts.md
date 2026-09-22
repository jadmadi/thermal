# Facts: Defensive Timestamp Parsing and Malformed Date Guards in Streak Computation

## Architectural Facts & Grounding

1. **Defensive Date Parsing (`golang-safety`)**:
   Under `golang-safety`, ignoring error returns with `_ = err` is an anti-pattern. When parsing temporal strings (`time.Parse`), a failure yields `time.Time{}` (the epoch zero value `0001-01-01T00:00:00Z`).

2. **Streak Computation Invariants**:
   - `ComputeStreaks` takes `map[string]bool` where keys are expected to be `YYYY-MM-DD`.
   - If a corrupted string enters the set, comparing against `prev.AddDate(0, 0, 1).Equal(t)` with a zero time causes streak calculations to break.
   - Inspecting `err != nil` and `continue` ensures that malformed tokens are dropped and do not disrupt valid runs.
