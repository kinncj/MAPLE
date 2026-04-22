![MAPLE](MAPLE_8bit.png)

[![CI](https://github.com/kinncj/MAPLE/actions/workflows/ci.yml/badge.svg)](https://github.com/kinncj/MAPLE/actions/workflows/ci.yml)
[![Integration Validation](https://github.com/kinncj/MAPLE/actions/workflows/validate-integrations.yml/badge.svg)](https://github.com/kinncj/MAPLE/actions/workflows/validate-integrations.yml)

**MAPLE** is the orchestration layer that connects Claude Code, OpenCode, and GitHub Copilot CLI into a unified, TDD-enforced development lifecycle. One binary installs everything: agents, skills, hooks, and a live project dashboard.

> Based on: [Building MAPLE: Orchestrated Multi-Agent Systems with Claude Code and OpenCode](./ARTICLE.md)

<div align="center">
  <img src="./demo.gif" alt="MAPLE demo — maple init scaffolding a project" width="860">
  <br/>
  <sub><code>maple init</code> — scaffolding a new project from the CLI</sub>
</div>

---

## Install

**macOS / Linux** — one line, no Go required:

```bash
curl -fsSL https://raw.githubusercontent.com/kinncj/MAPLE/main/scripts/install.sh | bash
```

Installs `maple` and `rtk` to `~/.tools/maple/bin/`. Add to `PATH`:

```bash
echo 'export PATH="$HOME/.tools/maple/bin:$PATH"' >> ~/.zshrc && source ~/.zshrc
```

**Windows** (PowerShell):

```powershell
irm https://raw.githubusercontent.com/kinncj/MAPLE/main/scripts/install.ps1 | iex
```

**Build from source** (Go 1.22+):

```bash
git clone https://github.com/kinncj/MAPLE.git && cd MAPLE
make build-tui        # → ./maple
sudo mv maple /usr/local/bin/
```

---

## Quick Start

```bash
cd your-project
maple init            # scaffold agents, skills, hooks, Makefile
maple                 # open the dashboard
```

Inside the dashboard press `n` to write requirements and generate a Gherkin story, then hand off to your harness:

```
/feature "user can reset password via email link"
```

---

## What is MAPLE?

**M**ulti-Agent · **A**rtifact-Driven · **P**hase-Gated · **L**ocal-First · **E**nforced.

### M — Multi-Agent Orchestration

27+ specialist agents, each with a defined role and restricted toolset. The orchestrator never writes code — it delegates to the right specialist every time.

- **TAFFY** — task-isolated, async, fault-tolerant, file-synced, YAML-driven workflow engine. `new-ui-feature` fires Spec-Kit → wireframe → mockup → component scaffold in one command. See [TAFFY](#taffy) below.
- **Capability hierarchy** — Agents (reasoning) → Skills (deterministic) → MCPs (last resort).

### A — Artifact-Driven Specification

Implementation never starts until human-approved artifacts exist.

- A Gherkin story file in `docs/stories/` is required before any code is written.
- `ui: true` stories require approved wireframes and mockups.
- Design tokens (`tokens.json`) and ADRs are generated and stored in `docs/design/` and `docs/adrs/`.

### P — Phase-Gated Pipeline

Eight phases, in order, no skipping:

**DISCOVER → ARCHITECT → PLAN → INFRA → IMPLEMENT → VALIDATE → DOCUMENT → FINAL GATE**

Agents prepare; humans approve at defined gates. File-based handoffs between the TUI and agents via `.claude/state/`.

### L — Local-First

- Self-contained `maple` binary — template embedded, no runtime dependencies.
- BubbleTea dashboard with live Stories, Agents, PRs, and QA panes.
- RTK token optimizer wired as a `PreToolUse` hook — 60–90% fewer tokens on build/grep/test output, transparent to all commands.

### E — Enforced

- TDD: failing tests before implementation, always.
- `lefthook` gates on pre-push: spec-kit, frontmatter, design-approved, a11y.
- WCAG 2.2 AA audit required for all `ui: true` stories before merge.

---

## `maple` TUI

Run `maple` inside any project initialized with `maple init`. Recommended: run inside **tmux** or **zellij** so harnesses open in new tabs without closing the dashboard.

```bash
tmux new-session -s work   # then: maple
# or
zellij                      # then: maple
```

### Keybindings

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Cycle panes |
| `j` / `k` | Move cursor down / up |
| `s` `a` `p` `Q` | Jump to Stories / Agents / PRs / QA pane |
| `Enter` | Open detail (story, session, PR, test file) |
| `o` | Open selected session + auto-pin it (`claude --resume` or `opencode --session`) |
| `p` | Pin selected session to `.claude/state/sessions.json` |
| `L` | Launch overlay — pick harness, type optional command, open in new tab |
| `x` | Taffy picker — select a named workflow to launch |
| `P` | Pipeline status — live view of active taffy workflow from `.claude/state/maple.json`; `[a]` to approve a gate, `[c]` to clear stale state |
| `R` | RTK harness selector — toggle which harnesses get `rtk init` wired |
| `r` | Run selected test (QA pane) / reload all panes |
| `d` | Design artifacts pane (full-screen toggle) |
| `l` | Logs pane (full-screen toggle) |
| `n` | Requirements wizard → new Gherkin story |
| `u` | Update — re-sync template files |
| `S` | `ship-safe` security audit |
| `F` | Skills marketplace — browse, install, remove |
| `/` | Search within active pane |
| `:` | Command mode (`:theme <name>`, `:update`, `:req`, `:help`) |
| `?` | Help overlay |
| `q` / `Ctrl+C` | Quit |

**Themes:** `tokyo-night` (default) · `catppuccin-mocha` · `gruvbox` · `nord` · `everforest`

Switch with `:theme <name>` or auto-detected from `~/.config/omarchy/current/theme`.

### CLI Commands

```bash
maple                          # boot check → dashboard
maple init                     # scaffold MAPLE into current directory
maple init --force             # overwrite existing files
maple req                      # requirements wizard → Gherkin story
maple resume-session           # resume pinned session (reads sessions.json)
maple resume-session claude    # resume specifically the pinned Claude session
maple labels                   # bootstrap GitHub label set
maple project                  # create GitHub Project v2
maple self-update              # upgrade maple to the latest release
maple --version                # print version
maple --no-animate             # skip animations (SSH / slow terminals)
```

---

## Agent Commands (inside Claude Code or OpenCode)

| Command | What it does |
|---------|-------------|
| `/feature "description"` | Full 8-phase pipeline |
| `/bugfix "description"` | Reproduce → fix → validate → CHANGELOG |
| `/validate` | Run full test suite |
| `/tdd "requirement"` | RED → GREEN → REFACTOR cycle |
| `/pipeline-runner <name>` | Launch a named taffy workflow |
| `/ship-safe` | Security/quality scan, reports blockers by severity |

---

## TAFFY

**T**ask-Isolated · **A**synchronous · **F**ault-Tolerant · **F**ile-Synced · **Y**AML-Driven

MAPLE sets the rules. TAFFY runs the jobs.

### T — Task-Isolated

Each agent job runs in a dedicated subprocess. A 60-second generation loop from Claude Code or OpenCode never freezes the TUI — you keep reviewing PRs or reading specs in other panes while the agent works.

### A — Asynchronous

Fire-and-forget from the orchestrator's perspective. MAPLE hands off a job and moves on. TAFFY manages the waiting, polling, and completion signal so the rest of the pipeline stays non-blocking.

### F — Fault-Tolerant

LLMs hallucinate, APIs rate-limit, agents loop. TAFFY is the circuit breaker:

- **Hard timeouts** — if an agent is stuck, TAFFY kills the process and flags the job `FAILED`.
- **Rate-limit elasticity** — on a `429`, TAFFY pauses, sets state to `RATE_LIMITED`, and resumes when the window clears.
- **3-strike escalation** — three consecutive failures on any stage → escalate to human.

### F — File-Synced

No Redis, no message broker. TAFFY broadcasts state by writing to `.claude/state/maple.json`. The TUI tails this file:

- `RUNNING` → spinner animates
- `PAUSED` → gate indicator, press `[a]` to approve
- `BLOCKED` / `RATE_LIMITED` → flagged yellow
- `DONE` / `FAILED` → final status

### Y — YAML-Driven

Workflows are stateless and deterministic. Each job is a YAML manifest — stage list, agent assignments, gates, guards, artifact expectations. No hidden state, no magic.

---

### Built-in workflows

| Name | What it runs |
|------|-------------|
| `new-ui-feature` | Spec-Kit → wireframe → mockup → component scaffold → TDD → a11y audit |
| `api-endpoint` | Spec-Kit → architect (ADR) → TDD → implement → contract test → docs |
| `bugfix` | Reproduce → root-cause analysis → fix → regression test → CHANGELOG |
| `design-refresh` | Visual identity → design tokens → component audit → mockup update |

### Running a workflow

**From the TUI** — press `x` to open the taffy picker, select a workflow, and it launches in your active harness.

**From any harness** — `/pipeline-runner <name>` works across Claude Code, OpenCode, and Copilot CLI:
```
/pipeline-runner new-ui-feature
/pipeline-runner api-endpoint
```

### Harness support

Workflow files are mirrored across all harnesses — same YAML, same behavior:

| Harness | Workflow dir | Skill |
|---------|-------------|-------|
| Claude Code | `.claude/taffy/` | `.claude/skills/pipeline-runner/` |
| OpenCode | `.opencode/taffy/` | `.opencode/skills/pipeline-runner/` |
| GitHub Copilot CLI | `.github/copilot-instructions.md` | Same `/pipeline-runner` command |

### Human-approval gates

Stages with `gate: human-approval` pause and write `PAUSED` to `maple.json`. The TUI shows the blocked stage in `[P]`. Press `a` to approve, or type "approved" in the harness chat.

### Custom workflows

Add a YAML file to `.claude/taffy/` (and mirror to `.opencode/taffy/`). Schema reference: `.claude/taffy/schema.yaml`.

```yaml
name: db-migration
description: "Safe database migration: schema → backfill → validate → deploy"
version: "1.0.0"
tags: [infra, database]
stages:
  - name: spec
    agent: spec-kit
    gate: human-approval
  - name: schema
    agent: architect
    depends_on: [spec]
  - name: tests
    agent: qa
    depends_on: [schema]
  - name: implement
    pipeline: standard
    depends_on: [tests]
    gate: human-approval
```

---

## Shared State Protocol

The TUI and agents communicate through files in `.claude/state/`:

| File | Owner | Purpose |
|------|-------|---------|
| `maple.json` | Skill writes pipeline fields; TUI writes `state`/`ts` | Taffy pipeline progress |
| `approval-pending.txt` | Skill creates; TUI deletes on approve | Human-in-the-loop gate handoff |
| `sessions.json` | TUI writes on `p`/`o`; skill reads for resume | Pinned harness session IDs |
| `rtk-harnesses.json` | TUI writes after `R` overlay; skill reads | Which harnesses have rtk wired |

Both sides **merge** rather than overwrite `maple.json` — the skill owns the taffy fields, the TUI owns `state` and `ts`.

---

## Skills Marketplace

`F` opens the skills.sh marketplace. Two tabs:

- **Installed** — all project and global skills; `d` to remove
- **Search** — type a query, `Enter` to search, `Enter` again to install

Skills install via `npx skills add <pkg> --all -y` and work across Claude Code, Cursor, and other editors.

---

## Documentation

| Doc | Contents |
|-----|---------|
| [Quickstart — Claude Code](./docs/quickstart-claude-code.md) | Install, scaffold, first feature |
| [Quickstart — Copilot CLI](./docs/quickstart-copilot-cli.md) | Install, enable Rubber Duck, first feature |
| [Quickstart — OpenCode](./docs/quickstart-opencode.md) | Install, configure providers, first feature |
| [The 8-Phase Pipeline](./docs/pipeline.md) | Phase details, TDD loop, Makefile contract |
| [The Agents](./docs/agents.md) | Full agent roster, skills, adding custom agents |
| [Customization Guide](./docs/customization.md) | Add agents, restrict permissions, extend skills |
| [Architecture Article](./ARTICLE.md) | Design decisions, why specialist agents |
| [Changelog](./CHANGELOG.md) | Full version history |

---

## Prerequisites

| Tool | Purpose | Required |
|------|---------|----------|
| [Claude Code](https://claude.ai/claude-code), [Copilot CLI](https://github.com/features/copilot/cli), or [OpenCode](https://opencode.ai) | Run the agents | At least one |
| [GitHub CLI `gh`](https://cli.github.com) | Issues, PRs, project management | Yes |
| [Go 1.22+](https://go.dev) | Build `maple` from source | Only for source builds |
| [Node.js](https://nodejs.org) | Cucumber E2E tests + `npx skills` | Optional |
| [Docker](https://docker.com) | Test infrastructure | Optional |

> Pre-built binaries for macOS/Linux/Windows are available on every [release](https://github.com/kinncj/MAPLE/releases). Go is only needed to build from source.

---

## License

AGPLv3 — see [LICENSE](./LICENSE) for details.

Copyright (C) 2025 Kinn Coelho Juliao <kinncj@protonmail.com>
