# Receipt remediation RCP-2026

## Context and files

P1. Ready, no prerequisites. Cwd `/mnt/Jad/github/lab/thermal-streak`.
Why: a command that never ran must not certify work. Prior completion remains historical.
Read `internal/thermal/receipt.go`, the audit's `probes.txt`, and `tasks/plan.md` common contract.
Edit `internal/thermal/receipt.go` and `internal/thermal/receipt_test.go`; permitted companion docs
are defined in `tasks/plan.md`. No ingestion adapter rewrite or spend accounting changes.

## Fixed behavior

Keep `ParseCommandEvidence(cmd string, exitCode int, output string) (EvidenceKind, string, bool)`.
Use token-boundary command recognition, not arbitrary substring presence. After existing shell
wrapper unwrapping, only a single recognized invocation can inherit a process exit status.
Commands containing shell lists, pipelines, substitutions or control operators remain unknown
unless the source adapter supplies separate command/result pairs. Do not implement a shell parser.
Explicit nonzero exits win. Missing exits remain unknown unless a complete recognized result line
establishes success/failure. Match success lines before generic failure words; `0 failed` alone
must not mean failure or prove an otherwise unfinished run. For pytest, a completed summary with
positive passed count and zero failed count is success; nonzero failed count is failure.
A standalone `The command exited with code 0` is success for a recognized command.
Retain sanitized labels, invocation correlation, deduplication and failure precedence.

## Example and acceptance

- RCP-201: `false && go test ./... || true`, exit 0, empty output -> unknown, no verified test.
  Include pipe-to-true, quoted mentions, `go tester`, empty and result-only records.
- RCP-202: `go test ./...`, unknown exit, completed zero-exit line -> test pass;
  `pytest`, unknown exit, `10 passed, 0 failed` -> pass; `1 passed, 2 failed` -> fail.
  Explicit exit 2 plus apparent pass text -> fail. Missing result -> unknown.
- RCP-203: add `TestReceiptCompoundCommandNotVerified` and
  `TestReceiptCompletedSummaryClassification`, including one synthetic Claude invocation/result
  fixture through `parseClaudeReceipt`, not only the helper. `go test -race ./internal/thermal
  -run 'Receipt|Evidence|Outcome'` exits 0; rerun audit receipt probes and common final checks.

Stop/return/review contract: `tasks/plan.md`; retain acceptance IDs above and historical IDs.
