# squad — AI Squad CLI

`squad` is the command-line entry point for AI Squad. It initialises a project with the full agent/skill/hook setup and provides a requirements-to-Gherkin helper for fast story creation.

The template is **embedded in the binary** — no separate template directory needed after installing a released build.

## Requirements

- Go 1.22+ (to build from source)
- A terminal — no truecolor required; works on Windows, Linux, macOS

No runtime dependencies beyond the Go standard library and the Bubble Tea family.

## Build

From the **repo root** (preferred):

```bash
make build-tui       # syncs template/ → tui/template/, then builds → ./squad
```

Or manually:

```bash
make sync-template   # copies template/ → tui/template/
cd tui && go build -ldflags "-X main.version=$(git describe --tags --always)" -o ../squad .
```

Cross-compile all platforms:

```bash
make build-tui-all
# produces: squad-darwin-amd64, squad-darwin-arm64,
#           squad-linux-amd64,  squad-linux-arm64,
#           squad-windows-amd64.exe
```

## Install

**Option A — move to system bin:**

```bash
sudo mv squad /usr/local/bin/squad
```

**Option B — personal install (recommended):**

```bash
mkdir -p ~/.tools/ai-squad/bin
mv squad ~/.tools/ai-squad/bin/squad
echo 'export PATH="$HOME/.tools/ai-squad/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

**Option C — one-liner from GitHub Releases (no Go required):**

```bash
curl -fsSL https://raw.githubusercontent.com/kinncj/AI-Squad/main/scripts/install.sh | bash
```

Windows:

```powershell
irm https://raw.githubusercontent.com/kinncj/AI-Squad/main/scripts/install.ps1 | iex
```

## Commands

| Command | Description |
|---|---|
| `squad init` | Set up AI Squad in the current directory |
| `squad init --force` | Overwrite existing files |
| `squad req` | Write requirements and generate a Gherkin story |
| `squad labels` | Bootstrap the canonical GitHub label set |
| `squad project` | Create a GitHub Project v2 and update project.config.yaml |
| `squad --version` | Print version |
| `squad --help` | Show usage |

## squad init

Detects which AI tools are installed (`claude`, `copilot`, `opencode`) and copies the matching template files:

- `.claude/` — agents, skills, hooks (if Claude Code or Copilot CLI detected)
- `.opencode/` — agents, skills (if OpenCode detected)
- `.github/` — copilot-instructions, instructions/, workflows
- `docs/` — story templates, specs, design structure
- `Makefile`, `lefthook.yml`, `scripts/sdlc/`
- `project.config.yaml` — written only if not already present
- Runs `lefthook install` to wire git hooks

Existing files are never overwritten (use `--force` to override).

## Template resolution

`squad init` finds the template in this order:

1. `AI_SQUAD_TEMPLATE` env var (resolved to absolute path)
2. `<binary_dir>/template/` if it exists on disk (dev checkout)
3. `./template/` in cwd (running from repo root)
4. **Embedded** — always works for released binaries (no external files needed)

## Themes

Built-in: `tokyo-night` (default), `catppuccin-mocha`, `gruvbox`, `nord`, `everforest`.  
Switch with `:theme <name>` in the TUI.

## Dependencies

```
github.com/charmbracelet/bubbletea   — TUI framework
github.com/charmbracelet/bubbles     — textarea, spinner
github.com/charmbracelet/lipgloss    — terminal styling
```
