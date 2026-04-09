# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Personal finance management REST API built with Go. Manages installment plans, monthly budgets, and person-to-person debts. Multi-tenant via JWT auth with user-scoped data access.

## Tech Stack

- **Go 1.25** with Gin web framework
- **MySQL 8.0** via GORM (auto-migration on startup)
- **JWT** (72h expiry) for auth, bcrypt for passwords
- **Docker Compose** for deployment (API + MySQL + Cloudflare Tunnel)

## Build & Run

```bash
# Local development
go run .                    # Starts on :8080 (reads .env)
go build -o server .        # Build binary

# Docker (production)
docker compose up -d --build

# Dependencies
go mod tidy
```

No tests exist in the codebase currently.

## Environment Variables

Configured via `.env` file (loaded by godotenv) or environment:
- `PORT` (default: 8080), `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `JWT_SECRET` - signing key for JWT tokens
- `STATIC_DIR` - path to frontend build for SPA serving (default: `./dist`)
- `GIN_MODE` - set to `release` in production

## Architecture

All code is in `package main` with a flat file structure:

| File | Purpose |
|------|---------|
| `main.go` | Server bootstrap, CORS config, SPA static file serving |
| `config.go` | Environment variable loading into `AppConfig` global |
| `database.go` | MySQL connection + GORM auto-migration, exposes `DB` global |
| `auth.go` | Register/Login handlers, JWT middleware (`AuthMiddleware()`) |
| `models.go` | All GORM models (User, InstallmentPlan, Installment, BudgetItem, BudgetMonthlyValue, PersonDebt, DebtPayment) |
| `routes.go` | Route registration under `/api` prefix |
| `handler_installment.go` | CRUD + toggle payment status for installment plans |
| `handler_budget.go` | CRUD for budget items with per-month values (Thai month abbreviations) |
| `handler_debt.go` | CRUD for person debts + payment recording with transaction |

## Key Patterns

- **User scoping**: All queries filter by `user_id` from JWT context (`c.GetString("user_id")`)
- **UUID primary keys**: All entities use `uuid.New().String()` for IDs
- **Thai Buddhist Era dates**: Debt dates use DD/MM/YYYY+543 format via `todayBE()`
- **Thai month names**: Budget months use Thai abbreviations (`ม.ค.`, `ก.พ.`, etc.)
- **Cascade deletes**: Child records (Installments, MonthlyValues, Payments) cascade on parent delete via GORM constraints
- **Budget year filtering**: Budget endpoints accept `?year=` query param to filter monthly values

## API Routes

All protected routes require `Authorization: Bearer <token>` header.

- `POST /api/auth/register`, `POST /api/auth/login` (public)
- `/api/installments` - CRUD + `PATCH /:planId/toggle/:installmentId`
- `/api/budget` - CRUD + `PATCH /:id/month` for individual month value updates
- `/api/debts` - CRUD + `POST /:id/payment` for recording payments

## Deployment

Self-hosted runner deploys on push to `main` via GitHub Actions. The workflow pulls both this repo and the frontend (`money-manage`), builds the frontend, copies `dist/` here, then runs `docker compose up -d --build`. Exposed via Cloudflare Tunnel.
