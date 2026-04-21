# CLAUDE.md — Project Rules

## Agent System

Default agent: `@orchestrator`. It never writes code — delegates everything to specialist agents via the Task tool.

Commands:
- `/feature {description}` — Full 8-phase pipeline
- `/build-feature {description}` — Alias for `/feature`
- `/bugfix {description}` — Reproduce → fix → validate → CHANGELOG
- `/validate` — Run full test suite
- `/tdd {requirement}` — Single RED → GREEN → REFACTOR cycle
- `/superpower-runner {name}` — Launch a named workflow (e.g. `new-ui-feature`, `api-endpoint`, `bugfix`, `design-refresh`)
- `/ship-safe` — Run `npx ship-safe audit .` security scan before shipping (**optional** — disabled by default; enable by setting repo variable `ENABLE_SHIP_SAFE=true`)

Pipeline rules:
1. Orchestrator never writes code. Delegates to specialist agents.
2. **A Gherkin story file must exist in `docs/stories/` before any implementation begins.** The story IS the spec.
3. Tests are written before implementation (TDD enforced).
4. QA writes failing tests. Implementation agents make them pass.
5. 3 consecutive failures on any task → escalate to human.
6. `make test-all` must pass before Phase 8 gate.
7. Every feature gets a GitHub issue. Agents update it via `gh` CLI.
8. Conventional Commits: `feat:`, `fix:`, `test:`, `docs:`, `infra:`, `refactor:`.

**Gate enforcement:**
- No implementation without a Gherkin story file (orchestrator HALTS and REFUSES if absent)
- `ui: true` stories require approved wireframe + mockup before IMPLEMENT phase
- A11y audit required after IMPLEMENT for all `ui: true` stories

---

## Communication Style

- Short sentences. Structured formatting (bullets, tables, sections).
- No marketing language, hype, filler, or motivational tone.
- Explicit about assumptions, constraints, and trade-offs.
- Audience: senior engineers, staff+, VP/Director level.
- Do not explain fundamentals unless asked.

---

## BusinessRepo Model (Always On)

This repository is a **BusinessRepo** — a domain-centric repo that owns everything required to build, test, deploy, operate, and evolve one business capability end-to-end.

**Ownership scope per repo:**
- Application code
- Domain-specific shared libraries
- Infrastructure (Terraform, Kubernetes, cloud configs)
- CI/CD definitions
- All test layers (unit, integration, e2e, contract)
- Documentation

**Naming:** `<domain>`, `<domain>-service`, `<domain>-api`, `<domain>-app`.
Examples: `payments`, `payments-api`, `identity-service`, `export-app`.

**Canonical layout:**
```
/app        # application code (modularized internally)
/common     # domain-scoped shared libraries (no cross-domain imports)
/infra      # Terraform, K8s manifests, cloud configs
/tests      # all test layers
/docs       # specs, ADRs, runbooks, stories, design artifacts
Makefile    # all CI/CD calls Makefile targets
```

**Anti-patterns — reject any design that:**
- Creates horizontal shared repos ("shared-utils", "platform-common")
- Hides ownership behind tool-driven names ("terraform-repo", "k8s-manifests")
- Introduces cross-domain data coupling
- Requires coordinated deploys across unrelated domains

---

## Architecture Standards

- **Clean Architecture** — domain logic has zero framework/infrastructure imports
- **SOLID** — every module. Call out violations explicitly.
  - Single Responsibility: one reason to change
  - Open/Closed: extend without modifying
  - Liskov Substitution: subtypes honor contracts
  - Interface Segregation: no fat interfaces
  - Dependency Inversion: depend on abstractions
- Composition over inheritance
- Testability, observability, reliability, security by default

---

## Code Review Standard (Staff+)

- Call out boundary violations and architectural risk.
- Prefer long-term maintainability over short-term convenience.
- No politeness padding.
- Distinguish tactical vs strategic trade-offs.
- Explain reasoning and second-order effects.

---

## ADR / RFC Format

1. Title
2. Context
3. Goals / Non-goals
4. Proposal
5. Alternatives
6. Trade-offs and Risks
7. Impact (FinOps, SRE, Security, Team)
8. Decision
9. Next Steps

---

## Cost and Ops (Every ADR)

**FinOps:** Cost drivers, scaling characteristics, cost visibility.
**SRE:** Failure modes, blast radius, observability, recovery.

---

## MCP Servers

- `context7`: Library documentation lookup (`use context7` in prompts)
- New MCP servers require an ADR. Use `docs/architecture/adr-template.md` and add a row to `docs/architecture/README.md`.

---

## Skills

Read skills from `.claude/skills/` before executing tasks.
Key skills: `tdd-workflow`, `playwright-cli`, `github-cli`, `mermaid-diagrams`, `superpower-runner`, `ship-safe`.
