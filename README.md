```

   ▄▄▄▄   ▄▄▄▄▄    ▄▄▄▄▄▄▄   ▄▄▄▄▄   ▄▄▄  ▄▄▄   ▄▄▄▄   ▄▄▄▄▄▄
  ▄██▀▀██▄  ███    █████▀▀▀ ▄███████▄ ███  ███ ▄██▀▀██▄ ███▀▀██▄
  ███  ███  ███     ▀████▄  ███   ███ ███  ███ ███  ███ ███  ███
  ███▀▀███  ███       ▀████ ███▄█▄███ ███▄▄███ ███▀▀███ ███  ███
  ███  ███ ▄███▄   ███████▀  ▀█████▀  ▀██████▀ ███  ███ ██████▀
                          ▀▀

```

[![CI](https://github.com/kinncj/AI-Development-Squad-Template/actions/workflows/ci.yml/badge.svg)](https://github.com/kinncj/AI-Development-Squad-Template/actions/workflows/ci.yml)
[![Integration Validation](https://github.com/kinncj/AI-Development-Squad-Template/actions/workflows/validate-integrations.yml/badge.svg)](https://github.com/kinncj/AI-Development-Squad-Template/actions/workflows/validate-integrations.yml)

A production-ready template for running an **orchestrated, phase-gated, TDD-enforced** development pipeline with **specialist AI agents**. Runs on two platforms: **Claude Code** and **OpenCode**.

> Based on: [Building an AI Development Squad: Orchestrated Multi-Agent Systems with Claude Code and OpenCode](./ARTICLE.md)

<div align="center">
  <img src="./demo.gif" alt="AI Squad demo — ai-squad init scaffolding a project" width="860">
  <br/>
  <sub><code>ai-squad init</code> — scaffolding a new project from the CLI</sub>
</div>

---

## What This Is

Single-agent AI coding breaks down at scale. Context gets polluted, tests get skipped, implementations diverge from requirements. The fix is structural: split the work across agents with **enforced boundaries**, just like a real engineering team.

- **Specialist agents** — each with a defined role, restricted tools, and a specific scope
- **8-phase pipeline** — DISCOVER → ARCHITECT → PLAN → INFRA → IMPLEMENT → VALIDATE → DOCUMENT → FINAL GATE
- **Spec-Kit layer** — Problem → Spec → Plan → Tasks before any agent writes code
- **Design & UX suite** — wireframes, mockups, design tokens, visual identity, a11y audit as first-class pipeline stages
- **Superpowers** — named, versioned workflows composed from skills (e.g. `new-ui-feature` fires 11 stages in one keystroke)
- **TDD enforced** — QA writes failing tests first; implementation agents make them pass
- **GitHub integration** — every feature tracked via `gh` CLI, Projects v2, Issues, PRs
- **Reusable skills** — token-efficient CLI wrappers for Playwright, Docker, kubectl, Stripe, Supabase, gh, and more
- **Dual platform** — identical agent prompts for Claude Code and OpenCode
- **`squad` TUI** — persistent panel dashboard replacing the one-shot CLI

---

## Quick Start

### Option 1 — `squad` TUI (interactive, recommended)

```bash
# Requires Go 1.22+
git clone https://github.com/kinncj/AI-Development-Squad-Template.git ai-squad
cd ai-squad/tui && go build -o ../squad .
sudo mv ../squad /usr/local/bin/squad

cd your-project
squad
```

The TUI shows Stories, Active Agents, PRs, QA status, Design artifacts, and Logs in a four-pane dashboard. Press `?` for keybindings.

### Option 2 — `ai-squad` CLI (non-interactive / CI)

```bash
git clone https://github.com/kinncj/AI-Development-Squad-Template.git ~/.ai-squad
echo 'export PATH="$HOME/.ai-squad/scripts:$PATH"' >> ~/.zshrc
source ~/.zshrc

mkdir my-project && cd my-project
ai-squad init

# Start a feature (inside Claude Code)
/feature "user registration with email and OAuth"
```

---

## `squad` TUI — Keybindings

| Key | Action |
|---|---|
| `Tab` / `Shift+Tab` | Cycle panes |
| `j` / `k` | Move down / up |
| `s` `a` `p` `q` `d` `l` | Jump to pane |
| `F` | Fire Superpower (fuzzy picker) |
| `n` | New story / spike / ADR |
| `/` | Search |
| `:` | Command mode (`:kickoff <id>`, `:theme <name>`, `:sync`, `:resume <sp>`) |
| `?` | Help overlay |
| `Ctrl+c` | Quit |

Themes: `tokyo-night` (default), `catppuccin-mocha`, `gruvbox`, `nord`, `everforest`. Switch with `:theme <name>`.

---

## Commands (inside Claude Code or OpenCode)

| Command | What it does |
|---|---|
| `/feature "description"` | Full 8-phase pipeline |
| `/bugfix "description"` | Reproduce → fix → validate → CHANGELOG |
| `/validate` | Run full test suite |
| `/tdd "requirement"` | Single RED → GREEN → REFACTOR cycle |

---

## Superpowers

Superpowers compose skills and agents into named one-keystroke workflows. Press `F` in the TUI to pick one.

| Superpower | What it does |
|---|---|
| `new-ui-feature` | Spec-Kit → wireframe → visual identity → mockup → component-scaffold → a11y → 8-phase pipeline |
| `api-endpoint` | Spec-Kit → Gherkin → backend agent → contract tests → OpenAPI |
| `bugfix` | Triage → red test → fix → green |
| `design-refresh` | Visual identity → tokens → mockup regeneration → a11y |

Declare your own in `template/.claude/superpowers/<name>.yaml`.

---

## Documentation

| Doc | Contents |
|---|---|
| [Quickstart — Claude Code](./docs/quickstart-claude-code.md) | Install, scaffold, run your first feature |
| [Quickstart — OpenCode](./docs/quickstart-opencode.md) | Install, configure providers, run your first feature |
| [The 8-Phase Pipeline](./docs/pipeline.md) | Phase details, TDD loop, Makefile contract, escalation policy |
| [The Agents](./docs/agents.md) | Agent roster, skills, adding custom agents |
| [Customization Guide](./docs/customization.md) | Add agents, restrict permissions, extend skills |
| [Architecture Article](./ARTICLE.md) | Design decisions, why specialist agents, CLI vs MCP |
| [TUI README](./tui/README.md) | Build, install, cross-compile, keybindings |
| [Examples](./docs/examples/) | UI feature, API endpoint, spike walk-throughs |

---

## Prerequisites

| Tool | Purpose | Install |
|---|---|---|
| [Claude Code](https://claude.ai/claude-code) | Primary platform | `npm install -g @anthropic-ai/claude-code` |
| [OpenCode](https://opencode.ai) | Alternate platform | See opencode.ai |
| [GitHub CLI](https://cli.github.com) | Issue and PR management | `brew install gh` |
| [Go 1.22+](https://go.dev) | Build the `squad` TUI | `brew install go` |
| [Docker](https://docker.com) | Test infrastructure | docker.com |
| [Node.js](https://nodejs.org) | Playwright / Cucumber E2E tests | nodejs.org |

---

## License

AGPLv3 — see [LICENSE](./LICENSE) for details.

Copyright (C) 2025 Kinn Coelho Juliao <kinncj@protonmail.com>
