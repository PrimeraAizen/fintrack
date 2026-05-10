# FinTrack

Personal finance tracking backend built with Go.

## Stack

- **Go 1.23** — Gin, pgx, go-redis, golang-jwt, shopspring/decimal
- **PostgreSQL 16** — primary store, goose migrations
- **Redis 7** — JWT blacklisting, exchange rate cache

## Project Layout

```
cmd/
  web/        # HTTP server entry point
  seed/       # DB seeder
config/       # config.yaml + config.example.yaml
internal/
  app/        # dependency wiring
  delivery/   # HTTP layer (handlers, middleware, DTOs)
  domain/     # domain structs + error sentinels
  repository/ # pgx queries
  service/    # business logic
migrations/postgres/
pkg/
  adapter/    # postgres + redis clients
  logger/
  response/   # success/error envelope helpers
```

## Getting Started

### Prerequisites

- Go 1.23+
- Docker + Docker Compose
- [goose](https://github.com/pressly/goose) — `go install github.com/pressly/goose/v3/cmd/goose@latest`

### Run with Docker

```bash
docker compose up -d
```

The API will be available at `http://localhost:8080`.

### Run locally

1. Copy and edit the config:

```bash
cp config/config.example.yaml config/config.yaml
```

2. Start dependencies:

```bash
make docker-up
```

3. Run migrations:

```bash
make migrate-up
```

4. (Optional) Seed the database:

```bash
make seed
```

5. Start the server:

```bash
make run
```

## Configuration

Config is loaded from `config/config.yaml`. Every field can be overridden with an environment variable using the `APP_` prefix and `_` as separator, e.g.:

| Env var | Config key |
|---|---|
| `APP_HTTP_PORT` | `http.port` |
| `APP_DATABASE_HOST` | `database.host` |
| `APP_DATABASE_PASSWORD` | `database.password` |
| `APP_REDIS_ADDR` | `redis.addr` |
| `APP_JWT_SECRET` | `jwt.secret` |
| `APP_CURRENCY_API_URL` | `currency.api_url` |

Key defaults (see `config/config.example.yaml`):

| Setting | Default |
|---|---|
| HTTP port | `8080` |
| JWT access TTL | `15m` |
| JWT refresh TTL | `168h` (7 days) |
| Exchange rate cache TTL | `1h` |

## API

All endpoints are prefixed with `/api/v1`.

### Auth

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/auth/register` | — | Register new user |
| POST | `/auth/login` | — | Login, returns user + tokens |
| POST | `/auth/refresh` | — | Rotate refresh token |
| POST | `/auth/logout` | Bearer | Revoke tokens |
| GET | `/me` | Bearer | Current user |

### Accounts

| Method | Path | Description |
|---|---|---|
| GET | `/accounts` | List accounts |
| POST | `/accounts` | Create account |
| GET | `/accounts/:id` | Get account |
| PATCH | `/accounts/:id` | Update name/type |
| DELETE | `/accounts/:id` | Delete account |

### Categories

| Method | Path | Description |
|---|---|---|
| GET | `/categories?type=income\|expense` | List categories |
| POST | `/categories` | Create category |
| PUT | `/categories/:id` | Update category |
| DELETE | `/categories/:id` | Delete category |

### Transactions

| Method | Path | Description |
|---|---|---|
| GET | `/transactions` | List with filters (`from`, `to`, `account`, `category`, `type`, `page`, `per_page`) |
| POST | `/transactions` | Create transaction |
| GET | `/transactions/:id` | Get transaction |
| PATCH | `/transactions/:id` | Update transaction |
| DELETE | `/transactions/:id` | Delete transaction |
| GET | `/transactions/export` | Export as CSV (same filter params) |
| POST | `/transactions/import` | Import from CSV (`multipart/form-data`, field: `file`) |

### Budgets

| Method | Path | Description |
|---|---|---|
| GET | `/budgets?period=weekly\|monthly` | List budgets |
| POST | `/budgets` | Create budget |
| PATCH | `/budgets/:id` | Update budget |
| DELETE | `/budgets/:id` | Delete budget |

### Transfers

| Method | Path | Description |
|---|---|---|
| GET | `/transfers` | List transfers |
| POST | `/transfers` | Create transfer |

### Saving Goals

| Method | Path | Description |
|---|---|---|
| GET | `/saving-goals` | List goals |
| POST | `/saving-goals` | Create goal |
| PUT | `/saving-goals/:id` | Update goal |
| POST | `/saving-goals/:id/contribute` | Add contribution |
| DELETE | `/saving-goals/:id` | Delete goal |

### Reports

| Method | Path | Description |
|---|---|---|
| GET | `/reports/weekly?weekOf=YYYY-MM-DD` | Weekly report |
| GET | `/reports/monthly?month=YYYY-MM` | Monthly report |

### Exchange Rates

| Method | Path | Description |
|---|---|---|
| GET | `/exchange-rates?from=USD&to=KZT` | Get exchange rate |

### Health

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/healthz` | Liveness |
| GET | `/api/v1/readyz` | Readiness (checks DB + Redis) |

## Response Format

All endpoints return a consistent JSON envelope:

```json
{ "error": false, "data": { ... } }
```

Errors:

```json
{ "error": true, "code": "not_found", "message": "resource not found" }
```

Paginated responses:

```json
{
  "error": false,
  "data": {
    "items": [...],
    "pagination": { "page": 1, "per_page": 20, "total": 100, "total_pages": 5 }
  }
}
```

## Migrations

```bash
make migrate-up          # apply all pending
make migrate-down        # roll back last
make migrate-status      # show applied/pending
make migrate-new name=add_indexes  # create new migration
```

## Makefile Reference

```bash
make run       # run the server
make build     # build binary to bin/fintrack
make test      # run tests
make lint      # run golangci-lint
make seed      # seed the database
make docker-up / docker-down
```
