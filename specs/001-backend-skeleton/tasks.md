# Tasks: Backend Skeleton

**Input**: Design documents from `/specs/001-backend-skeleton/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are included because the specification explicitly requires at least one health-endpoint test and architecture validation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Every task includes an exact file path

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Initialize project scaffold and developer baseline

- [ ] T001 Initialize Go module and base dependencies in `go.mod`
- [ ] T002 Create root ignore rules for local/dev artifacts in `.gitignore`
- [ ] T003 Create environment template for local stack defaults in `.env.example`
- [ ] T004 [P] Create API entrypoint scaffold in `cmd/api/main.go`
- [ ] T005 [P] Create feature docs index note in `specs/001-backend-skeleton/quickstart.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core architecture and infrastructure blocks required before user-story work

**CRITICAL**: No user story work can begin until this phase is complete

- [ ] T006 Define typed runtime configuration loader in `internal/config/config.go`
- [ ] T007 [P] Define health and infra ports in `internal/domain/ports/health.go`
- [ ] T008 [P] Define repository port placeholder in `internal/domain/ports/repository.go`
- [ ] T009 [P] Define storage provider port placeholder in `internal/domain/ports/storage.go`
- [ ] T010 [P] Define OCR provider port placeholder in `internal/domain/ports/ocr.go`
- [ ] T011 [P] Define cache port placeholder in `internal/domain/ports/cache.go`
- [ ] T012 Implement postgres client factory with pgx in `internal/adapter/postgres/client.go`
- [ ] T013 Implement redis client factory with go-redis in `internal/adapter/redis/client.go`

**Checkpoint**: Foundation ready - user story implementation can now begin in priority order

---

## Phase 3: User Story 1 - Start the full stack locally (Priority: P1) 🎯 MVP

**Goal**: Developers can run the entire backend stack locally with one command and no cloud accounts

**Independent Test**: Run `docker compose up --build` and verify API, PostgreSQL+PostGIS, Redis, and MinIO containers start successfully.

### Tests for User Story 1

- [ ] T014 [P] [US1] Add compose startup smoke test script in `scripts/smoke/compose-up.ps1`

### Implementation for User Story 1

- [ ] T015 [US1] Add multi-stage Go container build in `Dockerfile`
- [ ] T016 [US1] Define local stack services and healthchecks in `docker-compose.yml`
- [ ] T017 [US1] Add migration runner helper for local/dev in `scripts/migrate.sh`
- [ ] T018 [US1] Add initial migration up script in `migrations/000001_create_health_table.up.sql`
- [ ] T019 [US1] Add initial migration down script in `migrations/000001_create_health_table.down.sql`
- [ ] T020 [US1] Wire startup sequence (config, migrations, server boot) in `cmd/api/main.go`

**Checkpoint**: User Story 1 is independently testable via `docker compose up --build`

---

## Phase 4: User Story 2 - Verify all dependencies are alive (Priority: P1)

**Goal**: Developers can call `/api/v1/health` and see per-dependency status for PostgreSQL and Redis

**Independent Test**: Call `GET /api/v1/health` with all services up and with one service down; verify `status` and dependency fields match contract.

### Tests for User Story 2

- [ ] T021 [P] [US2] Add use-case unit tests for dependency status aggregation in `internal/usecase/health_test.go`
- [ ] T022 [P] [US2] Add HTTP handler integration test for `/api/v1/health` in `internal/adapter/http/health_handler_test.go`

### Implementation for User Story 2

- [ ] T023 [US2] Implement postgres health adapter using ping checks in `internal/adapter/postgres/health_check.go`
- [ ] T024 [US2] Implement redis health adapter using ping checks in `internal/adapter/redis/health_check.go`
- [ ] T025 [US2] Implement health use case orchestration in `internal/usecase/health.go`
- [ ] T026 [US2] Implement v1 health endpoint handler and response mapping in `internal/adapter/http/health_handler.go`
- [ ] T027 [US2] Register router, middleware, and `/api/v1/health` route in `internal/adapter/http/router.go`

**Checkpoint**: User Story 2 is independently testable via endpoint contract behavior

---

## Phase 5: User Story 3 - Trust architecture layer discipline (Priority: P2)

**Goal**: Domain dependencies point inward only; infra remains behind ports

**Independent Test**: Run architecture checks and confirm domain/usecase layers have zero direct infra imports and adapter-only implementations.

### Tests for User Story 3

- [ ] T028 [P] [US3] Add architecture dependency guard test for domain layer in `internal/domain/architecture_test.go`
- [ ] T029 [P] [US3] Add architecture dependency guard test for usecase layer in `internal/usecase/architecture_test.go`

### Implementation for User Story 3

- [ ] T030 [US3] Add storage adapter stub implementing domain port in `internal/adapter/storage/client.go`
- [ ] T031 [US3] Add OCR adapter stub implementing domain port in `internal/adapter/ocr/client.go`
- [ ] T032 [US3] Add dependency-check script for local validation in `scripts/check-architecture.ps1`
- [ ] T033 [US3] Wire architecture validation command into quickstart guidance in `specs/001-backend-skeleton/quickstart.md`

**Checkpoint**: User Story 3 is independently testable via architecture checks and adapter/port boundaries

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final hardening and consistency checks across all user stories

- [ ] T034 [P] Update health endpoint contract details and examples in `specs/001-backend-skeleton/contracts/health-contract.md`
- [ ] T035 [P] Align quickstart validation steps with actual commands and outputs in `specs/001-backend-skeleton/quickstart.md`
- [ ] T036 Run full verification suite (`go test ./...` and `docker compose up`) and record results in `specs/001-backend-skeleton/quickstart.md`
- [ ] T037 Perform final code cleanup and import-boundary review across `internal/`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies
- **Phase 2 (Foundational)**: Depends on Phase 1; blocks all user stories
- **Phase 3 (US1)**: Depends on Phase 2
- **Phase 4 (US2)**: Depends on Phase 2 and benefits from US1 stack readiness
- **Phase 5 (US3)**: Depends on Phase 2; can run after US2 wiring is present
- **Phase 6 (Polish)**: Depends on all story phases

### User Story Dependencies

- **US1 (P1)**: Independent after Foundational; establishes runnable environment
- **US2 (P1)**: Independent business value after Foundational; operationally validated with US1 stack
- **US3 (P2)**: Independent quality gate after Foundational; validates architecture regardless of future features

### Within Each User Story

- Tests first where included (US2, US3)
- Adapters/usecases before route wiring
- Story checkpoint must pass before moving on

### Parallel Opportunities

- Foundation ports tasks `T007` to `T011` are parallelizable
- US2 tests `T021` and `T022` are parallelizable
- US3 architecture tests `T028` and `T029` are parallelizable
- Polish docs updates `T034` and `T035` are parallelizable

---

## Parallel Example: User Story 2

```bash
# Run both US2 tests in parallel:
Task: "T021 [US2] internal/usecase/health_test.go"
Task: "T022 [US2] internal/adapter/http/health_handler_test.go"

# Then implement adapters in order:
Task: "T023 [US2] internal/adapter/postgres/health_check.go"
Task: "T024 [US2] internal/adapter/redis/health_check.go"
```

## Parallel Example: User Story 3

```bash
# Run architecture checks in parallel:
Task: "T028 [US3] internal/domain/architecture_test.go"
Task: "T029 [US3] internal/usecase/architecture_test.go"
```

---

## Implementation Strategy

### MVP First (US1 only)

1. Complete Phase 1 (Setup)
2. Complete Phase 2 (Foundational)
3. Complete Phase 3 (US1)
4. Validate stack startup with `docker compose up --build`
5. Demo runnable local environment

### Incremental Delivery

1. Setup + Foundational
2. Deliver US1 (local stack)
3. Deliver US2 (health endpoint contract)
4. Deliver US3 (architecture inward-dependency guarantees)
5. Polish and full validation

### Suggested MVP Scope

- **MVP**: Phase 1 + Phase 2 + Phase 3 (US1)
- **Operational MVP**: Add Phase 4 (US2) to get health observability

---

## Notes

- All tasks follow required checklist format: `- [ ] T### [P?] [US#?] Description with file path`
- Story labels are included only for user-story phases
- Paths are repository-relative and executable by an implementation agent
