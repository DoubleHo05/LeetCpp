# LeetCpp

> A grading-support platform for introductory C++ programming assignments.

## 1. Overview

LeetCpp helps teaching assistants (TAs) review and grade student submissions for programming assignments.

For an introductory C++ Sorting lab, TAs need to verify that students follow the assignment's constraints, such as implementing merge sort rather than using `std::sort` or another prohibited algorithm.

Manually reviewing submissions is time-consuming and repetitive. LeetCpp aims to streamline this process through automated compilation, test execution, constraint checking, and human review.

### Key features

- **Submission management:** Students submit their C++ solutions and receive compilation and test results.
- **Automated constraint checking:** Evaluate submissions against TA-defined assignment constraints using hard rules and LLM-assisted semantic analysis.
- **TA review workflow:** TAs review violations and uncertain results before publishing marks.
- **Appeals:** Students can appeal published results, and TAs can accept or reject appeals with a reason.
- **Evaluation and quality assurance:** Automated evaluation ensures that the constraint checker meets defined accuracy requirements.

## 2. Technology Stack

| Component | Technology |
|---|---|
| Frontend | Next.js, React, TypeScript |
| Frontend package manager | pnpm |
| Backend | Go |
| Backend dependency management | Go Modules |
| Database | PostgreSQL |
| Local development | Docker Compose |
| Submission sandbox | Isolated Docker containers |
| CI/CD | GitHub Actions |

Technology versions and additional dependencies are defined in their respective project configuration files.

## 3. Architecture

LeetCpp separates the frontend, backend, and submission execution environment.

| Component | Responsibility |
|---|---|
| Frontend | User interface for students, TAs, and administrators |
| Backend API | Authentication, authorization, business logic, and application APIs |
| Database | Persistent storage for users, classes, labs, submissions, reviews, and appeals |
| Worker | Processes submission jobs and coordinates execution |
| Sandbox | Compiles and executes untrusted student code in an isolated environment |

The submission execution environment must be isolated from the application infrastructure. Student code must not have access to application secrets, host resources, or external networks.

## 4. Getting Started

### Prerequisites

Install the following tools:

- Git
- Node.js (version specified in project configuration)
- pnpm
- Go (version specified in `go.mod`)
- Docker
- Docker Compose

### Clone the repository

```bash
git clone https://github.com/DoubleHo05/LeetCpp.git
cd LeetCpp
```

### Configure environment variables

Create your local environment file:

```bash
cp .env.example .env
```

Configure the required values in `.env`.

Never commit `.env` or expose its secrets.

### Install dependencies

Frontend:

```bash
cd frontend
pnpm install
```

Backend:

```bash
cd backend
go mod download
```

### Start infrastructure

From the repository root:

```bash
docker compose up -d
```

### Run the application

Frontend:

```bash
cd frontend
pnpm dev
```

Backend:

```bash
cd backend
go run ./cmd/api
```

Refer to the project configuration for the actual service addresses and ports.

## 5. Development and Quality Checks

Every check runs with one documented command. The commands are identical on
your machine and in CI — if `make check` is green locally, CI is green.

| Check | Command | What it runs |
|---|---|---|
| Lint | `make lint` | `go vet ./...` (backend), `pnpm lint` (frontend) |
| Format check | `make format-check` | `gofmt -l` (backend), `pnpm format:check` (frontend) |
| Type check | `make typecheck` | `go build ./...` (backend), `pnpm typecheck` (frontend) |
| Tests | `make test` | `go test ./...` (backend), `pnpm test` (frontend) |

**Unified quality-check command:** `make check` (from the repository root).
It runs lint, format check, type check, and tests, in that order. Run it
before every push. CI runs exactly `make check` on every pull request.

A fresh clone is green after the Getting Started steps above:

```bash
make check
```

## 6. Environment Variables and Secrets

All required environment variables must be documented in `.env.example` without actual secret values.

- Store local credentials in `.env`.
- Never commit `.env` or other files containing secrets.
- Validate required configuration when the application starts.
- Never expose model API keys or application secrets to the submission sandbox.

## 7. AI Model Usage

The LLM is used for semantic constraint checking when deterministic rules cannot establish whether a submission satisfies a constraint.

The model returns:

- `pass`
- `violation`
- `unsure`

Each result must include supporting code evidence where applicable. The model does not determine deduction amounts.

### Usage and spending limit

- Provider: To be determined.
- Model: To be determined.
- Spending cap: **USD 15 per calendar month**, set as a hard limit in the
  provider's billing console before the key is ever issued, and mirrored in
  `LLM_SPENDING_CAP_USD`. The key is never used without the cap in place.
  See `AI-LOG.md` for the dated decision record.

Unit tests must not make real LLM API calls. Use mocks or deterministic test responses instead.

## 8. Continuous Integration

The workflow in `.github/workflows/ci.yml` runs `make check` on every pull
request and on every push to `main`. A red check blocks the merge.

To make CI actually refuse a bad merge, enable branch protection on `main`
(GitHub → Settings → Branches → Branch protection rule):

1. Require a pull request before merging.
2. Require status checks to pass before merging; required check: **`checks`**.
3. Do not allow bypassing the above settings for administrators.

Nothing is pushed straight to `main` once protection is on.

Evidence of the gate biting:

- **Red check that stopped a merge:** [PR #1](https://github.com/DoubleHo05/LeetCpp/pull/1)
  — commit `b0bb518` went red on the `checks` status
  ([run 41311559092](https://github.com/DoubleHo05/LeetCpp/actions/runs/41311559092),
  `failure`, 2026-11-01 13:31 UTC). The PR stayed unmerged while it was red.
- **Commit that turned it green:**
  [`bb94186`](https://github.com/DoubleHo05/LeetCpp/pull/1/changes/bb94186a7064fd8306a55b804136801018de5fc5)
  — [run 41312466512](https://github.com/DoubleHo05/LeetCpp/actions/runs/41312466512)
  passed `make check` (`success`, 2026-11-01 13:42 UTC). PR #1 merged at
  14:00 UTC, after the green run.

Both PR #1 commits ran the same `checks` job. Nothing has been pushed straight
to `main` since branch protection was enabled; the only commits on `main`
after the initial commit arrived through PR #1.

## 9. Project Documentation

| Document | Description |
|---|---|
| `AGENTS.md` | Instructions, conventions, and permissions for AI coding agents |
| `AI-LOG.md` | Record of AI-assisted development decisions and incidents |
| `docs/contract/post-labs-labid-submissions.md` | Contract for `POST /labs/{labId}/submissions` |
| `docs/contract/` | API endpoint contracts |
| `docs/` | Architecture and technical documentation |
