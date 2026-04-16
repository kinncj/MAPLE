#!/usr/bin/env bash
# .claude/hooks/post-write.sh
# PostToolUse[Write|Edit] — runs after Claude writes or edits a file.
# Receives tool input+response as JSON on stdin.
# stdout is shown to Claude as context. Exit code is informational only.

INPUT=$(cat)
FILE=$(python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    ti = d.get('tool_input', d)
    print(ti.get('file_path', ti.get('path', '')))
except Exception:
    print('')
" <<< "$INPUT" 2>/dev/null || echo "")

[ -z "$FILE" ] && exit 0

# ── Story file: validate frontmatter ─────────────────────────────────────────
if echo "$FILE" | grep -qE '^docs/stories/.+\.md$' && [ "$FILE" != *"_template.md" ]; then
    if [ -f "scripts/sdlc/validate-frontmatter.sh" ]; then
        if bash scripts/sdlc/validate-frontmatter.sh "$FILE" 2>&1; then
            echo "✓ frontmatter OK: $FILE"
        else
            echo "⚠ frontmatter issues in $FILE — fix before committing."
        fi
    fi
    exit 0
fi

# ── Source file: lint + unit tests ───────────────────────────────────────────
if echo "$FILE" | grep -qE '\.(ts|tsx|js|jsx|py|java|cs|go|rs|rb)$'; then
    if [ -f "Makefile" ]; then
        echo "Running lint..."
        make lint 2>&1 | tail -8 || echo "⚠ lint found issues"

        echo "Running unit tests..."
        make test 2>&1 | tail -15 || echo "⚠ unit tests failed — fix before moving on"
    fi
fi

exit 0
