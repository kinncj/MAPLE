# AGENTS.md — Multi-Agent Squad Roster

## Quick Reference

| # | Agent | Role |
|---|---|---|
| 1 | orchestrator | Pipeline control (never codes) |
| 2 | product-owner | User stories, acceptance criteria |
| 3 | architect | ADR, contracts, threat models |
| 4 | qa | Write tests (RED) + validate (GREEN) |
| 4b | qa-cucumber | BDD: extract Gherkin → feature files, generate step stubs, run suite |
| 5 | dotnet | .NET backend implementation |
| 6 | javascript | Node.js / vanilla JS |
| 7 | typescript | TypeScript backend/libraries |
| 8 | react-vite | React + Vite + TypeScript SPA |
| 9 | nextjs | Next.js full-stack |
| 10 | java | Java backend (non-Spring) |
| 11 | springboot | Spring Boot applications |
| 12 | kubernetes | K8s manifests, Kustomize, Helm |
| 13 | terraform | Terraform IaC |
| 14 | docker | Dockerfiles, Compose |
| 15 | postgresql | Schema, migrations, RLS |
| 16 | redis | Caching, pub/sub, streams |
| 17 | supabase | Auth, RLS, Edge Functions |
| 18 | vercel | Deployment, edge, config |
| 19 | stripe | Payments, webhooks, billing |
| 20 | data-science | EDA, stats, visualization |
| 21 | data-engineer | Pipelines, ETL, orchestration |
| 22 | tensorflow | TF/Keras models, training |
| 23 | pytorch | PyTorch models, training |
| 24 | pandas-numpy | Data manipulation, arrays |
| 25 | scikit | Classical ML, pipelines |
| 26 | jupyter | Notebooks, papermill |
| 27 | docs | Feature docs, CHANGELOG, Mermaid |

## Pipeline Phases
1. DISCOVER → 2. ARCHITECT → 3. PLAN → 4. INFRA → 5. IMPLEMENT → 6. VALIDATE → 7. DOCUMENT → 8. FINAL GATE

## Makefile Contract
All agents use: `make build`, `make test`, `make test-integration`, `make test-e2e`,
`make test-contract`, `make test-all`, `make lint`, `make security-scan`, `make fmt`,
`make containers-up`, `make containers-down`, `make seed-test`, `make migrate`.

## Git Conventions
- Conventional Commits: `feat:`, `fix:`, `test:`, `docs:`, `infra:`, `refactor:`
- Branch naming: `feat/{slug}`, `fix/{slug}`
- Squash merge to main
