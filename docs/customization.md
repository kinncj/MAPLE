# Customization Guide

This guide covers how to tailor the AI Development Squad template to your specific stack — adding agents, adjusting models, and restricting what each agent can do.

---

## Adding a custom agent

### Step 1 — Create the Claude Code agent

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
- Invoke agents other than yourself
- Skip failing tests
```

No `model:` field needed — the agent uses your Claude Code default model.

### Step 2 — Create the OpenCode agent

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
    - bash: ["cargo", "rustc", "rustfmt", "clippy", "rustup"]
---

You are the Rust specialist...
```

No `model:` field needed — the agent inherits OpenCode's configured default.

### Step 3 — Register with the orchestrator

In `.opencode/agents/orchestrator.md`, add `rust` to the `permission.task` list:

```yaml
permission:
  task:
    - product-owner
    - architect
    - qa
    - rust        # ← add here
    - typescript
    # ...
```

Do the same in `.claude/agents/orchestrator.md` (add `rust` to the orchestrator's agent roster comment for documentation purposes).

### Step 4 — Update AGENTS.md

Add a row to the table in `AGENTS.md` so the orchestrator knows which tasks to route to the new agent.

---

## Removing an agent

1. Delete `.claude/agents/{name}.md`
2. Delete `.opencode/agents/{name}.md`
3. Remove the agent from the `permission.task` list in both orchestrator files
4. Remove the row from `AGENTS.md`

The orchestrator will no longer delegate to it.

---

## Choosing models

Agents have no hardcoded model — they use whatever model you configure in your AI tool of choice.

**Claude Code:** Set your preferred model in Claude Code's global settings. You can optionally add a `model:` field to individual agent frontmatter to override on a per-agent basis.

**OpenCode:** Configure your default model in `opencode.json` or via Settings → Providers. Add a `model:` field to individual agent frontmatter to override per-agent:

```json
{
  "model": {
    "default": "your-provider/your-model-id"
  }
}
```

---

## Restricting agent permissions

Each OpenCode agent has a `permission.allow` list that controls which shell commands it can run:

```yaml
permission:
  allow:
    - bash: ["cargo", "rustc", "rustfmt"]
```

This prevents the rust agent from running `npm`, `python`, or any other tool outside its domain. Claude Code uses tool permissions configured per-project in `CLAUDE.md`.

---

## Adding a custom skill

Skills are markdown files that encode specialized workflows (CLI patterns, TDD sequences, tool usage). They are loaded on demand.

Create `.claude/skills/{name}.md` (and `.opencode/skills/{name}.md`):

```markdown
# Skill: redis-pub-sub

Use this skill when implementing Redis pub/sub patterns.

## Setup
Always use connection pooling:
```bash
redis-cli -h localhost -p 6379
```

## Pattern: fan-out
...
```

Reference the skill in an agent prompt:

```
@skill redis-pub-sub
```

---

## Adjusting the plan.md format

Tasks in `plan.md` follow this format:

```
- [ ] Task N: @agent-name description of task
```

Any line matching `- [ ] Task [0-9]+:` with an `@agent-name` is a delegatable task. Checked tasks (`- [x]`) are skipped. Maintain this format if you customise the orchestrator's planning output.

---

## Makefile customization

The template ships with stub targets. Replace the recipe bodies for your stack:

```makefile
test:
	pytest tests/unit -v           # Python example

test-integration:
	docker compose -f docker-compose.test.yml up -d
	pytest tests/integration -v
	docker compose -f docker-compose.test.yml down

test-e2e:
	npx playwright test
```

The 13-target contract (`build`, `test`, `test-integration`, `test-e2e`, `test-contract`, `test-all`, `lint`, `security-scan`, `fmt`, `containers-up`, `containers-down`, `seed-test`, `migrate`) must remain intact — agents rely on these names.


This guide covers how to tailor the AI Development Squad template to your specific stack — adding agents, adjusting models, and restricting what each agent can do.

---

## Adding a custom agent

### Step 1 — Create the Claude Code agent

Create `.claude/agents/{name}.md`:

```markdown
---
name: rust
description: Rust systems programming specialist. Handles Cargo workspaces, async Tokio, and FFI.
model: claude-sonnet-4-6
---

You are the Rust specialist. You write idiomatic, safe Rust code.

## What you do
- Implement tasks assigned by the orchestrator
- Write unit tests with `#[test]` and integration tests in `tests/`
- Run `cargo build`, `cargo test`, `cargo clippy`, `cargo fmt`

## What you never do
- Modify files outside Rust source directories
- Invoke agents other than yourself
- Skip failing tests
```

### Step 2 — Create the OpenCode agent

Create `.opencode/agents/{name}.md`:

```markdown
---
name: rust
model: github-copilot/claude-sonnet-4.5
temperature: 0.2
mode: code
tools:
  - read
  - edit
  - write
  - bash
permission:
  allow:
    - bash: ["cargo", "rustc", "rustfmt", "clippy", "rustup"]
---

You are the Rust specialist...
```

### Step 3 — Register with the orchestrator

In `.opencode/agents/orchestrator.md`, add `rust` to the `permission.task` list:

```yaml
permission:
  task:
    - product-owner
    - architect
    - qa
    - rust        # ← add here
    - typescript
    # ...
```

Do the same in `.claude/agents/orchestrator.md` (Claude Code uses `description` for routing, not an explicit allowlist, but add `rust` to the orchestrator's agent roster comment for documentation purposes).

### Step 4 — Update AGENTS.md

Add a row to the table in `AGENTS.md` so the orchestrator knows which tasks to route to the new agent.

---

## Removing an agent

1. Delete `.claude/agents/{name}.md`
2. Delete `.opencode/agents/{name}.md`
3. Remove the agent from the `permission.task` list in both orchestrator files
4. Remove the row from `AGENTS.md`

The orchestrator will no longer delegate to it.

---

## Changing model assignments

### Claude Code

Edit the `model:` field in the agent's `.claude/agents/{name}.md` frontmatter:

```yaml
---
name: typescript
model: claude-opus-4-6    # upgrade to Opus for complex reasoning
---
```

Available Claude Code models:

| Model | ID | Use when |
|---|---|---|
| Claude Opus 4.6 | `claude-opus-4-6` | Complex design decisions, architecture |
| Claude Sonnet 4.6 | `claude-sonnet-4-6` | Fast, capable code generation |
| Claude Haiku 4.5 | `claude-haiku-4-5-20251001` | High-volume, simple tasks |

### OpenCode

Edit the `model:` field in `.opencode/agents/{name}.md` and update `opencode.json`:

```json
{
  "model": {
    "default": "github-copilot/claude-sonnet-4.5"
  }
}
```

Provider model ID strings:

| Provider | Model | ID string |
|---|---|---|
| Anthropic API | Claude Opus 4.6 | `anthropic/claude-opus-4-6` |
| Anthropic API | Claude Sonnet 4.6 | `anthropic/claude-sonnet-4-6` |
| GitHub Copilot Enterprise | Claude Sonnet 4.5 | `github-copilot/claude-sonnet-4.5` |
| GitHub Copilot Enterprise | GPT-4.1 | `copilot/gpt-4.1` |

Re-run `ai-squad init` (answer the provider prompts) to automatically rewrite all 34 agent files for your subscription.

---

## Restricting agent permissions

Each OpenCode agent has a `permission.allow` list that controls which shell commands it can run:

```yaml
permission:
  allow:
    - bash: ["cargo", "rustc", "rustfmt"]
```

This prevents the rust agent from running `npm`, `python`, or any other tool outside its domain. Claude Code uses tool permissions configured per-project in `CLAUDE.md`.

---

## Adding a custom skill

Skills are markdown files that encode specialized workflows (CLI patterns, TDD sequences, tool usage). They are loaded on demand.

Create `.claude/skills/{name}.md` (and `.opencode/skills/{name}.md`):

```markdown
# Skill: redis-pub-sub

Use this skill when implementing Redis pub/sub patterns.

## Setup
Always use connection pooling:
```bash
redis-cli -h localhost -p 6379
```

## Pattern: fan-out
...
```

Reference the skill in an agent prompt:

```
@skill redis-pub-sub
```

---

## Adjusting the plan.md format

The orchestrator and spec-kit agent parse `plan.md` for unchecked tasks using this pattern:

```
- [ ] Task N: @agent-name description of task
```

Any line matching `- [ ] Task [0-9]+:` with an `@agent-name` gets picked up. Checked tasks (`- [x]`) are skipped. Maintain this format if you customise the orchestrator's planning output.

---

## Changing the plan path

By default, spec-kit reads `docs/specs/<slug>/TASKS.md`. The orchestrator's pre-DISCOVER gate checks this path. Override by passing the path explicitly when invoking the spec-kit agent.

---

## Makefile customization

The template ships with stub targets. Replace the recipe bodies for your stack:

```makefile
test:
	pytest tests/unit -v           # Python example

test-integration:
	docker compose -f docker-compose.test.yml up -d
	pytest tests/integration -v
	docker compose -f docker-compose.test.yml down

test-e2e:
	npx playwright test
```

The 13-target contract (`build`, `test`, `test-integration`, `test-e2e`, `test-contract`, `test-all`, `lint`, `security-scan`, `fmt`, `containers-up`, `containers-down`, `seed-test`, `migrate`) must remain intact — agents rely on these names.

---

## Updating to new model versions

When Anthropic or GitHub releases a new model version:

1. Run `ai-squad init --yes` and answer the provider prompts — the wizard rewrites all agent files automatically
2. Or manually update:
   - `opencode.json` → `model.default`
   - Each `.opencode/agents/*.md` → `model:` field
   - Each `.claude/agents/*.md` → `model:` field (for Claude Code)
