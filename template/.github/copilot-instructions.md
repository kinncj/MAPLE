# Copilot CLI Instructions — AI Squad

## Agent System

Default agent: `@orchestrator`. It never writes code — delegates everything to specialist agents.

## Hard Rules (Non-Negotiable)

1. **TDD is mandatory.** `@qa` writes a failing test FIRST. Implementation follows. Never write implementation before a failing test exists.
2. **Orchestrator never writes code.** It decomposes, delegates, and gates.
3. **3 failures → escalate.** After 3 consecutive failures on any task, stop and surface to the human.
4. **No secrets in code.** Never write API keys, passwords, or tokens directly in source files.
5. **Conventional Commits.** All commits use `feat:`, `fix:`, `test:`, `docs:`, `infra:`, `refactor:`.
6. **`make test-all` must pass** before any PR is created (Phase 8 gate).

## Before Writing Any Code

Read `project.config.yaml`:
- Check `stack:` — if any key is null, run tech-stack discovery (ask the user).
- Check `sdlc.mode` — if `standard`, Spec-Kit intake applies before DISCOVER.

## After Writing Code

Run these in order:
```bash
make lint
make test
```

If either fails, **fix the failure before proceeding**. Do not move to the next task with a red build.

## Before `git commit` or `git push`

Run SDLC gates:
```bash
bash scripts/sdlc/spec-kit-gate.sh
bash scripts/sdlc/validate-frontmatter.sh $(find docs/stories -name "*.md" ! -name "_template.md" 2>/dev/null)
bash scripts/sdlc/design-approved-gate.sh $(find docs/stories -name "*.md" ! -name "_template.md" 2>/dev/null)
bash scripts/sdlc/a11y-gate.sh $(find docs/stories -name "*.md" ! -name "_template.md" 2>/dev/null)
```

If any gate fails, **do not commit**. Report the failure and wait for the human to resolve.

## Story Files (`docs/stories/**/*.md`)

Every story file must have valid YAML frontmatter:
- `id`, `title`, `epic`, `priority`, `ui`, `adr_required`, `milestone`, `labels`
- `ui: true` for ANY story involving a rendered UI element (pages, cards, modals, forms, navigation)
- `priority` must be `critical | high | medium | low`

After writing or editing a story file, validate:
```bash
bash scripts/sdlc/validate-frontmatter.sh <file>
```

## Phase Gates

| Phase | Gate condition |
|---|---|
| DISCOVER → ARCHITECT | Human approves stories |
| ARCHITECT → PLAN | Human approves architecture + ADR |
| PLAN → INFRA | plan.md complete, every impl task has a preceding test task |
| INFRA → IMPLEMENT | `docker compose up --wait` exits 0 |
| IMPLEMENT loop | RED (test fails) → GREEN (test passes) per task |
| IMPLEMENT → VALIDATE | All tests pass |
| VALIDATE → DOCUMENT | 100% test pass across all categories |
| DOCUMENT → FINAL GATE | Docs complete, CHANGELOG updated |
| FINAL GATE → PR | `make test-all` exits 0 |

## UI Stories (`ui: true`)

Insert design sub-pipeline before IMPLEMENT:
1. UX Research → personas + journey map
2. Wireframe → **PAUSE for human approval**
3. Visual Identity (if `docs/design/identity/tokens.json` missing) → **PAUSE for human approval**
4. Design Tokens → CSS vars, Tailwind config, Mantine theme
5. Mockup → **PAUSE for human approval**
6. Component Scaffold
7. After IMPLEMENT: A11y audit — no critical/serious WCAG 2.2 AA violations
