# Plan: Include Link response headers for agent discovery (RFC 8288)

## Execution Steps
- [x] Add `_headers` in `docs/pages/_headers` declaring `Link` response headers for `/*`.
- [x] Add corresponding `<link rel="...">` HTML headers in `<head>` of `docs/pages/index.html` and `docs/pages/404.html`.
- [x] Verify link target paths exist and headers match RFC 8288 format.
- [x] Commit with goal reference `(goals/agent-link-headers/goal.md)`.
