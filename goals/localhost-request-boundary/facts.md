# Facts: Reject untrusted hosts at the local telemetry API

## Observed Evidence
`internal/server/server.go:84-110` installs security headers but has no request-authority check. The synthetic handler request to http://untrusted.example/api/telemetry returned HTTP 200 with fixture telemetry. This establishes missing Host validation, not a demonstrated browser exploit. Go's Request.Host documentation recommends validating authority to prevent DNS rebinding: https://pkg.go.dev/net/http#Request.

Evaluated working tree: HEAD `88c24d5` plus pre-existing uncommitted changes, 2026-09-24. Re-read these locations before implementing; line numbers describe this snapshot. Existing goal completion is commit attribution, not evidence that these edge cases pass.

## Ownership and Constraints
- `internal/server/server.go`
- `internal/server/server_test.go`
- `docs/MIGRATION.md`

No account system, external identity provider or remote telemetry service. Coordinate with web-telemetry-parity to avoid concurrent edits to server.go.

Preserve unrelated work. Source ingestion stays read-only; normalized token categories remain disjoint; no prompts in output/cache; diagnostics stay off JSON stdout. Keep aggregation in internal/thermal, bounded worker pools, and no background service. Load the repository Go/UI skills when the implementation scope needs them.

## Verification Contract
`go test -race ./internal/server`

Then `go test -race ./...`, `go build -o /tmp/thermal-test ./cmd/thermal`, and `./scripts/simulated_user_gate.sh`. Use an isolated synthetic home, tool-root overrides and cached pricing. New benchmark paths are deliverables, not commands already present at evaluation time.
