# Quickstart — OpenCode

Get from zero to a running feature pipeline using **OpenCode** and AI Squad.

---

## Prerequisites

| Tool | Install |
|---|---|
| [OpenCode](https://opencode.ai) | See [opencode.ai](https://opencode.ai) |
| [GitHub CLI](https://cli.github.com) | `brew install gh` |
| [Git](https://git-scm.com) | pre-installed on macOS/Linux |
| [Go 1.22+](https://go.dev) | `brew install go` *(only to build from source)* |
| [Node.js](https://nodejs.org) | `brew install node` *(Playwright E2E tests)* |
| [Docker](https://docker.com) | [docker.com/get-started](https://docker.com/get-started) |

> **Any model works.** Quality and consistency come from the prompts, not hardcoded model IDs. Configure your preferred provider in OpenCode's Settings → Providers.

---

## 1. Install `squad`

**From source (preferred):**

```bash
git clone https://github.com/kinncj/AI-Squad.git ai-squad
cd ai-squad
make build-tui           # produces ./squad
```

Add to your PATH — pick one:

```bash
# Option A: move to a system bin
sudo mv squad /usr/local/bin/squad

# Option B: add repo directory to PATH (useful during development)
export PATH="$PWD:$PATH"   # add to ~/.zshrc / ~/.bashrc to persist

# Option C: install to ~/.tools/ai-squad/bin (recommended)
mkdir -p ~/.tools/ai-squad/bin
mv squad ~/.tools/ai-squad/bin/squad
echo 'export PATH="$HOME/.tools/ai-squad/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

**From a release (no Go required):**

```bash
curl -fsSL https://raw.githubusercontent.com/kinncj/AI-Squad/main/scripts/install.sh | bash
```

Verify:

```bash
squad --version
```

---

## 2. Scaffold a project

```bash
mkdir my-project && cd my-project
git init
squad init
squad labels       # bootstrap GitHub label set
```

---

## 3. Configure your provider

Open OpenCode in your project and go to **Settings → Providers**. Add your API key or authenticate with GitHub Copilot. The agents work with any model you configure — no model names are hardcoded.

---

## 4. Write your first requirement

```bash
squad req
```

Type your requirement in plain text, press `Ctrl+D`. AI Squad converts it to a Gherkin story under `docs/stories/`.

---

## 5. Run the pipeline

Open OpenCode in your project directory:

```bash
opencode
```

Then run:

```
/feature "describe your feature here"
```

The orchestrator drives all 8 phases. Sub-agents run as navigable child sessions via OpenCode's native `task` tool — use the session navigator to switch between them.

---

## Rubber Duck

The `@rubber-duck` agent is invoked by the orchestrator automatically at three checkpoints:

1. **After plan** — before implementation starts
2. **After complex multi-file implementations** — before tests run
3. **After tests written** — before executing them

You can also trigger it manually: ask OpenCode to "critique your work" or "get a second opinion."

---

## Available commands

| Command | What it does |
|---|---|
| `/feature "description"` | Full 8-phase pipeline from discovery to PR |
| `/bugfix "description"` | Reproduce → fix → validate → CHANGELOG |
| `/validate` | Run full test suite (skips discovery/architecture) |
| `/tdd "requirement"` | Single RED → GREEN → REFACTOR cycle |

---

## Project structure after `squad init`

```
my-project/
├── .opencode/
│   ├── agents/          # 35 agent definitions (OpenCode frontmatter)
│   ├── commands/        # /feature, /bugfix, /validate, /tdd
│   └── skills/          # 32 reusable skill files
├── .claude/             # Mirror for Claude Code
├── .github/
│   └── copilot-instructions.md
├── opencode.json        # OpenCode project config
├── Makefile
├── project.config.yaml
└── docs/stories/        # Gherkin stories (source of truth)
```

---

## OpenCode agent format

OpenCode agents use richer frontmatter than Claude Code. Example:

```markdown
---
name: typescript
temperature: 0.2
mode: code
tools:
  - read
  - edit
  - write
  - bash
permission:
  allow:
    - bash: ["npx", "node", "npm", "tsc", "jest", "vitest"]
---

You are the TypeScript specialist...
```

The `permission.allow` list restricts which shell commands each agent can run — preventing agents from touching each other's domains.

---

## Troubleshooting

**`opencode: command not found`**
Install from [opencode.ai](https://opencode.ai).

**Orchestrator cannot invoke sub-agents**
Ensure the agent name appears in `orchestrator.md` under `permission.task`. OpenCode enforces this allowlist strictly.

**API key errors**
Set your key in OpenCode's provider settings UI — not via environment variable. OpenCode manages credentials through its own config store.

**Agent uses a model you don't have access to**
Update the `model:` field in the relevant `.opencode/agents/*.md` file to a model ID available in your provider settings.
