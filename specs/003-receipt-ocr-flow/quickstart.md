# Quickstart: Receipt OCR Processing and Review

## Prerequisites

- Docker and Docker Compose available
- API service environment configured
- MinIO bucket and credentials configured via environment variables

## Environment Variables

### S3-Compatible Storage (used by same adapter in all environments)

- `S3_ENDPOINT`
- `S3_ACCESS_KEY`
- `S3_SECRET_KEY`
- `S3_BUCKET`
- `S3_USE_SSL`

### Development Values (MinIO)

- `S3_ENDPOINT=minio:9000`
- `S3_ACCESS_KEY=minioadmin`
- `S3_SECRET_KEY=minioadmin`
- `S3_BUCKET=receipts`
- `S3_USE_SSL=false`

### Production Values (Hetzner Object Storage)

- Same variable names; only endpoint and credentials differ.

## Local Run

1. `docker compose up -d --build`
2. Verify:
   - API health endpoint responds
   - OCR service health endpoint responds
   - MinIO console is reachable

Example checks:

```bash
curl http://localhost:8080/api/v1/health
curl http://localhost:8081/health
```

## End-to-End Validation

1. Upload receipt image -> expect `202` and `PENDING`.
2. Poll receipt detail -> expect `NEEDS_REVIEW` after processing.
3. Confirm corrected payload -> expect `204` and final `CONFIRMED` state.
4. Confirm without item currency -> expect `400` and no invalid observation persistence.
5. Upload duplicate same-user same-image -> expect idempotent response with same `receipt_id`.

## Parser Fixture Validation

- Run parser tests using supermarket OCR-text fixtures under `internal/usecase/fixtures/ocr/`.
- Validate readable, partial, and unreadable scenarios.

## Recorded Evidence (2026-06-24)

- `go test ./...` passed.
- `docker compose up -d --build` passed for `api`, `postgres`, `redis`, `minio`, `ocr`.
- Upload test:
  - response: `202`
  - `receipt_id`: `9400671d-fb34-4518-ae90-1b86bcc9e6b4`
  - initial status: `PENDING`
- Processing poll test:
  - transitioned to `NEEDS_REVIEW`
  - parsed sample returned store `SUPERMARKET CENTRAL` and one item.
- Confirmation test:
  - valid payload response: `204`
  - missing-currency payload response: `400`
