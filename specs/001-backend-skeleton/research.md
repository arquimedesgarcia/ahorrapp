# Research: Backend Skeleton

**Feature**: 001-backend-skeleton
**Phase**: 0 — Technology decisions and best practices

## 1. Go Project Layout

**Decision**: `cmd/` + `internal/` layout with Clean Architecture sub-packages.

**Rationale**:
- `cmd/api/` is the idiomatic Go entry point (one binary, one `main.go`).
- `internal/` enforces Go's module visibility boundary — code in `internal/` cannot be
  imported by other modules, providing a compile-time guarantee against accidental
  external coupling.
- Sub-packages follow Clean Architecture: `domain/ports/` (interfaces), `usecase/`
  (application logic), `adapter/` (infrastructure implementations).
- No `/pkg/` directory: the skeleton has no public library surface to export.

**Alternatives considered**:
- Flat `pkg/` layout: rejected — provides no visibility enforcement; downstream
  modules could import internals.
- `app/` or `server/` prefix: rejected — less idiomatic than `cmd/` for entry points.

## 2. HTTP Router

**Decision**: `chi` v5 (lightweight, idiomatic, stdlib-compatible).

**Rationale**:
- `chi` is a thin layer over `net/http` — it wraps `http.Handler` and `http.HandlerFunc`,
  not a custom type system. Handlers remain compatible with Go's standard library.
- Middleware chaining is built-in (logging, recovery, timeout).
- Route groups support `/api/v1` prefix naturally.
- Zero dependencies beyond `net/http`.
- The user explicitly requested "net/http with chi or similar lightweight".

**Alternatives considered**:
- `net/http` alone (stdlib ServeMux): rejected — Go 1.22+ ServeMux supports method
  routing but lacks middleware chaining and sub-routers needed for `/api/v1` grouping.
- `gin` or `echo`: rejected — both are heavier frameworks that violate the "simplest
  solution" principle (Art. VI). They introduce custom context types, making handlers
  incompatible with `net/http` testing.

## 3. PostgreSQL Driver & Connection Pool

**Decision**: `pgx` v5 (direct driver, no ORM), connection pool via `pgxpool`.

**Rationale**:
- `pgx` is the de facto standard PostgreSQL driver for Go (performance, native PostgreSQL
  type support, JSONB, PostGIS).
- `pgxpool` manages connection lifecycle, health checks (`Ping`), and connection reuse.
- No ORM (GORM, Bun): the skeleton has no entities to map; even when entities arrive in
  later epics, raw SQL or a thin query builder keeps the domain decoupled.
- PostGIS extension is baked into the `postgis/postgis:16-3.4` Docker image; no
  additional setup required.

**Alternatives considered**:
- `database/sql` + `lib/pq`: rejected — `pgx` is faster, supports PostgreSQL-specific
  features (COPY, LISTEN/NOTIFY), and avoids the `database/sql` interface overhead.
- ORM (GORM, Bun): rejected — violates Art. VI (simplicity), Art. I (domain must not
  depend on frameworks), and would require the domain to import the ORM.

## 4. Redis Client

**Decision**: `go-redis` v9.

**Rationale**:
- Most popular Redis client for Go; supports sentinel, cluster, and standalone modes.
- `Ping().Result()` provides a simple health-check method.
- The user explicitly requested "go-redis client".

**Alternatives considered**:
- `rueidis`: newer, higher throughput — overkill for this MVP; `go-redis` is the
  established choice and what the user requested.

## 5. Database Migrations

**Decision**: `golang-migrate` v4, run as a pre-start step.

**Rationale**:
- `golang-migrate` is the most widely used migration tool in Go; supports PostgreSQL
  natively, versioned up/down migrations with a `schema_migrations` tracking table.
- Can run as a CLI command, as a library embedded in `main.go`, or via an init
  container in Docker Compose.
- For this skeleton, migrations run via `cmd/api/main.go` at startup (call
  `migrate.Up()` before starting the HTTP server), keeping the local dev workflow as
  simple as `docker compose up`.
- The user explicitly requested "golang-migrate".

**Alternatives considered**:
- `atlas` or `goose`: both are viable but less widely adopted than `golang-migrate`.
  The user's decision is explicit: golang-migrate.
- Init container in Docker Compose: adds complexity (wait-for-it scripts, sequential
  dependency) — rejected in favor of in-process migration for the skeleton.

## 6. Environment Configuration

**Decision**: `caarlos0/env` v11 (struct-tag-based env parsing with validation).

**Rationale**:
- Parses environment variables into a typed Go struct using struct tags.
- Supports required-field validation, default values, and prefix namespacing.
- Zero boilerplate: define a struct, call `env.Parse()`, done.
- Lightweight library with no dependencies beyond stdlib.

**Alternatives considered**:
- `viper`: rejected — too heavy for env-var-only config (includes file formats,
  remote config, etc.) — violates Art. VI.
- `os.Getenv()` manually: rejected — verbose and error-prone as the number of env vars
  grows (at least 6: DB_URL, REDIS_ADDR, MINIO_ENDPOINT, MINIO_ACCESS_KEY,
  MINIO_SECRET_KEY, SERVER_PORT).

## 7. Docker Build

**Decision**: Multi-stage Dockerfile — Go 1.23 builder stage → distroless runtime stage.

**Rationale**:
- Builder stage: `golang:1.23-alpine`, compiles a statically linked binary with
  `CGO_ENABLED=0`.
- Runtime stage: `gcr.io/distroless/static-debian12` — the smallest possible image
  (~2 MB for the Go binary + distroless base). No shell, no package manager,
  minimal attack surface.
- Multi-stage keeps the final image small (~10 MB total) and separates build
  dependencies from runtime.

**Alternatives considered**:
- `scratch`: even smaller but lacks CA certificates (needed for TLS to cloud
  PostgreSQL/Redis later) and timezone data. Distroless includes both.
- `alpine`: includes a shell and package manager — larger image, larger attack
  surface. Distroless is preferred per Art. VI (simplest secure solution).

## 8. Docker Compose Health Checks

**Decision**: Docker native `healthcheck` on PostgreSQL and Redis containers;
  Go-level health endpoint for runtime verification.

**Rationale**:
- Docker `healthcheck` ensures containers are actually ready (not just "started")
  before `depends_on` conditions are satisfied.
- PostgreSQL healthcheck: `pg_isready -U user`.
- Redis healthcheck: `redis-cli ping`.
- The Go `/health` endpoint is for **runtime** verification (e.g., after a
  transient network partition), while Docker healthchecks are for **startup**
  ordering. Both are needed.

## 9. Testing Strategy

**Decision**: Standard library `testing` + `httptest`; no external framework for this
  phase.

**Rationale**:
- `usecase/health_test.go`: Unit tests with mock implementations of port interfaces
  (hand-coded mocks, no mock library — simple, readable, zero deps).
- `adapter/http/health_handler_test.go`: Integration test using `httptest.NewServer`
  to exercise the real handler with mock use cases.
- Both tests run with `go test ./...` — no special runner required.
- More sophisticated testing (testify, gomock) can be introduced later when the
  test surface grows, but for a single endpoint it's over-engineering (Art. VI).

**Alternatives considered**:
- `testify`: rejected for this phase — adds a dependency for features (assertions,
  mocks) that the standard library covers with a few extra lines. Can be adopted
  later if needed.
- `gomock` or `moq`: rejected — hand-coded mocks for 3-4 interfaces are trivial
  and more readable than generated code.

## 10. PostGIS in the Skeleton

**Decision**: Use `postgis/postgis:16-3.4` Docker image; no migration to enable
  the extension (it is already enabled in the image).

**Rationale**:
- The official PostGIS Docker image automatically enables the extension on database
  creation.
- No migration SQL needed for `CREATE EXTENSION postgis;` — it's already active.
- When the Store entity arrives (later epic), columns of type `geometry(Point, 4326)`
  are available immediately.

## Summary of Technology Stack

| Component | Choice | Version |
|-----------|--------|---------|
| Language | Go | 1.23+ |
| HTTP Router | chi | v5 |
| DB Driver | pgx | v5 |
| DB Migrations | golang-migrate | v4 |
| Redis Client | go-redis | v9 |
| Env Config | caarlos0/env | v11 |
| S3 Client | minio-go (stub) | v7 |
| Container base | distroless/static | debian12 |
| Test framework | stdlib testing | Go 1.23 |
