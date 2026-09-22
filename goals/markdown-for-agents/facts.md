# Facts: Return HTML responses as markdown when agents request it

## Issue
Site does not support Markdown for Agents.

## Fix & Requirements
Enable Markdown for Agents so requests with `Accept: text/markdown` return a markdown version of your HTML response while HTML stays the default for browsers. Confirm the response uses `Content-Type: text/markdown` (and `x-markdown-tokens` if available).

## Architectural Context & References
- Skill: `https://isitagentready.com/.well-known/agent-skills/markdown-negotiation/SKILL.md`
- Cloudflare Markdown for Agents Docs: `https://developers.cloudflare.com/fundamentals/reference/markdown-for-agents/`
- Implementation: Provide static `docs/pages/index.md` alongside `docs/pages/index.html` and configure Cloudflare Worker / Page Rule or headers.
