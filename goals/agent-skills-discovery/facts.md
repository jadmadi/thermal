# Facts: Publish an agent skills discovery index

## Issue
Agent Skills index not found.

## Fix & Requirements
Publish a skills discovery index at `/.well-known/agent-skills/index.json` (`docs/pages/.well-known/agent-skills/index.json`) per the Agent Skills Discovery RFC v0.2.0. Include a `$schema` field, and a `skills` array where each entry has `name`, `type`, `description`, `url`, and a `sha256` digest.

## Architectural Context & References
- Skill: `https://isitagentready.com/.well-known/agent-skills/agent-skills/SKILL.md`
- Cloudflare RFC: `https://github.com/cloudflare/agent-skills-discovery-rfc`
- Registry: `https://agentskills.io/`
