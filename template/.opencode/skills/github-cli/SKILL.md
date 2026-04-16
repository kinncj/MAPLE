---
name: github-cli
description: "Orchestrate the full GitHub workflow — issues, PRs, labels, and project board — using the gh CLI. Use when automating GitHub operations."
---

# SKILL: GitHub CLI

## Issue Lifecycle

```
PO creates issues → Orchestrator assigns & labels → Agents update status → QA closes on pass
```

## Key Commands by Agent Role

### Product Owner
```bash
# Create issue for each story
gh issue create \
  --title "Story: {title}" \
  --body-file docs/specs/{slug}/stories.md \
  --label "story,must-have" \
  --milestone "v1.0"
```

### Orchestrator
```bash
# Assign and label
gh issue edit {number} --add-label "in-progress" --add-assignee "@me"
gh issue edit {number} --add-label "phase:architect" --remove-label "phase:discover"
gh issue comment {number} --body "Phase 2 ARCHITECT: design complete. Artifacts in docs/specs/{slug}/"

# Block on failure
gh issue edit {number} --add-label "blocked"
gh issue comment {number} --body "BLOCKED: {agent} failed 3 times on {task}. Human needed."
```

### QA Agent
```bash
# After writing failing test
gh issue edit {number} --add-label "tdd:red"
gh issue comment {number} --body "RED: Failing test at tests/unit/{File}.ts — message: {failure}"

# After full validation passes
gh issue close {number} --comment "All acceptance criteria passing. Validation: {summary}"
```

### Any Agent
```bash
# View issue
gh issue view {number}

# List in-progress issues
gh issue list --label "in-progress" --assignee "@me"

# Comment with progress
gh issue comment {number} --body "Task 3/7 GREEN: tests/unit/UserTests.ts passing."
```

## PR Workflow
```bash
# Create feature branch
git checkout -b feat/{slug}

# Create draft PR
gh pr create \
  --title "feat: {description}" \
  --body "Closes #{number}" \
  --base main \
  --draft

# Mark ready
gh pr ready

# Merge (squash)
gh pr merge --squash --subject "feat: {description} (#{number})"
```

## Label Conventions

| Label | Applied by | Meaning |
|-------|-----------|---------|
| `story` | PO | User story |
| `bug` | PO/Human | Bug report |
| `must-have` | PO | MoSCoW: Must |
| `should-have` | PO | MoSCoW: Should |
| `could-have` | PO | MoSCoW: Could |
| `phase:discover` | Orchestrator | Phase 1 active |
| `phase:architect` | Orchestrator | Phase 2 active |
| `phase:plan` | Orchestrator | Phase 3 active |
| `phase:infra` | Orchestrator | Phase 4 active |
| `phase:implement` | Orchestrator | Phase 5 active |
| `phase:validate` | Orchestrator | Phase 6 active |
| `phase:document` | Orchestrator | Phase 7 active |
| `phase:done` | Orchestrator | Phase 8 complete |
| `in-progress` | Orchestrator | Work started |
| `blocked` | Any agent | Needs human |
| `validated` | QA | All tests pass |
| `ready-for-review` | Orchestrator | PR ready |
| `tdd:red` | QA | Failing test written |
| `tdd:green` | QA | Tests passing |

## CI Monitoring
```bash
gh run list --workflow=ci.yml --limit 5
gh run watch {run-id}
gh pr checks {pr-number}
```

## Bootstrap Labels (one-time per repo)
```bash
./scripts/bootstrap-labels.sh
# or with specific repo:
./scripts/bootstrap-labels.sh owner/repo
```
