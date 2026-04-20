# Example: API Endpoint via `superpower:api-endpoint`

This example shows how to use the `api-endpoint` superpower to scaffold a
REST endpoint end-to-end in a single workflow.

---

## Goal

Add `GET /api/users/:id/export-history` — returns paginated export history
for a user, respecting their subscription tier.

---

## Firing the Superpower

In the TUI, press `F` to open the Superpower picker:

```
  filter: api█

  ▸ api-endpoint       Scaffold REST/GraphQL endpoint with spec, tests, OpenAPI
    new-ui-feature     Full-stack UI feature with design, spec, and a11y
    bugfix             Reproduce → fix → validate
    design-refresh     Update design tokens and regenerate component mockups
```

Press `Enter`. A confirmation preview shows the stages:

```
  Superpower: api-endpoint
  ─────────────────────────────────────────
  Stage 1   spec-kit            problem → spec → plan → tasks
  Stage 2   gherkin-authoring   write Gherkin for each task
  Stage 3   @backend            implement endpoint + unit tests
  Stage 4   cucumber-automation sync .feature files
  Stage 5   gate: human-approval
  Stage 6   pipeline: standard-8-phase (validate → document → gate)
  ─────────────────────────────────────────
  Input: problem_statement required
  Fire? [Enter] / Cancel [Esc]
```

Enter the problem statement when prompted:

```
GET /api/users/:id/export-history
Returns paginated export records for a user.
Respects subscription tier: free users see last 10, pro users see all.
Requires: authenticated, user can only access own history.
```

---

## What the Superpower Produces

### Spec-Kit artifacts (`docs/specs/export-history-api/`)

```
PROBLEM.md   ← problem statement, user voice
SPEC.md      ← endpoint contract, auth requirements, pagination spec
PLAN.md      ← implementation approach, ADR for pagination strategy
TASKS.md     ← 2 tasks: endpoint + integration tests
```

### Story files

`docs/stories/export-history-endpoint-<ts>-0001/Story.md` — `ui: false`

```gherkin
Feature: User export history API

  @story:0045 @epic:export @priority:medium
  Scenario: Authenticated user retrieves their export history
    Given I am authenticated as user "u-123"
    And I have 15 export records
    When I GET /api/users/u-123/export-history?page=1&limit=10
    Then the response status is 200
    And the response body contains 10 records
    And the response includes a "next_page" cursor

  Scenario: Free tier user sees only last 10 records
    Given I am authenticated as a free-tier user with 20 export records
    When I GET /api/users/u-123/export-history
    Then the response contains exactly 10 records
    And the response does not include a "next_page" cursor

  Scenario: User cannot access another user's history
    Given I am authenticated as user "u-456"
    When I GET /api/users/u-123/export-history
    Then the response status is 403
```

### Backend implementation

`@backend` specialist opens PR `export/0045-export-history-api`:
- Route handler, service layer, repository query
- Unit tests (happy path + auth + pagination + tier logic)
- OpenAPI spec updated at `docs/specs/current/contracts/openapi.yaml`

### Feature files

`tests/features/export-history-endpoint.feature` — synced from story

### ADR (triggered by pagination decision)

`docs/specs/adrs/ADR-003-cursor-pagination-export-history.md`

---

## Key Differences from `/feature`

| `/feature` | `superpower:api-endpoint` |
|---|---|
| 8-phase pipeline, full scope | Scoped to API-only, skips design gate |
| Manual spec-kit progression | Automated spec-kit → story emit |
| No OpenAPI update | OpenAPI contract updated as part of pipeline |
| No Gherkin pre-generation | Gherkin authored before implementation |
