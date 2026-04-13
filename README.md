# Expense Intelligence App

> A modern household finance platform in progress — built to make budgeting clearer, collaborative, and more actionable.

[![CI](https://github.com/wahdan/expenses-project/actions/workflows/ci.yml/badge.svg)](https://github.com/wahdan/expenses-project/actions/workflows/ci.yml)

## Overview

Expense Intelligence App is a rising personal finance project focused on households, shared visibility, and smarter budgeting. The goal is simple: make it easier for people to track spending, stay aligned with family members, and understand what is driving their budget in real terms.

It combines shared expense management, flexible budgeting, and insight-driven reporting in one product direction that can grow from an MVP into a broader financial intelligence platform.

## Core Product Direction

- **Shared household finance** — Multi-user access with clear roles and personal or shared visibility
- **Smarter budgets** — Category budgets, custom pay-cycle timing, and clearer budget health signals
- **Reliable expense tracking** — Manual capture, edit history, and recovery-friendly workflows
- **Actionable insights** — Trends, category analysis, merchant-level patterns, and exportable reporting
- **Built for growth** — A foundation for future AI-assisted analysis, imports, and broader financial workflows

## Foundation

| Layer | Choice |
|-------|--------|
| Backend | Go 1.22, Chi router, PostgreSQL 16, sqlc, golang-migrate |
| Auth | JWT + refresh tokens |
| Frontend *(planned)* | Next.js, TypeScript, Tailwind CSS |
| Environment | Docker Compose |

## Quick Start

**Prerequisites:** Docker + Docker Compose

```bash
# 1. Clone
git clone https://github.com/wahdan/expenses-project.git
cd expenses-project

# 2. Copy env (defaults work out of the box)
cp backend/.env.example backend/.env

# 3. Build and start
docker compose build
docker compose up
```

The API runs at **`http://localhost:8080`**.

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

## Current State

The project is actively taking shape around a strong MVP for household expense management. The backend foundation is in place, the product scope is clearly defined, and the next steps focus on expanding the user experience and intelligence layer.

For implementation details and contributor workflows, see [`backend/README.md`](backend/README.md).

## Project Structure

```
backend/    Go REST API
frontend/   Next.js web app (planned)
project.md  Product scope and architecture direction
```

## Roadmap

This repository tracks the early growth of the product. The near-term focus is a polished MVP, with longer-term expansion into features like imports, deeper analytics, AI-assisted insights, and broader platform experiences documented in [`project.md`](project.md).
