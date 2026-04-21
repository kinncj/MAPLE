# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Conventional Commits](https://www.conventionalcommits.org/).

<!-- Agents append entries here using: gh issue comment + docs agent -->

## [4.4.2] — 2026-04-21

### Fixed
- `writeRecoveryMarker` now merges with existing `maple.json` instead of overwriting — TUI start/exit no longer wipes superpower pipeline state written by the `superpower-runner` skill, so the `P` pane shows the correct status

## [4.4.1] — 2026-04-21

### Fixed
- `install.sh` now finds the highest semver instead of using `/releases/latest` (which sorts by publish date, not version)
- Release fetch capped at `?per_page=100` — GitHub API max, single request
- Added `--version`/`-v` flag and `--install-dir` flag to `install.sh`

## [4.4.0] — 2026-04-21

### Changed — Gherkin-first cleanup sweep
- Rewrote `spec-kit` skill: story file IS the spec; dropped PROBLEM→SPEC→PLAN→TASKS state machine
- All doc examples (`ui-feature`, `api-endpoint`, `spike`) updated to Gherkin-first format
- Removed stale PROBLEM/SPEC/TASKS references from orchestrator, architect, and spec-kit agents
- Story frontmatter schema unified across template, validator, and spec-kit skill
- `copilot-instructions.md` and `stories.instructions` synced to canonical frontmatter
- Design and a11y gates made phase-aware (skip before design is approved)
- Pipeline status overlay (`P` key) added to TUI — reads `.claude/state/maple.json`
- `qa-cucumber` agent added; ADR enforcement gate wired into orchestrator
- `product-owner` agent deduplicated; `api-endpoint` example added to agents docs
- Added 8 missing agents to `AGENTS.md` quick-reference table

## [4.3.0] — 2026-04-16

### Added — Gherkin-first gates, orchestrator guardrails, superpowers
- `ship-safe` skill added to both `.claude/skills/` and `.opencode/skills/`
- `architect` and `orchestrator` agents updated to reference `ship-safe`
- OpenCode agents corrected to read `.opencode/skills/` (was wrongly reading `.claude/skills/`)
- `copilot-instructions.md` fleshed out with full commands, agents, and skills tables
- Ship-safe CI gate made opt-in via `ENABLE_SHIP_SAFE=true` repo variable
- TUI binary moved to `tui/`; `template/` stays at repo root via symlink
- Build dance (`rm symlink → cp -rL → go build → restore symlink`) wired into Makefile and CI
- Test suite updated: Go file checks target `tui/`, superpowers phase guarded by directory check
- `maple` binary and `.maple.json` added to `.gitignore`

## [3.7.0] — 2026-04-16

### Added — Phase VIII: Reference Implementations
- `docs/examples/ui-feature/` — full `new-ui-feature` superpower walkthrough (`ui: true` story, design gates, a11y)
- `docs/examples/api-endpoint/` — `api-endpoint` superpower (`ui: false`, no design gate)
- `docs/examples/spike/` — `spike/*` branch, Spec-Kit skip, no-production-merge pattern
- `README.md` updated: `maple` binary, `go build` quickstart, TUI keybindings, phase summary

## [3.6.0] — 2026-04-16

### Added — Phase VII: Enforcement Gates
- `template/lefthook.yml`: pre-push gates (spec-kit, frontmatter, design-approved, a11y), pre-commit no-secrets, post-merge log rotation
- `template/scripts/sdlc/validate-frontmatter.sh` — required fields + priority enum check
- `template/scripts/sdlc/a11y-gate.sh` — blocks on unresolved WCAG 2.2 AA critical/serious violations
- `template/scripts/sdlc/design-approved-gate.sh` — blocks if wireframe/mockup not approved for `ui: true` stories
- `template/scripts/sdlc/spec-kit-gate.sh` — checks spec progression integrity
- `template/scripts/sdlc/rotate-logs.sh` — rotates `.claude/logs/` at 10 MB, keeps last 5 compressed
- `template/.github/workflows/sdlc-gates.yml` — CI mirrors all lefthook gates
- `template/scripts/bootstrap-branch-protection.sh` — sets required checks + PR review on main

## [3.5.0] — 2026-04-16

### Added — Phase VI: TUI `maple`
- `tui/` — Go + Bubble Tea TUI binary; cross-compiles to macOS/Linux/Windows with zero runtime deps
- 8-pane dashboard: Stories, Agents, PRs, QA, Design, Logs, Help, Superpowers
- 5 built-in themes: `tokyo-night` (default), `catppuccin-mocha`, `gruvbox`, `nord`, `everforest`
- Animated logo: 3D depth shading, idle shimmer, theme-reactive palette per §5.10
- `--no-animate` flag; narrow terminal (<80 cols) degrades to single-pane scrolling mode
- Command mode (`:`), search mode (`/`), help overlay (`?`), superpower picker (`F`)
- `tui/README.md`: build, install, cross-compile, keybindings, theme switching

## [3.4.0] — 2026-04-16

### Added — Phase V: Superpowers
- `template/.claude/superpowers/schema.yaml` — YAML contract for all superpower declarations
- Built-in superpowers: `new-ui-feature`, `api-endpoint`, `bugfix`, `design-refresh`
- `superpower-runner` skill: YAML loader, stage executor, gate handler, resume logic (Claude Code + OpenCode)

## [3.3.0] — 2026-04-16

### Added — Phase IV: Spec-Kit
- `spec-kit` skill: PROBLEM→SPEC→PLAN→TASKS state machine, story emitter, skip logic for spike/chore branches
- `spec-kit.md` agent: approval-halted intake, hands off to orchestrator after TASKS.md approved
- `docs/specs/` directory standard; pre-DISCOVER integration into orchestrator (both platforms)

## [3.2.0] — 2026-04-16

### Added — Phase III: Design & UX Suite
- 6 new agents: `ux-researcher`, `wireframe-architect`, `visual-identity-designer`, `design-system-author`, `ui-mockup-builder`, `a11y-auditor` (Claude Code + OpenCode)
- 6 new skills: `wireframe`, `visual-identity`, `design-tokens`, `mockup`, `component-scaffold`, `a11y-audit`
- `docs/design/` directory standard: `research/`, `wireframes/`, `mockups/`, `identity/`, `system/components/`
- Orchestrator updated with design intake gate for `ui: true` stories

## [3.1.0] — 2026-04-16

### Added — Phase II: GitHub Integration Skills
- `gh-issues`, `gh-projects`, `gh-labels-milestones`, `gherkin-authoring`, `story-issue-sync`, `cucumber-automation` skills (Claude Code + OpenCode)
- `cmd_project` in `scripts/maple`: creates GitHub Project v2, stores `project_number` + `project_node_id` in `project.config.yaml`

## [3.0.0] — 2026-04-16

### Added — Phase I: Foundation
- `template/.claude/schemas/project.config.schema.json` — JSON Schema for `project.config.yaml`
- `template/project.config.yaml` — canonical per-project config with all v3 fields
- `template/docs/stories/_template.md` — story file template with frontmatter + Gherkin + DoD
- `template/docs/dod/definition-of-done.md` — default DoD including design and a11y gates
- Expanded label set: 11 label groups in `cmd_labels`

### Changed — Simplification track
- Removed swarm mode (`scripts/swarm-*.sh`) and hardcoded model names from all 54 agent files
- Removed empty scaffold stubs (`template/app/`, `template/common/`, `template/infra/`)
- Baked BusinessRepo/SOLID standards into `template/CLAUDE.md`, orchestrators, and architect agent
- Rewrote `scripts/maple`: 1,254 → ~700 lines

