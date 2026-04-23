---
name: pipeline-runner
description: "Universal dispatcher: run a named taffy workflow (.opencode/taffy/<name>.yaml), a skill, or a sub-agent. Falls back to skills.sh registry when a skill is not found locally. Tracks all runs in .claude/state/maple.json so the maple TUI shows live progress."
---

# SKILL: pipeline-runner

## What It Does

Dispatches any named workflow, skill, or agent. Resolution order:

1. **Taffy workflow** — look for `.opencode/taffy/<name>.yaml`; if found, execute each stage in order
2. **Skill (local)** — look for `.opencode/skills/<name>/`; if found, invoke the skill
3. **Agent** — look for `.opencode/agents/<name>.md`; if found, delegate to `@<name>`
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
/pipeline-runner bugfix
/pipeline-runner design-refresh
```

Available taffy workflows are in `.opencode/taffy/` — list them with:
```bash
ls .opencode/taffy/*.yaml | grep -v schema
```

## Execution Protocol

### 1. Resolve the target

```bash
# Check taffy first
[ -f ".opencode/taffy/<name>.yaml" ] && dispatch=taffy
# Then local skill
[ -d ".opencode/skills/<name>" ] && dispatch=skill
# Then agent
[ -f ".opencode/agents/<name>.md" ] && dispatch=agent
# Fallback: fetch from skills.sh registry, then retry
if [ -z "$dispatch" ] && command -v npx &>/dev/null; then
  echo "pipeline-runner: '<name>' not found locally — checking skills.sh…"
  npx --yes skills add kinncj/maple@<name> -a opencode -y 2>/dev/null \
    || npx --yes skills add <name> -a opencode -y 2>/dev/null \
    || true
  [ -d ".opencode/skills/<name>" ] && dispatch=skill
fi
```

If nothing matches: `pipeline-runner: no taffy workflow, skill, or agent named '<name>' (also checked skills.sh registry)`

### 2. Load the workflow (taffy only)

```bash
cat .opencode/taffy/<name>.yaml
```

Parse the `stages:` list. Resolve `depends_on` to execution order.

### 2. Initialise state

Write to `.claude/state/maple.json`:

```json
{
  "taffy": "<name>",
  "stage": "<first-stage-name>",
  "status": "RUNNING",
  "awaiting_approval": null,
  "started_at": "<iso8601>",
  "updated_at": "<iso8601>"
}
```

Create `.claude/state/` if it doesn't exist.

### 3. Execute each stage

For each stage in order:

**Check `when:` guard:**
- `when: ui:true` — read the current story's `ui:` frontmatter field. Skip stage if `ui: false`.
- `when: ui:false` — skip if `ui: true`.
- `when: always` — always run.

**Check `depends_on`:** All listed stages must have status `DONE` before this stage starts.

**Dispatch:**
- `agent: <name>` → delegate to `@<name>` with the current story context
- `skill: <name>` → invoke the skill directly
- `pipeline: standard` → run the full 8-phase orchestrator pipeline

**After each stage completes**, update `maple.json`:
```json
{
  "stage": "<current-stage>",
  "status": "RUNNING",
  "updated_at": "<iso8601>"
}
```

### 4. Human-approval gates

When a stage has `gate: human-approval`:

1. Complete the stage work (produce the artifact).
2. Write PAUSED state to `maple.json`:
```json
{
  "stage": "<stage-name>",
  "status": "PAUSED",
  "awaiting_approval": "<stage-name>",
  "updated_at": "<iso8601>"
}
```
3. Write the stage name to `.claude/state/approval-pending.txt` so the TUI can surface it:
```bash
echo "<stage-name>" > .claude/state/approval-pending.txt
```
4. Output:
```
TAFFY PAUSED — awaiting human approval
Stage:    <stage-name>
Artifact: <artifact path or description>

Approve via the maple TUI ([P] pipeline → [a] approve) or reply "approved" / "continue".
I will not advance to the next stage until approval is confirmed.
```
5. Poll for the approval signal — the TUI deletes the file when the user presses [a]:
```bash
until [ ! -f .claude/state/approval-pending.txt ]; do sleep 2; done
```
   Also accept an explicit "approved" or "continue" reply in chat as an alternative.
6. On resume: update `maple.json` to `RUNNING`, advance to next stage.

### Session context

On startup, read `.claude/state/sessions.json` if it exists — it contains pinned session IDs for each harness:

```json
{ "claude": "<uuid>", "opencode": "<id>", "copilot": "<id>" }
```

Use the matching session ID when the taffy workflow needs to resume or continue work within an existing agent session. If the file is absent or the relevant key is missing, start a new session normally.

### 5. Completion

When all stages are done:

```json
{
  "taffy": "<name>",
  "stage": "DONE",
  "status": "DONE",
  "awaiting_approval": null,
  "updated_at": "<iso8601>"
}
```

Output:
```
TAFFY COMPLETE — <name>
Stages run: N
Duration:   <elapsed>

Next steps: <PR creation if standard-8-phase was last stage>
```

## Failure Handling

If a stage fails after 3 attempts:

```json
{
  "stage": "<stage-name>",
  "status": "FAILED",
  "error": "<failure summary>",
  "updated_at": "<iso8601>"
}
```

Stop execution. Report to human with the failed stage name and error. Do not attempt subsequent stages.

## State File Reference

All state files live in `.claude/state/`. Both this skill and the maple TUI read and write these files — they are the shared communication channel between the running agent and the dashboard.

### `.claude/state/maple.json`

Written by the skill at every stage transition. The TUI reads it to display pipeline progress. **Do not overwrite fields you don't own** — the TUI writes `state` and `ts` recovery marker fields into this same file; merge your fields on top.

| Field | Owner | Values |
|---|---|---|
| `taffy` | skill | workflow name |
| `stage` | skill | current stage name |
| `status` | skill | `RUNNING`, `PAUSED`, `DONE`, `FAILED` |
| `awaiting_approval` | skill | stage name blocked on human approval, or `null` |
| `pipeline` | skill | `standard` if running 8-phase |
| `started_at` | skill | ISO 8601 |
| `updated_at` | skill | ISO 8601 |
| `state` | TUI | `running` or `exited` (recovery marker) |
| `ts` | TUI | ISO 8601 (recovery marker timestamp) |

### `.claude/state/approval-pending.txt`

Written by the skill: contains the stage name waiting for approval.
Deleted by the TUI: when the user presses `[a]` in the pipeline overlay.
The skill polls for deletion; TUI polls for creation.

### `.claude/state/sessions.json`

Written by the TUI: maps harness name → pinned session ID.
Read by the skill: use pinned session IDs when resuming within an existing session.

```json
{ "claude": "<uuid>", "opencode": "<id>", "copilot": "<id>" }
```

## Skip Conditions

- `spike/*` and `chore/*` branches: skip Spec-Kit stages but run implementation stages.
- Stage `when: ui:true` on a `ui: false` story: skip silently, log `[taffy] SKIP stage=<name> reason=ui:false`.
