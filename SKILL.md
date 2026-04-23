---
name: maple
description: "MAPLE AI Development Squad — install the full skill set for multi-agent development workflows: pipeline-runner, tdd-workflow, playwright-cli, github-cli, mermaid-diagrams, and more. Designed for Claude Code, OpenCode, and GitHub Copilot."
tags:
  - maple
  - ai-squad
  - multi-agent
  - tdd
  - pipeline
  - workflow
---

# MAPLE — AI Development Squad Skills

MAPLE is a multi-agent development framework. This package installs the core skill set.

## Install

```bash
# Full package (all skills)
npx skills add kinncj/maple --all -y

# Individual skills
npx skills add kinncj/maple@pipeline-runner --all -y
npx skills add kinncj/maple@tdd-workflow --all -y
npx skills add kinncj/maple@github-cli --all -y
```

## Skills included

| Skill | What it does |
|-------|-------------|
| `pipeline-runner` | Universal dispatcher: taffy workflows, skills, agents. Falls back to skills.sh registry. |
| `tdd-workflow` | Red → green → refactor TDD cycle enforced before implementation |
| `playwright-cli` | Browser automation and E2E test authoring |
| `github-cli` | Full `gh` CLI reference: issues, PRs, Actions, projects |
| `mermaid-diagrams` | Architecture diagrams: component, sequence, ER, state |
| `gherkin-authoring` | Gherkin story files — the spec before any implementation |
| `spec-kit` | Story scaffolding with correct IDs and frontmatter |
| `rfc-adr` | Architecture Decision Records in `docs/specs/` |
| `threat-modeling` | STRIDE threat modeling for every feature boundary |
| `sre-review` | SLOs, alerting, runbooks, rollback before launch |

## Harness support

Works with Claude Code, OpenCode, GitHub Copilot, and Cursor.

## Learn more

- [GitHub](https://github.com/kinncj/maple)
- [Install maple](https://github.com/kinncj/maple/releases/latest)
