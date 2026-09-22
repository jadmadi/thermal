# Facts: Publish Auth.md metadata for agent registration

## Issue
`auth.md` not found.

## Fix & Requirements
Serve `/auth.md` at the site root (`docs/pages/auth.md`) with agent registration instructions, publish `/.well-known/oauth-protected-resource`, and include an `agent_auth` block in `/.well-known/oauth-authorization-server` with `register_uri`, supported identity types, credential types, and claim/revocation URLs where applicable.

## Architectural Context & References
- Skill: `https://isitagentready.com/.well-known/agent-skills/auth-md/SKILL.md`
- WorkOS Auth.md: `https://workos.com/auth-md`
- Spec Repository: `https://github.com/workos/auth.md`
