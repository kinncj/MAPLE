# Quickstart — Claude Code

This guide gets you from zero to a running feature pipeline using **Claude Code** and the AI Development Squad template.

---

## Prerequisites

| Tool | Install |
|---|---|
| [Claude Code](https://claude.ai/claude-code) | `npm install -g @anthropic-ai/claude-code` |
| [GitHub CLI](https://cli.github.com) | `brew install gh` |
| [Git](https://git-scm.com) | pre-installed on macOS/Linux |
| [Node.js](https://nodejs.org) | `brew install node` (for Playwright E2E tests) |
| [Docker](https://docker.com) | [docker.com/get-started](https://docker.com/get-started) |

---

## Installation

```mermaid
flowchart TD
    A["Clone the template globally"] --> B["Add scripts/ to PATH"]
    B --> C["Scaffold a new project"]
    C --> D["Authenticate gh CLI"]
    D --> E["Start your first feature"]
```

### 1. Install the CLI globally

```bash
git clone https://github.com/kinncj/AI-Development-Squad-Template.git ~/.ai-squad
echo 'export PATH="$HOME/.ai-squad/scripts:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

Verify:

```bash
ai-squad help
```

### 2. Scaffold a new project

```bash
mkdir my-project && cd my-project
ai-squad init
```

`ai-squad init` will:
- Copy all agent definitions, commands, and skills into the project
- Initialize a git repository
- Install npm dependencies (Playwright)
- Offer to bootstrap GitHub labels

### 3. Connect a remote repository

```bash
gh auth login
gh repo create my-project --public --push --source=.
```

### 4. Bootstrap GitHub labels (optional but recommended)

```bash
ai-squad labels
```

This creates the full label set the orchestrator uses to track pipeline phases.

---

## Running your first feature

Open Claude Code in your project directory and run:

```
/feature "describe your feature here"
```

Claude Code will invoke the orchestrator, which drives all 8 phases sequentially in a single session. Use Claude Code's built-in sub-agent support to run specialist agents in parallel if needed.

---

## Available commands

Run these inside Claude Code with `/command-name`:

| Command | What it does |
|---|---|
| `/feature "description"` | Full 8-phase pipeline from discovery to PR |
| `/build-feature "description"` | Alias for `/feature` |
| `/bugfix "description"` | Reproduce → fix → validate → CHANGELOG |
| `/validate` | Run the full test suite (no discovery/architecture) |
| `/tdd "requirement"` | Single RED → GREEN → REFACTOR cycle |

---

## Workflow overview

```mermaid
sequenceDiagram
    participant H as Human
    participant ORC as Orchestrator
    participant PO as @product-owner
    participant ARCH as @architect
    participant QA as @qa
    participant DEV as Specialist agents
    participant DOCS as @docs

    H->>ORC: /feature "description"
    ORC->>PO: Phase 1 — write stories & acceptance criteria
    PO-->>H: stories.md (human gate)
    H-->>ORC: approved
    ORC->>ARCH: Phase 2 — ADR, contracts, threat model
    ARCH-->>H: architecture.md (human gate)
    H-->>ORC: approved
    ORC->>ORC: Phase 3 — decompose into plan.md
    ORC->>DEV: Phase 4 — spin up infra
    loop Each task in plan.md
        ORC->>QA: write failing test (RED)
        ORC->>DEV: make test pass (GREEN)
        ORC->>DEV: refactor
    end
    ORC->>QA: Phase 6 — full validation suite
    ORC->>DOCS: Phase 7 — docs, CHANGELOG
    ORC->>H: Phase 8 — PR created ✅
```

---

## Project structure after `ai-squad init`

```
my-project/
├── .claude/
│   ├── agents/          # 27 agent definitions
│   ├── commands/        # /feature, /bugfix, /validate, /tdd
│   └── skills/          # 17 reusable skill files
├── .opencode/           # Mirror for OpenCode platform
├── CLAUDE.md            # Project rules loaded by Claude Code
├── Makefile             # 13-target build/test contract
├── docker-compose.test.yml
└── docs/specs/          # Pipeline artifact output
```

---

## Customizing for your stack

See [Customization Guide](./customization.md).

---

## Troubleshooting

**`claude: command not found`**
Install Claude Code: `npm install -g @anthropic-ai/claude-code`

**`ai-squad: command not found`**
Check your PATH: `echo $PATH | grep ai-squad`
Re-run: `source ~/.zshrc`

**Agent doesn't start a phase**
Ensure `CLAUDE.md` is present in your project root — it contains the project rules the orchestrator reads.


---

## Installation

```mermaid
flowchart TD
    A["Clone the template globally"] --> B["Add scripts/ to PATH"]
    B --> C["Scaffold a new project"]
    C --> D["Authenticate gh CLI"]
    D --> E["Start your first feature"]
```

### 1. Install the CLI globally

```bash
git clone https://github.com/kinncj/AI-Development-Squad-Template.git ~/.ai-squad
echo 'export PATH="$HOME/.ai-squad/scripts:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

Verify:

```bash
ai-squad help
```

### 2. Scaffold a new project

```bash
mkdir my-project && cd my-project
ai-squad init
```

`ai-squad init` will:
- Copy all agent definitions, commands, and skills into the project
- Initialize a git repository
- Install npm dependencies (Playwright)
- Offer to bootstrap GitHub labels

### 3. Connect a remote repository

```bash
gh auth login
gh repo create my-project --public --push --source=.
```

### 4. Bootstrap GitHub labels (optional but recommended)

```bash
ai-squad labels
```

This creates the full label set the orchestrator uses to track pipeline phases.

---

## Running your first feature

### Option A — Single-agent pipeline (simplest)

Open Claude Code in your project directory and run:

```
/feature "describe your feature here"
```

Claude Code will invoke the orchestrator, which drives all 8 phases sequentially in a single session.

### Option B — Swarm mode (parallel agents, faster)

Swarm mode opens each agent in its own Zellij tab so all Phase 5 tasks run simultaneously.

```bash
ai-squad swarm feature "describe your feature here"
```

See [Swarm Mode](./swarm.md) for full details.

---

## Available commands

Run these inside Claude Code with `/command-name`:

| Command | What it does |
|---|---|
| `/feature "description"` | Full 8-phase pipeline from discovery to PR |
| `/build-feature "description"` | Alias for `/feature` |
| `/bugfix "description"` | Reproduce → fix → validate → CHANGELOG |
| `/validate` | Run the full test suite (no discovery/architecture) |
| `/tdd "requirement"` | Single RED → GREEN → REFACTOR cycle |

---

## Workflow overview

```mermaid
sequenceDiagram
    participant H as Human
    participant ORC as Orchestrator
    participant PO as @product-owner
    participant ARCH as @architect
    participant QA as @qa
    participant DEV as Specialist agents
    participant DOCS as @docs

    H->>ORC: /feature "description"
    ORC->>PO: Phase 1 — write stories & acceptance criteria
    PO-->>H: stories.md (human gate)
    H-->>ORC: approved
    ORC->>ARCH: Phase 2 — ADR, contracts, threat model
    ARCH-->>H: architecture.md (human gate)
    H-->>ORC: approved
    ORC->>ORC: Phase 3 — decompose into plan.md
    ORC->>DEV: Phase 4 — spin up infra
    loop Each task in plan.md
        ORC->>QA: write failing test (RED)
        ORC->>DEV: make test pass (GREEN)
        ORC->>DEV: refactor
    end
    ORC->>QA: Phase 6 — full validation suite
    ORC->>DOCS: Phase 7 — docs, CHANGELOG
    ORC->>H: Phase 8 — PR created ✅
```

---

## Project structure after `ai-squad init`

```
my-project/
├── .claude/
│   ├── agents/          # 27 agent definitions
│   ├── commands/        # /feature, /bugfix, /validate, /tdd
│   └── skills/          # 17 reusable skill files
├── .opencode/           # Mirror for OpenCode platform
├── CLAUDE.md            # Project rules loaded by Claude Code
├── Makefile             # 13-target build/test contract
├── docker-compose.test.yml
└── docs/specs/          # Pipeline artifact output
```

---

## Customizing for your stack

See [Customization Guide](./customization.md).

---

## Troubleshooting

**`claude: command not found`**
Install Claude Code: `npm install -g @anthropic-ai/claude-code`

**`ai-squad: command not found`**
Check your PATH: `echo $PATH | grep ai-squad`
Re-run: `source ~/.zshrc`

**Agent doesn't start a phase**
Ensure `CLAUDE.md` is present in your project root — it contains the project rules the orchestrator reads.

**Swarm mode: `There is no active session!`**
You are already inside a Zellij session. Press `Ctrl-o d` to detach first, then re-run the command. See [Swarm Mode](./swarm.md#troubleshooting).
