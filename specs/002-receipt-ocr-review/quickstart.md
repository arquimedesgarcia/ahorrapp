# Quickstart Validation: Receipt OCR Review Flow

## Prerequisites

- Backend skeleton and auth already running
- Local dependencies available (Postgres, Redis, MinIO, OCR service)
- Auth context available via `X-User-ID` header

## 1. Upload a Receipt

Send `POST /api/v1/receipts` with an authenticated image upload.

Example:

```bash
curl -X POST http://localhost:8080/api/v1/receipts \
  -H "X-User-ID: smoke-user" \
  -F "image=@receipt.jpg"
```

**Expected**:
- Response `202` with `receipt_id`
- Receipt persisted in `PENDING`
- One OCR job enqueued

## 2. Wait for Processing

Poll `GET /api/v1/receipts/{id}`.

**Expected**:
- Status becomes `NEEDS_REVIEW`
- Editable summary includes store/date/total/items (or partial/empty items for unreadable OCR)

## 3. Confirm Corrected Summary

Send `POST /api/v1/receipts/{id}/confirm` with corrected payload.

**Expected**:
- Response `204`
- Receipt status changes to `CONFIRMED`
- Canonical products resolved
- `PriceObservation` records created with mandatory currency

## 4. Edge-Case Validation

### Unreadable image
- Upload unreadable image.
- Expect `NEEDS_REVIEW` with editable manual-completion path.

### Unknown merchant
- Use receipt with unrecognized store text.
- Expect ability to create/associate new store during review.

### Duplicate same-user same-image upload
- Upload same image twice as same user.
- Expect idempotent duplicate behavior: existing `receipt_id` returned, no extra queue job.

### Missing currency in confirmation
- Attempt confirm with at least one item missing currency.
- Expect `400` validation error and receipt remains `NEEDS_REVIEW`.

## 5. Regression Safety Checks

- Swapping OCR adapter implementation does not require use-case modifications.
- Receipt state transitions remain valid (`PENDING -> NEEDS_REVIEW -> CONFIRMED`).
- No `PriceObservation` persisted without currency.

## 6. Recorded Validation Evidence (2026-06-24)

- `docker compose up -d --build` completed with all services up (`api`, `postgres`, `redis`, `minio`).
- Upload smoke test returned `202` with:
  - `receipt_id`: `00644c79-d64e-49e4-875d-c09a6b621162`
  - `status`: `PENDING`
  - `duplicate`: `false`
- Confirm smoke test returned `204`, and subsequent `GET /api/v1/receipts/{id}` returned `CONFIRMED`.
- Duplicate upload test returned same `receipt_id` with `duplicate: true`.
- Missing-currency confirm test returned `400`.
