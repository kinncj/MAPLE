# Quickstart — Claude Code

Get from zero to a running feature pipeline using **Claude Code** and AI Squad.

---

## Prerequisites

| Tool | Install |
|---|---|
| [Claude Code](https://claude.ai/claude-code) | `npm install -g @anthropic-ai/claude-code` |
| [GitHub CLI](https://cli.github.com) | `brew install gh` |
| [Git](https://git-scm.com) | pre-installed on macOS/Linux |
| [Go 1.22+](https://go.dev) | `brew install go` *(only to build from source)* |
| [Node.js](https://nodejs.org) | `brew install node` *(Playwright E2E tests)* |
| [Docker](https://docker.com) | [docker.com/get-started](https://docker.com/get-started) |

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

# Option C: install to ~/.tools/ai-squad/bin (recommended for personal installs)
mkdir -p ~/.tools/ai-squad/bin
mv squad ~/.tools/ai-squad/bin/squad
echo 'export PATH="$HOME/.tools/ai-squad/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

**From a release (no Go required):**

```bash
curl -fsSL https://raw.githubusercontent.com/kinncj/AI-Squad/main/scripts/install.sh | bash
# installs to ~/.tools/ai-squad/bin/ and prints PATH instructions
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
```

`squad init` detects which AI tools are installed and copies the matching agent definitions, skills, hooks, Makefile, and config. Existing files are never overwritten.

```bash
squad labels       # bootstrap GitHub label set (requires gh auth login)
squad project      # create a GitHub Project v2 (optional)
```

---

## 3. Write your first requirement

```bash
squad req
```

Type your requirement in plain text, press `Ctrl+D`. AI Squad converts it to a Gherkin story saved under `docs/stories/`.

---

## 4. Run the pipeline

Open Claude Code in your project directory:

```bash
claude
```

Then run:

```
/feature "describe your feature here"
```

The orchestrator drives all 8 phases: DISCOVER → ARCHITECT → PLAN → INFRA → IMPLEMENT → VALIDATE → DOCUMENT → FINAL GATE. It pauses at each human-approval gate and surfaces results as GitHub Issues.

---

## Rubber Duck (second opinion)

Claude Code invokes `@rubber-duck` automatically at three checkpoints:

1. **After plan** — before implementation starts
2. **After complex multi-file implementations** — before tests run
3. **After tests written** — before executing them

You can also trigger it manually: just ask Claude to "critique your work" or "get a second opinion."

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
├── .claude/
│   ├── agents/          # 35 agent definitions (incl. rubber-duck)
│   ├── commands/        # /feature, /bugfix, /validate, /tdd
│   ├── hooks/           # pre/post tool-use enforcement
│   └── skills/          # 32 reusable skill files
├── .opencode/           # Mirror for OpenCode
├── .github/
│   ├── copilot-instructions.md   # Copilot CLI rules
│   └── instructions/             # path-specific rules
├── CLAUDE.md            # Project rules loaded on every Claude Code session
├── Makefile             # build/test/lint contract
├── project.config.yaml  # stack detection, SDLC mode
└── docs/stories/        # Gherkin stories (source of truth)
```

---

## Troubleshooting

**`squad: command not found`**
Check your PATH. If you used Option C above: `source ~/.zshrc` then retry.

**`claude: command not found`**
`npm install -g @anthropic-ai/claude-code`

**Agent skips a phase gate**
Ensure `CLAUDE.md` is present in your project root — it contains the project rules the orchestrator reads on every session start.

**Hook blocked my commit**
The pre-bash hook runs SDLC gates before every `git commit`. Fix the reported issue (failing test, missing frontmatter, secret in staged files) and retry.
