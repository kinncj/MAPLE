# tests/features/

Gherkin `.feature` files are **generated at build time** by the `qa-cucumber` skill.

Do not manually edit files in this directory. They are derived from story files in `docs/stories/`.

## How it works

The `qa-cucumber` skill reads `docs/stories/*.md`, extracts fenced ` ```gherkin ` blocks,
and writes one `.feature` file per story to this directory.

Run manually:
```bash
# From Claude Code or OpenCode
@qa-cucumber extract-features
```

Or via Makefile:
```bash
make test-features-sync
```

Step definition scaffolding:
```bash
make test-features-scaffold
```

Step definitions live in `tests/step_definitions/` (TypeScript) or `tests/steps/` (Python).
They are NOT overwritten by re-running the sync — only new scenarios get new step stubs.
