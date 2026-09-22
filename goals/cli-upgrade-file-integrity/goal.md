# Goal: Archive Extraction Deduplication and File Write Close Verification in CLI Self-Updater

## Goal Description
In `cmd/thermal/upgrade.go`, binary extraction from `.tar.gz` archives had duplicate logic paths: an initial check for root-level `thermal` binaries followed immediately by an identical branch matching any `thermal` binary, with both blocks executing the exact same extraction code. Furthermore, file creation functions (`downloadFile` and `extractBinary`) deferred `out.Close()` without checking write errors on close, which can silently discard file truncation or disk-full errors during filesystem flush.

This goal consolidates the archive extractor into a single, clean branch and explicitly captures and returns `out.Close()` errors upon file write completion.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship

## Files to Touch
- `cmd/thermal/upgrade.go`
- `cmd/thermal/upgrade_test.go`

## References
- **Shared Understanding & Fact Sheet**: [`goals/cli-upgrade-file-integrity/facts.md`](facts.md)
- **Execution Plan**: [`goals/cli-upgrade-file-integrity/plan.md`](plan.md)

## Done Condition
1. Archive binary extraction eliminates dead or redundant condition branches in `extractBinary`.
2. All file write routines in `upgrade.go` check and propagate `out.Close()` return values.
3. Unit tests pass cleanly: `go test -v ./cmd/thermal/...`.
4. Simulated user gate passes: `./scripts/simulated_user_gate.sh`.
