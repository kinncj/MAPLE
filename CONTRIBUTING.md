# Contributing to MAPLE

Thank you for contributing. This is an open-source project licensed under AGPLv3.

---

## What You Can Contribute

| Area | Examples |
|---|---|
| **New agents** | Add a specialist for a new language, framework, or service |
| **Improved agent prompts** | Sharpen existing agent instructions, tool restrictions, or escalation rules |
| **New skills** | Add a `SKILL.md` for a tool or workflow not yet covered |
| **New commands** | Add a `/command.md` for a common pipeline pattern |
| **CLI improvements** | Improve `scripts/maple` (new subcommands, UX, error handling) |
| **Tests** | Add tests in `tests/` for new or existing functionality |
| **Documentation** | Improve `README.md`, `ARTICLE.md`, or agent descriptions |
| **Bug fixes** | Fix incorrect model IDs, broken frontmatter, bad shell syntax |

---

## Getting Started

```bash
# 1. Fork and clone
git clone https://github.com/YOUR_USERNAME/MAPLE.git
cd MAPLE

# 2. Install the CLI globally from your fork
echo 'export PATH="'$(pwd)'/scripts:$PATH"' >> ~/.zshrc
source ~/.zshrc

# 3. Run the test suite to confirm everything passes
bash tests/run_all.sh
```

---

## Repository Structure

```
.
├── scripts/
│   └── maple              # Global CLI (init · labels · project)
├── tui/                      # maple TUI binary (Go + Bubble Tea)
├── template/                 # Everything copied on maple init
│   ├── .claude/agents/       # 34 Claude Code agent definitions
│   ├── .claude/commands/     # slash commands
│   ├── .claude/skills/       # 31 skill files
│   ├── .claude/superpowers/  # composed named workflows
│   ├── .opencode/agents/     # 34 OpenCode agent definitions (mirrored)
│   ├── .opencode/commands/
│   ├── .opencode/skills/
│   ├── .github/workflows/    # sdlc-gates CI
│   ├── scripts/sdlc/         # gate scripts (validate-frontmatter, a11y, design, spec-kit, rotate-logs)
│   ├── lefthook.yml          # pre-push, pre-commit, post-merge hooks
│   ├── Makefile
│   ├── CLAUDE.md
│   ├── AGENTS.md
│   └── opencode.json
├── docs/
│   ├── examples/             # ui-feature, api-endpoint, spike walk-throughs
│   ├── specs/                # spec-kit outputs
│   └── design/               # design artifacts (wireframes, mockups, tokens)
├── tests/
│   ├── cli/                  # CLI tests (218 assertions)
│   └── features/             # generated .feature files (qa-cucumber output)
└── .github/
    ├── workflows/            # ci.yml, validate-integrations.yml
    ├── copilot-instructions.md
    └── pull_request_template.md
```

---

## CI/CD

Every pull request runs:

| Job | What it checks |
|---|---|
| **Lint** | `shellcheck` on all `.sh` files |
| **CLI Tests** | `maple help`, `init`, file count assertions |
| **Template Validation** | Structure, agent frontmatter, skills, commands |
| **Model ID Audit** | No stale date-suffixed IDs; Orchestrator/Architect use Opus |

The `ci-gate` job fails the PR if any of the above fail. PRs cannot be merged with a failing gate.

---

## Adding an Agent

### 1. Claude Code (`.claude/agents/{name}.md`)

```markdown
---
name: my-agent
description: One-line description shown in the agent picker.
model: claude-sonnet-4-6
---

You are a specialist agent in an orchestrated, phase-gated pipeline.

Hard rules:
- Only work on tasks assigned by @orchestrator.
- Never skip quality gates.
- Read relevant skills from `.claude/skills/` before executing tasks.
- After 3 consecutive failures, stop and report to human.
```

### 2. OpenCode (`template/.opencode/agents/{name}.md`)

```markdown
---
description: One-line description.
mode: subagent
model: github-copilot/claude-sonnet-4.5
temperature: 0.2
tools:
  write: true
  edit: true
  bash: true
permission:
  edit: ask
  bash:
    "*": ask
    "make *": allow
    "git status*": allow
    "git diff*": allow
    "gh *": allow
  webfetch: deny
---

[Same body as Claude Code agent]
```

### 3. Register the agent

- Add to `permission.task` list in `template/.opencode/agents/orchestrator.md`
- Add a row to `template/AGENTS.md`
- Update the agent count in `README.md`

### 4. Add a test

The test in `tests/template/test_agents.sh` validates frontmatter automatically for all agents. If your agent has special requirements, add an assertion there.

---

## Model Assignment Rules

| Role | Claude Code | OpenCode |
|---|---|---|
| Orchestrator, Architect | `claude-opus-4-6` | `anthropic/claude-opus-4-6` |
| All implementation agents | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |
| Kubernetes, Terraform, Docker | `claude-sonnet-4-6` | `copilot/gpt-4.1` |

Never use date-suffixed model IDs (e.g., `claude-sonnet-4-20250514`). The CI model audit will catch these.

---

## Conventional Commits

```
feat:     new agent, skill, command, or CLI subcommand
fix:      bug fix in existing agent, script, or frontmatter
test:     new or updated test files
docs:     README, ARTICLE, CONTRIBUTING, or agent description
infra:    Makefile, docker-compose, GitHub Actions
chore:    dependency updates, formatting, renames
```

---

## Pull Request Checklist

- [ ] `bash tests/run_all.sh` passes locally
- [ ] New agents have correct frontmatter on both platforms
- [ ] No stale model IDs
- [ ] `shellcheck` passes on any new/modified `.sh` files
- [ ] `CHANGELOG.md` updated if user-visible change

---

## Code of Conduct

Be direct, technical, and respectful. This project values correctness and clarity over quantity of output. Reviewers will prioritize correctness of agent prompts, model assignments, and CI behavior.

---

## License

By contributing you agree your work will be licensed under AGPLv3.

Copyright (C) 2025 Kinn Coelho Juliao <kinncj@protonmail.com>
