```
                                        ▄█▄
                                       █████
                                      ███████
                                     █████████
                                    ███████████
                        ▄▄         █████████████         ▄▄
                       ████▄      ███████████████      ▄████
                       ███████▄  █████████████████  ▄███████
                       ██████████████████████████████████████
                        ████████████████████████████████████
              ▄▄        ████████████████████████████████████        ▄▄
             ████▄      ████████████████████████████████████      ▄████
      ▄▄    ███████▄   ██████████████████████████████████████   ▄███████    ▄▄
     █████▄▄██████████ ██████████████████████████████████████ ██████████▄▄█████
      ████████████████████████████████████████████████████████████████████████
       ██████████████████████████████████████████████████████████████████████
        ████████████████████████████████████████████████████████████████████
         ██████████████████████████████████████████████████████████████████
      ▄▄▄███████████████████████████████████████████████████████████████████▄▄▄
         ██████████████████████████████████████████████████████████████████
           ██████████████████████████████████████████████████████████████
             ██████████████████████████████████████████████████████████
                ████████████████████████████████████████████████████
                   ██████████████████████████████████████████████
                      ████████████████████████████████████████
                         ██████████████████████████████████
                           ████████████████████████████████
                         ▀████████████  ██████  ████████████▀
                                        ██████
                                        ▓▓▓▓▓▓
                                        ▓▓▓▓▓▓
                                        ▒▒▒▒▒▒
                                        ▒▒▒▒▒▒
                                        ░░░░░░
                                        ░░░░░░


🍁 ▗▖  ▗▖ ▗▄▖ ▗▄▄▖ ▗▖   ▗▄▄▄▖ 🍁
🍁 ▐▛▚▞▜▌▐▌ ▐▌▐▌ ▐▌▐▌   ▐▌    🍁
🍁 ▐▌  ▐▌▐▛▀▜▌▐▛▀▘ ▐▌   ▐▛▀▀▘ 🍁
🍁 ▐▌  ▐▌▐▌ ▐▌▐▌   ▐▙▄▄▖▐▙▄▄▖ 🍁
```

[![CI](https://github.com/kinncj/MAPLE/actions/workflows/ci.yml/badge.svg)](https://github.com/kinncj/MAPLE/actions/workflows/ci.yml)
[![Integration Validation](https://github.com/kinncj/MAPLE/actions/workflows/validate-integrations.yml/badge.svg)](https://github.com/kinncj/MAPLE/actions/workflows/validate-integrations.yml)

**MAPLE** is the orchestration layer that connects Claude Code, OpenCode, and GitHub Copilot CLI into a unified, TDD-enforced development lifecycle.

> Based on: [Building MAPLE: Orchestrated Multi-Agent Systems with Claude Code and OpenCode](./ARTICLE.md)

<div align="center">
  <img src="./demo.gif" alt="MAPLE demo — maple init scaffolding a project" width="860">
  <br/>
  <sub><code>maple init</code> — scaffolding a new project from the CLI</sub>
</div>

---

## What is MAPLE?

**MAPLE** stands for **M**ulti-Agent · **A**rtifact-Driven · **P**hase-Gated · **L**ocal-First · **E**nforced.

### M — Multi-Agent Orchestration

The core engine runs on a federated network of specialist agents, replacing monolithic prompting with targeted expertise.

- **Specialist squads** — 27+ agents each with a defined role, restricted toolset, and specific scope. Extend with the Design & UX suite (`ux-researcher`, `visual-identity-designer`, `a11y-auditor`) for frontend-heavy projects.
- **Superpowers (composability)** — agents and deterministic skills compose into named workflows. A single keystroke in the TUI can fire `new-ui-feature`, chaining Spec-Kit, wireframing, mockup, and component scaffolding seamlessly.
- **Capability hierarchy** — Agents (reasoning) → Skills (deterministic mechanics) → MCPs (last resort). The orchestrator never writes code; it only delegates.

### A — Artifact-Driven Specification

Before a single line of implementation code is written, the system demands clear, human-approved artifacts.

- **Spec-Kit layer** — enforces a linear progression of Problem → Spec → Plan → Tasks. The orchestrator refuses to enter implementation until these artifacts are materialized.
- **Design & brand tokens** — UI-bearing stories (`ui: true`) trigger design intake gates. ASCII/SVG wireframes, W3C DTCG design tokens (`tokens.json`), and high-fidelity mockups are generated and stored canonically in `docs/design/`.
- **Gherkin scenarios** — specifications are written as embedded Gherkin in Markdown story files, so product owners and automated QA agents speak the exact same language.

### P — Phase-Gated Pipeline

The SDLC is a rigid, visible state machine. No phase can be skipped.

**DISCOVER → ARCHITECT → PLAN → INFRA → IMPLEMENT → VALIDATE → DOCUMENT → FINAL GATE**

- **Automated sync** — bidirectional sync between local Markdown story files and GitHub Issues/Projects via the `story-issue-sync` skill. GitHub Project board is always the authoritative status source; local files own the narrative.
- **Human-in-the-loop** — agents prepare PRs and artifacts, humans approve them. The pipeline halts at predefined gates (wireframe approval, ADRs) to preserve technological sovereignty and oversight.

### L — Local-First Ecosystem

The developer experience prioritizes terminal dominance and minimal reliance on external SaaS.

- **`maple` TUI** — a 4-pane Go/BubbleTea dashboard (Stories · Agents · PRs · QA) inspired by lazydocker/lazygit. Boot check, animated maple leaf, live dashboard, skills marketplace (`F`), and a requirements wizard (`n`) — all in a self-contained binary with the template embedded.
- **Preserved aesthetics** — Canada-red animated maple leaf at boot, compact block-char wordmark in the dashboard, truecolor themes mapped to your local environment (Omarchy, Tokyo Night, and more).
- **Offline tolerance** — actions requiring network access degrade gracefully; local data (stories, agents, QA) always loads instantly.

### E — Enforced Execution & Guardrails

Rules, testing, and compliance are enforced, not just encouraged.

- **TDD & BDD automation** — the `qa-cucumber` agent extracts Gherkin from stories at build time, generating step definitions and running tests. Merges are hard-blocked until scenarios are green.
- **Strict Definitions of Done** — hook-enforced checks (`lefthook`) ensure DoD checklist items and WCAG 2.2 AA accessibility audits for UI components are verified before a push is allowed.
- **Architectural guardrails** — introducing an MCP or making cross-boundary data changes automatically triggers an ADR requirement. The Appendix C checklist gates against vendor lock-in and uncontrolled scope creep.

---

## Quick Start

**Pre-built binaries** — if you don't have Go installed:
```bash
curl -fsSL https://raw.githubusercontent.com/kinncj/MAPLE/main/scripts/install.sh | bash
```
Installs `maple` to `~/.tools/maple/bin/`. Add that to your `PATH`.

**Build from source**
```bash
git clone https://github.com/kinncj/MAPLE.git maple-src
cd maple-src
make build-tui              # produces ./maple
export PATH="$PWD:$PATH"   # or move to any directory on your PATH

cd your-project
maple init
maple req                   # write requirements → Gherkin story
```

---

## `maple` TUI

The `maple` binary is a self-contained Go/BubbleTea TUI. Run it inside any project initialized with `maple init`.

```
maple          # boot check → dashboard (if project.config.yaml exists)
maple init     # scaffold agents, skills, hooks, docs into current directory
maple req      # requirements wizard → Gherkin story file
maple --help   # all flags
```

### Keybindings

| Key | Action |
|---|---|
| `Tab` / `Shift+Tab` | Cycle panes |
| `j` / `k` | Move down / up |
| `s` `a` `p` `Q` | Jump to Stories / Agents / PRs / QA pane |
| `d` | Toggle Design artifacts pane (full-screen) |
| `l` | Toggle Skill Logs pane (full-screen) |
| `F` | Skills marketplace — browse, install, remove via skills.sh |
| `S` | Run `ship-safe` audit — security/quality scan with colored findings |
| `x` | Superpowers picker — browse and launch named agent/skill workflows |
| `P` | Pipeline status — show active superpower progress from `.claude/state/maple.json` |
| `o` | Open selected session in its agent (Agents pane: `claude --resume` or `opencode`) |
| `n` | New story → Gherkin requirements wizard |
| `u` | Update — re-sync template files (preserves your Makefile edits) |
| `r` | Reload all pane data |
| `/` | Search within active pane |
| `:` | Command mode (`:theme <name>`, `:update`, `:req`, `:help`) |
| `?` | Help overlay |
| `q` / `Ctrl+C` | Quit |

**Themes:** `tokyo-night` (default), `catppuccin-mocha`, `gruvbox`, `nord`, `everforest`. Switch with `:theme <name>` or auto-detected from `~/.config/omarchy/current/theme`.

---

## Commands (inside Claude Code or OpenCode)

| Command | What it does |
|---|---|
| `/feature "description"` | Full 8-phase pipeline |
| `/bugfix "description"` | Reproduce → fix → validate → CHANGELOG |
| `/validate` | Run full test suite |
| `/tdd "requirement"` | Single RED → GREEN → REFACTOR cycle |
| `/ship-safe` | Run `npx ship-safe audit .` — security scan, reports blockers by severity |

---

## Skills Marketplace

The `F` key in the TUI opens the skills.sh marketplace browser. Two tabs:

- **Installed** — shows all project and global skills, `d` to remove
- **Search** — type a query, `Enter` to search, `Enter` again to install

Skills are installed via `npx skills add <pkg> --all -y` and work across Claude Code, Cursor, and other supported editors.

MAPLE installs `obra/superpowers` automatically during `maple init` if `npx` is available.

---

## Documentation

| Doc | Contents |
|---|---|
| [Quickstart — Claude Code](./docs/quickstart-claude-code.md) | Install, scaffold, run your first feature |
| [Quickstart — Copilot CLI](./docs/quickstart-copilot-cli.md) | Install, enable Rubber Duck, run your first feature |
| [Quickstart — OpenCode](./docs/quickstart-opencode.md) | Install, configure providers, run your first feature |
| [The 8-Phase Pipeline](./docs/pipeline.md) | Phase details, TDD loop, Makefile contract, escalation policy |
| [The Agents](./docs/agents.md) | Agent roster, skills, adding custom agents |
| [Customization Guide](./docs/customization.md) | Add agents, restrict permissions, extend skills |
| [Architecture Article](./ARTICLE.md) | Design decisions, why specialist agents, CLI vs MCP |
| [Examples](./template/docs/specs/examples/) | UI feature, API endpoint, spike walk-throughs |

---

## Prerequisites

| Tool | Purpose | Install |
|---|---|---|
| [Go 1.22+](https://go.dev) | Build `maple` from source | `brew install go` |
| [Claude Code](https://claude.ai/claude-code) or [Copilot CLI](https://github.com/features/copilot/cli) or [OpenCode](https://opencode.ai) | Run the agents | see each link |
| [GitHub CLI](https://cli.github.com) | Issue, PR, project management | `brew install gh` |
| [Node.js](https://nodejs.org) | Playwright / Cucumber E2E tests + `npx skills` | `brew install node` |
| [Docker](https://docker.com) | Test infrastructure | docker.com |

> Go is only needed to build `maple` from source. Use the one-liner installer for a pre-built binary.

---

## License

AGPLv3 — see [LICENSE](./LICENSE) for details.

Copyright (C) 2025 Kinn Coelho Juliao <kinncj@protonmail.com>
