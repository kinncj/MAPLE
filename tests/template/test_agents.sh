#!/usr/bin/env bash
# tests/template/test_agents.sh — agent frontmatter validation
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TEMPLATE="$REPO_ROOT/template"
PASS=0
FAIL=0

# ─── helpers ──────────────────────────────────────────────────────────────────
ok()   { printf "  \033[1;32m✓\033[0m  %s\n" "$1"; PASS=$((PASS + 1)); }
fail() { printf "  \033[1;31m✗\033[0m  %s\n" "$1"; FAIL=$((FAIL + 1)); }

# Extract a frontmatter field value from a markdown file
# Usage: fm_field <file> <field>
fm_field() {
  local file="$1" field="$2"
  awk "/^---$/{if(found) exit; found=1; next} found && /^${field}:/{print; exit}" "$file" \
    | sed "s/^${field}:[[:space:]]*//"
}

# Check frontmatter has the field with a non-empty value
assert_fm() {
  local file="$1" field="$2"
  local val
  val=$(fm_field "$file" "$field")
  if [[ -n "$val" ]]; then
    ok "$(basename "$file"): $field = $val"
  else
    fail "$(basename "$file"): missing frontmatter '$field'"
  fi
}

# ─── Claude Code agents ───────────────────────────────────────────────────────
printf "\n\033[1m  Claude Code Agent Frontmatter\033[0m\n\n"

for f in "$TEMPLATE/.claude/agents/"*.md; do
  [[ -f "$f" ]] || continue
  assert_fm "$f" "name"
  assert_fm "$f" "description"
done

# ─── OpenCode agents ──────────────────────────────────────────────────────────
printf "\n\033[1m  OpenCode Agent Frontmatter\033[0m\n\n"

for f in "$TEMPLATE/.opencode/agents/"*.md; do
  [[ -f "$f" ]] || continue
  assert_fm "$f" "description"
done

# ─── No stale date-suffixed model IDs ─────────────────────────────────────────
printf "\n\033[1m  Stale Model ID Check\033[0m\n\n"

STALE=$(grep -rE "claude-.*-2[0-9]{7}" \
  "$TEMPLATE/.claude/agents/" \
  "$TEMPLATE/.opencode/agents/" 2>/dev/null || true)

if [[ -z "$STALE" ]]; then
  ok "No stale date-suffixed model IDs found"
else
  printf "%s\n" "$STALE"
  fail "Stale model IDs found (remove date suffixes)"
fi

# ─── summary ──────────────────────────────────────────────────────────────────
printf "\n  ────────────────────────────────────────\n"
printf "  \033[1;32m%d passed\033[0m  ·  " "$PASS"
if [[ "$FAIL" -gt 0 ]]; then
  printf "\033[1;31m%d failed\033[0m\n\n" "$FAIL"
  exit 1
else
  printf "\033[2m0 failed\033[0m\n\n"
fi
