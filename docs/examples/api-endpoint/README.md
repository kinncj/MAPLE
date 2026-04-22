# Example: API Endpoint — Export CSV

This example traces an API-only story (`ui: false`) through the `api-endpoint` taffy workflow: Spec-Kit → Gherkin → contract → 8-phase pipeline.

## Step 1 — Problem Statement

**Who is affected:** Data analysts who need to export user activity data.
**Current situation:** Data is only accessible via the database. No API endpoint exists.
**Impact:** Analysts run manual SQL queries against production. Read-only risk and ops overhead.
**Desired outcome:** A `GET /api/v1/exports/activity.csv` endpoint that returns the last 90 days of activity in CSV format, gated by API key.

## Step 2 — Story File

Frontmatter:
```yaml
id: "data-export-0001"
title: "API endpoint: export activity as CSV"
epic: "data-export"
priority: "medium"
ui: false
adr_required: false
milestone: "v1.1"
labels:
  - "type:feature"
  - "priority:medium"
  - "phase:discover"
```

No design phase — `ui: false` skips all design gates.

## Step 3 — Gherkin

```gherkin
@story:data-export-0001 @epic:data-export @priority:medium
Feature: Activity CSV Export

  Scenario: Successful export with valid API key
    Given the request includes a valid API key
    When GET /api/v1/exports/activity.csv is called
    Then the response status is 200
    And the Content-Type is "text/csv"
    And the body contains activity rows for the last 90 days

  Scenario: Rejected without API key
    Given no API key is present
    When GET /api/v1/exports/activity.csv is called
    Then the response status is 401

  Scenario: Empty dataset
    Given no activity exists in the last 90 days
    When GET /api/v1/exports/activity.csv is called
    Then the response status is 200
    And the body contains only the CSV header row
```

## Step 4 — Contract (OpenAPI fragment)

```yaml
/api/v1/exports/activity.csv:
  get:
    summary: Export user activity as CSV
    security:
      - ApiKeyAuth: []
    parameters:
      - name: days
        in: query
        schema:
          type: integer
          default: 90
    responses:
      "200":
        content:
          text/csv:
            schema:
              type: string
      "401":
        description: Missing or invalid API key
```

## Step 5 — Pipeline

No design agents invoked. Orchestrator routes directly:
DISCOVER → ARCHITECT (contract) → PLAN → INFRA → IMPLEMENT → VALIDATE → DOCUMENT → FINAL GATE

## Step 6 — SDLC Gates

```
[frontmatter]    OK    docs/stories/data-export-...-0001.md
[design-gate]    SKIP  ui:false — no design gate required
[a11y-gate]      SKIP  ui:false — no a11y gate required
[adr-required]   SKIP  adr_required:false
```
