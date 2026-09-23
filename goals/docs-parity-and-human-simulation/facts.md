# Facts: Comprehensive Documentation Parity, Full Scan & Simulated Human User Audit

## Architectural Invariants & Constraints
- 100% Feature Parity: Every CLI command (thermal stats, thermal serve, thermal share, thermal yield, thermal audit, thermal receipt), flag, and dashboard surface must be documented with accurate usage examples in both Markdown and HTML.
- Simulated Human Gate Pass: Update scripts/simulated_user_gate.sh to simulate real human workflows for all new features and require all checks to pass (100% green).
- Zero Regressions & Link Integrity: HTML docs must pass validation with zero broken internal links, correct asset paths, valid JSON-LD schema, and synchronized version metadata.

## File & Interface Contracts
- Relevant files:
  - `README.md`
  - `AGENTS.md`
  - `docs/pages/index.html`
  - `docs/pages/404.html`
  - `docs/pages/index.md`
  - `scripts/simulated_user_gate.sh`
