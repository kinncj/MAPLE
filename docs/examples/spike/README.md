# Example: Spike — Performance Investigation

This example shows a `spike/*` branch. Spec-Kit, design gates, and a11y gates are all skipped. The spike produces a document, not production code.

## When to use a spike

- Uncertain technical feasibility
- Need to benchmark competing approaches
- Exploring an unfamiliar library or service
- Time-boxed: 1–2 days maximum

## Step 1 — Story File

Branch: `spike/perf-audit-export`

Frontmatter:
```yaml
id: "spike-perf-0001"
title: "Spike: assess export performance at 1M rows"
epic: "data-export"
priority: "medium"
ui: false
adr_required: true
milestone: "v1.1"
labels:
  - "type:spike"
  - "priority:medium"
  - "phase:discover"
```

## Step 2 — What the spike skips

| Gate | Applies? | Reason |
|---|---|---|
| Spec-Kit (PROBLEM/SPEC/PLAN/TASKS) | No | `spike/*` branch |
| Wireframe / mockup approval | No | `ui: false` |
| A11y audit | No | `ui: false` |
| Frontmatter validation | Yes | Always runs |

## Step 3 — Spike output

The spike produces **one document only**:

```
docs/specs/data-export-export-csv/spike.md
```

Structure:
```markdown
## Spike: {title}

### Time-box
{N hours / days}

### Question(s) to answer
- {question 1}
- {question 2}

### Findings
{what was learned}

### Recommendation
{go / no-go / need more info}

### Follow-on stories
- {story title} — feeds back into Spec-Kit as a new PROBLEM.md
```

## Step 4 — Spike DoD

From `docs/dod/definition-of-done.md` (spike section):

- [ ] Spike document written at `docs/specs/<slug>/spike.md`
- [ ] Key findings summarised for the follow-on story
- [ ] No production code committed (spike branch only)

## Step 5 — No production merge

Spike branches do not merge to main. The `spike.md` document is cherry-picked or referenced by the follow-on story issue. The branch is deleted after the document is reviewed.
