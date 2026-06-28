# API Contract: Receipt OCR Review Flow (`/api/v1`)

## 1) Upload Receipt

### `POST /api/v1/receipts`

**Auth**: Required

**Request**:
- `multipart/form-data`
- `image` (image, required)

**Response**: `202 Accepted`

```json
{
  "receipt_id": "rcpt_123",
  "status": "PENDING",
  "duplicate": false
}
```

**Idempotent duplicate response** (`same user + same image`):

```json
{
  "receipt_id": "rcpt_123",
  "status": "PENDING",
  "duplicate": true
}
```

## 2) Get Editable Summary

### `GET /api/v1/receipts/{id}`

**Auth**: Required (owner-only)

**Response**: `200 OK`

```json
{
  "id": "rcpt_123",
  "status": "NEEDS_REVIEW",
  "store": {
    "name": "Central Market",
    "branch": "Downtown",
    "address": "Av. X #123"
  },
  "purchase_date": "2026-06-21",
  "total": 42.50,
  "items": [
    {
      "raw_text": "ARROZ 1KG",
      "quantity": 1,
      "unit_price": 2.40,
      "currency": "USD"
    }
  ]
}
```

Unreadable OCR can still return `NEEDS_REVIEW` with empty/partial `items`.

## 3) Confirm Corrected Receipt

### `POST /api/v1/receipts/{id}/confirm`

**Auth**: Required (owner-only)

**Request body**:

```json
{
  "store": {
    "name": "Central Market",
    "branch": "Downtown",
    "address": "Av. X #123"
  },
  "purchase_date": "2026-06-21",
  "total": 42.50,
  "items": [
    {
      "raw_text": "ARROZ 1KG",
      "quantity": 1,
      "unit_price": 2.40,
      "currency": "USD"
    }
  ]
}
```

**Success response**: `204 No Content`

**Validation failure** (`missing currency`): `400 Bad Request`

```json
{
  "error": "confirm receipt: item currency is required"
}
```

## 4) State & Behavior Rules

- Upload always asynchronous (`202`) if accepted.
- OCR provider implementation is swappable behind `OCRProvider` without use-case changes.
- Receipt status transitions:
  - upload accepted -> `PENDING`
  - OCR/parsing attempt complete -> `NEEDS_REVIEW`
  - successful confirm -> `CONFIRMED`
- Confirmation emits downstream integration signals (loyalty, aggregate recompute) but does not implement those modules.
