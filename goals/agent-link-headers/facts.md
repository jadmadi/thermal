# Facts: Include Link response headers for agent discovery (RFC 8288)

## Issue
No Link headers found on target page (`https://thermal.jadmadi.net/`).

## Fix & Requirements
Add Link response headers to the homepage (and `<link rel="...">` in HTML `<head>` for static GitHub Pages) pointing agents to useful discovery resources:
- `Link: </.well-known/api-catalog>; rel="api-catalog"`
- `Link: </.well-known/agent-skills/index.json>; rel="agent-skills"`
- `Link: </.well-known/mcp/server-card.json>; rel="mcp-server-card"`
- `Link: </docs/pages/index.html>; rel="service-doc"`

## Architectural Context & References
- Skill: `https://isitagentready.com/.well-known/agent-skills/link-headers/SKILL.md`
- RFC 8288 (Web Linking): `https://www.rfc-editor.org/rfc/rfc8288`
- RFC 9727 (API Catalog Link Relations): `https://www.rfc-editor.org/rfc/rfc9727#section-3`
- Host environment: GitHub Pages / Cloudflare Pages custom headers file (`docs/pages/_headers`).
