# Session — 2026-04-22

## What was done

- **superpowers → TAFFY rename** — complete across all harnesses (Claude Code, OpenCode, Copilot CLI):
  - `.claude/superpowers/` → `.claude/taffy/`
  - `.opencode/superpowers/` → `.opencode/taffy/`
  - `superpower-runner` skill → `pipeline-runner` (universal dispatcher: YAML workflow / skill / agent)
  - All SKILL.md files rewritten with new paths, name, and universal dispatch protocol
  - `schema.yaml` updated in both harnesses
  - `spec-kit.md` agents updated in both harnesses
  - All docs, examples, and walkthroughs updated

- **TAFFY acronym introduced** — Task-Isolated · Asynchronous · Fault-Tolerant · File-Synced · YAML-Driven

- **TUI changes for taffy rename**:
  - `pipelineState.Superpower` field → `Taffy` (`json:"taffy"`)
  - `isSuperpower()` → `isTaffy()`
  - `writeQuickLaunchState` writes `"taffy"` key instead of `"superpower"`
  - `buildQuickPromptCmd` tracking block updated
  - `[P]` pipeline overlay UI text updated ("No active taffy pipeline", etc.)
  - `maple init` drops `obra/superpowers` install; taffy ships with the template
  - `init.go` logs "taffy workflows ready"

- **`[x]` quick launch — TAFFY-first picker with `[t]` toggle**:
  - Default mode: TAFFY workflows from `.claude/taffy/*.yaml`
  - `[t]` toggles to skills/agents list and back; search resets on toggle
  - Taffy badge shows stage count: `[taffy 6 stages]`
  - Taffy items launch as `/pipeline-runner <name>`
  - `loadTaffyItems()` added to `quick_prompt.go` — parses name, description, tags, stageCount from YAML
  - `quickMode` and `taffyItems` fields added to model

- **RTK hook-audit wiring**:
  - `maple init`: runs `rtk hook-audit` after `rtk init` to verify hooks are wired
  - `maple rtk-audit`: new CLI subcommand = `rtk hook-audit`
  - `maple resume-session`: sets `RTK_HOOK_AUDIT=1` in child process env when rtk is on PATH
  - `buildLaunchCmd`: prepends `env RTK_HOOK_AUDIT=1` to all TUI-spawned harness commands when rtk is present

- **README revamp** — full restructure:
  - Clean top-level sections with logical flow
  - MAPLE acronym as a compact table
  - Explicit "Harness Support" section (Claude Code / OpenCode / Copilot CLI)
  - TAFFY as a T/A/F/F/Y table with property descriptions
  - TAFFY workflows section: built-ins, running, gates, custom YAML example

- **Test suite updated** — `tests/cli/test_ai_squad.sh` checks `.claude/taffy/` and `pipeline-runner` instead of superpowers

## Decisions made

- **TAFFY over "superpowers"** — avoids confusion with `obra/superpowers` (NPM package); our concept is distinct and deserves its own name
- **`pipeline-runner` as universal dispatcher** — resolves target in order: taffy YAML → skill → agent; single entry point for everything
- **`[x]` defaults to taffy** — taffy workflows are the primary intent; skills/agents accessible via `[t]` toggle
- **`RTK_HOOK_AUDIT=1` always set when rtk present** — makes hook activity automatically logged on every session launch; no manual export needed
- **Taffy ships with template, not installed** — dropped `obra/superpowers` install from `maple init`; taffy YAML files are embedded in the binary

## Fixes applied

- Background agent had partially renamed files but left `.opencode/superpowers/` as deleted-not-renamed; cleaned up manually with `git add template/.opencode/taffy/` and git mv
- `template/template/` (gitignored stale copy inside template/) updated to match — all superpower references removed
- Orphaned `---` separator in README TAFFY section causing messy rendering — fixed in revamp
- `isSuperpower()` callers in `dashboard_views.go` updated to `isTaffy()` — would have caused compile error

## Unfinished / follow-up

- **`[x]` quick prompt `[t]` toggle — not yet in keybinding help overlay** (`?` key) — should document the toggle
- **`maple rtk-audit` not in README** — `docs/quickstart-claude-code.md` RTK section could mention `maple rtk-audit`
- **GitHub issues #14–#18** remain open (v4.11.0 milestone) — none were addressed this session
- **TUI `[x]` footer hint update** — `x` description in `[?]` help still says "Quick Prompt — pick a skill or agent"; should say "TAFFY picker / skills / agents"
- **OpenCode taffy** — `loadTaffyItems()` reads `.claude/taffy/`; if running under OpenCode, should probably also read `.opencode/taffy/` (same content, but harness-aware loading would be cleaner)

## Commits

- `85d9ebf` add rtk hook-audit wiring across maple
- `4c874cc` [x] quick launch: TAFFY-first picker with [t] toggle
- `34e6125` clean up background agent work: drop opencode/superpowers, sync SKILL.md and schema across harnesses
- `84bbd12` revamp README: clean structure, TAFFY as table, harness support section
- `8051b8c` rename superpowers → taffy; introduce TAFFY acronym

## Releases

- v4.10.8 — rename superpowers to TAFFY; revamp README
- v4.10.9 — [x] TAFFY-first picker with [t] toggle
- v4.10.10 — rtk hook-audit wiring
