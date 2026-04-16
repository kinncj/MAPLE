# The 27 Agents

Every agent is a single markdown file with a defined role, restricted toolset, and a specific set of instructions. No agent can exceed its role — the orchestrator enforces boundaries. Use whichever model your subscription supports — quality comes from the prompts, not hardcoded model IDs.

---

## Architecture

```mermaid
graph TB
    Human["👤 Human"] -->|feature request| ORC

    subgraph Coordination
        ORC["orchestrator\nnever writes code"]
    end

    subgraph "Discovery & Design"
        PO["product-owner"]
        ARCH["architect"]
    end

    subgraph "Quality & Docs"
        QA["qa"]
        DOCS["docs"]
    end

    subgraph "Backend"
        DN[".NET"]
        JS["javascript"]
        TS["typescript"]
        JV["java"]
        SB["springboot"]
    end

    subgraph "Frontend"
        RV["react-vite"]
        NJ["nextjs"]
    end

    subgraph "Infrastructure"
        K8["kubernetes"]
        TF["terraform"]
        DK["docker"]
    end

    subgraph "Data & Services"
        PG["postgresql"]
        RD["redis"]
        SU["supabase"]
        VC["vercel"]
        ST["stripe"]
    end

    subgraph "ML / Data"
        DS["data-science"]
        DE["data-engineer"]
        TN["tensorflow"]
        PT["pytorch"]
        PA["pandas-numpy"]
        SK["scikit"]
        JP["jupyter"]
    end

    ORC --> PO & ARCH & QA & DOCS
    ORC --> DN & JS & TS & JV & SB & RV & NJ
    ORC --> K8 & TF & DK & PG & RD & SU & VC & ST
    ORC --> DS & DE & TN & PT & PA & SK & JP
```

---

## Full roster

| # | Agent | Role |
|---|---|---|
| 1 | `orchestrator` | Pipeline control — never writes code |
| 2 | `product-owner` | User stories, acceptance criteria |
| 3 | `architect` | ADR, contracts, threat models |
| 4 | `qa` | Write failing tests (RED) + validate (GREEN) |
| 5 | `dotnet` | .NET backend |
| 6 | `javascript` | Node.js / vanilla JS |
| 7 | `typescript` | TypeScript backend/libraries |
| 8 | `react-vite` | React + Vite SPA |
| 9 | `nextjs` | Next.js full-stack |
| 10 | `java` | Java backend (non-Spring) |
| 11 | `springboot` | Spring Boot |
| 12 | `kubernetes` | K8s manifests, Kustomize, Helm |
| 13 | `terraform` | Terraform IaC |
| 14 | `docker` | Dockerfiles, Compose |
| 15 | `postgresql` | Schema, migrations, RLS |
| 16 | `redis` | Caching, pub/sub, streams |
| 17 | `supabase` | Auth, RLS, Edge Functions |
| 18 | `vercel` | Deployment, edge config |
| 19 | `stripe` | Payments, webhooks, billing |
| 20 | `data-science` | EDA, stats, visualization |
| 21 | `data-engineer` | Pipelines, ETL, orchestration |
| 22 | `tensorflow` | TF/Keras models |
| 23 | `pytorch` | PyTorch models |
| 24 | `pandas-numpy` | Data manipulation, arrays |
| 25 | `scikit` | Classical ML, pipelines |
| 26 | `jupyter` | Notebooks, papermill |
| 27 | `docs` | Feature docs, CHANGELOG, Mermaid |

---

## The 17 skills

Skills are markdown files agents read before executing specific tasks. They encode CLI patterns, workflows, and tool usage. They are loaded on demand — not injected on every turn — keeping token costs low.

| Category | Skills |
|---|---|
| Process | `tdd-workflow`, `rfc-adr`, `threat-modeling` |
| Output | `mermaid-diagrams`, `finops-review`, `sre-review` |
| Tool / CLI | `playwright-cli`, `github-cli`, `docker-patterns`, `kubernetes-patterns`, `terraform-patterns`, `supabase-patterns`, `stripe-patterns`, `vercel-patterns`, `postgresql-patterns`, `redis-patterns`, `jupyter-patterns` |

Skills live in `.claude/skills/` and `.opencode/skills/` — both directories are kept in sync.

---

## Adding a custom agent

### Claude Code (`.claude/agents/`)

```markdown
---
name: rust
description: Rust systems programming specialist. Handles Cargo workspaces, async Tokio, and FFI.
---

You are the Rust specialist...
```

### OpenCode (`.opencode/agents/`)

```markdown
---
name: rust
temperature: 0.2
mode: code
tools:
  - read
  - edit
  - write
  - bash
permission:
  allow:
    - bash: ["cargo", "rustc", "rustfmt", "clippy"]
---

You are the Rust specialist...
```

Then add `rust` to the `permission.task` list in `.opencode/agents/orchestrator.md` and update `AGENTS.md`.

---

## Removing an agent

Delete the `.claude/agents/{name}.md` and `.opencode/agents/{name}.md` files, then remove the agent from the `permission.task` list in `orchestrator.md`. The orchestrator will no longer delegate to it.


Every agent is a single markdown file with a defined role, restricted toolset, and a specific model assignment. No agent can exceed its role — the orchestrator enforces boundaries.

---

## Architecture

```mermaid
graph TB
    Human["👤 Human"] -->|feature request| ORC

    subgraph Coordination
        ORC["orchestrator\nnever writes code"]
    end

    subgraph "Discovery & Design"
        PO["product-owner"]
        ARCH["architect"]
    end

    subgraph "Quality & Docs"
        QA["qa"]
        DOCS["docs"]
    end

    subgraph "Backend"
        DN[".NET"]
        JS["javascript"]
        TS["typescript"]
        JV["java"]
        SB["springboot"]
    end

    subgraph "Frontend"
        RV["react-vite"]
        NJ["nextjs"]
    end

    subgraph "Infrastructure"
        K8["kubernetes"]
        TF["terraform"]
        DK["docker"]
    end

    subgraph "Data & Services"
        PG["postgresql"]
        RD["redis"]
        SU["supabase"]
        VC["vercel"]
        ST["stripe"]
    end

    subgraph "ML / Data"
        DS["data-science"]
        DE["data-engineer"]
        TN["tensorflow"]
        PT["pytorch"]
        PA["pandas-numpy"]
        SK["scikit"]
        JP["jupyter"]
    end

    ORC --> PO & ARCH & QA & DOCS
    ORC --> DN & JS & TS & JV & SB & RV & NJ
    ORC --> K8 & TF & DK & PG & RD & SU & VC & ST
    ORC --> DS & DE & TN & PT & PA & SK & JP
```

---

## Full roster

| # | Agent | Role | Claude Code model | OpenCode model |
|---|---|---|---|---|
| 1 | `orchestrator` | Pipeline control — never writes code | `claude-opus-4-6` | `anthropic/claude-opus-4-6` |
| 2 | `product-owner` | User stories, acceptance criteria | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |
| 3 | `architect` | ADR, contracts, threat models | `claude-opus-4-6` | `anthropic/claude-opus-4-6` |
| 4 | `qa` | Write failing tests (RED) + validate (GREEN) | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |
| 5 | `dotnet` | .NET backend | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |
| 6 | `javascript` | Node.js / vanilla JS | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |
| 7 | `typescript` | TypeScript backend/libraries | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |
| 8 | `react-vite` | React + Vite SPA | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |
| 9 | `nextjs` | Next.js full-stack | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |
| 10 | `java` | Java backend (non-Spring) | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |
| 11 | `springboot` | Spring Boot | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |
| 12 | `kubernetes` | K8s manifests, Kustomize, Helm | `claude-sonnet-4-6` | `copilot/gpt-4.1` |
| 13 | `terraform` | Terraform IaC | `claude-sonnet-4-6` | `copilot/gpt-4.1` |
| 14 | `docker` | Dockerfiles, Compose | `claude-sonnet-4-6` | `copilot/gpt-4.1` |
| 15 | `postgresql` | Schema, migrations, RLS | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |
| 16 | `redis` | Caching, pub/sub, streams | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |
| 17 | `supabase` | Auth, RLS, Edge Functions | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |
| 18 | `vercel` | Deployment, edge config | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |
| 19 | `stripe` | Payments, webhooks, billing | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |
| 20 | `data-science` | EDA, stats, visualization | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |
| 21 | `data-engineer` | Pipelines, ETL, orchestration | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |
| 22 | `tensorflow` | TF/Keras models | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |
| 23 | `pytorch` | PyTorch models | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |
| 24 | `pandas-numpy` | Data manipulation, arrays | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |
| 25 | `scikit` | Classical ML, pipelines | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |
| 26 | `jupyter` | Notebooks, papermill | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |
| 27 | `docs` | Feature docs, CHANGELOG, Mermaid | `claude-sonnet-4-6` | `github-copilot/claude-sonnet-4.5` |

---

## Model routing rationale

```mermaid
graph LR
    subgraph "High reasoning — Anthropic API / Claude Code Max"
        OP["claude-opus-4-6"]
    end
    subgraph "Fast code gen — GitHub Copilot Enterprise"
        SON["claude-sonnet-4.5"]
    end
    subgraph "Config & manifests — GitHub Copilot"
        GPT["gpt-4.1"]
    end

    ORC2["orchestrator\narchitect"] --> OP
    IMPL2["25 implementation\nagents"] --> SON
    INFRA2["kubernetes\nterraform\ndocker"] --> GPT
```

- **Orchestrator and Architect** use the highest-reasoning model — they make design decisions that all other agents follow
- **Implementation agents** use the fast model — high-throughput code generation within well-defined task boundaries
- **Infrastructure agents** use GPT-4.1 — strong at manifest, HCL, and YAML generation under GitHub Copilot

---

## The 17 skills

Skills are markdown files agents read before executing specific tasks. They encode CLI patterns, workflows, and tool usage. They are loaded on demand — not injected on every turn — keeping token costs low.

| Category | Skills |
|---|---|
| Process | `tdd-workflow`, `rfc-adr`, `threat-modeling` |
| Output | `mermaid-diagrams`, `finops-review`, `sre-review` |
| Tool / CLI | `playwright-cli`, `github-cli`, `docker-patterns`, `kubernetes-patterns`, `terraform-patterns`, `supabase-patterns`, `stripe-patterns`, `vercel-patterns`, `postgresql-patterns`, `redis-patterns`, `jupyter-patterns` |

Skills live in `.claude/skills/` and `.opencode/skills/` — both directories are kept in sync.

---

## Adding a custom agent

### Claude Code (`.claude/agents/`)

```markdown
---
name: rust
description: Rust systems programming specialist. Handles Cargo workspaces, async Tokio, and FFI.
model: claude-sonnet-4-6
---

You are the Rust specialist...
```

### OpenCode (`.opencode/agents/`)

```markdown
---
name: rust
model: github-copilot/claude-sonnet-4.5
temperature: 0.2
mode: code
tools:
  - read
  - edit
  - write
  - bash
permission:
  allow:
    - bash: ["cargo", "rustc", "rustfmt", "clippy"]
---

You are the Rust specialist...
```

Then add `rust` to the `permission.task` list in `.opencode/agents/orchestrator.md` and update `AGENTS.md`.

---

## Removing an agent

Delete the `.claude/agents/{name}.md` and `.opencode/agents/{name}.md` files, then remove the agent from the `permission.task` list in `orchestrator.md`. The orchestrator will no longer delegate to it.
