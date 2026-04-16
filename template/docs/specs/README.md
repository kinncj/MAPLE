# docs/specs/

Feature specification artifacts live here. One subdirectory per epic+feature, produced by the `spec-kit` agent before DISCOVER.

## Structure

```
docs/specs/
└── <epic>-<feature-slug>/
    ├── PROBLEM.md    # Problem statement — user voice, no solution
    ├── SPEC.md       # Formal specification — goals, non-goals, acceptance criteria
    ├── PLAN.md       # Technical plan — approach, ADRs, risks, test strategy
    └── TASKS.md      # Task decomposition → story files
```

## Progression

Each file must be `status: approved` before the next is created.

```
PROBLEM → SPEC → PLAN → TASKS → story files → DISCOVER
```

## Key Rules

- Spec-Kit is **skipped** for `spike/*` and `chore/*` branches and `type:bug` stories.
- TASKS.md emits story files into `docs/stories/` — do not create story files manually for spec-kit features.
- Once `stories_emitted: true` is set in TASKS.md, do not re-emit.
- All four files are committed to the repo and linked to the feature's GitHub Issue.
