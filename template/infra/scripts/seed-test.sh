#!/usr/bin/env bash
# seed-test.sh — populate the test database with fixtures before BDD/integration runs.
# Customize this script for your project's data model.
set -euo pipefail

echo "Seeding test database..."

# Example: run migrations and load fixtures
# psql "$DATABASE_URL" < infra/scripts/fixtures.sql

echo "Seed complete."
