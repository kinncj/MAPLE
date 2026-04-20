# Example: UI-Bearing Feature — End to End

This example traces a `ui: true` story from intake through merge using the full
Spec-Kit + Design Suite + 8-phase pipeline + TUI workflow.

---

## Story: Export Filtered Results as CSV

### 1. Intake via `squad req`

Launch the TUI and select **Requirements**. Type the problem statement:

```
As an analyst, I want to export the current filtered table to CSV
so that I can share results with stakeholders offline.

The export must respect active filters.
Only the last 90 days of data should be included.
File must be downloadable immediately (no email).
```

Press `Ctrl+D`. Claude converts this to a Gherkin story saved at:
`docs/stories/export-filtered-results-as-csv-<ts>-0001/Story.md`

The story gets `ui: true` added (detected from "export" being a data/UI action),
and the DoD checklist includes wireframe + mockup + a11y items automatically.

---

### 2. Spec-Kit: Problem → Spec → Plan → Tasks

Run `/feature export filtered CSV` in Claude Code. The orchestrator dispatches
`@spec-kit` to produce four artifacts in `docs/specs/export-filtered-csv/`:

**PROBLEM.md** (approved by user before SPEC begins):
```markdown
---
status: approved
approved_by: kinncj
approved_at: 2026-04-20
---
Analysts need to share query results. Current workflow requires screenshots.
CSV export tied to active filters prevents data leakage.
```

**SPEC.md** (approved before PLAN):
- Goals: single-click CSV download, filter-respecting, 90-day window
- Non-goals: Excel format, email delivery, scheduled exports
- Acceptance criteria: 5 Gherkin scenarios (happy path + edge cases)

**PLAN.md** (approved before TASKS):
- Frontend: new ExportButton component, uses React Query mutation
- Backend: `/api/export/csv` endpoint, streams response
- ADR required: response streaming approach

**TASKS.md** (emits story files on approval):
- Story 0042: ExportButton UI component
- Story 0043: CSV export API endpoint
- Story 0044: Filter-to-query serialisation

---

### 3. Design Intake Gate (triggered by `ui: true`)

Orchestrator detects `ui: true` in stories 0042 and automatically dispatches
`@wireframe-architect`.

**Wireframe** saved at `docs/design/wireframes/export-button.wireframe.md`:
```
┌─ Table Header ──────────────────────────────────────────────┐
│  Showing 47 results  (filter: status=shipped, last 90d)     │
│                                          [ Export CSV ↓ ]   │
└─────────────────────────────────────────────────────────────┘
```

User approves in TUI (`squad` → Stories pane → Enter → approve).

Since no `tokens.json` exists, `@visual-identity-designer` is dispatched,
producing `docs/design/identity/palette.json` and `docs/design/identity/tokens.json`.

**Mockup** built by `@ui-mockup-builder` using Mantine + tokens:
`docs/design/mockups/export-button.mockup.tsx`

User approves. `component-scaffold` skill runs, generating:
```
app/components/ExportButton/
├── index.tsx
├── ExportButton.stories.tsx
├── ExportButton.test.tsx
└── ExportButton.spec.ts
```

---

### 4. Standard 8-Phase Pipeline

DISCOVER → ARCHITECT → PLAN → INFRA → IMPLEMENT → VALIDATE → DOCUMENT → GATE

Key checkpoints:
- `@typescript` specialist opens PR `export/0042-export-button` from IMPLEMENT
- `@qa` writes failing Cucumber scenarios before implementation
- `@a11y-auditor` runs WCAG 2.2 AA audit; contrast ratio ✓, focus order ✓
- `@orchestrator` reviews PR, leaves comments, does not push to specialist branch
- All gates green → human approves → merge

---

### 5. Monitoring in the TUI

While work is in progress, `squad` dashboard shows:

```
┌─ Stories ─────────────────────┐  ┌─ Recent Agents ────────────────────┐
│ ▸ 0042 export-button  implement│  │ typescript   ExportButton.tsx      │
│   0043 export-api    implement │  │ qa           export.spec.ts        │
│   0044 filter-serial discover  │  │ a11y-auditor contrast check        │
└────────────────────────────────┘  └────────────────────────────────────┘
┌─ PRs ─────────────────────────┐  ┌─ QA / Gherkin ─────────────────────┐
│ ● #131 export-button ● open   │  │  3 feature file(s)                 │
│ ● #132 export-api    ● open   │  │  11 scenario(s) total              │
└────────────────────────────────┘  └────────────────────────────────────┘
```

Press `d` to view wireframes and tokens in the Design pane.
Press `F` to fire the `new-ui-feature` superpower on a new story.

---

### Artifacts Produced

| Artifact | Location |
|---|---|
| Spec-Kit | `docs/specs/export-filtered-csv/` |
| Story files | `docs/stories/export-*/Story.md` |
| Wireframe | `docs/design/wireframes/export-button.wireframe.md` |
| Design tokens | `docs/design/identity/tokens.json` |
| Mockup | `docs/design/mockups/export-button.mockup.tsx` |
| Component scaffold | `app/components/ExportButton/` |
| Feature files | `tests/features/export-*.feature` |
| ADR | `docs/specs/adrs/ADR-002-streaming-csv-export.md` |
| PR | `#131 export/0042-export-button` |
