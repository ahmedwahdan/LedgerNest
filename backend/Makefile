# =============================================================================
# Expense Intelligence API — Makefile
#
# All commands run inside the Docker dev container.
# Prerequisites: Docker + Docker Compose (no local Go required).
#
# Usage:
#   make <target>
#   make help        — list all targets
# =============================================================================

COMPOSE := docker compose
# Use 'exec' when the container is already running (fastest path).
EXEC    := $(COMPOSE) exec api
# Use 'run --rm' to spin up a one-off container (works without 'make dev' running).
RUN     := $(COMPOSE) run --rm api

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

# ── Docker ────────────────────────────────────────────────────────────────────

.PHONY: image
image: ## Build / rebuild Docker images
	$(COMPOSE) build

.PHONY: up
up: ## Start all services detached
	$(COMPOSE) up -d

.PHONY: down
down: ## Stop all services
	$(COMPOSE) down

.PHONY: logs
logs: ## Tail api service logs
	$(COMPOSE) logs -f api

# ── Development ───────────────────────────────────────────────────────────────

.PHONY: dev
dev: ## Start all services with hot reload (foreground)
	$(COMPOSE) up

.PHONY: shell
shell: ## Open a shell inside the running api container
	$(EXEC) sh

# ── Build & Lint ──────────────────────────────────────────────────────────────

.PHONY: compile
compile: ## Compile the Go binary inside the container
	$(RUN) go build ./...

.PHONY: lint
lint: ## Run golangci-lint inside the container
	$(RUN) golangci-lint run ./...

.PHONY: vet
vet: ## Run go vet
	$(RUN) go vet ./...

# ── Tests ─────────────────────────────────────────────────────────────────────

.PHONY: test
test: ## Run all tests (unit + integration)
	$(RUN) go test -race -count=1 ./...

.PHONY: test-unit
test-unit: ## Run unit tests only (no DB, fast)
	$(RUN) go test -race -count=1 -short ./...

.PHONY: test-integration
test-integration: ## Run integration tests only
	$(RUN) go test -race -count=1 -run Integration ./test/integration/...

.PHONY: test-cover
test-cover: ## Run all tests and produce a coverage report (coverage.html)
	$(RUN) sh -c 'go test -race -count=1 -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html'
	@echo "Coverage report: backend/coverage.html"

# ── Database & Migrations ─────────────────────────────────────────────────────

.PHONY: migrate-up
migrate-up: ## Apply all pending migrations
	$(RUN) sh -c 'migrate -path db/migrations -database "$$DATABASE_URL" up'

.PHONY: migrate-down
migrate-down: ## Roll back the last applied migration
	$(RUN) sh -c 'migrate -path db/migrations -database "$$DATABASE_URL" down 1'

.PHONY: migrate-new
migrate-new: ## Create a new migration pair  e.g. make migrate-new NAME=create_users
	@test -n "$(NAME)" || (echo "Usage: make migrate-new NAME=<snake_case_name>"; exit 1)
	$(RUN) migrate create -ext sql -dir db/migrations -seq $(NAME)

.PHONY: migrate-status
migrate-status: ## Show current migration version
	$(RUN) sh -c 'migrate -path db/migrations -database "$$DATABASE_URL" version'

.PHONY: db-reset
db-reset: ## Drop all tables and reapply migrations (dev only)
	@echo "Resetting dev database…"
	$(RUN) sh -c 'migrate -path db/migrations -database "$$DATABASE_URL" drop -f'
	$(MAKE) migrate-up
	@echo "Done."

.PHONY: seed
seed: ## Run the dev seed script
	$(RUN) go run ./scripts/seed/main.go

# ── Code Generation ───────────────────────────────────────────────────────────

.PHONY: sqlc-gen
sqlc-gen: ## Generate type-safe DB code from SQL queries (sqlc)
	$(RUN) sqlc generate

# ── Modules ───────────────────────────────────────────────────────────────────

.PHONY: mod-tidy
mod-tidy: ## Tidy go.mod and regenerate go.sum
	$(RUN) go mod tidy

.PHONY: mod-download
mod-download: ## Download all modules declared in go.mod
	$(RUN) go mod download

# ── Smoke Test ────────────────────────────────────────────────────────────────

.PHONY: smoke
smoke: ## Run smoke tests against the running API (requires curl on host)
	@bash scripts/smoke.sh
