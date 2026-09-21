# Plan: Claude Code Message-Level Deduplication

## Tasks
- [ ] **Phase 1 — Schema & Message ID Extraction**: Update record unmarshaling in `internal/loaders/claude.go` to capture `rec.Message.ID`.
- [ ] **Phase 2 — Deduplication Logic**: Implement per-file `seenMsgIDs` set to filter out duplicate messages and track removed count.
- [ ] **Phase 3 — Warning & Diagnostic Reporting**: Report deduplication counts via `Summary.Warnings` when `--verbose` is enabled.
- [ ] **Phase 4 — Unit Tests**: Add test cases in `claude_test.go` with synthetic transcripts containing duplicate message IDs and verify exact counts.
- [ ] **Phase 5 — Full Test Suite Pass**: Verify with `go test -v -race ./...` and `golangci-lint`.
