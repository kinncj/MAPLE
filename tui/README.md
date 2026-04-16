# squad — AI-Squad CLI

`squad` is the command-line entry point for AI-Squad. It initialises a project with the full agent/skill/hook setup, and provides a requirements-to-Gherkin helper for fast story creation.

## Requirements

- Go 1.22+  (to build from source)
- A terminal — no truecolor required; works on Windows, Linux, macOS

No runtime dependencies beyond the Go standard library and the Bubble Tea family.

## Build

```bash
cd tui
go build -o squad .
```

Cross-compile:

```bash
# macOS
GOOS=darwin  GOARCH=amd64  go build -o squad-darwin-amd64  .
GOOS=darwin  GOARCH=arm64  go build -o squad-darwin-arm64  .

# Linux
GOOS=linux   GOARCH=amd64  go build -o squad-linux-amd64   .
GOOS=linux   GOARCH=arm64  go build -o squad-linux-arm64   .

# Windows
GOOS=windows GOARCH=amd64  go build -o squad-windows-amd64.exe .
```

Or from the repo root:

```bash
make build-tui   # builds ./squad
```

## Install

```bash
# macOS / Linux
cd tui && go build -o squad . && sudo mv squad /usr/local/bin/squad

# Windows — add the .exe to a directory on %PATH%
```

## Commands

| Command | Description |
|---|---|
| `squad init` | Set up AI-Squad in the current directory |
| `squad init --force` | Overwrite existing files |
| `squad req` | Write requirements and generate a Gherkin story |
| `squad labels` | Bootstrap the canonical GitHub label set |
| `squad project` | Create a GitHub Project v2 and update project.config.yaml |
| `squad --version` | Print version |
| `squad --help` | Show usage |

## squad init

Detects which AI tools are installed (claude, opencode, gh copilot) and copies the matching template files:

- `.claude/` — agents, skills, hooks (if Claude Code detected)
- `.opencode/` — agents, skills (if OpenCode detected)
- `.github/` — copilot-instructions, workflows
- `docs/` — story templates, specs, design structure
- `Makefile`, `lefthook.yml`, `scripts/sdlc/`
- `project.config.yaml` — written only if not already present
- Runs `lefthook install` to wire git hooks

Existing files are never overwritten (use `--force` to override).

## squad req

Opens an interactive Bubble Tea editor. Type plain-text requirements, press **Ctrl+D** to convert. The AI tool detected on the machine converts the text to a Gherkin `.feature` file saved under `docs/stories/`.

## Template resolution

`squad init` finds templates in this order:

1. `AI_SQUAD_TEMPLATE` environment variable
2. `<binary>/../template/` (works when installed alongside the repo)
3. `./template/` (cwd fallback — works when running from repo root)
4. `~/.ai-squad/template/` (global install)

## Themes

Built-in: `tokyo-night` (default), `catppuccin-mocha`, `gruvbox`, `nord`, `everforest`.

## Dependencies

```
github.com/charmbracelet/bubbletea   — TUI framework
github.com/charmbracelet/bubbles     — textarea, spinner
github.com/charmbracelet/lipgloss    — terminal styling
```

