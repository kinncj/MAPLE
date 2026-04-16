#!/usr/bin/env bash
# tests/cli/test_ai_squad.sh — CLI smoke tests for ai-squad
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CLI="$REPO_ROOT/scripts/ai-squad"
TEMPLATE_DIR="$REPO_ROOT/template"
PASS=0
FAIL=0

# ─── helpers ──────────────────────────────────────────────────────────────────
ok()   { printf "  \033[1;32m✓\033[0m  %s\n" "$1"; PASS=$((PASS + 1)); }
fail() { printf "  \033[1;31m✗\033[0m  %s\n" "$1"; FAIL=$((FAIL + 1)); }

assert_exit_ok() {
  local label="$1"; shift
  if "$@" >/dev/null 2>&1; then
    ok "$label"
  else
    fail "$label (expected exit 0, got $?)"
  fi
}

assert_exit_fail() {
  local label="$1"; shift
  if ! "$@" >/dev/null 2>&1; then
    ok "$label"
  else
    fail "$label (expected non-zero exit, got 0)"
  fi
}

# ─── tests ────────────────────────────────────────────────────────────────────
printf "\n\033[1m  CLI Tests\033[0m  →  %s\n\n" "$CLI"

# 1. Executable
if [[ -x "$CLI" ]]; then
  ok "ai-squad is executable"
else
  fail "ai-squad is not executable"
fi

# 2. help exits 0
assert_exit_ok  "ai-squad help exits 0"         "$CLI" help
assert_exit_ok  "ai-squad --help exits 0"       "$CLI" --help
assert_exit_ok  "ai-squad -h exits 0"           "$CLI" -h

# 3. Unknown command exits non-zero
assert_exit_fail "unknown command exits non-zero" "$CLI" totally-not-a-command

# 4. init copies template files
TEST_TMP="$(mktemp -d)"
TEST_PROJECT="$TEST_TMP/test-project"
trap 'rm -rf "$TEST_TMP"' EXIT

# Pipe "n" to skip labels prompt; pass absolute path to init
printf 'n\n' | "$CLI" init "$TEST_PROJECT" >/dev/null 2>&1 || true

if [[ -d "$TEST_PROJECT" ]]; then
  ok "init creates project directory"
else
  fail "init did not create project directory"
fi

# 5. init copies Claude Code agents
CLAUDE_AGENTS="$TEST_PROJECT/.claude/agents"
if [[ -d "$CLAUDE_AGENTS" ]]; then
  COUNT=$(find "$CLAUDE_AGENTS" -name "*.md" | wc -l | tr -d ' ')
  EXPECTED=$(find "$TEMPLATE_DIR/.claude/agents" -name "*.md" | wc -l | tr -d ' ')
  if [[ "$COUNT" -eq "$EXPECTED" ]]; then
    ok "init copies $COUNT Claude Code agent files"
  else
    fail "init copied $COUNT agents, expected $EXPECTED"
  fi
else
  fail "init did not create .claude/agents/"
fi

# 6. init copies OpenCode agents
OC_AGENTS="$TEST_PROJECT/.opencode/agents"
if [[ -d "$OC_AGENTS" ]]; then
  COUNT=$(find "$OC_AGENTS" -name "*.md" | wc -l | tr -d ' ')
  EXPECTED=$(find "$TEMPLATE_DIR/.opencode/agents" -name "*.md" | wc -l | tr -d ' ')
  if [[ "$COUNT" -eq "$EXPECTED" ]]; then
    ok "init copies $COUNT OpenCode agent files"
  else
    fail "init copied $COUNT agents, expected $EXPECTED"
  fi
else
  fail "init did not create .opencode/agents/"
fi

# 7. init copies Makefile
if [[ -f "$TEST_PROJECT/Makefile" ]]; then
  ok "init copies Makefile"
else
  fail "init did not copy Makefile"
fi

# 8. init copies CLAUDE.md
if [[ -f "$TEST_PROJECT/CLAUDE.md" ]]; then
  ok "init copies CLAUDE.md"
else
  fail "init did not copy CLAUDE.md"
fi

# 9. init copies opencode.json
if [[ -f "$TEST_PROJECT/opencode.json" ]]; then
  ok "init copies opencode.json"
else
  fail "init did not copy opencode.json"
fi

# 10. no agent file contains a model: line (model names are not hardcoded)
CLAUDE_MODEL_COUNT=$(grep -rl '^model:' "$TEST_PROJECT/.claude/agents/" 2>/dev/null | wc -l | tr -d ' ') || true
OC_MODEL_COUNT=$(grep -rl '^model:' "$TEST_PROJECT/.opencode/agents/" 2>/dev/null | wc -l | tr -d ' ') || true
if [[ "$CLAUDE_MODEL_COUNT" -eq 0 && "$OC_MODEL_COUNT" -eq 0 ]]; then
  ok "no agent file contains a hardcoded model: line"
else
  fail "found hardcoded model: lines ($CLAUDE_MODEL_COUNT Claude Code, $OC_MODEL_COUNT OpenCode)"
fi

# 11. opencode.json does not contain a top-level model field
if python3 -c "import json,sys; d=json.load(open('$TEST_PROJECT/opencode.json')); sys.exit(1 if 'model' in d else 0)" 2>/dev/null; then
  ok "opencode.json has no hardcoded model field"
else
  fail "opencode.json still contains a top-level model field"
fi

# 12. ai-squad help output does not mention swarm
MAIN_HELP=$("$CLI" help 2>&1 || true)
if printf '%s' "$MAIN_HELP" | grep -qi 'swarm'; then
  fail "ai-squad help still mentions 'swarm'"
else
  ok "ai-squad help contains no swarm references"
fi

# 13. ai-squad help lists init, labels, and project commands
if printf '%s' "$MAIN_HELP" | grep -q 'init' && \
   printf '%s' "$MAIN_HELP" | grep -q 'labels' && \
   printf '%s' "$MAIN_HELP" | grep -q 'project'; then
  ok "ai-squad help lists init, labels, and project commands"
else
  fail "ai-squad help is missing init, labels, or project"
fi

# 14. ai-squad project command exists and exits non-zero without a repo (no gh auth in CI)
PROJECT_OUT=$("$CLI" project 2>&1 || true)
if printf '%s' "$PROJECT_OUT" | grep -qiE 'repository|owner|project|Could not'; then
  ok "ai-squad project command exists and produces meaningful output"
else
  fail "ai-squad project command missing or silent"
fi

# 15. template includes story template
if [[ -f "$TEMPLATE_DIR/docs/stories/_template.md" ]]; then
  ok "template includes docs/stories/_template.md"
else
  fail "template is missing docs/stories/_template.md"
fi

# 16. template includes DoD document
if [[ -f "$TEMPLATE_DIR/docs/dod/definition-of-done.md" ]]; then
  ok "template includes docs/dod/definition-of-done.md"
else
  fail "template is missing docs/dod/definition-of-done.md"
fi

# 17. init copies story template and dod into new project
if [[ -f "$TEST_PROJECT/docs/stories/_template.md" ]]; then
  ok "init copies docs/stories/_template.md"
else
  fail "init did not copy docs/stories/_template.md"
fi

if [[ -f "$TEST_PROJECT/docs/dod/definition-of-done.md" ]]; then
  ok "init copies docs/dod/definition-of-done.md"
else
  fail "init did not copy docs/dod/definition-of-done.md"
fi

# 18. template .gitignore contains Claude logs entries
GITIGNORE_CONTENT=$(cat "$TEMPLATE_DIR/.gitignore")
if printf '%s' "$GITIGNORE_CONTENT" | grep -q '.claude/logs/' && \
   printf '%s' "$GITIGNORE_CONTENT" | grep -q 'pending-sync.jsonl' && \
   printf '%s' "$GITIGNORE_CONTENT" | grep -q '.claude/state/squad.json'; then
  ok ".gitignore includes .claude/logs/, pending-sync.jsonl, state/squad.json"
else
  fail ".gitignore missing one or more Claude runtime entries"
fi

# 19. label command output covers new groups (dry check via grep on script source)
for group in "type:feature" "type:spike" "priority:high" "priority:medium" \
             "spec:problem" "spec:approved" "design:pending" "design:a11y-passed" \
             "adr:required" "adr:complete" "ui:required"; do
  if grep -q "\"$group\"" "$CLI"; then
    ok "label group present: $group"
  else
    fail "label group missing: $group"
  fi
done

# ─── summary ──────────────────────────────────────────────────────────────────
printf "\n  ────────────────────────────────────────\n"
printf "  \033[1;32m%d passed\033[0m  ·  " "$PASS"
if [[ "$FAIL" -gt 0 ]]; then
  printf "\033[1;31m%d failed\033[0m\n\n" "$FAIL"
  exit 1
else
  printf "\033[2m0 failed\033[0m\n\n"
fi
