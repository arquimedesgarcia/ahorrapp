# Implementation Plan: Receipt OCR Processing and Review

**Branch**: `003-receipt-ocr-flow` | **Date**: 2026-06-24 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/003-receipt-ocr-flow/spec.md`

## Summary

Deliver an end-to-end receipt workflow where authenticated users upload receipt images, a concurrent Go worker consumes OCR jobs from Redis, receipt text is extracted through a replaceable `OCRProvider`, parsed into editable fields, and user confirmation persists normalized observations with mandatory currency. Storage remains S3-compatible through one adapter configured by environment variables for MinIO (development) and Hetzner Object Storage (production).

## Technical Context

**Language/Version**:
- Go 1.23+ (API + worker)
- Python 3.11+ (OCR microservice)

**Primary Dependencies**:
- Go API: chi, pgx, go-redis, minio-go, golang-migrate
- OCR service: FastAPI, Uvicorn, PaddleOCR

**Storage**:
- PostgreSQL 16 + PostGIS (core business entities)
- Redis (job queue and retry/dead-letter lists)
- S3-compatible object storage via one adapter (MinIO in dev, Hetzner in prod)

**Testing**:
- Go unit tests for parser and use cases
- HTTP integration tests for receipt endpoints
- OCR provider contract tests against a mocked FastAPI response

**Target Platform**: Dockerized Linux services, local-first

**Project Type**: Backend web service + background worker + sidecar OCR microservice

**Performance Goals**:
- Upload acceptance p95 < 2s
- Queue-to-review transition p95 < 60s under nominal local load

**Constraints**:
- Maintain strict state machine: `PENDING -> NEEDS_REVIEW -> CONFIRMED | REJECTED`
- Keep extraction/storage replaceable through domain ports only
- Ensure currency is mandatory for persisted `PriceObservation`

**Scale/Scope**:
- MVP receipt ingestion and review flow
- No loyalty points implementation (emit-only behavior remains outside scope)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Article I — Clean Architecture in Go backend**
  - `OCRProvider` and `StorageProvider` stay as domain ports.
  - Redis queue, S3 storage, FastAPI OCR are adapters only.
  - Use cases depend on ports/entities, not infrastructure.
  - **Status**: PASS.

- **Article IV — Explicit, versioned contracts**
  - Receipt endpoints remain under `/api/v1/...`.
  - Contract file documents request/response/error shapes.
  - OCR microservice contract documented separately as internal integration.
  - **Status**: PASS.

- **Article V — Data, currency, and normalization**
  - Migrations include `stores`, `products`, `receipts`, `receipt_items`, `price_observations`.
  - Confirmation flow enforces mandatory currency on every observation.
  - Product normalization handled before observation persistence.
  - **Status**: PASS.

**Gate Result**: PASS - no constitutional deviations identified.

## Project Structure

### Documentation (this feature)

```text
specs/003-receipt-ocr-flow/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── receipt-api-contract.md
│   └── ocr-service-contract.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/
└── api/

internal/
├── domain/
│   ├── entities/
│   └── ports/
├── usecase/
│   └── fixtures/ocr/
├── adapter/
│   ├── http/
│   ├── postgres/
│   ├── redis/
│   ├── storage/
│   └── ocr/
└── config/

ocr-service/
├── app/
│   └── main.py
├── requirements.txt
└── Dockerfile

migrations/
```

**Structure Decision**: Keep a single Go service with in-process worker and add a separately deployable OCR microservice container. This satisfies local-first constraints while preserving adapter boundaries.

## Phase 0: Research

1. Queue model over Redis (job list + retry/backoff + dead-letter handling)
2. OCR provider contract design using `RawOCRResult`
3. S3 adapter compatibility strategy for MinIO/Hetzner
4. Parser strategy and fixture format for supermarket OCR text

Output: `research.md`

## Phase 1: Design and Contracts

1. Define entities and state transitions in `data-model.md`
2. Define `/api/v1` contract in `contracts/receipt-api-contract.md`
3. Define OCR microservice integration contract in `contracts/ocr-service-contract.md`
4. Define validation flow and environment setup in `quickstart.md`

Output: design docs ready for `/speckit.tasks`

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
