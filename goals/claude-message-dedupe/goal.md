# Goal: Claude Code Message-Level Deduplication

## Goal Description
Implement message-level deduplication in the Claude Code loader ([`internal/loaders/claude.go`](file:///mnt/Jad/github/lab/thermal-streak/internal/loaders/claude.go)) to eliminate duplicate streaming chunks, tool retries, and repeated turns from session transcripts.

Claude Code writes session transcripts as JSONL lines in `~/.claude/projects/*/*.jsonl`. In sessions with streaming turn records, tool execution retries, or subagent message propagation, assistant message events can be logged with identical message IDs or duplicate usage structures.

Thermal currently iterates through each JSONL line matching `rec.Type == "assistant"` and adds its tokens directly. If identical messages are logged across the transcript, tokens are overcounted.

Tyler Southwick's analysis of local AI coding histories in StackReplay specifically highlighted the necessity of a deduplication pass against raw logs, discovering 1,267 exact duplicate events (~1.3%) across local session files.

This goal adds:
1. Message-level deduplication tracking in `claude.go` using `rec.Message.ID` (and fallback to content/timestamp signature if ID is missing).
2. Diagnostic counting of duplicate events removed under `--verbose` (`summary.Warnings`).
3. Unit test coverage with synthetic JSONL transcripts containing duplicated message records.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Shape**: ship

## References
- **Shared Understanding & Fact Sheet**: [`goals/claude-message-dedupe/facts.md`](facts.md)
- **Execution Plan**: [`goals/claude-message-dedupe/plan.md`](plan.md)

## Done Condition
1. `claude.go` deduplicates assistant messages within each session transcript file based on message ID.
2. Duplicate counts are tracked and reported in `summary.Warnings` under `--verbose`.
3. Unit tests in `claude_test.go` assert that duplicate message blocks are skipped and counted once.
4. `go test -v -race ./...` passes cleanly with zero race conditions.
5. `golangci-lint` passes with zero issues.
