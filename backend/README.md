# LedgerNest API

Go REST API backend for LedgerNest.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) + Docker Compose  
  No local Go, sqlc, or migrate installation required.

## Setup

```bash
# 1. Copy and adjust env vars (defaults work out of the box with compose)
cp .env.example .env

# 2. Build the dev Docker image
make image

# 3. Start all services (postgres, redis, api with hot reload)
make dev
```

The API will be available at `http://localhost:8080`.

## Common Commands

| Command | Description |
|---------|-------------|
| `make dev` | Start all services with hot reload |
| `make logs` | Tail the API log stream |
| `make shell` | Open a shell inside the running container |
| `make compile` | Compile the binary (smoke-check for build errors) |
| `make test` | Run all tests |
| `make test-unit` | Run unit tests only (fast, no DB) |
| `make lint` | Run golangci-lint |
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Roll back the last migration |
| `make migrate-new NAME=x` | Create a new migration pair |
| `make db-reset` | Wipe dev DB and reapply all migrations |
| `make sqlc-gen` | Regenerate type-safe DB code from SQL queries |
| `make mod-tidy` | Tidy go.mod / go.sum |
| `make smoke` | Run smoke tests against the running API |

## Project Structure

```
cmd/api/            entry point
internal/
  auth/             JWT signing, password hashing
  config/           typed config loader
  db/               pgxpool, transaction helpers
  handler/          HTTP handlers (grouped by domain)
  httpx/            response/error helpers
  middleware/       auth, CORS, request-id, rate-limit, logging
  model/            domain types
  repository/       sqlc-generated DB code + hand-written repos
  service/          business logic per domain
  validator/        request DTOs + validation
  worker/           cron jobs
  mail/             email interface + dev stub
db/
  migrations/       *.up.sql / *.down.sql
  queries/          sqlc query definitions
scripts/            dev utilities (seed, smoke)
test/integration/   end-to-end integration tests
```

## Environment Variables

See [`.env.example`](.env.example) for the full list with descriptions.

## Tech Stack

| Concern | Choice |
|---------|--------|
| Language | Go 1.22 |
| Router | Chi v5 |
| Database | PostgreSQL 16 |
| DB driver | pgx/v5 |
| SQL codegen | sqlc |
| Migrations | golang-migrate |
| Auth | JWT (golang-jwt/jwt v5) |
| Hot reload | air |
| Lint | golangci-lint |
