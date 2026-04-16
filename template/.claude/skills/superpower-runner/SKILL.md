# SKILL: superpower-runner

## Purpose

Load a superpower YAML declaration, resolve its stages in order, and dispatch each stage to the appropriate agent or skill. Respects `when:` conditions, halts at `gate: human-approval` stages, and streams progress to the caller.

## Inputs

| Field | Source | Example |
|---|---|---|
| `superpower_name` | CLI arg or TUI picker | `new-ui-feature` |
| `inputs` | User-provided key-value map | `{epic: user-auth, feature_slug: reset-password}` |

## Superpower Discovery

```bash
find .claude/superpowers -name "*.yaml" ! -name "schema.yaml" | sort
```

The TUI calls this to populate the superpower picker.

## Load and Validate a Superpower

```python
import yaml, sys, os

def load_superpower(name):
    path = f".claude/superpowers/{name}.yaml"
    if not os.path.exists(path):
        raise FileNotFoundError(f"Superpower not found: {path}")
    sp = yaml.safe_load(open(path))
    required = ["name", "description", "version", "stages"]
    for field in required:
        if field not in sp:
            raise ValueError(f"Superpower {name} missing required field: {field}")
    return sp

sp = load_superpower(sys.argv[1])
print(f"[superpower] LOADED  {sp['name']}  v{sp['version']}  stages={len(sp['stages'])}")
```

## Evaluate a `when:` Condition

```python
def eval_when(condition, context):
    """
    context = {
      "identity.tokens_missing": not os.path.exists("docs/design/identity/tokens.json"),
      "story.ui_true": current_story_has_ui_true(),
      "branch.not_spike": not current_branch().startswith("spike/"),
      "always": True,
    }
    For template expressions like '{{scope}} != tokens-only', evaluate inline.
    """
    if condition is None or condition == "always":
        return True
    # Template expression (e.g. "{{scope}} != mockups-only")
    import re
    m = re.match(r'\{\{(\w+)\}\}\s*(!=|==)\s*["\']?([\w-]+)["\']?', condition)
    if m:
        var, op, val = m.groups()
        actual = context.get("inputs", {}).get(var, "")
        return (actual != val) if op == "!=" else (actual == val)
    # Lookup key
    return context.get(condition, True)
```

## Execute Stages

```python
def run_superpower(sp, user_inputs):
    context = build_context(user_inputs)
    total = len(sp["stages"])

    for i, stage in enumerate(sp["stages"], 1):
        label = stage.get("description", str(stage))
        when  = stage.get("when", "always")

        if not eval_when(when, context):
            print(f"[superpower] SKIP  stage {i}/{total}: {label}  (when={when})")
            continue

        print(f"[superpower] STAGE {i}/{total}: {label}")

        if "gate" in stage:
            # human-approval gate — halt execution
            print(f"[superpower] GATE  Human approval required: {label}")
            print(f"[superpower] PAUSED — resume by running: squad :resume {sp['name']}")
            return "paused"

        elif "agent" in stage:
            dispatch_agent(stage["agent"], stage.get("inputs", {}), user_inputs)

        elif "skill" in stage:
            dispatch_skill(stage["skill"], stage.get("inputs", {}), user_inputs)

        elif "pipeline" in stage:
            if stage["pipeline"] == "standard-8-phase":
                dispatch_agent("orchestrator", {"mode": "standard-8-phase"}, user_inputs)

    print(f"[superpower] COMPLETE  {sp['name']}")
    return "complete"
```

## Dispatch to Agent

```python
def dispatch_agent(agent_name, stage_inputs, user_inputs):
    """
    For Claude Code: uses Task tool syntax.
    For OpenCode: uses task tool with subagent_type.
    This function generates the prompt — actual dispatch is done by the calling agent.
    """
    merged = {**user_inputs, **stage_inputs}
    prompt = build_prompt(agent_name, merged)
    print(f"[superpower] DISPATCH  agent={agent_name}")
    # The orchestrator handles actual Task invocation
    return {"agent": agent_name, "prompt": prompt}
```

## Resume After Gate

Superpower state is persisted in `.claude/state/squad.json`:

```json
{
  "active_superpower": "new-ui-feature",
  "current_stage": 4,
  "status": "paused",
  "inputs": { "epic": "user-auth", "feature_slug": "reset-password" }
}
```

Resume command (from TUI or CLI):
```bash
squad :resume new-ui-feature
# Reads .claude/state/squad.json, continues from current_stage
```

## Failure Modes

| Condition | Action |
|---|---|
| Superpower YAML not found | Log error. List available superpowers. |
| Required input missing | Prompt for missing input before starting. |
| Stage agent fails 3× | Mark stage `failed`. Pause superpower. Surface to human. |
| YAML schema violation | Reject before starting. Log specific field error. |
| `gate` not acknowledged | Do not advance. State is persisted. |

## Logging

```
[superpower] LOADED  new-ui-feature  v1.0.0  stages=11
[superpower] STAGE 1/11: Spec intake: PROBLEM → SPEC → PLAN → TASKS → story files
[superpower] STAGE 2/11: Await TASKS.md approval before design begins
[superpower] GATE   Human approval required
[superpower] PAUSED — resume with: squad :resume new-ui-feature
[superpower] SKIP   stage 5/11: Visual identity  (when=identity.tokens_missing → false)
[superpower] COMPLETE  new-ui-feature
```
