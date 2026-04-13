# =============================================================================
# Stage 1 — dev
# Includes hot-reload (air) and all dev tooling.
# Source is volume-mounted at runtime; nothing is compiled at build time.
# =============================================================================
FROM golang:1.26.1-alpine AS dev

WORKDIR /app

RUN apk --no-cache add git curl bash

# Pin tool versions so the dev image stays stable across rebuilds.
ARG AIR_VERSION=v1.52.3
ARG SQLC_VERSION=v1.28.0
ARG MIGRATE_VERSION=v4.18.3
ARG GOLANGCI_LINT_VERSION=v1.64.8

RUN go install github.com/air-verse/air@${AIR_VERSION} && \
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@${SQLC_VERSION} && \
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@${MIGRATE_VERSION} && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}

# Pre-download Go module dependencies as a cached layer.
# This layer is only invalidated when go.mod or go.sum changes.
COPY go.mod go.sum ./
RUN go mod download

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]


# =============================================================================
# Stage 2 — builder
# Compiles a static binary for production.
# =============================================================================
FROM golang:1.26.1-alpine AS builder

WORKDIR /app

RUN apk --no-cache add git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /api \
    ./cmd/api


# =============================================================================
# Stage 3 — prod
# Minimal runtime image — binary only.
# =============================================================================
FROM alpine:3.19 AS prod

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /api /api

EXPOSE 8080

ENTRYPOINT ["/api"]
