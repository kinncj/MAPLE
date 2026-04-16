#!/usr/bin/env bash
# .claude/hooks/post-bash.sh
# PostToolUse[Bash] — runs after every shell command Claude executes.
# Surfaces clear pass/fail signals after test and lint commands.
# stdout is shown to Claude as context.

INPUT=$(cat)

COMMAND=$(python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('tool_input', d).get('command', ''))
except Exception:
    print('')
" <<< "$INPUT" 2>/dev/null || echo "")

EXIT_CODE=$(python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    r = d.get('tool_response', {})
    print(r.get('exit_code', r.get('exitCode', 0)))
except Exception:
    print(0)
" <<< "$INPUT" 2>/dev/null || echo "0")

# ── Test commands ─────────────────────────────────────────────────────────────
if echo "$COMMAND" | grep -qE '(make test|pytest|jest|vitest|npx playwright|gradle test|mvn test|dotnet test)'; then
    if [ "$EXIT_CODE" -eq 0 ]; then
        echo "✓ Tests passed."
    else
        echo "✗ Tests failed (exit $EXIT_CODE). The QA gate is NOT satisfied. Do not proceed to the next pipeline phase."
    fi
fi

# ── Lint commands ─────────────────────────────────────────────────────────────
if echo "$COMMAND" | grep -qE '(make lint|eslint|pylint|flake8|golint|dotnet format)'; then
    if [ "$EXIT_CODE" -eq 0 ]; then
        echo "✓ Lint passed."
    else
        echo "✗ Lint failed (exit $EXIT_CODE). Fix lint errors before proceeding."
    fi
fi

# ── make test-all (Phase 8 gate) ──────────────────────────────────────────────
if echo "$COMMAND" | grep -q 'make test-all'; then
    if [ "$EXIT_CODE" -eq 0 ]; then
        echo "✓ Phase 8 gate passed — all test layers green. Safe to create PR."
    else
        echo "✗ Phase 8 gate FAILED. Return to Phase 5 for the failing component."
    fi
fi

exit 0
