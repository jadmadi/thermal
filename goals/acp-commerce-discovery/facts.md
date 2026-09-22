# Facts: Publish ACP discovery metadata so agents can discover your commerce API

## Issue
ACP discovery document not found.

## Fix & Requirements
Serve `/.well-known/acp.json` (`docs/pages/.well-known/acp.json`) at the origin root with `protocol.name "acp"`, `protocol.version`, `api_base_url`, supported transports, and `capabilities.services` so agents can discover your ACP implementation without creating a checkout session first.

## Architectural Context & References
- Skill: `https://isitagentready.com/.well-known/agent-skills/acp/SKILL.md`
- ACP Portal: `https://agenticcommerce.dev`
- GitHub Discovery RFC: `https://github.com/agentic-commerce-protocol/agentic-commerce-protocol/blob/main/rfcs/rfc.discovery.md`
