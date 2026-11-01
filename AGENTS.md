# AGENTS.md

## 1. Stack

- Frontend: Next.js, React, TypeScript
- Frontend package manager: pnpm
- Backend: Go
- Backend dependency management: Go Modules
- Database: PostgreSQL
- Local infrastructure: Docker Compose

Follow the versions defined in project configuration and lockfiles.

## 2. Commands

### Frontend

- Development: `pnpm dev`
- Build: `pnpm build`
- Tests: `pnpm test`
- Lint: `pnpm lint`
- Type check: `pnpm typecheck`

### Backend

- Development: `go run ./cmd/api`
- Tests: `go test ./...`
- Build: `go build ./...`
- Format: `gofmt`

### Quality gate

- Run everything: `make check` (lint, format check, type check, tests)
- Individual checks: `make lint`, `make format-check`, `make typecheck`, `make test`

CI runs exactly `make check`. Run it before every push; a red check is a
refused merge.

### Infrastructure

- Start: `docker compose up -d`
- Stop: `docker compose down`
- Logs: `docker compose logs -f`

Use existing project scripts and commands. Do not invent commands that are not configured.

## 3. Conventions

- Follow the existing project structure and coding style.
- Keep changes small and focused.
- Use TypeScript for frontend code; avoid `any` unless justified.
- Follow standard Go conventions and use `gofmt`.
- Validate input at system boundaries.
- Handle errors explicitly.
- Write or update tests when changing behavior.
- Keep frontend, backend, and sandbox responsibilities separate.

### Agent permissions

**Tier 1 — Allowed**
- Read and inspect project files.
- Create and modify source code.
- Run local tests, lint, formatting, and type checks.
- Read project documentation.

**Tier 2 — Ask first**
- Install, remove, or upgrade dependencies.
- Modify database schemas or migrations.
- Delete or move existing files.
- Change CI/CD or infrastructure configuration.
- Execute commands that modify shared resources.

**Tier 3 — Never**
- Access production systems or databases.
- Expose the Docker socket to student submission containers.
- Bypass sandbox security restrictions.
- Commit or push changes without explicit authorization.

## 4. Never

- Never expose secrets, credentials, or model API keys to the submission sandbox.
- Never make real LLM API calls in unit tests.
- Never modify tests just to make implementation pass — fix the code, or change
  the contract first (AI-LOG.md 2026-10-31 #2 and #4).
- Never modify evaluation data or labels to make evaluation pass.
- Never weaken sandbox restrictions or CI checks to obtain a passing result.
- Never commit `.env` files or secrets to the repository.
- Never reorder the error-precedence list in `docs/contract/` in code alone —
  contract, implementation, and tests change together (AI-LOG.md 2026-10-31 #5).
