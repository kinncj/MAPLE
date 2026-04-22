---
name: pipeline-runner
description: "Universal dispatcher: run a named taffy workflow (.claude/taffy/<name>.yaml), a skill (/skill-name), or a sub-agent (@agent-name). Tracks all runs in .claude/state/maple.json so the maple TUI shows live progress."
---

# SKILL: pipeline-runner

## What It Does

Dispatches any named workflow, skill, or agent from a single entry point. Resolution order:

1. **Taffy workflow** — look for `.claude/taffy/<name>.yaml`; if found, execute each stage in order
2. **Skill** — look for `.claude/skills/<name>/`; if found, invoke the skill
3. **Agent** — look for `.claude/agents/<name>.md`; if found, delegate to `@<name>`

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
# Then skill
[ -d ".claude/skills/<name>" ] && dispatch=skill
# Then agent
[ -f ".claude/agents/<name>.md" ] && dispatch=agent
```

If nothing matches, report: `pipeline-runner: no taffy workflow, skill, or agent named '<name>'`

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
5. Poll: `timeout 540 bash -c 'until [ ! -f .claude/state/approval-pending.txt ]; do sleep 2; done'`
   - On timeout (exit 124), re-run the same poll. The Bash tool caps at 10 min per call; re-polling across calls lets approval delays exceed that bound.
   - Also accept an explicit "approved" / "continue" reply in chat.
   - When the user approves via the maple TUI ([P] → [a]), the TUI deletes the pending file **and** sends a "continue" keystroke to the agent's pane via the active multiplexer (outer tmux/zellij, or a detached `tmux new-session` wrapper). Either signal is sufficient to resume.
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
