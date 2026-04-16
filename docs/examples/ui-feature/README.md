# Example: UI Feature — Password Reset

This example traces a `ui: true` story through the full `new-ui-feature` superpower: Spec-Kit intake → design → 8-phase pipeline → a11y gate.

## Step 1 — Problem Statement

```
docs/specs/user-auth-reset-password/PROBLEM.md  (status: approved)
```

**Who is affected:** Registered users who have forgotten their password.
**Current situation:** No self-service password reset exists. Users email support.
**Impact:** ~40 support tickets/week are password resets. 2-day average resolution time.
**Desired outcome:** User receives a reset link by email within 60 seconds and can set a new password without contacting support.

## Step 2 — Spec (Gherkin)

```
docs/specs/user-auth-reset-password/SPEC.md  (status: approved)
```

Key acceptance criterion (Gherkin):

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
```

## Step 3 — Story File (emitted by Spec-Kit)

```
docs/stories/user-auth-reset-password-20250416143000-0001.md
```

Frontmatter:
```yaml
id: "user-auth-0001"
title: "User can reset their password via email"
epic: "user-auth"
priority: "high"
ui: true
adr_required: false
milestone: "v1.0"
labels:
  - "type:feature"
  - "priority:high"
  - "phase:discover"
issue_number: null
issue_url: null
```

## Step 4 — Design Flow (ui: true gates)

| Artifact | Location | Status needed |
|---|---|---|
| Wireframe | `docs/design/wireframes/user-auth-0001.wireframe.md` | `approved` |
| Palette | `docs/design/identity/palette.json` | `approved` |
| Mockup | `docs/design/mockups/user-auth-0001.mockup.tsx` | `approved` |
| A11y report | `docs/design/mockups/user-auth-0001.a11y.json` | no violations |

## Step 5 — Pipeline

Once design gates pass, orchestrator runs standard 8-phase pipeline:
DISCOVER → ARCHITECT → PLAN → INFRA → IMPLEMENT → VALIDATE → DOCUMENT → FINAL GATE

## Step 6 — SDLC Gates (pre-push)

```
[frontmatter]    OK    docs/stories/user-auth-reset-password-...-0001.md
[design-gate]    OK    wireframe:user-auth-0001  status=approved
[design-gate]    OK    mockup:user-auth-0001     status=approved
[a11y-gate]      OK    docs/stories/...  violations=0
[spec-kit-gate]  OK    user-auth-reset-password  TASKS.md approved
```
