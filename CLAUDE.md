# CLAUDE.md — MAPLE Repository

This file is for agents and contributors working on the MAPLE codebase itself (the `maple` binary, install scripts, template files). It is not the CLAUDE.md that gets copied into user projects — that lives at `template/CLAUDE.md`.

---

## Repository Layout

```
.
├── tui/                    # Go source for the maple binary (Bubble Tea TUI)
│   ├── main.go             # CLI entry point, subcommands, runDashboardLoop
│   ├── dashboard.go        # Model, Update(), View(), all key handlers
│   ├── dashboard_views.go  # Every overlay/pane render function
│   ├── test_discovery.go   # Multi-framework test scanner (QA pane)
│   ├── terminal_spawn.go   # spawnInNewTerminal — 8-env detection chain
│   ├── loaders.go          # File readers: stories, sessions, pipeline state, RTK registry
│   ├── detect.go           # Tool detection (claude, opencode, gh, rtk, npx …)
│   ├── boot.go             # Boot check screen shown before dashboard
│   ├── init.go             # maple init — copies template files
│   ├── pipeline_state.go   # pipelineState struct + isStale()
│   ├── embed.go            # go:embed for tui/template/
│   └── template/           # symlink → ../template (real copy required for go:embed)
├── template/               # Everything copied on `maple init`
│   ├── .claude/agents/     # Agent definitions
│   ├── .claude/skills/     # Skill definitions (including superpower-runner)
│   ├── .claude/superpowers/# Named workflow YAML files
│   ├── .opencode/          # Mirror of .claude/ for OpenCode harness
│   ├── .github/            # Copilot instructions, workflows
│   ├── docs/               # Story templates, pipeline, design specs
│   ├── Makefile            # Project Makefile copied on init
│   └── lefthook.yml        # Git hook wiring
├── scripts/
│   ├── install.sh          # Bash installer (macOS + Linux)
│   └── install.ps1         # PowerShell installer (Windows)
├── tests/                  # Test suite for the repo itself
└── CHANGELOG.md            # Updated with every release
```

---

## Build

The `tui/template` directory is a **symlink** to `../template` for development. The Go `go:embed` directive cannot embed symlinks, so the build dance replaces it with a real copy before building:

```bash
# From repo root (preferred — Makefile handles the dance):
make build-tui

# Manually:
rm tui/template
cp -rL template tui/template
cd tui && go build -ldflags "-X main.version=$(git describe --tags --always)" -o ../maple .
cd .. && rm -rf tui/template && ln -s ../template tui/template
```

**Quick compile check** (no dance needed — just checking for errors):

```bash
cd tui && go build -o /tmp/maple_test .
```

If `embed.go` complains about `irregular file template`, the symlink is in place and the build dance is required.

---

## Test After Every Change

Before committing any change to `tui/`:

```bash
cd tui && go build -o /tmp/maple_test . && echo "OK"
```

No exceptions. If it doesn't compile, it doesn't commit.

For install script changes, syntax-check before committing:

```bash
bash -n scripts/install.sh && echo "OK"
```

---

## Release Process

1. Commit and push to `main`
2. Wait for CI to go green (`gh run list --limit 3`)
3. Create the release tag — this triggers `release.yml` which cross-compiles all 5 platforms:

```bash
gh release create vX.Y.Z --title "vX.Y.Z — short description" --notes "..."
```

The release workflow builds:
- `maple-darwin-amd64.tar.gz`
- `maple-darwin-arm64.tar.gz`
- `maple-linux-amd64.tar.gz`
- `maple-linux-arm64.tar.gz`
- `maple-windows-amd64.zip`

**Never push a tag directly** — always use `gh release create`. The release action reads `on: release: types: [published]`.

---

## Versioning

Semver (`vMAJOR.MINOR.PATCH`). Current stream: `v4.x.x`.

| Change type | Bump |
|-------------|------|
| Bug fix, single broken behaviour | patch (`v4.9.1 → v4.9.2`) |
| New feature or TUI overlay | minor (`v4.9.x → v4.10.0`) |
| Breaking change to template schema or protocol | major (rare) |

Multiple related bug fixes in one release are still patch. A release that adds one new keybinding/overlay counts as minor.

---

## Commit Messages

Imperative, lowercase, under 72 characters. No AI phrasing.

**Banned:** enhance, leverage, ensure, implement, utilize, facilitate, improve maintainability, align with best practices, `Co-Authored-By: Claude`

**Good examples:**
```
fix rtk install: find binary in archive tree, isolate from set -e
fix gherkin runner: match 'test-features:' target not just substring
maple never exits when launching a harness
add zellij detection to terminal spawning, document multiplexer setup
fix test discovery: ** globs, Go dedup, Python unittest cmd, id[:8] panic
```

Stage specific files. Never `git add -A`.

---

## GitHub Issues and Project Board

Every bug fix and feature gets a GitHub issue before or immediately after the commit. Use the v4.10.0 milestone for current work.

```bash
gh issue create --title "..." --body "..." --label "bug"       # or "enhancement"
gh api repos/kinncj/MAPLE/issues/N --method PATCH -f milestone=1
```

Roadmap: https://github.com/users/kinncj/projects/67/views/1

After shipping a fix, close the issue:

```bash
gh issue close N --comment "Fixed in vX.Y.Z"
```

---

## Architecture: Key Invariants

### Harness launching never calls `tea.Quit`

`o` (open session), `L` (launcher), and the superpower overlay all use `trySpawnCmd()` — an async `tea.Cmd`. If `spawnInNewTerminal` succeeds, a status bar message appears. If it fails, the `showManualLaunch` modal shows the pasteable command. The TUI never exits for a harness launch.

```go
// correct
return m, trySpawnCmd(cmd)

// wrong — breaks "maple never exits"
m.openTarget = cmd
m.exitAction = dashActionOpenAgent
return m, tea.Quit
```

### `maple.json` is merge-not-overwrite

`writeRecoveryMarker` and any TUI write to `maple.json` must read the existing file first, unmarshal into `map[string]interface{}`, update only its own keys (`state`, `ts`), and re-marshal. The skill owns all other fields.

### `reload()` runs every 5 seconds

`reload()` is called on every `dashTickMsg`. It refreshes stories, sessions, QA entries, design tree, logs, pipeline state, approval pending, and pinned sessions. Adding a new file-based state field means adding it here too.

### `spawnInNewTerminal` detection order

`$ZELLIJ` → `$TMUX` → `$STY` (screen) → `$WEZTERM_PANE` → `$KITTY_PID` → macOS osascript → Linux emulators → Windows Terminal → `errNoNewTerminal`

When `errNoNewTerminal` is returned, `trySpawnCmd` sends `spawnFailedMsg` and the manual-launch modal appears. Never swallow the error silently.

### Test discovery uses `filepath.WalkDir`, not `filepath.Glob`

`filepath.Glob` does not support `**` in Go's stdlib — it silently matches nothing. All test detectors that need to recurse use `filepath.WalkDir`.

---

## RTK Token Optimizer

RTK is installed alongside `maple` by both `install.sh` and `maple init`. It wires a `PreToolUse` hook that compresses Bash tool output before it reaches the LLM.

`maple init` calls `rtk init` via `tools.RTK` (the resolved binary path from `detect.go`). If RTK is not found, init prints a soft warning and continues — it is not required.

The `R` key in the dashboard opens a per-harness RTK wiring overlay. State is saved to `.claude/state/rtk-harnesses.json`.

---

## Shared State Protocol (TUI ↔ Agents)

All communication between the TUI and running agents goes through files in `.claude/state/`:

| File | Writer | Reader | Purpose |
|------|--------|--------|---------|
| `maple.json` | Skill (pipeline fields) + TUI (`state`/`ts`) | Both | Superpower pipeline progress |
| `approval-pending.txt` | Skill | TUI | Human gate — TUI deletes to approve |
| `sessions.json` | TUI (`p`/`o` keys) | Skill (resume logic) | Pinned session IDs per harness |
| `rtk-harnesses.json` | TUI (`R` overlay) | — | Which harnesses have rtk wired |

---

## Open Issues (v4.10.0 — due 2026-05-31)

| # | Title | Target |
|---|-------|--------|
| [#17](https://github.com/kinncj/MAPLE/issues/17) | Verify `o` key opens session end-to-end on v4.9.x | Apr 30 |
| [#18](https://github.com/kinncj/MAPLE/issues/18) | Remove dead `dashActionOpenAgent`/`dashActionLaunch` cases | Apr 25 |
| [#16](https://github.com/kinncj/MAPLE/issues/16) | Automate `tui/template` build dance in Makefile | May 7 |
| [#14](https://github.com/kinncj/MAPLE/issues/14) | `manualLaunchCopied` auto-reset after 2s | May 7 |
| [#15](https://github.com/kinncj/MAPLE/issues/15) | PostToolUse hooks for near-instant TUI refresh | May 31 |

---

## Style

- No comments unless the *why* is non-obvious. Never comment what the code does.
- No multi-line docstrings.
- No backwards-compatibility shims for removed behaviour.
- Unused exported symbols get deleted, not commented out.
- Keep overlay handlers grouped: check `if m.showX` before the global key switch, same order as the `View()` function.
