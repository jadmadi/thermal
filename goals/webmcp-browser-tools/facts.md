# Facts: Support WebMCP to expose site tools to AI agents via the browser

## Issue
No WebMCP tools detected on page load.

## Fix & Requirements
Implement the WebMCP API by calling `navigator.modelContext.provideContext()` in `docs/pages/index.html` with tool definitions that expose key client actions/queries to AI agents:
- `getStreakStatus`: Returns current streak and activity stats.
- `listSupportedTools`: Returns list of 14 supported AI agents.
- `calculateEstimate`: Calculates estimated cost for given token volume and model.
Each tool defines `name`, `description`, `inputSchema` (JSON Schema), and an `execute` callback function.

## Architectural Context & References
- Skill: `https://isitagentready.com/.well-known/agent-skills/webmcp/SKILL.md`
- WebMCP Spec: `https://webmachinelearning.github.io/webmcp/`
- Chrome Developer Blog: `https://developer.chrome.com/blog/webmcp-epp`
