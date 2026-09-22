# Plan: Return HTML responses as markdown when agents request it

## Execution Steps
- [x] Generate clean, structured markdown representation of the documentation site as `docs/pages/index.md`.
- [x] Configure Cloudflare content negotiation or add Cloudflare Worker script `docs/pages/_worker.js` to inspect `Accept: text/markdown` and serve markdown.
- [x] Ensure `Content-Type: text/markdown; charset=utf-8` and `x-markdown-tokens` headers are emitted.
- [x] Verify using `curl -H "Accept: text/markdown" https://thermal.jadmadi.net/`.
- [x] Commit with goal reference `(goals/markdown-for-agents/goal.md)`.
