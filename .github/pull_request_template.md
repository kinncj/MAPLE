## What is this PR doing?

<!-- One sentence. Be direct. -->

## Type of change

- [ ] Feature
- [ ] Bug fix
- [ ] Refactor
- [ ] Docs
- [ ] Spike

## Story / Issue

Closes #<!-- issue number -->

## Checklist

- [ ] Tests added or updated
- [ ] All existing tests pass (`make test`)
- [ ] SDLC gates green (frontmatter, spec-kit, design-approved, a11y)
- [ ] DoD checklist complete in linked Issue
- [ ] ADRs linked if required (`adr_required: true` in story frontmatter)
- [ ] No secrets committed

---

## MCP Adoption Checklist (complete ONLY if this PR introduces an MCP server)

> Leave this section blank and unchecked if no MCP is involved.
> PRs introducing MCP without completing this checklist are blocked per §4.1 of the platform PRD.

- [ ] **Capability:** Describe what this MCP provides.
- [ ] **Skill gap:** Explain specifically why a skill cannot satisfy this need.
- [ ] **CLI coverage:** Confirm no `gh`, language tooling, or local CLI covers this.
- [ ] **ADR linked:** Link the ADR documenting the decision (required — no ADR, no merge).
- [ ] **Owner named:** Identify the owner for uptime, auth rotation, and version upgrades.
- [ ] **Degraded mode:** Describe offline/degraded-mode behavior.
- [ ] **Auth secrets:** Confirm secrets are stored per repo security standards (no hardcoded tokens).
