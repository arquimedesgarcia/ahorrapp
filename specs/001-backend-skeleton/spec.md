# Feature Specification: Backend Skeleton

**Feature Branch**: `001-backend-skeleton`

**Created**: 2026-06-21

**Status**: Draft

**Input**: User description: "Feature: Go backend skeleton with Clean Architecture. The service
exposes a versioned REST API at /api/v1 with a health endpoint confirming PostgreSQL and Redis
reachability. The project structure follows Clean Architecture with domain ports. The full stack
runs locally via Docker Compose (Go API, PostgreSQL 16 + PostGIS, Redis, MinIO) with zero cloud
accounts. Includes versioned database migrations and environment-variable configuration. No
authentication or receipt logic — foundation only."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Start the full stack locally (Priority: P1)

A developer clones the repository and wants to bring up every backend service on their machine
with a single command, without creating cloud accounts or paying for services. They run the
startup command and all containers (API, database, cache, object storage) become operational.

**Why this priority**: This is the entry point for all development and testing. Without a working
local stack, no other feature can be built or verified.

**Independent Test**: Run `docker compose up` on a clean checkout. All containers reach a healthy
state without crashing. The developer can then call the health endpoint.

**Acceptance Scenarios**:

1. **Given** Docker is installed and running on the developer's machine, **When** the developer
   executes the startup command from the project root, **Then** within 60 seconds the API
   container, PostgreSQL container, Redis container, and MinIO container are all running without
   errors.
2. **Given** a previous shutdown left no residual state, **When** the developer starts the stack
   again, **Then** all services start cleanly with no manual cleanup required.

---

### User Story 2 - Verify all dependencies are alive (Priority: P1)

A developer starts the stack and needs to confirm that every dependency the backend relies on is
reachable before doing any further work or troubleshooting.

**Why this priority**: Health verification is the first diagnostic tool after startup. It gives
immediate feedback on whether the environment is correct, without requiring domain-specific tests
or user accounts.

**Independent Test**: Call `GET /api/v1/health` after `docker compose up`. The response reports
status of PostgreSQL and Redis individually.

**Acceptance Scenarios**:

1. **Given** the full stack is running and all dependencies are healthy, **When** the developer
   sends `GET /api/v1/health`, **Then** the response is `200 OK` with a body indicating
   PostgreSQL is reachable and Redis is reachable.
2. **Given** the stack is running but PostgreSQL has stopped, **When** the developer sends
   `GET /api/v1/health`, **Then** the response indicates PostgreSQL is unreachable while any other
   dependency status is reported accurately.
3. **Given** the stack is running but Redis has stopped, **When** the developer sends
   `GET /api/v1/health`, **Then** the response indicates Redis is unreachable while any other
   dependency status is reported accurately.

---

### User Story 3 - Trust the architecture's layer discipline (Priority: P2)

A developer wants assurance that the source code respects Clean Architecture: the domain layer
does not leak infrastructure concerns, and external dependencies are accessible only through
interfaces defined in the domain. This guarantee is structural — it prevents future features from
accidentally coupling business logic to frameworks or databases.

**Why this priority**: Layer discipline is a constitution requirement (Article I). It is verified
through automated structural checks that run as part of the project's build process. This story
validates the foundation is correct before more complex features are added on top.

**Independent Test**: Inspect the domain layer's dependencies. Confirm it references no HTTP
router, database driver, or infrastructure library. Run the project's structural validation and
verify zero infrastructure dependencies in the domain.

**Acceptance Scenarios**:

1. **Given** the project is built from source, **When** a reviewer examines the domain layer's
   dependencies, **Then** no direct reference to any HTTP framework, database driver, Redis
   client, S3 library, or OCR library is found in the domain code.
2. **Given** a new developer attempts to add an infrastructure dependency directly to the domain
   layer, **When** the project's structural validation runs, **Then** the validation rejects the
   change with an error explaining which layer boundary was violated.
3. **Given** at least one external dependency interface (port) is defined in the domain,
   **When** the project builds successfully, **Then** that interface is implemented by an adapter
   in an outer layer, not self-implemented in the domain.

---

### Edge Cases

- **PostgreSQL unreachable at startup**: What happens if the database container fails to start or
  crashes? The health endpoint must report the dependency as unhealthy rather than crashing the
  whole service. The API container itself remains alive and diagnosable.
- **Redis unreachable mid-session**: If Redis is available at startup but becomes unreachable
  later, the health endpoint must reflect the current (degraded) state on each request, not cache
  an old healthy result.
- **Port conflict on host**: If the default ports are already in use on the developer's machine,
  the startup must fail with a clear, actionable error message (not a silent hang or obscure
  connection-refused trace).
- **Missing or incomplete environment configuration**: If the developer has not provided required
  environment variables (e.g., database URL), the service must refuse to start with a clear
  message listing which variables are missing — not crash with a cryptic nil-pointer or
  connection error.
- **Multiple sequential starts and stops**: Running `docker compose down` followed by
  `docker compose up` must restore the stack to the same working state as a first-time start.
  No data corruption or stale state from the prior run.
- **Database migration on first run**: The first time the stack starts with an empty database, the
  migration system must apply all pending migrations automatically or via an explicit command that
  is clearly documented.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST expose a health-check endpoint at `GET /api/v1/health`.
- **FR-002**: The health endpoint MUST verify PostgreSQL connectivity and report its status
  (`reachable` or `unreachable`) in the response body.
- **FR-003**: The health endpoint MUST verify Redis connectivity and report its status
  (`reachable` or `unreachable`) in the response body.
- **FR-004**: System MUST start all services (Go API, PostgreSQL 16 + PostGIS, Redis, MinIO)
  via a single `docker compose up` command from the project root with no additional cloud
  accounts, sign-ups, or paid services.
- **FR-005**: All connection parameters (database URL, Redis address, MinIO endpoint and
  credentials, server port) MUST be configured through environment variables. Source code MUST
  NOT contain hardcoded credentials, connection strings, or secrets.
- **FR-006**: System MUST include a versioned database migration mechanism that applies schema
  changes in order and tracks which migrations have been applied.
- **FR-007**: The domain/application layer (entities and use cases) MUST NOT depend on or
  reference any HTTP framework, database driver, Redis client, object-storage SDK, OCR library,
  or any other infrastructure package.
- **FR-008**: External dependencies (PostgreSQL, Redis, Object Storage) MUST be accessed through
  interfaces (ports) defined in the domain layer, with concrete adapter implementations in an
  outer layer.
- **FR-009**: The PostgreSQL container MUST run with the PostGIS extension enabled from the first
  start, so store-geolocation features can be added later without altering the database
  provisioning step.
- **FR-010**: At least one automated test MUST exist that calls the health endpoint and verifies
  the response shape and status code.

### Key Entities

This feature establishes infrastructure scaffolding; it does not introduce domain business
entities (those belong to later epics — User, Receipt, Product, etc.). The following structural
artifacts are defined:

- **Health status response**: A structured response containing per-dependency status fields. Each
  dependency has at minimum a name and a reachable/unreachable indicator.
- **Migration state**: A system table tracking applied migrations (typically a schema version
  number and a timestamp). This is infrastructure metadata, not a domain entity.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer on a machine with only Docker and Git installed can go from `git clone`
  to a successful `GET /api/v1/health → 200` in under 5 minutes, including container image pulls
  on a first run.
- **SC-002**: `docker compose up` brings all containers to a ready state in under 60 seconds on
  subsequent starts (images already cached).
- **SC-003**: The health endpoint responds within 2 seconds when all dependencies are healthy.
- **SC-004**: When a dependency is down, the health endpoint still responds (does not crash) and
  correctly labels the down dependency, within 5 seconds (the connection-timeout threshold).
- **SC-005**: The domain package compiles with zero imports from infrastructure packages — 100%
  compliance with Article I of the constitution.
- **SC-006**: The entire stack operates with zero cloud accounts, zero paid services, and zero
  external API calls during development.

## Assumptions

- Docker and Docker Compose are already installed on the developer's machine (as documented in
  `docs/01_GETTING_STARTED.md`, Phase 0 step 0.5).
- The developer runs the stack on Windows, macOS, or Linux natively. Docker Desktop provides the
  engine; the compose file uses platform-agnostic configuration.
- The PostgreSQL container automatically enables PostGIS via its official Docker image tag (e.g.,
  `postgis/postgis:16-3.4`), no manual extension activation needed.
- MinIO is included as a running container but is not health-checked by the `/health` endpoint at
  this stage; its liveness can be verified through its own web console (port 9001) during
  development.
- Environment variables are provided via a `.env` file at the project root (not committed to git)
  with sensible defaults for local development.
- The database migration system runs as a pre-start step — either an init container in compose or
  a startup command in the API binary — before the API begins serving requests.
- The OCR service is not included in this skeleton; it will be added in Epic E4. Only its
  interface (port) is declared in the domain. No concrete OCR adapter is implemented yet.
- Authentication (JWT, user registration), receipt upload, and any business logic are explicitly
  out of scope. This spec only validates the infrastructure foundation.
