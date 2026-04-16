#!/usr/bin/env bash
# scripts/sdlc/spec-kit-gate.sh
# Blocks push if any feature has an incomplete spec-kit progression.
# Skips spike/*, chore/* branches and branches with no docs/specs/ entries.
set -euo pipefail

BRANCH=$(git branch --show-current 2>/dev/null || echo "")
if echo "$BRANCH" | grep -qE '^(spike|chore)/'; then
  echo "[spec-kit-gate] SKIP  branch=$BRANCH"
  exit 0
fi

[ -d "docs/specs" ] || exit 0

FAIL=0

for dir in docs/specs/*/; do
  [ -d "$dir" ] || continue
  FEATURE=$(basename "$dir")

  for artifact in PROBLEM SPEC PLAN TASKS; do
    FILE="$dir${artifact}.md"
    if [ ! -f "$FILE" ]; then
      # Missing artifacts before TASKS are only a failure if a later one exists
      NEXT=""
      case $artifact in
        PROBLEM) NEXT="$dir/SPEC.md" ;;
        SPEC)    NEXT="$dir/PLAN.md" ;;
        PLAN)    NEXT="$dir/TASKS.md" ;;
      esac
      [ -n "$NEXT" ] && [ -f "$NEXT" ] && {
        echo "[spec-kit-gate] FAIL  $FEATURE  $artifact.md missing but later artifact exists"
        FAIL=1
      }
      continue
    fi

    STATUS=$(python3 -c "
import re
m = re.search(r'^status:\s*(\w+)', open('$FILE').read(), re.MULTILINE)
print(m.group(1) if m else 'draft')
" 2>/dev/null || echo "draft")

    if [ "$STATUS" = "rejected" ]; then
      echo "[spec-kit-gate] FAIL  $FEATURE  $artifact.md rejected — resolve before pushing"
      FAIL=1
    elif [ "$STATUS" = "draft" ]; then
      # draft is acceptable unless it's TASKS (the terminal artifact)
      if [ "$artifact" = "TASKS" ] && [ -f "$dir/TASKS.md" ]; then
        EMITTED=$(python3 -c "
import re
m = re.search(r'^stories_emitted:\s*(true|false)', open('$FILE').read(), re.MULTILINE)
print(m.group(1) if m else 'false')
" 2>/dev/null || echo "false")
        if [ "$EMITTED" = "false" ]; then
          echo "[spec-kit-gate] WARN  $FEATURE  TASKS.md not yet approved/emitted — stories may not exist"
        fi
      fi
    fi
  done
done

exit $FAIL
