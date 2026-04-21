#!/usr/bin/env bash
# tests/run_all.sh — run every test suite and report overall result
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TOTAL_PASS=0
TOTAL_FAIL=0

# ─── helpers ──────────────────────────────────────────────────────────────────
section() { printf "\n\033[1;36m══  %s  ══\033[0m\n" "$1"; }
suite_ok()   { printf "\n  \033[1;32m✓  PASSED\033[0m  %s\n" "$1"; }
suite_fail() { printf "\n  \033[1;31m✗  FAILED\033[0m  %s\n" "$1"; }

run_suite() {
  local label="$1" script="$2"
  section "$label"
  if bash "$script"; then
    suite_ok "$label"
    TOTAL_PASS=$((TOTAL_PASS + 1))
  else
    suite_fail "$label"
    TOTAL_FAIL=$((TOTAL_FAIL + 1))
  fi
}

# ─── suites ───────────────────────────────────────────────────────────────────
printf "\n\033[1m  MAPLE — Test Suite\033[0m\n"
printf "  %s\n" "$REPO_ROOT"

run_suite "CLI Tests"              "$REPO_ROOT/tests/cli/test_ai_squad.sh"
run_suite "Template Structure"     "$REPO_ROOT/tests/template/test_structure.sh"
run_suite "Agent Frontmatter"      "$REPO_ROOT/tests/template/test_agents.sh"
run_suite "Skills"                 "$REPO_ROOT/tests/template/test_skills.sh"
run_suite "Commands"               "$REPO_ROOT/tests/template/test_commands.sh"

# ─── summary ──────────────────────────────────────────────────────────────────
printf "\n\033[1m  ══════════════════════════════════════\033[0m\n"
printf "  Suites:  \033[1;32m%d passed\033[0m" "$TOTAL_PASS"
if [[ "$TOTAL_FAIL" -gt 0 ]]; then
  printf "  ·  \033[1;31m%d failed\033[0m\n\n" "$TOTAL_FAIL"
  exit 1
else
  printf "  ·  \033[2m0 failed\033[0m\n\n"
fi
