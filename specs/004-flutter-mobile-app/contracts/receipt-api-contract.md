# Contract: Receipt API (`/api/v1/receipts`)

> **Source**: Derived from `specs/003-receipt-ocr-flow/contracts/receipt-api-contract.md`.
> Flutter app MUST consume ONLY these documented shapes. No undocumented fields.

## POST `/api/v1/receipts`

Upload a receipt image for processing.

- **Auth**: Required (`Authorization: Bearer <token>`)
- **Content-Type**: `multipart/form-data`

### Form Fields

| Field | Type | Required |
|-------|------|----------|
| `image` | file (jpg/png) | yes |

### Success (`202`)

```json
{
  "receipt_id": "uuid",
  "status": "PENDING",
  "duplicate": false
}
```

### Duplicate Idempotent Success (`202`)

```json
{
  "receipt_id": "uuid",
  "status": "PENDING",
  "duplicate": true
}
```

### Errors

| Status | Body | Condition |
|--------|------|-----------|
| `401` | `{"error": "invalid or expired token"}` | Token invalid |
| `400` | `{"error": "image is required"}` | No image file |
| `413` | `{"error": "image too large"}` | Exceeds max size |

---

## GET `/api/v1/receipts/{id}`

Retrieve receipt detail for review.

- **Auth**: Required (owner only)

### Success (`200`) — NEEDS_REVIEW state

```json
{
  "receipt_id": "uuid",
  "status": "NEEDS_REVIEW",
  "store": {
    "name": "Central Market",
    "branch": "Downtown",
    "address": "Av. X #123"
  },
  "purchase_date": "2026-06-24",
  "total": 42.5,
  "items": [
    {
      "raw_text": "ARROZ 1KG",
      "quantity": 1,
      "unit_price": 2.4,
      "currency": "USD"
    }
  ]
}
```

Note: `PENDING` receipts return `status: "PENDING"` with null store/date/total/items. Flutter app polls until status changes to `NEEDS_REVIEW`.

### Errors

| Status | Body | Condition |
|--------|------|-----------|
| `401` | `{"error": "invalid or expired token"}` | Token invalid |
| `403` | `{"error": "not your receipt"}` | Not owner |
| `404` | `{"error": "receipt not found"}` | Invalid ID |

---

## POST `/api/v1/receipts/{id}/confirm`

Submit corrected receipt data for confirmation.

- **Auth**: Required (owner only)
- **Content-Type**: `application/json`

### Request Body

```json
{
  "store": {
    "name": "Central Market",
    "branch": "Downtown",
    "address": "Av. X #123"
  },
  "purchase_date": "2026-06-24",
  "total": 42.5,
  "items": [
    {
      "raw_text": "ARROZ 1KG",
      "quantity": 1,
      "unit_price": 2.4,
      "currency": "USD"
    }
  ]
}
```

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `store.name` | string | yes | Non-empty |
| `purchase_date` | string | no | ISO date format |
| `total` | number | no | Positive |
| `items[].raw_text` | string | yes | Non-empty |
| `items[].currency` | string | yes | "USD" or "Bs." |

### Success (`200`)

```json
{
  "points_earned": 10
}
```

### Errors

| Status | Body | Condition |
|--------|------|-----------|
| `400` | `{"error": "confirm receipt: item currency is required"}` | Missing currency |
| `401` | `{"error": "invalid or expired token"}` | Token invalid |
| `403` | `{"error": "not your receipt"}` | Not owner |
| `409` | `{"error": "receipt already confirmed"}` | Already confirmed |
