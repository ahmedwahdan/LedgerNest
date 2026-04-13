# Expense Intelligence App

A household expense tracking and budgeting application.

## Structure

| Directory | Contents |
|-----------|----------|
| `/backend` | Go REST API |
| `/frontend` | Next.js web app *(coming later)* |
| `project.md` | Full product spec and architecture |

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) + Docker Compose  
  *(No local Go or Node required — everything runs in containers)*

## Quick Start

```bash
# 1. Copy env file (adjust values if needed)
cp backend/.env.example backend/.env

# 2. Build images and start services
docker compose build
docker compose up
```

API is available at `http://localhost:8080/health`.

See [`backend/README.md`](backend/README.md) for the full developer guide.
