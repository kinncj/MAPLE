# SKILL: spec-kit

## Purpose

Enforce a structured Problem → Spec → Plan → Tasks progression before any agent enters the DISCOVER phase. Each artifact requires explicit human approval before the next begins. TASKS.md is the terminal artifact — it emits the story files that feed the 8-phase pipeline.

## State Machine

```
PROBLEM (draft) → PROBLEM (approved)
               → SPEC (draft) → SPEC (approved)
                              → PLAN (draft) → PLAN (approved)
                                             → TASKS (draft) → TASKS (approved)
                                                             → story files emitted → DISCOVER
```

No step may begin until the previous step's artifact is approved. Approval = human sets `status: approved` in the artifact frontmatter, or reacts with ✅ on the linked GitHub Issue.

## Artifact Locations

```
docs/specs/<epic>-<feature-slug>/
├── PROBLEM.md    # raw problem statement — user voice, no solution
├── SPEC.md       # formal spec — goals, non-goals, acceptance criteria
├── PLAN.md       # technical plan — approach, ADR triggers, risks
└── TASKS.md      # task decomposition — maps to story files
```

## PROBLEM.md Template

```markdown
---
epic: "{epic-slug}"
feature: "{feature-slug}"
status: draft        # draft | approved | rejected
approved_by: null
approved_at: null
---

## Problem Statement

<!-- Write in user voice. No solution language. Describe the pain, not the fix. -->

**Who is affected:** {user role(s)}
**Current situation:** {what happens today}
**Impact:** {why this matters — quantify if possible}
**Desired outcome:** {what good looks like, from the user's perspective}

## Open Questions

<!-- Things that must be answered before writing SPEC.md -->
- [ ] {question 1}
- [ ] {question 2}

## Out of Scope

<!-- Explicitly list what this problem statement does NOT cover -->
- {out of scope item}
```

## SPEC.md Template

```markdown
---
epic: "{epic-slug}"
feature: "{feature-slug}"
status: draft        # draft | approved | rejected
approved_by: null
approved_at: null
problem_approved: true   # must be true before this file is created
---

## Goals

- {measurable goal 1}
- {measurable goal 2}

## Non-Goals

- {explicitly excluded scope item}

## Acceptance Criteria

```gherkin
Feature: {Feature title}

  Scenario: {primary success path}
    Given {precondition}
    When {action}
    Then {outcome}

  Scenario: {failure / edge case}
    ...
```

## Constraints

- {technical constraint}
- {business constraint}

## Dependencies

- {system or team dependency}

## ADR Triggers

<!-- List decisions that will require an ADR in the PLAN phase -->
- [ ] {decision to make}
```

## PLAN.md Template

```markdown
---
epic: "{epic-slug}"
feature: "{feature-slug}"
status: draft        # draft | approved | rejected
approved_by: null
approved_at: null
spec_approved: true
---

## Technical Approach

{2–4 paragraph summary. Clean Architecture layers affected. SOLID concerns.}

## Component Breakdown

| Component | Layer | Agent | Notes |
|---|---|---|---|
| {name} | {domain/infra/ui} | {agent} | {notes} |

## ADRs Required

| Decision | Status | File |
|---|---|---|
| {decision} | pending | docs/architecture/{slug}-adr.md |

## Risk Register

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| {risk} | low/med/high | low/med/high | {mitigation} |

## Test Strategy

- Unit: {what and where}
- Integration: {what and where}
- E2E: {what and where}
- BDD: {Gherkin scenario files}
```

## TASKS.md Template

```markdown
---
epic: "{epic-slug}"
feature: "{feature-slug}"
status: draft        # draft | approved | rejected
approved_by: null
approved_at: null
plan_approved: true
stories_emitted: false
---

## Tasks

<!-- Each task maps to one story file. task_id becomes the story NNNN. -->

| ID | Title | Agent | Priority | ui | adr_required |
|---|---|---|---|---|---|
| 0001 | {story title} | {primary agent} | high | false | false |
| 0002 | {story title} | {primary agent} | medium | true | false |

## Emitted Stories

<!-- Populated automatically when TASKS is approved -->
- [ ] docs/stories/{epic}-{slug}-{timestamp}-0001.md
- [ ] docs/stories/{epic}-{slug}-{timestamp}-0002.md
```

## Validate an Artifact

```bash
validate_spec_artifact() {
  local file="$1"
  local required_status="$2"   # approved

  if [ ! -f "$file" ]; then
    echo "[spec-kit] MISSING  $file"
    return 1
  fi

  STATUS=$(python3 -c "
import re
m = re.search(r'^status:\s*(\w+)', open('$file').read(), re.MULTILINE)
print(m.group(1) if m else 'draft')
")

  if [ "$STATUS" != "$required_status" ]; then
    echo "[spec-kit] BLOCKED  $file  status=$STATUS  required=$required_status"
    return 1
  fi

  echo "[spec-kit] OK  $file  status=approved"
}

# Example: check full chain before emitting stories
EPIC="user-auth"
SLUG="reset-password"
BASE="docs/specs/${EPIC}-${SLUG}"

validate_spec_artifact "$BASE/PROBLEM.md" "approved" || exit 1
validate_spec_artifact "$BASE/SPEC.md"    "approved" || exit 1
validate_spec_artifact "$BASE/PLAN.md"    "approved" || exit 1
validate_spec_artifact "$BASE/TASKS.md"   "approved" || exit 1
```

## Emit Story Files from TASKS.md

When TASKS.md is approved, emit one story file per task row:

```bash
BASE="docs/specs/${EPIC}-${SLUG}"
TS=$(date -u +"%Y%m%d%H%M%S")

python3 - <<'EOF'
import re, os, sys
from datetime import datetime, timezone

tasks_path = sys.argv[1]
epic = sys.argv[2]
ts = sys.argv[3]
text = open(tasks_path).read()

# Parse table rows (skip header and separator)
rows = re.findall(
    r'^\|\s*(\d{4})\s*\|\s*(.*?)\s*\|\s*(.*?)\s*\|\s*(.*?)\s*\|\s*(true|false)\s*\|\s*(true|false)\s*\|',
    text, re.MULTILINE
)

os.makedirs('docs/stories', exist_ok=True)
emitted = []

for task_id, title, agent, priority, ui, adr_req in rows:
    slug = re.sub(r'[^a-z0-9]+', '-', title.lower()).strip('-')[:40]
    filename = f"docs/stories/{epic}-{slug}-{ts}-{task_id}.md"
    if os.path.exists(filename):
        print(f"[spec-kit] SKIP  {filename}  (exists)")
        continue
    content = f"""---
id: "{epic}-{task_id}"
title: "{title}"
epic: "{epic}"
priority: "{priority}"
ui: {ui}
adr_required: {adr_req}
milestone: null
labels:
  - "type:feature"
  - "priority:{priority}"
  - "phase:discover"
issue_number: null
issue_url: null
---

## Story

**As a** user,
**I want** {title.lower()},
**so that** TODO: define business outcome.

## Acceptance Criteria

```gherkin
@story:{epic}-{task_id} @epic:{epic} @priority:{priority}
Feature: {title}

  Scenario: TODO
    Given TODO
    When TODO
    Then TODO
```

## Definition of Done

- [ ] All Gherkin scenarios have passing step implementations
- [ ] Unit tests written and passing
- [ ] Code reviewed and approved

## ADR Links

<!-- Add ADR links here -->
"""
    with open(filename, 'w') as f:
        f.write(content)
    emitted.append(filename)
    print(f"[spec-kit] EMIT  {filename}")

# Mark stories_emitted in TASKS.md
text2 = open(tasks_path).read()
text2 = text2.replace('stories_emitted: false', 'stories_emitted: true')
open(tasks_path, 'w').write(text2)
EOF
"$BASE/TASKS.md" "$EPIC" "$TS"
```

## Skip Conditions

Spec-Kit is **skipped** (not required) for:
- Branches matching `spike/*` or `chore/*`
- Stories with `type:bug` label (use bugfix superpower instead)
- When `sdlc.mode: spike` in `project.config.yaml`

```bash
BRANCH=$(git branch --show-current)
if echo "$BRANCH" | grep -qE '^(spike|chore)/'; then
  echo "[spec-kit] SKIP  branch=$BRANCH — spike/chore exempt"
  exit 0
fi
```

## Failure Modes

| Condition | Action |
|---|---|
| Artifact missing | Create from template. Set `status: draft`. Halt until approved. |
| Artifact `status: rejected` | Stop. Surface rejection reason to human. Do not create next artifact. |
| `stories_emitted: true` already | Skip emit. Log `SKIP — stories already emitted`. |
| TASKS table empty | Reject TASKS.md. Require at least one task row. |
| Duplicate task IDs | Reject TASKS.md. IDs must be unique 4-digit values. |

## Logging

```
[spec-kit] VALIDATE  docs/specs/user-auth-reset/PROBLEM.md  status=approved
[spec-kit] VALIDATE  docs/specs/user-auth-reset/SPEC.md     status=approved
[spec-kit] BLOCKED   docs/specs/user-auth-reset/PLAN.md     status=draft
[spec-kit] EMIT      docs/stories/user-auth-reset-password-20250416143000-0001.md
[spec-kit] SKIP      spike/* branch — spec-kit not required
```
