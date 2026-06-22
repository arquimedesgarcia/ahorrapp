# Data Model: Backend Skeleton

**Feature**: 001-backend-skeleton
**Phase**: 1 — Entities and configuration schema

## Overview

This feature establishes infrastructure scaffolding. It does **not** introduce domain business
entities (User, Receipt, Product, Store, etc.). Those belong to later epics. The data model
here covers:

1. **Health status** — the in-memory shape of the `/api/v1/health` response.
2. **Migration tracking** — the `schema_migrations` table managed by golang-migrate.
3. **Configuration** — the env-var schema loaded at startup.

## 1. Health Status (Response Entity)

Not persisted. Built at request time by the `HealthCheck` use case and serialized to JSON.

| Field | Type | Description |
|-------|------|-------------|
| `status` | string | `"ok"` if all dependencies are reachable; `"degraded"` if any is down |
| `dependencies` | object | Map of dependency name → dependency status |
| `dependencies.postgres` | object | `{ name: "postgres", reachable: true/false, error?: string }` |
| `dependencies.redis` | object | `{ name: "redis", reachable: true/false, error?: string }` |

**Validation rules**:
- `status` MUST be `"ok"` when all `dependencies.*.reachable` are `true`.
- `status` MUST be `"degraded"` when any `dependencies.*.reachable` is `false`.
- HTTP status code: `200` when `ok`, `200` when `degraded` (the endpoint itself is alive).
  The service does NOT return 5xx for degraded dependencies — 5xx means the handler itself
  crashed.
- `error` field is present only when `reachable` is `false`; omitted when `true`.

**Example (all healthy)**:
```json
{
  "status": "ok",
  "dependencies": {
    "postgres": { "name": "postgres", "reachable": true },
    "redis": { "name": "redis", "reachable": true }
  }
}
```

**Example (postgres down)**:
```json
{
  "status": "degraded",
  "dependencies": {
    "postgres": { "name": "postgres", "reachable": false, "error": "dial tcp: connection refused" },
    "redis": { "name": "redis", "reachable": true }
  }
}
```

## 2. Migration Tracking Table

Managed entirely by `golang-migrate`. This table is infrastructure metadata, not a domain
entity.

**Table**: `schema_migrations`

| Column | Type | Description |
|--------|------|-------------|
| `version` | bigint | Migration version number (e.g., `1` for `000001`) |
| `dirty` | boolean | Whether the migration failed partway through |

**Behavior**:
- `golang-migrate` creates this table automatically on first run.
- Each migration increments `version` and sets `dirty = false` on success.
- If a migration fails, `dirty = true` and the version stays at the last successful one.
- No manual SQL against this table is required.

The initial migration (`000001`) in this skeleton creates no application tables — it exists
to validate the migration pipeline works end-to-end. It could insert a sentinel row or simply
be a no-op (`SELECT 1`). The first real table creation happens in E2 (User auth).

## 3. Configuration Schema

Loaded from environment variables at startup. All fields are required unless marked
optional.

| Env Variable | Type | Default | Description |
|-------------|------|---------|-------------|
| `SERVER_PORT` | int | `8080` | HTTP listen port |
| `DATABASE_URL` | string | `postgres://ahorrapp:ahorrapp@localhost:5432/ahorrapp?sslmode=disable` | PostgreSQL connection string |
| `REDIS_ADDR` | string | `localhost:6379` | Redis host:port |
| `REDIS_PASSWORD` | string | `""` | Redis password (empty for local dev) |
| `MINIO_ENDPOINT` | string | `localhost:9000` | MinIO S3 endpoint |
| `MINIO_ACCESS_KEY` | string | `minioadmin` | MinIO access key |
| `MINIO_SECRET_KEY` | string | `minioadmin` | MinIO secret key |
| `MINIO_BUCKET` | string | `receipts` | Default bucket for receipt images |
| `LOG_LEVEL` | string | `info` | Log level (debug, info, warn, error) |

**Validation at startup**:
- `DATABASE_URL` must be a valid PostgreSQL connection string.
- `REDIS_ADDR` must be a valid `host:port`.
- `SERVER_PORT` must be in range 1024–65535.
- If any required variable is missing or invalid, the service prints a clear message listing
  which variables are missing and exits with code 1.

**Security**:
- `.env` file MUST be listed in `.gitignore` — never committed.
- `.env.example` is committed with default local-dev values (no production secrets).
- All secrets live in env vars per Art. VII and FR-005.

## 4. What This Feature Does NOT Model

| Excluded entity | Rationale |
|----------------|-----------|
| User | E2 — authentication epic |
| Store (merchant) | E5 — receipt parsing epic |
| Receipt | E3 — receipt upload epic |
| ReceiptItem | E5 — receipt parsing epic |
| Product | E5 — receipt parsing epic |
| PriceObservation | E7 — price engine epic |
| PriceAggregate | E7 — price engine epic |
| LoyaltyTransaction | E9 — gamification epic |

Each of these will define its own migration, entity, and port in its respective epic. The
skeleton ensures the pipeline (migrations, ports, adapters, config) is in place so later
epics only need to add — never refactor the foundation.
