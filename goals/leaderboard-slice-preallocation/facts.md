# Facts: Preallocated Slice Capacity in Leaderboard Rendering

## Architectural Facts & Grounding

1. **Slice Capacity Growth (`golang-performance`, `golang-data-structures`)**:
   Appending to a slice created with zero capacity (`make([]T, 0)`) causes Go's runtime to double slice capacity and reallocate the underlying array repeatedly as elements are appended. When the bound or count is known (`len(results)`), preallocating with `make([]T, 0, len(results))` eliminates all intermediate allocations.

2. **Leaderboard Filtering Invariant**:
   `results` contains at most the registered tools (around 15 today). Both `tokenResults` and `activityResults` are subsets whose combined lengths equal `len(results)`. Sizing both to `len(results)` capacity is negligible in memory while completely eliminating resizing reallocations.
