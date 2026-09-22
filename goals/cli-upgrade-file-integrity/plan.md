# Plan: Archive Extraction Deduplication and File Write Close Verification in CLI Self-Updater

## Execution Steps
- [x] **Task 1 — Refactor `downloadFile`**: Capture `out.Close()` after `io.Copy(out, resp.Body)` so flush errors are returned.
- [x] **Task 2 — Refactor `extractBinary`**: Remove redundant if-branch and ensure `out.Close()` error is returned when writing the extracted binary.
- [x] **Task 3 — CLI Test Validation**: Run `go test -v ./cmd/thermal/...` to verify self-updater and flag parsing integrity.
- [x] **Task 4 — Simulated User Gate**: Run `./scripts/simulated_user_gate.sh` to confirm overall CLI release readiness.
