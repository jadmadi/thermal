# Plan: Support WebMCP to expose site tools to AI agents via the browser

## Execution Steps
- [x] Add client-side script to `docs/pages/index.html` guarding on `if ('modelContext' in navigator)`.
- [x] Register tools (`getStreakStatus`, `listSupportedTools`, `getToolDetails`) via `navigator.modelContext.provideContext()`.
- [x] Test in compatible browser environment (Chrome with experimental WebMCP enabled).
- [x] Commit with goal reference `(goals/webmcp-browser-tools/goal.md)`.
