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

A production-ready template for running an **orchestrated, phase-gated, TDD-enforced** development pipeline with **specialist AI agents**. Runs on three platforms: **Claude Code**, **GitHub Copilot CLI**, and **OpenCode**.

> Based on: [Building an AI Development Squad: Orchestrated Multi-Agent Systems with Claude Code and OpenCode](./ARTICLE.md)

<div align="center">
  <img src="./demo.gif" alt="AI Squad demo — squad init scaffolding a project" width="860">
  <br/>
  <sub><code>squad init</code> — scaffolding a new project from the CLI</sub>
</div>

---

## What This Is

Single-agent AI coding breaks down at scale. Context gets polluted, tests get skipped, implementations diverge from requirements. The fix is structural: split the work across agents with **enforced boundaries**, just like a real engineering team.

- **Specialist agents** — each with a defined role, restricted tools, and a specific scope
- **8-phase pipeline** — DISCOVER → ARCHITECT → PLAN → INFRA → IMPLEMENT → VALIDATE → DOCUMENT → FINAL GATE
- **Spec-Kit layer** — Problem → Spec → Plan → Tasks before any agent writes code
- **Design & UX suite** — wireframes, mockups, design tokens, visual identity, a11y audit as first-class pipeline stages
- **Rubber Duck** — second-opinion reviewer invoked at plan, code, and test checkpoints; backed by Copilot CLI's built-in cross-model reviewer when using `/experimental`
- **Superpowers** — named, versioned workflows composed from skills (e.g. `new-ui-feature` fires 11 stages in one keystroke)
- **TDD enforced** — QA writes failing tests first; implementation agents make them pass; proper Playwright patterns enforced (no `window.fetch` overrides)
- **GitHub integration** — every feature tracked via `gh` CLI, Projects v2, Issues, PRs; stories auto-sync on write
- **Reusable skills** — token-efficient CLI wrappers for Playwright, Docker, kubectl, Stripe, Supabase, gh, and more
- **Three platforms** — identical agent prompts for Claude Code, GitHub Copilot CLI, and OpenCode
- **`squad` TUI** — interactive dashboard and `init` / `req` wizard; self-contained binary with template embedded

---

## Quick Start

```bash
git clone https://github.com/kinncj/AI-Development-Squad-Template.git ai-squad
cd ai-squad
make build-tui              # produces ./squad
export PATH="$PWD:$PATH"   # or move to any directory on your PATH

cd your-project
squad init
squad req                   # write requirements → Gherkin story
```

Open your project in **Claude Code**, **GitHub Copilot CLI**, or **OpenCode**, then run `/feature "your feature description"`.

> **Releases** — if you don't have Go installed, grab a pre-built binary:
> ```bash
> curl -fsSL https://raw.githubusercontent.com/kinncj/AI-Squad/main/scripts/install.sh | bash
> ```
> This installs `squad` to `~/.tools/ai-squad/bin/`. Add that to your `PATH`.

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
| [Quickstart — Copilot CLI](./docs/quickstart-copilot-cli.md) | Install, enable Rubber Duck, run your first feature |
| [Quickstart — OpenCode](./docs/quickstart-opencode.md) | Install, configure providers, run your first feature |
| [The 8-Phase Pipeline](./docs/pipeline.md) | Phase details, TDD loop, Makefile contract, escalation policy |
| [The Agents](./docs/agents.md) | Agent roster, skills, adding custom agents |
| [Customization Guide](./docs/customization.md) | Add agents, restrict permissions, extend skills |
| [Architecture Article](./ARTICLE.md) | Design decisions, why specialist agents, CLI vs MCP |
| [Examples](./docs/examples/) | UI feature, API endpoint, spike walk-throughs |

---

## Prerequisites

| Tool | Purpose | Install |
|---|---|---|
| [Go 1.22+](https://go.dev) | Build `squad` from source | `brew install go` |
| [Claude Code](https://claude.ai/claude-code) or [Copilot CLI](https://github.com/features/copilot/cli) or [OpenCode](https://opencode.ai) | Run the agents | see each link |
| [GitHub CLI](https://cli.github.com) | Issue, PR, project management | `brew install gh` |
| [Docker](https://docker.com) | Test infrastructure | docker.com |
| [Node.js](https://nodejs.org) | Playwright / Cucumber E2E tests | nodejs.org |

> Go is only needed to build `squad` from source. If you prefer a pre-built binary, use the one-liner installer above.

---

## License

AGPLv3 — see [LICENSE](./LICENSE) for details.

Copyright (C) 2025 Kinn Coelho Juliao <kinncj@protonmail.com>
