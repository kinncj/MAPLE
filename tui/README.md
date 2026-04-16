# squad — AI-Squad TUI

`squad` is the interactive terminal dashboard for AI-Squad. It surfaces stories, active agents, PRs, QA status, design artifacts, and logs in a single four-pane view. It is the default interactive entry point; the existing `ai-squad` CLI stays for CI and non-interactive use.

## Requirements

- Go 1.22+
- A terminal with truecolor support (degrades to 256-color with a notice)

No other runtime dependencies. The binary is statically linked.

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

## Install

```bash
# macOS / Linux
cd tui && go build -o squad . && sudo mv squad /usr/local/bin/squad

# Windows — add the .exe to a directory on %PATH%
```

## Usage

```
squad                   Launch interactive dashboard
squad :resume <name>    Resume a paused superpower
squad --version         Print version
squad --help            Show keybindings
```

## Keybindings

| Key | Action |
|---|---|
| `Tab` / `Shift+Tab` | Cycle panes |
| `j` / `k` | Move down / up |
| `Enter` | Detail view |
| `s` | Stories pane |
| `a` | Agents pane |
| `p` | PRs pane |
| `q` | QA pane |
| `d` | Design pane |
| `l` | Logs pane |
| `F` | Fire superpower |
| `n` | New story / spike / ADR |
| `/` | Search |
| `:` | Command mode |
| `r` | Refresh pane |
| `?` | Help overlay |
| `Ctrl+c` | Quit |

## Commands (`:` mode)

| Command | Action |
|---|---|
| `:kickoff <story-id>` | Start 8-phase pipeline for story |
| `:sync` | Sync all stories ↔ GitHub Issues |
| `:a11y <story-id>` | Run a11y audit for a story |
| `:theme <name>` | Switch theme |
| `:resume <superpower>` | Resume paused superpower |
| `:debug` | Toggle debug log tee to `.claude/logs/tui.log` |

## Themes

Built-in: `tokyo-night` (default), `catppuccin-mocha`, `gruvbox`, `nord`, `everforest`.

Switch: `:theme gruvbox`

## State

TUI state is stored in `.claude/state/squad.json` (gitignored). It contains only ephemeral runtime state (active superpower, current stage). All persistent data (stories, issues, tokens) lives in git or GitHub.

## Dependencies

```
github.com/charmbracelet/bubbletea   — TUI framework
github.com/charmbracelet/bubbles     — reusable TUI components
github.com/charmbracelet/lipgloss    — terminal styling
```

That's it. Three packages.
