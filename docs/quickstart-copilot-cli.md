# Quickstart — GitHub Copilot CLI

Get from zero to a running feature pipeline using **GitHub Copilot CLI** (`copilot`) and MAPLE.

---

## Prerequisites

| Tool | Install |
|---|---|
| [GitHub Copilot CLI](https://github.com/features/copilot/cli) | `gh extension install github/gh-copilot` |
| [GitHub CLI](https://cli.github.com) | `brew install gh` |
| [Git](https://git-scm.com) | pre-installed on macOS/Linux |
| [Go 1.22+](https://go.dev) | `brew install go` *(only to build from source)* |
| [Node.js](https://nodejs.org) | `brew install node` *(Playwright E2E tests)* |
| [Docker](https://docker.com) | [docker.com/get-started](https://docker.com/get-started) |

Authenticate:

```bash
gh auth login
gh copilot --version     # verify Copilot CLI is working
```

---

## 1. Install `maple`

**From source (preferred):**

```bash
git clone https://github.com/kinncj/AI-Squad.git maple
cd maple
make build-tui           # produces ./maple
```

Add to your PATH — pick one:

```bash
# Option A: move to a system bin
sudo mv maple /usr/local/bin/maple

# Option B: add repo directory to PATH (useful during development)
export PATH="$PWD:$PATH"   # add to ~/.zshrc / ~/.bashrc to persist

# Option C: install to ~/.tools/maple/bin (recommended)
mkdir -p ~/.tools/maple/bin
mv maple ~/.tools/maple/bin/maple
echo 'export PATH="$HOME/.tools/maple/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

**From a release (no Go required):**

```bash
curl -fsSL https://raw.githubusercontent.com/kinncj/AI-Squad/main/scripts/install.sh | bash
```

Verify:

```bash
maple --version
```

---

## 2. Scaffold a project

```bash
mkdir my-project && cd my-project
git init
maple init
maple labels       # bootstrap GitHub label set
```

---

## 3. Write your first requirement

```bash
maple req
```

Type your requirement in plain text, press `Ctrl+D` to generate a Gherkin story saved under `docs/stories/`.

---

## 4. Open Copilot CLI and enable Rubber Duck

```bash
copilot
```

Enable experimental mode for the **Rubber Duck** second-opinion reviewer:

```
/experimental
```

Then select a Claude model from the model picker. Rubber Duck will automatically use GPT-5.4 as the cross-family reviewer.

---

## 5. Run the pipeline

```
/feature "describe your feature here"
```

The orchestrator agent (`@orchestrator`) drives all 8 phases. It pauses at human-approval gates and updates GitHub Issues as it goes.

---

## Rubber Duck

Copilot CLI's built-in Rubber Duck activates **automatically** at three checkpoints when `/experimental` is enabled:

| Checkpoint | What it catches |
|---|---|
| After plan (Phase 3) | Missing components, wrong task ordering, flawed TDD sequencing |
| After complex multi-file implementation | Logic bugs, cross-file contract breaks, missing error handling |
| After writing tests | Weak assertions, missing scenarios, JS mock antipatterns |

You can also trigger it manually at any time — just say "critique your work" or "get a second opinion."

> **Without `/experimental`:** The `@rubber-duck` agent defined in `.github/copilot-instructions.md` provides equivalent coverage — the orchestrator invokes it at the same three checkpoints.

---

## Skills

MAPLE skills load automatically from `.github/extensions/` and `.claude/skills/`. All skill files have the required YAML frontmatter (`name:` + `description:`) for Copilot CLI.

Reference a skill in your session:

```
use the rubber-duck skill to review this plan
use the tdd-workflow skill
```

---

## Available commands

| Command | What it does |
|---|---|
| `/feature "description"` | Full 8-phase pipeline from discovery to PR |
| `/bugfix "description"` | Reproduce → fix → validate → CHANGELOG |
| `/validate` | Run full test suite (skips discovery/architecture) |
| `/tdd "requirement"` | Single RED → GREEN → REFACTOR cycle |

---

## Project structure after `maple init`

```
my-project/
├── .github/
│   ├── copilot-instructions.md   # Copilot CLI rules (loaded automatically)
│   └── instructions/             # path-specific rules for stories, etc.
├── .claude/
│   ├── agents/                   # 35 agent definitions
│   └── skills/                   # 32 skill files (all with YAML frontmatter)
├── .opencode/                    # Mirror for OpenCode
├── Makefile
├── project.config.yaml
└── docs/stories/                 # Gherkin stories
```

---

## Troubleshooting

**`copilot: command not found`**
Install: `gh extension install github/gh-copilot` then `gh copilot --version`.

**Skill shows "missing or malformed YAML frontmatter"**
Each skill file must start with `---\nname: skill-name\ndescription: ...\n---`. All MAPLE skills already have this — if you added a custom skill, add the frontmatter.

**Rubber Duck not appearing**
Type `/experimental` in your Copilot CLI session and make sure a Claude model is selected in the model picker.

**Agent doesn't follow pipeline phases**
Ensure `.github/copilot-instructions.md` is present — it's the enforcement file Copilot CLI reads for project-level rules.
