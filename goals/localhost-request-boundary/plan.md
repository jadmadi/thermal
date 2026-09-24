# Plan: Reject untrusted hosts at the local telemetry API

## Execution Steps
- [ ] **Task 1 — Define permitted request authorities**: Validate Request.Host against the actual listening address/port and intentional local aliases before collecting telemetry. Parse IPv4/IPv6 and ports structurally. Preserve explicit --host behavior; wildcard bind must not mean every DNS name is trusted.

  Acceptance and verification: 127.0.0.1, localhost and configured IPv6 cases work where intended; foreign names, malformed authorities and invalid ports are rejected before collection.

- [ ] **Task 2 — Apply the boundary to every telemetry route**: Cover REST and SSE, avoid trusting forwarded host headers, and avoid enabling broad CORS. Specify a same-origin policy for browser requests while keeping direct local CLI clients functional.

  Acceptance and verification: Foreign Host GET /api/telemetry returns a bounded 4xx with no data; the same policy covers cached and uncached paths and /api/stream.

- [ ] **Task 3 — Verify and document request-contract changes**: Add fixture-only handler tests and one real loopback integration test; document affected reverse-proxy/custom-host usage using six-part migration anatomy.

  Acceptance and verification: Run server tests with -race; valid browser/SSE flow works and attacker-controlled authority does not trigger data loading.

- [ ] **Checkpoint — Review scope and release evidence**: Inspect only owned files; confirm acceptance criteria, source-data safety, required six-part migration entry, focused checks, full race tests, build and simulated gate. Record actual command outcomes, including any environment limits.
- [ ] **Attribution commit**: Stage only this goal's implementation and evidence; inspect `git diff --cached --name-only`; create an impact-first Conventional Commit ending in `(goals/localhost-request-boundary/goal.md)`, then run `sila goals`. Do not push without explicit authorization. Planning commits must not close this goal.

## Dependencies and Handoff
Start after `web-telemetry-parity`. Read prerequisite contracts before editing shared code.

## Verification
`go test -race ./internal/server`

No account system, external identity provider or remote telemetry service. Coordinate with web-telemetry-parity to avoid concurrent edits to server.go.
