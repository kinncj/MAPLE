# Example: UI Feature — Password Reset

This example traces a `ui: true` story through the `new-ui-feature` taffy workflow: Spec-Kit intake → design → 8-phase pipeline → a11y gate.

## Step 1 — Story File (written by Spec-Kit)

Run `/feature user can reset their password via email` in Claude Code. The orchestrator dispatches `@spec-kit`, which writes the Gherkin story directly to `docs/stories/`.

```
docs/stories/user-auth-0001.md
```

Frontmatter:
```yaml
id: "user-auth-0001"
title: "User can reset their password via email"
epic: "user-auth"
priority: "high"
ui: true
adr_required: false
phase: discover
labels:
  - "type:feature"
  - "priority:high"
status: draft
issue_number: null
```

Story body (Gherkin):
```gherkin
Feature: Password Reset

  Scenario: Successful reset request
    Given the user has a verified account with email "alice@example.com"
    When they submit the password reset form with that email
    Then they receive a reset link within 60 seconds
    And the link expires after 24 hours

  Scenario: Unknown email address
    Given no account exists for "ghost@example.com"
    When they submit the password reset form
    Then they see a generic confirmation message
    And no email is sent

  Scenario: Unauthenticated access to reset endpoint
    Given the user is not logged in
    When they attempt to call the password reset endpoint directly
    Then the response is 401 Unauthorized
```

Human approves story → `status: approved` in frontmatter.

## Step 2 — Design Flow (ui: true gates)

| Artifact | Location | Status needed |
|---|---|---|
| Wireframe | `docs/design/wireframes/user-auth-0001.wireframe.md` | `approved` |
| Palette | `docs/design/identity/palette.json` | `approved` |
| Mockup | `docs/design/mockups/user-auth-0001.mockup.tsx` | `approved` |
| A11y report | `docs/design/mockups/user-auth-0001.a11y.json` | no violations |

## Step 3 — Pipeline

Once design gates pass, orchestrator runs standard 8-phase pipeline:
DISCOVER → ARCHITECT → PLAN → INFRA → IMPLEMENT → VALIDATE → DOCUMENT → FINAL GATE

## Step 4 — SDLC Gates (pre-push)

```
[frontmatter]    OK    docs/stories/user-auth-0001.md
[design-gate]    OK    wireframe:user-auth-0001  status=approved
[design-gate]    OK    mockup:user-auth-0001     status=approved
[a11y-gate]      OK    docs/stories/user-auth-0001.md  violations=0
[adr-required]   SKIP  adr_required:false
```
