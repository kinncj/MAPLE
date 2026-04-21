# Session — 2026-04-21

## What was done

- **ShipSafe integration** — `S` key in the TUI opens a colored popup running `npx ship-safe audit .`; green/yellow/red severity coloring
- **Session detail popups** — Agents pane now uses centered rounded-border lipgloss popups (not full-screen overlays)
- **Open-in-agent** (`o` key) — Agents pane quits MAPLE and execs `claude --resume <uuid>` or `opencode` depending on session source
- **PR approve error fix** — `footer()` now checks `m.status != ""` first, so error messages show through overlay hint text
- **Go source moved to `tui/`** — `main.go`, `dashboard.go`, `themes.go` etc. all moved from repo root into `tui/`; `template/` stays at root; `tui/template` is a relative symlink
- **Build dance** — Makefile and CI both do `rm -f tui/template && cp -rL template tui/template && go build && restore symlink` to satisfy `go:embed`
- **demo.gif** regenerated from `tui_usage.cast` at 0.6× speed with `agg`
- **`kinncj/MAPLE` CI fixed** — `release.yml` was building from `.` (no Go files); `validate-integrations.yml` was hitting "copy into itself" on the symlink; both fixed and pushed
- **`/ship-safe` skill** — added to `template/.claude/skills/` and `template/.opencode/skills/` with full SKILL.md
- **All harnesses updated** — `architect.md` and `orchestrator.md` in both `.claude/agents/` and `.opencode/agents/` reference `ship-safe`
- **OpenCode path fix** — `.opencode/agents/` were wrongly reading `.claude/skills/`; corrected to `.opencode/skills/`
- **Copilot instructions** — `template/.github/copilot-instructions.md` now has full commands table, full specialist agents table, and full skills table
- **Template CI** — `template/.github/workflows/sdlc-gates.yml` gained a `ship-safe` job (`npx --yes ship-safe audit .`)
- **README + docs** updated — `S` and `o` keybindings, `/ship-safe` command, `ship-safe` in skills table
- **Test suite fixed** — Go file checks updated to look in `tui/`; superpowers section skipped when absent; maple binary build uses the cp-rL dance

## Decisions made

- **Superpowers are "not yet built"** — test checks are wrapped in a directory-existence guard rather than hard-failing; the feature is planned but not implemented
- **template/ stays at repo root permanently** — `tui/template` is always a relative symlink; build steps swap it for a real copy only during `go build`
- **OpenCode agents use `.opencode/skills/` paths** — not `.claude/skills/`; was a latent bug in all opencode agent files

## Fixes applied

| Bug | Root cause | Fix |
|---|---|---|
| `kinncj/MAPLE` release CI — "no Go files" | `go build .` from repo root after files moved to `tui/` | Changed to `go build ./tui` |
| `kinncj/MAPLE` validate-integrations — "copy into itself" | `cp -r template tui/template` when `tui/template` is a symlink resolving to `template/` | Added `rm -rf tui/template` before cp |
| Test suite — `main.go missing` | Tests checked root, files now in `tui/` | Updated checks to `tui/$f` |
| Test suite — superpowers failures | Files don't exist yet | Wrapped in `if [[ -d superpowers ]]` guard |
| Test suite — maple binary build failed | `go build` from `tui/` without the symlink-swap | Added `cp -rL` dance to `_build_maple()` |
| OpenCode agents reading wrong skill paths | Copy-paste from claude agents, path not updated | `perl -pi` replaced all `.claude/skills/` → `.opencode/skills/` |

## Unfinished / follow-up

- **Superpowers** — the test infrastructure for `schema.yaml`, `new-ui-feature.yaml`, `superpower-runner` skill is in place (skipped, not failing); the actual feature is not built
- **`.maple.json` gitignore** — the runtime state file is untracked; `maple` binary is now in `.gitignore` but `.maple.json` is not; worth adding

## Commits

```
15237d6 fix test suite: tui/ paths, skip superpowers, fix binary build step
4117d27 Remove bin
847a02a add maple binary to .gitignore
80340ed fix opencode skill paths and flesh out copilot instructions
9e2d0b0 add ship-safe skill to all harnesses and CI gate
377d5bd Updated README
887af68 fix: build from ./tui; rm symlink before cp for embed (kinncj/MAPLE)
ef42c0c move Go source to tui/, template/ stays at root via symlink
de10313 refresh demo.gif from tui_usage.cast at 0.6x speed
f957000 add ShipSafe audit, session popups, open-in-agent, fix PR approve error
```
