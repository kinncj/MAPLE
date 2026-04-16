#!/usr/bin/env bash
# seed-test.sh — Seed the test database with fixture data
# Usage: ./infra/scripts/seed-test.sh
# Called by: make seed-test
#
# Copyright (C) 2025 Kinn Coelho Juliao <kinncj@protonmail.com>
# SPDX-License-Identifier: AGPL-3.0-or-later
set -euo pipefail

# ─── Colour palette ───────────────────────────────────────────────────────────
if [[ -n "${NO_COLOR:-}" || "${TERM:-}" == "dumb" || ! -t 1 ]]; then
  R=''; B=''; D=''
  BGRN=''; BRED=''; BCYN=''; BMGT=''
else
  R='\033[0m';    B='\033[1m';    D='\033[2m'
  BGRN='\033[1;32m'; BRED='\033[1;31m'
  BCYN='\033[1;36m'; BMGT='\033[1;35m'
fi

HR="  ${D}$(printf '─%.0s' {1..60})${R}"

# ─── UI primitives ────────────────────────────────────────────────────────────
header() {
  printf "\n"
  printf "  ${B}${BMGT}AI Development Squad${R}  ${D}·${R}  ${B}%s${R}\n" "$1"
  printf "%b\n\n" "$HR"
}

step()  { printf "\n  ${BCYN}›${R}  ${B}%s${R}\n" "$1"; }
ok()    { printf "  ${BGRN}✓${R}  %-40s  ${D}%s${R}\n" "$1" "${2:-}"; }
skip()  { printf "  ${D}–  %-40s  not found, skipping${R}\n" "$1"; }
info()  { printf "  ${D}  %-20s${R}  ${B}%s${R}\n" "$1" "${2:-}"; }
fail()  { printf "\n  ${BRED}✗${R}  ${B}%s${R}\n\n" "$*" >&2; exit 1; }

# ─── Config ───────────────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

DB_HOST="${POSTGRES_HOST:-localhost}"
DB_PORT="${POSTGRES_PORT:-5433}"
DB_USER="${POSTGRES_USER:-testuser}"
DB_PASSWORD="${POSTGRES_PASSWORD:-testpassword}"
DB_NAME="${POSTGRES_DB:-testdb}"

export PGPASSWORD="$DB_PASSWORD"

# ─── Main ─────────────────────────────────────────────────────────────────────
header "Seed Test Database"

info "Host"     "${DB_HOST}:${DB_PORT}"
info "Database" "$DB_NAME"
info "User"     "$DB_USER"

# Wait for the database to be ready
step "Connecting to database"
for i in $(seq 1 30); do
  if pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" &>/dev/null; then
    ok "Database ready"
    break
  fi
  if [[ "$i" -eq 30 ]]; then
    fail "Database not ready after 30 seconds  (${DB_HOST}:${DB_PORT})"
  fi
  printf "  ${D}  Waiting... %d/30${R}\r" "$i"
  sleep 1
done

# SQL helper
run_sql() {
  psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
       -f "$1" --quiet 2>&1 || fail "SQL error in: $1"
}

# Migrations
MIGRATION_DIR="$PROJECT_DIR/migrations"
step "Running migrations"
if compgen -G "${MIGRATION_DIR}/*.sql" &>/dev/null; then
  for f in "$MIGRATION_DIR"/*.sql; do
    run_sql "$f"
    ok "$(basename "$f")"
  done
else
  skip "$MIGRATION_DIR/*.sql"
fi

# Contract seed data
step "Applying contract seed data"
SEED_FILE="$PROJECT_DIR/docs/specs/current/contracts/seed-data.sql"
if [[ -f "$SEED_FILE" ]]; then
  run_sql "$SEED_FILE"
  ok "$(basename "$SEED_FILE")"
else
  skip "docs/specs/current/contracts/seed-data.sql"
fi

# Test fixtures
FIXTURES_DIR="$PROJECT_DIR/tests/fixtures"
step "Applying test fixtures"
if compgen -G "${FIXTURES_DIR}/*.sql" &>/dev/null; then
  for f in "$FIXTURES_DIR"/*.sql; do
    [[ -f "$f" ]] || continue
    run_sql "$f"
    ok "$(basename "$f")"
  done
else
  skip "tests/fixtures/*.sql"
fi

# Summary
printf "\n%b\n" "$HR"
printf '%s\n\n' "  ${BGRN}✓${R}  ${B}Test database seeded.${R}"
