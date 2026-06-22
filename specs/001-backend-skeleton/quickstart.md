# Quickstart: Backend Skeleton Validation

**Feature**: 001-backend-skeleton
**Purpose**: End-to-end validation that the skeleton works before proceeding to `/speckit.implement`.

## Prerequisites

- Docker & Docker Compose installed and running (see `docs/01_GETTING_STARTED.md`, step 0.5)
- Git installed
- Go 1.23+ installed (for running tests outside containers, optional)

## 1. Clone and Configure

```bash
git clone <repo-url> ahorrapp
cd ahorrapp
git checkout 001-backend-skeleton
```

Copy the environment template:

```bash
cp .env.example .env
```

> The default values in `.env.example` match the Docker Compose service names. No editing
> is needed for local development.

## 2. Start the Stack

```bash
docker compose up --build
```

**Expected outcome** (within 60 seconds):

```
[+] Running 4/4
 ✔ Container ahorrapp-postgres  Healthy
 ✔ Container ahorrapp-redis     Healthy
 ✔ Container ahorrapp-minio     Started
 ✔ Container ahorrapp-api       Started
```

> The `--build` flag is only needed the first time or after code changes. Subsequent starts
> can use `docker compose up` alone and should complete in under 15 seconds.

**If something fails**:
- `ahorrapp-postgres` fails: check port 5432 is free.
- `ahorrapp-redis` fails: check port 6379 is free.
- `ahorrapp-api` fails: run `docker compose logs api` to see error details. Most common
  cause: `.env` missing or `DATABASE_URL` pointing to unreachable host.

## 3. Verify the Health Endpoint

```bash
curl -s http://localhost:8080/api/v1/health | python -m json.tool
```

**Expected output** (all healthy):

```json
{
  "status": "ok",
  "dependencies": {
    "postgres": {
      "name": "postgres",
      "reachable": true
    },
    "redis": {
      "name": "redis",
      "reachable": true
    }
  }
}
```

> Windows note: if `python -m json.tool` is unavailable, just `curl http://localhost:8080/api/v1/health`
> to see the raw JSON. The point is that you get a `200` with reachable dependencies.

## 4. Verify Degraded Behavior

Stop one dependency and confirm the health endpoint reports it accurately:

```bash
docker compose stop postgres
curl -s http://localhost:8080/api/v1/health
```

**Expected**: `"postgres": { "reachable": false, "error": "..." }`, `"status": "degraded"`,
HTTP status still `200`.

Restore the dependency:

```bash
docker compose start postgres
sleep 2
curl -s http://localhost:8080/api/v1/health
```

**Expected**: Back to `"status": "ok"` with both reachable.

## 5. Verify Architecture Discipline

From the project root, confirm the domain layer has zero infrastructure dependencies:

```bash
# List all non-stdlib imports in the domain
go list -f '{{ join .Imports "\n" }}' ./internal/domain/... 2>/dev/null | Select-String -NotMatch '^ahorrapp/' | Select-String -NotMatch '^$'
```

Or on macOS/Linux:
```bash
go list -f '{{ join .Imports "\n" }}' ./internal/domain/... | grep -v '^ahorrapp/' | grep -v '^$'
```

**Expected**: No output (or only stdlib packages like `context`, `time`, `errors`). If any
line contains `pgx`, `chi`, `go-redis`, `minio-go`, or any other infrastructure package,
the domain layer is compromised — fix before proceeding.

## 6. Run Tests

```bash
go test ./...
```

**Expected**: All tests pass. At minimum, the health use case test and the health handler
integration test must pass.

```bash
go test -v ./internal/usecase/
go test -v ./internal/adapter/http/
```

## 7. Tear Down

```bash
docker compose down
```

Rerun `docker compose up` from step 2 to confirm a clean restart. The migration should
be idempotent (no errors on re-run).

## Validation Checklist

- [ ] `docker compose up --build` starts all 4 containers
- [ ] `GET /api/v1/health` returns `200` with `"status": "ok"`
- [ ] Stopping Postgres → health reports `degraded` with postgres unreachable
- [ ] Restarting Postgres → health returns to `ok`
- [ ] `go list` confirms domain layer has zero infrastructure imports
- [ ] `go test ./...` passes (at least 2 tests)
- [ ] `docker compose down` + `docker compose up` clean restart succeeds

All checked → Skeleton is validated. Ready for E2 (User Auth).
