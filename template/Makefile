# Makefile — Unified build/test contract for all agents
.PHONY: build test test-integration test-e2e test-contract test-all lint security-scan fmt containers-up containers-down seed-test migrate help

## Build
build:
	@echo "Override this target with your project's build command"
	@echo "Examples:"
	@echo "  dotnet build"
	@echo "  mvn package"
	@echo "  npm run build"
	@echo "  cargo build --release"

## Unit tests
test:
	@echo "Override this target with your project's test command"
	@echo "Examples:"
	@echo "  dotnet test --filter 'Category=Unit'"
	@echo "  mvn test -Dgroups=unit"
	@echo "  npx vitest run tests/unit"
	@echo "  pytest tests/unit"

## Integration tests (requires containers)
test-integration:
	@echo "Override this target with your project's integration test command"
	@echo "Examples:"
	@echo "  dotnet test --filter 'Category=Integration'"
	@echo "  mvn test -Dgroups=integration"
	@echo "  npx vitest run tests/integration"
	@echo "  pytest tests/integration"

## E2E tests
test-e2e:
	npx playwright test tests/e2e/ --reporter=html

## Contract tests
test-contract:
	@echo "Override this target with your project's contract test command"
	@echo "Example: npx playwright test tests/contract/"

## Run all tests (gate for Phase 8)
test-all: test test-integration test-e2e test-contract
	@echo "All test suites passed."

## Linting
lint:
	@echo "Override this target with your project's lint command"
	@echo "Examples:"
	@echo "  dotnet format --verify-no-changes"
	@echo "  npx eslint . --ext .ts,.tsx"
	@echo "  ruff check ."

## Security scanning
security-scan:
	@echo "Override this target with your project's security scan command"
	@echo "Examples:"
	@echo "  dotnet-sonarscanner"
	@echo "  npm audit --audit-level=high"
	@echo "  trivy fs ."

## Format code
fmt:
	@echo "Override this target with your project's format command"
	@echo "Examples:"
	@echo "  dotnet format"
	@echo "  npx prettier --write ."
	@echo "  ruff format ."

## Start test containers
containers-up:
	docker compose -f docker-compose.test.yml up -d --wait

## Stop test containers
containers-down:
	docker compose -f docker-compose.test.yml down -v

## Seed test data
seed-test:
	./scripts/seed-test.sh

## Run database migrations
migrate:
	@echo "Override this target with your project's migration command"
	@echo "Examples:"
	@echo "  dotnet ef database update"
	@echo "  mvn flyway:migrate"
	@echo "  npx prisma migrate deploy"
	@echo "  supabase db push"

## Show available targets
help:
	@echo "Available make targets:"
	@echo "  build           - Build the project"
	@echo "  test            - Run unit tests"
	@echo "  test-integration - Run integration tests (requires containers)"
	@echo "  test-e2e        - Run E2E tests with Playwright"
	@echo "  test-contract   - Run contract/schema tests"
	@echo "  test-all        - Run all test suites (Phase 8 gate)"
	@echo "  lint            - Run linter"
	@echo "  security-scan   - Run security scanning"
	@echo "  fmt             - Format code"
	@echo "  containers-up   - Start test containers"
	@echo "  containers-down - Stop and remove test containers"
	@echo "  seed-test       - Seed test data"
	@echo "  migrate         - Run database migrations"
