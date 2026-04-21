# maple — MAPLE CLI

`maple` is the command-line entry point for MAPLE. It initialises a project with the full agent/skill/hook setup and provides a requirements-to-Gherkin helper for fast story creation.

The template is **embedded in the binary** — no separate template directory needed after installing a released build.

## Requirements

- Go 1.22+ (to build from source)
- A terminal — no truecolor required; works on Windows, Linux, macOS

No runtime dependencies beyond the Go standard library and the Bubble Tea family.

## Recommended setup — terminal multiplexer

When you launch a harness (Claude Code, OpenCode, Copilot) from maple, it tries to open it in a **new tab or window** so the dashboard stays visible. This works automatically if you run maple inside a multiplexer:

| Multiplexer | How to start | How maple opens harnesses |
|---|---|---|
| **tmux** | `tmux new-session -s work` then `maple` | `tmux new-window` |
| **zellij** | `zellij` then `maple` | `zellij action new-tab` |

Both are cross-platform (macOS, Linux; tmux also works in WSL on Windows).

If no multiplexer is detected, maple falls back through: WezTerm → Kitty → macOS (iTerm2, Terminal.app) → Linux terminal emulators → Windows Terminal → suspend-and-resume in the same terminal.

You don't need a multiplexer — maple works fine without one — but it's the lowest-friction setup.

## Build

From the **repo root** (preferred):

```bash
make build-tui       # syncs template/ → tui/template/, then builds → ./maple
```

Or manually:

```bash
make sync-template   # copies template/ → tui/template/
cd tui && go build -ldflags "-X main.version=$(git describe --tags --always)" -o ../maple .
```

Cross-compile all platforms:

```bash
make build-tui-all
# produces: maple-darwin-amd64, maple-darwin-arm64,
#           maple-linux-amd64,  maple-linux-arm64,
#           maple-windows-amd64.exe
```

## Install

**Option A — move to system bin:**

```bash
sudo mv maple /usr/local/bin/maple
```

**Option B — personal install (recommended):**

```bash
mkdir -p ~/.tools/maple/bin
mv maple ~/.tools/maple/bin/maple
echo 'export PATH="$HOME/.tools/maple/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

**Option C — one-liner from GitHub Releases (no Go required):**

```bash
curl -fsSL https://raw.githubusercontent.com/kinncj/MAPLE/main/scripts/install.sh | bash
```

Windows:

```powershell
irm https://raw.githubusercontent.com/kinncj/MAPLE/main/scripts/install.ps1 | iex
```

## Commands

| Command | Description |
|---|---|
| `maple init` | Set up MAPLE in the current directory |
| `maple init --force` | Overwrite existing files |
| `maple req` | Write requirements and generate a Gherkin story |
| `maple labels` | Bootstrap the canonical GitHub label set |
| `maple project` | Create a GitHub Project v2 and update project.config.yaml |
| `maple resume-session` | Resume the pinned session (reads `.claude/state/sessions.json`) |
| `maple resume-session claude` | Resume specifically the pinned claude session |
| `maple --version` | Print version |
| `maple --help` | Show usage |

## maple init

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

`maple init` finds the template in this order:

1. `MAPLE_TEMPLATE` env var (resolved to absolute path)
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
