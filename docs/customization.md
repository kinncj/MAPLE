# Customization Guide

How to tailor MAPLE to your specific stack — adding agents, adjusting permissions, extending skills, and customizing the Makefile.

---

## Adding a custom agent

### Step 1 — Claude Code agent

Create `.claude/agents/{name}.md`:

```markdown
---
name: rust
description: Rust systems programming specialist. Handles Cargo workspaces, async Tokio, and FFI.
---

You are the Rust specialist. You write idiomatic, safe Rust code.

## What you do
- Implement tasks assigned by the orchestrator
- Write unit tests with `#[test]` and integration tests in `tests/`
- Run `cargo build`, `cargo test`, `cargo clippy`, `cargo fmt`

## What you never do
- Modify files outside Rust source directories
- Invoke other agents
- Skip failing tests
```

No `model:` field needed — the agent uses your Claude Code default model.

### Step 2 — OpenCode agent

Create `.opencode/agents/{name}.md`:

```markdown
---
name: rust
temperature: 0.2
mode: code
tools:
  - read
  - edit
  - write
  - bash
permission:
  allow:
    - bash: ["cargo", "rustc", "rustfmt", "clippy"]
---

You are the Rust specialist...
```

### Step 3 — Register with the orchestrator

Add your agent to the `permission.task` list in `.opencode/agents/orchestrator.md`:

```yaml
permission:
  task:
    - rust    # add here
    - dotnet
    - javascript
    # ...
```

Then update `AGENTS.md` so the orchestrator knows what expertise is available.

---

## Restricting agent permissions

OpenCode agents support fine-grained bash restrictions. Only allow the commands an agent legitimately needs:

```yaml
permission:
  allow:
    - bash: ["npm", "npx", "node"]   # node.js agent — only npm/node
  deny:
    - bash: ["rm", "git push"]       # never delete or push directly
```

Claude Code agents inherit permissions from your `settings.json`. To lock down a specific agent, add it to the `agentPermissions` section in `.claude/settings.json`.

---

## Customizing the Makefile

The template Makefile ships with stub targets. Replace the recipe bodies with your actual stack commands:

```makefile
## Unit tests
test:
    dotnet test --filter 'Category=Unit'

## Integration tests
test-integration:
    dotnet test --filter 'Category=Integration'

## Build
build:
    dotnet build
```

The section between `# ─── BEGIN MAPLE MANAGED` and `# ─── END MAPLE MANAGED` is managed by `maple update`. Your custom targets outside those markers are never touched on update.

---

## Adding custom skills

Skills are markdown files read by agents before task execution. Create one at `.claude/skills/{name}/SKILL.md`:

```markdown
# my-workflow

## When to use
Use this skill when implementing X.

## Steps
1. Run `some-tool init`
2. Edit the config at `./config.yml`
3. Run `some-tool validate`

## Common patterns
...
```

Reference it in an agent prompt: `Read .claude/skills/my-workflow/SKILL.md before starting.`

---

## Installing skills from the marketplace

The skills.sh marketplace has community-built skills for Claude Code, Cursor, and other editors.

**From the TUI** — press `F` to open the browser:
- **Installed** tab — see all installed skills, `d` to remove
- **Search** tab — type a query, `Enter` to find and install

**From the CLI:**
```bash
npx skills find <query>
npx skills add owner/repo@skill --all -y
npx skills remove <name> --all -y
npx skills ls
```

MAPLE ships built-in taffy workflows in `.claude/taffy/` — no extra install needed.

---

## Switching themes

The `maple` TUI supports five built-in themes. Switch from the dashboard:

```
:theme tokyo-night       (default)
:theme catppuccin-mocha
:theme gruvbox
:theme nord
:theme everforest
```

Or auto-detect from your Omarchy config: if `~/.config/omarchy/current/theme` exists, `maple` selects the matching theme at launch.

---

## Updating MAPLE

Pull the latest template files into your project without overwriting your customizations:

```bash
maple update        # or: u from the dashboard
```

This re-syncs:
- `.claude/` agents, skills, hooks
- `.opencode/` agents
- `docs/` story templates and structure
- `scripts/sdlc/`
- The MAPLE-managed section of your `Makefile`

Your custom Makefile targets, story files, and project-specific settings are left untouched.

To upgrade the `maple` binary itself:

```bash
maple self-update   # fetches latest release from GitHub
```
