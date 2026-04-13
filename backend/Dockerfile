# =============================================================================
# Stage 1 — dev
# Includes hot-reload (air) and all dev tooling.
# Source is volume-mounted at runtime; nothing is compiled at build time.
# =============================================================================
FROM golang:1.22-alpine AS dev

WORKDIR /app

RUN apk --no-cache add git curl bash

# Install dev tools (pinned to avoid breaking changes)
RUN go install github.com/air-verse/air@latest && \
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest && \
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

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
FROM golang:1.22-alpine AS builder

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
