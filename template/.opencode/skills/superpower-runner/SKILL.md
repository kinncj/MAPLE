---
name: superpower-runner
description: "Execute a named superpower workflow from .claude/superpowers/<name>.yaml. Chains agents and skills in declared stage order, pausing at human-approval gates. Tracks pipeline state in .claude/state/maple.json."
---

# SKILL: superpower-runner

## What It Does

Reads a superpower workflow definition (`.claude/superpowers/<name>.yaml`), executes each stage in order by delegating to the appropriate agent or skill, and pauses at `gate: human-approval` stages until the human confirms.

Pipeline state is written to `.claude/state/maple.json` at every transition so the TUI can show progress.

## Usage

```
/superpower-runner <name>
```

Examples:
```
/superpower-runner new-ui-feature
/superpower-runner api-endpoint
/superpower-runner bugfix
/superpower-runner design-refresh
```

Available superpowers are in `.claude/superpowers/` — list them with:
```bash
ls .claude/superpowers/*.yaml | grep -v schema
```

## Execution Protocol

### 1. Load the workflow

```bash
cat .claude/superpowers/<name>.yaml
```

Parse the `stages:` list. Resolve `depends_on` to execution order.

### 2. Initialise state

Write to `.claude/state/maple.json`:

```json
{
  "superpower": "<name>",
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
3. Output:
```
SUPERPOWER PAUSED — awaiting human approval
Stage:    <stage-name>
Artifact: <artifact path or description>

Review the output above. When approved, reply "approved" or "continue".
I will not advance to the next stage until you confirm.
```
4. Wait for human response. Resume only on explicit approval.
5. On resume: update `maple.json` to `RUNNING`, advance to next stage.

### 5. Completion

When all stages are done:

```json
{
  "superpower": "<name>",
  "stage": "DONE",
  "status": "DONE",
  "awaiting_approval": null,
  "updated_at": "<iso8601>"
}
```

Output:
```
SUPERPOWER COMPLETE — <name>
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

`.claude/state/maple.json` fields:

| Field | Type | Values |
|---|---|---|
| `superpower` | string | workflow name |
| `stage` | string | current stage name |
| `status` | string | `RUNNING`, `PAUSED`, `DONE`, `FAILED` |
| `awaiting_approval` | string\|null | stage name blocked on human approval |
| `pipeline` | string | `standard` if running 8-phase |
| `started_at` | string | ISO 8601 |
| `updated_at` | string | ISO 8601 |

## Skip Conditions

- `spike/*` and `chore/*` branches: skip Spec-Kit stages but run implementation stages.
- Stage `when: ui:true` on a `ui: false` story: skip silently, log `[superpower] SKIP stage=<name> reason=ui:false`.
