---
name: pipeline-runner
description: "Universal dispatcher: run a named taffy workflow (.claude/taffy/<name>.yaml), a skill (/skill-name), or a sub-agent (@agent-name). Falls back to skills.sh registry when a skill is not found locally. Tracks all runs in .claude/state/maple.json so the maple TUI shows live progress."
tags:
  - pipeline
  - workflow
  - dispatcher
  - maple
  - taffy
---

# SKILL: pipeline-runner

## What It Does

Dispatches any named workflow, skill, or agent from a single entry point. Resolution order:

1. **Taffy workflow** — look for `.claude/taffy/<name>.yaml`; if found, execute each stage in order
2. **Skill (local)** — look for `.claude/skills/<name>/`; if found, invoke the skill
3. **Agent** — look for `.claude/agents/<name>.md`; if found, delegate to `@<name>`
4. **Skill (registry)** — if not found locally, try to install from skills.sh, then retry step 2

Pipeline state is written to `.claude/state/maple.json` at every transition.

## Usage

```
/pipeline-runner <name>
```

Examples:
```
/pipeline-runner new-ui-feature
/pipeline-runner api-endpoint
/pipeline-runner tdd-workflow
/pipeline-runner orchestrator
```

List available taffy workflows:
```bash
ls .claude/taffy/*.yaml | grep -v schema
```

List available skills:
```bash
ls .claude/skills/
```

## Dispatch Protocol

### Step 1: Resolve the target

```bash
# Check taffy first
[ -f ".claude/taffy/<name>.yaml" ] && dispatch=taffy
# Then local skill
[ -d ".claude/skills/<name>" ] && dispatch=skill
# Then agent
[ -f ".claude/agents/<name>.md" ] && dispatch=agent
# Fallback: fetch from skills.sh registry, then retry
if [ -z "$dispatch" ] && command -v npx &>/dev/null; then
  echo "pipeline-runner: '<name>' not found locally — checking skills.sh…"
  npx --yes skills add kinncj/maple@<name> -a claude-code -y 2>/dev/null \
    || npx --yes skills add <name> -a claude-code -y 2>/dev/null \
    || true
  [ -d ".claude/skills/<name>" ] && dispatch=skill
fi
```

If nothing matches after the registry fallback, report:
`pipeline-runner: no taffy workflow, skill, or agent named '<name>' (also checked skills.sh registry)`

### Step 2: Initialise state

Write to `.claude/state/maple.json` (merge — do not overwrite unowned fields):

```json
{
  "taffy": "<name>",
  "stage": "<first-stage or skill-name>",
  "status": "RUNNING",
  "awaiting_approval": null,
  "started_at": "<iso8601>",
  "updated_at": "<iso8601>"
}
```

### Step 3a: Taffy workflow execution

Load `.claude/taffy/<name>.yaml`, parse `stages:`, resolve `depends_on` order.

For each stage:

**`when:` guard:**
- `when: ui:true` — skip if story has `ui: false`
- `when: ui:false` — skip if story has `ui: true`
- `when: always` — always run

**`depends_on`:** all listed stages must be `DONE` before this one starts.

**Dispatch:**
- `agent: <name>` → delegate to `@<name>` with current story context
- `skill: <name>` → invoke the skill
- `pipeline: standard` → run the full 8-phase orchestrator pipeline

After each stage: update `maple.json` with current stage + `RUNNING`.

### Step 3b: Skill invocation

Invoke the skill directly. Update `maple.json` on start and completion.

### Step 3c: Agent delegation

Delegate to `@<name>`. Update `maple.json` on start and completion.

### Step 4: Human-approval gates (taffy only)

When a stage has `gate: human-approval`:

1. Complete stage work (produce artifact).
2. Write PAUSED state:
```json
{ "stage": "<name>", "status": "PAUSED", "awaiting_approval": "<name>", "updated_at": "<iso8601>" }
```
3. Write stage name to `.claude/state/approval-pending.txt`.
4. Output:
```
TAFFY PAUSED — awaiting human approval
Stage:    <stage-name>
Artifact: <artifact path or description>

Approve via the maple TUI ([P] pipeline → [a] approve) or reply "approved" / "continue".
I will not advance to the next stage until approval is confirmed.
```
5. Poll: `until [ ! -f .claude/state/approval-pending.txt ]; do sleep 2; done`
   Also accept explicit "approved" / "continue" reply in chat.
6. On resume: update to `RUNNING`, advance to next stage.

### Step 5: Completion

```json
{ "taffy": "<name>", "stage": "DONE", "status": "DONE", "awaiting_approval": null, "updated_at": "<iso8601>" }
```

Output:
```
TAFFY COMPLETE — <name>
Stages run: N
Duration:   <elapsed>
```

## Failure Handling

After 3 consecutive failures on any stage:

```json
{ "stage": "<name>", "status": "FAILED", "error": "<summary>", "updated_at": "<iso8601>" }
```

Stop. Report failed stage and error to human. Do not proceed.

## Session Context

On startup, read `.claude/state/sessions.json` if it exists:

```json
{ "claude": "<uuid>", "opencode": "<id>", "copilot": "<id>" }
```

Use the matching session ID when resuming work within an existing agent session.

## State File Reference

All state in `.claude/state/`. TUI and skill share these files.

### `.claude/state/maple.json`

| Field | Owner | Values |
|---|---|---|
| `taffy` | skill | workflow/skill/agent name |
| `stage` | skill | current stage name |
| `status` | skill | `RUNNING`, `PAUSED`, `DONE`, `FAILED` |
| `awaiting_approval` | skill | stage name or `null` |
| `pipeline` | skill | `standard` if running 8-phase |
| `started_at` | skill | ISO 8601 |
| `updated_at` | skill | ISO 8601 |
| `state` | TUI | `running` or `exited` |
| `ts` | TUI | ISO 8601 |

**Merge-not-overwrite:** read existing file, update only owned fields, re-write.

### `.claude/state/approval-pending.txt`

Skill writes stage name. TUI deletes when user presses `[a]`.

### `.claude/state/sessions.json`

TUI writes harness→session-ID map. Skill reads for session resume.

## Skip Conditions

- `spike/*` and `chore/*` branches: skip Spec-Kit stages, run implementation stages.
- Stage `when: ui:true` on a `ui: false` story: skip silently, log `[pipeline-runner] SKIP stage=<name> reason=ui:false`.
