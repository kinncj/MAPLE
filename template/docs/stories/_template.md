---
epic: __EPIC__
story_id: "0001"
story_slug: __SLUG__
created_at: __TIMESTAMP__
priority: medium
domain: __DOMAIN__
specialist_hints: []
ui: false
adr_required: false
issue_number: null
---

# __TITLE__

## Narrative

As a __ROLE__, I want to __ACTION__ so that __OUTCOME__.

## Scenarios

```gherkin
@story:__STORY_ID__ @epic:__EPIC__ @priority:__PRIORITY__
Feature: __FEATURE_TITLE__

  Scenario: __HAPPY_PATH__
    Given __PRECONDITION__
    When __ACTION__
    Then __EXPECTED_OUTCOME__

  Scenario: __EDGE_CASE__
    Given __PRECONDITION__
    When __EDGE_ACTION__
    Then __EDGE_OUTCOME__
```

## Definition of Done

- [ ] Unit tests green
- [ ] Integration tests green
- [ ] Cucumber/Behave scenarios green
- [ ] Wireframe approved (required when `ui: true`)
- [ ] Mockup approved (required when `ui: true`)
- [ ] A11y audit passed (required when `ui: true`)
- [ ] ADRs linked where required
- [ ] CHANGELOG entry added
- [ ] PR description references this story

## ADR Links

<!-- populated by adr-author agent -->
