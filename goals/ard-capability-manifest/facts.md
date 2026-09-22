# Facts: Publish an ARD (Agentic Resource Discovery) manifest so agents can discover your site's capabilities

## Issue
ARD capability manifest not found.

## Fix & Requirements
Serve `/.well-known/ai-catalog.json` (`docs/pages/.well-known/ai-catalog.json`) at the origin root with `Content-Type: application/json` and `Access-Control-Allow-Origin: *`. Include `specVersion`, a `host` object, and an `entries` array. Give each entry a `urn:air:thermal.jadmadi.net:<namespace>:<name>` identifier, a `displayName`, an IANA media type in `"type"`, and exactly one of `url` or `data`. Add 2-5 `representativeQueries` per entry so registries can build semantic embeddings.

## Architectural Context & References
- Skill: `https://isitagentready.com/.well-known/agent-skills/ard/SKILL.md`
- ARD Spec: `https://agenticresourcediscovery.org/`
- GitHub Spec: `https://github.com/ards-project/ard-spec`
- Schema: `https://github.com/Agent-Card/ai-catalog`
