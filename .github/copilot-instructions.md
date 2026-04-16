# GitHub Copilot Instructions

You are assisting with a **BusinessRepo** — a domain-centric repository that owns all code, infra, CI/CD, tests, and docs for a single business capability end-to-end.

## Communication style

- Short sentences. Structured formatting (bullets, tables, sections).
- No marketing language, hype, filler, or motivational tone.
- State assumptions and trade-offs explicitly.
- Senior technical + executive audience. Skip fundamentals unless asked.

## Architecture standards

- **Clean Architecture**: domain logic must not depend on frameworks or infrastructure.
- **SOLID**: every class/module/function follows SRP, OCP, LSP, ISP, DIP.
- **Composition over inheritance**.
- **Testability, observability, reliability, security** by default.

## BusinessRepo layout (enforced)

```
/app        Application code
/common     Domain-specific shared libraries (within this repo only)
/infra      Terraform, Kubernetes, cloud configs
/tests      All test layers: unit, integration, e2e, contract
/docs       ADRs, stories, specs, design artifacts, DoD
Makefile    Standardized targets: build, test, test-integration, lint, security-scan
```

## Capability hierarchy (binding)

1. **Agent** — orchestration, reasoning, cross-step coordination
2. **Skill** — deterministic, reusable mechanics (CLI wrapping, templating, validation)
3. **MCP** — last resort only; requires ADR (see `.github/pull_request_template.md`)

## What Copilot completions MUST NOT do

- Suggest hardcoded tokens, credentials, or secrets.
- Suggest imports that cross domain/layer boundaries.
- Suggest `// Co-authored-by: <any AI attribution>` in commit messages.
- Suggest skipping error handling.
- Suggest `any` / untyped returns without a comment explaining why.

## Story + spec context

Story files live at `docs/stories/<epic>-<story>-<timestamp>-NNNN.md`.
Specs live at `docs/specs/<epic>-<slug>/` (PROBLEM → SPEC → PLAN → TASKS).
When completing code for a story, respect the Gherkin scenarios in the story file as the source of truth for behavior.

## Test expectations

- Unit tests for all pure functions and domain logic.
- Integration tests for infrastructure adapters.
- Gherkin/Cucumber scenarios for user-facing behavior (extracted into `tests/features/` at build time).
- A11y audit required for any component with `ui: true` in story frontmatter.

## Design tokens

`docs/design/identity/tokens.json` is the canonical source (W3C DTCG format).
CSS vars, Tailwind config, Mantine theme are generated from it — never manually edited.
Token naming: `{category}.{group}.{role}` e.g. `color.brand.primary`.
